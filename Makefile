# Dwell Property Management API Makefile

.PHONY: help build run test clean docker-build docker-run docker-stop lint format swagger aws-setup aws-test

# Default target
help:
	@echo "Dwell Property Management API - Available Commands:"
	@echo ""
	@echo "Development:"
	@echo "  build          - Build the application"
	@echo "  run            - Run the application locally"
	@echo "  test           - Run tests"
	@echo "  test-coverage  - Run tests with coverage report"
	@echo "  clean          - Clean build artifacts"
	@echo ""
	@echo "Docker:"
	@echo "  docker-build   - Build Docker image"
	@echo "  docker-run     - Run with Docker Compose"
	@echo "  docker-stop    - Stop Docker Compose services"
	@echo "  docker-logs    - View Docker logs"
	@echo ""
	@echo "Code Quality:"
	@echo "  lint           - Run linter"
	@echo "  format         - Format Go code"
	@echo "  swagger        - Generate Swagger documentation"
	@echo ""
	@echo "Database:"
	@echo "  db-migrate     - Run database migrations"
	@echo "  db-seed        - Seed database with sample data"
	@echo ""
	@echo "Dependencies:"
	@echo "  deps           - Download Go dependencies"
	@echo "  deps-update    - Update Go dependencies"
	@echo ""
	@echo "AWS Setup:"
	@echo "  aws-setup      - Run AWS setup script"
	@echo "  aws-test       - Test AWS services connectivity"
	@echo ""
	@echo "Use 'make <command>' to run a specific command"

# Development commands
build:
	@echo "Building Dwell API..."
	@go build -o bin/dwell main.go
	@echo "Build complete! Binary available at bin/dwell"

run:
	@echo "Starting Dwell API..."
	@go run main.go

test:
	@echo "Running tests..."
	@go test ./...

test-coverage:
	@echo "Running tests with coverage..."
	@go test -coverprofile=coverage.out ./...
	@go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

clean:
	@echo "Cleaning build artifacts..."
	@rm -rf bin/
	@rm -f coverage.out coverage.html
	@go clean

# Docker commands
docker-build:
	@echo "Building Docker image..."
	@docker build -t dwell:latest .

docker-run:
	@echo "Starting services with Docker Compose..."
	@docker-compose up -d

docker-stop:
	@echo "Stopping Docker Compose services..."
	@docker-compose down

docker-logs:
	@echo "Viewing Docker logs..."
	@docker-compose logs -f

docker-clean:
	@echo "Cleaning Docker resources..."
	@docker-compose down -v
	@docker system prune -f

# AWS Setup commands
aws-setup:
	@echo "Setting up AWS integration..."
	@if [ "$(OS)" = "Windows_NT" ]; then \
		powershell -ExecutionPolicy Bypass -File setup_aws.ps1; \
	else \
		chmod +x setup_aws.sh && ./setup_aws.sh; \
	fi

aws-test:
	@echo "Testing AWS services connectivity..."
	@if [ "$(OS)" = "Windows_NT" ]; then \
		powershell -ExecutionPolicy Bypass -File setup_aws.ps1 -SkipTests; \
	else \
		chmod +x setup_aws.sh && ./setup_aws.sh; \
	fi

# Code quality commands
lint:
	@echo "Running linter..."
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run; \
	else \
		echo "golangci-lint not found. Install with: go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest"; \
	fi

format:
	@echo "Formatting Go code..."
	@go fmt ./...
	@if command -v goimports >/dev/null 2>&1; then \
		find . -name "*.go" -not -path "./vendor/*" -exec goimports -w {} \;; \
	else \
		echo "goimports not found. Install with: go install golang.org/x/tools/cmd/goimports@latest"; \
	fi

swagger:
	@echo "Generating Swagger documentation..."
	@if command -v swag >/dev/null 2>&1; then \
		swag init -g main.go -o docs; \
	else \
		echo "swag not found. Install with: go install github.com/swaggo/swag/cmd/swag@latest"; \
	fi

# Database commands
db-migrate:
	@echo "Running database migrations..."
	@echo "Note: Migrations are automatically run when starting the application"

db-seed:
	@echo "Seeding database with sample data..."
	@echo "Note: This feature is not yet implemented"

# Dependency management
deps:
	@echo "Downloading Go dependencies..."
	@go mod download

deps-update:
	@echo "Updating Go dependencies..."
	@go get -u ./...
	@go mod tidy

# Development setup
setup:
	@echo "Setting up development environment..."
	@make deps
	@make docker-build
	@echo "Development environment setup complete!"

# Production build
build-prod:
	@echo "Building production binary..."
	@CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o bin/dwell-prod main.go
	@echo "Production binary available at bin/dwell-prod"

# Health check
health:
	@echo "Checking application health..."
	@curl -f http://localhost:8080/api/v1/health || echo "Application is not responding"

# Install development tools
install-tools:
	@echo "Installing development tools..."
	@go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	@go install golang.org/x/tools/cmd/goimports@latest
	@go install github.com/swaggo/swag/cmd/swag@latest
	@echo "Development tools installed!"

# Quick development workflow
dev: deps format lint test run

# Production deployment check
prod-check: build-prod test lint
	@echo "Production build check complete!"

