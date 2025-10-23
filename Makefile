.PHONY: help run build test test-race lint fmt migrate-up migrate-down docker-up docker-down clean

# Variables
APP_NAME=tourneyrank
MAIN_PATH=./cmd/service
MIGRATION_PATH=./migrations
DATABASE_URL?=postgresql://tourneyrank:tourneyrank@localhost:5432/tourneyrank?sslmode=disable

help: ## Show this help message
	@echo 'Usage: make [target]'
	@echo ''
	@echo 'Available targets:'
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  %-15s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

run: ## Run the application locally
	go run $(MAIN_PATH)/main.go

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

migrate-up: ## Run database migrations up
	@echo "Running migrations..."
	migrate -path $(MIGRATION_PATH) -database "$(DATABASE_URL)" up

migrate-down: ## Rollback last migration
	@echo "Rolling back migration..."
	migrate -path $(MIGRATION_PATH) -database "$(DATABASE_URL)" down 1

migrate-create: ## Create new migration (usage: make migrate-create NAME=create_users)
	@if [ -z "$(NAME)" ]; then echo "NAME is required. Usage: make migrate-create NAME=migration_name"; exit 1; fi
	migrate create -ext sql -dir $(MIGRATION_PATH) -seq $(NAME)

docker-up: ## Start Docker containers
	docker-compose up -d

docker-down: ## Stop Docker containers
	docker-compose down

docker-logs: ## Show Docker logs
	docker-compose logs -f

docker-clean: ## Remove Docker containers and volumes
	docker-compose down -v

clean: ## Clean build artifacts
	rm -rf bin/
	rm -f coverage.out coverage.html

install-tools: ## Install development tools
	@echo "Installing development tools..."
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	go install golang.org/x/tools/cmd/goimports@latest
	go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest

setup: install-tools docker-up migrate-up ## Setup development environment
	@echo "Development environment ready!"

.DEFAULT_GOAL := help
