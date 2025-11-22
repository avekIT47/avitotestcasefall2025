#!/bin/bash

# Performance testing with vegeta
# Install: go install github.com/tsenart/vegeta@latest

BASE_URL=${BASE_URL:-"http://localhost:8080"}
DURATION=${DURATION:-"30s"}
RATE=${RATE:-"5"}

echo "==================================================="
echo "Performance Testing with Vegeta"
echo "Base URL: $BASE_URL"
echo "Duration: $DURATION"
echo "Rate: $RATE req/sec"
echo "==================================================="

# Создание целей для тестирования
cat > /tmp/vegeta_targets.txt <<EOF
GET ${BASE_URL}/health
GET ${BASE_URL}/teams
GET ${BASE_URL}/users
GET ${BASE_URL}/pull-requests
GET ${BASE_URL}/statistics
EOF

echo ""
echo "Starting performance test..."
echo "Target SLI: 99.9% success rate, 300ms response time"
echo ""

# Запуск теста
vegeta attack -duration=${DURATION} -rate=${RATE} -targets=/tmp/vegeta_targets.txt | \
vegeta report -type=text

# Генерация HTML отчёта
echo ""
echo "Generating detailed report..."
vegeta attack -duration=${DURATION} -rate=${RATE} -targets=/tmp/vegeta_targets.txt | \
vegeta encode | \
vegeta report -type=html > performance_report.html

echo ""
echo "HTML report saved to performance_report.html"

# Генерация графика задержек
echo "Generating latency plot..."
vegeta attack -duration=${DURATION} -rate=${RATE} -targets=/tmp/vegeta_targets.txt | \
vegeta encode | \
vegeta plot > latency_plot.html

echo "Latency plot saved to latency_plot.html"

# Очистка
rm /tmp/vegeta_targets.txt

echo ""
echo "Performance testing complete!"
echo ""
echo "To test specific endpoint with higher load:"
echo "  echo \"GET ${BASE_URL}/health\" | vegeta attack -duration=60s -rate=100 | vegeta report"
echo ""
echo "To test with custom body:"
echo "  echo \"POST ${BASE_URL}/teams\" | vegeta attack -body='{\"name\":\"Test\"}' -header=\"Content-Type: application/json\" -duration=30s -rate=10 | vegeta report"
