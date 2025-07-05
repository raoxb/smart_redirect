# Smart Redirect Service

A high-performance URL shortening and redirection service built with Go and PostgreSQL, featuring advanced traffic management, weighted distribution, geographic targeting, and comprehensive analytics.

## 🌟 Features

### Core Functionality
- **URL Shortening**: Generate unique short links with UUID-based IDs
- **Smart Redirection**: 302 redirects with configurable business logic
- **Multi-target Support**: Weight-based traffic distribution across multiple targets
- **Geographic Targeting**: Country-based target allocation using IP geolocation
- **Parameter Transformation**: Dynamic URL parameter mapping and static injection

### Traffic Management
- **Business Logic Rate Limiting**: IP-based intervals for intelligent target allocation
- **Caps Management**: Per-target and total visit limits with backup URL fallback
- **Access Logging**: Comprehensive tracking of all redirects with analytics
- **Real-time Statistics**: Live monitoring of link performance and traffic distribution

### Management Interface
- **Web-based Admin Panel**: React + TypeScript frontend for link management
- **JWT Authentication**: Secure access control for administrative functions
- **Enhanced UI**: Direct copy URL functionality, expanded action buttons
- **Batch Operations**: Bulk link creation, CSV import/export capabilities

## 🏗️ Tech Stack

- **Backend**: Go + Gin Framework + GORM
- **Database**: PostgreSQL + Redis (caching & rate limiting)
- **Frontend**: React + TypeScript + Vite + Ant Design 5
- **State Management**: Zustand + TanStack Query
- **Authentication**: JWT-based stateless auth
- **Testing**: Go testing framework + Python stress testing
- **Deployment**: Docker Compose development environment

## 🚀 Quick Start

### Prerequisites
- Go 1.21+
- Node.js 18+
- PostgreSQL 15+
- Redis 7+
- Docker & Docker Compose (recommended)

### Development Setup

1. **Clone the repository**
   ```bash
   git clone git@github.com:raoxb/smart_redirect.git
   cd smart_redirect
   ```

2. **Start infrastructure services**
   ```bash
   docker-compose up -d postgres redis adminer
   ```

3. **Initialize the database**
   ```bash
   go run scripts/init_db.go
   ```

4. **Start the backend server**
   ```bash
   go run cmd/server/main.go
   ```

5. **Start the frontend development server**
   ```bash
   cd frontend
   npm install
   npm run dev
   ```

6. **Access the application**
   - Admin Panel: http://localhost:3001
   - API Endpoints: http://localhost:8080
   - Database Admin: http://localhost:8081 (adminer)

### Default Credentials
- **Username**: `admin`
- **Password**: `admin123`

## 📖 Usage Examples

### URL Format
Generated short links follow the format:
```
http://your-domain.com/v1/{business_unit}/{link_id}
```

Example: `http://localhost:8080/v1/marketing/abc123`

### Creating Links via API
```bash
curl -X POST http://localhost:8080/api/v1/links \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "business_unit": "marketing",
    "network": "facebook",
    "targets": [
      {
        "url": "https://example.com/landing1",
        "weight": 70,
        "countries": ["US", "CA"]
      },
      {
        "url": "https://example.com/landing2", 
        "weight": 30,
        "countries": ["GB", "DE"]
      }
    ]
  }'
```

```

## 🧪 Testing & Monitoring

### Stress Testing
The project includes comprehensive stress testing capabilities:

```bash
# Basic stress test (1 hour, 4 threads)
python3 stress_test.py 1 4

# Extended stability test (24 hours, 8 threads)  
python3 stress_test.py 24 8

# Monitor test progress
python3 monitor_test.py
tail -f stress_test_*.log
```

**Test Features:**
- 6 short links with 21 targets total
- 16 countries/regions IP simulation
- Real-world traffic patterns
- Geographic distribution validation

### Built-in Analytics
- **Dashboard**: Real-time statistics and charts
- **Access Logs**: Detailed per-request tracking  
- **Statistics Page**: Link performance and target distribution
- **Geographic Reports**: Country-based traffic analysis

## ⚙️ Configuration

Configuration is managed through YAML files in the `config/` directory:

```yaml
# config/local.yaml
server:
  port: 8080
  mode: debug

database:
  postgres:
    host: localhost
    port: 5432
    user: smartredirect
    password: smart123
    dbname: smart_redirect

redis:
  addr: localhost:6379
  db: 0

security:
  jwt_secret: "your-secret-key"
  jwt_expire_hours: 24
```

## 📋 Key Design Principles

Based on [302.md](302.md) requirements:

1. **Smart Rate Limiting**: IP-based 12-hour intervals for target allocation (not blocking admin access)
2. **Geographic Targeting**: Country-based access control for personalized redirects
3. **Caps Management**: Per-target and total visit limits with backup URL redirection
4. **Parameter Flexibility**: Dynamic parameter mapping and static parameter injection
5. **Performance**: Redis caching, efficient algorithms, minimal redirect latency

## 🤝 Contributing

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'feat: add amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## 📚 Documentation

- **[CLAUDE.md](CLAUDE.md)**: Development guidance and architecture details
- **[302.md](302.md)**: Feature specifications and requirements
- **API Documentation**: Available in the admin panel

## 📄 License

This project is proprietary software.