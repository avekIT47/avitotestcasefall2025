package circuitbreaker

import (
	"context"
	"fmt"
	"time"

	"github.com/sony/gobreaker"
	"github.com/user/pr-reviewer/internal/logger"
)

// CircuitBreaker обертка над gobreaker с логированием
type CircuitBreaker struct {
	cb     *gobreaker.CircuitBreaker
	logger *logger.Logger
	name   string
}

// Config конфигурация circuit breaker
type Config struct {
	Name          string
	MaxRequests   uint32        // Максимум запросов в half-open состоянии
	Interval      time.Duration // Интервал для сброса счетчиков
	Timeout       time.Duration // Время в open состоянии
	ReadyToTrip   func(counts gobreaker.Counts) bool
	OnStateChange func(name string, from gobreaker.State, to gobreaker.State)
}

// New создает новый circuit breaker
func New(cfg Config, log *logger.Logger) *CircuitBreaker {
	if cfg.MaxRequests == 0 {
		cfg.MaxRequests = 1
	}
	if cfg.Interval == 0 {
		cfg.Interval = 60 * time.Second
	}
	if cfg.Timeout == 0 {
		cfg.Timeout = 30 * time.Second
	}
	if cfg.ReadyToTrip == nil {
		// По умолчанию: открываем если > 5 ошибок или > 50% error rate
		cfg.ReadyToTrip = func(counts gobreaker.Counts) bool {
			failureRatio := float64(counts.TotalFailures) / float64(counts.Requests)
			return counts.Requests >= 3 && (counts.ConsecutiveFailures > 5 || failureRatio >= 0.5)
		}
	}

	cb := &CircuitBreaker{
		logger: log,
		name:   cfg.Name,
	}

	settings := gobreaker.Settings{
		Name:        cfg.Name,
		MaxRequests: cfg.MaxRequests,
		Interval:    cfg.Interval,
		Timeout:     cfg.Timeout,
		ReadyToTrip: cfg.ReadyToTrip,
		OnStateChange: func(name string, from gobreaker.State, to gobreaker.State) {
			cb.logger.Warnw("Circuit breaker state changed",
				"name", name,
				"from", from.String(),
				"to", to.String(),
			)
			if cfg.OnStateChange != nil {
				cfg.OnStateChange(name, from, to)
			}
		},
	}

	cb.cb = gobreaker.NewCircuitBreaker(settings)
	return cb
}

// Execute выполняет функцию через circuit breaker
func (cb *CircuitBreaker) Execute(fn func() (interface{}, error)) (interface{}, error) {
	result, err := cb.cb.Execute(fn)
	if err != nil {
		if err == gobreaker.ErrOpenState {
			cb.logger.Warnw("Circuit breaker is open",
				"name", cb.name,
			)
		} else if err == gobreaker.ErrTooManyRequests {
			cb.logger.Warnw("Circuit breaker: too many requests",
				"name", cb.name,
			)
		}
	}
	return result, err
}

// ExecuteContext выполняет функцию с контекстом
func (cb *CircuitBreaker) ExecuteContext(ctx context.Context, fn func() (interface{}, error)) (interface{}, error) {
	// Проверяем контекст перед выполнением
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	// Выполняем с таймаутом из контекста
	done := make(chan struct {
		result interface{}
		err    error
	}, 1)

	go func() {
		result, err := cb.Execute(fn)
		done <- struct {
			result interface{}
			err    error
		}{result, err}
	}()

	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case res := <-done:
		return res.result, res.err
	}
}

// State возвращает текущее состояние
func (cb *CircuitBreaker) State() gobreaker.State {
	return cb.cb.State()
}

// Counts возвращает статистику
func (cb *CircuitBreaker) Counts() gobreaker.Counts {
	return cb.cb.Counts()
}

// NewDefaultConfig создает конфигурацию по умолчанию
func NewDefaultConfig(name string) Config {
	return Config{
		Name:        name,
		MaxRequests: 1,
		Interval:    60 * time.Second,
		Timeout:     30 * time.Second,
	}
}

// Example: Использование для external API
func ExampleDatabaseCall(cb *CircuitBreaker) error {
	_, err := cb.Execute(func() (interface{}, error) {
		// Ваш код работы с БД
		// Если БД недоступна, вернется ошибка
		// Circuit breaker откроется после нескольких ошибок
		return nil, nil
	})
	return err
}

// Example: Использование для HTTP запросов
func ExampleHTTPCall(cb *CircuitBreaker) ([]byte, error) {
	result, err := cb.Execute(func() (interface{}, error) {
		// HTTP запрос к внешнему API
		// Example: return http.Get(url)
		return nil, nil
	})
	if err != nil {
		return nil, err
	}
	// return result.(*http.Response), nil
	return result.([]byte), nil
}

// WithFallback выполняет функцию с fallback при ошибке
func (cb *CircuitBreaker) WithFallback(fn func() (interface{}, error), fallback func() (interface{}, error)) (interface{}, error) {
	result, err := cb.Execute(fn)
	if err != nil {
		cb.logger.Warnw("Primary function failed, using fallback",
			"name", cb.name,
			"error", err,
		)
		return fallback()
	}
	return result, nil
}

// Manager управляет несколькими circuit breakers
type Manager struct {
	breakers map[string]*CircuitBreaker
	logger   *logger.Logger
}

// NewManager создает менеджер circuit breakers
func NewManager(log *logger.Logger) *Manager {
	return &Manager{
		breakers: make(map[string]*CircuitBreaker),
		logger:   log,
	}
}

// Register регистрирует новый circuit breaker
func (m *Manager) Register(name string, cfg Config) *CircuitBreaker {
	cb := New(cfg, m.logger)
	m.breakers[name] = cb
	return cb
}

// Get возвращает circuit breaker по имени
func (m *Manager) Get(name string) (*CircuitBreaker, error) {
	cb, exists := m.breakers[name]
	if !exists {
		return nil, fmt.Errorf("circuit breaker not found: %s", name)
	}
	return cb, nil
}

// GetStates возвращает состояния всех breakers
func (m *Manager) GetStates() map[string]string {
	states := make(map[string]string)
	for name, cb := range m.breakers {
		states[name] = cb.State().String()
	}
	return states
}
