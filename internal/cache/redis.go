package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/user/pr-reviewer/internal/logger"
)

// Cache интерфейс для кеширования
type Cache interface {
	Get(ctx context.Context, key string, dest interface{}) error
	Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error
	Delete(ctx context.Context, key string) error
	DeletePattern(ctx context.Context, pattern string) error
	Exists(ctx context.Context, key string) (bool, error)
	Close() error
}

// RedisCache реализация кеша на Redis
type RedisCache struct {
	client *redis.Client
	logger *logger.Logger
	prefix string
}

// NewRedisCache создает новый Redis кеш
func NewRedisCache(addr, password string, db int, prefix string, log *logger.Logger) (*RedisCache, error) {
	client := redis.NewClient(&redis.Options{
		Addr:         addr,
		Password:     password,
		DB:           db,
		DialTimeout:  5 * time.Second,
		ReadTimeout:  3 * time.Second,
		WriteTimeout: 3 * time.Second,
		PoolSize:     10,
		MinIdleConns: 5,
	})

	// Проверяем подключение
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("redis connection failed: %w", err)
	}

	log.Info("Connected to Redis cache")

	return &RedisCache{
		client: client,
		logger: log,
		prefix: prefix,
	}, nil
}

// Get получает значение из кеша
func (c *RedisCache) Get(ctx context.Context, key string, dest interface{}) error {
	fullKey := c.prefix + key

	val, err := c.client.Get(ctx, fullKey).Result()
	if err == redis.Nil {
		return fmt.Errorf("cache miss: %s", key)
	}
	if err != nil {
		c.logger.WithError(err).Warnw("Cache get error", "key", key)
		return err
	}

	if err := json.Unmarshal([]byte(val), dest); err != nil {
		c.logger.WithError(err).Errorw("Cache unmarshal error", "key", key)
		return err
	}

	c.logger.Debugw("Cache hit", "key", key)
	return nil
}

// Set сохраняет значение в кеш
func (c *RedisCache) Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	fullKey := c.prefix + key

	data, err := json.Marshal(value)
	if err != nil {
		c.logger.WithError(err).Errorw("Cache marshal error", "key", key)
		return err
	}

	if err := c.client.Set(ctx, fullKey, data, ttl).Err(); err != nil {
		c.logger.WithError(err).Warnw("Cache set error", "key", key)
		return err
	}

	c.logger.Debugw("Cache set", "key", key, "ttl", ttl)
	return nil
}

// Delete удаляет значение из кеша
func (c *RedisCache) Delete(ctx context.Context, key string) error {
	fullKey := c.prefix + key

	if err := c.client.Del(ctx, fullKey).Err(); err != nil {
		c.logger.WithError(err).Warnw("Cache delete error", "key", key)
		return err
	}

	c.logger.Debugw("Cache delete", "key", key)
	return nil
}

// DeletePattern удаляет все ключи по паттерну
func (c *RedisCache) DeletePattern(ctx context.Context, pattern string) error {
	fullPattern := c.prefix + pattern

	iter := c.client.Scan(ctx, 0, fullPattern, 0).Iterator()
	for iter.Next(ctx) {
		if err := c.client.Del(ctx, iter.Val()).Err(); err != nil {
			c.logger.WithError(err).Warnw("Cache delete pattern error", "pattern", pattern)
		}
	}

	if err := iter.Err(); err != nil {
		return err
	}

	c.logger.Debugw("Cache delete pattern", "pattern", pattern)
	return nil
}

// Exists проверяет существование ключа
func (c *RedisCache) Exists(ctx context.Context, key string) (bool, error) {
	fullKey := c.prefix + key

	n, err := c.client.Exists(ctx, fullKey).Result()
	if err != nil {
		return false, err
	}

	return n > 0, nil
}

// Close закрывает подключение к Redis
func (c *RedisCache) Close() error {
	return c.client.Close()
}

// CacheKey генерирует ключ кеша
func CacheKey(parts ...string) string {
	key := ""
	for i, part := range parts {
		if i > 0 {
			key += ":"
		}
		key += part
	}
	return key
}

// NoOpCache заглушка для кеша (когда Redis недоступен)
type NoOpCache struct{}

func NewNoOpCache() *NoOpCache {
	return &NoOpCache{}
}

func (c *NoOpCache) Get(ctx context.Context, key string, dest interface{}) error {
	return fmt.Errorf("cache miss (noop)")
}

func (c *NoOpCache) Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	return nil
}

func (c *NoOpCache) Delete(ctx context.Context, key string) error {
	return nil
}

func (c *NoOpCache) DeletePattern(ctx context.Context, pattern string) error {
	return nil
}

func (c *NoOpCache) Exists(ctx context.Context, key string) (bool, error) {
	return false, nil
}

func (c *NoOpCache) Close() error {
	return nil
}
