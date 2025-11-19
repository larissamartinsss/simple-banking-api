.PHONY: help build run test clean docker-build docker-run docker-stop install-deps

# Variables
APP_NAME=banking-api
BINARY_DIR=bin
DOCKER_IMAGE=simple-banking-api
DATA_DIR=./data

help: ## Show this help message
	@echo 'Usage: make [target]'
	@echo ''
	@echo 'Available targets:'
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-20s\033[0m %s\n", $$1, $$2}'

install-deps: ## Install Go dependencies
	@echo "📦 Installing dependencies..."
	go mod download
	go mod tidy

build: ## Build the application
	@echo "🔨 Building application..."
	@mkdir -p $(BINARY_DIR)
	go build -o $(BINARY_DIR)/$(APP_NAME) ./cmd/api
	@echo "✅ Build complete: $(BINARY_DIR)/$(APP_NAME)"

run: build ## Run the application locally
	@echo "🚀 Starting application..."
	@mkdir -p $(DATA_DIR)
	./$(BINARY_DIR)/$(APP_NAME)

test: ## Run tests
	@echo "🧪 Running tests..."
	go test -v -race -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "✅ Tests complete. Coverage report: coverage.html"

clean: ## Clean build artifacts
	@echo "🧹 Cleaning..."
	rm -rf $(BINARY_DIR)
	rm -rf $(DATA_DIR)
	rm -f coverage.out coverage.html
	@echo "✅ Clean complete"

mocks: ## Generate mocks using mockery
	@echo "🔧 Generating mocks..."
	@$(HOME)/go/bin/mockery
	@echo "✅ Mocks generated successfully!"

docker-build: ## Build Docker image
	@echo "🐳 Building Docker image..."
	docker build -t $(DOCKER_IMAGE):latest .
	@echo "✅ Docker image built: $(DOCKER_IMAGE):latest"

docker-run: ## Run Docker container
	@echo "🐳 Starting Docker container..."
	docker-compose up -d
	@echo "✅ Container started. API available at http://localhost:8080"

docker-stop: ## Stop Docker container
	@echo "🛑 Stopping Docker container..."
	docker-compose down
	@echo "✅ Container stopped"

docker-logs: ## View Docker container logs
	docker-compose logs -f

docker-restart: docker-stop docker-run ## Restart Docker container

lint: ## Run linter
	@echo "🔍 Running linter..."
	golangci-lint run ./...

format: ## Format code
	@echo "✨ Formatting code..."
	go fmt ./...
	gofmt -s -w .

all: clean install-deps build test ## Clean, install deps, build, and test
