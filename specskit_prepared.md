```angular2html
  /speckit.specify Build a Lithuanian grocery flyer aggregation system "Kainuguru" as a Go monolith with these features:

  CORE FUNCTIONALITY:
  - Scrape weekly flyers from Lithuanian stores (IKI, Maxima, Rimi) with PDF-to-image conversion
  - Extract products using ChatGPT/GPT-4 Vision API handling inconsistent Lithuanian text
  - Implement product master catalog solving weekly product identity changes
  - PostgreSQL full-text search with Lithuanian language support and trigram fuzzy matching
  - Smart shopping lists using 3-tier identification (product→tag→text) surviving flyer rotations
  - GraphQL API (gqlgen) for all operations with JWT authentication

  TECHNICAL ARCHITECTURE:
  - Go 1.22+ with Fiber v2 framework in modular monolith structure
  - PostgreSQL 15+ for everything (database, search, queue, partitioned product tables)
  - Redis for sessions and distributed locking preventing race conditions
  - Bun ORM with migrations, bulk operations, connection pooling (25 max)
  - Docker containerization, structured logging with zerolog (no fancy monitoring)

  DATA CHALLENGES SOLVED:
  - Product names varying weekly handled via normalized_name and product_masters table
  - Lithuanian diacritics (ą,č,ę,ė,į,š,ų,ū,ž) normalized for search
  - Shopping list items auto-migrate when products expire using confidence scoring
  - Concurrent scraper conflicts prevented with Redis distributed locks
  - Failed extractions queued in PostgreSQL job_queue for manual review

  SIMPLIFIED MVP APPROACH:
  - Direct ChatGPT usage without complex cost controls (TODOs for later)
  - PostgreSQL-based queue instead of NATS
  - Basic error handling returning empty arrays on failure
  - Rate limiting (100 req/min) without complex metrics
  - No Elasticsearch, Grafana, Prometheus - just logs

  DELIVERABLES:
  - GraphQL schema with shops, flyers, products, shopping lists queries/mutations
  - Store-specific scrapers with duplicate detection
  - User system with OAuth-ready schema
  - BDD tests using Godog/Cucumber
  - 6-week implementation timeline

```

```angular2html
/plan Implement Kainuguru MVP in 6 weeks with this structure and timeline:

  PROJECT STRUCTURE:
  kainuguru-go/
  ├── cmd/api (GraphQL server), cmd/scraper (worker), cmd/migrator (database)
  ├── internal/config (viper config), internal/models (Bun ORM entities), internal/repository (data access)
  ├── internal/services/auth (JWT), /product (CRUD+search), /scraper (store-specific), /ai (ChatGPT)
  ├── internal/handlers (GraphQL resolvers with DataLoaders)
  ├── pkg/openai (ChatGPT client), pkg/normalize (Lithuanian text)
  ├── migrations/ (Goose SQL files), configs/ (YAML), tests/bdd (Ginkgo tests)

  WEEK 1-2 FOUNDATION:
  - Setup: Docker compose with PostgreSQL 15 + Redis, project structure, environment configs
  - Database: Users (OAuth-ready), shops (with location), flyers, products (partitioned), product_masters, tags, shopping_lists schema
  - Models: Bun ORM entities with nullable fields, JSONB configs, array types
  - Services: UserService, AuthService (JWT+bcrypt), ProductService with bulk operations
  - GraphQL: Schema definition, gqlgen setup, authentication middleware, basic resolvers

  WEEK 3 EXTRACTION:
  - Scrapers: IKIScraper, MaximaScraper, RimiScraper with PDF download
  - Processing: pdftoppm for PDF→images, distributed locking, duplicate detection
  - AI: ChatGPT integration with structured prompts for Lithuanian products
  - Validation: Price sanity checks, normalize Lithuanian text (ą→a), brand extraction
  - Queue: PostgreSQL job_queue with SKIP LOCKED for concurrent workers

  WEEK 4 FEATURES:
  - Search: PostgreSQL FTS with Lithuanian config, trigram indexes, multi-strategy (exact→FTS→fuzzy→tag)
  - Shopping Lists: CRUD operations, 3-tier matching, auto-migration for expired items
  - Matching: Confidence scoring, suggestion algorithm, bulk operations
  - GraphQL: Complete resolvers, DataLoaders, cursor pagination, rate limiting

  WEEK 5 PRODUCTION:
  - Safeguards: Connection pooling, rate limiting middleware, health checks
  - Optimization: Bulk upserts (1000 chunks), materialized views for popular products
  - Testing: BDD specs for critical flows, integration tests for GraphQL
  - Logging: Structured zerolog, error handling, TODO comments for monitoring

  WEEK 6 DEPLOYMENT:
  - Docker: Multi-stage build, docker-compose for services
  - Config: Environment variables for production
  - Deploy: DigitalOcean droplet (4GB RAM), managed PostgreSQL
  - Documentation: API docs, deployment guide, TODOs for post-MVP

  DELIVERABLES:
  - Working GraphQL API at /graphql
  - Automated scraping with ChatGPT extraction
  - PostgreSQL search returning results <500ms
  - Smart shopping lists persisting across weeks
  - Basic auth with JWT tokens
  - Structured logs, no complex monitoring
  - €80/month base infrastructure cost
  - make file prepared to run project, to seed project, to install project, project is ran via docker compose


```

