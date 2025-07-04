.PHONY: build run test clean migrate lint

# Variables
BINARY_NAME=smart_redirect
MAIN_PATH=cmd/server/main.go

# Build the application
build:
	go build -o $(BINARY_NAME) $(MAIN_PATH)

# Run the application
run:
	go run $(MAIN_PATH)

# Run tests
test:
	go test -v ./...

# Run tests with coverage
test-coverage:
	go test -v -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html

# Run unit tests only
test-unit:
	go test -v ./test/unit/...

# Run integration tests only
test-integration:
	go test -v ./test/integration/...

# Run all tests with detailed reporting
test-all:
	./scripts/run_tests.sh

# Run tests with load testing
test-load:
	./scripts/run_tests.sh --load-tests

# Run benchmark tests
bench:
	go test -bench=. -benchmem ./...

# Clean build artifacts
clean:
	rm -f $(BINARY_NAME)
	rm -f coverage.out coverage.html

# Run database migrations
migrate-up:
	go run scripts/migrate.go up

migrate-down:
	go run scripts/migrate.go down

# Run linter
lint:
	golangci-lint run

# Install dependencies
deps:
	go mod download
	go mod tidy

# Development mode with hot reload
dev:
	air -c .air.toml

# Docker commands
docker-build:
	docker build -t smart-redirect:latest .

docker-run:
	docker-compose up -d

docker-stop:
	docker-compose down