#!/bin/bash

# Smart Redirect Deployment Script
# Usage: ./scripts/deploy.sh [environment] [action]
# Environment: dev, prod
# Action: start, stop, restart, logs, status

set -e

ENVIRONMENT=${1:-dev}
ACTION=${2:-start}
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_DIR="$(cd "$SCRIPT_DIR/.." && pwd)"

# Color codes for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

log() {
    echo -e "${BLUE}[$(date +'%Y-%m-%d %H:%M:%S')]${NC} $1"
}

success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Check if docker and docker-compose are installed
check_dependencies() {
    log "Checking dependencies..."
    
    if ! command -v docker &> /dev/null; then
        error "Docker is not installed. Please install Docker first."
        exit 1
    fi
    
    if ! command -v docker-compose &> /dev/null; then
        error "Docker Compose is not installed. Please install Docker Compose first."
        exit 1
    fi
    
    success "Dependencies check passed"
}

# Validate environment
validate_environment() {
    if [[ "$ENVIRONMENT" != "dev" && "$ENVIRONMENT" != "prod" ]]; then
        error "Invalid environment: $ENVIRONMENT. Use 'dev' or 'prod'"
        exit 1
    fi
}

# Create necessary directories
create_directories() {
    log "Creating necessary directories..."
    
    mkdir -p "$PROJECT_DIR/logs"
    mkdir -p "$PROJECT_DIR/data/postgres"
    mkdir -p "$PROJECT_DIR/data/redis"
    mkdir -p "$PROJECT_DIR/ssl"
    
    success "Directories created"
}

# Generate SSL certificates for development
generate_dev_ssl() {
    if [[ "$ENVIRONMENT" == "dev" ]]; then
        log "Generating development SSL certificates..."
        
        SSL_DIR="$PROJECT_DIR/ssl"
        
        if [[ ! -f "$SSL_DIR/cert.pem" || ! -f "$SSL_DIR/key.pem" ]]; then
            openssl req -x509 -newkey rsa:4096 -keyout "$SSL_DIR/key.pem" -out "$SSL_DIR/cert.pem" \
                -days 365 -nodes -subj "/C=US/ST=State/L=City/O=Organization/CN=localhost" \
                2>/dev/null || true
            
            success "Development SSL certificates generated"
        else
            log "SSL certificates already exist"
        fi
    fi
}

# Check environment file
check_env_file() {
    ENV_FILE="$PROJECT_DIR/.env.$ENVIRONMENT"
    
    if [[ ! -f "$ENV_FILE" ]]; then
        warning "Environment file $ENV_FILE not found. Creating template..."
        
        cat > "$ENV_FILE" << EOF
# Database Configuration
DB_HOST=postgres
DB_PORT=5432
DB_NAME=smart_redirect
DB_USER=postgres
DB_PASSWORD=your_secure_password_here

# Redis Configuration
REDIS_HOST=redis
REDIS_PORT=6379
REDIS_PASSWORD=your_redis_password_here

# JWT Configuration
JWT_SECRET=your_jwt_secret_here_change_this_in_production

# Server Configuration
SERVER_PORT=8080
SERVER_HOST=0.0.0.0

# Rate Limiting
RATE_LIMIT_REQUESTS=100
RATE_LIMIT_WINDOW=3600

# Logging
LOG_LEVEL=info
LOG_FORMAT=json
EOF
        
        warning "Please update $ENV_FILE with your actual configuration values"
    fi
}

# Start services
start_services() {
    log "Starting services in $ENVIRONMENT environment..."
    
    cd "$PROJECT_DIR"
    
    if [[ "$ENVIRONMENT" == "prod" ]]; then
        docker-compose -f docker-compose.prod.yml --env-file .env.prod up -d
    else
        docker-compose -f docker-compose.dev.yml --env-file .env.dev up -d
    fi
    
    success "Services started successfully"
    
    # Wait for services to be ready
    log "Waiting for services to be ready..."
    sleep 10
    
    # Check service health
    check_service_health
}

# Stop services
stop_services() {
    log "Stopping services..."
    
    cd "$PROJECT_DIR"
    
    if [[ "$ENVIRONMENT" == "prod" ]]; then
        docker-compose -f docker-compose.prod.yml down
    else
        docker-compose -f docker-compose.dev.yml down
    fi
    
    success "Services stopped successfully"
}

# Restart services
restart_services() {
    log "Restarting services..."
    stop_services
    start_services
}

# Show logs
show_logs() {
    log "Showing logs for $ENVIRONMENT environment..."
    
    cd "$PROJECT_DIR"
    
    if [[ "$ENVIRONMENT" == "prod" ]]; then
        docker-compose -f docker-compose.prod.yml logs -f
    else
        docker-compose -f docker-compose.dev.yml logs -f
    fi
}

# Check service status
check_status() {
    log "Checking service status..."
    
    cd "$PROJECT_DIR"
    
    if [[ "$ENVIRONMENT" == "prod" ]]; then
        docker-compose -f docker-compose.prod.yml ps
    else
        docker-compose -f docker-compose.dev.yml ps
    fi
}

# Check service health
check_service_health() {
    log "Checking service health..."
    
    # Check if backend is responding
    if curl -s -f http://localhost:8080/health > /dev/null 2>&1; then
        success "Backend service is healthy"
    else
        warning "Backend service is not responding"
    fi
    
    # Check if frontend is accessible (only in dev)
    if [[ "$ENVIRONMENT" == "dev" ]]; then
        if curl -s -f http://localhost:3000 > /dev/null 2>&1; then
            success "Frontend service is healthy"
        else
            warning "Frontend service is not responding"
        fi
    fi
    
    # Check database connection
    if docker-compose -f docker-compose.$ENVIRONMENT.yml exec -T postgres pg_isready -U postgres > /dev/null 2>&1; then
        success "Database is ready"
    else
        warning "Database is not ready"
    fi
    
    # Check Redis connection
    if docker-compose -f docker-compose.$ENVIRONMENT.yml exec -T redis redis-cli ping > /dev/null 2>&1; then
        success "Redis is ready"
    else
        warning "Redis is not ready"
    fi
}

# Initialize database
init_database() {
    log "Initializing database..."
    
    cd "$PROJECT_DIR"
    
    # Wait for database to be ready
    log "Waiting for database to be ready..."
    sleep 5
    
    # Run database initialization
    if [[ -f "scripts/init_db.go" ]]; then
        docker-compose -f docker-compose.$ENVIRONMENT.yml exec smart_redirect go run scripts/init_db.go
        success "Database initialized"
    else
        warning "Database initialization script not found"
    fi
}

# Main execution
main() {
    log "Smart Redirect Deployment Script"
    log "Environment: $ENVIRONMENT"
    log "Action: $ACTION"
    
    validate_environment
    check_dependencies
    create_directories
    check_env_file
    
    case $ACTION in
        "start")
            generate_dev_ssl
            start_services
            ;;
        "stop")
            stop_services
            ;;
        "restart")
            restart_services
            ;;
        "logs")
            show_logs
            ;;
        "status")
            check_status
            ;;
        "health")
            check_service_health
            ;;
        "init-db")
            init_database
            ;;
        *)
            error "Invalid action: $ACTION"
            echo "Usage: $0 [environment] [action]"
            echo "Environment: dev, prod"
            echo "Action: start, stop, restart, logs, status, health, init-db"
            exit 1
            ;;
    esac
}

# Run main function
main "$@"