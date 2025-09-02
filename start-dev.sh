#!/bin/bash

# ========================================
# DWELL PROPERTY MANAGEMENT MVP
# Development Environment Startup Script
# ========================================

echo "🚀 Starting Dwell Property Management Development Environment..."

# Check if Docker is running
if ! docker info > /dev/null 2>&1; then
    echo "❌ Docker is not running. Please start Docker and try again."
    exit 1
fi

# Check if .env file exists
if [ ! -f .env ]; then
    echo "⚠️  No .env file found. Creating one from template..."
    if [ -f env.example ]; then
        cp env.example .env
        echo "✅ Created .env from env.example"
        echo "📝 Please edit .env with your actual AWS credentials and configuration"
    else
        echo "❌ No env.example found. Please create a .env file manually."
        exit 1
    fi
fi

# Create necessary directories
echo "📁 Creating necessary directories..."
mkdir -p logs
mkdir -p deploy/nginx/logs

# Start services
echo "🐳 Starting Docker Compose services..."
docker-compose up -d

# Wait for services to be healthy
echo "⏳ Waiting for services to be healthy..."
sleep 10

# Check service status
echo "🔍 Checking service status..."
docker-compose ps

echo ""
echo "✅ Development environment is starting up!"
echo ""
echo "📊 Service URLs:"
echo "   - API: http://localhost:5000"
echo "   - pgAdmin: http://localhost:5050 (admin@dwell.com / admin)"
echo "   - PostgreSQL: localhost:5433"
echo "   - Redis: localhost:6380"
echo ""
echo "📝 To view logs: docker-compose logs -f [service-name]"
echo "🛑 To stop: docker-compose down"
echo "🔄 To restart: docker-compose restart"
echo ""
echo "🎯 Happy coding!"
