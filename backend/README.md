# AiOpsHub Backend

Go backend service for AiOpsHub - Intelligent Operations Platform

## Architecture

- **API Server**: RESTful API service (port 8080)

## Quick Start

### Prerequisites

- Go 1.24+
- PostgreSQL 14+
- Redis 7+

### Run API Server

```bash
# Copy config
cp config/config.yaml.example config/config.yaml

# Build
go build -o bin/api-server ./cmd/api-server

# Run
./bin/api-server
```

## Project Structure

```
backend/
├── cmd/                    # Application entry points
│   ├── api-server/        # API Server
│   └── temporal-worker/   # Temporal Worker
├── internal/              # Private application code
│   ├── config/           # Configuration management
│   ├── handler/          # HTTP handlers
│   ├── middleware/       # HTTP middlewares
│   ├── model/            # Data models
│   ├── repository/       # Data access layer
│   └── service/           # Business logic layer
├── pkg/                   # Public packages
│   └── logger/           # Logging utilities
└── config/                # Configuration files
```

## API Endpoints

- `GET /health` - Health check
- `POST /api/v1/auth/login` - User login
- `POST /api/v1/auth/register` - User registration
- `/api/v1/agents/*` - Agent management
- `/api/v1/workflows/*` - Workflow management
- `/api/v1/alerts/*` - Alert management
- `/api/v1/knowledge/*` - Knowledge base management
- `/api/v1/datasources/*` - Datasource management
- `/api/v1/tools/*` - Tool management
- `/api/v1/monitor/*` - Monitoring and statistics

## Development Status

- [x] Project structure
- [x] Configuration management
- [x] Logger
- [x] HTTP middleware
- [x] Basic handlers
- [ ] Database integration (GORM)
- [ ] Langchaingo Agent implementation
- [ ] Full API implementation
- [ ] Authentication and authorization
- [ ] Tests

## Technology Stack

- **Web Framework**: Gin
- **Configuration**: Viper
- **Logging**: Zap
- **Agent Framework**: langchaingo (planned)
- **Database**: PostgreSQL with GORM
- **Cache**: Redis