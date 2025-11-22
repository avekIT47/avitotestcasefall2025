package metrics

import (
	"strconv"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// Metrics содержит все метрики приложения
type Metrics struct {
	// HTTP метрики
	HTTPRequestsTotal    *prometheus.CounterVec
	HTTPRequestDuration  *prometheus.HistogramVec
	HTTPRequestsInFlight prometheus.Gauge

	// Database метрики
	DBQueryDuration    *prometheus.HistogramVec
	DBConnectionsOpen  prometheus.Gauge
	DBConnectionsInUse prometheus.Gauge

	// Business метрики
	PRCreatedTotal         prometheus.Counter
	PRMergedTotal          prometheus.Counter
	ReviewersAssignedTotal prometheus.Counter
	UsersDeactivatedTotal  prometheus.Counter

	// Application метрики
	AppUptime prometheus.Gauge
	AppInfo   *prometheus.GaugeVec
}

var metrics *Metrics

// Init инициализирует метрики Prometheus
func Init(namespace string) *Metrics {
	metrics = &Metrics{
		// HTTP метрики
		HTTPRequestsTotal: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: namespace,
				Name:      "http_requests_total",
				Help:      "Total number of HTTP requests",
			},
			[]string{"method", "endpoint", "status"},
		),
		HTTPRequestDuration: promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Namespace: namespace,
				Name:      "http_request_duration_seconds",
				Help:      "HTTP request latencies in seconds",
				Buckets:   []float64{0.001, 0.005, 0.01, 0.05, 0.1, 0.25, 0.5, 1, 2.5, 5, 10},
			},
			[]string{"method", "endpoint"},
		),
		HTTPRequestsInFlight: promauto.NewGauge(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Name:      "http_requests_in_flight",
				Help:      "Current number of HTTP requests being served",
			},
		),

		// Database метрики
		DBQueryDuration: promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Namespace: namespace,
				Name:      "db_query_duration_seconds",
				Help:      "Database query duration in seconds",
				Buckets:   []float64{0.001, 0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1},
			},
			[]string{"query_type"},
		),
		DBConnectionsOpen: promauto.NewGauge(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Name:      "db_connections_open",
				Help:      "Number of open database connections",
			},
		),
		DBConnectionsInUse: promauto.NewGauge(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Name:      "db_connections_in_use",
				Help:      "Number of database connections currently in use",
			},
		),

		// Business метрики
		PRCreatedTotal: promauto.NewCounter(
			prometheus.CounterOpts{
				Namespace: namespace,
				Name:      "pull_requests_created_total",
				Help:      "Total number of pull requests created",
			},
		),
		PRMergedTotal: promauto.NewCounter(
			prometheus.CounterOpts{
				Namespace: namespace,
				Name:      "pull_requests_merged_total",
				Help:      "Total number of pull requests merged",
			},
		),
		ReviewersAssignedTotal: promauto.NewCounter(
			prometheus.CounterOpts{
				Namespace: namespace,
				Name:      "reviewers_assigned_total",
				Help:      "Total number of reviewers assigned",
			},
		),
		UsersDeactivatedTotal: promauto.NewCounter(
			prometheus.CounterOpts{
				Namespace: namespace,
				Name:      "users_deactivated_total",
				Help:      "Total number of users deactivated",
			},
		),

		// Application метрики
		AppUptime: promauto.NewGauge(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Name:      "app_uptime_seconds",
				Help:      "Application uptime in seconds",
			},
		),
		AppInfo: promauto.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Name:      "app_info",
				Help:      "Application information",
			},
			[]string{"version", "go_version", "environment"},
		),
	}

	return metrics
}

// Get возвращает глобальный экземпляр метрик
func Get() *Metrics {
	return metrics
}

// RecordHTTPRequest записывает метрики HTTP запроса
func (m *Metrics) RecordHTTPRequest(method, endpoint string, statusCode int, duration time.Duration) {
	m.HTTPRequestsTotal.WithLabelValues(method, endpoint, strconv.Itoa(statusCode)).Inc()
	m.HTTPRequestDuration.WithLabelValues(method, endpoint).Observe(duration.Seconds())
}

// RecordDBQuery записывает метрики запроса к БД
func (m *Metrics) RecordDBQuery(queryType string, duration time.Duration) {
	m.DBQueryDuration.WithLabelValues(queryType).Observe(duration.Seconds())
}

// IncrementInFlightRequests увеличивает счетчик активных запросов
func (m *Metrics) IncrementInFlightRequests() {
	m.HTTPRequestsInFlight.Inc()
}

// DecrementInFlightRequests уменьшает счетчик активных запросов
func (m *Metrics) DecrementInFlightRequests() {
	m.HTTPRequestsInFlight.Dec()
}

// SetDBStats устанавливает статистику БД
func (m *Metrics) SetDBStats(open, inUse int) {
	m.DBConnectionsOpen.Set(float64(open))
	m.DBConnectionsInUse.Set(float64(inUse))
}

// SetAppInfo устанавливает информацию о приложении
func (m *Metrics) SetAppInfo(version, goVersion, environment string) {
	m.AppInfo.WithLabelValues(version, goVersion, environment).Set(1)
}

// UpdateUptime обновляет метрику uptime
func (m *Metrics) UpdateUptime(startTime time.Time) {
	m.AppUptime.Set(time.Since(startTime).Seconds())
}
