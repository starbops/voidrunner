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

### Local Development

```bash
# Build and run
make run

# Or build and run manually
make build
./bin/voidrunner
```

### With Docker Compose

```bash
# Start application with PostgreSQL database
make docker-compose-up

# Or start in detached mode
make docker-compose-up-detached
```

Server runs on `http://localhost:8080`

**📖 View API Documentation:** http://localhost:8080/docs/

## API Documentation

### Interactive Documentation
📖 **Swagger UI:** http://localhost:8080/docs/

The API provides comprehensive OpenAPI documentation with interactive examples, authentication flows, and request/response schemas.

### Authentication Flow

All `/api/v1/users/*` and `/api/v1/tasks/*` endpoints require JWT authentication.

#### 1. Register a new account
```bash
curl -X POST http://localhost:8080/api/v1/register \
  -H "Content-Type: application/json" \
  -d '{
    "username": "john_doe",
    "email": "john@example.com",
    "password": "securepass123",
    "first_name": "John",
    "last_name": "Doe"
  }'
```

#### 2. Login to get a JWT token
```bash
curl -X POST http://localhost:8080/api/v1/login \
  -H "Content-Type: application/json" \
  -d '{
    "identifier": "john_doe",
    "password": "securepass123"
  }'
```

Save the `token` from the response.

#### 3. Use the token for API calls
```bash
# Get current user profile
curl -H "Authorization: Bearer YOUR_TOKEN" \
  http://localhost:8080/api/v1/users/me

# Update current user profile
curl -X PUT http://localhost:8080/api/v1/users/me \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "first_name": "John Updated",
    "last_name": "Doe"
  }'

# Create a task
curl -X POST http://localhost:8080/api/v1/tasks/ \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "My Task",
    "description": "Task description",
    "status": "pending"
  }'

# Get all user tasks
curl -H "Authorization: Bearer YOUR_TOKEN" \
  http://localhost:8080/api/v1/tasks/

# Update a task
curl -X PUT http://localhost:8080/api/v1/tasks/1/ \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "status": "completed"
  }'
```

#### 4. Logout (invalidates token)
```bash
curl -X POST http://localhost:8080/api/v1/logout \
  -H "Authorization: Bearer YOUR_TOKEN"
```

## API Endpoints

📖 **Complete documentation with examples:** http://localhost:8080/docs/

### Public Endpoints (no authentication required)
- `GET /api/v1/welcome` - Welcome message and health check
- `POST /api/v1/register` - User registration
- `POST /api/v1/login` - User authentication

### Protected Endpoints (JWT token required)

#### Authentication
- `POST /api/v1/logout` - User logout (invalidates token)

#### User Management
- `GET /api/v1/users/me` - Get current user profile
- `PUT /api/v1/users/me` - Update current user profile
- `DELETE /api/v1/users/me` - Delete current user account

#### Task Management (user-scoped)
- `GET /api/v1/tasks/` - List user's tasks
- `POST /api/v1/tasks/` - Create new task
- `GET /api/v1/tasks/{id}/` - Get specific task
- `PUT /api/v1/tasks/{id}/` - Update task
- `DELETE /api/v1/tasks/{id}/` - Delete task

## Configuration

### Environment Variables

#### Core Configuration
- `STORAGE_BACKEND` - "memory" (default) or "postgres"
- `PORT` - Server port (default: 8080)
- `GO_ENV` - "development" or "production"

#### Authentication
- `JWT_SECRET` - JWT signing secret (⚠️ **required for production**)
- `JWT_EXPIRATION` - Token expiration duration (default: 24h)

#### PostgreSQL Configuration (when using postgres backend)
- `PG_HOST` - PostgreSQL host (default: localhost)
- `PG_PORT` - PostgreSQL port (default: 5432)
- `PG_USER` - Database user
- `PG_PASSWORD` - Database password
- `PG_DBNAME` - Database name

### Example Configuration

```bash
# Development (memory backend)
export STORAGE_BACKEND=memory
export JWT_SECRET=your-secret-key

# Production (PostgreSQL backend)
export STORAGE_BACKEND=postgres
export JWT_SECRET=your-strong-secret-key
export PG_HOST=localhost
export PG_PORT=5432
export PG_USER=voidrunner
export PG_PASSWORD=password
export PG_DBNAME=voidrunner
```

## Containerization

### Docker Compose (Recommended)

```bash
# Start application with PostgreSQL database
make docker-compose-up

# Start in detached mode
make docker-compose-up-detached

# View logs
make docker-compose-logs

# Stop services
make docker-compose-down
```

### Docker Only

```bash
# Build Docker image
make docker-build

# Run with memory backend
make docker-run

# Run tests in container
make docker-test
```

## Database Setup

### PostgreSQL with Docker Compose
```bash
# Start PostgreSQL and application together
make docker-compose-up
```

### PostgreSQL with Local Development
```bash
# Start PostgreSQL container only
make db-up

# Run database migrations
make db-migrate-up

# Run application with PostgreSQL (includes above steps)
make run-postgres
```

### Manual Migration Commands
```bash
# Install golang-migrate tool
brew install golang-migrate

# Run all migrations
migrate -path db/migrations -database "postgres://user:pass@localhost:5432/dbname?sslmode=disable" up

# Rollback last migration
migrate -path db/migrations -database "postgres://user:pass@localhost:5432/dbname?sslmode=disable" down 1

# Reset database (drop and recreate)
make db-migrate-reset
```

## Development

### Building and Testing

```bash
# Run all tests (unit, integration, E2E)
make test-all

# Run specific test types
make test                # Unit tests only
make test-integration    # Integration tests
make test-e2e           # End-to-end tests

# Build application
make build

# Clean build artifacts
make clean
```

### API Documentation

```bash
# Generate OpenAPI documentation
make docs-generate

# Validate documentation format
make docs-validate

# Start server with documentation at /docs
make docs-serve

# Clean generated docs
make docs-clean
```

### Database Commands

```bash
make db-up          # Start PostgreSQL container
make db-down        # Stop PostgreSQL container
make db-migrate-up  # Run database migrations
make db-migrate-down # Rollback migrations
make db-migrate-reset # Reset database (drop and recreate)
```

### Docker Commands

```bash
make docker-build              # Build Docker image
make docker-run               # Run in container (memory backend)
make docker-compose-up        # Start with docker-compose
make docker-compose-down      # Stop docker-compose services
make docker-test             # Run tests in container
make docker-clean            # Clean Docker artifacts
```

### Complete Development Workflow

```bash
# 1. Start services with docker-compose
make docker-compose-up

# 2. View API documentation
open http://localhost:8080/docs/

# 3. Run tests
make test-all

# 4. Generate/update documentation
make docs-generate

# 5. Clean up
make docker-compose-down
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
