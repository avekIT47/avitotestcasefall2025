package webhook

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/user/pr-reviewer/internal/logger"
	"github.com/user/pr-reviewer/internal/models"
)

// EventType тип webhook события
type EventType string

const (
	EventPRCreated        EventType = "pr.created"
	EventPRMerged         EventType = "pr.merged"
	EventPRClosed         EventType = "pr.closed"
	EventReviewerAssigned EventType = "reviewer.assigned"
	EventReviewerChanged  EventType = "reviewer.changed"
	EventUserDeactivated  EventType = "user.deactivated"
)

// Payload данные webhook события
type Payload struct {
	Event     EventType              `json:"event"`
	Timestamp time.Time              `json:"timestamp"`
	Data      map[string]interface{} `json:"data"`
}

// Subscription подписка на webhook
type Subscription struct {
	ID        int64       `json:"id"`
	URL       string      `json:"url"`
	Events    []EventType `json:"events"`
	Secret    string      `json:"secret,omitempty"`
	Active    bool        `json:"active"`
	CreatedAt time.Time   `json:"created_at"`
}

// Deliverer интерфейс для доставки webhook
type Deliverer interface {
	Deliver(ctx context.Context, sub *Subscription, payload *Payload) error
}

// HTTPDeliverer HTTP реализация доставки webhook
type HTTPDeliverer struct {
	client *http.Client
	logger *logger.Logger
}

// NewHTTPDeliverer создает новый HTTP deliverer
func NewHTTPDeliverer(log *logger.Logger) *HTTPDeliverer {
	return &HTTPDeliverer{
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
		logger: log,
	}
}

// Deliver отправляет webhook
func (d *HTTPDeliverer) Deliver(ctx context.Context, sub *Subscription, payload *Payload) error {
	// Сериализуем payload
	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}

	// Создаем HTTP запрос
	req, err := http.NewRequestWithContext(ctx, "POST", sub.URL, bytes.NewBuffer(body))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	// Устанавливаем headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Webhook-Event", string(payload.Event))
	req.Header.Set("X-Webhook-Timestamp", payload.Timestamp.Format(time.RFC3339))

	// Добавляем HMAC signature если есть secret
	if sub.Secret != "" {
		signature := generateSignature(body, sub.Secret)
		req.Header.Set("X-Webhook-Signature", signature)
	}

	// Отправляем запрос
	resp, err := d.client.Do(req)
	if err != nil {
		d.logger.Errorw("Failed to deliver webhook",
			"subscription_id", sub.ID,
			"url", sub.URL,
			"error", err,
		)
		return fmt.Errorf("failed to deliver webhook: %w", err)
	}
	defer resp.Body.Close()

	// Проверяем статус код
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		d.logger.Warnw("Webhook delivery failed with non-2xx status",
			"subscription_id", sub.ID,
			"url", sub.URL,
			"status_code", resp.StatusCode,
		)
		return fmt.Errorf("webhook returned status %d", resp.StatusCode)
	}

	d.logger.Debugw("Webhook delivered successfully",
		"subscription_id", sub.ID,
		"url", sub.URL,
		"event", payload.Event,
	)

	return nil
}

// Manager управляет webhook подписками и доставкой
type Manager struct {
	deliverer     Deliverer
	subscriptions []*Subscription
	logger        *logger.Logger
	queue         chan *webhookJob
}

type webhookJob struct {
	subscription *Subscription
	payload      *Payload
}

// NewManager создает новый webhook manager
func NewManager(deliverer Deliverer, log *logger.Logger) *Manager {
	m := &Manager{
		deliverer:     deliverer,
		subscriptions: make([]*Subscription, 0),
		logger:        log,
		queue:         make(chan *webhookJob, 100),
	}

	// Запускаем воркеры для обработки webhook
	for i := 0; i < 5; i++ {
		go m.worker()
	}

	return m
}

