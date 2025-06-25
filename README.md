# voidrunner

[![CI](https://github.com/starbops/voidrunner/actions/workflows/ci.yml/badge.svg)](https://github.com/starbops/voidrunner/actions/workflows/ci.yml)

A Go HTTP API server for multi-user task management with JWT authentication, comprehensive OpenAPI documentation, and containerized deployment.

## Features

- 🔐 **JWT Authentication** - Secure user authentication and authorization
- 📋 **Task Management** - User-scoped CRUD operations for tasks
- 👤 **User Management** - User profile management and account operations
- 📖 **OpenAPI Documentation** - Interactive Swagger UI with comprehensive API docs
- 🐳 **Containerized** - Docker support with multi-stage builds
- 🗄️ **Multiple Storage Backends** - Memory (development) and PostgreSQL (production)
- ✅ **Comprehensive Testing** - Unit, integration, and E2E tests
- 🚀 **CI/CD Ready** - GitHub Actions with automated testing and validation

## Quick Start

### Prerequisites
- **Go 1.23+** (for local development)
- **Docker & Docker Compose** (recommended for full setup)
- **golang-migrate** (for manual database operations)

### Option 1: Docker Compose (Recommended)

Get up and running in 30 seconds:

```bash
# 1. Clone the repository
git clone https://github.com/starbops/voidrunner.git
cd voidrunner

# 2. Start application with PostgreSQL database
make docker-up

# 3. Enable API documentation (development only)
ENABLE_DOCS=true make docker-up
```

**🚀 Server is now running at:** http://localhost:8080

**📖 View API Documentation:** http://localhost:8080/docs/ (when `ENABLE_DOCS=true`)  
**Note:** The docs URL will redirect to `/docs/index.html` automatically

### Option 2: Local Development

```bash
# 1. Build and run with memory storage
make run

# 2. Or run with PostgreSQL
make run-postgres
```

### Option 3: Docker Only

```bash
# Build and run in container (memory backend)
make docker-run
```

### Your First API Call

```bash
# Health check
curl http://localhost:8080/api/v1/welcome

# Register a new user
curl -X POST http://localhost:8080/api/v1/register \
  -H "Content-Type: application/json" \
  -d '{
    "username": "john_doe",
    "email": "john@example.com", 
    "password": "securepass123",
    "first_name": "John",
    "last_name": "Doe"
  }'

# Login to get your token
curl -X POST http://localhost:8080/api/v1/login \
  -H "Content-Type: application/json" \
  -d '{
    "identifier": "john_doe",
    "password": "securepass123"
  }'
```

Save the `token` from the login response and use it for authenticated requests!

## API Endpoints

📖 **Complete interactive documentation:** http://localhost:8080/docs/ (requires `ENABLE_DOCS=true`)

### Public Endpoints (No Authentication Required)

| Method | Endpoint | Description |
|--------|----------|-------------|
| `GET` | `/api/v1/welcome` | Welcome message and health check |
| `POST` | `/api/v1/register` | User registration |
| `POST` | `/api/v1/login` | User authentication (returns JWT token) |

### Protected Endpoints (JWT Token Required)

All protected endpoints require the `Authorization: Bearer <token>` header.

**Note:** The logout endpoint (`/api/v1/logout`) requires authentication but handles JWT validation manually rather than using the standard auth middleware.

#### Authentication
| Method | Endpoint | Description |
|--------|----------|-------------|
| `POST` | `/api/v1/logout` | User logout (invalidates token) - requires Bearer token in Authorization header |

#### User Management
| Method | Endpoint | Description |
|--------|----------|-------------|
| `GET` | `/api/v1/users/me` | Get current user profile |
| `PUT` | `/api/v1/users/me` | Update current user profile |
| `DELETE` | `/api/v1/users/me` | Delete current user account |

#### Task Management (User-Scoped)
| Method | Endpoint | Description |
|--------|----------|-------------|
| `GET` | `/api/v1/tasks/` | List user's tasks |
| `POST` | `/api/v1/tasks/` | Create new task |
| `GET` | `/api/v1/tasks/{id}/` | Get specific task |
| `PUT` | `/api/v1/tasks/{id}/` | Update task |
| `DELETE` | `/api/v1/tasks/{id}/` | Delete task |

### Example Usage

```bash
# Set your token (from login response)
export TOKEN="your-jwt-token-here"

# Create a task (status will be set to "pending" automatically)
curl -X POST http://localhost:8080/api/v1/tasks/ \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Complete project",
    "description": "Finish the final report"
  }'

# List all your tasks
curl -H "Authorization: Bearer $TOKEN" \
  http://localhost:8080/api/v1/tasks/

# Update a task status
curl -X PUT http://localhost:8080/api/v1/tasks/1/ \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"status": "completed"}'
```

## Configuration

### Environment Variables

#### Core Configuration
- `STORAGE_BACKEND` - "memory" (default) or "postgres"
- `PORT` - Server port (default: 8080)

#### Authentication
- `JWT_SECRET` - JWT signing secret (⚠️ **required for production**)
- JWT tokens have a fixed expiration of 24 hours (not configurable)

#### Documentation (Security)
- `ENABLE_DOCS` - Enable Swagger UI at `/docs/` (default: false, only enable in development)

#### PostgreSQL Configuration (when using postgres backend)
- `PG_HOST` - PostgreSQL host (default: localhost)
- `PG_PORT` - PostgreSQL port (default: 5432)
- `PG_USER` - Database user
- `PG_PASSWORD` - Database password
- `PG_DBNAME` - Database name

### Configuration Examples

```bash
# Development (memory backend with docs)
export STORAGE_BACKEND=memory
export ENABLE_DOCS=true
export JWT_SECRET=dev-secret-key

# Production (PostgreSQL backend, docs disabled)
export STORAGE_BACKEND=postgres
export JWT_SECRET=your-strong-production-secret
export PG_HOST=localhost
export PG_PORT=5432
export PG_USER=voidrunner
export PG_PASSWORD=secure-password
export PG_DBNAME=voidrunner
```

⚠️ **Security Note:** The documentation endpoint is disabled by default. Only enable `ENABLE_DOCS=true` in development environments.

**Docker Usage:** When using `make docker-up`, set `ENABLE_DOCS=true` as an environment variable to enable documentation:
```bash
ENABLE_DOCS=true make docker-up
```

## Development

### Make Commands Reference

#### Core Build Commands
```bash
make build          # Build the application to ./bin/voidrunner
make clean          # Remove build artifacts from ./bin
make run            # Build and run the application (starts server on :8080)
```

#### Test Commands
```bash
make test           # Run unit tests with coverage
make test-integration # Run integration tests
make test-e2e       # Run end-to-end tests (use STORAGE_BACKEND env var for backend selection)
make test-all       # Run all test suites (unit, integration, E2E)

# Run E2E tests with specific backend
STORAGE_BACKEND=memory make test-e2e
STORAGE_BACKEND=postgres make test-e2e
```

#### Database Commands
```bash
make db-up          # Start PostgreSQL container with docker-compose
make db-down        # Stop PostgreSQL container
make db-migrate     # Run database migrations
make db-reset       # Reset database (drop all tables and re-migrate)
make run-postgres   # Start database, migrate, and run application with PostgreSQL backend
```

#### Docker Commands
```bash
make docker-build   # Build Docker image for the application
make docker-run     # Build and run application in Docker container (memory backend)
make docker-up      # Start all services with docker-compose
make docker-down    # Stop all docker-compose services
make docker-logs    # Show logs from all services

# Start in detached mode
DETACH=1 make docker-up
```

#### Docker Test Commands
```bash
make docker-test              # Run unit tests in Docker container
make docker-test-integration  # Run integration tests with PostgreSQL in Docker
make docker-test-e2e          # Run E2E tests with both backends in Docker
make docker-test-all          # Run all test suites in Docker containers

# Note: Docker test commands use isolated test containers and databases
# They automatically handle service dependencies and cleanup
```

#### Documentation Commands
```bash
make docs           # Generate and validate OpenAPI documentation
make docs-clean     # Remove generated documentation files
```

### Complete Development Workflow

```bash
# 1. Start services with PostgreSQL
ENABLE_DOCS=true make docker-up

# 2. View API documentation (development only)
open http://localhost:8080/docs/

# 3. Run all tests
make test-all

# 4. Generate/update documentation
make docs

# 5. Clean up
make docker-down
```

### Testing Strategy

The project uses a comprehensive three-tier testing approach with both local and Docker execution options:

#### Test Types
- **Unit Tests**: Mock-based testing with coverage reporting (no external dependencies)
- **Integration Tests**: Real PostgreSQL database with complete HTTP testing
- **E2E Tests**: Full application testing with both memory and PostgreSQL backends

#### Execution Options
**Local Testing** (requires local Go and optionally PostgreSQL):
```bash
make test              # Unit tests only
make test-integration  # Integration tests (requires PostgreSQL)
make test-e2e          # E2E tests (supports both backends)
make test-all          # All test suites
```

**Docker Testing** (fully containerized, no local dependencies):
```bash
make docker-test              # Unit tests in Docker
make docker-test-integration  # Integration tests with containerized PostgreSQL
make docker-test-e2e          # E2E tests with both backends in Docker
make docker-test-all          # All test suites in Docker containers
```

The Docker test commands provide isolated environments and automatic cleanup, making them ideal for CI/CD pipelines or when you don't want to set up local PostgreSQL.

### Database Operations

#### Development Setup
```bash
# Start PostgreSQL only
make db-up

# Run migrations
make db-migrate

# Reset database (useful for development)
make db-reset
```

#### Manual Migration Operations
```bash
# Install golang-migrate
brew install golang-migrate

# Run migrations manually
migrate -path db/migrations \
  -database "postgres://voidrunner:password@localhost:5432/voidrunner?sslmode=disable" up

# Rollback last migration
migrate -path db/migrations \
  -database "postgres://voidrunner:password@localhost:5432/voidrunner?sslmode=disable" down 1
```

## Architecture

VoidRunner follows a clean, layered architecture with dependency injection:

- **`cmd/`** - Application entry points and server setup
- **`internal/handlers/`** - HTTP request handlers with OpenAPI annotations
- **`internal/services/`** - Business logic layer
- **`internal/repositories/`** - Data access layer with pluggable backends
- **`internal/models/`** - Domain models and API request/response types
- **`internal/middleware/`** - Authentication and request processing middleware
- **`pkg/`** - Shared packages (auth, config)
- **`docs/`** - Generated OpenAPI documentation
- **`test/`** - Comprehensive test suites (unit, integration, E2E)

## Technologies

- **Go 1.23** - Primary programming language
- **JWT** - Authentication and authorization
- **PostgreSQL** - Production database
- **Docker & Docker Compose** - Containerization
- **OpenAPI 3.0 / Swagger** - API documentation
- **GitHub Actions** - CI/CD pipeline
- **golang-migrate** - Database migrations

## License

This project is licensed under the MIT License - see the LICENSE file for details.