package config

import (
	"os"
	"testing"
)

func TestLoad(t *testing.T) {
	// Set test environment variables
	os.Setenv("SERVER_PORT", "8080")
	os.Setenv("DATABASE_URL", "postgres://test")
	os.Setenv("LOG_LEVEL", "info")
	os.Setenv("MIGRATIONS_PATH", "file://test/migrations")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("failed to load config: %v", err)
	}

	if cfg.ServerPort != "8080" {
		t.Errorf("expected port '8080', got '%s'", cfg.ServerPort)
	}

	if cfg.DatabaseURL != "postgres://test" {
		t.Errorf("expected database URL 'postgres://test', got '%s'", cfg.DatabaseURL)
	}

	if cfg.LogLevel != "info" {
		t.Errorf("expected log level 'info', got '%s'", cfg.LogLevel)
	}

	if cfg.MigrationsPath != "file://test/migrations" {
		t.Errorf("expected migrations path 'file://test/migrations', got '%s'", cfg.MigrationsPath)
	}

	// Clean up
	os.Unsetenv("SERVER_PORT")
	os.Unsetenv("DATABASE_URL")
	os.Unsetenv("LOG_LEVEL")
	os.Unsetenv("MIGRATIONS_PATH")
}

func TestLoad_Defaults(t *testing.T) {
	// Clear any existing env vars
	os.Unsetenv("SERVER_PORT")
	os.Unsetenv("DATABASE_URL")
	os.Unsetenv("LOG_LEVEL")
	os.Unsetenv("MIGRATIONS_PATH")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("failed to load config: %v", err)
	}

	// Check defaults
	if cfg.ServerPort != "8080" {
		t.Errorf("expected default port '8080', got '%s'", cfg.ServerPort)
	}

	if cfg.LogLevel != "info" {
		t.Errorf("expected default log level 'info', got '%s'", cfg.LogLevel)
	}

	if cfg.MigrationsPath != "file://migrations" {
		t.Errorf("expected default migrations path 'file://migrations', got '%s'", cfg.MigrationsPath)
	}
}

func TestConfig_Structure(t *testing.T) {
	cfg := &Config{
		ServerPort:     "8080",
		DatabaseURL:    "postgres://test",
		MigrationsPath: "file://migrations",
		LogLevel:       "info",
	}

	if cfg.ServerPort != "8080" {
		t.Errorf("expected port '8080', got '%s'", cfg.ServerPort)
	}

	if cfg.DatabaseURL != "postgres://test" {
		t.Errorf("expected database URL 'postgres://test', got '%s'", cfg.DatabaseURL)
	}

	if cfg.LogLevel != "info" {
		t.Errorf("expected log level 'info', got '%s'", cfg.LogLevel)
	}

	if cfg.MigrationsPath != "file://migrations" {
		t.Errorf("expected migrations path 'file://migrations', got '%s'", cfg.MigrationsPath)
	}
}

func TestGetEnv(t *testing.T) {
	os.Setenv("TEST_VAR", "test_value")
	defer os.Unsetenv("TEST_VAR")

	value := getEnv("TEST_VAR", "default")
	if value != "test_value" {
		t.Errorf("expected 'test_value', got '%s'", value)
	}

	value = getEnv("NON_EXISTENT", "default")
	if value != "default" {
		t.Errorf("expected 'default', got '%s'", value)
	}
}
