.PHONY: build
build:
	@echo "Building..."
	@go build -o ./bin/voidrunner ./cmd/main.go

.PHONY: test
test:
	@echo "Running tests..."
	@go test -v ./...

.PHONY: run
run: build
	@echo "Running the application..."
	@./bin/voidrunner

.PHONY: clean
clean:
	@echo "Cleaning up..."
	@rm -rf ./bin
