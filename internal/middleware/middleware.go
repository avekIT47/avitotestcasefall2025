package middleware

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/user/pr-reviewer/internal/logger"
	"github.com/user/pr-reviewer/internal/metrics"
	"golang.org/x/time/rate"
)

const (
	// Request ID header
	RequestIDHeader = "X-Request-ID"

	// Context keys
	RequestIDKey = "request_id"
	StartTimeKey = "start_time"

	// Limits
	MaxRequestBodySize = 1 << 20 // 1MB
)

// Middleware содержит все middleware для HTTP handlers
type Middleware struct {
	logger  *logger.Logger
	metrics *metrics.Metrics
	limiter *RateLimiter
}

// New создает новый экземпляр middleware
func New(log *logger.Logger, met *metrics.Metrics) *Middleware {
	return &Middleware{
		logger:  log,
		metrics: met,
		limiter: NewRateLimiter(100, 200), // 100 RPS burst 200
	}
}

// RequestID добавляет уникальный ID к каждому запросу
func (m *Middleware) RequestID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestID := r.Header.Get(RequestIDHeader)
		if requestID == "" {
			requestID = uuid.New().String()
		}

		// Добавляем request ID в контекст
		ctx := context.WithValue(r.Context(), RequestIDKey, requestID)

		// Добавляем request ID в ответ
		w.Header().Set(RequestIDHeader, requestID)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// Logging логирует все HTTP запросы
func (m *Middleware) Logging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		startTime := time.Now()
		ctx := context.WithValue(r.Context(), StartTimeKey, startTime)

		// Оборачиваем ResponseWriter для захвата статус кода
	wrapped := &responseWriter{
		ResponseWriter: w,
		statusCode:     http.StatusOK,
	}

	requestID, ok := r.Context().Value(RequestIDKey).(string)
	if !ok {
		requestID = "unknown"
	}
	reqLogger := m.logger.WithRequestID(requestID)

		reqLogger.Infow("HTTP request started",
			"method", r.Method,
			"path", r.URL.Path,
			"remote_addr", r.RemoteAddr,
			"user_agent", r.UserAgent(),
		)

		next.ServeHTTP(wrapped, r.WithContext(ctx))

		duration := time.Since(startTime)
		reqLogger.Infow("HTTP request completed",
			"method", r.Method,
			"path", r.URL.Path,
			"status", wrapped.statusCode,
			"duration_ms", duration.Milliseconds(),
		)
	})
}

// Metrics записывает метрики для каждого запроса
func (m *Middleware) Metrics(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		m.metrics.IncrementInFlightRequests()
		defer m.metrics.DecrementInFlightRequests()

		startTime := time.Now()

		wrapped := &responseWriter{
			ResponseWriter: w,
			statusCode:     http.StatusOK,
		}

		next.ServeHTTP(wrapped, r)

		duration := time.Since(startTime)
		m.metrics.RecordHTTPRequest(
			r.Method,
			sanitizePath(r.URL.Path),
			wrapped.statusCode,
			duration,
		)
	})
}

// RateLimit ограничивает количество запросов
func (m *Middleware) RateLimit(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Используем IP адрес как ключ
		ip := getIP(r)

		if !m.limiter.Allow(ip) {
			requestID, ok := r.Context().Value(RequestIDKey).(string)
			if !ok {
				requestID = "unknown"
			}
			m.logger.WithRequestID(requestID).Warnw("Rate limit exceeded",
				"ip", ip,
				"path", r.URL.Path,
			)

			w.Header().Set("Retry-After", "60")
			http.Error(w, "Too Many Requests", http.StatusTooManyRequests)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// SecurityHeaders добавляет security headers к ответу
func (m *Middleware) SecurityHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Защита от XSS
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "DENY")
		w.Header().Set("X-XSS-Protection", "1; mode=block")

		// Content Security Policy
		w.Header().Set("Content-Security-Policy", "default-src 'self'")

		// HTTPS enforcement (в production за load balancer)
		if r.Header.Get("X-Forwarded-Proto") == "https" {
			w.Header().Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
		}

		// Referrer policy
		w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")

		// Permissions policy
		w.Header().Set("Permissions-Policy", "geolocation=(), microphone=(), camera=()")

		next.ServeHTTP(w, r)
	})
}

