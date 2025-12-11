.PHONY: help install format run test scrape clean seed-data db-reset build validate-all validate-database validate-graphql validate-auth validate-search validate-shopping validate-pricing validate-performance validate-quick validate-setup test-snapshots update-snapshots

# Default target
help:
	@echo "🍎 Kainuguru API - Available Commands:"
	@echo "=================================="
	@echo "🚀 DEVELOPMENT:"
	@echo "  install      - Spin up Docker development environment"
	@echo "  seed-data    - Load test fixtures into database (unified data population)"
	@echo "  db-reset     - Reset database and reload all fixtures"
	@echo "  format       - Clean up and format code"
	@echo "  build        - Build all binaries"
	@echo "  run          - Run API server locally"
	@echo "  test         - Run all tests"
	@echo "  test-snapshots    - Run GraphQL snapshot tests only"
	@echo "  update-snapshots  - Update GraphQL snapshot golden files"
	@echo "  clean        - Stop containers and clean up"
	@echo ""
	@echo "🔍 VALIDATION FRAMEWORK:"
	@echo "  validate-all        - Run complete system validation (10-15 min)"
	@echo "  validate-database   - Validate database integrity and schema"
	@echo "  validate-graphql    - Validate GraphQL endpoints and resolvers"
	@echo "  validate-auth       - Validate authentication flows"
	@echo "  validate-search     - Validate Lithuanian search functionality"
	@echo "  validate-shopping   - Validate shopping list operations"
	@echo "  validate-pricing    - Validate price history and trends"
	@echo "  validate-quick      - Quick validation (essential tests)"
	@echo "  validate-setup      - Setup validation framework"
	@echo ""
	@echo "Quick start: make install && make seed-data && make validate-quick"

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

seed-data:
	@echo "📦 Loading test fixtures into database..."
	@echo "Running migrations first..."
	@docker-compose exec -T api go run cmd/migrator/main.go -action=up || true
	@echo "Loading fixtures..."
	@docker exec -e DATABASE_URL="postgres://kainuguru:kainuguru_password@db:5432/kainuguru_db?sslmode=disable" kainuguru-api-api-1 go run tests/scripts/load_complete/load_complete_fixtures.go
	@echo "✅ Test fixtures loaded successfully!"

db-reset:
	@echo "🔄 Resetting database and reloading fixtures..."
	@echo "Stopping containers and removing volumes..."
	@docker-compose down --volumes
	@echo "Starting containers..."
	@docker-compose up -d
	@echo "Waiting for database to be ready..."
	@sleep 15
	@make seed-data
	@echo "✅ Database reset completed!"

build:
	@echo "🔨 Building binaries..."
	@mkdir -p bin/
	@go build -o bin/api cmd/api/main.go
	@go build -o bin/seeder cmd/seeder/main.go
	@go build -o bin/enrich-flyers cmd/enrich-flyers/*.go
	@go build -o bin/archive-flyers cmd/archive-flyers/*.go
	@echo "✅ Binaries built successfully!"

build-enrich:
	@echo "🤖 Building enrichment command..."
	@mkdir -p bin/
	@go build -o bin/enrich-flyers cmd/enrich-flyers/*.go
	@echo "✅ Enrichment command built: bin/enrich-flyers"

build-archive:
	@echo "📦 Building archive command..."
	@mkdir -p bin/
	@go build -o bin/archive-flyers cmd/archive-flyers/*.go
	@echo "✅ Archive command built: bin/archive-flyers"

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

test-snapshots:
	@echo "📸 Running GraphQL snapshot tests..."
	@go test -v ./internal/graphql/resolvers -run Snapshot
	@echo "✅ Snapshot tests passed! Connection payloads are stable."

update-snapshots:
	@echo "🔄 Updating GraphQL snapshot test data..."
	@go test ./internal/graphql/resolvers -run Snapshot -update_graphql_snapshots
	@echo "⚠️  Snapshots updated! Review changes with: git diff internal/graphql/resolvers/testdata/"
	@echo "   Commit only if changes are intentional."

clean:
	@echo "🧹 Stopping containers and cleaning up..."
	@docker-compose down --remove-orphans --volumes
	@docker system prune -f 2>/dev/null || true
	@rm -rf test_output/ coverage.out coverage.html bin/
	@echo "✅ Environment cleaned!"

_test-unit:
	@echo "Running unit tests..."
	@go test -v ./... -tags=unit -short

_test-integration:
	@echo "Running integration tests..."
	@go test -v ./tests/... -tags=integration

_docker-logs:
	@echo "Showing Docker logs..."
	@docker-compose logs -f

_status:
	@echo "Docker container status:"
	@docker-compose ps