.PHONY: help run build test test-race lint fmt infra-up infra-down docker-up docker-down clean docker-logs install-tools setup

# Variables
APP_NAME=tourneyrank
MAIN_PATH=./cmd/service

# Docker
DOCKER_COMPOSE=docker compose

help: ## Show this help message
	@echo 'Usage: make [target]'
	@echo ''
	@echo 'Available targets:'
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  %-15s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

run: ## Run the application locally (requires infra-up)
	HTTP_PORT=8080 MONGODB_URI="mongodb://tourneyrank:tourneyrank@localhost:27017/tourneyrank?authSource=admin" MONGODB_DATABASE=tourneyrank REDIS_URL=redis://localhost:6379 go run $(MAIN_PATH)/main.go

build: ## Build the application binary
	go build -o bin/$(APP_NAME) $(MAIN_PATH)/main.go

test: ## Run unit tests
	go test ./... -v -cover

test-race: ## Run tests with race detector
	go test ./... -race -coverprofile=coverage.out

coverage: test-race ## Generate coverage report
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

lint: ## Run linter (golangci-lint required)
	golangci-lint run --enable-all --disable exhaustivestruct,exhaustruct,gochecknoglobals

fmt: ## Format Go code
	gofmt -s -w .
	goimports -w .

vet: ## Run go vet
	go vet ./...

tidy: ## Tidy go modules
	go mod tidy

infra-up: ## Start infrastructure (MongoDB, Redis)
	$(DOCKER_COMPOSE) up -d mongodb redis

infra-down: ## Stop infrastructure
	$(DOCKER_COMPOSE) stop mongodb redis

docker-up: ## Start everything with Docker
	$(DOCKER_COMPOSE) up --build

docker-down: ## Stop everything
	$(DOCKER_COMPOSE) down

clean: ## Clean build artifacts
	rm -rf bin/ coverage.out coverage.html

docker-logs: ## Show Docker logs
	$(DOCKER_COMPOSE) logs -f

install-tools: ## Install development tools
	@echo "Installing development tools..."
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	go install golang.org/x/tools/cmd/goimports@latest

setup: install-tools infra-up ## Setup development environment
	@echo "Development environment ready! Run 'make run' to start the backend."

.DEFAULT_GOAL := help