// RequestValidation валидирует входящие запросы
func (m *Middleware) RequestValidation(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Проверка размера тела запроса
		r.Body = http.MaxBytesReader(w, r.Body, MaxRequestBodySize)

	// Проверка Content-Type для POST/PUT/PATCH
	if r.Method == http.MethodPost || r.Method == http.MethodPut || r.Method == http.MethodPatch {
		contentType := r.Header.Get("Content-Type")
		if !strings.HasPrefix(contentType, "application/json") {
			requestID, ok := r.Context().Value(RequestIDKey).(string)
			if !ok {
				requestID = "unknown"
			}
			m.logger.WithRequestID(requestID).Warnw("Invalid content type",
					"content_type", contentType,
					"path", r.URL.Path,
				)
				http.Error(w, "Content-Type must be application/json", http.StatusUnsupportedMediaType)
				return
			}
		}

		next.ServeHTTP(w, r)
	})
}

// Recovery восстанавливается после panic
func (m *Middleware) Recovery(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				requestID := r.Context().Value(RequestIDKey)
				if requestID == nil {
					requestID = "unknown"
				}

				m.logger.WithRequestID(requestID.(string)).Errorw("Panic recovered",
					"error", err,
					"path", r.URL.Path,
				)

				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusInternalServerError)
				fmt.Fprintf(w, `{"error":"Internal server error","request_id":"%s"}`, requestID)
			}
		}()

		next.ServeHTTP(w, r)
	})
}

// Chain объединяет несколько middleware
func Chain(middlewares ...func(http.Handler) http.Handler) func(http.Handler) http.Handler {
	return func(final http.Handler) http.Handler {
		for i := len(middlewares) - 1; i >= 0; i-- {
			final = middlewares[i](final)
		}
		return final
	}
}

// responseWriter обертка для захвата статус кода
type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

// RateLimiter реализует rate limiting per IP
type RateLimiter struct {
	visitors map[string]*rate.Limiter
	mu       sync.RWMutex
	rps      int
	burst    int
}

// NewRateLimiter создает новый rate limiter
func NewRateLimiter(rps, burst int) *RateLimiter {
	rl := &RateLimiter{
		visitors: make(map[string]*rate.Limiter),
		rps:      rps,
		burst:    burst,
	}

	// Периодическая очистка старых visitor'ов
	go rl.cleanup()

	return rl
}

// Allow проверяет, разрешен ли запрос для данного IP
func (rl *RateLimiter) Allow(ip string) bool {
	rl.mu.Lock()
	limiter, exists := rl.visitors[ip]
	if !exists {
		limiter = rate.NewLimiter(rate.Limit(rl.rps), rl.burst)
		rl.visitors[ip] = limiter
	}
	rl.mu.Unlock()

	return limiter.Allow()
}

// cleanup периодически удаляет неактивных visitor'ов
func (rl *RateLimiter) cleanup() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		rl.mu.Lock()
		// В production здесь нужна более сложная логика с отслеживанием last access time
		if len(rl.visitors) > 10000 {
			rl.visitors = make(map[string]*rate.Limiter)
		}
		rl.mu.Unlock()
	}
}

// getIP извлекает IP адрес из запроса
func getIP(r *http.Request) string {
	// Проверяем X-Forwarded-For header (если за прокси/балансировщиком)
	forwarded := r.Header.Get("X-Forwarded-For")
	if forwarded != "" {
		// Берем первый IP из списка
		ips := strings.Split(forwarded, ",")
		return strings.TrimSpace(ips[0])
	}

	// Проверяем X-Real-IP header
	realIP := r.Header.Get("X-Real-IP")
	if realIP != "" {
		return realIP
	}

	// Используем RemoteAddr
	ip := r.RemoteAddr
	if idx := strings.LastIndex(ip, ":"); idx != -1 {
		ip = ip[:idx]
	}
	return ip
}

// sanitizePath удаляет параметры из пути для метрик
func sanitizePath(path string) string {
	// Заменяем ID на placeholder для агрегации метрик
	// Например: /teams/123 -> /teams/:id
	parts := strings.Split(path, "/")
	for i, part := range parts {
		// Простая эвристика: если часть пути - число, заменяем на :id
		if len(part) > 0 && isNumeric(part) {
			parts[i] = ":id"
		}
	}
	return strings.Join(parts, "/")
}

// isNumeric проверяет, является ли строка числом
func isNumeric(s string) bool {
	for _, c := range s {
		if c < '0' || c > '9' {
			return false
		}
	}
	return len(s) > 0
}
