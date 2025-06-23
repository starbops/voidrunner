# End-to-End (E2E) Tests

This directory contains end-to-end tests for the VoidRunner API that validate complete user workflows and cross-backend compatibility.

## Overview

E2E tests differ from integration tests in that they:
- Use real HTTP server instances (not test servers)
- Test complete user journeys from client perspective
- Validate cross-backend compatibility (memory vs PostgreSQL)
- Test data persistence across server restarts
- Validate real authentication flows with JWT tokens

## Test Structure

### Test Files

- **`helpers.go`**: E2E test infrastructure and utilities
- **`app_test.go`**: Full application E2E tests
- **`workflow_test.go`**: User workflow E2E tests
- **`backend_test.go`**: Cross-backend compatibility tests

### Test Categories

#### 1. Full Application E2E Tests (`app_test.go`)
- Complete user workflow: Registration → Login → Task CRUD → Logout
- Server health and endpoint validation
- Authentication lifecycle testing
- Real HTTP client-server communication

#### 2. User Workflow E2E Tests (`workflow_test.go`)
- Task management workflows (create → update → complete → delete)
- Concurrent user operations
- Error scenario testing
- Data consistency validation

#### 3. Backend Compatibility E2E Tests (`backend_test.go`)
- Memory backend functionality testing
- PostgreSQL backend functionality testing
- Cross-backend consistency validation
- Data persistence verification

## Running E2E Tests

### Prerequisites

For PostgreSQL backend tests:
- PostgreSQL server running on localhost:5432
- Database user `voidrunner` with password `password`
- Permission to create/drop databases

### Local Testing

```bash
# Run all E2E tests
make test-e2e

# Run specific test file
go test -v ./test/e2e/app_test.go ./test/e2e/helpers.go

# Run with specific backend
STORAGE_BACKEND=memory go test -v ./test/e2e/...
STORAGE_BACKEND=postgres go test -v ./test/e2e/...
```

### Environment Variables

- `TEST_PG_HOST`: PostgreSQL host (default: localhost)
- `TEST_PG_PORT`: PostgreSQL port (default: 5432)
- `TEST_PG_USER`: PostgreSQL user (default: voidrunner)
- `TEST_PG_PASSWORD`: PostgreSQL password (default: password)
- `TEST_PG_DBNAME`: Base name for test databases (default: voidrunner_test)

## Test Infrastructure

### E2ETestHelper

The `E2ETestHelper` provides:
- Real server process management
- Dynamic port allocation
- Database setup/cleanup
- HTTP client utilities
- Authentication helpers

### Server Management

- Builds the application binary
- Starts real server instances
- Uses dynamic port allocation to avoid conflicts
- Properly terminates server processes

### Database Management

- Creates isolated test databases
- Runs migrations programmatically
- Cleans up test data between tests
- Handles database connection lifecycle

## Test Isolation

Each test:
- Uses a unique database (for PostgreSQL tests)
- Uses a unique server port
- Registers unique users to avoid conflicts
- Cleans up resources after completion

## Troubleshooting

### Common Issues

1. **Server startup failures**: Check that the application builds correctly
2. **Database connection errors**: Verify PostgreSQL is running and credentials are correct
3. **Port conflicts**: Tests use dynamic port allocation, but ensure no other services conflict
4. **Authentication failures**: Verify JWT tokens are properly generated and used

### Debug Mode

Add logging to tests:
```go
t.Logf("Server URL: %s", helper.ServerURL)
t.Logf("Backend: %s", helper.BackendType)
```

### Logs

Server logs are not captured by default. To see server output, modify the helper to redirect stdout/stderr.

## CI/CD Integration

E2E tests are integrated into GitHub Actions workflows:
- Run on both push and pull request events
- Test both memory and PostgreSQL backends
- Use PostgreSQL service containers for database tests
- Collect test artifacts on failure

## Performance Considerations

E2E tests are slower than unit/integration tests because they:
- Start real server processes
- Use real database connections
- Perform full HTTP request/response cycles
- Include network latency

Consider running E2E tests separately or in parallel to optimize CI/CD pipeline performance.