# Development stage
FROM golang:1.24-alpine AS development

# Install build dependencies and PDF processing tools
RUN apk add --no-cache git ca-certificates tzdata poppler-utils imagemagick

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
FROM golang:1.24-alpine AS builder

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

# Build test utilities
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o test-scraper cmd/test-scraper/main.go
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o test-full-pipeline cmd/test-full-pipeline/main.go

# Production stage
FROM alpine:latest AS production

# Install runtime dependencies for PDF processing and image manipulation
RUN apk --no-cache add \
    ca-certificates \
    tzdata \
    poppler-utils \
    imagemagick \
    curl

# Create non-root user
RUN addgroup -g 1001 -S kainuguru && \
    adduser -u 1001 -S kainuguru -G kainuguru

# Set working directory
WORKDIR /app

# Create temp directories for PDF processing
RUN mkdir -p /tmp/kainuguru/pdf && \
    chown -R kainuguru:kainuguru /tmp/kainuguru

# Copy binaries from builder stage
COPY --from=builder /app/api /app/test-scraper /app/test-full-pipeline ./

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