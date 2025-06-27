# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Build Commands

### Core Build Commands
- `make build` - Build the application to `./bin/voidrunner`
- `make clean` - Remove build artifacts from `./bin`
- `make run` - Build and run the application (starts server on :8080)

### Test Commands
- `make test` - Run unit tests with coverage
- `make test-integration` - Run integration tests (uses -count=1 flag to prevent caching)
- `make test-e2e` - Run end-to-end tests (use STORAGE_BACKEND env var for backend selection)
- `make test-all` - Run all test suites (unit, integration, E2E)
- `go test -v ./test/integration/...` - Run integration tests directly
- `go test -v ./test/e2e/...` - Run E2E tests directly
- `STORAGE_BACKEND=postgres make test-e2e` - Run E2E tests with specific backend

### Database Commands
- `make db-up` - Start PostgreSQL container with docker-compose
- `make db-down` - Stop PostgreSQL container
- `make db-migrate` - Run database migrations
- `make db-reset` - Reset database (drop all tables and re-migrate)
- `make run-postgres` - Start database, migrate, and run application with PostgreSQL backend

### Docker Commands
- `make docker-build` - Build Docker image for the application
- `make docker-run` - Build and run application in Docker container (memory backend)
- `make docker-up` - Start all services with docker-compose (use DETACH=1 for detached mode)
- `make docker-down` - Stop all docker-compose services
- `make docker-logs` - Show logs from all services
- `make docker-test` - Run unit tests inside Docker container
- `make docker-test-integration` - Run integration tests with PostgreSQL in Docker
- `make docker-test-e2e` - Run E2E tests with both backends in Docker
- `make docker-test-all` - Run all test suites in Docker containers

### Documentation Commands
- `make docs` - Generate and validate OpenAPI documentation
- `make docs-clean` - Remove generated documentation files

## Architecture Overview

VoidRunner is a Go HTTP API server for multi-user task management with JWT authentication, pluggable storage backends, and containerized deployment.

### Go Version Requirements
- **Minimum**: Go 1.23.10 (for security patches)
- **CI/CD**: All GitHub Actions workflows use Go 1.23.10 to address govulncheck vulnerabilities
- **Security**: Older versions have known vulnerabilities (GO-2025-3751, GO-2025-3750)

