# Docker команды для PR Reviewer Service

## Основные команды

### Запуск сервиса
```bash
# Быстрый старт
./quickstart.sh

# Или вручную
docker-compose up -d
```

### Просмотр логов
```bash
# Все логи
docker-compose logs

# Логи в реальном времени
docker-compose logs -f

# Логи конкретного сервиса
docker-compose logs app
docker-compose logs db
```

### Остановка сервиса
```bash
# Остановить контейнеры
docker-compose stop

# Остановить и удалить контейнеры
docker-compose down

# Остановить и удалить контейнеры + volumes
docker-compose down -v
```

### Перезапуск
```bash
# Перезапустить все сервисы
docker-compose restart

# Перезапустить конкретный сервис
docker-compose restart app
```

## Работа с базой данных

### Подключение к PostgreSQL
```bash
# Подключиться к БД в контейнере
docker-compose exec db psql -U postgres -d pr_reviewer

# Выполнить SQL команду
docker-compose exec db psql -U postgres -d pr_reviewer -c "SELECT * FROM teams;"
```

### Backup и восстановление
```bash
# Создать backup
docker-compose exec db pg_dump -U postgres pr_reviewer > backup.sql

# Восстановить из backup
docker-compose exec db psql -U postgres pr_reviewer < backup.sql
```

## Отладка

### Войти в контейнер
```bash
# Войти в контейнер приложения
docker-compose exec app sh

# Войти в контейнер БД
docker-compose exec db bash
```

### Проверка состояния
```bash
# Статус контейнеров
docker-compose ps

# Использование ресурсов
docker stats

# Проверка сети
docker network ls
docker network inspect avito_testcase_default
```

## Сборка и обновление

### Пересборка образов
```bash
# Пересобрать без кеша
docker-compose build --no-cache

# Пересобрать и запустить
docker-compose up -d --build
```

### Обновление зависимостей
```bash
# Обновить только зависимости Go
docker-compose exec app go mod download
docker-compose exec app go mod tidy
```

## Тестирование

### Запуск тестов в контейнере
```bash
# Unit тесты
docker-compose exec app go test ./...

# Интеграционные тесты
docker-compose -f docker-compose.test.yaml up --abort-on-container-exit
```

### Проверка API
```bash
# Health check
curl http://localhost:8080/health

# Создать команду
curl -X POST http://localhost:8080/teams \
  -H "Content-Type: application/json" \
  -d '{"name":"Docker Test Team"}'

# Получить статистику
curl http://localhost:8080/statistics | jq
```

## Очистка

### Удаление неиспользуемых ресурсов
```bash
# Удалить остановленные контейнеры
docker container prune

# Удалить неиспользуемые образы
docker image prune

# Удалить всё неиспользуемое
docker system prune -a

# Удалить volumes
docker volume prune
```

## Production-подобное окружение

### Запуск с ограничениями ресурсов
```bash
# Создать docker-compose.prod.yaml с лимитами
docker-compose -f docker-compose.yaml -f docker-compose.prod.yaml up -d
```

### Мониторинг
```bash
# Запустить с Prometheus и Grafana
docker-compose -f docker-compose.yaml -f docker-compose.monitoring.yaml up -d
```

## Полезные алиасы

Добавьте в `.bashrc` или `.zshrc`:
```bash
alias dc='docker-compose'
alias dcup='docker-compose up -d'
alias dcdown='docker-compose down'
alias dclogs='docker-compose logs -f'
alias dcrestart='docker-compose restart'
alias dcexec='docker-compose exec'
```

## Troubleshooting

### Порт занят
```bash
# Проверить, что занимает порт 8080
lsof -i :8080
# или
netstat -tulpn | grep 8080
```

### Проблемы с правами
```bash
# Дать права на выполнение скриптов
chmod +x quickstart.sh
chmod +x tests/load_test.sh
chmod +x scripts/seed_db.sh
```

### Контейнер постоянно перезапускается
```bash
# Проверить логи
docker-compose logs app --tail 50

# Проверить exit code
docker-compose ps
```
