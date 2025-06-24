# Core build targets
.PHONY: build clean run
build:
	@echo "Building..."
	@go build -o ./bin/voidrunner ./cmd/main.go

clean:
	@echo "Cleaning up..."
	@rm -rf ./bin

run: build
	@echo "Running the application..."
	@./bin/voidrunner

# Test targets
.PHONY: test test-integration test-e2e test-all
test:
	@echo "Running unit tests..."
	@go test -cover $(shell go list ./... | grep -v test/integration | grep -v test/e2e)

test-integration:
	@echo "Running integration tests..."
	@go test -v ./test/integration/...

test-e2e:
	@echo "Running E2E tests..."
	@go test -v ./test/e2e/...

test-all: test test-integration test-e2e

# Database targets
.PHONY: db-up db-down db-migrate db-reset run-postgres
db-up:
	@echo "Starting PostgreSQL..."
	@docker-compose up -d postgres
	@echo "Waiting for PostgreSQL to be ready..."
	@until docker-compose exec -T postgres pg_isready -U voidrunner -d voidrunner >/dev/null 2>&1; do \
		echo "PostgreSQL is not ready yet, waiting..."; \
		sleep 2; \
	done
	@echo "PostgreSQL is ready!"

db-down:
	@echo "Stopping PostgreSQL..."
	@docker-compose down

db-migrate:
	@echo "Running migrations..."
	@migrate -path db/migrations -database "postgres://voidrunner:password@localhost:5432/voidrunner?sslmode=disable" up

db-reset:
	@echo "Resetting database..."
	@migrate -path db/migrations -database "postgres://voidrunner:password@localhost:5432/voidrunner?sslmode=disable" drop
	@migrate -path db/migrations -database "postgres://voidrunner:password@localhost:5432/voidrunner?sslmode=disable" up

run-postgres: db-up db-migrate build
	@echo "Running with PostgreSQL..."
	@STORAGE_BACKEND=postgres PG_HOST=localhost PG_PORT=5432 PG_USER=voidrunner PG_PASSWORD=password PG_DBNAME=voidrunner ./bin/voidrunner

# Docker targets  
.PHONY: docker-build docker-run docker-up docker-down docker-logs docker-test
docker-build:
	@echo "Building Docker image..."
	@docker build -t voidrunner:latest .

docker-run: docker-build
	@echo "Running Docker container..."
	@docker run -p 8080:8080 --env STORAGE_BACKEND=memory voidrunner:latest

docker-up:
	@echo "Starting services with docker-compose..."
	@docker-compose up --build $(if $(DETACH),-d,)

docker-down:
	@echo "Stopping docker-compose services..."
	@docker-compose down

docker-logs:
	@echo "Showing docker-compose logs..."
	@docker-compose logs -f

docker-test:
	@echo "Running tests in Docker container..."
	@docker build --target builder -t voidrunner-test .
	@docker run --rm voidrunner-test go test -cover ./...

# Documentation targets
.PHONY: docs docs-clean
docs:
	@echo "Generating OpenAPI documentation..."
	@swag init -g cmd/main.go -o docs
	@swag fmt

docs-clean:
	@echo "Cleaning generated documentation..."
	@rm -rf docs/
