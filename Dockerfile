# Development stage
FROM golang:1.22-alpine AS development

# Install build dependencies
RUN apk add --no-cache git ca-certificates tzdata

# Set working directory
WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Default command for development
CMD ["go", "run", "cmd/api/main.go"]

# Build stage
FROM golang:1.22-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git ca-certificates tzdata

# Set working directory
WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build API server
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o api cmd/api/main.go

# Build scraper worker
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o scraper cmd/scraper/main.go

# Build migrator
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o migrator cmd/migrator/main.go

# Production stage
FROM alpine:latest AS production

# Install runtime dependencies
RUN apk --no-cache add ca-certificates tzdata poppler-utils

# Create non-root user
RUN addgroup -g 1001 -S kainuguru && \
    adduser -u 1001 -S kainuguru -G kainuguru

# Set working directory
WORKDIR /app

# Copy binaries from builder stage
COPY --from=builder /app/api /app/scraper /app/migrator ./

# Copy configuration files
COPY --from=builder /app/configs ./configs

# Change ownership to non-root user
RUN chown -R kainuguru:kainuguru /app

# Switch to non-root user
USER kainuguru

# Expose port
EXPOSE 8080

# Default command (can be overridden)
CMD ["./api"]