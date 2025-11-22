//go:build production
// +build production

package main

// Пример интеграции всех production-ready компонентов
// Используйте с флагом: go build -tags production

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"runtime"
	"syscall"
	"time"

	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/rs/cors"

	"github.com/user/pr-reviewer/internal/auth"
	"github.com/user/pr-reviewer/internal/cache"
	"github.com/user/pr-reviewer/internal/config"
	"github.com/user/pr-reviewer/internal/database"
	"github.com/user/pr-reviewer/internal/handler"
	"github.com/user/pr-reviewer/internal/health"
	"github.com/user/pr-reviewer/internal/logger"
	"github.com/user/pr-reviewer/internal/metrics"
	"github.com/user/pr-reviewer/internal/middleware"
	"github.com/user/pr-reviewer/internal/service"
)

const (
	Version     = "1.0.0"
	ServiceName = "pr-reviewer"
)

func main() {
	// Загрузка конфигурации
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to load config: %v\n", err)
		os.Exit(1)
	}

	// Инициализация структурированного логгера
	environment := getEnv("ENVIRONMENT", "development")
	log, err := logger.New(cfg.LogLevel, environment)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create logger: %v\n", err)
		os.Exit(1)
	}
	defer log.Close()

	log.Infow("Starting PR Reviewer Service",
		"version", Version,
		"go_version", runtime.Version(),
		"environment", environment,
		"port", cfg.ServerPort,
	)

	// Инициализация метрик
	met := metrics.Init(ServiceName)
	met.SetAppInfo(Version, runtime.Version(), environment)

	// Старт времени для uptime
	startTime := time.Now()
	go func() {
		ticker := time.NewTicker(30 * time.Second)
		defer ticker.Stop()
		for range ticker.C {
			met.UpdateUptime(startTime)
		}
	}()

	// Подключение к базе данных
	db, err := database.New(cfg.DatabaseURL)
	if err != nil {
		log.Fatalw("Failed to connect to database", "error", err)
	}
	defer db.Close()

	log.Info("Connected to database")

	// Периодическое обновление метрик БД
	go func() {
		ticker := time.NewTicker(15 * time.Second)
		defer ticker.Stop()
		for range ticker.C {
			stats := db.DB.Stats()
			met.SetDBStats(stats.OpenConnections, stats.InUse)
		}
	}()

	// Выполнение миграций
	if err := db.Migrate(cfg.MigrationsPath); err != nil {
		log.Fatalw("Failed to run migrations", "error", err)
	}
	log.Info("Database migrations completed")

	// Инициализация кеша
	var cacheClient cache.Cache
	redisAddr := getEnv("REDIS_ADDR", "")
	if redisAddr != "" {
		redisPassword := getEnv("REDIS_PASSWORD", "")
		redisDB := getEnvAsInt("REDIS_DB", 0)

		cacheClient, err = cache.NewRedisCache(
			redisAddr,
			redisPassword,
			redisDB,
			ServiceName+":",
			log,
		)
		if err != nil {
			log.Warnw("Failed to connect to Redis, using no-op cache", "error", err)
			cacheClient = cache.NewNoOpCache()
		}
	} else {
		log.Info("Redis not configured, using no-op cache")
		cacheClient = cache.NewNoOpCache()
	}
	defer cacheClient.Close()

	// Инициализация health checks
	healthChecker := health.New(Version, log)
	healthChecker.RegisterChecker(health.NewDatabaseChecker(db.DB))
	healthChecker.RegisterChecker(health.NewSystemChecker())

	// Инициализация JWT аутентификации (опционально)
	var jwtAuth *auth.JWTAuth
	jwtSecret := getEnv("JWT_SECRET", "")
	if jwtSecret != "" && jwtSecret != "change_me_in_production" {
		jwtExpiration := getEnvAsDuration("JWT_EXPIRATION", 24*time.Hour)
		jwtAuth = auth.NewJWTAuth(jwtSecret, jwtExpiration, log)
		log.Info("JWT authentication enabled")
	} else {
		log.Warn("JWT authentication disabled (JWT_SECRET not set)")
	}

	// Инициализация сервисов
	svc := service.New(db)

	// Инициализация HTTP обработчиков
	h := handler.New(svc, log)

	// Настройка middleware
	mw := middleware.New(log, met)

	// Настройка маршрутов
	router := mux.NewRouter()

	// Health endpoints (без аутентификации)
	router.HandleFunc("/health", healthChecker.Handler()).Methods("GET")
	router.HandleFunc("/health/live", healthChecker.LivenessHandler()).Methods("GET")
	router.HandleFunc("/health/ready", healthChecker.ReadinessHandler()).Methods("GET")

	// Metrics endpoint (для Prometheus)
	router.Handle("/metrics", promhttp.Handler()).Methods("GET")

	// API routes с middleware
	apiRouter := router.PathPrefix("/").Subrouter()

	// Применяем middleware
	middlewareChain := middleware.Chain(
		mw.RequestID,
		mw.Logging,
		mw.Metrics,
		mw.Recovery,
		mw.SecurityHeaders,
		mw.RateLimit,
		mw.RequestValidation,
	)

	// Регистрируем API routes
	h.RegisterRoutes(apiRouter)

	// Применяем middleware ко всему API router
	router.Use(middlewareChain)

	// Опционально: добавляем JWT аутентификацию
	if jwtAuth != nil {
		// Можно сделать селективную аутентификацию:
		// - Публичные эндпоинты (GET /teams, GET /users) - без аутентификации
		// - Мутирующие эндпоинты - с аутентификацией
		// router.Use(jwtAuth.OptionalMiddleware)
	}

	// Настройка CORS
	corsHandler := cors.New(cors.Options{
		AllowedOrigins:   getAllowedOrigins(),
		AllowedMethods:   []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Content-Type", "Authorization", "X-Requested-With", "Accept", "X-Request-ID"},
		ExposedHeaders:   []string{"X-Request-ID"},
		AllowCredentials: true,
		MaxAge:           3600,
	}).Handler(router)

	// Настройка HTTP сервера
	srv := &http.Server{
		Addr:         ":" + cfg.ServerPort,
		Handler:      corsHandler,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
		// MaxHeaderBytes: 1 << 20, // 1MB
	}

	// Запуск сервера в отдельной горутине
	go func() {
		log.Infow("Server listening",
			"port", cfg.ServerPort,
			"environment", environment,
		)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalw("Failed to start server", "error", err)
		}
	}()

	// Ожидание сигнала завершения
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	sig := <-quit

	log.Infow("Shutting down server", "signal", sig)

	// Graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Errorw("Server forced to shutdown", "error", err)
	}

	log.Info("Server exited gracefully")
}

