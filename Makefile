.PHONY: help build run test test-unit test-coverage test-verbose test-race clean docker-build docker-run

help: ## Show this help message
	@echo 'Usage: make [target]'
	@echo ''
	@echo 'Available targets:'
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  %-20s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

build: ## Build the Go binary
	@echo "Building Go application..."
	go build -o bin/server ./cmd/server

run: ## Run the application locally
	@echo "Running application..."
	go run ./cmd/server

test: ## Run all tests
	@echo "Running all tests..."
	@chmod +x run_tests.sh
	@./run_tests.sh all

test-unit: ## Run unit tests only
	@echo "Running unit tests..."
	@chmod +x run_tests.sh
	@./run_tests.sh unit

test-coverage: ## Run tests with coverage report
	@echo "Running tests with coverage..."
	@chmod +x run_tests.sh
	@./run_tests.sh coverage

test-verbose: ## Run tests in verbose mode
	@echo "Running tests in verbose mode..."
	@chmod +x run_tests.sh
	@./run_tests.sh verbose

test-race: ## Run tests with race detection
	@echo "Running tests with race detection..."
	go test -race ./...

test-bench: ## Run benchmark tests
	@echo "Running benchmark tests..."
	@chmod +x run_tests.sh
	@./run_tests.sh bench

clean: ## Clean build artifacts
	@echo "Cleaning..."
	rm -rf bin/
	rm -f coverage.out coverage.html
	go clean

docker-build: ## Build Docker image
	@echo "Building Docker image..."
	docker build -f Dockerfile -t document-smbrelay:latest .

docker-run: ## Run Docker container
	@echo "Running Docker container..."
	docker run --rm -p 8080:8080 \
		-e SMB_SERVER_NAME=testserver \
		-e SMB_SERVER_IP=127.0.0.1 \
		-e SMB_SHARE_NAME=testshare \
		-e SMB_USERNAME=testuser \
		-e SMB_PASSWORD=testpass \
		-e LOG_LEVEL=DEBUG \
		document-smbrelay:latest

deps: ## Download dependencies
	@echo "Downloading dependencies..."
	go mod download
	go mod tidy

fmt: ## Format code
	@echo "Formatting code..."
	go fmt ./...

vet: ## Run go vet
	@echo "Running go vet..."
	go vet ./...

lint: ## Run golangci-lint (if installed)
	@echo "Running golangci-lint..."
	golangci-lint run

check: fmt vet lint test ## Run format, vet, and tests

ci: deps fmt vet test-coverage ## Run full CI pipeline locally
