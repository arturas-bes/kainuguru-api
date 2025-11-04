# Implementation Plan: Kainuguru Grocery Flyer Aggregation System

**Branch**: `001-kainuguru-core` | **Date**: 2025-11-04 | **Spec**: [spec.md](./spec.md)
**Input**: Feature specification from `/specs/001-kainuguru-core/spec.md`

**Note**: This template is filled in by the `/speckit.plan` command. See `.specify/templates/commands/plan.md` for the execution workflow.

## Summary

Build a Lithuanian grocery flyer aggregation system that collects weekly promotional materials from all major grocery stores, extracts product information using ChatGPT Vision API, and provides a GraphQL API for browsing flyers, searching products with Lithuanian language support, and managing persistent shopping lists. The system will use Go as a monolith with PostgreSQL for everything (database, search, queue), implementing a pragmatic MVP approach with structured logging and TODO comments for future optimizations.

## Technical Context

**Language/Version**: Go 1.22+
**Primary Dependencies**: Fiber v2 (web framework), gqlgen (GraphQL), Bun ORM (database), ChatGPT/GPT-4 Vision API
**Storage**: PostgreSQL 15+ (database, search, queue, partitioned tables), Redis (sessions, distributed locks)
**Testing**: Ginkgo/Godog for BDD tests, standard Go testing for unit tests
**Target Platform**: Linux server (Docker containers), DigitalOcean deployment
**Project Type**: web - GraphQL API monolith
**Performance Goals**: <500ms search response, 100 concurrent users, 4-hour flyer processing
**Constraints**: €80/month infrastructure, no complex monitoring, Lithuanian language support required
**Scale/Scope**: 10,000 registered users, all major Lithuanian grocery stores, 6-week delivery

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

✅ **Simplicity First**: Using PostgreSQL for everything (database, search, queue). No Elasticsearch, no NATS, no fancy tools.
✅ **Ship Over Perfect**: Direct ChatGPT usage without cost controls initially, TODOs throughout for optimizations.
✅ **User Value**: Every feature directly helps find deals and manage shopping lists.
✅ **Pragmatic Choices**: ChatGPT for extraction instead of complex OCR logic, PostgreSQL FTS instead of Elasticsearch.
✅ **Monolith First**: Single Go application, no microservices.
✅ **PostgreSQL Everything**: Database, search (FTS + trigrams), queue (job_queue table).
✅ **GraphQL From Start**: Using gqlgen from the beginning, no REST phase.
✅ **Basic Logging**: Zerolog for structured logs, no Grafana/Prometheus.
✅ **Fail Gracefully**: Failed extractions display page as-is, continue processing other pages.

## Project Structure

### Documentation (this feature)

```text
specs/001-kainuguru-core/
├── plan.md              # This file (/speckit.plan command output)
├── research.md          # Phase 0 output (/speckit.plan command)
├── data-model.md        # Phase 1 output (/speckit.plan command)
├── quickstart.md        # Phase 1 output (/speckit.plan command)
├── contracts/           # Phase 1 output (/speckit.plan command)
└── tasks.md             # Phase 2 output (/speckit.tasks command - NOT created by /speckit.plan)
```

### Source Code (repository root)

```text
kainuguru-api/
├── cmd/
│   ├── api/             # GraphQL server entry point
│   ├── scraper/         # Scraper worker entry point
│   └── migrator/        # Database migration runner
├── internal/
│   ├── config/          # Viper configuration management
│   ├── models/          # Bun ORM entities
│   ├── repository/      # Data access layer
│   ├── services/
│   │   ├── auth/        # JWT authentication service
│   │   ├── product/     # Product CRUD and search
│   │   ├── scraper/     # Store-specific scrapers
│   │   └── ai/          # ChatGPT integration
│   └── handlers/        # GraphQL resolvers with DataLoaders
├── pkg/
│   ├── openai/          # ChatGPT client wrapper
│   └── normalize/       # Lithuanian text normalization
├── migrations/          # Goose SQL migration files
├── configs/             # YAML configuration files
├── tests/
│   └── bdd/            # Ginkgo BDD test specs
├── docker-compose.yml   # Development environment
├── Dockerfile          # Multi-stage build
├── Makefile           # Build and run commands
├── go.mod             # Go dependencies
└── README.md          # Setup and deployment guide
```

**Structure Decision**: Web application monolith structure chosen due to GraphQL API requirement and single-service architecture. All components (API, scraper, migrator) share the same codebase but have separate entry points in cmd/ directory. This aligns with the "Monolith First" principle while maintaining clear separation of concerns.

## Complexity Tracking

> **Fill ONLY if Constitution Check has violations that must be justified**

| Violation | Why Needed | Simpler Alternative Rejected Because |
|-----------|------------|-------------------------------------|
| Redis addition | Distributed locking for concurrent scrapers, session management | PostgreSQL advisory locks don't work well across multiple processes; session storage in PostgreSQL would add unnecessary database load |

Note: Redis is the only deviation from "PostgreSQL Everything" but is justified for distributed locking to prevent race conditions in concurrent scrapers and efficient session management. This is still simpler than adding a full message queue system.