func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}

func getEnvAsInt(key string, defaultValue int) int {
	valueStr := getEnv(key, "")
	if valueStr == "" {
		return defaultValue
	}
	var value int
	if _, err := fmt.Sscanf(valueStr, "%d", &value); err == nil {
		return value
	}
	return defaultValue
}

func getEnvAsDuration(key string, defaultValue time.Duration) time.Duration {
	valueStr := getEnv(key, "")
	if valueStr == "" {
		return defaultValue
	}
	if value, err := time.ParseDuration(valueStr); err == nil {
		return value
	}
	return defaultValue
}

func getAllowedOrigins() []string {
	originsStr := getEnv("ALLOWED_ORIGINS", "http://localhost:3000,http://localhost:80")
	origins := []string{}
	for _, origin := range splitAndTrim(originsStr, ",") {
		if origin != "" {
			origins = append(origins, origin)
		}
	}
	return origins
}

func splitAndTrim(s, sep string) []string {
	parts := []string{}
	for _, part := range split(s, sep) {
		trimmed := trim(part)
		if trimmed != "" {
			parts = append(parts, trimmed)
		}
	}
	return parts
}

func split(s, sep string) []string {
	result := []string{}
	current := ""
	for _, ch := range s {
		if string(ch) == sep {
			result = append(result, current)
			current = ""
		} else {
			current += string(ch)
		}
	}
	if current != "" {
		result = append(result, current)
	}
	return result
}

func trim(s string) string {
	start := 0
	end := len(s)

	for start < end && (s[start] == ' ' || s[start] == '\t' || s[start] == '\n') {
		start++
	}

	for end > start && (s[end-1] == ' ' || s[end-1] == '\t' || s[end-1] == '\n') {
		end--
	}

	return s[start:end]
}
