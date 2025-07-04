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
- **Frontend**: React + TypeScript + Ant Design
- **State Management**: Zustand + React Query
- **Build Tool**: Vite
- **Styling**: Ant Design + CSS

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
│   ├── auth/           # JWT authentication
│   └── geoip/          # IP geolocation
├── frontend/           # React admin dashboard
│   ├── src/
│   │   ├── pages/      # Page components
│   │   ├── components/ # Reusable components
│   │   ├── services/   # API services
│   │   ├── hooks/      # Custom hooks
│   │   ├── store/      # State management
│   │   └── utils/      # Utilities
│   └── public/         # Static assets
├── migrations/         # Database migrations
├── scripts/           # Build and deployment scripts
├── test/              # Test suites
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

5. Start the backend server
```bash
go run cmd/server/main.go
```

6. Start the frontend development server
```bash
cd frontend
npm install
npm run dev
```

7. Access the application
- Admin Dashboard: http://localhost:3000
- API: http://localhost:8080

## License

This project is proprietary software.