```angular2html
 /constitution Kainuguru MVP Development Constitution:

  MISSION:
  Build a simple, working Lithuanian grocery flyer aggregation system that helps users find deals and manage shopping lists, shipping in 6 weeks without over-engineering.

  CORE PRINCIPLES:
  1. SIMPLICITY FIRST - Use boring, proven technology. PostgreSQL for everything. No fancy tools.
  2. SHIP OVER PERFECT - Working MVP beats perfect architecture. TODOs for optimization.
  3. USER VALUE - Every feature must directly help Lithuanian shoppers save money.
  4. PRAGMATIC CHOICES - ChatGPT without cost controls initially. Monitor, then optimize.

  TECHNICAL VALUES:
  - MONOLITH FIRST - Start simple, split later if needed
  - POSTGRESQL EVERYTHING - Database, search, queue in one place
  - GRAPHQL FROM START - Better developer experience, no REST phase
  - BASIC LOGGING - Structured logs only, no Grafana/Prometheus
  - FAIL GRACEFULLY - Return empty arrays, log errors, continue processing

  CONSTRAINTS:
  - NO COMPLEX METRICS - Ship first, measure later
  - NO OVER-ENGINEERING - Resist adding "nice to have" features
  - LITHUANIAN FOCUS - All features must handle Lithuanian language properly

  QUALITY STANDARDS:
  - SEARCH WORKS - Results in <500ms with Lithuanian text
  - LISTS PERSIST - Shopping lists survive weekly data rotation
  - EXTRACTION CONTINUES - Failed pages don't stop processing
  - AUTH SECURE - Proper JWT + bcrypt, no shortcuts
  - TESTS EXIST - BDD for critical paths, not 100% coverage

  DECISION FRAMEWORK:
  When in doubt, ask:
  1. Does this help users find grocery deals? If no, skip it.
  2. Can PostgreSQL do it? If yes, don't add new infrastructure.
  4. Is it simpler to use ChatGPT? If yes, don't build complex logic.
  5. Can we monitor it with logs? If yes, don't add metrics tools.

  SUCCESS METRICS (Simple):
  ✓ Flyers scraped successfully
  ✓ Products extracted (any success rate)
  ✓ Search returns results
  ✓ Shopping lists work
  ✓ Users can register and login
  ✓ Site doesn't crash

  POST-MVP TODOS:
  - Add cost monitoring for ChatGPT
  - Implement OCR fallback strategies
  - Add proper monitoring tools
  - Optimize based on real usage
  - Scale when we have users

  TEAM COMMITMENTS:
  - Reference KAINUGURU_MVP_FINAL.md for all decisions
  - Use TODO comments liberally for future work
  - Test critical paths, not edge cases

```