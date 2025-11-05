# 🛒 Kainuguru API

> **Discover the best deals in Lithuanian grocery stores**

A powerful GraphQL API for browsing weekly grocery flyers, tracking price history, and managing smart shopping lists across major Lithuanian retail chains.

[![Go Version](https://img.shields.io/badge/Go-1.24+-00ADD8?style=flat&logo=go)](https://golang.org/)
[![Fiber](https://img.shields.io/badge/Fiber-v2-00ADB5?style=flat&logo=fiber)](https://gofiber.io/)
[![PostgreSQL](https://img.shields.io/badge/PostgreSQL-15+-4169E1?style=flat&logo=postgresql&logoColor=white)](https://postgresql.org/)
[![GraphQL](https://img.shields.io/badge/GraphQL-E10098?style=flat&logo=graphql&logoColor=white)](https://graphql.org/)

## ✨ Features

### 🏪 Multi-Store Support
- **Maxima, Rimi, IKI, Lidl** and other major Lithuanian grocery chains
- Unified API across different store formats and pricing systems
- Real-time flyer extraction and product cataloging

### 📊 Smart Price Analytics
- **Price History Tracking**: Comprehensive historical price data with trend analysis
- **Buying Recommendations**: ML-powered suggestions based on price patterns
- **Seasonal Analysis**: Detect price cycles and optimal buying windows
- **Multi-Store Comparison**: Find the best deals across different retailers

### 🛍️ Intelligent Shopping Lists
- **Smart Item Matching**: 3-tier fuzzy matching system for products
- **Cross-Store Shopping**: Optimize your shopping across multiple stores
- **Collaborative Lists**: Share and collaborate on shopping lists
- **Auto-Suggestions**: AI-powered product recommendations

### 🔍 Advanced Search
- **Lithuanian Full-Text Search**: Optimized for Lithuanian language with trigram support
- **Fuzzy Matching**: Find products even with typos or alternative names
- **Category Filtering**: Browse by product categories and tags
- **Price Range Search**: Filter by budget and promotional offers

### 🔐 Secure Authentication
- **JWT-based Authentication**: Secure access token and refresh token system
- **Session Management**: Device tracking and session management
- **Email Verification**: Secure account registration and password reset
- **Role-based Access**: Different permission levels for users and admins

## 🚀 Quick Start

### Prerequisites

- **Go 1.24+**
- **PostgreSQL 15+**
- **Redis 6+** (for caching and sessions)

### Installation

1. **Clone the repository**
   ```bash
   git clone https://github.com/kainuguru/kainuguru-api.git
   cd kainuguru-api
   ```

2. **Install dependencies**
   ```bash
   go mod download
   ```

3. **Set up environment**
   ```bash
   cp .env.example .env
   # Edit .env with your database and Redis credentials
   ```

4. **Run database migrations**
   ```bash
   go run cmd/migrate/main.go up
   ```

5. **Start the server**
   ```bash
   go run cmd/server/main.go
   ```

The API will be available at `http://localhost:8080`

### Docker Setup (Recommended)

1. **Start with Docker Compose**
   ```bash
   docker-compose up -d
   ```

This will start:
- API server on port 8080
- PostgreSQL database on port 5432
- Redis cache on port 6379
- GraphQL Playground at `http://localhost:8080/graphql`

## 📖 API Documentation

### GraphQL Playground

Visit `http://localhost:8080/graphql` for the interactive GraphQL playground with:
- Complete schema documentation
- Query and mutation examples
- Real-time query testing

### Core Operations

#### 🔍 Search Products
```graphql
query SearchProducts {
  advancedSearch(input: {
    query: "pienas"
    storeIDs: [1, 2]
    onSaleOnly: true
    limit: 20
  }) {
    products {
      id
      name
      currentPrice
      isOnSale
      store {
        name
      }
    }
  }
}
```

#### 📊 Price History
```graphql
query PriceHistory {
  priceHistory(productID: 123) {
    edges {
      node {
        price
        recordedAt
        isOnSale
        store {
          name
        }
      }
    }
  }

  analyzeTrend(productID: 123, period: "30_days") {
    direction
    trendPercentage
    confidence
  }
}
```

#### 🛍️ Shopping Lists
```graphql
mutation CreateShoppingList {
  createShoppingList(input: {
    name: "Weekly Shopping"
    description: "Groceries for the week"
  }) {
    id
    name
    itemCount
  }
}

mutation AddItem {
  createShoppingListItem(input: {
    listID: 1
    description: "Duona"
    quantity: 2
  }) {
    id
    description
    matchedProduct {
      name
      currentPrice
    }
  }
}
```

#### 🔐 Authentication
```graphql
mutation Register {
  register(input: {
    email: "user@example.com"
    password: "securepassword"
    fullName: "Jonas Jonaitis"
  }) {
    user {
      id
      email
    }
    accessToken
  }
}

mutation Login {
  login(input: {
    email: "user@example.com"
    password: "securepassword"
  }) {
    user {
      id
      email
    }
    accessToken
    refreshToken
  }
}
```

## 🏗️ Architecture

### Tech Stack

- **Framework**: [Fiber v2](https://gofiber.io/) - Fast HTTP framework
- **Database**: [PostgreSQL 15+](https://postgresql.org/) with [Bun ORM](https://bun.uptrace.dev/)
- **Cache**: [Redis](https://redis.io/) for sessions and query caching
- **GraphQL**: Custom GraphQL implementation with DataLoader for N+1 prevention
- **Authentication**: JWT with bcrypt password hashing
- **Testing**: BDD with [Godog](https://github.com/cucumber/godog)

### Project Structure

```
kainuguru-api/
├── cmd/                    # Application entry points
│   ├── server/            # HTTP server
│   └── migrate/           # Database migrations
├── internal/              # Internal application code
│   ├── config/           # Configuration management
│   ├── database/         # Database connection and setup
│   ├── handlers/         # HTTP handlers and GraphQL resolvers
│   ├── middleware/       # HTTP middleware (auth, CORS, etc.)
│   ├── models/           # Data models and business logic
│   ├── repositories/     # Data access layer
│   ├── services/         # Business logic services
│   │   ├── auth/        # Authentication services
│   │   ├── price/       # Price analysis services
│   │   ├── shopping/    # Shopping list services
│   │   └── search/      # Search and filtering services
│   └── graphql/         # GraphQL schema and resolvers
├── migrations/           # Database migration files
├── tests/               # Test files
│   └── bdd/            # BDD feature files
├── docs/               # Documentation
└── scripts/           # Utility scripts
```

### Database Schema

#### Core Entities
- **Stores**: Retail chain information and locations
- **Flyers**: Weekly promotional flyers from stores
- **Products**: Individual products with pricing and metadata
- **Users**: User accounts with authentication
- **Shopping Lists**: User-created shopping lists with items

#### Advanced Features
- **Price History**: Historical pricing data with trend analysis
- **Product Masters**: Unified product catalog across stores
- **User Sessions**: Secure session management
- **Search Indexes**: Optimized search with trigram support

## 🔧 Configuration

### Environment Variables

```bash
# Server Configuration
PORT=8080
ENV=development

# Database
DB_HOST=localhost
DB_PORT=5432
DB_NAME=kainuguru
DB_USER=kainuguru
DB_PASSWORD=your_password
DB_SSL_MODE=disable

# Redis Cache
REDIS_HOST=localhost
REDIS_PORT=6379
REDIS_PASSWORD=

# JWT Authentication
JWT_SECRET=your-secret-key
JWT_ACCESS_EXPIRY=15m
JWT_REFRESH_EXPIRY=24h

# Email Configuration
SMTP_HOST=smtp.gmail.com
SMTP_PORT=587
SMTP_USERNAME=your-email@gmail.com
SMTP_PASSWORD=your-app-password

# External APIs
OPENAI_API_KEY=your-openai-key
```

### Performance Tuning

```bash
# Database Connection Pool
DB_MAX_OPEN_CONNS=25
DB_MAX_IDLE_CONNS=10
DB_CONN_MAX_LIFETIME=1h

# Redis Configuration
REDIS_POOL_SIZE=10
REDIS_MIN_IDLE_CONNS=5

# Server Settings
READ_TIMEOUT=30s
WRITE_TIMEOUT=30s
BODY_LIMIT=4MB
```

## 🧪 Testing

### Running Tests

```bash
# Unit tests
go test ./...

# BDD tests
go test ./tests/bdd

# Integration tests with test database
ENV=test go test ./tests/integration

# Load testing
go run scripts/loadtest/main.go
```

### Test Coverage

```bash
# Generate coverage report
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out -o coverage.html
```

### BDD Features

The API includes comprehensive BDD test scenarios:

- **Authentication**: Registration, login, password reset
- **Search**: Product search with various filters
- **Shopping Lists**: List creation, item management, sharing
- **Price History**: Price tracking and trend analysis
- **Store Management**: Multi-store operations

## 🚀 Deployment

### Production Deployment

1. **Build the application**
   ```bash
   go build -o bin/server cmd/server/main.go
   ```

2. **Run database migrations**
   ```bash
   ./bin/server migrate
   ```

3. **Start the server**
   ```bash
   ./bin/server
   ```

### Docker Production

```bash
# Build production image
docker build -t kainuguru-api:latest .

# Run with production compose
docker-compose -f docker-compose.prod.yml up -d
```

### Health Checks

The API provides health check endpoints:

- `GET /health` - Basic health check
- `GET /health/db` - Database connectivity
- `GET /health/redis` - Redis connectivity

## 🔍 Monitoring

### Metrics

The API exposes Prometheus metrics at `/metrics`:

- Request duration and count
- Database connection pool stats
- Cache hit/miss ratios
- Authentication success/failure rates

### Logging

Structured logging with different levels:

```bash
# Development
LOG_LEVEL=debug

# Production
LOG_LEVEL=info
LOG_FORMAT=json
```

## 🤝 Contributing

1. **Fork the repository**
2. **Create a feature branch** (`git checkout -b feature/amazing-feature`)
3. **Write tests** for your changes
4. **Commit your changes** (`git commit -m 'Add amazing feature'`)
5. **Push to the branch** (`git push origin feature/amazing-feature`)
6. **Open a Pull Request**

### Development Guidelines

- Follow Go best practices and idioms
- Write comprehensive tests for new features
- Update documentation for API changes
- Use conventional commit messages
- Ensure all tests pass before submitting PR

### Code Style

```bash
# Format code
go fmt ./...

# Lint code
golangci-lint run

# Vet code
go vet ./...
```

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## 🙏 Acknowledgments

- [Fiber](https://gofiber.io/) for the excellent HTTP framework
- [Bun](https://bun.uptrace.dev/) for the powerful PostgreSQL ORM
- [Godog](https://github.com/cucumber/godog) for BDD testing support
- Lithuanian grocery stores for providing public flyer data

## 📞 Support

- **Issues**: [GitHub Issues](https://github.com/kainuguru/kainuguru-api/issues)
- **Discussions**: [GitHub Discussions](https://github.com/kainuguru/kainuguru-api/discussions)
- **Email**: support@kainuguru.lt

---

**Built with ❤️ for the Lithuanian shopping community**