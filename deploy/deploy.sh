#!/bin/bash

# 🚀 Dwell Backend Deployment Script for EC2
# This script automates the deployment process

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Function to print colored output
print_status() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

print_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Configuration
APP_NAME="dwell"
CONTAINER_NAME="dwell-api-prod"
HEALTH_CHECK_URL="http://localhost:8080/api/v1/health"
DOCKER_COMPOSE_FILE="docker-compose.prod.yml"

# Check if we're in the right directory
if [ ! -f "docker-compose.prod.yml" ]; then
    print_error "docker-compose.prod.yml not found. Please run this script from the project root."
    exit 1
fi

# Check if .env.prod exists
if [ ! -f ".env.prod" ]; then
    print_error ".env.prod file not found. Please create it from env.example first."
    exit 1
fi

print_status "Starting deployment of $APP_NAME..."

# Function to check if container is healthy
check_health() {
    local max_attempts=30
    local attempt=1
    
    print_status "Waiting for application to be healthy..."
    
    while [ $attempt -le $max_attempts ]; do
        if curl -f -s "$HEALTH_CHECK_URL" > /dev/null 2>&1; then
            print_success "Application is healthy!"
            return 0
        fi
        
        print_status "Attempt $attempt/$max_attempts - Application not ready yet..."
        sleep 10
        attempt=$((attempt + 1))
    done
    
    print_error "Application failed to become healthy after $max_attempts attempts"
    return 1
}

# Function to backup current deployment
backup_deployment() {
    if docker ps -q -f name="$CONTAINER_NAME" | grep -q .; then
        print_status "Creating backup of current deployment..."
        
        # Create backup directory
        mkdir -p backups/$(date +%Y%m%d_%H%M%S)
        
        # Save current image
        docker commit "$CONTAINER_NAME" "dwell:backup-$(date +%Y%m%d_%H%M%S)"
        
        print_success "Backup created successfully"
    fi
}

# Function to rollback deployment
rollback_deployment() {
    print_error "Deployment failed. Rolling back..."
    
    # Find latest backup
    local latest_backup=$(docker images --format "table {{.Repository}}:{{.Tag}}" | grep "dwell:backup" | tail -1 | awk '{print $1}')
    
    if [ -n "$latest_backup" ]; then
        print_status "Rolling back to: $latest_backup"
        
        # Stop current containers
        docker-compose -f "$DOCKER_COMPOSE_FILE" down
        
        # Start with backup image
        docker run -d --name "$CONTAINER_NAME" -p 8080:8080 "$latest_backup"
        
        print_success "Rollback completed"
    else
        print_error "No backup found for rollback"
    fi
}

# Main deployment function
deploy() {
    print_status "Starting deployment process..."
    
    # Backup current deployment
    backup_deployment
    
    # Pull latest code
    print_status "Pulling latest code..."
    git pull origin main
    
    # Build new image
    print_status "Building new Docker image..."
    docker-compose -f "$DOCKER_COMPOSE_FILE" build --no-cache
    
    # Stop existing services
    print_status "Stopping existing services..."
    docker-compose -f "$DOCKER_COMPOSE_FILE" down
    
    # Start with new image
    print_status "Starting new services..."
    docker-compose -f "$DOCKER_COMPOSE_FILE" up -d
    
    # Wait for services to be ready
    print_status "Waiting for services to start..."
    sleep 20
    
    # Health check
    if check_health; then
        print_success "Deployment completed successfully!"
        
        # Show service status
        print_status "Service status:"
        docker-compose -f "$DOCKER_COMPOSE_FILE" ps
        
        # Show logs
        print_status "Recent logs:"
        docker-compose -f "$DOCKER_COMPOSE_FILE" logs --tail=20
        
        return 0
    else
        print_error "Deployment failed health check"
        rollback_deployment
        return 1
    fi
}

# Function to show deployment status
show_status() {
    print_status "Current deployment status:"
    docker-compose -f "$DOCKER_COMPOSE_FILE" ps
    
    print_status "Recent logs:"
    docker-compose -f "$DOCKER_COMPOSE_FILE" logs --tail=10
}

# Function to show logs
show_logs() {
    print_status "Showing logs for all services:"
    docker-compose -f "$DOCKER_COMPOSE_FILE" logs -f
}

# Function to restart services
restart_services() {
    print_status "Restarting services..."
    docker-compose -f "$DOCKER_COMPOSE_FILE" restart
    
    print_success "Services restarted"
    show_status
}

# Function to clean up old images
cleanup() {
    print_status "Cleaning up old Docker images..."
    
    # Remove unused images
    docker image prune -f
    
    # Remove old backup images
    docker images --format "table {{.Repository}}:{{.Tag}}" | grep "dwell:backup" | head -n -3 | awk '{print $1}' | xargs -r docker rmi
    
    print_success "Cleanup completed"
}

# Main script logic
case "${1:-deploy}" in
    "deploy")
        deploy
        ;;
    "status")
        show_status
        ;;
    "logs")
        show_logs
        ;;
    "restart")
        restart_services
        ;;
    "cleanup")
        cleanup
        ;;
    "rollback")
        rollback_deployment
        ;;
    "help"|"-h"|"--help")
        echo "Usage: $0 [command]"
        echo ""
        echo "Commands:"
        echo "  deploy    - Deploy the application (default)"
        echo "  status    - Show deployment status"
        echo "  logs      - Show service logs"
        echo "  restart   - Restart all services"
        echo "  cleanup   - Clean up old Docker images"
        echo "  rollback  - Rollback to previous deployment"
        echo "  help      - Show this help message"
        ;;
    *)
        print_error "Unknown command: $1"
        echo "Use '$0 help' for usage information"
        exit 1
        ;;
esac




