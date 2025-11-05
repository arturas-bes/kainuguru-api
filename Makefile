.PHONY: help install format run test scrape clean

# Default target
help:
	@echo "🍎 Kainuguru API - Available Commands:"
	@echo "=================================="
	@echo "  install      - Spin up Docker development environment"
	@echo "  format       - Clean up and format code"
	@echo "  run          - Run API server locally"
	@echo "  test         - Run all tests"
	@echo "  scrape       - Run scraper tests and integrations"
	@echo "  clean        - Stop containers and clean up"
	@echo ""
	@echo "Quick start: make install && make run"

# 🚀 MAIN COMMANDS
# These are the primary commands the user requested

install:
	@echo "🐳 Spinning up Docker development environment..."
	@docker-compose down --remove-orphans 2>/dev/null || true
	@docker-compose up -d
	@echo "✅ Development environment ready!"
	@echo "   Database: postgres://kainuguru_user:kainuguru_pass@localhost:5432/kainuguru"
	@echo "   Redis: redis://localhost:6379"
	@echo "   API will start once containers are healthy"

format:
	@echo "🧹 Cleaning up and formatting code..."
	@go fmt ./...
	@go mod tidy
	@echo "✅ Code formatted and dependencies cleaned!"

run:
	@echo "🚀 Running API server locally..."
	@go run cmd/api/main.go

test:
	@echo "🧪 Running all tests..."
	@go test -v ./... -race
	@echo "✅ All tests completed!"

scrape:
	@echo "🤖 Running scraper tests and integrations..."
	@echo ""
	@echo "1. Testing IKI Scraper (Local)..."
	@go run cmd/test-scraper/main.go
	@echo ""
	@echo "2. Testing Full Pipeline (Docker - includes PDF processing)..."
	@docker run --rm -v $(PWD):/app -w /app golang:1.24-alpine sh -c 'apk add --no-cache poppler-utils imagemagick && go run cmd/test-full-pipeline/main.go'
	@echo ""
	@echo "✅ Scraper integration tests completed!"

clean:
	@echo "🧹 Stopping containers and cleaning up..."
	@docker-compose down --remove-orphans --volumes
	@docker system prune -f 2>/dev/null || true
	@rm -rf test_output/ coverage.out coverage.html bin/
	@echo "✅ Environment cleaned!"

# 🔧 HELPER COMMANDS (internal use)

_test-unit:
	@echo "Running unit tests..."
	@go test -v ./... -tags=unit -short

_test-integration:
	@echo "Running integration tests..."
	@go test -v ./tests/... -tags=integration

_build:
	@echo "Building binaries..."
	@mkdir -p bin/
	@go build -o bin/api cmd/api/main.go
	@go build -o bin/scraper cmd/scraper/main.go
	@go build -o bin/migrator cmd/migrator/main.go

_docker-logs:
	@echo "Showing Docker logs..."
	@docker-compose logs -f

_status:
	@echo "Docker container status:"
	@docker-compose ps