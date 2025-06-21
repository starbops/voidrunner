# voidrunner

[![CI](https://github.com/starbops/voidrunner/actions/workflows/ci.yml/badge.svg)](https://github.com/starbops/voidrunner/actions/workflows/ci.yml)

A Go HTTP API server for task and user management with JWT authentication.

## Quick Start

```bash
# Build and run
make run

# Or build and run manually
make build
./bin/voidrunner
```

Server runs on `http://localhost:8080`

## Authentication

All `/api/v1/users/*` and `/api/v1/tasks/*` endpoints require authentication.

### 1. Register a new account
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

### 2. Login to get a token
```bash
curl -X POST http://localhost:8080/api/v1/login \
  -H "Content-Type: application/json" \
  -d '{
    "identifier": "john_doe",
    "password": "securepass123"
  }'
```

Save the `token` from the response.

### 3. Use the token for API calls
```bash
# Get all users
curl -H "Authorization: Bearer YOUR_TOKEN" \
  http://localhost:8080/api/v1/users/

# Create a task
curl -X POST http://localhost:8080/api/v1/tasks/ \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "My Task",
    "description": "Task description"
  }'

# Get all tasks
curl -H "Authorization: Bearer YOUR_TOKEN" \
  http://localhost:8080/api/v1/tasks/
```

### 4. Logout (invalidates token)
```bash
curl -X POST http://localhost:8080/api/v1/logout \
  -H "Authorization: Bearer YOUR_TOKEN"
```

## API Endpoints

**Public (no auth required):**
- `GET /api/v1/welcome` - Welcome message
- `POST /api/v1/register` - User registration  
- `POST /api/v1/login` - User login

**Protected (auth required):**
- `POST /api/v1/logout` - User logout
- `/api/v1/users/*` - User management (CRUD)
- `/api/v1/tasks/*` - Task management (CRUD)

## Configuration

Set environment variables:
- `JWT_SECRET` - JWT signing secret (required for production)
- `STORAGE_BACKEND` - "memory" (default) or "postgres"
- PostgreSQL config: `PG_HOST`, `PG_PORT`, `PG_USER`, `PG_PASSWORD`, `PG_DBNAME`

## Development

```bash
# Run tests
make test

# Build
make build

# Clean
make clean
```
