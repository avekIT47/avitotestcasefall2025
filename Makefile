# ⚠️ ВНИМАНИЕ: Этот Makefile использует Unix-команды и не работает в Windows PowerShell
# Для Windows используйте: make -f Makefile.windows <команда>
# Или используйте WSL/Git Bash для полной совместимости
# Подробности: см. MAKEFILE_WINDOWS_COMPATIBILITY.md

.PHONY: help build run test clean docker-build docker-up docker-down migrate lint fmt load-test

# Переменные
APP_NAME=pr-reviewer
DOCKER_COMPOSE=docker-compose
GO=go
GOFLAGS=-v
MAIN_PATH=cmd/server/main.go
BINARY_NAME=main

# Цвета для вывода
GREEN=\033[0;32m
NC=\033[0m # No Color

help: ## Показать справку
	@echo "PR Reviewer Service - Makefile Commands"
	@echo ""
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "$(GREEN)%-20s$(NC) %s\n", $$1, $$2}'

build: ## Собрать приложение
	@echo "Building application..."
	@$(GO) build $(GOFLAGS) -o $(BINARY_NAME) $(MAIN_PATH)
	@echo "Build complete: $(BINARY_NAME)"

run: ## Запустить приложение локально
	@echo "Running application..."
	@$(GO) run $(MAIN_PATH)

test: ## Запустить unit тесты
	@echo "Running tests..."
	@$(GO) test $(GOFLAGS) ./...

test-coverage: ## Запустить тесты с покрытием
	@echo "Running tests with coverage..."
	@$(GO) test -coverprofile=coverage.out ./...
	@$(GO) tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report saved to coverage.html"

test-integration: ## Запустить интеграционные тесты
	@echo "Running integration tests..."
	@$(GO) test -tags=integration ./tests/...

clean: ## Очистить build артефакты
	@echo "Cleaning..."
	@rm -f $(BINARY_NAME)
	@rm -f coverage.out coverage.html
	@$(GO) clean
	@echo "Clean complete"

docker-build: ## Собрать Docker образы
	@echo "Building Docker images..."
	@$(DOCKER_COMPOSE) build
	@echo "Docker build complete"

docker-build-frontend: ## Собрать Docker образ фронтенда
	@echo "Building frontend Docker image..."
	@$(DOCKER_COMPOSE) build frontend
	@echo "Frontend Docker build complete"

docker-build-backend: ## Собрать Docker образ бэкенда
	@echo "Building backend Docker image..."
	@$(DOCKER_COMPOSE) build backend
	@echo "Backend Docker build complete"

docker-up: ## Запустить приложение через Docker Compose
	@echo "Starting services..."
	@$(DOCKER_COMPOSE) up -d
	@echo "Services started:"
	@echo "  - Frontend: http://localhost:3000"
	@echo "  - Backend API: http://localhost:8080"
	@echo "Waiting for services to be ready..."
	@sleep 5
	@curl -s http://localhost:8080/health > /dev/null && echo "✓ Backend service is healthy" || echo "✗ Backend service is not responding"
	@curl -s http://localhost:3000/health > /dev/null && echo "✓ Frontend service is healthy" || echo "✗ Frontend service is not responding"

docker-down: ## Остановить Docker Compose сервисы
	@echo "Stopping services..."
	@$(DOCKER_COMPOSE) down
	@echo "Services stopped"

docker-logs: ## Показать логи всех Docker контейнеров
	@$(DOCKER_COMPOSE) logs -f

docker-logs-frontend: ## Показать логи фронтенда
	@$(DOCKER_COMPOSE) logs -f frontend

docker-logs-backend: ## Показать логи бэкенда
	@$(DOCKER_COMPOSE) logs -f backend

docker-restart: docker-down docker-up ## Перезапустить Docker сервисы

docker-restart-frontend: ## Перезапустить фронтенд
	@$(DOCKER_COMPOSE) restart frontend

docker-restart-backend: ## Перезапустить бэкенд
	@$(DOCKER_COMPOSE) restart backend

migrate: ## Выполнить миграции БД
	@echo "Running migrations..."
	@migrate -path migrations -database "$(DATABASE_URL)" up
	@echo "Migrations complete"

migrate-down: ## Откатить последнюю миграцию
	@echo "Rolling back migration..."
	@migrate -path migrations -database "$(DATABASE_URL)" down 1
	@echo "Rollback complete"

lint: ## Запустить линтер
	@echo "Running linter..."
	@golangci-lint run
	@echo "Linting complete"

fmt: ## Форматировать код
	@echo "Formatting code..."
	@$(GO) fmt ./...
	@goimports -w .
	@echo "Formatting complete"

deps: ## Установить зависимости
	@echo "Installing dependencies..."
	@$(GO) mod download
	@$(GO) mod tidy
	@echo "Dependencies installed"

load-test: ## Запустить нагрузочное тестирование
	@echo "Running load tests..."
	@chmod +x tests/load_test.sh
	@./tests/load_test.sh

api-docs: ## Открыть API документацию (Swagger UI)
	@echo "Opening API documentation..."
	@npx @redocly/openapi-cli preview-docs openapi.yaml

install-tools: ## Установить необходимые инструменты
	@echo "Installing tools..."
	@go install github.com/golang-migrate/migrate/v4/cmd/migrate@latest
	@go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	@go install golang.org/x/tools/cmd/goimports@latest
	@echo "Tools installed"

seed-db: ## Заполнить БД тестовыми данными
	@echo "Seeding database with test data..."
	@./scripts/seed_db.sh
	@echo "Database seeded"

.DEFAULT_GOAL := help
