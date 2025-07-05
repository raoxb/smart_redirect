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

### Frontend Structure
- **frontend/src/pages**: Page components (Dashboard, Links, etc.)
- **frontend/src/components**: Reusable UI components
- **frontend/src/services**: API client and service layer
- **frontend/src/hooks**: Custom React hooks for data fetching
- **frontend/src/store**: Zustand state management
- **frontend/src/types**: TypeScript type definitions
- **frontend/src/utils**: Utility functions and helpers

### Data Flow
1. User accesses short URL → Redirect service
2. Service checks Redis cache → Falls back to PostgreSQL
3. Apply rate limiting and caps checking
4. Select target based on weights and rules
5. Transform parameters and redirect

## Common Commands

### Backend Development
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

### Frontend Development
```bash
# Navigate to frontend directory
cd frontend

# Install dependencies
npm install

# Run development server
npm run dev

# Build for production
npm run build

# Type checking
npm run type-check

# Linting
npm run lint
```

### Database Initialization
```bash
# Initialize database with test data
go run scripts/init_db.go

# Add test targets for links
python3 add_test_targets.py

# Create test data via API
python3 create_test_data.py
```

### Stress Testing
```bash
# Run stress test (duration_hours, threads)
python3 stress_test.py 1 4

# Run extended stability test  
python3 stress_test.py 24 8

# Monitor test progress
tail -f stress_test_*.log
python3 monitor_test.py
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
- **Business Logic Rate Limiting**: IP-based 12-hour intervals for target allocation
- **Geographic Targeting**: Country-based target access control  
- **Caps Management**: Per-target and total visit limits with backup URL redirection
- **Abuse Prevention**: Smart IP blocking for malicious traffic
- **Note**: Management API endpoints are not rate-limited to ensure admin access

### Link Management & UI
- **Enhanced Links Page**: Expanded action buttons with direct copy URL functionality
- **Copy URL Integration**: One-click copy of actual redirect URLs (not hardcoded domains)
- **Multi-port Support**: Automatic port detection for dev environments (3000, 3001, 5173 → 8080)
- **Inline Actions**: View Details and Delete buttons directly visible
- **Real-time URL Generation**: Dynamic URL generation based on current environment

### Target Distribution & Testing
- **Multi-target Support**: Each link can have multiple targets with weight-based distribution
- **Geographic Routing**: IP-based country targeting for personalized redirects
- **Parameter Transformation**: Dynamic parameter mapping and static parameter injection
- **Comprehensive Testing**: 16-country IP pools for global traffic simulation

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
- **Stress Tests**: Root directory - Load testing and stability validation
- **Test Scripts**: Python scripts for data creation and multi-link testing

### Stress Testing Infrastructure
- **Multi-link Testing**: 6 short links with 21 targets total
- **Global IP Simulation**: 16 countries/regions IP pools (US, CN, GB, DE, AU, CA, FR, IT, JP, KR, BR, IN, RU, ES, NL, SE)
- **Concurrent Load**: Configurable thread count for parallel testing
- **Real-world Simulation**: Varied user agents, referers, and request patterns
- **Geographic Distribution**: Country-based target allocation testing

### Running Tests
```bash
# Run unit and integration tests
make test-all
make test-unit
make test-integration

# Run stress tests
python3 stress_test.py <hours> <threads>

# Examples:
python3 stress_test.py 1 4      # 1 hour, 4 threads
python3 stress_test.py 24 8     # 24 hours, 8 threads
python3 stress_test.py 0.1 2    # 6 minutes, 2 threads

# Monitor running tests
python3 monitor_test.py
tail -f stress_test_*.log
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
- Merge to master branch
- Release tag creation

## Recent Important Changes

### Rate Limiting Architecture Fix (e4975e5)
**Critical**: Removed inappropriate rate limiting from management API endpoints
- **Problem**: Rate limiting was blocking admin login and operations
- **Solution**: Rate limiting now only applies to redirect business logic per 302.md spec
- **Impact**: Admin panel and API management now work without artificial limits

### Enhanced Link Management UI (7c2ad0a)
- **Copy URL Functionality**: Real redirect URLs instead of hardcoded domains
- **Multi-port Support**: Automatic dev environment port detection (3000/3001/5173 → 8080)
- **Expanded Actions**: Direct action buttons instead of dropdown menus
- **Better UX**: One-click copy URL from Link ID column

### Comprehensive Testing Infrastructure (70906d2)
- **Multi-link Testing**: 6 links with 21 targets across different business units
- **Global Coverage**: 16 countries/regions IP simulation
- **Stress Testing**: Configurable duration and thread count
- **Real-world Patterns**: Varied user agents, referers, and request timing

## Important Implementation Notes

### Rate Limiting Design
According to 302.md requirements, rate limiting should ONLY affect redirect logic:
- ✅ IP-based 12-hour intervals for target allocation
- ✅ Geographic targeting for country-based access
- ✅ Caps management with backup URL redirection
- ❌ NOT applied to admin/management endpoints

### URL Generation
Frontend generateShortUrl() function automatically handles:
- Development environments (ports 3000, 3001, 5173 → 8080)
- Production environments (uses actual domain)
- Optional network parameters
- Proper HTTPS/HTTP protocol handling

### Database Relationships
- Links → Targets (one-to-many)
- Links → AccessLogs (one-to-many) 
- Targets → AccessLogs (one-to-many)
- Users → Links (many-to-many via permissions)

### Testing Best Practices
- Always test both management and redirect functionality
- Use stress_test.py for load testing with real-world patterns
- Monitor Redis memory usage during extended tests
- Check geographic distribution in analytics after testing