### Core Architecture Pattern
The codebase follows a layered architecture with dependency injection:
- **cmd/main.go**: Application entry point with structured JSON logging, initializes repositories via factory pattern
- **cmd/api/**: HTTP server setup and routing with authentication middleware, dependency injection flow
- **internal/handlers/**: HTTP request handlers with OpenAPI annotations and JWT context extraction
- **internal/services/**: Business logic layer with user-scoped operations
- **internal/repositories/**: Data access layer with interface-based factory pattern for backend selection
- **internal/models/**: Domain models with rich type definitions (TaskStatus enum) and API-specific DTOs
- **internal/middleware/**: JWT authentication middleware with context-based user propagation
- **pkg/config/**: Environment-first configuration with production validation
- **pkg/auth/**: JWT token management with thread-safe revocation and user context injection

### User Management & Authentication System
JWT-based authentication with context propagation architecture:
- **Token Management**: Thread-safe JWT generation, validation, and revocation in `pkg/auth`
- **Middleware Integration**: Validates Bearer tokens and injects user context (ID, username, email) into request context
- **User Scoping**: All operations user-scoped via context extraction using `GetUserIDFromContext()`
- **Password Security**: bcrypt hashing with secure defaults
- **Session Management**: Token-based logout with server-side invalidation

### Storage Backend System
Repository factory pattern with runtime backend selection:
- **Memory backend** (default): In-memory storage for development and testing
- **Postgres backend**: Production database with migrations and ACID compliance
- **Factory Pattern**: `NewTaskRepository()` and `NewUserRepository()` functions select backend at runtime
- **Interface Abstraction**: Common repository interfaces enable seamless backend switching
- **User Isolation**: Repository methods support both general queries and user-scoped operations

### Configuration
Environment-based configuration with validation and security warnings:

#### Core Configuration
- `STORAGE_BACKEND`: "memory" (default) or "postgres"
- `PORT`: Server port (default: 8080)
- `GO_ENV`: "development" or "production" (affects logging level)

#### PostgreSQL Configuration (required when using postgres backend)
- `PG_HOST`: PostgreSQL server host
- `PG_PORT`: PostgreSQL server port
- `PG_USER`: Database user
- `PG_PASSWORD`: Database password
- `PG_DBNAME`: Database name

#### Authentication Configuration
- `JWT_SECRET`: Secret key for JWT token signing (REQUIRED in production)
- `JWT_EXPIRATION`: Token expiration duration (default: 24h)

#### Documentation Configuration
- `ENABLE_DOCS`: Enable Swagger documentation endpoint at `/docs/` (default: false)

**SECURITY NOTES**: 
- Always set a strong `JWT_SECRET` in production environments
- Documentation endpoint is disabled by default for security - only enable in development or when needed
- Configuration loader validates required settings and provides production warnings

### API Structure
RESTful API with JSON responses and JWT authentication:

#### Public Endpoints (no authentication required)
- `POST /api/v1/register` - User registration
- `POST /api/v1/login` - User authentication
- `GET /api/v1/welcome` - Welcome message and health check

#### Protected Endpoints (JWT token required)
- `POST /api/v1/logout` - User logout (token invalidation)
- `GET /api/v1/users/me` - Get current user profile
- `PUT /api/v1/users/me` - Update current user profile
- `DELETE /api/v1/users/me` - Delete current user account
- `GET /api/v1/tasks` - List user's tasks
- `POST /api/v1/tasks` - Create new task
- `GET /api/v1/tasks/{id}` - Get specific task
- `PUT /api/v1/tasks/{id}` - Update task
- `DELETE /api/v1/tasks/{id}` - Delete task

### Database Schema
SQL migrations in `db/migrations/` with up/down pattern:
1. **001_create_tasks_table**: Initial task management schema
2. **002_create_users_table**: User management schema
3. **003_add_user_password**: Password field for authentication
4. **004_add_user_id_to_tasks**: User-task relationships with foreign key constraints

### Testing Strategy
Three-tier testing approach with backend compatibility:

#### Test Structure
- **Unit Tests**: Mock-based testing with `sqlmock`, located alongside source files
- **Integration Tests** (`test/integration/`): Real PostgreSQL with `httptest`, database setup/teardown helpers
- **E2E Tests** (`test/e2e/`): Real server process with HTTP client, tests both memory and PostgreSQL backends

#### Key Testing Patterns
- **Backend Agnostic**: E2E tests run against both storage backends automatically
- **Authentication Testing**: Complete JWT flow testing with token validation
- **User Isolation**: Tests verify user-scoped data access across all layers
- **Database Migration**: Integration tests use real migrations and rollback
- **Concurrent Testing**: Goroutine-based tests for authentication and data consistency

### Dependencies and Tools
Key external dependencies:
- **github.com/golang-jwt/jwt/v5**: JWT authentication with standard claims
- **golang.org/x/crypto**: Password hashing (bcrypt) with secure defaults
- **github.com/lib/pq**: PostgreSQL driver for production backend
- **github.com/DATA-DOG/go-sqlmock**: Database testing mocks for unit tests
- **github.com/swaggo/swag**: OpenAPI documentation generation from code annotations

### Development Workflow
1. Use `make docker-up` for full-stack development with PostgreSQL
2. Use `ENABLE_DOCS=true make docker-up` to enable API documentation in development
3. Use `make test-all` to run complete test suite across all backends
4. Use `make docs` to generate and validate OpenAPI documentation
5. Database migrations are automatically applied on startup
6. Integration tests use `-count=1` flag to prevent caching and ensure fresh database state detection

### OpenAPI Documentation Integration
Code-first documentation with automated generation:
- **Swagger Annotations**: Documentation maintained directly in handler code with examples
- **Build Integration**: `make docs` generates and validates documentation from code
- **Security Documentation**: JWT Bearer auth documented in OpenAPI spec with examples
- **Conditional Serving**: Documentation UI at `/docs/` when `ENABLE_DOCS=true`
- **CI/CD Integration**: Documentation generation and validation in GitHub Actions

### Containerization Architecture
Multi-stage Docker builds with production security:
- **Multi-stage Build**: Separate build and runtime containers for optimized image size
- **Security**: Non-root user execution (voidrunner:1001), Alpine Linux base image
- **Development Support**: Local development with `make run` or `make run-postgres` for full PostgreSQL setup
- **Health Monitoring**: Built-in health checks for containerized deployments

### Error Handling and Logging
Structured approach to error management:
- **Structured Logging**: JSON logging with `slog` package, debug level in development
- **Error Responses**: Standardized JSON error responses with appropriate HTTP status codes
- **Authentication Errors**: Clear distinction between unauthorized and forbidden operations
- **Validation Errors**: Detailed error messages for request validation failures

### GitHub Actions CI/CD
Three-tier GitHub Actions workflow structure:
- **quality.yml**: Linting, security scanning (GoSec, govulncheck), and Docker security (Trivy)
- **reusable-test.yml**: Reusable workflow for comprehensive testing (unit, integration, E2E)
- **ci.yml**: Main branch workflow (uploads artifacts, runs Docker tests)
- **pr.yml**: Pull request workflow (lighter testing, no artifact uploads)

### Production Deployment
1. Build with `make docker-build`
2. Deploy with `docker-compose up` using production configuration
3. Ensure `JWT_SECRET` is set to a strong, unique value
4. PostgreSQL backend is recommended for production
5. Monitor health endpoints for service status
6. Documentation endpoint is disabled by default (set `ENABLE_DOCS=true` only if needed)