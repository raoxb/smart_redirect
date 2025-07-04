# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

Smart Redirect is a high-performance URL shortening and redirection service built with Go and PostgreSQL. It provides advanced traffic management features including weighted distribution, rate limiting, and parameter transformation.

## Architecture

### Backend Structure
- **cmd/server**: Application entry point
- **internal/api**: HTTP handlers for redirect and admin APIs
- **internal/models**: GORM models for database entities (Link, Target, User, AccessLog)
- **internal/services**: Business logic layer
- **internal/middleware**: Authentication, rate limiting, logging
- **internal/database**: PostgreSQL and Redis connection management
- **internal/config**: Configuration management using Viper

### Data Flow
1. User accesses short URL → Redirect service
2. Service checks Redis cache → Falls back to PostgreSQL
3. Apply rate limiting and caps checking
4. Select target based on weights and rules
5. Transform parameters and redirect

## Common Commands

### Development
```bash
# Install dependencies
go mod download

# Run the server (requires config/local.yaml)
go run cmd/server/main.go

# Run with custom config
go run cmd/server/main.go -config=config/dev.yaml

# Build binary
make build

# Run tests
make test

# Run with hot reload (requires air)
make dev
```

### Database
```bash
# Run migrations up
make migrate-up

# Rollback migrations
make migrate-down
```

### Git Workflow
```bash
# Remote repository
git@github.com:raoxb/smart_redirect.git

# Commit changes (always include descriptive messages)
git add .
git commit -m "feat: implement rate limiting logic"
git push origin main
```

## Key Design Decisions

1. **UUID-based Link IDs**: First 6 characters of UUID for uniqueness
2. **Redis Caching**: All active links cached for performance
3. **Weighted Random Selection**: Efficient algorithm for traffic distribution
4. **JSONB for Parameters**: Flexible parameter storage in PostgreSQL
5. **JWT Authentication**: Stateless auth for admin APIs
6. **Multi-layer Rate Limiting**: IP-based, link-specific, and global caps
7. **Geo-targeting**: Country-based target filtering using IP geolocation
8. **Batch Operations**: Efficient bulk operations for link management

## Advanced Features

### Rate Limiting & Security
- **IP Rate Limiting**: 100 requests/hour per IP (configurable)
- **Link-specific Limits**: 10 requests/12h per IP per link
- **Automatic IP Blocking**: Abuse detection and blocking
- **Geo-targeting**: Country-based access control
- **Global Traffic Caps**: Per-link visit limits

### Batch Operations
- **Bulk Link Creation**: Create multiple links with targets
- **CSV Import/Export**: Bulk data management via CSV files
- **Batch Updates**: Update multiple links simultaneously
- **Template System**: Create reusable link configurations

### Statistics & Monitoring
- **Real-time Analytics**: Link performance tracking
- **IP Access Monitoring**: Detailed IP access logs
- **Hourly Statistics**: Time-based analytics
- **Country-based Stats**: Geographic traffic analysis

## Testing Strategy

### Test Structure
- **Unit Tests**: `/test/unit/` - Business logic testing with mocks
- **Integration Tests**: `/test/integration/` - End-to-end API testing
- **Fixtures**: `/test/fixtures/` - Test data and helpers
- **Test Utils**: `/test/testutil/` - Shared testing utilities

### Running Tests
```bash
# Run all tests
make test-all

# Run specific test types
make test-unit
make test-integration

# Run with coverage
make test-coverage

# Run load tests
make test-load

# Run benchmarks
make bench
```

### Test Configuration
- Test database: SQLite in-memory or PostgreSQL test DB
- Redis test DB: Database 1 (separate from main)
- JWT secret: `test-secret-key-for-testing`
- Reduced rate limits for faster testing

### Coverage Targets
- Overall coverage: >80%
- Critical paths (redirect logic): >90%
- Error handling: >85%

### Continuous Integration
Tests run automatically on:
- Pull request creation
- Merge to main branch
- Release tag creation