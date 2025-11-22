package featureflags

import (
	"context"
	"encoding/json"
	"sync"
	"time"

	"github.com/user/pr-reviewer/internal/cache"
	"github.com/user/pr-reviewer/internal/logger"
)

// Flag представляет feature flag
type Flag struct {
	Key         string                 `json:"key"`
	Enabled     bool                   `json:"enabled"`
	Description string                 `json:"description"`
	Rollout     *Rollout               `json:"rollout,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
	UpdatedAt   time.Time              `json:"updated_at"`
}

// Rollout конфигурация постепенного раската
type Rollout struct {
	Percentage int      `json:"percentage"` // 0-100
	UserIDs    []int64  `json:"user_ids,omitempty"`
	TeamIDs    []int64  `json:"team_ids,omitempty"`
	Whitelist  []string `json:"whitelist,omitempty"` // email patterns
}

// Context контекст для проверки флага
type Context struct {
	UserID int64
	TeamID int64
	Email  string
}

// Manager управляет feature flags
type Manager struct {
	flags  map[string]*Flag
	mu     sync.RWMutex
	cache  cache.Cache
	logger *logger.Logger
}

// NewManager создает новый feature flags manager
func NewManager(cacheClient cache.Cache, log *logger.Logger) *Manager {
	m := &Manager{
		flags:  make(map[string]*Flag),
		cache:  cacheClient,
		logger: log,
	}

	// Инициализируем дефолтные флаги
	m.initDefaultFlags()

	return m
}

// initDefaultFlags инициализирует флаги по умолчанию
func (m *Manager) initDefaultFlags() {
	defaultFlags := []*Flag{
		{
			Key:         "jwt_auth",
			Enabled:     false,
			Description: "Enable JWT authentication",
			UpdatedAt:   time.Now(),
		},
		{
			Key:         "redis_cache",
			Enabled:     true,
			Description: "Enable Redis caching",
			UpdatedAt:   time.Now(),
		},
		{
			Key:         "webhooks",
			Enabled:     false,
			Description: "Enable webhook notifications",
			UpdatedAt:   time.Now(),
		},
		{
			Key:         "audit_log",
			Enabled:     true,
			Description: "Enable audit logging",
			UpdatedAt:   time.Now(),
		},
		{
			Key:         "advanced_metrics",
			Enabled:     true,
			Description: "Enable advanced Prometheus metrics",
			UpdatedAt:   time.Now(),
		},
		{
			Key:         "rate_limiting",
			Enabled:     true,
			Description: "Enable rate limiting",
			UpdatedAt:   time.Now(),
		},
		{
			Key:         "circuit_breaker",
			Enabled:     false,
			Description: "Enable circuit breaker for external calls",
			UpdatedAt:   time.Now(),
		},
		{
			Key:         "distributed_tracing",
			Enabled:     false,
			Description: "Enable distributed tracing with Jaeger",
			UpdatedAt:   time.Now(),
		},
	}

	for _, flag := range defaultFlags {
		m.flags[flag.Key] = flag
	}

	m.logger.Infow("Default feature flags initialized", "count", len(defaultFlags))
}

// IsEnabled проверяет включен ли флаг
func (m *Manager) IsEnabled(key string) bool {
	return m.IsEnabledWithContext(key, nil)
}

// IsEnabledWithContext проверяет флаг с контекстом для rollout
func (m *Manager) IsEnabledWithContext(key string, ctx *Context) bool {
	m.mu.RLock()
	flag, exists := m.flags[key]
	m.mu.RUnlock()

	if !exists {
		m.logger.Warnw("Feature flag not found", "key", key)
		return false
	}

	// Если флаг полностью выключен
	if !flag.Enabled {
		return false
	}

	// Если нет rollout конфигурации - флаг включен для всех
	if flag.Rollout == nil || ctx == nil {
		return true
	}

	// Проверяем whitelist по user ID
	if ctx.UserID > 0 {
		for _, id := range flag.Rollout.UserIDs {
			if id == ctx.UserID {
				return true
			}
		}
	}

	// Проверяем whitelist по team ID
	if ctx.TeamID > 0 {
		for _, id := range flag.Rollout.TeamIDs {
			if id == ctx.TeamID {
				return true
			}
		}
	}

	// Проверяем percentage rollout
	if flag.Rollout.Percentage > 0 && ctx.UserID > 0 {
		// Детерминированный hash для consistency
		hash := int(ctx.UserID % 100)
		if hash < flag.Rollout.Percentage {
			return true
		}
	}

	return false
}

// SetFlag устанавливает значение флага
func (m *Manager) SetFlag(flag *Flag) {
	flag.UpdatedAt = time.Now()

	m.mu.Lock()
	m.flags[flag.Key] = flag
	m.mu.Unlock()

	// Инвалидируем кеш
	if m.cache != nil {
		_ = m.cache.Delete(context.Background(), "feature_flags")
	}

	m.logger.Infow("Feature flag updated",
		"key", flag.Key,
		"enabled", flag.Enabled,
	)
}

// GetFlag возвращает флаг
func (m *Manager) GetFlag(key string) (*Flag, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	flag, exists := m.flags[key]
	return flag, exists
}

// GetAllFlags возвращает все флаги
func (m *Manager) GetAllFlags() map[string]*Flag {
	m.mu.RLock()
	defer m.mu.RUnlock()

	// Копируем map чтобы избежать race conditions
	result := make(map[string]*Flag, len(m.flags))
	for k, v := range m.flags {
		result[k] = v
	}
	return result
}

// EnableFlag включает флаг
func (m *Manager) EnableFlag(key string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	flag, exists := m.flags[key]
	if !exists {
		m.logger.Warnw("Feature flag not found", "key", key)
		return nil
	}

	flag.Enabled = true
	flag.UpdatedAt = time.Now()

	m.logger.Infow("Feature flag enabled", "key", key)
	return nil
}

// DisableFlag выключает флаг
func (m *Manager) DisableFlag(key string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	flag, exists := m.flags[key]
	if !exists {
		m.logger.Warnw("Feature flag not found", "key", key)
		return nil
	}

	flag.Enabled = false
	flag.UpdatedAt = time.Now()

	m.logger.Infow("Feature flag disabled", "key", key)
	return nil
}

// SetRollout устанавливает rollout конфигурацию
func (m *Manager) SetRollout(key string, rollout *Rollout) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	flag, exists := m.flags[key]
	if !exists {
		m.logger.Warnw("Feature flag not found", "key", key)
		return nil
	}

	flag.Rollout = rollout
	flag.UpdatedAt = time.Now()

	m.logger.Infow("Feature flag rollout updated",
		"key", key,
		"percentage", rollout.Percentage,
	)
	return nil
}

// LoadFromCache загружает флаги из кеша
func (m *Manager) LoadFromCache(ctx context.Context) error {
	if m.cache == nil {
		return nil
	}

	var flags map[string]*Flag
	err := m.cache.Get(ctx, "feature_flags", &flags)
	if err != nil {
		return err
	}

	m.mu.Lock()
	m.flags = flags
	m.mu.Unlock()

	m.logger.Info("Feature flags loaded from cache")
	return nil
}

// SaveToCache сохраняет флаги в кеш
func (m *Manager) SaveToCache(ctx context.Context) error {
	if m.cache == nil {
		return nil
	}

	m.mu.RLock()
	flags := m.flags
	m.mu.RUnlock()

	return m.cache.Set(ctx, "feature_flags", flags, 1*time.Hour)
}

// MarshalJSON сериализует флаги в JSON
func (m *Manager) MarshalJSON() ([]byte, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return json.Marshal(m.flags)
}

// Helper: Middleware для проверки feature flag
func RequireFeatureFlag(m *Manager, flagKey string) func(next func()) func() {
	return func(next func()) func() {
		return func() {
			if m.IsEnabled(flagKey) {
				next()
			}
		}
	}
}

// Примеры использования

// Example 1: Простая проверка
func ExampleSimpleCheck(m *Manager) {
	if m.IsEnabled("jwt_auth") {
		// JWT authentication включена
	}
}

// Example 2: Проверка с контекстом
func ExampleContextCheck(m *Manager, userID, teamID int64) {
	ctx := &Context{
		UserID: userID,
		TeamID: teamID,
	}

	if m.IsEnabledWithContext("new_feature", ctx) {
		// Новая фича включена для этого пользователя
	}
}

// Example 3: Gradual rollout
func ExampleGradualRollout(m *Manager) {
	// Включаем фичу для 25% пользователей
	m.SetFlag(&Flag{
		Key:         "new_dashboard",
		Enabled:     true,
		Description: "New dashboard UI",
		Rollout: &Rollout{
			Percentage: 25, // 25% пользователей
		},
	})
}

// Example 4: Whitelist rollout
func ExampleWhitelistRollout(m *Manager) {
	// Включаем только для определенных пользователей
	m.SetFlag(&Flag{
		Key:         "beta_feature",
		Enabled:     true,
		Description: "Beta feature",
		Rollout: &Rollout{
			UserIDs: []int64{1, 2, 3, 10, 15},
			TeamIDs: []int64{5, 7},
		},
	})
}
