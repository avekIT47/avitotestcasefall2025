package logger

import (
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Logger обертка над zap logger для структурированного логирования
type Logger struct {
	*zap.SugaredLogger
}

// New создает новый структурированный logger
func New(level string, env string) (*Logger, error) {
	var zapLevel zapcore.Level
	switch level {
	case "debug":
		zapLevel = zapcore.DebugLevel
	case "info":
		zapLevel = zapcore.InfoLevel
	case "warn":
		zapLevel = zapcore.WarnLevel
	case "error":
		zapLevel = zapcore.ErrorLevel
	default:
		zapLevel = zapcore.InfoLevel
	}

	// Конфигурация для разных окружений
	var config zap.Config
	if env == "production" {
		// JSON формат для production
		config = zap.NewProductionConfig()
		config.EncoderConfig.TimeKey = "timestamp"
		config.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	} else {
		// Readable формат для development
		config = zap.NewDevelopmentConfig()
		config.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	}

	config.Level = zap.NewAtomicLevelAt(zapLevel)
	config.OutputPaths = []string{"stdout"}
	config.ErrorOutputPaths = []string{"stderr"}

	// Добавляем caller info для трейсинга
	config.EncoderConfig.CallerKey = "caller"
	config.EncoderConfig.EncodeCaller = zapcore.ShortCallerEncoder

	// Добавляем stacktrace для errors
	config.EncoderConfig.StacktraceKey = "stacktrace"

	zapLogger, err := config.Build(
		zap.AddCaller(),
		zap.AddCallerSkip(1),
		zap.AddStacktrace(zapcore.ErrorLevel),
	)
	if err != nil {
		return nil, err
	}

	// Добавляем общие поля для всех логов
	zapLogger = zapLogger.With(
		zap.String("service", "pr-reviewer"),
		zap.String("environment", env),
		zap.Int("pid", os.Getpid()),
	)

	return &Logger{
		SugaredLogger: zapLogger.Sugar(),
	}, nil
}

// WithRequestID добавляет request ID в контекст логов
func (l *Logger) WithRequestID(requestID string) *Logger {
	return &Logger{
		SugaredLogger: l.With(zap.String("request_id", requestID)),
	}
}

// WithUser добавляет информацию о пользователе
func (l *Logger) WithUser(userID int64) *Logger {
	return &Logger{
		SugaredLogger: l.With(zap.Int64("user_id", userID)),
	}
}

// WithError добавляет информацию об ошибке
func (l *Logger) WithError(err error) *Logger {
	return &Logger{
		SugaredLogger: l.With(zap.Error(err)),
	}
}

// LogHTTPRequest логирует HTTP запрос
func (l *Logger) LogHTTPRequest(method, path, ip string, statusCode, duration int) {
	l.Infow("HTTP request",
		"method", method,
		"path", path,
		"ip", ip,
		"status", statusCode,
		"duration_ms", duration,
	)
}

// LogDBQuery логирует запрос к БД
func (l *Logger) LogDBQuery(query string, duration int64, err error) {
	if err != nil {
		l.Errorw("Database query failed",
			"query", query,
			"duration_ms", duration,
			"error", err,
		)
	} else {
		l.Debugw("Database query",
			"query", query,
			"duration_ms", duration,
		)
	}
}

// Close корректно закрывает logger
func (l *Logger) Close() error {
	return l.Sync()
}
