.PHONY: build
build:
	@echo "Building..."
	@go build -o ./bin/voidrunner ./cmd/main.go

.PHONY: test
test:
	@echo "Running unit tests..."
	@go test -cover $(shell go list ./... | grep -v test/integration | grep -v test/e2e)

.PHONY: test-integration
test-integration:
	@echo "Running integration tests..."
	@go test -v ./test/integration/...

.PHONY: test-e2e
test-e2e:
	@echo "Running E2E tests..."
	@go test -v ./test/e2e/...

.PHONY: test-e2e-memory
test-e2e-memory:
	@echo "Running E2E tests with memory backend..."
	@STORAGE_BACKEND=memory go test -v ./test/e2e/...

.PHONY: test-e2e-postgres
test-e2e-postgres:
	@echo "Running E2E tests with PostgreSQL backend..."
	@STORAGE_BACKEND=postgres go test -v ./test/e2e/...

.PHONY: test-all
test-all: test test-integration test-e2e

.PHONY: run
run: build
	@echo "Running the application..."
	@./bin/voidrunner

.PHONY: clean
clean:
	@echo "Cleaning up..."
	@rm -rf ./bin

# Database commands
.PHONY: db-up
db-up:
	@echo "Starting PostgreSQL..."
	@docker-compose up -d

.PHONY: db-down
db-down:
	@echo "Stopping PostgreSQL..."
	@docker-compose down

.PHONY: db-migrate-up
db-migrate-up:
	@echo "Running migrations..."
	@migrate -path db/migrations -database "postgres://voidrunner:password@localhost:5432/voidrunner?sslmode=disable" up

.PHONY: db-migrate-down
db-migrate-down:
	@echo "Rolling back migrations..."
	@migrate -path db/migrations -database "postgres://voidrunner:password@localhost:5432/voidrunner?sslmode=disable" down

.PHONY: db-migrate-reset
db-migrate-reset:
	@echo "Resetting database..."
	@migrate -path db/migrations -database "postgres://voidrunner:password@localhost:5432/voidrunner?sslmode=disable" drop
	@migrate -path db/migrations -database "postgres://voidrunner:password@localhost:5432/voidrunner?sslmode=disable" up

.PHONY: run-postgres
run-postgres: db-up db-migrate-up
	@echo "Running with PostgreSQL..."
	@STORAGE_BACKEND=postgres PG_HOST=localhost PG_PORT=5432 PG_USER=voidrunner PG_PASSWORD=password PG_DBNAME=voidrunner ./bin/voidrunner

# Docker commands
.PHONY: docker-build
docker-build:
	@echo "Building Docker image..."
	@docker build -t voidrunner:latest .

.PHONY: docker-run
docker-run: docker-build
	@echo "Running Docker container..."
	@docker run -p 8080:8080 --env STORAGE_BACKEND=memory voidrunner:latest

.PHONY: docker-compose-up
docker-compose-up:
	@echo "Starting services with docker-compose..."
	@docker-compose up --build

.PHONY: docker-compose-up-detached
docker-compose-up-detached:
	@echo "Starting services with docker-compose (detached)..."
	@docker-compose up --build -d

.PHONY: docker-compose-down
docker-compose-down:
	@echo "Stopping docker-compose services..."
	@docker-compose down

.PHONY: docker-compose-logs
docker-compose-logs:
	@echo "Showing docker-compose logs..."
	@docker-compose logs -f

.PHONY: docker-test
docker-test:
	@echo "Running tests in Docker container..."
	@docker build --target builder -t voidrunner-test .
	@docker run --rm voidrunner-test go test -cover ./...

.PHONY: docker-clean
docker-clean:
	@echo "Cleaning Docker images and containers..."
	@docker system prune -f
	@docker rmi voidrunner:latest voidrunner-test 2>/dev/null || true

# Documentation commands
.PHONY: docs-generate
docs-generate:
	@echo "Generating OpenAPI documentation..."
	@swag init -g cmd/main.go -o docs

.PHONY: docs-validate
docs-validate: docs-generate
	@echo "Validating OpenAPI documentation..."
	@swag fmt

.PHONY: docs-clean
docs-clean:
	@echo "Cleaning generated documentation..."
	@rm -rf docs/

.PHONY: docs-serve
docs-serve: build
	@echo "Starting server with documentation available at http://localhost:8080/docs/"
	@./bin/voidrunner
