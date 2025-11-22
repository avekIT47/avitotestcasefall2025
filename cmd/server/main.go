//go:build !production
// +build !production

package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gorilla/mux"
	"github.com/rs/cors"
	"github.com/user/pr-reviewer/internal/config"
	"github.com/user/pr-reviewer/internal/database"
	"github.com/user/pr-reviewer/internal/handler"
	"github.com/user/pr-reviewer/internal/service"
)

func main() {
	// Инициализация логгера
	logger := log.New(os.Stdout, "[PR-REVIEWER] ", log.LstdFlags|log.Lshortfile)

	// Загрузка конфигурации
	cfg, err := config.Load()
	if err != nil {
		logger.Fatalf("Failed to load config: %v", err)
	}

	logger.Printf("Starting PR Reviewer Service on port %s", cfg.ServerPort)

	// Подключение к базе данных
	db, err := database.New(cfg.DatabaseURL)
	if err != nil {
		logger.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	logger.Println("Connected to database")

	// Выполнение миграций
	if err := db.Migrate(cfg.MigrationsPath); err != nil {
		logger.Fatalf("Failed to run migrations: %v", err)
	}

	logger.Println("Database migrations completed")

	// Инициализация сервисов
	svc := service.New(db)

	// Инициализация HTTP обработчиков
	h := handler.New(svc, logger)

	// Настройка маршрутов
	router := mux.NewRouter()
	h.RegisterRoutes(router)

	// Настройка CORS
	corsHandler := cors.New(cors.Options{
		AllowedOrigins:   []string{"http://localhost:3000", "http://localhost:80"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Content-Type", "Authorization", "X-Requested-With", "Accept"},
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
	}

	// Запуск сервера в отдельной горутине
	go func() {
		logger.Printf("Server listening on port %s", cfg.ServerPort)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatalf("Failed to start server: %v", err)
		}
	}()

	// Ожидание сигнала завершения
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Println("Shutting down server...")

	// Graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		logger.Printf("Server forced to shutdown: %v", err)
	}

	logger.Println("Server exited")
}

func init() {
	// Проверка обязательных переменных окружения при старте
	requiredEnvVars := []string{
		"DATABASE_URL",
	}

	for _, envVar := range requiredEnvVars {
		if os.Getenv(envVar) == "" {
			// Если переменная не установлена, используем значение по умолчанию
			switch envVar {
			case "DATABASE_URL":
				os.Setenv("DATABASE_URL", "postgres://postgres:postgres@db:5432/pr_reviewer?sslmode=disable")
			}
		}
	}

	fmt.Println("PR Reviewer Service - initializing...")
}
