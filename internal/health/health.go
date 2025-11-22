package health

import (
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"runtime"
	"sync"
	"time"

	"github.com/user/pr-reviewer/internal/logger"
)

// Status представляет статус компонента
type Status string

const (
	StatusHealthy   Status = "healthy"
	StatusDegraded  Status = "degraded"
	StatusUnhealthy Status = "unhealthy"
)

// CheckResult результат проверки здоровья
type CheckResult struct {
	Status    Status                 `json:"status"`
	Timestamp time.Time              `json:"timestamp"`
	Duration  string                 `json:"duration"`
	Details   map[string]interface{} `json:"details,omitempty"`
	Error     string                 `json:"error,omitempty"`
}

// HealthResponse общий ответ health check
type HealthResponse struct {
	Status    Status                 `json:"status"`
	Timestamp time.Time              `json:"timestamp"`
	Version   string                 `json:"version"`
	Uptime    string                 `json:"uptime"`
	Checks    map[string]CheckResult `json:"checks"`
}

// Checker интерфейс для проверки здоровья компонента
type Checker interface {
	Check(ctx context.Context) CheckResult
	Name() string
}

// Health управляет проверками здоровья
type Health struct {
	version   string
	startTime time.Time
	checkers  []Checker
	logger    *logger.Logger
	mu        sync.RWMutex
}

// New создает новый health checker
func New(version string, log *logger.Logger) *Health {
	return &Health{
		version:   version,
		startTime: time.Now(),
		checkers:  make([]Checker, 0),
		logger:    log,
	}
}

// RegisterChecker регистрирует новую проверку
func (h *Health) RegisterChecker(checker Checker) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.checkers = append(h.checkers, checker)
}

// Check выполняет все проверки здоровья
func (h *Health) Check(ctx context.Context) HealthResponse {
	checks := make(map[string]CheckResult)
	overallStatus := StatusHealthy

	// Выполняем все проверки параллельно
	var wg sync.WaitGroup
	var mu sync.Mutex

	h.mu.RLock()
	checkers := h.checkers
	h.mu.RUnlock()

	for _, checker := range checkers {
		wg.Add(1)
		go func(c Checker) {
			defer wg.Done()

			// Timeout для каждой проверки
			checkCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
			defer cancel()

			result := c.Check(checkCtx)

			mu.Lock()
			checks[c.Name()] = result
			if result.Status == StatusUnhealthy {
				overallStatus = StatusUnhealthy
			} else if result.Status == StatusDegraded && overallStatus != StatusUnhealthy {
				overallStatus = StatusDegraded
			}
			mu.Unlock()
		}(checker)
	}

	wg.Wait()

	return HealthResponse{
		Status:    overallStatus,
		Timestamp: time.Now(),
		Version:   h.version,
		Uptime:    time.Since(h.startTime).String(),
		Checks:    checks,
	}
}

// Handler HTTP handler для health check
func (h *Health) Handler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		response := h.Check(ctx)

		w.Header().Set("Content-Type", "application/json")

		// Устанавливаем HTTP статус в зависимости от результата
		switch response.Status {
		case StatusHealthy:
			w.WriteHeader(http.StatusOK)
		case StatusDegraded:
			w.WriteHeader(http.StatusOK) // 200 но с предупреждением
		case StatusUnhealthy:
			w.WriteHeader(http.StatusServiceUnavailable)
		}

		if err := json.NewEncoder(w).Encode(response); err != nil {
			h.logger.WithError(err).Error("Failed to encode health response")
		}
	}
}

// ReadinessHandler HTTP handler для Kubernetes readiness probe
func (h *Health) ReadinessHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		response := h.Check(ctx)

		// Readiness probe должен возвращать 200 только если все healthy
		if response.Status == StatusHealthy {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("ready"))
		} else {
			w.WriteHeader(http.StatusServiceUnavailable)
			_, _ = w.Write([]byte("not ready"))
		}
	}
}

// LivenessHandler HTTP handler для Kubernetes liveness probe
func (h *Health) LivenessHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Liveness probe проверяет только что приложение живо
		// Не выполняем проверки зависимостей
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("alive"))
	}
}

// DatabaseChecker проверяет подключение к БД
type DatabaseChecker struct {
	db *sql.DB
}

func NewDatabaseChecker(db *sql.DB) *DatabaseChecker {
	return &DatabaseChecker{db: db}
}

func (c *DatabaseChecker) Name() string {
	return "database"
}

func (c *DatabaseChecker) Check(ctx context.Context) CheckResult {
	start := time.Now()
	result := CheckResult{
		Timestamp: start,
		Details:   make(map[string]interface{}),
	}

	// Проверяем ping
	if err := c.db.PingContext(ctx); err != nil {
		result.Status = StatusUnhealthy
		result.Error = err.Error()
		result.Duration = time.Since(start).String()
		return result
	}

	// Получаем статистику пула соединений
	stats := c.db.Stats()
	result.Details["open_connections"] = stats.OpenConnections
	result.Details["in_use"] = stats.InUse
	result.Details["idle"] = stats.Idle
	result.Details["max_open_connections"] = stats.MaxOpenConnections

	// Проверяем что есть доступные соединения
	if stats.OpenConnections >= stats.MaxOpenConnections {
		result.Status = StatusDegraded
		result.Error = "connection pool exhausted"
	} else {
		result.Status = StatusHealthy
	}

	result.Duration = time.Since(start).String()
	return result
}

// SystemChecker проверяет системные ресурсы
type SystemChecker struct{}

func NewSystemChecker() *SystemChecker {
	return &SystemChecker{}
}

func (c *SystemChecker) Name() string {
	return "system"
}

func (c *SystemChecker) Check(ctx context.Context) CheckResult {
	start := time.Now()
	result := CheckResult{
		Status:    StatusHealthy,
		Timestamp: start,
		Details:   make(map[string]interface{}),
	}

	// Memory stats
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	result.Details["goroutines"] = runtime.NumGoroutine()
	result.Details["memory_alloc_mb"] = m.Alloc / 1024 / 1024
	result.Details["memory_sys_mb"] = m.Sys / 1024 / 1024
	result.Details["gc_runs"] = m.NumGC

	// Проверяем количество горутин (простая эвристика)
	if runtime.NumGoroutine() > 10000 {
		result.Status = StatusDegraded
		result.Error = "too many goroutines"
	}

	// Disk space (platform-specific)
	if available, total, err := getDiskStats(); err == nil {
		usedPercent := float64(total-available) / float64(total) * 100

		result.Details["disk_available_gb"] = available / 1024 / 1024 / 1024
		result.Details["disk_total_gb"] = total / 1024 / 1024 / 1024
		result.Details["disk_used_percent"] = int(usedPercent)

		// Предупреждение если диск заполнен более чем на 80%
		if usedPercent > 80 {
			result.Status = StatusDegraded
			result.Error = "disk space low"
		}
	}

	result.Duration = time.Since(start).String()
	return result
}
