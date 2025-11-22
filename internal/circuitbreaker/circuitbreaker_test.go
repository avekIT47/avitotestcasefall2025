package circuitbreaker

import (
	"errors"
	"testing"
	"time"

	"github.com/sony/gobreaker"
	"github.com/user/pr-reviewer/internal/logger"
)

func TestCircuitBreaker_Success(t *testing.T) {
	log, _ := logger.New("error", "test")
	cfg := NewDefaultConfig("test")
	cb := New(cfg, log)

	// Успешный вызов
	result, err := cb.Execute(func() (interface{}, error) {
		return "success", nil
	})

	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if result != "success" {
		t.Errorf("Expected 'success', got %v", result)
	}

	if cb.State() != gobreaker.StateClosed {
		t.Errorf("Expected Closed state, got %v", cb.State())
	}
}

func TestCircuitBreaker_Opens(t *testing.T) {
	log, _ := logger.New("error", "test")
	cfg := Config{
		Name:        "test",
		MaxRequests: 1,
		Interval:    1 * time.Second,
		Timeout:     100 * time.Millisecond,
		ReadyToTrip: func(counts gobreaker.Counts) bool {
			// Открываем после 3 ошибок подряд
			return counts.ConsecutiveFailures >= 3
		},
	}
	cb := New(cfg, log)

	// Генерируем 3 ошибки
	testErr := errors.New("test error")
	for i := 0; i < 3; i++ {
		_, err := cb.Execute(func() (interface{}, error) {
			return nil, testErr
		})
		if err != testErr {
			t.Errorf("Expected test error, got %v", err)
		}
	}

	// Circuit breaker должен быть открыт
	if cb.State() != gobreaker.StateOpen {
		t.Errorf("Expected Open state, got %v", cb.State())
	}

	// Следующий запрос должен сразу вернуть ErrOpenState
	_, err := cb.Execute(func() (interface{}, error) {
		return "should not execute", nil
	})

	if err != gobreaker.ErrOpenState {
		t.Errorf("Expected ErrOpenState, got %v", err)
	}
}

func TestCircuitBreaker_HalfOpen(t *testing.T) {
	log, _ := logger.New("error", "test")
	cfg := Config{
		Name:        "test",
		MaxRequests: 1,
		Interval:    1 * time.Second,
		Timeout:     50 * time.Millisecond, // Короткий timeout для теста
		ReadyToTrip: func(counts gobreaker.Counts) bool {
			return counts.ConsecutiveFailures >= 2
		},
	}
	cb := New(cfg, log)

	// Генерируем 2 ошибки чтобы открыть breaker
	testErr := errors.New("test error")
	for i := 0; i < 2; i++ {
		cb.Execute(func() (interface{}, error) {
			return nil, testErr
		})
	}

	// Ждем timeout чтобы breaker перешел в half-open
	time.Sleep(60 * time.Millisecond)

	// В half-open состоянии успешный запрос закроет breaker
	result, err := cb.Execute(func() (interface{}, error) {
		return "recovered", nil
	})

	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if result != "recovered" {
		t.Errorf("Expected 'recovered', got %v", result)
	}

	// После успеха должен быть Closed
	time.Sleep(10 * time.Millisecond)
	if cb.State() != gobreaker.StateClosed {
		t.Errorf("Expected Closed state after recovery, got %v", cb.State())
	}
}

func TestCircuitBreaker_WithFallback(t *testing.T) {
	log, _ := logger.New("error", "test")
	cfg := NewDefaultConfig("test")
	cb := New(cfg, log)

	// Основная функция возвращает ошибку
	result, err := cb.WithFallback(
		func() (interface{}, error) {
			return nil, errors.New("primary failed")
		},
		func() (interface{}, error) {
			return "fallback value", nil
		},
	)

	if err != nil {
		t.Errorf("Expected no error with fallback, got %v", err)
	}

	if result != "fallback value" {
		t.Errorf("Expected 'fallback value', got %v", result)
	}
}

func TestCircuitBreakerManager(t *testing.T) {
	log, _ := logger.New("error", "test")
	manager := NewManager(log)

	// Регистрируем несколько breakers
	cfg1 := NewDefaultConfig("service1")
	cfg2 := NewDefaultConfig("service2")

	manager.Register("service1", cfg1)
	manager.Register("service2", cfg2)

	// Проверяем что можем получить breakers
	cb1, err := manager.Get("service1")
	if err != nil {
		t.Errorf("Expected to find service1, got error: %v", err)
	}
	if cb1 == nil {
		t.Error("Expected non-nil circuit breaker")
	}

	cb2, err := manager.Get("service2")
	if err != nil {
		t.Errorf("Expected to find service2, got error: %v", err)
	}
	if cb2 == nil {
		t.Error("Expected non-nil circuit breaker")
	}

	// Проверяем что несуществующий breaker возвращает ошибку
	_, err = manager.Get("nonexistent")
	if err == nil {
		t.Error("Expected error for nonexistent circuit breaker")
	}

	// Проверяем GetStates
	states := manager.GetStates()
	if len(states) != 2 {
		t.Errorf("Expected 2 states, got %d", len(states))
	}
}
