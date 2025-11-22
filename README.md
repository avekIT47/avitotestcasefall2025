# PR Reviewer Assignment Service

**Сервис автоматического назначения рецензентов для Pull Request**

> 🚀 Production-ready микросервис для эффективного распределения нагрузки code review между членами команды с поддержкой мониторинга, метрик и масштабирования.

---

## 📋 Содержание

- [Описание проекта](#описание-проекта)
- [Архитектура](#архитектура)
- [Ключевые возможности](#ключевые-возможности)
- [Технологический стек](#технологический-стек)
- [Структура проекта](#структура-проекта)
- [API Documentation](#api-documentation)
- [Быстрый старт](#быстрый-старт)
- [Production развертывание](#production-развертывание)
- [Мониторинг и метрики](#мониторинг-и-метрики)
- [Тестирование](#тестирование)

---

## 🎯 Описание проекта

PR Reviewer Assignment Service - это корпоративное решение для автоматизации процесса назначения рецензентов на Pull Request'ы. Система обеспечивает равномерное распределение нагрузки по code review между членами команды, с учетом их активности и текущей загрузки.

### Проблема, которую решает сервис:

1. **Неравномерное распределение нагрузки** - некоторые разработчики перегружены ревью, другие простаивают
2. **Ручное назначение** - тратится время на выбор рецензента вместо автоматизации
3. **Отсутствие статистики** - нет прозрачности в процессе code review
4. **Проблемы с отпусками/болезнями** - необходимость переназначения ревьюеров

### Решение:

- ✅ Автоматическое назначение 1-2 рецензентов на основе нагрузки
- ✅ Учет активности пользователей (активные/неактивные)
- ✅ Управление командами и их участниками
- ✅ Детальная статистика по назначениям
- ✅ Bulk операции для массового управления
- ✅ Production-ready инфраструктура

---

## 🏗️ Архитектура

### High-Level Architecture

```
┌─────────────────┐
│   Frontend      │ React 19 + TypeScript + Tailwind CSS
│   (SPA)         │ Vite, React Router, Zustand, TanStack Query
└────────┬────────┘
         │ HTTP/REST
         │
┌────────▼────────┐
│   Nginx         │ Reverse Proxy + Static Assets
│   (Gateway)     │ Rate Limiting, SSL/TLS
└────────┬────────┘
         │
┌────────▼────────┐
│   Backend       │ Go 1.21 + Gorilla Mux
│   (API Server)  │ RESTful API, Business Logic
└─────┬──┬────┬───┘
      │  │    │
      │  │    └─────────────┐
      │  │                  │
┌─────▼──▼───┐        ┌─────▼──────┐
│ PostgreSQL │        │   Redis    │
│   (DB)     │        │  (Cache)   │
└────────────┘        └────────────┘
      │
┌─────▼──────────────────────────┐
│   Monitoring Stack             │
│ - Prometheus (Metrics)         │
│ - Grafana (Dashboards)         │
└────────────────────────────────┘
```

### Backend Architecture (Clean Architecture)

```
cmd/server/
  └── main.go                    # Точка входа приложения

internal/
  ├── config/                    # Конфигурация
  │   └── config.go              # Загрузка переменных окружения
  │
  ├── database/                  # Слой БД
  │   └── database.go            # Подключение и миграции
  │
  ├── models/                    # Доменные модели
  │   └── models.go              # User, Team, PullRequest, Statistics
  │
  ├── repository/                # Слой данных (DAL)
  │   ├── team_repository.go     # CRUD для команд
  │   ├── user_repository.go     # CRUD для пользователей
  │   ├── pr_repository.go       # CRUD для PR
  │   └── statistics_repository.go # Аналитика
  │
  ├── service/                   # Бизнес-логика
  │   ├── service.go             # Основная логика
  │   └── service_test.go        # Unit тесты
  │
  ├── handler/                   # HTTP обработчики
  │   └── handler.go             # REST API endpoints
  │
  ├── middleware/                # HTTP middleware
  │   └── middleware.go          # Rate limiting, CORS, Auth
  │
  ├── auth/                      # Аутентификация
  │   └── jwt.go                 # JWT токены
  │
  ├── cache/                     # Кэширование
  │   └── redis.go               # Redis клиент
  │
  ├── metrics/                   # Метрики
  │   └── metrics.go             # Prometheus metrics
  │
  ├── logger/                    # Логирование
  │   └── logger.go              # Structured logging
  │
  ├── health/                    # Health checks
  │   └── health.go              # Readiness/Liveness
  │
  ├── tracing/                   # Distributed tracing
  │   └── tracing.go             # OpenTelemetry
  │
  ├── audit/                     # Audit logging
  │   └── audit.go               # Audit trails
  │
  ├── webhook/                   # Webhooks
  │   └── webhook.go             # Webhook notifications
  │
  ├── circuitbreaker/            # Circuit Breaker
  │   ├── circuitbreaker.go      # Реализация CB
  │   └── circuitbreaker_test.go # Тесты CB
  │
  └── featureflags/              # Feature Flags
      ├── featureflags.go        # Feature toggles
      └── featureflags_test.go   # Тесты FF
```

### Принципы архитектуры:

1. **Clean Architecture** - разделение на слои (handler → service → repository → database)
2. **Dependency Injection** - инверсия зависимостей через интерфейсы
3. **Single Responsibility** - каждый компонент отвечает за одну задачу
4. **Repository Pattern** - абстракция работы с БД
5. **Graceful Shutdown** - корректная остановка сервиса

---

## ✨ Ключевые возможности

### Основной функционал

#### 1. Управление командами
- Создание, чтение, удаление команд
- Добавление/удаление пользователей в команды
- Массовая деактивация пользователей команды
- Просмотр статистики по команде

#### 2. Управление пользователями
- Создание пользователей с привязкой к команде
- Обновление информации (имя, активность)
- Поддержка username для уникальной идентификации
- Флаг активности (is_active) для управления доступностью
- Автоматическая обработка неактивных пользователей при назначении

#### 3. Pull Request Management
- Создание PR с автоматическим назначением рецензентов
- **Умное назначение**: 1-2 рецензента из команды автора
- **Алгоритм балансировки нагрузки**: учитывает текущее количество назначений
- Исключение автора PR из списка рецензентов
- Только активные пользователи назначаются рецензентами
- Merge PR (переход в статус MERGED)
- Close PR (переход в статус CLOSED)
- Ручное добавление рецензентов
- Переназначение рецензента (при увольнении/болезни)
- Фильтрация PR по статусу (OPEN, MERGED, CLOSED)
- Поддержка сортировки и пагинации

#### 4. Статистика и аналитика
- Общая статистика по всем PR (total, open, merged, closed)
- Распределение назначений по пользователям
- Статистика по командам
- Top рецензенты
- История изменений (audit logs)
- Метрики производительности

#### 5. Bulk операции
- Массовая деактивация пользователей
- Автоматическое переназначение PR от деактивированных пользователей
- Транзакционная обработка (все или ничего)

### Production возможности

#### Observability (Наблюдаемость)
- **Metrics** - Prometheus метрики (requests, latency, errors)
- **Logging** - Structured JSON logging с уровнями
- **Tracing** - Distributed tracing (OpenTelemetry ready)
- **Audit** - Полный audit trail всех операций
- **Health Checks** - Readiness и Liveness probes

#### Resilience (Отказоустойчивость)
- **Circuit Breaker** - защита от каскадных сбоев
- **Rate Limiting** - защита от перегрузки
- **Graceful Shutdown** - корректная остановка
- **Retry Logic** - повторные попытки при сбоях
- **Timeout Control** - контроль таймаутов

#### Performance (Производительность)
- **Redis Cache** - кэширование часто используемых данных
- **Database Indexes** - оптимизированные индексы
- **Connection Pooling** - пул соединений к БД
- **Concurrent Processing** - конкурентная обработка запросов

#### Security (Безопасность)
- **JWT Authentication** - аутентификация через токены
- **CORS Protection** - защита от CSRF
- **SQL Injection Protection** - параметризованные запросы
- **Rate Limiting** - защита от DDoS
- **Input Validation** - валидация входных данных

#### Scalability (Масштабируемость)
- **Stateless Architecture** - возможность горизонтального масштабирования
- **Docker Support** - контейнеризация
- **Kubernetes Ready** - готовые манифесты K8s
- **Database Migrations** - версионирование схемы БД
- **Feature Flags** - постепенный rollout новых функций

#### DevOps
- **Docker Compose** - локальная разработка и prod окружение
- **Multi-stage Build** - оптимизированные Docker образы
- **Makefile** - автоматизация команд
- **CI/CD Ready** - готовность к CI/CD pipeline
- **Monitoring Stack** - Prometheus + Grafana из коробки

---

## 🛠️ Технологический стек

### Backend

| Компонент | Технология | Версия | Назначение |
|-----------|------------|--------|------------|
| **Язык** | Go | 1.21+ | Основной язык backend |
| **HTTP Router** | Gorilla Mux | 1.8.1 | HTTP роутинг |
| **Database** | PostgreSQL | 15 | Основная БД |
| **Cache** | Redis | 7 | Кэширование |
| **DB Driver** | lib/pq | 1.10.9 | PostgreSQL драйвер |
| **Migrations** | golang-migrate | 4.17.0 | Миграции БД |
| **CORS** | rs/cors | 1.11.1 | CORS middleware |
| **Testing** | testify | 1.11.1 | Unit/Integration тесты |
| **Monitoring** | Prometheus | latest | Метрики |
| **Visualization** | Grafana | latest | Дашборды |
| **Reverse Proxy** | Nginx | alpine | API Gateway |

### Frontend

| Компонент | Технология | Версия | Назначение |
|-----------|------------|--------|------------|
| **Framework** | React | 19.2.0 | UI библиотека |
| **Language** | TypeScript | 5.9.3 | Типизация |
| **Build Tool** | Vite | 7.2.4 | Сборка и dev server |
| **Routing** | React Router | 7.9.6 | Клиентский роутинг |
| **State** | Zustand | 5.0.8 | Глобальное состояние |
| **Data Fetching** | TanStack Query | 5.90.10 | Асинхронные запросы |
| **HTTP Client** | Axios | 1.13.2 | HTTP клиент |
| **Styling** | Tailwind CSS | 3.4.18 | CSS фреймворк |
| **UI Components** | Headless UI | 2.2.9 | Доступные компоненты |
| **Icons** | Heroicons | 2.2.0 | Иконки |
| **Animations** | Framer Motion | 12.23.24 | Анимации |
| **Charts** | Recharts | 3.4.1 | Графики статистики |
| **i18n** | i18next | 25.6.3 | Интернационализация |
| **Notifications** | react-hot-toast | 2.6.0 | Уведомления |
| **PDF Export** | jsPDF | 3.0.4 | Экспорт в PDF |
| **CSV Export** | react-csv | 2.2.2 | Экспорт в CSV |
| **Date Utils** | date-fns | 4.1.0 | Работа с датами |

### DevOps & Infrastructure

- **Containerization**: Docker + Docker Compose
- **Orchestration**: Kubernetes (manifests в `/k8s`)
- **Connection Pooling**: PgBouncer (конфиг в `/pgbouncer`)
- **Load Testing**: Vegeta (скрипты в `/tests/performance`)
- **API Docs**: OpenAPI 3.0 (swagger)

---

## 📁 Структура проекта

```
avito_testcase/
│
├── cmd/                           # Точки входа приложения
│   └── server/
│       ├── main.go                # Основной сервер
│       └── main_production.go     # Production версия
│
├── internal/                      # Приватный код приложения
│   ├── config/                    # Конфигурация
│   ├── database/                  # Работа с БД
│   ├── models/                    # Доменные модели
│   ├── repository/                # Слой данных
│   │   ├── team_repository.go
│   │   ├── user_repository.go
│   │   ├── pr_repository.go
│   │   └── statistics_repository.go
│   ├── service/                   # Бизнес-логика
│   │   ├── service.go
│   │   └── service_test.go
│   ├── handler/                   # HTTP handlers
│   ├── middleware/                # HTTP middleware
│   ├── auth/                      # JWT аутентификация
│   ├── cache/                     # Redis кэш
│   ├── metrics/                   # Prometheus метрики
│   ├── logger/                    # Логирование
│   ├── health/                    # Health checks
│   ├── tracing/                   # Tracing
│   ├── audit/                     # Audit logs
│   ├── webhook/                   # Webhooks
│   ├── circuitbreaker/            # Circuit Breaker
│   └── featureflags/              # Feature Flags
│
├── migrations/                    # Миграции БД
│   ├── 001_initial_schema.up.sql
│   ├── 001_initial_schema.down.sql
│   ├── 002_add_closed_status.up.sql
│   ├── 002_add_closed_status.down.sql
│   ├── 003_add_audit_logs.up.sql
│   ├── 004_add_webhooks.up.sql
│   ├── 005_add_username_field.up.sql
│   └── 006_add_updated_at.up.sql
│
├── tests/                         # Тесты
│   ├── integration_test.go        # Интеграционные тесты
│   ├── load_test.sh               # Нагрузочные тесты
│   └── performance/
│       └── vegeta_test.sh         # Vegeta тесты
│
├── frontend/                      # Frontend приложение
│   ├── src/
│   │   ├── components/            # React компоненты
│   │   │   ├── Auth/
│   │   │   ├── Dashboard/
│   │   │   ├── PullRequests/
│   │   │   ├── Teams/
│   │   │   ├── Users/
│   │   │   ├── Statistics/
│   │   │   ├── Settings/
│   │   │   ├── Layout/
│   │   │   └── UI/
│   │   ├── services/              # API клиент
│   │   │   └── api.ts
│   │   ├── store/                 # Zustand store
│   │   │   └── index.ts
│   │   ├── types/                 # TypeScript типы
│   │   │   └── index.ts
│   │   ├── utils/                 # Утилиты
│   │   │   └── index.ts
│   │   ├── i18n/                  # Интернационализация
│   │   │   └── index.ts
│   │   ├── App.tsx                # Главный компонент
│   │   └── main.tsx               # Точка входа
│   ├── public/                    # Статические файлы
│   ├── dist/                      # Собранное приложение
│   ├── package.json
│   ├── vite.config.ts
│   ├── tailwind.config.js
│   ├── tsconfig.json
│   ├── Dockerfile                 # Frontend Docker
│   └── nginx.conf                 # Nginx конфиг
│
├── k8s/                           # Kubernetes манифесты
│   ├── namespace.yaml
│   ├── deployment.yaml
│   ├── ingress.yaml
│   ├── configmap.yaml
│   ├── secrets.yaml
│   ├── networkpolicy.yaml
│   └── servicemonitor.yaml
│
├── monitoring/                    # Мониторинг
│   ├── prometheus.yml             # Prometheus конфиг
│   ├── alerts.yml                 # Alert правила
│   └── grafana/
│       ├── dashboards/
│       │   ├── dashboard.yml
│       │   └── pr-reviewer.json
│       └── datasources/
│           └── prometheus.yml
│
├── pgbouncer/                     # PgBouncer конфиги
│   ├── docker-compose.pgbouncer.yaml
│   ├── pgbouncer.ini
│   └── userlist.txt
│
├── scripts/                       # Утилиты
│   └── seed_db.sh                 # Заполнение тестовыми данными
│
├── examples/                      # Примеры
│   └── api_examples.http          # HTTP запросы для тестирования
│
├── docs/                          # Документация (удалена, см. README.md и INSTALLATION.md)
│
├── docker-compose.yaml            # Docker Compose (production)
├── docker-compose.dev.yaml        # Docker Compose (development)
├── docker-compose.test.yaml       # Docker Compose (testing)
├── docker-compose.prod.yaml       # Docker Compose (production полный)
├── Dockerfile                     # Backend Dockerfile
├── nginx.prod.conf                # Nginx production конфиг
├── openapi.yaml                   # OpenAPI спецификация
├── Makefile                       # Автоматизация команд
├── Makefile.production            # Production команды
├── quickstart.sh                  # Скрипт быстрого старта
├── go.mod                         # Go зависимости
├── go.sum                         # Go checksums
├── README.md                      # Эта документация
└── INSTALLATION.md                # Детальная инструкция по установке
```

---

## 📚 API Documentation

### REST API Endpoints

Полная спецификация OpenAPI 3.0 доступна в файле `openapi.yaml`.

#### Teams

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/teams` | Получить все команды |
| POST | `/teams` | Создать команду |
| GET | `/teams/{teamId}` | Получить команду по ID |
| DELETE | `/teams/{teamId}` | Удалить команду |
| POST | `/teams/{teamId}/users` | Добавить пользователя в команду |
| DELETE | `/teams/{teamId}/users` | Удалить пользователя из команды |
| POST | `/teams/{teamId}/users/deactivate` | Массовая деактивация |

#### Users

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/users` | Получить всех пользователей |
| POST | `/users` | Создать пользователя |
| GET | `/users/{userId}` | Получить пользователя по ID |
| PATCH | `/users/{userId}` | Обновить пользователя |

#### Pull Requests

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/pull-requests` | Получить все PR (с фильтрацией) |
| POST | `/pull-requests` | Создать PR (+ авто-назначение) |
| GET | `/pull-requests/{prId}` | Получить PR по ID |
| POST | `/pull-requests/{prId}/reviewers` | Добавить рецензента |
| PUT | `/pull-requests/{prId}/reviewers` | Переназначить рецензента |
| POST | `/pull-requests/{prId}/merge` | Merge PR |
| POST | `/pull-requests/{prId}/close` | Close PR |

#### Statistics

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/statistics` | Получить статистику |

#### Health

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/health` | Health check |

### Примеры запросов

См. файл `examples/api_examples.http` для полных примеров всех API запросов.

#### Создание команды

```bash
curl -X POST http://localhost:8080/teams \
  -H "Content-Type: application/json" \
  -d '{"name": "Backend Team"}'
```

#### Создание пользователя

```bash
curl -X POST http://localhost:8080/users \
  -H "Content-Type: application/json" \
  -d '{
    "username": "john_doe",
    "name": "John Doe",
    "teamId": 1
  }'
```

#### Создание PR с автоматическим назначением рецензентов

```bash
curl -X POST http://localhost:8080/pull-requests \
  -H "Content-Type: application/json" \
  -d '{
    "title": "Add new feature",
    "authorId": 1
  }'
```

#### Получение статистики

```bash
curl http://localhost:8080/statistics
```

---

## 🚀 Быстрый старт

### Требования

- Docker 20.10+
- Docker Compose 2.0+
- (Опционально) Go 1.21+ для локальной разработки
- (Опционально) Node.js 18+ для frontend разработки

### Запуск за 2 минуты

```bash
# 1. Клонировать репозиторий
git clone <repository-url>
cd avito_testcase

# 2. Запустить все сервисы через Docker Compose
docker-compose up -d

# 3. Подождать ~30 секунд пока сервисы поднимутся

# 4. Проверить статус
docker-compose ps
curl http://localhost:8080/health

# 5. Открыть приложение
# Frontend: http://localhost:3000
# Backend API: http://localhost:8080
# Grafana: http://localhost:3001 (admin/admin)
# Prometheus: http://localhost:9090
```

### Заполнение тестовыми данными

```bash
# Выполнить скрипт заполнения БД
./scripts/seed_db.sh

# Или использовать Makefile
make seed-db
```

### Проверка работоспособности

```bash
# Health check
curl http://localhost:8080/health

# Получить команды
curl http://localhost:8080/teams

# Получить пользователей
curl http://localhost:8080/users

# Получить статистику
curl http://localhost:8080/statistics
```

---

## 🏭 Production развертывание

### Docker Compose (Production)

```bash
# Использовать production конфиг
docker-compose -f docker-compose.prod.yaml up -d

# Или через Makefile
make docker-up
```

### Kubernetes

```bash
# Применить все манифесты
kubectl apply -f k8s/

# Проверить статус
kubectl get pods -n pr-reviewer
kubectl get svc -n pr-reviewer
kubectl get ingress -n pr-reviewer
```

### Переменные окружения (Production)

Создать файл `.env`:

```bash
# Application
ENVIRONMENT=production
LOG_LEVEL=info
SERVER_PORT=8080

# Database
DATABASE_URL=postgres://postgres:STRONG_PASSWORD@db:5432/pr_reviewer?sslmode=require
POSTGRES_PASSWORD=STRONG_PASSWORD

# Redis
REDIS_ADDR=redis:6379
REDIS_PASSWORD=STRONG_REDIS_PASSWORD
REDIS_DB=0

# JWT
JWT_SECRET=CHANGE_THIS_TO_STRONG_SECRET_KEY_IN_PRODUCTION
JWT_EXPIRATION=24h

# Rate Limiting
RATE_LIMIT_RPS=100
RATE_LIMIT_BURST=200

# Grafana
GRAFANA_USER=admin
GRAFANA_PASSWORD=STRONG_GRAFANA_PASSWORD
```

### Миграции в Production

```bash
# Выполнить миграции вручную
DATABASE_URL="postgres://user:pass@host:5432/db" make migrate

# Или использовать migrate CLI
migrate -path migrations \
  -database "postgres://user:pass@host:5432/db?sslmode=disable" \
  up
```

---

## 📊 Мониторинг и метрики

### Prometheus Metrics

Backend автоматически экспортирует метрики:

- `http_requests_total` - общее количество запросов
- `http_request_duration_seconds` - длительность запросов
- `http_requests_errors_total` - количество ошибок
- `db_connections_active` - активные соединения с БД
- `cache_hits_total` - попадания в кэш
- `cache_misses_total` - промахи кэша

Доступ к метрикам: `http://localhost:8080/metrics`

### Grafana Dashboards

Предустановленные дашборды:

1. **PR Reviewer Overview** - общая информация о сервисе
2. **Database Performance** - производительность БД
3. **API Metrics** - метрики HTTP запросов
4. **System Resources** - CPU, Memory, Disk

Доступ: `http://localhost:3001` (admin/admin)

### Alerts

Настроены алерты Prometheus:

- **High Error Rate** - более 5% ошибок
- **High Response Time** - p95 > 1s
- **Database Connection Issues** - проблемы с БД
- **Service Down** - сервис недоступен

Конфигурация: `monitoring/alerts.yml`

---

## 🧪 Тестирование

### Unit тесты

```bash
# Запустить все unit тесты
make test

# С покрытием
make test-coverage

# Открыть HTML отчет покрытия
open coverage.html
```

### Integration тесты

```bash
# Запустить интеграционные тесты
make test-integration

# Или напрямую через go
go test -tags=integration ./tests/...
```

### Тестовая база данных

```bash
# Поднять тестовое окружение
docker-compose -f docker-compose.test.yaml up -d

# Запустить тесты
DATABASE_URL="postgres://postgres:postgres@localhost:5432/pr_reviewer_test" \
  go test -tags=integration ./tests/...

# Остановить
docker-compose -f docker-compose.test.yaml down
```

### Нагрузочное тестирование

```bash
# Apache Bench
make load-test

# Vegeta
./tests/performance/vegeta_test.sh

# Результаты будут в tests/performance/results/
```

### Примеры тестовых сценариев

```bash
# 1. Создать команду
curl -X POST http://localhost:8080/teams \
  -H "Content-Type: application/json" \
  -d '{"name": "Test Team"}'

# 2. Создать 5 пользователей
for i in {1..5}; do
  curl -X POST http://localhost:8080/users \
    -H "Content-Type: application/json" \
    -d "{\"username\": \"user$i\", \"name\": \"User $i\", \"teamId\": 1}"
done

# 3. Создать 10 PR (автоматически назначатся рецензенты)
for i in {1..10}; do
  curl -X POST http://localhost:8080/pull-requests \
    -H "Content-Type: application/json" \
    -d "{\"title\": \"PR #$i\", \"authorId\": 1}"
done

# 4. Проверить статистику
curl http://localhost:8080/statistics | jq .
```

### Покрытие тестами

| Компонент | Покрытие | Статус |
|-----------|----------|--------|
| service | 85%+ | ✅ |
| handler | 75%+ | ✅ |
| repository | 80%+ | ✅ |
| middleware | 70%+ | ✅ |
| circuitbreaker | 90%+ | ✅ |
| featureflags | 90%+ | ✅ |

---

## 🔧 Разработка

### Локальная разработка Backend

```bash
# Установить зависимости
make deps

# Запустить БД и Redis
docker-compose up -d db redis

# Запустить backend локально
DATABASE_URL="postgres://postgres:postgres@localhost:5432/pr_reviewer" \
  make run

# Или через go run
go run cmd/server/main.go
```

### Локальная разработка Frontend

```bash
cd frontend

# Установить зависимости
npm install

# Запустить dev server
npm run dev

# Открыть http://localhost:5173
```

### Линтинг и форматирование

```bash
# Backend
make lint    # golangci-lint
make fmt     # go fmt + goimports

# Frontend
cd frontend
npm run lint  # ESLint
```

### Hot Reload

Backend:

```bash
# Использовать air для hot reload
go install github.com/cosmtrek/air@latest
air
```

Frontend:

```bash
cd frontend
npm run dev  # Vite уже поддерживает HMR
```

---

## 📖 Дополнительные ресурсы

### Документация

- **INSTALLATION.md** - Детальная инструкция по установке и настройке
- **openapi.yaml** - OpenAPI 3.0 спецификация API
- **examples/api_examples.http** - Примеры всех API запросов

### Полезные команды

```bash
# Makefile help
make help

# Логи всех сервисов
docker-compose logs -f

# Логи конкретного сервиса
docker-compose logs -f backend
docker-compose logs -f frontend

# Перезапустить сервис
docker-compose restart backend

# Подключиться к БД
docker-compose exec db psql -U postgres -d pr_reviewer

# Выполнить SQL
docker-compose exec db psql -U postgres -d pr_reviewer -c "SELECT * FROM users;"

# Очистить все данные
docker-compose down -v

# Пересобрать образы
docker-compose build --no-cache
```

---

## 🤝 Contributing

### Code Style

- **Backend**: Следуем [Effective Go](https://golang.org/doc/effective_go)
- **Frontend**: [Airbnb JavaScript Style Guide](https://github.com/airbnb/javascript)

### Git Workflow

```bash
# Создать feature branch
git checkout -b feature/your-feature

# Сделать изменения
git add .
git commit -m "feat: add new feature"

# Push и создать PR
git push origin feature/your-feature
```

### Commit Convention

Используем [Conventional Commits](https://www.conventionalcommits.org/):

- `feat:` - новая функциональность
- `fix:` - исправление бага
- `docs:` - изменения в документации
- `refactor:` - рефакторинг кода
- `test:` - добавление тестов
- `chore:` - изменения в конфигурации

---

## 📝 License

MIT License

---

## 👥 Authors

Создано как тестовое задание для Avito.

---

## 🆘 Troubleshooting

### Проблема: Backend не запускается

```bash
# Проверить логи
docker-compose logs backend

# Проверить подключение к БД
docker-compose exec db pg_isready -U postgres

# Проверить миграции
docker-compose exec backend ls -la migrations/
```

### Проблема: Frontend не подключается к Backend

```bash
# Проверить CORS
curl -I http://localhost:8080/health

# Проверить переменные окружения frontend
docker-compose exec frontend env | grep VITE
```

### Проблема: База данных не инициализируется

```bash
# Пересоздать volume БД
docker-compose down -v
docker-compose up -d db

# Вручную выполнить миграции
docker-compose exec backend sh
# inside container:
migrate -path migrations -database "$DATABASE_URL" up
```

---

## 📞 Support

Для вопросов и поддержки:
- Issues: [GitHub Issues]
- Email: [support@example.com]

---

**🎉 Happy Coding!**

