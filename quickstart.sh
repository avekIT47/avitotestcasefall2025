#!/bin/bash

# Quick start script for PR Reviewer Service

set -e

echo "==================================================="
echo "  PR Reviewer Service - Quick Start"
echo "==================================================="
echo ""

# Проверка Docker
echo -n "Checking Docker... "
if ! command -v docker &> /dev/null; then
    echo "FAILED"
    echo "Docker is not installed. Please install Docker first."
    exit 1
fi
echo "OK"

# Проверка Docker Compose
echo -n "Checking Docker Compose... "
if ! command -v docker-compose &> /dev/null; then
    echo "FAILED"
    echo "Docker Compose is not installed. Please install Docker Compose first."
    exit 1
fi
echo "OK"

echo ""
echo "Starting services..."
echo ""

# Остановка существующих контейнеров
docker-compose down 2>/dev/null || true

# Сборка и запуск
docker-compose up -d --build

echo ""
echo "Waiting for services to be ready..."

# Ждём готовности сервиса
MAX_ATTEMPTS=30
ATTEMPT=0

while [ $ATTEMPT -lt $MAX_ATTEMPTS ]; do
    if curl -f -s http://localhost:8080/health > /dev/null 2>&1; then
        echo ""
        echo "✅ Service is ready!"
        break
    fi
    echo -n "."
    sleep 2
    ATTEMPT=$((ATTEMPT + 1))
done

if [ $ATTEMPT -eq $MAX_ATTEMPTS ]; then
    echo ""
    echo "❌ Service failed to start. Check logs with: docker-compose logs"
    exit 1
fi

echo ""
echo "==================================================="
echo "  Service is running!"
echo "==================================================="
echo ""
echo "📍 API URL: http://localhost:8080"
echo "📍 Health: http://localhost:8080/health"
echo ""
echo "📚 Available commands:"
echo "  make help              - Show all available commands"
echo "  make docker-logs       - View logs"
echo "  make docker-down       - Stop services"
echo "  docker-compose logs -f - Follow logs"
echo ""
echo "📝 API Examples:"
echo "  See examples/api_examples.http"
echo ""
echo "🧪 Run tests:"
echo "  make test              - Unit tests"
echo "  make load-test         - Load testing"
echo ""
echo "🎯 Quick test - Creating a team:"
curl -s -X POST http://localhost:8080/teams \
    -H "Content-Type: application/json" \
    -d '{"name":"Quick Test Team"}' | python3 -m json.tool 2>/dev/null || echo "Created team (install python3 for pretty JSON)"

echo ""
echo "==================================================="
echo "  Happy coding! 🚀"
echo "===================================================">
