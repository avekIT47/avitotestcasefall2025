package config

import (
	"fmt"
	"os"
	"strconv"
)

// Config содержит конфигурацию приложения
type Config struct {
	ServerPort     string
	DatabaseURL    string
	MigrationsPath string
	LogLevel       string
}

// Load загружает конфигурацию из окружения
func Load() (*Config, error) {
	cfg := &Config{
		ServerPort:     getEnv("SERVER_PORT", "8080"),
		DatabaseURL:    getEnv("DATABASE_URL", "postgres://postgres:postgres@db:5432/pr_reviewer?sslmode=disable"),
		MigrationsPath: getEnv("MIGRATIONS_PATH", "file://migrations"),
		LogLevel:       getEnv("LOG_LEVEL", "info"),
	}

	if cfg.DatabaseURL == "" {
		return nil, fmt.Errorf("DATABASE_URL is required")
	}

	return cfg, nil
}

// getEnv получает значение переменной окружения или возвращает значение по умолчанию
func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}

// getEnvAsInt получает значение переменной окружения как int
func getEnvAsInt(name string, defaultVal int) int {
	valueStr := getEnv(name, "")
	if value, err := strconv.Atoi(valueStr); err == nil {
		return value
	}
	return defaultVal
}

// getEnvAsBool получает значение переменной окружения как bool
func getEnvAsBool(name string, defaultVal bool) bool {
	valStr := getEnv(name, "")
	if val, err := strconv.ParseBool(valStr); err == nil {
		return val
	}
	return defaultVal
}
