# Smart Redirect Deployment Guide

This guide provides comprehensive instructions for deploying Smart Redirect in different environments.

## Table of Contents

1. [Quick Start](#quick-start)
2. [Environment Setup](#environment-setup)
3. [Development Deployment](#development-deployment)
4. [Production Deployment](#production-deployment)
5. [Monitoring Setup](#monitoring-setup)
6. [SSL Configuration](#ssl-configuration)
7. [Backup and Recovery](#backup-and-recovery)
8. [Troubleshooting](#troubleshooting)

## Quick Start

### Prerequisites

- Docker and Docker Compose
- Git
- OpenSSL (for SSL certificates)
- Python 3.7+ (for testing scripts)

### Clone and Setup

```bash
git clone https://github.com/raoxb/smart_redirect.git
cd smart_redirect
```

### Development Deployment

```bash
# Start development environment
./scripts/deploy.sh dev start

# Initialize database
./scripts/deploy.sh dev init-db

# Check service status
./scripts/deploy.sh dev status
```

### Production Deployment

```bash
# Setup environment files
cp .env.example .env.prod
# Edit .env.prod with production values

# Generate SSL certificates
./scripts/ssl.sh yourdomain.com prod

# Start production environment
./scripts/deploy.sh prod start

# Initialize database
./scripts/deploy.sh prod init-db
```

## Environment Setup

### Environment Files

Create environment-specific configuration files:

**`.env.dev`** (Development)
```env
# Database Configuration
DB_HOST=postgres
DB_PORT=5432
DB_NAME=smart_redirect
DB_USER=postgres
DB_PASSWORD=dev_password

# Redis Configuration
REDIS_HOST=redis
REDIS_PORT=6379
REDIS_PASSWORD=dev_redis_password

# JWT Configuration
JWT_SECRET=dev_jwt_secret_change_in_production

# Server Configuration
SERVER_PORT=8080
SERVER_HOST=0.0.0.0

# Rate Limiting
RATE_LIMIT_REQUESTS=100
RATE_LIMIT_WINDOW=3600

# Logging
LOG_LEVEL=debug
LOG_FORMAT=text
```

**`.env.prod`** (Production)
```env
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
RATE_LIMIT_REQUESTS=1000
RATE_LIMIT_WINDOW=3600

# Logging
LOG_LEVEL=info
LOG_FORMAT=json
```

### Directory Structure

```
smart_redirect/
├── config/
│   ├── local.yaml
│   ├── dev.yaml
│   └── production.yaml
├── docker-compose.dev.yml
├── docker-compose.prod.yml
├── docker-compose.monitoring.yml
├── scripts/
│   ├── deploy.sh
│   ├── backup.sh
│   └── ssl.sh
├── monitoring/
│   ├── prometheus/
│   ├── grafana/
│   ├── loki/
│   └── alertmanager/
├── nginx/
│   └── nginx.prod.conf
└── ssl/
    ├── cert.pem
    ├── key.pem
    └── ca-bundle.pem
```

## Development Deployment

### Using Docker Compose

```bash
# Start all services
docker-compose -f docker-compose.dev.yml up -d

# View logs
docker-compose -f docker-compose.dev.yml logs -f

# Stop services
docker-compose -f docker-compose.dev.yml down
```

### Using Deployment Script

```bash
# Start development environment
./scripts/deploy.sh dev start

# View logs
./scripts/deploy.sh dev logs

# Check status
./scripts/deploy.sh dev status

# Stop services
./scripts/deploy.sh dev stop
```

### Development Services

- **Backend**: http://localhost:8080
- **Frontend**: http://localhost:3000
- **Database**: localhost:5432
- **Redis**: localhost:6379

### Hot Reload

The development environment includes hot reload for both frontend and backend:

- Frontend: Vite dev server with HMR
- Backend: Air for Go hot reload

## Production Deployment

### System Requirements

- **CPU**: 2+ cores
- **RAM**: 4GB minimum, 8GB recommended
- **Storage**: 50GB+ SSD
- **Network**: 100Mbps+ bandwidth

### Production Services

- **Application**: Smart Redirect backend
- **Database**: PostgreSQL with connection pooling
- **Cache**: Redis with persistence
- **Proxy**: Nginx with SSL termination
- **Monitoring**: Prometheus, Grafana, Loki

### Deployment Steps

1. **Prepare Environment**
   ```bash
   # Create production environment file
   cp .env.example .env.prod
   vim .env.prod
   ```

2. **Generate SSL Certificates**
   ```bash
   # For development (self-signed)
   ./scripts/ssl.sh localhost dev
   
   # For production (Let's Encrypt recommended)
   ./scripts/ssl.sh yourdomain.com prod
   ```

3. **Start Services**
   ```bash
   ./scripts/deploy.sh prod start
   ```

4. **Initialize Database**
   ```bash
   ./scripts/deploy.sh prod init-db
   ```

5. **Verify Deployment**
   ```bash
   ./scripts/deploy.sh prod health
   ```

### Production URLs

- **Application**: https://yourdomain.com
- **Admin Panel**: https://yourdomain.com/admin
- **Health Check**: https://yourdomain.com/health

## Monitoring Setup

### Start Monitoring Stack

```bash
# Start monitoring services
docker-compose -f docker-compose.monitoring.yml up -d

# Check monitoring status
docker-compose -f docker-compose.monitoring.yml ps
```

### Monitoring Services

- **Prometheus**: http://localhost:9090
- **Grafana**: http://localhost:3001 (admin/admin)
- **Alertmanager**: http://localhost:9093
- **Loki**: http://localhost:3100

### Metrics Endpoints

- **Application Metrics**: http://localhost:8080/metrics
- **Node Metrics**: http://localhost:9100/metrics
- **Redis Metrics**: http://localhost:9121/metrics
- **PostgreSQL Metrics**: http://localhost:9187/metrics

### Grafana Dashboards

Import these dashboard IDs for monitoring:

- **Go Application**: 10826
- **PostgreSQL**: 9628
- **Redis**: 763
- **Nginx**: 12559
- **Node Exporter**: 1860

### Alerting Configuration

Alerts are configured for:

- High response time (>500ms)
- High error rate (>10%)
- Database/Redis connection issues
- High CPU/Memory usage
- Disk space usage
- Application downtime

## SSL Configuration

### Development SSL

```bash
# Generate self-signed certificate
./scripts/ssl.sh localhost dev
```

### Production SSL

For production, use Let's Encrypt:

```bash
# Install certbot
sudo apt-get install certbot python3-certbot-nginx

# Generate certificate
sudo certbot certonly --standalone -d yourdomain.com

# Copy certificates
sudo cp /etc/letsencrypt/live/yourdomain.com/fullchain.pem ./ssl/cert.pem
sudo cp /etc/letsencrypt/live/yourdomain.com/privkey.pem ./ssl/key.pem
sudo cp /etc/letsencrypt/live/yourdomain.com/chain.pem ./ssl/ca-bundle.pem

# Set up auto-renewal
sudo crontab -e
# Add: 0 12 * * * /usr/bin/certbot renew --quiet
```

### SSL Security Features

- TLS 1.2 and 1.3 support
- Strong cipher suites
- OCSP stapling
- HSTS headers
- Security headers (CSP, X-Frame-Options, etc.)

## Backup and Recovery

### Automated Backups

```bash
# Create backup
./scripts/backup.sh prod

# Backups are stored in ./backups/ directory
```

### Manual Backup

```bash
# Database backup
docker-compose -f docker-compose.prod.yml exec postgres pg_dump -U postgres smart_redirect > backup.sql

# Redis backup
docker-compose -f docker-compose.prod.yml exec redis redis-cli BGSAVE
```

### Recovery

```bash
# Restore database
docker-compose -f docker-compose.prod.yml exec -T postgres psql -U postgres smart_redirect < backup.sql

# Restore Redis
docker cp backup.rdb $(docker-compose -f docker-compose.prod.yml ps -q redis):/data/dump.rdb
docker-compose -f docker-compose.prod.yml restart redis
```

### Backup Schedule

Set up automated backups with cron:

```bash
# Edit crontab
crontab -e

# Add backup job (daily at 2 AM)
0 2 * * * /path/to/smart_redirect/scripts/backup.sh prod
```

## Troubleshooting

### Common Issues

**Service Won't Start**
```bash
# Check logs
./scripts/deploy.sh [env] logs

# Check status
./scripts/deploy.sh [env] status

# Check health
./scripts/deploy.sh [env] health
```

**Database Connection Issues**
```bash
# Check database logs
docker-compose -f docker-compose.[env].yml logs postgres

# Test connection
docker-compose -f docker-compose.[env].yml exec postgres pg_isready -U postgres
```

**Redis Connection Issues**
```bash
# Check Redis logs
docker-compose -f docker-compose.[env].yml logs redis

# Test connection
docker-compose -f docker-compose.[env].yml exec redis redis-cli ping
```

**SSL Certificate Issues**
```bash
# Check certificate
openssl x509 -in ssl/cert.pem -text -noout

# Verify certificate chain
openssl verify -CAfile ssl/ca-bundle.pem ssl/cert.pem
```

### Performance Tuning

**Database Performance**
```yaml
# In docker-compose.prod.yml, add to postgres environment:
POSTGRES_SHARED_BUFFERS: 256MB
POSTGRES_EFFECTIVE_CACHE_SIZE: 1GB
POSTGRES_WORK_MEM: 4MB
```

**Redis Performance**
```yaml
# In docker-compose.prod.yml, add to redis command:
- --maxmemory 512mb
- --maxmemory-policy allkeys-lru
```

**Nginx Performance**
```nginx
# In nginx.prod.conf, add to http block:
worker_processes auto;
worker_connections 2048;
keepalive_requests 1000;
```

### Log Analysis

**Application Logs**
```bash
# View application logs
docker-compose logs smart_redirect

# Filter by level
docker-compose logs smart_redirect | grep ERROR
```

**Access Logs**
```bash
# View Nginx access logs
docker-compose exec nginx tail -f /var/log/nginx/access.log

# Analyze traffic patterns
docker-compose exec nginx awk '{print $1}' /var/log/nginx/access.log | sort | uniq -c | sort -nr | head -10
```

### Health Checks

**Application Health**
```bash
curl -s http://localhost:8080/health | jq .
```

**Database Health**
```bash
docker-compose exec postgres pg_isready -U postgres
```

**Redis Health**
```bash
docker-compose exec redis redis-cli ping
```

## Security Considerations

### Network Security

- Use internal Docker networks
- Expose only necessary ports
- Implement proper firewall rules
- Use strong passwords and secrets

### Application Security

- Keep dependencies updated
- Use secure JWT secrets
- Implement rate limiting
- Monitor for suspicious activity

### Data Security

- Encrypt sensitive data at rest
- Use encrypted connections (SSL/TLS)
- Regular security audits
- Proper backup encryption

## Maintenance

### Regular Tasks

- **Daily**: Check service health and logs
- **Weekly**: Review monitoring alerts and metrics
- **Monthly**: Update dependencies and security patches
- **Quarterly**: Review and update SSL certificates

### Updates

```bash
# Update application
git pull origin main
docker-compose build --no-cache
./scripts/deploy.sh [env] restart

# Update monitoring
docker-compose -f docker-compose.monitoring.yml pull
docker-compose -f docker-compose.monitoring.yml up -d
```

### Scaling

For high traffic, consider:

- Multiple backend instances
- Database read replicas
- CDN integration
- Load balancing
- Horizontal scaling with Kubernetes

## Support

For issues and questions:

- Check the troubleshooting section
- Review application logs
- Open an issue on GitHub
- Contact the development team

---

This deployment guide provides comprehensive instructions for deploying Smart Redirect in various environments. Follow the appropriate sections based on your deployment needs.