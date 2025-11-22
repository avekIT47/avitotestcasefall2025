package logger

import (
	"bytes"
	"testing"
)

func TestNew(t *testing.T) {
	logger, err := New("info", "json")
	if err != nil {
		t.Fatalf("failed to create logger: %v", err)
	}

	if logger == nil {
		t.Error("expected non-nil logger")
	}
}

func TestNew_InvalidLevel(t *testing.T) {
	// Invalid level defaults to "info" - no error
	logger, err := New("invalid", "json")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if logger == nil {
		t.Error("expected non-nil logger")
	}
}

func TestLogger_Levels(t *testing.T) {
	tests := []struct {
		name  string
		level string
	}{
		{"debug level", "debug"},
		{"info level", "info"},
		{"warn level", "warn"},
		{"error level", "error"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger, err := New(tt.level, "json")
			if err != nil {
				t.Fatalf("failed to create logger: %v", err)
			}

			if logger == nil {
				t.Error("expected non-nil logger")
			}
		})
	}
}

func TestLogger_Formats(t *testing.T) {
	tests := []struct {
		name   string
		format string
	}{
		{"json format", "json"},
		{"console format", "console"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger, err := New("info", tt.format)
			if err != nil {
				t.Fatalf("failed to create logger: %v", err)
			}

			if logger == nil {
				t.Error("expected non-nil logger")
			}
		})
	}
}

func TestLogger_Methods(t *testing.T) {
	logger, err := New("debug", "json")
	if err != nil {
		t.Fatalf("failed to create logger: %v", err)
	}

	// Test that methods don't panic
	logger.Info("test info")
	logger.Infof("test %s", "info")
	logger.Warn("test warn")
	logger.Error("test error")
	logger.Debug("test debug")
	logger.Infow("test", "key", "value")
	logger.Warnw("test", "key", "value")
	logger.Errorw("test", "key", "value")
	logger.Debugw("test", "key", "value")

	// Test WithError
	withErr := logger.WithError(bytes.ErrTooLarge)
	if withErr == nil {
		t.Error("expected non-nil logger with error")
	}

	withErr.Warnw("test")
	withErr.Errorw("test")
}

func TestLogger_Sync(t *testing.T) {
	logger, err := New("info", "json")
	if err != nil {
		t.Fatalf("failed to create logger: %v", err)
	}

	// Test that Sync doesn't panic
	_ = logger.Sync()
}

func TestLogger_Output(t *testing.T) {
	// This is a basic test - in production you'd want to capture actual output
	logger, err := New("info", "console")
	if err != nil {
		t.Fatalf("failed to create logger: %v", err)
	}

	// Just verify the logger works without panicking
	logger.Info("test message")
	logger.Infof("test %s", "formatted")
	logger.Infow("test", "key", "value")
}

func TestLogLevels(t *testing.T) {
	levels := []string{"debug", "info", "warn", "error"}
	for _, level := range levels {
		t.Run(level, func(t *testing.T) {
			logger, err := New(level, "json")
			if err != nil {
				t.Fatalf("failed to create logger with level %s: %v", level, err)
			}
			if logger == nil {
				t.Error("expected non-nil logger")
			}
		})
	}
}

func TestLogFormats(t *testing.T) {
	formats := []string{"json", "console"}
	for _, format := range formats {
		t.Run(format, func(t *testing.T) {
			logger, err := New("info", format)
			if err != nil {
				t.Fatalf("failed to create logger with format %s: %v", format, err)
			}
			if logger == nil {
				t.Error("expected non-nil logger")
			}
		})
	}
}

func TestVariousConfigurations(t *testing.T) {
	tests := []struct {
		name   string
		level  string
		format string
	}{
		{"invalid level defaults to info", "invalid", "json"},
		{"production env", "info", "production"},
		{"development env", "debug", "development"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger, err := New(tt.level, tt.format)
			if err != nil {
				t.Fatalf("unexpected error for level=%s format=%s: %v", tt.level, tt.format, err)
			}
			if logger == nil {
				t.Error("expected non-nil logger")
			}
		})
	}
}
