.PHONY: help build run dev test clean docker-build docker-run lint fmt

# Variables
APP_NAME := relay-api
MAIN_PATH := ./cmd/api/main.go
BINARY_OUTPUT := ./bin/api
GO := go

# Colors for output
BLUE := \033[0;34m
GREEN := \033[0;32m
YELLOW := \033[0;33m
NC := \033[0m # No Color

help: ## Display this help message
	@echo "$(BLUE)Relay Backend - Available Commands$(NC)"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "$(GREEN)%-20s$(NC) %s\n", $$1, $$2}'

dev: ## Run server in development mode with hot reload (requires air)
	@echo "$(YELLOW)Checking for air...$(NC)"
	@which air > /dev/null || (echo "$(BLUE)Installing air (hot reload)...$(NC)" && go install github.com/cosmtrek/air@latest)
	@echo "$(BLUE)Starting development server...$(NC)"
	air

build: ## Build the application binary
	@echo "$(BLUE)Building $(APP_NAME)...$(NC)"
	@mkdir -p bin
	go build -o $(BINARY_OUTPUT) $(MAIN_PATH)
	@echo "$(GREEN)Build complete: $(BINARY_OUTPUT)$(NC)"

run: build ## Build and run the server
	@echo "$(BLUE)Running $(APP_NAME)...$(NC)"
	@$(BINARY_OUTPUT)

test: ## Run tests
	@echo "$(BLUE)Running tests...$(NC)"
	go test -v -cover ./...

test-coverage: ## Run tests with coverage report
	@echo "$(BLUE)Running tests with coverage...$(NC)"
	go test -v -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "$(GREEN)Coverage report: coverage.html$(NC)"

lint: ## Run linters (requires golangci-lint)
	@echo "$(BLUE)Running linters...$(NC)"
	@which golangci-lint > /dev/null || (echo "$(YELLOW)Installing golangci-lint...$(NC)" && go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest)
	golangci-lint run ./...

fmt: ## Format code
	@echo "$(BLUE)Formatting code...$(NC)"
	go fmt ./...
	@echo "$(GREEN)Code formatted$(NC)"

vet: ## Run go vet
	@echo "$(BLUE)Running go vet...$(NC)"
	go vet ./...

tidy: ## Tidy go.mod and go.sum
	@echo "$(BLUE)Tidying dependencies...$(NC)"
	go mod tidy
	@echo "$(GREEN)Dependencies tidied$(NC)"

deps-update: ## Update all dependencies to latest versions
	@echo "$(BLUE)Updating dependencies...$(NC)"
	go get -u ./...
	go mod tidy

clean: ## Clean build artifacts and temporary files
	@echo "$(BLUE)Cleaning...$(NC)"
	rm -rf bin/
	rm -f coverage.out coverage.html
	@echo "$(GREEN)Clean complete$(NC)"

install-tools: ## Install development tools (air, golangci-lint, etc.)
	@echo "$(BLUE)Installing development tools...$(NC)"
	go install github.com/cosmtrek/air@latest
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	@echo "$(GREEN)Tools installed$(NC)"

# Docker targets
docker-build: ## Build Docker image
	@echo "$(BLUE)Building Docker image...$(NC)"
	docker build -t $(APP_NAME):latest .
	@echo "$(GREEN)Docker image built: $(APP_NAME):latest$(NC)"

docker-run: docker-build ## Build and run Docker container
	@echo "$(BLUE)Running Docker container...$(NC)"
	docker run -p 8080:8080 --env-file configs/.env $(APP_NAME):latest

docker-compose-up: ## Run full stack with docker-compose
	@echo "$(BLUE)Starting docker-compose stack...$(NC)"
	docker-compose up

docker-compose-down: ## Stop docker-compose stack
	@echo "$(BLUE)Stopping docker-compose stack...$(NC)"
	docker-compose down

# Database targets
db-migrate-up: ## Run pending migrations (requires migrate CLI)
	@echo "$(BLUE)Running database migrations...$(NC)"
	migrate -path ./migrations -database "$(DATABASE_URL)" up

db-migrate-down: ## Rollback last migration
	@echo "$(BLUE)Rolling back database migration...$(NC)"
	migrate -path ./migrations -database "$(DATABASE_URL)" down 1

# Common workflow targets
setup: tidy install-tools ## Setup project: install dependencies and tools
	@echo "$(GREEN)Setup complete!$(NC)"
	@echo "$(BLUE)Next steps:$(NC)"
	@echo "  1. Copy configs/.env.example to configs/.env"
	@echo "  2. Update configuration in configs/.env"
	@echo "  3. Run 'make dev' to start development server"
