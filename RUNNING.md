# 🚀 Инструкция по запуску PR Reviewer Service

> Подробное руководство по запуску и развертыванию сервиса в различных окружениях

---

## 📋 Содержание

- [Системные требования](#системные-требования)
- [Быстрый старт через Docker](#быстрый-старт-через-docker)
- [Локальная разработка](#локальная-разработка)
- [Запуск через Make](#запуск-через-make)
- [Ручная настройка окружения](#ручная-настройка-окружения)
- [Production развертывание](#production-развертывание)
- [Проверка работоспособности](#проверка-работоспособности)
- [Troubleshooting](#troubleshooting)

---

## 🖥️ Системные требования

### Минимальные требования

| Компонент | Версия | Назначение |
|-----------|--------|------------|
| **Docker** | 20.10+ | Контейнеризация |
| **Docker Compose** | 2.0+ | Оркестрация контейнеров |
| **Make** | 3.81+ | Автоматизация команд (опционально) |

### Для локальной разработки (опционально)

| Компонент | Версия | Назначение |
|-----------|--------|------------|
| **Go** | 1.21+ | Backend разработка |
| **Node.js** | 18+ | Frontend разработка |
| **PostgreSQL** | 15+ | База данных |
| **Redis** | 7+ | Кэширование |
| **golangci-lint** | 1.55+ | Линтинг кода |

### Операционные системы

- ✅ **Linux** (Ubuntu 20.04+, Debian 11+)
- ✅ **macOS** (10.15+)
- ✅ **Windows** 10/11 (с WSL2 или Git Bash для Make команд)

> **⚠️ Для Windows:** Makefile использует Unix-команды. Используйте WSL2, Git Bash или `Makefile.windows`

---

## 🎯 Быстрый старт через Docker

### Вариант 1: Один Docker Compose файл (Production)

Самый простой способ запустить весь стек приложения:

```bash
# 1. Клонировать репозиторий
git clone <repository-url>
cd avito_testcase

# 2. Запустить все сервисы
docker-compose up -d

# 3. Подождать ~30 секунд для инициализации
sleep 30

# 4. Проверить статус сервисов
docker-compose ps

# 5. Проверить здоровье backend
curl http://localhost:8080/health
```

**Результат:**
- ✅ Backend API: http://localhost:8080
- ✅ Frontend: http://localhost:3000
- ✅ Prometheus: http://localhost:9090
- ✅ Grafana: http://localhost:3001 (admin/admin)

### Вариант 2: Используя Makefile

```bash
# Запустить все сервисы
make docker-up

# Посмотреть логи
make docker-logs

# Остановить сервисы
make docker-down
```

### Порты по умолчанию

| Сервис | Порт | URL | Описание |
|--------|------|-----|----------|
| Frontend | 3000 | http://localhost:3000 | React SPA |
| Backend | 8080 | http://localhost:8080 | REST API |
| PostgreSQL | 5432 | localhost:5432 | База данных |
| Redis | 6379 | localhost:6379 | Кэш |
| Prometheus | 9090 | http://localhost:9090 | Метрики |
| Grafana | 3001 | http://localhost:3001 | Дашборды |

---

## 💻 Локальная разработка

### Backend (Go)

#### 1. Установка зависимостей

```bash
# Установить Go зависимости
go mod download
go mod tidy

# Установить необходимые инструменты
make install-tools

# Или вручную
go install github.com/golang-migrate/migrate/v4/cmd/migrate@latest
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
go install golang.org/x/tools/cmd/goimports@latest
```

#### 2. Запустить инфраструктуру (БД, Redis)

```bash
# Запустить только БД и Redis
docker-compose up -d db redis

# Проверить статус
docker-compose ps

# Проверить что БД доступна
docker-compose exec db pg_isready -U postgres
```

#### 3. Настроить переменные окружения

Создать файл `.env` в корне проекта:

```bash
# Application
SERVER_PORT=8080
LOG_LEVEL=debug
ENVIRONMENT=development

# Database
DATABASE_URL=postgres://postgres:postgres@localhost:5432/pr_reviewer?sslmode=disable
MIGRATIONS_PATH=file://migrations

# Redis
REDIS_ADDR=localhost:6379
REDIS_PASSWORD=
REDIS_DB=0

# JWT
JWT_SECRET=dev_secret_key_change_in_production
JWT_EXPIRATION=24h

# Rate limiting
RATE_LIMIT_RPS=100
RATE_LIMIT_BURST=200
```

#### 4. Выполнить миграции

```bash
# Через Makefile
make migrate

# Или вручную
migrate -path migrations \
  -database "postgres://postgres:postgres@localhost:5432/pr_reviewer?sslmode=disable" \
  up
```

#### 5. Запустить backend

```bash
# Через Makefile
make run

# Или напрямую через Go
go run cmd/server/main.go

# Или с переменными окружения
DATABASE_URL="postgres://postgres:postgres@localhost:5432/pr_reviewer?sslmode=disable" \
  go run cmd/server/main.go
```

Backend будет доступен на http://localhost:8080

#### 6. Проверить работоспособность

```bash
# Health check
curl http://localhost:8080/health

# Получить статистику
curl http://localhost:8080/statistics

# Создать команду
curl -X POST http://localhost:8080/teams \
  -H "Content-Type: application/json" \
  -d '{"name": "Test Team"}'
```

### Frontend (React + TypeScript)

#### 1. Установить зависимости

```bash
cd frontend

# Установить npm пакеты
npm install
```

#### 2. Настроить переменные окружения

Создать файл `frontend/.env`:

```bash
# API URL (backend)
VITE_API_URL=http://localhost:8080
VITE_API_BASE_URL=http://localhost:8080

# Environment
VITE_ENVIRONMENT=development
```

#### 3. Запустить dev сервер

```bash
# Запустить Vite dev server с HMR
npm run dev

# Dev server будет доступен на http://localhost:5173
```

#### 4. Сборка production версии

```bash
# Собрать для production
npm run build

# Превью production сборки
npm run preview
```

---

## 🛠️ Запуск через Make

Makefile предоставляет удобные команды для работы с проектом.

### Просмотр всех команд

```bash
make help
```

### Основные команды

| Команда | Описание |
|---------|----------|
| `make help` | Показать справку по всем командам |
| `make build` | Собрать Go приложение локально |
| `make run` | Запустить backend локально |
| `make test` | Запустить unit тесты |
| `make test-coverage` | Тесты с покрытием (coverage.html) |
| `make test-integration` | Интеграционные тесты |
| `make lint` | Запустить golangci-lint |
| `make fmt` | Форматировать Go код |
| `make deps` | Установить/обновить зависимости |
| `make migrate` | Выполнить миграции БД |
| `make clean` | Очистить build артефакты |

### Docker команды

| Команда | Описание |
|---------|----------|
| `make docker-build` | Собрать Docker образы |
| `make docker-up` | Запустить все сервисы |
| `make docker-down` | Остановить все сервисы |
| `make docker-logs` | Показать логи всех контейнеров |
| `make docker-logs-backend` | Логи backend |
| `make docker-logs-frontend` | Логи frontend |
| `make docker-restart` | Перезапустить все сервисы |
| `make docker-restart-backend` | Перезапустить backend |
| `make docker-restart-frontend` | Перезапустить frontend |

### Дополнительные команды

| Команда | Описание |
|---------|----------|
| `make load-test` | Нагрузочное тестирование |
| `make seed-db` | Заполнить БД тестовыми данными |
| `make install-tools` | Установить dev инструменты |
| `make api-docs` | Открыть API документацию |

### Примеры использования

```bash
# 1. Полный цикл разработки
make deps              # Установить зависимости
make fmt               # Форматировать код
make lint              # Проверить линтером
make test              # Запустить тесты
make build             # Собрать приложение

# 2. Работа с Docker
make docker-build      # Собрать образы
make docker-up         # Запустить
make docker-logs       # Посмотреть логи

# 3. Работа с БД
make migrate           # Применить миграции
make seed-db           # Заполнить тестовыми данными

# 4. Тестирование
make test-coverage     # Unit тесты с покрытием
make test-integration  # Интеграционные тесты
make load-test         # Нагрузочное тестирование
```

### Windows пользователям

```bash
# Если Makefile не работает в PowerShell:

# Вариант 1: Использовать Windows-версию
make -f Makefile.windows <команда>

# Вариант 2: Использовать WSL2 (рекомендуется)
wsl make <команда>

# Вариант 3: Использовать Git Bash
# Запустить Git Bash и использовать make как обычно
```

---

## ⚙️ Ручная настройка окружения

Если вы не хотите использовать Docker, можно настроить окружение вручную.

### 1. Установить PostgreSQL

#### Ubuntu/Debian

```bash
sudo apt update
sudo apt install postgresql postgresql-contrib
sudo systemctl start postgresql
sudo systemctl enable postgresql
```

#### macOS

```bash
brew install postgresql@15
brew services start postgresql@15
```

#### Создать базу данных

```bash
# Подключиться к PostgreSQL
sudo -u postgres psql

# Создать базу и пользователя
CREATE DATABASE pr_reviewer;
CREATE USER pr_user WITH PASSWORD 'pr_password';
GRANT ALL PRIVILEGES ON DATABASE pr_reviewer TO pr_user;
\q
```

### 2. Установить Redis

#### Ubuntu/Debian

```bash
sudo apt install redis-server
sudo systemctl start redis-server
sudo systemctl enable redis-server
```

#### macOS

```bash
brew install redis
brew services start redis
```

### 3. Установить Go

```bash
# Ubuntu/Debian
wget https://go.dev/dl/go1.21.0.linux-amd64.tar.gz
sudo tar -C /usr/local -xzf go1.21.0.linux-amd64.tar.gz
echo 'export PATH=$PATH:/usr/local/go/bin' >> ~/.bashrc
source ~/.bashrc

# macOS
brew install go@1.21
```

### 4. Установить golang-migrate

```bash
go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest
```

### 5. Выполнить миграции

```bash
migrate -path migrations \
  -database "postgres://pr_user:pr_password@localhost:5432/pr_reviewer?sslmode=disable" \
  up
```

### 6. Запустить приложение

```bash
# Установить переменные окружения
export DATABASE_URL="postgres://pr_user:pr_password@localhost:5432/pr_reviewer?sslmode=disable"
export REDIS_ADDR="localhost:6379"
export SERVER_PORT="8080"
export LOG_LEVEL="info"

# Запустить
go run cmd/server/main.go
```

---

## 🏭 Production развертывание

### Docker Compose (Production)

#### 1. Создать .env файл

```bash
# .env
ENVIRONMENT=production
LOG_LEVEL=info
SERVER_PORT=8080

# Используйте СИЛЬНЫЕ пароли!
POSTGRES_PASSWORD=STRONG_SECURE_PASSWORD_HERE
REDIS_PASSWORD=STRONG_REDIS_PASSWORD_HERE
JWT_SECRET=SUPER_STRONG_JWT_SECRET_256_BITS

# Rate limiting
RATE_LIMIT_RPS=100
RATE_LIMIT_BURST=200

# Grafana
GRAFANA_USER=admin
GRAFANA_PASSWORD=STRONG_GRAFANA_PASSWORD

# Версия образа
VERSION=1.0.0
```

#### 2. Запустить production стек

```bash
# Собрать образы
docker-compose -f docker-compose.prod.yaml build

# Запустить сервисы
docker-compose -f docker-compose.prod.yaml up -d

# Проверить статус
docker-compose -f docker-compose.prod.yaml ps

# Посмотреть логи
docker-compose -f docker-compose.prod.yaml logs -f
```

#### 3. Проверить health checks

```bash
# Backend health
curl http://localhost:8080/health

# Проверить метрики
curl http://localhost:8080/metrics

# Проверить frontend
curl http://localhost:3000
```

### Kubernetes

#### 1. Создать namespace

```bash
kubectl apply -f k8s/namespace.yaml
```

#### 2. Создать secrets

```bash
# Создать секреты для БД
kubectl create secret generic postgres-secret \
  --from-literal=password=STRONG_PASSWORD \
  -n pr-reviewer

# Создать секреты для JWT
kubectl create secret generic jwt-secret \
  --from-literal=secret=SUPER_STRONG_JWT_SECRET \
  -n pr-reviewer
```

#### 3. Применить манифесты

```bash
# Применить все манифесты
kubectl apply -f k8s/

# Проверить pods
kubectl get pods -n pr-reviewer

# Проверить services
kubectl get svc -n pr-reviewer

# Проверить ingress
kubectl get ingress -n pr-reviewer
```

#### 4. Масштабирование

```bash
# Увеличить количество реплик backend
kubectl scale deployment pr-reviewer-backend --replicas=3 -n pr-reviewer

# Включить автоскейлинг
kubectl autoscale deployment pr-reviewer-backend \
  --cpu-percent=70 \
  --min=2 \
  --max=10 \
  -n pr-reviewer
```

### Мониторинг в Production

```bash
# Проверить метрики Prometheus
curl http://localhost:9090/api/v1/targets

# Открыть Grafana
# http://localhost:3001
# Логин: admin
# Пароль: см. GRAFANA_PASSWORD в .env

# Проверить алерты
curl http://localhost:9090/api/v1/alerts
```

---

## ✅ Проверка работоспособности

### Автоматическая проверка

```bash
# Скрипт быстрой проверки
./quickstart.sh
```

### Ручная проверка

#### 1. Health Check

```bash
curl http://localhost:8080/health

# Ожидаемый ответ:
# {
#   "status": "healthy",
#   "version": "1.0.0",
#   "database": "connected",
#   "redis": "connected"
# }
```

#### 2. Создать команду

```bash
curl -X POST http://localhost:8080/teams \
  -H "Content-Type: application/json" \
  -d '{"name": "Backend Team"}'

# Ответ:
# {"id": 1, "name": "Backend Team", "createdAt": "..."}
```

#### 3. Создать пользователя

```bash
curl -X POST http://localhost:8080/users \
  -H "Content-Type: application/json" \
  -d '{
    "username": "john_doe",
    "name": "John Doe",
    "teamId": 1
  }'
```

#### 4. Создать Pull Request

```bash
curl -X POST http://localhost:8080/pull-requests \
  -H "Content-Type: application/json" \
  -d '{
    "title": "Add new feature",
    "authorId": 1
  }'

# Ответ будет содержать автоматически назначенных рецензентов
```

#### 5. Получить статистику

```bash
curl http://localhost:8080/statistics

# Ответ:
# {
#   "total": 1,
#   "open": 1,
#   "merged": 0,
#   "closed": 0,
#   "assignmentsByUser": {...}
# }
```

#### 6. Проверить frontend

Открыть в браузере:
- Frontend: http://localhost:3000
- Grafana: http://localhost:3001 (admin/admin)
- Prometheus: http://localhost:9090

### Заполнение тестовыми данными

```bash
# Выполнить скрипт заполнения
./scripts/seed_db.sh

# Или через Makefile
make seed-db
```

Скрипт создаст:
- 3 команды (Backend, Frontend, DevOps)
- 15 пользователей (по 5 в каждой команде)
- 20 Pull Request'ов с назначенными рецензентами

---

## 🔧 Troubleshooting

### Проблема: Backend не запускается

**Симптомы:**
```
Error: failed to connect to database
```

**Решение:**

```bash
# 1. Проверить что БД запущена
docker-compose ps db

# 2. Проверить логи БД
docker-compose logs db

# 3. Проверить подключение к БД
docker-compose exec db pg_isready -U postgres

# 4. Проверить что миграции выполнены
docker-compose exec db psql -U postgres -d pr_reviewer -c "\dt"

# 5. Перезапустить БД
docker-compose restart db

# 6. Если не помогает - пересоздать volume
docker-compose down -v
docker-compose up -d
```

### Проблема: Frontend не может подключиться к Backend

**Симптомы:**
- Frontend отображает ошибки API
- В консоли браузера: `CORS error` или `Network error`

**Решение:**

```bash
# 1. Проверить что backend запущен
curl http://localhost:8080/health

# 2. Проверить переменные окружения frontend
cat frontend/.env
# Должно быть: VITE_API_URL=http://localhost:8080

# 3. Проверить CORS настройки backend
curl -I http://localhost:8080/health

# 4. Перезапустить frontend
docker-compose restart frontend
# или для dev
cd frontend && npm run dev
```

### Проблема: Порт уже занят

**Симптомы:**
```
Error: bind: address already in use
```

**Решение:**

```bash
# Найти процесс, использующий порт (например 8080)
# Linux/macOS:
lsof -i :8080
kill -9 <PID>

# Windows:
netstat -ano | findstr :8080
taskkill /PID <PID> /F

# Или изменить порт в .env или docker-compose.yaml
```

### Проблема: Миграции не применяются

**Симптомы:**
```
Error: Dirty database version
```

**Решение:**

```bash
# 1. Проверить статус миграций
migrate -path migrations \
  -database "$DATABASE_URL" \
  version

# 2. Сбросить dirty state
migrate -path migrations \
  -database "$DATABASE_URL" \
  force <version>

# 3. Или пересоздать БД
docker-compose down -v
docker-compose up -d db
make migrate
```

### Проблема: Redis недоступен

**Симптомы:**
```
Error: failed to connect to redis
```

**Решение:**

```bash
# 1. Проверить что Redis запущен
docker-compose ps redis

# 2. Проверить подключение
docker-compose exec redis redis-cli ping
# Должен ответить: PONG

# 3. Если Redis требует пароль
docker-compose exec redis redis-cli -a "$REDIS_PASSWORD" ping

# 4. Перезапустить Redis
docker-compose restart redis
```

### Проблема: Docker Compose не запускается на Windows

**Симптомы:**
- Ошибки с путями
- Проблемы с line endings

**Решение:**

```bash
# 1. Убедитесь что используете WSL2 backend для Docker
# Docker Desktop -> Settings -> General -> Use WSL 2 based engine

# 2. Конвертировать line endings
dos2unix docker-compose.yaml
dos2unix Dockerfile

# 3. Или клонировать репозиторий с правильными настройками
git config --global core.autocrlf false
git clone <repo>

# 4. Использовать WSL2
wsl
cd /mnt/c/path/to/project
docker-compose up -d
```

### Проблема: Медленная работа приложения

**Решение:**

```bash
# 1. Проверить метрики
curl http://localhost:8080/metrics

# 2. Проверить логи
docker-compose logs backend | tail -100

# 3. Проверить ресурсы
docker stats

# 4. Увеличить лимиты в docker-compose.yaml
# resources:
#   limits:
#     cpus: '1'
#     memory: 1G

# 5. Проверить индексы БД
docker-compose exec db psql -U postgres -d pr_reviewer -c "\d+ users"
```

### Получение помощи

```bash
# Посмотреть все доступные команды
make help

# Собрать диагностическую информацию
docker-compose ps
docker-compose logs --tail=100
curl http://localhost:8080/health
curl http://localhost:8080/metrics

# Проверить версии
docker --version
docker-compose --version
go version
```

---

## 📚 Дополнительные ресурсы

- **README.md** - Основная документация проекта
- **openapi.yaml** - API спецификация
- **examples/api_examples.http** - Примеры всех API запросов
- **LINTER.md** - Описание конфигурации линтера

---

## 🎉 Готово!

После выполнения инструкций сервис должен быть запущен и доступен:

- ✅ Backend API: http://localhost:8080
- ✅ Frontend: http://localhost:3000
- ✅ Prometheus: http://localhost:9090
- ✅ Grafana: http://localhost:3001

**Happy Coding! 🚀**

