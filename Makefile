.PHONY: help test build tidy fmt vet lint clean check setup-hooks pre-commit

# Default target
help: ## Show this help message
	@echo "Available commands:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-15s\033[0m %s\n", $$1, $$2}'

test: ## Run all tests
	@echo "Running tests..."
	cd shared && go test ./...
	cd services/block-synchronizer && go test ./...
	cd services/event-processor && go test ./...

test-v: ## Run all tests with verbose output
	@echo "Running tests with verbose output..."
	cd shared && go test -v ./...
	cd services/block-synchronizer && go test -v ./...
	cd services/event-processor && go test -v ./...

build: ## Build all services
	@echo "Building services..."
	@mkdir -p bin
	cd services/block-synchronizer && go build -o ../../bin/block-synchronizer .
	cd services/event-processor && go build -o ../../bin/event-processor .

tidy: ## Run go mod tidy for all modules
	@echo "Running go mod tidy..."
	go work sync
	cd shared && go mod tidy
	cd services/block-synchronizer && go mod tidy
	cd services/event-processor && go mod tidy

fmt: ## Format all Go files
	@echo "Formatting code..."
	cd shared && go fmt ./...
	cd services/block-synchronizer && go fmt ./...
	cd services/event-processor && go fmt ./...

clean: ## Clean build artifacts and test cache
	@echo "Cleaning..."
	rm -rf bin/
	go clean -testcache
	go clean -cache

setup-hooks: ## Install pre-commit hooks
	@echo "Installing pre-commit hooks..."
	pre-commit install

pre-commit: ## Run pre-commit on all files
	@echo "Running pre-commit checks..."
	pre-commit run --all-files
