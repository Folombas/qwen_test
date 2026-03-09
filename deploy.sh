#!/bin/bash
# Go Quiz — Deploy Script
# Автоматический деплой на Ubuntu сервер

set -e

# Configuration
APP_NAME="goquiz"
REMOTE_USER="${DEPLOY_USER:-root}"
REMOTE_HOST="${DEPLOY_HOST:-}"
REMOTE_DIR="/opt/${APP_NAME}"

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Functions
log_info() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

log_warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

check_prerequisites() {
    log_info "Checking prerequisites..."
    
    if [ -z "$REMOTE_HOST" ]; then
        log_error "DEPLOY_HOST environment variable is not set"
        exit 1
    fi
    
    if ! command -v docker &> /dev/null; then
        log_error "Docker is not installed"
        exit 1
    fi
    
    if ! command -v docker-compose &> /dev/null; then
        log_error "Docker Compose is not installed"
        exit 1
    fi
    
    log_info "Prerequisites check passed"
}

build() {
    log_info "Building Docker image..."
    docker-compose build
    log_info "Build completed"
}

deploy() {
    log_info "Deploying to $REMOTE_HOST..."
    
    # Create remote directory
    ssh ${REMOTE_USER}@${REMOTE_HOST} "mkdir -p ${REMOTE_DIR}"
    
    # Copy files
    log_info "Copying files..."
    scp -r docker-compose.yml Dockerfile .env.example nginx/ ${REMOTE_USER}@${REMOTE_HOST}:${REMOTE_DIR}/
    
    # Create .env if not exists
    ssh ${REMOTE_USER}@${REMOTE_HOST} "cd ${REMOTE_DIR} && [ -f .env ] || cp .env.example .env"
    
    # Deploy on remote server
    ssh ${REMOTE_USER}@${REMOTE_HOST} << EOF
        cd ${REMOTE_DIR}
        
        log_info "Pulling latest changes..."
        docker-compose pull
        
        log_info "Starting services..."
        docker-compose up -d
        
        log_info "Cleaning up old images..."
        docker image prune -f
        
        log_info "Deployment completed!"
EOF
    
    log_info "Deployment to $REMOTE_HOST completed successfully"
}

start() {
    log_info "Starting services..."
    docker-compose up -d
    log_info "Services started"
}

stop() {
    log_info "Stopping services..."
    docker-compose down
    log_info "Services stopped"
}

restart() {
    stop
    start
}

logs() {
    docker-compose logs -f
}

backup() {
    log_info "Creating backup..."
    BACKUP_FILE="backup_$(date +%Y%m%d_%H%M%S).tar.gz"
    docker-compose exec app tar -czf /app/backups/${BACKUP_FILE} /app/data
    log_info "Backup created: ${BACKUP_FILE}"
}

show_help() {
    echo "Go Quiz Deploy Script"
    echo ""
    echo "Usage: $0 [command]"
    echo ""
    echo "Commands:"
    echo "  build     Build Docker image"
    echo "  deploy    Deploy to remote server"
    echo "  start     Start services locally"
    echo "  stop      Stop services"
    echo "  restart   Restart services"
    echo "  logs      View logs"
    echo "  backup    Create backup"
    echo "  help      Show this help"
    echo ""
    echo "Environment Variables:"
    echo "  DEPLOY_USER  Remote SSH user (default: root)"
    echo "  DEPLOY_HOST  Remote server hostname/IP"
    echo ""
}

# Main
case "${1:-help}" in
    build)
        check_prerequisites
        build
        ;;
    deploy)
        check_prerequisites
        build
        deploy
        ;;
    start)
        start
        ;;
    stop)
        stop
        ;;
    restart)
        restart
        ;;
    logs)
        logs
        ;;
    backup)
        backup
        ;;
    help|--help|-h)
        show_help
        ;;
    *)
        log_error "Unknown command: $1"
        show_help
        exit 1
        ;;
esac
