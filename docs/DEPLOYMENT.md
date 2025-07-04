# Smart Redirect Deployment Guide

This guide covers various deployment options for the Smart Redirect service.

## Table of Contents

- [Prerequisites](#prerequisites)
- [Docker Deployment](#docker-deployment)
- [Kubernetes Deployment](#kubernetes-deployment)
- [Manual Deployment](#manual-deployment)
- [Environment Configuration](#environment-configuration)
- [Performance Tuning](#performance-tuning)
- [Monitoring](#monitoring)
- [Troubleshooting](#troubleshooting)

## Prerequisites

### System Requirements

- **CPU**: 2+ cores recommended
- **Memory**: 4GB+ RAM recommended
- **Storage**: 20GB+ available space
- **Network**: Stable internet connection for geolocation services

### Software Dependencies

- Docker 20.10+ and Docker Compose 2.0+
- PostgreSQL 14+ (if not using Docker)
- Redis 6+ (if not using Docker)
- Go 1.21+ (for manual deployment)

## Docker Deployment

### Quick Start

1. **Clone the repository**:
   ```bash
   git clone git@github.com:raoxb/smart_redirect.git
   cd smart_redirect
   ```

2. **Configure environment**:
   ```bash
   cp config/example.yaml config/docker.yaml
   # Edit config/docker.yaml with your settings
   ```

3. **Start services**:
   ```bash
   docker-compose up -d
   ```

4. **Initialize admin user**:
   ```bash
   docker-compose exec app ./smart_redirect -config=config/local.yaml
   # Run: go run scripts/init_db.go -config=config/docker.yaml
   ```

5. **Verify deployment**:
   ```bash
   curl http://localhost:8080/health
   ```

### Production Docker Setup

1. **Set production environment variables**:
   ```bash
   export JWT_SECRET="your-super-secure-jwt-secret-key"
   export POSTGRES_PASSWORD="secure-database-password"
   ```

2. **Use production compose file**:
   ```bash
   docker-compose -f docker-compose.prod.yml up -d
   ```

3. **Configure SSL (recommended)**:
   - Place SSL certificates in `nginx/ssl/`
   - Uncomment SSL configuration in `nginx/nginx.conf`
   - Restart nginx: `docker-compose restart nginx`

## Kubernetes Deployment

### 1. Create Namespace
```yaml
apiVersion: v1
kind: Namespace
metadata:
  name: smart-redirect
```

### 2. PostgreSQL Deployment
```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: postgres
  namespace: smart-redirect
spec:
  replicas: 1
  selector:
    matchLabels:
      app: postgres
  template:
    metadata:
      labels:
        app: postgres
    spec:
      containers:
      - name: postgres
        image: postgres:15-alpine
        env:
        - name: POSTGRES_DB
          value: "smart_redirect"
        - name: POSTGRES_USER
          value: "postgres"
        - name: POSTGRES_PASSWORD
          valueFrom:
            secretKeyRef:
              name: postgres-secret
              key: password
        ports:
        - containerPort: 5432
        volumeMounts:
        - name: postgres-storage
          mountPath: /var/lib/postgresql/data
      volumes:
      - name: postgres-storage
        persistentVolumeClaim:
          claimName: postgres-pvc
```

### 3. Redis Deployment
```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: redis
  namespace: smart-redirect
spec:
  replicas: 1
  selector:
    matchLabels:
      app: redis
  template:
    metadata:
      labels:
        app: redis
    spec:
      containers:
      - name: redis
        image: redis:7-alpine
        ports:
        - containerPort: 6379
```

### 4. Smart Redirect Application
```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: smart-redirect-app
  namespace: smart-redirect
spec:
  replicas: 3
  selector:
    matchLabels:
      app: smart-redirect-app
  template:
    metadata:
      labels:
        app: smart-redirect-app
    spec:
      containers:
      - name: app
        image: smart-redirect:latest
        ports:
        - containerPort: 8080
        env:
        - name: JWT_SECRET
          valueFrom:
            secretKeyRef:
              name: app-secret
              key: jwt-secret
        livenessProbe:
          httpGet:
            path: /health
            port: 8080
          initialDelaySeconds: 30
          periodSeconds: 10
        readinessProbe:
          httpGet:
            path: /health
            port: 8080
          initialDelaySeconds: 5
          periodSeconds: 5
```

## Manual Deployment

### 1. Install Dependencies

```bash
# Ubuntu/Debian
sudo apt update
sudo apt install postgresql redis-server golang-go

# CentOS/RHEL
sudo yum install postgresql-server redis golang

# Start services
sudo systemctl start postgresql redis
sudo systemctl enable postgresql redis
```

### 2. Database Setup

```bash
# Create database
sudo -u postgres createdb smart_redirect
sudo -u postgres createuser smart_redirect_user

# Set password
sudo -u postgres psql -c "ALTER USER smart_redirect_user PASSWORD 'your_password';"
sudo -u postgres psql -c "GRANT ALL PRIVILEGES ON DATABASE smart_redirect TO smart_redirect_user;"
```

### 3. Build and Deploy Application

```bash
# Clone and build
git clone git@github.com:raoxb/smart_redirect.git
cd smart_redirect

# Install dependencies
go mod download

# Build application
make build

# Copy configuration
cp config/example.yaml config/production.yaml
# Edit config/production.yaml with your settings

# Create systemd service
sudo tee /etc/systemd/system/smart-redirect.service > /dev/null <<EOF
[Unit]
Description=Smart Redirect Service
After=network.target postgresql.service redis.service

[Service]
Type=simple
User=smartredirect
WorkingDirectory=/opt/smart-redirect
ExecStart=/opt/smart-redirect/smart_redirect -config=config/production.yaml
Restart=always
RestartSec=5

[Install]
WantedBy=multi-user.target
EOF

# Create user and setup directories
sudo useradd -r -s /bin/false smartredirect
sudo mkdir -p /opt/smart-redirect
sudo cp smart_redirect config/ /opt/smart-redirect/
sudo chown -R smartredirect:smartredirect /opt/smart-redirect

# Enable and start service
sudo systemctl enable smart-redirect
sudo systemctl start smart-redirect
```

## Environment Configuration

### Required Environment Variables

```bash
# JWT Configuration
JWT_SECRET="your-super-secure-secret-key"

# Database Configuration  
DB_HOST="localhost"
DB_PORT="5432"
DB_USER="postgres"
DB_PASSWORD="your-db-password"
DB_NAME="smart_redirect"

# Redis Configuration
REDIS_ADDR="localhost:6379"
REDIS_PASSWORD=""

# Application Configuration
GIN_MODE="release"
SERVER_PORT="8080"
```

### Configuration File Structure

```yaml
server:
  port: 8080
  mode: release

database:
  postgres:
    host: ${DB_HOST:-localhost}
    port: ${DB_PORT:-5432}
    user: ${DB_USER:-postgres}
    password: ${DB_PASSWORD}
    dbname: ${DB_NAME:-smart_redirect}
    sslmode: disable

redis:
  addr: ${REDIS_ADDR:-localhost:6379}
  password: ${REDIS_PASSWORD:-""}
  db: 0

security:
  jwt_secret: ${JWT_SECRET}
  jwt_expire_hours: 24

rate_limit:
  ip_limit_per_hour: 1000
  global_daily_cap: 10000000
```

## Performance Tuning

### Database Optimization

```sql
-- PostgreSQL optimizations
ALTER SYSTEM SET shared_buffers = '256MB';
ALTER SYSTEM SET effective_cache_size = '1GB';
ALTER SYSTEM SET maintenance_work_mem = '64MB';
ALTER SYSTEM SET wal_buffers = '16MB';
ALTER SYSTEM SET checkpoint_completion_target = 0.9;

-- Reload configuration
SELECT pg_reload_conf();

-- Create indexes for better performance
CREATE INDEX CONCURRENTLY idx_access_logs_created_at ON access_logs(created_at);
CREATE INDEX CONCURRENTLY idx_access_logs_ip ON access_logs(ip);
CREATE INDEX CONCURRENTLY idx_links_link_id ON links(link_id);
```

### Redis Configuration

```conf
# /etc/redis/redis.conf
maxmemory 512mb
maxmemory-policy allkeys-lru
save 900 1
save 300 10
save 60 10000
```

### Application Tuning

```yaml
# config/production.yaml
database:
  postgres:
    max_idle_conns: 25
    max_open_conns: 100
    conn_max_lifetime: 3600

redis:
  pool_size: 50

rate_limit:
  ip_limit_per_hour: 5000
  redirect_limit_per_hour: 10000
```

## Monitoring

### Health Checks

```bash
# Application health
curl http://localhost:8080/health

# Database health
psql -h localhost -U postgres -c "SELECT 1;"

# Redis health
redis-cli ping
```

### Metrics Collection

1. **Prometheus Configuration**:
```yaml
# prometheus.yml
scrape_configs:
  - job_name: 'smart-redirect'
    static_configs:
      - targets: ['localhost:8080']
    metrics_path: '/metrics'
```

2. **Grafana Dashboard**: Import the dashboard from `docs/grafana-dashboard.json`

### Log Monitoring

```bash
# Application logs
tail -f /var/log/smart-redirect/app.log

# Docker logs
docker-compose logs -f app

# Kubernetes logs
kubectl logs -f deployment/smart-redirect-app -n smart-redirect
```

## Troubleshooting

### Common Issues

1. **Database Connection Failed**:
   ```bash
   # Check PostgreSQL status
   sudo systemctl status postgresql
   
   # Check connectivity
   telnet postgres-host 5432
   
   # Verify credentials
   psql -h host -U user -d database
   ```

2. **Redis Connection Failed**:
   ```bash
   # Check Redis status
   sudo systemctl status redis
   
   # Test connection
   redis-cli -h redis-host ping
   ```

3. **High Memory Usage**:
   ```bash
   # Check Redis memory
   redis-cli info memory
   
   # Clear Redis cache if needed
   redis-cli flushdb
   ```

4. **Rate Limiting Issues**:
   ```bash
   # Check rate limit counters
   redis-cli keys "rate_limit:*"
   
   # Clear rate limits for IP
   redis-cli del "rate_limit:ip:192.168.1.1"
   ```

### Performance Issues

1. **Slow Redirects**:
   - Check database indexes
   - Monitor Redis hit ratio
   - Analyze slow query logs

2. **High CPU Usage**:
   - Review rate limiting settings
   - Check for inefficient queries
   - Monitor goroutine count

3. **Memory Leaks**:
   - Enable pprof profiling
   - Monitor heap growth
   - Check for unclosed connections

### Recovery Procedures

1. **Database Recovery**:
   ```bash
   # Restore from backup
   pg_restore -h localhost -U postgres -d smart_redirect backup.sql
   
   # Rebuild indexes
   REINDEX DATABASE smart_redirect;
   ```

2. **Cache Warming**:
   ```bash
   # Warm up cache after Redis restart
   curl -X POST http://localhost:8080/api/v1/admin/cache/warm
   ```

For additional support, check the logs and refer to the application metrics for detailed diagnostics.