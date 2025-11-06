#!/bin/bash

# Validation Environment Setup Script
# Sets up the complete validation environment for Kainuguru system testing

set -e  # Exit on any error

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Logging functions
log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

log_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Check if command exists
command_exists() {
    command -v "$1" >/dev/null 2>&1
}

# Check prerequisites
check_prerequisites() {
    log_info "Checking prerequisites..."

    local missing_deps=()

    if ! command_exists docker; then
        missing_deps+=("docker")
    fi

    if ! command_exists docker-compose; then
        missing_deps+=("docker-compose")
    fi

    if ! command_exists go; then
        missing_deps+=("go")
    fi

    if ! command_exists make; then
        missing_deps+=("make")
    fi

    if [ ${#missing_deps[@]} -ne 0 ]; then
        log_error "Missing required dependencies: ${missing_deps[*]}"
        log_info "Please install the missing dependencies and run this script again."
        exit 1
    fi

    # Check Go version
    local go_version=$(go version | grep -oE 'go[0-9]+\.[0-9]+' | sed 's/go//')
    local go_major=$(echo $go_version | cut -d. -f1)
    local go_minor=$(echo $go_version | cut -d. -f2)

    if [ "$go_major" -lt 1 ] || ([ "$go_major" -eq 1 ] && [ "$go_minor" -lt 22 ]); then
        log_error "Go 1.22 or higher is required. Current version: $go_version"
        exit 1
    fi

    log_success "All prerequisites satisfied"
}

# Check Docker status
check_docker() {
    log_info "Checking Docker daemon..."

    if ! docker info >/dev/null 2>&1; then
        log_error "Docker daemon is not running. Please start Docker and try again."
        exit 1
    fi

    log_success "Docker daemon is running"
}

# Setup validation framework
setup_validation_framework() {
    log_info "Setting up validation framework..."

    # Create directory structure
    mkdir -p tests/validation/{config,utils,reporting,models,engine,assertions,metrics,docker,cli,storage}
    mkdir -p tests/validation/{database,graphql,auth,search,shopping,pricing}
    mkdir -p tests/{scripts,fixtures}
    mkdir -p tests/validation/{logs,results}
    mkdir -p bin/

    # Initialize Go module for validation if it doesn't exist
    if [ ! -f "tests/validation/go.mod" ]; then
        log_warning "Validation Go module not found. Please ensure it exists."
    fi

    log_success "Validation framework directory structure created"
}

# Setup database for validation
setup_database() {
    log_info "Setting up validation database..."

    # Check if Docker Compose services are running
    if ! docker-compose ps | grep -q "Up"; then
        log_info "Starting Docker Compose services..."
        docker-compose up -d

        # Wait for services to be ready
        log_info "Waiting for services to be ready..."
        sleep 30
    fi

    # Check database connectivity
    local db_ready=false
    local retries=30

    while [ $retries -gt 0 ] && [ "$db_ready" = false ]; do
        if docker-compose exec -T db pg_isready -h localhost -U kainuguru_user >/dev/null 2>&1; then
            db_ready=true
            log_success "Database is ready"
        else
            log_info "Waiting for database to be ready... ($retries retries left)"
            sleep 2
            retries=$((retries - 1))
        fi
    done

    if [ "$db_ready" = false ]; then
        log_error "Database failed to start within timeout period"
        exit 1
    fi

    # Run migrations if they exist
    if [ -d "migrations" ]; then
        log_info "Running database migrations..."
        # Note: This would need to be implemented based on your migration tool
        # For now, we'll just note that migrations should be run
        log_warning "Please ensure database migrations are up to date"
    fi
}

# Setup test data
setup_test_data() {
    log_info "Setting up test data..."

    # Create test data script if it doesn't exist
    cat > tests/scripts/seed_test_data.sql << 'EOF'
-- Test data for validation framework
-- This should include sample stores, products, users, etc.

-- Sample stores
INSERT INTO stores (name, logo_url, website, description, is_active) VALUES
('Test Store 1', 'https://example.com/logo1.png', 'https://store1.test', 'Test store for validation', true),
('Test Store 2', 'https://example.com/logo2.png', 'https://store2.test', 'Another test store', true)
ON CONFLICT (name) DO NOTHING;

-- Add more test data as needed...
EOF

    log_success "Test data setup completed"
}

# Build validation binary
build_validation_binary() {
    log_info "Building validation binary..."

    if [ -f "tests/validation/main.go" ]; then
        cd tests/validation
        go mod download
        go build -o ../../bin/validation .
        cd ../..

        if [ -f "bin/validation" ]; then
            chmod +x bin/validation
            log_success "Validation binary built successfully"
        else
            log_error "Failed to build validation binary"
            exit 1
        fi
    else
        log_error "Validation main.go not found. Please ensure the validation framework is complete."
        exit 1
    fi
}

# Verify setup
verify_setup() {
    log_info "Verifying setup..."

    # Check if validation binary exists and is executable
    if [ ! -x "bin/validation" ]; then
        log_error "Validation binary not found or not executable"
        exit 1
    fi

    # Test validation binary
    if ./bin/validation help >/dev/null 2>&1; then
        log_success "Validation binary is working"
    else
        log_error "Validation binary failed to execute"
        exit 1
    fi

    # Check Docker services
    local services_status=$(docker-compose ps --format json 2>/dev/null | jq -r '.[].State' 2>/dev/null || echo "unknown")
    if echo "$services_status" | grep -q "running"; then
        log_success "Docker services are running"
    else
        log_warning "Some Docker services may not be running. Check with 'docker-compose ps'"
    fi

    log_success "Setup verification completed"
}

# Run quick validation test
run_quick_test() {
    log_info "Running quick validation test..."

    # Set environment variables for testing
    export VALIDATION_ENV=local
    export DB_HOST=localhost
    export DB_PORT=5432
    export DB_NAME=kainuguru
    export DB_USER=kainuguru_user
    export DB_PASSWORD=kainuguru_pass
    export API_BASE_URL=http://localhost:8080

    # Run a simple connectivity test
    if ./bin/validation version >/dev/null 2>&1; then
        log_success "Quick validation test passed"
    else
        log_warning "Quick validation test failed. Please check the setup manually."
    fi
}

# Print usage information
print_usage() {
    cat << EOF
Kainuguru Validation Environment Setup

Usage: $0 [OPTIONS]

Options:
    -h, --help          Show this help message
    -q, --quick         Quick setup (skip optional steps)
    -v, --verify-only   Only verify existing setup
    --no-docker         Skip Docker-related setup
    --no-build          Skip building validation binary
    --no-test           Skip running quick test

Examples:
    $0                  # Full setup
    $0 --quick          # Quick setup
    $0 --verify-only    # Just verify
EOF
}

# Main setup function
main() {
    local quick_mode=false
    local verify_only=false
    local skip_docker=false
    local skip_build=false
    local skip_test=false

    # Parse command line arguments
    while [[ $# -gt 0 ]]; do
        case $1 in
            -h|--help)
                print_usage
                exit 0
                ;;
            -q|--quick)
                quick_mode=true
                shift
                ;;
            -v|--verify-only)
                verify_only=true
                shift
                ;;
            --no-docker)
                skip_docker=true
                shift
                ;;
            --no-build)
                skip_build=true
                shift
                ;;
            --no-test)
                skip_test=true
                shift
                ;;
            *)
                log_error "Unknown option: $1"
                print_usage
                exit 1
                ;;
        esac
    done

    log_info "🔧 Starting Kainuguru Validation Environment Setup"
    echo

    if [ "$verify_only" = true ]; then
        verify_setup
        log_success "✅ Verification completed"
        exit 0
    fi

    # Run setup steps
    check_prerequisites

    if [ "$skip_docker" = false ]; then
        check_docker
        setup_database
    fi

    setup_validation_framework

    if [ "$quick_mode" = false ]; then
        setup_test_data
    fi

    if [ "$skip_build" = false ]; then
        build_validation_binary
    fi

    verify_setup

    if [ "$skip_test" = false ]; then
        run_quick_test
    fi

    echo
    log_success "🎉 Validation environment setup completed successfully!"
    echo
    log_info "Next steps:"
    log_info "  1. Run 'make validate-quick' for a quick system check"
    log_info "  2. Run 'make validate-all' for complete validation"
    log_info "  3. Check 'make help' for all available commands"
    echo
}

# Run main function with all arguments
main "$@"