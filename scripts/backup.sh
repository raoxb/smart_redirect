#!/bin/bash

# Smart Redirect Backup Script
# Usage: ./scripts/backup.sh [environment]

set -e

ENVIRONMENT=${1:-prod}
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_DIR="$(cd "$SCRIPT_DIR/.." && pwd)"
BACKUP_DIR="$PROJECT_DIR/backups"
TIMESTAMP=$(date +%Y%m%d_%H%M%S)

# Color codes
GREEN='\033[0;32m'
BLUE='\033[0;34m'
NC='\033[0m'

log() {
    echo -e "${BLUE}[$(date +'%Y-%m-%d %H:%M:%S')]${NC} $1"
}

success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

# Create backup directory
mkdir -p "$BACKUP_DIR"

log "Starting backup for $ENVIRONMENT environment..."

# Backup database
log "Backing up PostgreSQL database..."
docker-compose -f docker-compose.$ENVIRONMENT.yml exec -T postgres pg_dump -U postgres smart_redirect > "$BACKUP_DIR/postgres_backup_$TIMESTAMP.sql"
success "PostgreSQL backup completed"

# Backup Redis data
log "Backing up Redis data..."
docker-compose -f docker-compose.$ENVIRONMENT.yml exec -T redis redis-cli --rdb /data/dump.rdb
docker cp $(docker-compose -f docker-compose.$ENVIRONMENT.yml ps -q redis):/data/dump.rdb "$BACKUP_DIR/redis_backup_$TIMESTAMP.rdb"
success "Redis backup completed"

# Backup configuration files
log "Backing up configuration files..."
tar -czf "$BACKUP_DIR/config_backup_$TIMESTAMP.tar.gz" \
    config/ \
    docker-compose.*.yml \
    nginx/ \
    .env.* 2>/dev/null || true
success "Configuration backup completed"

# Cleanup old backups (keep last 7 days)
log "Cleaning up old backups..."
find "$BACKUP_DIR" -name "*backup_*" -type f -mtime +7 -delete 2>/dev/null || true
success "Cleanup completed"

log "Backup completed successfully"
log "Backup files stored in: $BACKUP_DIR"