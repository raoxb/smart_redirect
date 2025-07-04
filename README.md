# Smart Redirect

A high-performance URL shortening and redirection service with advanced traffic management features.

## Features

- Dynamic URL generation with business unit and channel parameters
- Multi-target traffic distribution with configurable weights
- Advanced rate limiting (IP-based, geographic, caps)
- Parameter transformation and injection
- Admin dashboard for link management
- Multi-user support with permission management
- RESTful API for batch operations

## Tech Stack

- **Backend**: Go + Gin Framework
- **Database**: PostgreSQL + Redis
- **Frontend**: React + Ant Design

## Project Structure

```
smart_redirect/
├── cmd/server/          # Application entry point
├── internal/            # Internal packages
│   ├── api/            # HTTP handlers
│   ├── config/         # Configuration management
│   ├── database/       # Database connections
│   ├── middleware/     # HTTP middlewares
│   ├── models/         # Data models
│   ├── services/       # Business logic
│   └── utils/          # Utility functions
├── pkg/                # Public packages
│   ├── logger/         # Logging utilities
│   └── validator/      # Validation utilities
├── migrations/         # Database migrations
├── scripts/           # Build and deployment scripts
└── docs/             # Documentation
```

## Getting Started

### Prerequisites

- Go 1.21+
- PostgreSQL 14+
- Redis 6+
- Node.js 18+ (for frontend)

### Installation

1. Clone the repository
```bash
git clone git@github.com:raoxb/smart_redirect.git
cd smart_redirect
```

2. Install dependencies
```bash
go mod download
```

3. Set up configuration
```bash
cp config/example.yaml config/local.yaml
# Edit config/local.yaml with your settings
```

4. Run database migrations
```bash
go run scripts/migrate.go up
```

5. Start the server
```bash
go run cmd/server/main.go
```

## License

This project is proprietary software.