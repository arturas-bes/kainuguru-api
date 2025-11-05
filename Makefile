.PHONY: help build run test clean migrate-up migrate-down migrate-create docker-up docker-down format lint

# Default target
help:
	@echo "Available commands:"
	@echo "  build        - Build all binaries"
	@echo "  run-api      - Run API server"
	@echo "  run-scraper  - Run scraper worker"
	@echo "  test         - Run all tests"
	@echo "  test-bdd     - Run BDD tests only"
	@echo "  test-unit    - Run unit tests only"
	@echo "  format       - Format code with gofmt"
	@echo "  lint         - Run linters"
	@echo "  clean        - Clean build artifacts"
	@echo "  migrate-up   - Run database migrations up"
	@echo "  migrate-down - Run database migrations down"
	@echo "  migrate-create NAME=<name> - Create new migration"
	@echo "  docker-up    - Start development environment"
	@echo "  docker-down  - Stop development environment"
	@echo "  deps         - Install dependencies"

# Build targets
build: build-api build-scraper build-migrator

build-api:
	@echo "Building API server..."
	@go build -o bin/api cmd/api/main.go

build-scraper:
	@echo "Building scraper worker..."
	@go build -o bin/scraper cmd/scraper/main.go

build-migrator:
	@echo "Building migrator..."
	@go build -o bin/migrator cmd/migrator/main.go

# Run targets
run-api:
	@echo "Starting API server..."
	@go run cmd/api/main.go

run-scraper:
	@echo "Starting scraper worker..."
	@go run cmd/scraper/main.go

# Test targets
test: test-unit test-bdd

test-unit:
	@echo "Running unit tests..."
	@go test -v ./... -tags=unit

test-bdd:
	@echo "Running BDD tests..."
	@go test -v ./tests/bdd/... -tags=bdd

test-coverage:
	@echo "Running tests with coverage..."
	@go test -v -coverprofile=coverage.out ./...
	@go tool cover -html=coverage.out -o coverage.html

# Code quality
format:
	@echo "Formatting code..."
	@go fmt ./...

lint:
	@echo "Running linters..."
	@golangci-lint run

# Dependencies
deps:
	@echo "Installing dependencies..."
	@go mod download
	@go mod tidy

# Clean
clean:
	@echo "Cleaning build artifacts..."
	@rm -rf bin/
	@rm -f coverage.out coverage.html

# Database migrations
migrate-up:
	@echo "Running migrations up..."
	@go run cmd/migrator/main.go up

migrate-down:
	@echo "Running migrations down..."
	@go run cmd/migrator/main.go down

migrate-create:
	@echo "Creating migration: $(NAME)"
	@go run cmd/migrator/main.go create $(NAME)

# Docker commands
docker-up:
	@echo "Starting development environment..."
	@docker-compose up -d

docker-down:
	@echo "Stopping development environment..."
	@docker-compose down

docker-build:
	@echo "Building Docker images..."
	@docker-compose build

docker-logs:
	@echo "Showing Docker logs..."
	@docker-compose logs -f

# GraphQL
generate-graphql:
	@echo "Generating GraphQL code..."
	@go run github.com/99designs/gqlgen generate

# Development shortcuts
dev-api: docker-up run-api

dev-scraper: docker-up run-scraper

# Production builds
build-prod:
	@echo "Building production binaries..."
	@CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o bin/api-prod cmd/api/main.go
	@CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o bin/scraper-prod cmd/scraper/main.go
	@CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o bin/migrator-prod cmd/migrator/main.go

# Install tools
install-tools:
	@echo "Installing development tools..."
	@go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	@go install github.com/99designs/gqlgen@latest
	@go install github.com/pressly/goose/v3/cmd/goose@latest