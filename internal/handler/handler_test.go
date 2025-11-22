package handler

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

// mockLogger для тестирования
type mockLogger struct{}

func (m *mockLogger) Printf(format string, args ...interface{}) {}

func TestHandler_Structure(t *testing.T) {
	h := &Handler{
		service: nil, // Would be a real service in production
		logger:  &mockLogger{},
	}

	if h == nil {
		t.Error("expected non-nil Handler")
	}
}

func TestHealthCheck(t *testing.T) {
	h := &Handler{
		service: nil,
		logger:  &mockLogger{},
	}

	req := httptest.NewRequest("GET", "/health", nil)
	w := httptest.NewRecorder()

	h.HealthCheck(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	body := w.Body.String()
	if body == "" {
		t.Error("expected non-empty response body")
	}

	// Check that response contains "healthy"
	if len(body) < 10 {
		t.Error("response body too short")
	}
}

func TestGetIntParam(t *testing.T) {
	h := &Handler{
		service: nil,
		logger:  &mockLogger{},
	}

	// Test with invalid value
	req := httptest.NewRequest("GET", "/test/abc", nil)
	_, err := h.getIntParam(req, "id")
	if err == nil {
		t.Error("expected error for non-numeric ID")
	}
}

func TestGetIntQuery(t *testing.T) {
	h := &Handler{
		service: nil,
		logger:  &mockLogger{},
	}

	// Test with empty query
	req := httptest.NewRequest("GET", "/test", nil)
	_, err := h.getIntQuery(req, "id")
	if err == nil {
		t.Error("expected error for missing query parameter")
	}

	// Test with invalid value
	req = httptest.NewRequest("GET", "/test?id=abc", nil)
	_, err = h.getIntQuery(req, "id")
	if err == nil {
		t.Error("expected error for non-numeric query parameter")
	}
}

func TestGetBoolQuery(t *testing.T) {
	h := &Handler{
		service: nil,
		logger:  &mockLogger{},
	}

	// Test with empty query
	req := httptest.NewRequest("GET", "/test", nil)
	_, err := h.getBoolQuery(req, "active")
	if err == nil {
		t.Error("expected error for missing query parameter")
	}

	// Test with valid value
	req = httptest.NewRequest("GET", "/test?active=true", nil)
	value, err := h.getBoolQuery(req, "active")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if !value {
		t.Error("expected true")
	}
}

func TestSendError(t *testing.T) {
	h := &Handler{
		service: nil,
		logger:  &mockLogger{},
	}

	w := httptest.NewRecorder()
	h.sendError(w, http.StatusBadRequest, "test error")

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", w.Code)
	}

	body := w.Body.String()
	if body == "" {
		t.Error("expected non-empty error response")
	}
}

func TestLoggingMiddleware(t *testing.T) {
	h := &Handler{
		service: nil,
		logger:  &mockLogger{},
	}

	nextCalled := false
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		nextCalled = true
		w.WriteHeader(http.StatusOK)
	})

	middleware := h.loggingMiddleware(next)

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()

	middleware.ServeHTTP(w, req)

	if !nextCalled {
		t.Error("expected next handler to be called")
	}

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}
}
