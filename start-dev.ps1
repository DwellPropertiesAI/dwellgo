# ========================================
# DWELL PROPERTY MANAGEMENT MVP
# Development Environment Startup Script (Windows PowerShell)
# ========================================

Write-Host "🚀 Starting Dwell Property Management Development Environment..." -ForegroundColor Green

# Check if Docker is running
try {
    docker info | Out-Null
} catch {
    Write-Host "❌ Docker is not running. Please start Docker Desktop and try again." -ForegroundColor Red
    exit 1
}

# Check if .env file exists
if (-not (Test-Path ".env")) {
    Write-Host "⚠️  No .env file found. Creating one from template..." -ForegroundColor Yellow
    if (Test-Path "env.example") {
        Copy-Item "env.example" ".env"
        Write-Host "✅ Created .env from env.example" -ForegroundColor Green
        Write-Host "📝 Please edit .env with your actual AWS credentials and configuration" -ForegroundColor Yellow
    } else {
        Write-Host "❌ No env.example found. Please create a .env file manually." -ForegroundColor Red
        exit 1
    }
}

# Create necessary directories
Write-Host "📁 Creating necessary directories..." -ForegroundColor Blue
New-Item -ItemType Directory -Force -Path "logs" | Out-Null
New-Item -ItemType Directory -Force -Path "deploy/nginx/logs" | Out-Null

# Start services
Write-Host "🐳 Starting Docker Compose services..." -ForegroundColor Blue
docker-compose up -d

# Wait for services to be healthy
Write-Host "⏳ Waiting for services to be healthy..." -ForegroundColor Yellow
Start-Sleep -Seconds 10

# Check service status
Write-Host "🔍 Checking service status..." -ForegroundColor Blue
docker-compose ps

Write-Host ""
Write-Host "✅ Development environment is starting up!" -ForegroundColor Green
Write-Host ""
Write-Host "📊 Service URLs:" -ForegroundColor Cyan
Write-Host "   - API: http://localhost:5000" -ForegroundColor White
Write-Host "   - pgAdmin: http://localhost:5050 (admin@dwell.com / admin)" -ForegroundColor White
Write-Host "   - PostgreSQL: localhost:5433" -ForegroundColor White
Write-Host "   - Redis: localhost:6380" -ForegroundColor White
Write-Host ""
Write-Host "📝 To view logs: docker-compose logs -f [service-name]" -ForegroundColor Yellow
Write-Host "🛑 To stop: docker-compose down" -ForegroundColor Yellow
Write-Host "🔄 To restart: docker-compose restart" -ForegroundColor Yellow
Write-Host ""
Write-Host "🎯 Happy coding!" -ForegroundColor Green