// Subscribe добавляет подписку
func (m *Manager) Subscribe(sub *Subscription) {
	m.subscriptions = append(m.subscriptions, sub)
	m.logger.Infow("Webhook subscription added",
		"id", sub.ID,
		"url", sub.URL,
		"events", sub.Events,
	)
}

// Trigger отправляет webhook всем подписчикам
func (m *Manager) Trigger(event EventType, data map[string]interface{}) {
	payload := &Payload{
		Event:     event,
		Timestamp: time.Now(),
		Data:      data,
	}

	// Находим все активные подписки на это событие
	for _, sub := range m.subscriptions {
		if !sub.Active {
			continue
		}

		// Проверяем что подписка слушает это событие
		subscribed := false
		for _, e := range sub.Events {
			if e == event {
				subscribed = true
				break
			}
		}

		if subscribed {
			// Добавляем в очередь
			select {
			case m.queue <- &webhookJob{
				subscription: sub,
				payload:      payload,
			}:
			default:
				m.logger.Warnw("Webhook queue is full, dropping event",
					"subscription_id", sub.ID,
					"event", event,
				)
			}
		}
	}
}

// worker обрабатывает webhook из очереди
func (m *Manager) worker() {
	for job := range m.queue {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)

		// Пытаемся доставить с retry
		err := m.deliverWithRetry(ctx, job.subscription, job.payload, 3)
		if err != nil {
			m.logger.Errorw("Failed to deliver webhook after retries",
				"subscription_id", job.subscription.ID,
				"event", job.payload.Event,
				"error", err,
			)
		}

		cancel()
	}
}

// deliverWithRetry пытается доставить webhook с повторами
func (m *Manager) deliverWithRetry(ctx context.Context, sub *Subscription, payload *Payload, maxRetries int) error {
	var lastErr error

	for i := 0; i < maxRetries; i++ {
		if i > 0 {
			// Exponential backoff
			delay := time.Duration(i*i) * time.Second
			select {
			case <-time.After(delay):
			case <-ctx.Done():
				return ctx.Err()
			}
		}

		err := m.deliverer.Deliver(ctx, sub, payload)
		if err == nil {
			return nil
		}

		lastErr = err
		m.logger.Warnw("Webhook delivery attempt failed",
			"subscription_id", sub.ID,
			"attempt", i+1,
			"max_retries", maxRetries,
			"error", err,
		)
	}

	return lastErr
}

// generateSignature генерирует HMAC signature
func generateSignature(body []byte, secret string) string {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(body)
	return hex.EncodeToString(mac.Sum(nil))
}

// VerifySignature проверяет HMAC signature
func VerifySignature(body []byte, signature, secret string) bool {
	expected := generateSignature(body, secret)
	return hmac.Equal([]byte(signature), []byte(expected))
}

// Helper функции для создания webhook events

// TriggerPRCreated отправляет событие создания PR
func (m *Manager) TriggerPRCreated(pr *models.PullRequest) {
	payload := map[string]interface{}{
		"pr_id":      pr.ID,
		"title":      pr.Title,
		"author_id":  pr.AuthorID,
		"status":     pr.Status,
		"reviewers":  pr.Reviewers,
		"created_at": pr.CreatedAt,
	}

	// Add team_id if team is available
	if pr.Team != nil {
		payload["team_id"] = pr.Team.ID
	}

	m.Trigger(EventPRCreated, payload)
}

// TriggerPRMerged отправляет событие слияния PR
func (m *Manager) TriggerPRMerged(pr *models.PullRequest) {
	m.Trigger(EventPRMerged, map[string]interface{}{
		"pr_id":     pr.ID,
		"title":     pr.Title,
		"author_id": pr.AuthorID,
		"merged_at": pr.MergedAt,
	})
}

// TriggerReviewerAssigned отправляет событие назначения рецензента
func (m *Manager) TriggerReviewerAssigned(prID int64, reviewerID int64) {
	m.Trigger(EventReviewerAssigned, map[string]interface{}{
		"pr_id":       prID,
		"reviewer_id": reviewerID,
		"assigned_at": time.Now(),
	})
}
