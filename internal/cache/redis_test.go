package cache

import (
	"context"
	"testing"
	"time"
)

func TestKey(t *testing.T) {
	tests := []struct {
		name     string
		parts    []string
		expected string
	}{
		{
			name:     "single part",
			parts:    []string{"user"},
			expected: "user",
		},
		{
			name:     "multiple parts",
			parts:    []string{"user", "123", "profile"},
			expected: "user:123:profile",
		},
		{
			name:     "empty",
			parts:    []string{},
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Key(tt.parts...)
			if result != tt.expected {
				t.Errorf("expected '%s', got '%s'", tt.expected, result)
			}
		})
	}
}

func TestNoOpCache(t *testing.T) {
	cache := NewNoOpCache()
	ctx := context.Background()

	// Test Get - should always return error
	var dest interface{}
	err := cache.Get(ctx, "test", &dest)
	if err == nil {
		t.Error("expected error from NoOpCache.Get")
	}

	// Test Set - should not error
	err = cache.Set(ctx, "test", "value", 1*time.Minute)
	if err != nil {
		t.Errorf("NoOpCache.Set should not error, got: %v", err)
	}

	// Test Delete - should not error
	err = cache.Delete(ctx, "test")
	if err != nil {
		t.Errorf("NoOpCache.Delete should not error, got: %v", err)
	}

	// Test DeletePattern - should not error
	err = cache.DeletePattern(ctx, "test:*")
	if err != nil {
		t.Errorf("NoOpCache.DeletePattern should not error, got: %v", err)
	}

	// Test Exists - should always return false
	exists, err := cache.Exists(ctx, "test")
	if err != nil {
		t.Errorf("NoOpCache.Exists should not error, got: %v", err)
	}
	if exists {
		t.Error("NoOpCache.Exists should always return false")
	}

	// Test Close - should not error
	err = cache.Close()
	if err != nil {
		t.Errorf("NoOpCache.Close should not error, got: %v", err)
	}
}

func TestCacheInterface(t *testing.T) {
	// Test that NoOpCache implements Cache interface
	var _ Cache = (*NoOpCache)(nil)
}

