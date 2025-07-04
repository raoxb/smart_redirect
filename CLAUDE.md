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

## Testing Strategy

1. Unit tests for business logic in services layer
2. Integration tests for API endpoints
3. Load testing for redirect performance
4. Mock Redis/PostgreSQL for isolated testing