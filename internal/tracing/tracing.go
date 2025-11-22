package tracing

import (
	"context"
	"fmt"
	"net/http"

	"github.com/user/pr-reviewer/internal/logger"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.21.0"
	"go.opentelemetry.io/otel/trace"
)

// Config конфигурация трейсинга
type Config struct {
	Enabled     bool
	ServiceName string
	Environment string
	OTLPURL     string // OTLP endpoint URL (например, http://localhost:4318/v1/traces)
	SampleRate  float64
}

// Tracer обертка над OpenTelemetry tracer
type Tracer struct {
	tracer trace.Tracer
	logger *logger.Logger
	config Config
}

// Init инициализирует OpenTelemetry tracing
func Init(cfg Config, log *logger.Logger) (*Tracer, error) {
	if !cfg.Enabled {
		log.Info("Distributed tracing is disabled")
		return &Tracer{
			tracer: otel.Tracer(cfg.ServiceName),
			logger: log,
			config: cfg,
		}, nil
	}

	// Создаем OTLP HTTP exporter (поддерживает Jaeger через OTLP)
	exp, err := otlptracehttp.New(
		context.Background(),
		otlptracehttp.WithEndpoint(cfg.OTLPURL),
		otlptracehttp.WithInsecure(), // Для локальной разработки
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create OTLP exporter: %w", err)
	}

	// Создаем resource с информацией о сервисе
	res, err := resource.New(
		context.Background(),
		resource.WithAttributes(
			semconv.ServiceName(cfg.ServiceName),
			semconv.DeploymentEnvironment(cfg.Environment),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create resource: %w", err)
	}

	// Создаем trace provider
	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exp),
		sdktrace.WithResource(res),
		sdktrace.WithSampler(sdktrace.TraceIDRatioBased(cfg.SampleRate)),
	)

	// Регистрируем глобальный trace provider
	otel.SetTracerProvider(tp)

	// Настраиваем propagator для передачи контекста между сервисами
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	))

	log.Infow("Distributed tracing initialized",
		"service", cfg.ServiceName,
		"otlp_url", cfg.OTLPURL,
		"sample_rate", cfg.SampleRate,
	)

	return &Tracer{
		tracer: otel.Tracer(cfg.ServiceName),
		logger: log,
		config: cfg,
	}, nil
}

// Start начинает новый span
func (t *Tracer) Start(ctx context.Context, spanName string, opts ...trace.SpanStartOption) (context.Context, trace.Span) {
	return t.tracer.Start(ctx, spanName, opts...)
}

// Middleware HTTP middleware для трейсинга
func (t *Tracer) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !t.config.Enabled {
			next.ServeHTTP(w, r)
			return
		}

		// Извлекаем контекст из headers (если есть)
		ctx := otel.GetTextMapPropagator().Extract(r.Context(), propagation.HeaderCarrier(r.Header))

		// Создаем span для HTTP запроса
		ctx, span := t.Start(ctx, fmt.Sprintf("%s %s", r.Method, r.URL.Path),
			trace.WithSpanKind(trace.SpanKindServer),
			trace.WithAttributes(
				semconv.HTTPMethod(r.Method),
				semconv.HTTPRoute(r.URL.Path),
				semconv.HTTPScheme(r.URL.Scheme),
				semconv.HTTPTarget(r.URL.RequestURI()),
				attribute.String("net.peer.ip", r.RemoteAddr),
				semconv.UserAgentOriginal(r.UserAgent()),
			),
		)
		defer span.End()

		// Оборачиваем ResponseWriter для захвата статус кода
		wrapped := &statusWriter{ResponseWriter: w, statusCode: http.StatusOK}

		// Выполняем запрос
		next.ServeHTTP(wrapped, r.WithContext(ctx))

		// Добавляем статус код в span
		span.SetAttributes(semconv.HTTPStatusCode(wrapped.statusCode))

		// Отмечаем ошибки
		if wrapped.statusCode >= 400 {
			span.SetAttributes(attribute.Bool("error", true))
		}
	})
}

// TraceDBQuery трейсит запрос к БД
func (t *Tracer) TraceDBQuery(ctx context.Context, query string, queryType string) (context.Context, func(error)) {
	if !t.config.Enabled {
		return ctx, func(error) {}
	}

	ctx, span := t.Start(ctx, "db.query",
		trace.WithSpanKind(trace.SpanKindClient),
		trace.WithAttributes(
			attribute.String("db.system", "postgresql"),
			attribute.String("db.operation", queryType),
			attribute.String("db.statement", query),
		),
	)

	return ctx, func(err error) {
		if err != nil {
			span.SetAttributes(attribute.Bool("error", true))
			span.RecordError(err)
		}
		span.End()
	}
}

// TraceExternalCall трейсит внешний HTTP вызов
func (t *Tracer) TraceExternalCall(ctx context.Context, method, url string) (context.Context, func(int, error)) {
	if !t.config.Enabled {
		return ctx, func(int, error) {}
	}

	ctx, span := t.Start(ctx, fmt.Sprintf("%s %s", method, url),
		trace.WithSpanKind(trace.SpanKindClient),
		trace.WithAttributes(
			semconv.HTTPMethod(method),
			attribute.String("http.url", url),
		),
	)

	return ctx, func(statusCode int, err error) {
		span.SetAttributes(semconv.HTTPStatusCode(statusCode))
		if err != nil || statusCode >= 400 {
			span.SetAttributes(attribute.Bool("error", true))
			if err != nil {
				span.RecordError(err)
			}
		}
		span.End()
	}
}

// AddEvent добавляет событие в текущий span
func (t *Tracer) AddEvent(ctx context.Context, name string, attrs ...attribute.KeyValue) {
	if !t.config.Enabled {
		return
	}
	span := trace.SpanFromContext(ctx)
	span.AddEvent(name, trace.WithAttributes(attrs...))
}

// SetAttributes добавляет атрибуты в текущий span
func (t *Tracer) SetAttributes(ctx context.Context, attrs ...attribute.KeyValue) {
	if !t.config.Enabled {
		return
	}
	span := trace.SpanFromContext(ctx)
	span.SetAttributes(attrs...)
}

// RecordError записывает ошибку в span
func (t *Tracer) RecordError(ctx context.Context, err error) {
	if !t.config.Enabled || err == nil {
		return
	}
	span := trace.SpanFromContext(ctx)
	span.RecordError(err)
	span.SetAttributes(attribute.Bool("error", true))
}

// statusWriter обертка для захвата статус кода
type statusWriter struct {
	http.ResponseWriter
	statusCode int
}

func (w *statusWriter) WriteHeader(code int) {
	w.statusCode = code
	w.ResponseWriter.WriteHeader(code)
}

// Helper: извлечение trace ID для логирования
func GetTraceID(ctx context.Context) string {
	span := trace.SpanFromContext(ctx)
	if !span.SpanContext().IsValid() {
		return ""
	}
	return span.SpanContext().TraceID().String()
}

// Helper: извлечение span ID
func GetSpanID(ctx context.Context) string {
	span := trace.SpanFromContext(ctx)
	if !span.SpanContext().IsValid() {
		return ""
	}
	return span.SpanContext().SpanID().String()
}
