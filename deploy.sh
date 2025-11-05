#!/bin/bash

# Kainuguru API Deployment Script
# Usage: ./deploy.sh [environment]

set -e

ENVIRONMENT="${1:-production}"
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

echo "=================================="
echo "  Kainuguru API Deployment"
echo "=================================="
echo "Environment: $ENVIRONMENT"
echo "=================================="

# Validate environment
case $ENVIRONMENT in
  "production"|"staging"|"development")
    echo "✅ Valid environment: $ENVIRONMENT"
    ;;
  *)
    echo "❌ Invalid environment: $ENVIRONMENT"
    echo "Valid environments: production, staging, development"
    exit 1
    ;;
esac

# Check if Docker is available
if ! command -v docker &> /dev/null; then
    echo "❌ Docker is not installed"
    exit 1
fi

if ! command -v docker-compose &> /dev/null; then
    echo "❌ Docker Compose is not installed"
    exit 1
fi

# Load environment file
ENV_FILE=".env.${ENVIRONMENT}"
if [ -f "$ENV_FILE" ]; then
    echo "✅ Loading environment from: $ENV_FILE"
    set -a
    source "$ENV_FILE"
    set +a
else
    echo "⚠️  Environment file not found: $ENV_FILE"
    echo "Using default configuration"
fi

# Validate required environment variables
required_vars=("DB_PASSWORD" "REDIS_PASSWORD" "JWT_SECRET")
for var in "${required_vars[@]}"; do
    if [ -z "${!var}" ]; then
        echo "❌ Required environment variable not set: $var"
        exit 1
    fi
done

echo "✅ All required environment variables are set"

# Create necessary directories
mkdir -p logs backups ssl

# Pre-deployment checks
echo "🔍 Running pre-deployment checks..."

# Check if services are healthy
if [ "$ENVIRONMENT" = "production" ]; then
    echo "⚠️  Production deployment detected"
    echo "This will update the production environment"
    read -p "Are you sure you want to continue? (type 'YES' to confirm): " confirm
    if [ "$confirm" != "YES" ]; then
        echo "❌ Deployment cancelled"
        exit 1
    fi
fi

# Build and deploy
echo "🏗️  Building application..."
if [ "$ENVIRONMENT" = "production" ]; then
    COMPOSE_FILE="docker-compose.prod.yml"
else
    COMPOSE_FILE="docker-compose.yml"
fi

# Pull latest images
echo "📥 Pulling latest images..."
docker-compose -f "$COMPOSE_FILE" pull

# Build application
echo "🔨 Building application..."
docker-compose -f "$COMPOSE_FILE" build

# Run database migrations
echo "📊 Running database migrations..."
# docker-compose -f "$COMPOSE_FILE" run --rm kainuguru-api ./main migrate

# Start services
echo "🚀 Starting services..."
docker-compose -f "$COMPOSE_FILE" up -d

# Wait for services to be healthy
echo "⏳ Waiting for services to be healthy..."
timeout=120
elapsed=0
while [ $elapsed -lt $timeout ]; do
    if docker-compose -f "$COMPOSE_FILE" ps | grep -q "healthy"; then
        echo "✅ Services are healthy"
        break
    fi
    sleep 5
    elapsed=$((elapsed + 5))
    echo "⏳ Waiting... ($elapsed/${timeout}s)"
done

if [ $elapsed -ge $timeout ]; then
    echo "❌ Services failed to become healthy within timeout"
    echo "Checking service status:"
    docker-compose -f "$COMPOSE_FILE" ps
    exit 1
fi

# Post-deployment checks
echo "🔍 Running post-deployment checks..."

# Health check
API_URL="http://localhost:8080"
if [ "$ENVIRONMENT" = "production" ]; then
    API_URL="https://kainuguru.lt"
fi

if curl -sf "$API_URL/health" >/dev/null; then
    echo "✅ API health check passed"
else
    echo "❌ API health check failed"
    exit 1
fi

# Show service status
echo "📊 Service status:"
docker-compose -f "$COMPOSE_FILE" ps

echo ""
echo "🎉 Deployment completed successfully!"
echo "Environment: $ENVIRONMENT"
echo "API URL: $API_URL"
echo ""
echo "Useful commands:"
echo "  View logs: docker-compose -f $COMPOSE_FILE logs -f"
echo "  Stop services: docker-compose -f $COMPOSE_FILE down"
echo "  Restart services: docker-compose -f $COMPOSE_FILE restart"