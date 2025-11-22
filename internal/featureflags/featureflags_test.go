package featureflags

import (
	"context"
	"testing"

	"github.com/user/pr-reviewer/internal/cache"
	"github.com/user/pr-reviewer/internal/logger"
)

func TestFeatureFlags_IsEnabled(t *testing.T) {
	log, _ := logger.New("error", "test")
	cacheClient := cache.NewNoOpCache()
	manager := NewManager(cacheClient, log)

	// Проверяем дефолтный флаг
	if !manager.IsEnabled("redis_cache") {
		t.Error("Expected redis_cache to be enabled by default")
	}

	if manager.IsEnabled("jwt_auth") {
		t.Error("Expected jwt_auth to be disabled by default")
	}
}

func TestFeatureFlags_SetFlag(t *testing.T) {
	log, _ := logger.New("error", "test")
	cacheClient := cache.NewNoOpCache()
	manager := NewManager(cacheClient, log)

	// Создаем новый флаг
	flag := &Flag{
		Key:         "test_feature",
		Enabled:     true,
		Description: "Test feature",
	}

	manager.SetFlag(flag)

	// Проверяем что флаг сохранился
	if !manager.IsEnabled("test_feature") {
		t.Error("Expected test_feature to be enabled")
	}

	// Получаем флаг
	retrieved, exists := manager.GetFlag("test_feature")
	if !exists {
		t.Error("Expected flag to exist")
	}

	if retrieved.Key != "test_feature" {
		t.Errorf("Expected key 'test_feature', got '%s'", retrieved.Key)
	}
}

func TestFeatureFlags_EnableDisable(t *testing.T) {
	log, _ := logger.New("error", "test")
	cacheClient := cache.NewNoOpCache()
	manager := NewManager(cacheClient, log)

	// Выключаем дефолтный включенный флаг
	err := manager.DisableFlag("redis_cache")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if manager.IsEnabled("redis_cache") {
		t.Error("Expected redis_cache to be disabled")
	}

	// Включаем обратно
	err = manager.EnableFlag("redis_cache")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if !manager.IsEnabled("redis_cache") {
		t.Error("Expected redis_cache to be enabled")
	}
}

func TestFeatureFlags_RolloutPercentage(t *testing.T) {
	log, _ := logger.New("error", "test")
	cacheClient := cache.NewNoOpCache()
	manager := NewManager(cacheClient, log)

	// Создаем флаг с 50% rollout
	flag := &Flag{
		Key:         "new_feature",
		Enabled:     true,
		Description: "New feature with 50% rollout",
		Rollout: &Rollout{
			Percentage: 50,
		},
	}

	manager.SetFlag(flag)

	// Проверяем для разных пользователей
	enabledCount := 0
	for userID := int64(1); userID <= 100; userID++ {
		ctx := &Context{UserID: userID}
		if manager.IsEnabledWithContext("new_feature", ctx) {
			enabledCount++
		}
	}

	// Должно быть примерно 50% (допускаем погрешность)
	if enabledCount < 40 || enabledCount > 60 {
		t.Errorf("Expected ~50%% enabled, got %d%%", enabledCount)
	}
}

func TestFeatureFlags_RolloutWhitelist(t *testing.T) {
	log, _ := logger.New("error", "test")
	cacheClient := cache.NewNoOpCache()
	manager := NewManager(cacheClient, log)

	// Создаем флаг только для определенных пользователей
	flag := &Flag{
		Key:         "beta_feature",
		Enabled:     true,
		Description: "Beta feature",
		Rollout: &Rollout{
			UserIDs: []int64{1, 2, 3},
			TeamIDs: []int64{5},
		},
	}

	manager.SetFlag(flag)

	// Пользователь в whitelist
	ctx1 := &Context{UserID: 1}
	if !manager.IsEnabledWithContext("beta_feature", ctx1) {
		t.Error("Expected user 1 to have access")
	}

	// Пользователь не в whitelist
	ctx2 := &Context{UserID: 99}
	if manager.IsEnabledWithContext("beta_feature", ctx2) {
		t.Error("Expected user 99 to NOT have access")
	}

	// Team в whitelist
	ctx3 := &Context{TeamID: 5}
	if !manager.IsEnabledWithContext("beta_feature", ctx3) {
		t.Error("Expected team 5 to have access")
	}

	// Team не в whitelist
	ctx4 := &Context{TeamID: 99}
	if manager.IsEnabledWithContext("beta_feature", ctx4) {
		t.Error("Expected team 99 to NOT have access")
	}
}

func TestFeatureFlags_GetAllFlags(t *testing.T) {
	log, _ := logger.New("error", "test")
	cacheClient := cache.NewNoOpCache()
	manager := NewManager(cacheClient, log)

	// Получаем все флаги
	flags := manager.GetAllFlags()

	// Должны быть дефолтные флаги
	if len(flags) < 5 {
		t.Errorf("Expected at least 5 default flags, got %d", len(flags))
	}

	// Проверяем наличие известных флагов
	if _, exists := flags["redis_cache"]; !exists {
		t.Error("Expected redis_cache flag to exist")
	}

	if _, exists := flags["jwt_auth"]; !exists {
		t.Error("Expected jwt_auth flag to exist")
	}
}

func TestFeatureFlags_CacheIntegration(t *testing.T) {
	log, _ := logger.New("error", "test")
	cacheClient := cache.NewNoOpCache()
	manager := NewManager(cacheClient, log)

	ctx := context.Background()

	// Сохраняем в кеш
	err := manager.SaveToCache(ctx)
	if err != nil {
		t.Errorf("Unexpected error saving to cache: %v", err)
	}

	// Загружаем из кеша (с NoOpCache это ничего не делает, но не должно быть ошибок)
	err = manager.LoadFromCache(ctx)
	// NoOpCache всегда возвращает cache miss, это нормально
	if err == nil {
		t.Log("Cache load attempted (NoOpCache always returns miss)")
	}
}

func TestFeatureFlags_ConsistentHashing(t *testing.T) {
	log, _ := logger.New("error", "test")
	cacheClient := cache.NewNoOpCache()
	manager := NewManager(cacheClient, log)

	// Создаем флаг с rollout
	flag := &Flag{
		Key:         "consistent_feature",
		Enabled:     true,
		Description: "Feature with consistent rollout",
		Rollout: &Rollout{
			Percentage: 30,
		},
	}

	manager.SetFlag(flag)

	// Проверяем что результат консистентный для одного пользователя
	ctx := &Context{UserID: 42}

	firstResult := manager.IsEnabledWithContext("consistent_feature", ctx)

	// Проверяем 10 раз
	for i := 0; i < 10; i++ {
		result := manager.IsEnabledWithContext("consistent_feature", ctx)
		if result != firstResult {
			t.Error("Feature flag result is not consistent for the same user")
		}
	}
}
