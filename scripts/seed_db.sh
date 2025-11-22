#!/bin/bash

# Скрипт для заполнения БД тестовыми данными
# Используется для демонстрации и тестирования

BASE_URL=${BASE_URL:-"http://localhost:8080"}

echo "==================================================="
echo "Seeding database with test data"
echo "Base URL: $BASE_URL"
echo "==================================================="

# Проверка доступности сервиса
echo -n "Checking service availability... "
curl -f -s "$BASE_URL/health" > /dev/null
if [ $? -ne 0 ]; then
    echo "FAILED"
    echo "Service is not available at $BASE_URL"
    echo "Please ensure the service is running (make docker-up)"
    exit 1
fi
echo "OK"

# Создание команд
echo ""
echo "Creating teams..."

TEAMS=("Backend" "Frontend" "DevOps" "QA" "Mobile")
TEAM_IDS=()

for team in "${TEAMS[@]}"; do
    echo -n "  Creating team '$team'... "
    RESPONSE=$(curl -s -X POST "$BASE_URL/teams" \
        -H "Content-Type: application/json" \
        -d "{\"name\":\"$team Team\"}" 2>/dev/null)
    
    if [ $? -eq 0 ]; then
        TEAM_ID=$(echo $RESPONSE | grep -o '"id":[0-9]*' | grep -o '[0-9]*')
        if [ ! -z "$TEAM_ID" ]; then
            TEAM_IDS+=($TEAM_ID)
            echo "OK (ID: $TEAM_ID)"
        else
            echo "FAILED (possibly already exists)"
        fi
    else
        echo "FAILED"
    fi
done

# Создание пользователей
echo ""
echo "Creating users..."

FIRST_NAMES=("Алексей" "Мария" "Дмитрий" "Анна" "Сергей" "Елена" "Иван" "Ольга" "Андрей" "Наталья")
LAST_NAMES=("Иванов" "Петров" "Сидоров" "Смирнов" "Кузнецов" "Попов" "Васильев" "Павлов" "Семёнов" "Голубев")

USER_IDS=()
USER_COUNT=0

for team_id in "${TEAM_IDS[@]}"; do
    echo "  Creating users for team ID $team_id:"
    
    # Создаём 3-5 пользователей для каждой команды
    NUM_USERS=$((3 + RANDOM % 3))
    
    for i in $(seq 1 $NUM_USERS); do
        FIRST_NAME=${FIRST_NAMES[$RANDOM % ${#FIRST_NAMES[@]}]}
        LAST_NAME=${LAST_NAMES[$RANDOM % ${#LAST_NAMES[@]}]}
        FULL_NAME="$FIRST_NAME $LAST_NAME"
        
        echo -n "    Creating user '$FULL_NAME'... "
        
        RESPONSE=$(curl -s -X POST "$BASE_URL/users" \
            -H "Content-Type: application/json" \
            -d "{\"name\":\"$FULL_NAME\",\"teamId\":$team_id}" 2>/dev/null)
        
        if [ $? -eq 0 ]; then
            USER_ID=$(echo $RESPONSE | grep -o '"id":[0-9]*' | grep -o '[0-9]*')
            if [ ! -z "$USER_ID" ]; then
                USER_IDS+=($USER_ID)
                USER_COUNT=$((USER_COUNT + 1))
                echo "OK (ID: $USER_ID)"
            else
                echo "FAILED"
            fi
        else
            echo "FAILED"
        fi
    done
done

# Создание Pull Requests
echo ""
echo "Creating Pull Requests..."

PR_TITLES=(
    "Добавить новый эндпоинт для отчётов"
    "Исправить баг в авторизации"
    "Обновить зависимости проекта"
    "Рефакторинг модуля платежей"
    "Добавить unit тесты для сервиса"
    "Оптимизация запросов к БД"
    "Новая фича: экспорт в Excel"
    "Исправить проблему с кешированием"
    "Обновить документацию API"
    "Миграция на новую версию фреймворка"
)

PR_COUNT=0
for i in $(seq 1 15); do
    if [ ${#USER_IDS[@]} -eq 0 ]; then
        echo "No users available, skipping PR creation"
        break
    fi
    
    # Выбираем случайного автора
    AUTHOR_ID=${USER_IDS[$RANDOM % ${#USER_IDS[@]}]}
    TITLE=${PR_TITLES[$RANDOM % ${#PR_TITLES[@]}]}
    
    echo -n "  Creating PR '$TITLE' by user $AUTHOR_ID... "
    
    RESPONSE=$(curl -s -X POST "$BASE_URL/pull-requests" \
        -H "Content-Type: application/json" \
        -d "{\"title\":\"$TITLE\",\"authorId\":$AUTHOR_ID}" 2>/dev/null)
    
    if [ $? -eq 0 ]; then
        PR_ID=$(echo $RESPONSE | grep -o '"id":[0-9]*' | grep -o '[0-9]*')
        if [ ! -z "$PR_ID" ]; then
            PR_COUNT=$((PR_COUNT + 1))
            echo "OK (ID: $PR_ID)"
            
            # Случайно мержим некоторые PR (30% вероятность)
            if [ $((RANDOM % 100)) -lt 30 ]; then
                echo -n "    Merging PR $PR_ID... "
                curl -s -X POST "$BASE_URL/pull-requests/$PR_ID/merge" > /dev/null 2>&1
                if [ $? -eq 0 ]; then
                    echo "OK"
                else
                    echo "FAILED"
                fi
            fi
        else
            echo "FAILED"
        fi
    else
        echo "FAILED"
    fi
done

# Получение статистики
echo ""
echo "==================================================="
echo "Fetching statistics..."
echo "==================================================="

STATS=$(curl -s "$BASE_URL/statistics" 2>/dev/null)
if [ $? -eq 0 ] && [ ! -z "$STATS" ]; then
    echo "$STATS" | python3 -m json.tool 2>/dev/null || echo "$STATS"
else
    echo "Failed to fetch statistics"
fi

echo ""
echo "==================================================="
echo "Database seeding complete!"
echo "Summary:"
echo "  - Teams created: ${#TEAM_IDS[@]}"
echo "  - Users created: $USER_COUNT"
echo "  - Pull Requests created: $PR_COUNT"
echo "==================================================="
