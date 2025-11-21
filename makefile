.PHONY: help build run test test-integration clean docker-build docker-run docker-stop

help: ## Show this help message
	@echo 'Usage: make [target]'
	@echo ''
	@echo 'Available targets:'
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-20s\033[0m %s\n", $$1, $$2}'

build: ## Build the application
	@echo "🔨 Building application..."
	go build -o bin/banking-api ./cmd/api
	@echo "✅ Build complete: bin/banking-api"

run: ## Run the application locally
	@echo "🚀 Starting application..."
	go run cmd/api/*.go

test: ## Run unit tests
	@echo "🧪 Running unit tests..."
	go test ./... -v -cover

test-integration: ## Run integration tests (requires server running)
	@echo "🧪 Running integration tests..."
	@if ! curl -s http://localhost:8080/health > /dev/null 2>&1; then \
		echo "❌ Server is not running. Please run 'make run' first."; \
		exit 1; \
	fi
	@echo "✅ Server is running"
	@chmod +x run-local/test-simple.sh
	@./run-local/test-simple.sh

clean: ## Remove build artifacts and database
	@echo "🧹 Cleaning up..."
	rm -rf bin/
	rm -rf data/
	rm -f *.log test-results.log server.log
	@echo "✅ Cleanup completed"

clean-db: ## Clean only database files
	@echo "🗑️  Cleaning database..."
	rm -rf data/
	@echo "✅ Database cleaned"

docker-build: ## Build Docker image
	@echo "🐳 Building Docker image..."
	docker build -t simple-banking-api:latest .
	@echo "✅ Docker image built"

docker-run: ## Run application in Docker
	@echo "🐳 Starting application in Docker..."
	docker-compose up -d
	@echo "✅ Application running on http://localhost:8080"

docker-stop: ## Stop Docker containers
	@echo "🛑 Stopping Docker containers..."
	docker-compose down
	@echo "✅ Containers stopped"

docker-logs: ## Show Docker logs
	docker-compose logs -f

docker-restart: ## Restart Docker containers
	@$(MAKE) docker-stop
	@$(MAKE) docker-run
