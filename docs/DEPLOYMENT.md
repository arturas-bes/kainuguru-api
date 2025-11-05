# Kainuguru API Deployment Guide

## Docker-based Production Deployment

All dependencies have been containerized for easy migration to remote servers.

### Prerequisites

- Docker & Docker Compose
- Environment variables file (.env)

### Included Dependencies

✅ **PDF Processing:**
- poppler-utils (pdftoppm v25.04.0)
- ImageMagick (v7.1.2-3)

✅ **Runtime:**
- Go 1.24 runtime
- Alpine Linux base
- All compiled binaries

✅ **Infrastructure:**
- PostgreSQL 15
- Redis 7
- Nginx reverse proxy

### Quick Start

1. **Development Mode:**
```bash
docker-compose up -d
```

2. **Production Mode:**
```bash
docker-compose -f docker-compose.prod.yml up -d
```

### Environment Variables (.env)

Create a `.env` file in the project root:

```bash
# Database
DB_PASSWORD=secure_password_here

# Redis
REDIS_PASSWORD=redis_password_here

# JWT
JWT_SECRET=your_jwt_secret_here

# OpenAI
OPENAI_API_KEY=your_openai_api_key_here
```

### Available Services

- **API Server** (`kainuguru-api`) - Main application server
- **Scraper Worker** (`kainuguru-scraper`) - PDF processing and scraping
- **PostgreSQL** (`postgres`) - Primary database
- **Redis** (`redis`) - Caching and job queues
- **Nginx** (`nginx`) - Reverse proxy and load balancer

### PDF Processing Features

The containerized application can:
- Download flyer PDFs from IKI and other stores
- Convert PDFs to high-quality JPEG images
- Process 35+ MB files with 59+ pages
- Handle Lithuanian text and date parsing
- Extract product information using AI

### Testing

Run the test script to verify all dependencies:
```bash
./test-docker-pdf.sh
```

### Migration to Remote Server

1. Copy project files to remote server
2. Set up environment variables
3. Run: `docker-compose -f docker-compose.prod.yml up -d`
4. All dependencies are included - no additional setup required

### Container Specifications

- **Base Image:** Alpine Linux (production) / Go 1.24 Alpine (development)
- **Size:** ~45MB runtime dependencies
- **User:** Non-root `kainuguru` user (UID: 1001)
- **Temp Storage:** `/tmp/kainuguru/pdf` with proper permissions
- **Health Checks:** Included for all services

### Security Features

- Non-root container execution
- Secure environment variable handling
- Isolated network (kainuguru-network)
- SSL/TLS ready with Nginx
- Health monitoring for all services

Ready for production deployment! 🚀