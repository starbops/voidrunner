.PHONY: build
build:
	@echo "Building..."
	@go build -o ./bin/voidrunner ./cmd/main.go

.PHONY: test
test:
	@echo "Running unit tests..."
	@go test -cover $(shell go list ./... | grep -v test/integration)

.PHONY: test-integration
test-integration:
	@echo "Running integration tests..."
	@go test -v ./test/integration/...

.PHONY: test-all
test-all: test test-integration

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
