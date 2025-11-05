#!/bin/bash

# Load testing script for Kainuguru API
# Usage: ./run.sh [config_file]

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
CONFIG_FILE="${1:-${SCRIPT_DIR}/load_test.conf}"

# Load configuration if file exists
if [ -f "$CONFIG_FILE" ]; then
    echo "Loading configuration from: $CONFIG_FILE"
    source "$CONFIG_FILE"
fi

# Set default values if not configured
API_URL="${API_URL:-http://localhost:8080}"
CONCURRENT_USERS="${CONCURRENT_USERS:-100}"
REQUESTS_PER_USER="${REQUESTS_PER_USER:-10}"
TEST_DURATION_SECONDS="${TEST_DURATION_SECONDS:-30}"

echo "=================================="
echo "  Kainuguru API Load Test"
echo "=================================="
echo "API URL: $API_URL"
echo "Concurrent Users: $CONCURRENT_USERS"
echo "Requests per User: $REQUESTS_PER_USER"
echo "Test Duration: ${TEST_DURATION_SECONDS}s"
echo "=================================="

# Check if API is reachable
echo "Checking API health..."
if curl -sf "$API_URL/health" >/dev/null 2>&1; then
    echo "✅ API is reachable"
else
    echo "❌ API is not reachable at $API_URL"
    echo "Please ensure the server is running before running load tests"
    exit 1
fi

# Export environment variables for the load test
export API_URL
export CONCURRENT_USERS
export REQUESTS_PER_USER
export TEST_DURATION_SECONDS

# Build and run the load test
echo "Building load test..."
cd "$SCRIPT_DIR"
go build -o loadtest main.go

echo "Starting load test..."
./loadtest

# Cleanup
rm -f loadtest

echo "Load test completed!"