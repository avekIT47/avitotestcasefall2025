#!/bin/bash

# Load Testing Script for PR Reviewer Service
# Использует Apache Bench (ab) для нагрузочного тестирования

echo "==================================================="
echo "PR Reviewer Service - Load Testing"
echo "==================================================="

BASE_URL="http://localhost:8080"

# Проверка доступности сервиса
echo "Checking service availability..."
curl -f -s "$BASE_URL/health" > /dev/null
if [ $? -ne 0 ]; then
    echo "Service is not available at $BASE_URL"
    exit 1
fi
echo "Service is available"
echo ""

# Создание тестовых данных
echo "Creating test data..."

# Создание команды
TEAM_ID=$(curl -s -X POST "$BASE_URL/teams" \
    -H "Content-Type: application/json" \
    -d '{"name":"Load Test Team"}' | jq -r '.id')

echo "Created team with ID: $TEAM_ID"

# Создание пользователей
USER_IDS=()
for i in {1..10}; do
    USER_ID=$(curl -s -X POST "$BASE_URL/users" \
        -H "Content-Type: application/json" \
        -d "{\"name\":\"Load Test User $i\",\"teamId\":$TEAM_ID}" | jq -r '.id')
    USER_IDS+=($USER_ID)
done
echo "Created ${#USER_IDS[@]} users"

# Создание PR для тестирования
PR_ID=$(curl -s -X POST "$BASE_URL/pull-requests" \
    -H "Content-Type: application/json" \
    -d "{\"title\":\"Load Test PR\",\"authorId\":${USER_IDS[0]}}" | jq -r '.id')
echo "Created PR with ID: $PR_ID"
echo ""

echo "==================================================="
echo "Starting Load Tests"
echo "==================================================="
echo ""

# Тест 1: Health Check Endpoint
echo "Test 1: Health Check Endpoint"
echo "Concurrent requests: 10, Total requests: 100"
ab -n 100 -c 10 "$BASE_URL/health" 2>/dev/null | grep -E "Requests per second|Time per request|Failed requests|Non-2xx responses"
echo ""

# Тест 2: Get Teams (чтение)
echo "Test 2: Get Teams Endpoint (Read)"
echo "Concurrent requests: 5, Total requests: 50"
ab -n 50 -c 5 "$BASE_URL/teams" 2>/dev/null | grep -E "Requests per second|Time per request|Failed requests|Non-2xx responses"
echo ""

# Тест 3: Get Users с фильтрацией
echo "Test 3: Get Users with Filter (Read)"
echo "Concurrent requests: 5, Total requests: 50"
ab -n 50 -c 5 "$BASE_URL/users?teamId=$TEAM_ID" 2>/dev/null | grep -E "Requests per second|Time per request|Failed requests|Non-2xx responses"
echo ""

# Тест 4: Create PR (запись)
echo "Test 4: Create Pull Request (Write)"
echo "Concurrent requests: 2, Total requests: 20"

# Создание временного файла с данными для POST запроса
echo "{\"title\":\"Performance Test PR\",\"authorId\":${USER_IDS[1]}}" > /tmp/pr_data.json

ab -n 20 -c 2 -p /tmp/pr_data.json -T "application/json" "$BASE_URL/pull-requests" 2>/dev/null | grep -E "Requests per second|Time per request|Failed requests|Non-2xx responses"
rm /tmp/pr_data.json
echo ""

# Тест 5: Get Statistics
echo "Test 5: Get Statistics (Aggregation)"
echo "Concurrent requests: 5, Total requests: 50"
ab -n 50 -c 5 "$BASE_URL/statistics" 2>/dev/null | grep -E "Requests per second|Time per request|Failed requests|Non-2xx responses"
echo ""

# Тест 6: Смешанная нагрузка
echo "Test 6: Mixed Load Test"
echo "Running parallel requests to different endpoints..."

# Запуск параллельных тестов
(ab -n 30 -c 3 "$BASE_URL/health" 2>/dev/null | grep "Requests per second" | sed 's/^/  Health: /') &
(ab -n 30 -c 3 "$BASE_URL/teams" 2>/dev/null | grep "Requests per second" | sed 's/^/  Teams: /') &
(ab -n 30 -c 3 "$BASE_URL/users" 2>/dev/null | grep "Requests per second" | sed 's/^/  Users: /') &
(ab -n 30 -c 3 "$BASE_URL/pull-requests" 2>/dev/null | grep "Requests per second" | sed 's/^/  PRs: /') &

# Ждём завершения всех фоновых процессов
wait
echo ""

echo "==================================================="
echo "Load Testing Complete"
echo "==================================================="
echo ""
echo "Summary:"
echo "- Target RPS: 5"
echo "- Target response time: < 300ms"
echo "- Target success rate: 99.9%"
echo ""
echo "Note: For more detailed results, use specialized tools like:"
echo "- wrk: wrk -t12 -c400 -d30s --latency $BASE_URL/health"
echo "- vegeta: echo \"GET $BASE_URL/health\" | vegeta attack -duration=30s -rate=5 | vegeta report"
