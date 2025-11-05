# Tasks: Kainuguru Grocery Flyer Aggregation System

**Input**: Design documents from `/specs/001-kainuguru-core/`
**Prerequisites**: plan.md (required), spec.md (required), research.md, data-model.md, contracts/

**Tests**: BDD tests included for critical user flows per constitution requirements

**Organization**: Tasks are grouped by user story to enable independent implementation and testing of each story.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (e.g., US1, US2, US3)
- Include exact file paths in descriptions

## Path Conventions

- **Go project**: `cmd/`, `internal/`, `pkg/` at repository root
- **Migrations**: `migrations/` at repository root
- **Config**: `configs/` at repository root
- **Tests**: `tests/bdd/` for BDD specs

---

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: Project initialization and basic structure

- [X] T001 Create project structure per implementation plan
- [X] T002 Initialize Go module with go.mod and core dependencies
- [X] T003 [P] Create docker-compose.yml with PostgreSQL 15 and Redis services
- [X] T004 [P] Create Dockerfile with multi-stage build for Go application
- [X] T005 [P] Create Makefile with build, run, test, and migration commands
- [X] T006 [P] Setup environment configuration files in configs/
- [X] T007 Initialize Goose migration structure in migrations/

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Core infrastructure that MUST be complete before ANY user story can be implemented

**⚠️ CRITICAL**: No user story work can begin until this phase is complete

- [ ] T008 Create database connection pool with pgx in internal/database/connection.go
- [ ] T009 Setup Viper configuration management in internal/config/config.go
- [ ] T010 [P] Implement structured logging with zerolog in pkg/logger/logger.go
- [ ] T011 [P] Create error handling utilities in pkg/errors/errors.go
- [ ] T012 Create base repository interface in internal/repository/base.go
- [ ] T013 Setup Bun ORM initialization in internal/database/bun.go
- [ ] T014 [P] Create Redis client wrapper in internal/cache/redis.go
- [ ] T015 Setup Fiber v2 web framework in cmd/api/main.go
- [ ] T016 Initialize gqlgen GraphQL server in internal/handlers/graphql.go
- [ ] T017 Create health check endpoint in internal/handlers/health.go
- [ ] T018 Implement rate limiting middleware in internal/middleware/ratelimit.go
- [ ] T019 Create CORS middleware configuration in internal/middleware/cors.go
- [ ] T020 Setup graceful shutdown handling in cmd/api/server.go

**Checkpoint**: Foundation ready - user story implementation can now begin in parallel

---

## Phase 3: User Story 1 - Browse Weekly Grocery Flyers (Priority: P1) 🎯 MVP

**Goal**: Enable users to browse current weekly flyers from all major grocery stores

**Independent Test**: Can access the system and view current week's flyers from at least one store with product information

### BDD Tests for User Story 1

> **NOTE: Write these tests FIRST, ensure they FAIL before implementation**

- [ ] T021 [P] [US1] BDD test for viewing current flyers list in tests/bdd/features/browse_flyers.feature
- [ ] T022 [P] [US1] BDD test for viewing flyer pages in tests/bdd/features/view_flyer_pages.feature
- [ ] T023 [P] [US1] BDD test for anonymous access to flyers in tests/bdd/features/public_access.feature

### Database Schema for User Story 1

- [ ] T024 [P] [US1] Create stores table migration in migrations/001_create_stores.sql
- [ ] T025 [P] [US1] Create flyers table migration in migrations/002_create_flyers.sql
- [ ] T026 [P] [US1] Create flyer_pages table migration in migrations/003_create_flyer_pages.sql
- [ ] T027 [P] [US1] Create products partitioned table migration in migrations/004_create_products.sql
- [ ] T028 [US1] Create weekly partition function in migrations/005_partition_function.sql
- [ ] T029 [US1] Create extraction_jobs table migration in migrations/006_create_extraction_jobs.sql
- [ ] T030 [US1] Seed initial store data in migrations/007_seed_stores.sql

### Models for User Story 1

- [ ] T031 [P] [US1] Create Store model with Bun ORM in internal/models/store.go
- [ ] T032 [P] [US1] Create Flyer model with Bun ORM in internal/models/flyer.go
- [ ] T033 [P] [US1] Create FlyerPage model with Bun ORM in internal/models/flyer_page.go
- [ ] T034 [P] [US1] Create Product model with Bun ORM in internal/models/product.go
- [ ] T035 [P] [US1] Create ExtractionJob model in internal/models/extraction_job.go

### Repositories for User Story 1

- [ ] T036 [US1] Implement StoreRepository in internal/repository/store_repository.go
- [ ] T037 [US1] Implement FlyerRepository in internal/repository/flyer_repository.go
- [ ] T038 [US1] Implement ProductRepository with partition handling in internal/repository/product_repository.go
- [ ] T039 [US1] Implement JobRepository with SKIP LOCKED in internal/repository/job_repository.go

### Scraping Infrastructure for User Story 1

- [ ] T040 [US1] Create base scraper interface in internal/services/scraper/scraper.go
- [ ] T041 [P] [US1] Implement IKI scraper in internal/services/scraper/iki_scraper.go
- [ ] T042 [P] [US1] Implement Maxima scraper in internal/services/scraper/maxima_scraper.go
- [ ] T043 [P] [US1] Implement Rimi scraper in internal/services/scraper/rimi_scraper.go
- [ ] T044 [US1] Create scraper factory in internal/services/scraper/factory.go
- [ ] T045 [US1] Implement PDF processor with pdftoppm in pkg/pdf/processor.go
- [ ] T046 [US1] Create image optimizer for API calls in pkg/image/optimizer.go

### AI Integration for User Story 1

- [ ] T047 [US1] Create OpenAI client wrapper in pkg/openai/client.go
- [ ] T048 [US1] Implement Lithuanian prompt builder in internal/services/ai/prompt_builder.go
- [ ] T049 [US1] Create product extractor service in internal/services/ai/extractor.go
- [ ] T050 [US1] Implement extraction result validator in internal/services/ai/validator.go
- [ ] T051 [US1] Add cost tracking for API calls in internal/services/ai/cost_tracker.go

### Text Processing for User Story 1

- [ ] T052 [P] [US1] Create Lithuanian text normalizer in pkg/normalize/lithuanian.go
- [ ] T053 [P] [US1] Implement unit extractor in pkg/normalize/units.go
- [ ] T054 [P] [US1] Create brand name mapper in pkg/normalize/brands.go

### Worker Implementation for User Story 1

- [ ] T055 [US1] Create job queue worker in cmd/scraper/main.go
- [ ] T056 [US1] Implement job processor with retry logic in internal/services/worker/processor.go
- [ ] T057 [US1] Add distributed locking with Redis in internal/services/worker/locker.go
- [ ] T058 [US1] Create job scheduler for weekly updates in internal/services/worker/scheduler.go

### GraphQL API for User Story 1

- [ ] T059 [US1] Define GraphQL schema for stores and flyers in graph/schema.graphqls
- [ ] T060 [US1] Generate GraphQL code with gqlgen in graph/generated/
- [ ] T061 [US1] Implement Store resolver in internal/handlers/store_resolver.go
- [ ] T062 [US1] Implement Flyer resolver in internal/handlers/flyer_resolver.go
- [ ] T063 [US1] Implement Product resolver in internal/handlers/product_resolver.go
- [ ] T064 [US1] Add DataLoader for N+1 query prevention in internal/handlers/dataloader.go

### Caching for User Story 1

- [ ] T065 [US1] Implement Redis caching for flyer data in internal/cache/flyer_cache.go
- [ ] T066 [US1] Add extraction result caching in internal/cache/extraction_cache.go

**Checkpoint**: User Story 1 complete - users can browse flyers and view products

---

## Phase 4: User Story 2 - Search for Products Across All Flyers (Priority: P1)

**Goal**: Enable product search across all current flyers with Lithuanian language support

**Independent Test**: Can search for products and get relevant results from available flyers

### BDD Tests for User Story 2

- [ ] T067 [P] [US2] BDD test for product search in tests/bdd/features/search_products.feature
- [ ] T068 [P] [US2] BDD test for Lithuanian diacritics handling in tests/bdd/features/lithuanian_search.feature
- [ ] T069 [P] [US2] BDD test for fuzzy search in tests/bdd/features/fuzzy_search.feature

### Database Schema for User Story 2

- [ ] T070 [US2] Add Lithuanian FTS configuration in migrations/008_fts_config.sql
- [ ] T071 [US2] Create search indexes on products in migrations/009_search_indexes.sql
- [ ] T072 [US2] Add trigram extension and indexes in migrations/010_trigram_indexes.sql

### Search Implementation for User Story 2

- [ ] T073 [US2] Create search service interface in internal/services/search/search.go
- [ ] T074 [US2] Implement PostgreSQL FTS search in internal/services/search/fts_search.go
- [ ] T075 [US2] Add fuzzy search with trigrams in internal/services/search/fuzzy_search.go
- [ ] T076 [US2] Create search result ranker in internal/services/search/ranker.go
- [ ] T077 [US2] Implement search facets aggregation in internal/services/search/facets.go

### GraphQL API for User Story 2

- [ ] T078 [US2] Add search schema to GraphQL in graph/schema.graphqls
- [ ] T079 [US2] Implement searchProducts resolver in internal/handlers/search_resolver.go
- [ ] T080 [US2] Add search suggestions resolver in internal/handlers/suggestions_resolver.go
- [ ] T081 [US2] Create price comparison resolver in internal/handlers/price_resolver.go

### Performance Optimization for User Story 2

- [ ] T082 [US2] Add search result caching in internal/cache/search_cache.go
- [ ] T083 [US2] Implement search query optimization in internal/services/search/optimizer.go

**Checkpoint**: User Story 2 complete - users can search products with Lithuanian support

---

## Phase 5: User Story 5 - User Registration and Authentication (Priority: P2)

**Goal**: Enable user account creation and secure authentication for personalized features

**Independent Test**: Can register an account, log out, and log back in to access saved data

### BDD Tests for User Story 5

- [ ] T084 [P] [US5] BDD test for user registration in tests/bdd/features/registration.feature
- [ ] T085 [P] [US5] BDD test for login/logout in tests/bdd/features/authentication.feature
- [ ] T086 [P] [US5] BDD test for password reset in tests/bdd/features/password_reset.feature

### Database Schema for User Story 5

- [ ] T087 [P] [US5] Create users table migration in migrations/011_create_users.sql
- [ ] T088 [P] [US5] Create user_sessions table migration in migrations/012_create_sessions.sql

### Models for User Story 5

- [ ] T089 [P] [US5] Create User model with Bun ORM in internal/models/user.go
- [ ] T090 [P] [US5] Create UserSession model in internal/models/user_session.go

### Authentication Service for User Story 5

- [ ] T091 [US5] Create auth service interface in internal/services/auth/auth.go
- [ ] T092 [US5] Implement JWT token generation in internal/services/auth/jwt.go
- [ ] T093 [US5] Add bcrypt password hashing in internal/services/auth/password.go
- [ ] T094 [US5] Create session manager in internal/services/auth/session.go
- [ ] T095 [US5] Implement email verification in internal/services/auth/email_verify.go
- [ ] T096 [US5] Add password reset functionality in internal/services/auth/password_reset.go

### User Repository for User Story 5

- [ ] T097 [US5] Implement UserRepository in internal/repository/user_repository.go
- [ ] T098 [US5] Add session repository methods in internal/repository/session_repository.go

### Authentication Middleware for User Story 5

- [ ] T099 [US5] Create JWT validation middleware in internal/middleware/auth.go
- [ ] T100 [US5] Add session validation in internal/middleware/session.go

### GraphQL API for User Story 5

- [ ] T101 [US5] Add authentication schema to GraphQL in graph/schema.graphqls
- [ ] T102 [US5] Implement register mutation in internal/handlers/auth_resolver.go
- [ ] T103 [US5] Add login/logout mutations in internal/handlers/auth_resolver.go
- [ ] T104 [US5] Create me query resolver in internal/handlers/user_resolver.go
- [ ] T105 [US5] Implement password reset mutations in internal/handlers/auth_resolver.go

**Checkpoint**: User Story 5 complete - users can register and authenticate

---

## Phase 6: User Story 3 - Create and Manage Shopping Lists (Priority: P2)

**Goal**: Enable users to create persistent shopping lists that survive weekly updates

**Independent Test**: Can create a list, add items, and verify persistence after flyer updates

### BDD Tests for User Story 3

- [ ] T106 [P] [US3] BDD test for shopping list CRUD in tests/bdd/features/shopping_lists.feature
- [ ] T107 [P] [US3] BDD test for list item management in tests/bdd/features/list_items.feature
- [ ] T108 [P] [US3] BDD test for list persistence in tests/bdd/features/list_persistence.feature

### Database Schema for User Story 3

- [ ] T109 [US3] Create shopping_lists table migration in migrations/013_create_shopping_lists.sql
- [ ] T110 [US3] Create shopping_list_items table migration in migrations/014_create_list_items.sql
- [ ] T111 [US3] Create product_masters table migration in migrations/015_create_product_masters.sql
- [ ] T112 [US3] Create product_tags table migration in migrations/016_create_tags.sql

### Models for User Story 3

- [ ] T113 [P] [US3] Create ShoppingList model in internal/models/shopping_list.go
- [ ] T114 [P] [US3] Create ShoppingListItem model in internal/models/shopping_list_item.go
- [ ] T115 [P] [US3] Create ProductMaster model in internal/models/product_master.go
- [ ] T116 [P] [US3] Create ProductTag model in internal/models/product_tag.go

### Shopping List Service for User Story 3

- [ ] T117 [US3] Create shopping list service in internal/services/shopping/shopping_list.go
- [ ] T118 [US3] Implement 3-tier item matching in internal/services/shopping/item_matcher.go
- [ ] T119 [US3] Add item availability tracker in internal/services/shopping/availability.go
- [ ] T120 [US3] Create alternative suggester in internal/services/shopping/suggester.go
- [ ] T121 [US3] Implement list sharing functionality in internal/services/shopping/sharing.go

### Product Matching for User Story 3

- [ ] T122 [US3] Create product master service in internal/services/product/master.go
- [ ] T123 [US3] Implement tag-based matching in internal/services/product/tag_matcher.go
- [ ] T124 [US3] Add confidence scoring in internal/services/product/confidence.go

### Repositories for User Story 3

- [ ] T125 [US3] Implement ShoppingListRepository in internal/repository/shopping_list_repository.go
- [ ] T126 [US3] Create ProductMasterRepository in internal/repository/product_master_repository.go

### GraphQL API for User Story 3

- [ ] T127 [US3] Add shopping list schema to GraphQL in graph/schema.graphqls
- [ ] T128 [US3] Implement shopping list mutations in internal/handlers/shopping_list_resolver.go
- [ ] T129 [US3] Add list item mutations in internal/handlers/list_item_resolver.go
- [ ] T130 [US3] Create list sharing resolvers in internal/handlers/sharing_resolver.go

**Checkpoint**: User Story 3 complete - users can manage persistent shopping lists

---

## Phase 7: User Story 4 - Track Price History for Products (Priority: P3)

**Goal**: Enable users to view historical price trends for products

**Independent Test**: Can view a product and see its price history over time

### BDD Tests for User Story 4

- [ ] T131 [P] [US4] BDD test for price history viewing in tests/bdd/features/price_history.feature
- [ ] T132 [P] [US4] BDD test for price trend analysis in tests/bdd/features/price_trends.feature

### Price History Service for User Story 4

- [ ] T133 [US4] Create price history service in internal/services/price/history.go
- [ ] T134 [US4] Implement price aggregation in internal/services/price/aggregator.go
- [ ] T135 [US4] Add trend calculator in internal/services/price/trends.go

### Data Archival for User Story 4

- [ ] T136 [US4] Create archival service in internal/services/archive/archiver.go
- [ ] T137 [US4] Implement image removal for archives in internal/services/archive/cleaner.go

### GraphQL API for User Story 4

- [ ] T138 [US4] Add price history schema to GraphQL in graph/schema.graphqls
- [ ] T139 [US4] Implement price history resolver in internal/handlers/price_history_resolver.go

**Checkpoint**: User Story 4 complete - users can track price history

---

## Phase 8: Polish & Cross-Cutting Concerns

**Purpose**: Improvements that affect multiple user stories

- [ ] T140 [P] Create README.md with setup and deployment instructions
- [ ] T141 [P] Add API documentation generation in docs/api.md
- [ ] T142 Run full BDD test suite and fix any failures
- [ ] T143 Performance testing with 100 concurrent users
- [ ] T144 [P] Add monitoring dashboards configuration in configs/monitoring/
- [ ] T145 Security audit for authentication and SQL injection
- [ ] T146 Add database backup scripts in scripts/backup/
- [ ] T147 Create production deployment configuration in docker-compose.prod.yml
- [ ] T148 Run quickstart.md validation for developer onboarding
- [ ] T149 Add TODO comments for post-MVP optimizations throughout codebase
- [ ] T150 Final testing of all user stories in integration

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: No dependencies - can start immediately
- **Foundational (Phase 2)**: Depends on Setup completion - BLOCKS all user stories
- **User Stories (Phase 3-7)**: All depend on Foundational phase completion
  - US1 & US2 (P1): Can run in parallel after Foundation
  - US5 (P2): Can start after Foundation, needed for US3
  - US3 (P2): Depends on US5 for authentication
  - US4 (P3): Can start after US1 (needs product data)
- **Polish (Phase 8)**: Depends on all desired user stories being complete

### User Story Dependencies

- **User Story 1 (P1)**: Foundation only - establishes core data and scraping
- **User Story 2 (P1)**: Needs US1 for product data to search
- **User Story 5 (P2)**: Foundation only - independent auth system
- **User Story 3 (P2)**: Needs US5 for user authentication
- **User Story 4 (P3)**: Needs US1 for historical product data

### Within Each User Story

- BDD tests MUST be written and FAIL before implementation
- Database schema before models
- Models before services
- Services before API endpoints
- Core implementation before optimization

### Parallel Opportunities

- All Setup tasks marked [P] can run in parallel
- US1 and US2 can be developed in parallel after Foundation
- Within each story, all [P] tasks can run in parallel
- Different user stories can be worked on by different team members

---

## Parallel Example: User Story 1

```bash
# Launch all BDD tests together:
Task T021: "BDD test for viewing current flyers"
Task T022: "BDD test for viewing flyer pages"
Task T023: "BDD test for anonymous access"

# Launch all models together:
Task T031: "Create Store model"
Task T032: "Create Flyer model"
Task T033: "Create FlyerPage model"
Task T034: "Create Product model"
Task T035: "Create ExtractionJob model"

# Launch all scrapers together:
Task T041: "Implement IKI scraper"
Task T042: "Implement Maxima scraper"
Task T043: "Implement Rimi scraper"
```

---

## Implementation Strategy

### MVP First (User Stories 1 & 2 Only)

1. Complete Phase 1: Setup
2. Complete Phase 2: Foundational (CRITICAL)
3. Complete Phase 3: User Story 1 (Browse Flyers)
4. Complete Phase 4: User Story 2 (Search)
5. **STOP and VALIDATE**: Test core functionality
6. Deploy/demo if ready

### Full Implementation

1. Complete MVP (US1 & US2)
2. Add Authentication (US5)
3. Add Shopping Lists (US3)
4. Add Price History (US4)
5. Polish and deploy

### Parallel Team Strategy

With 3 developers:
1. All work on Setup + Foundation together (2 days)
2. Split after Foundation:
   - Dev A: User Story 1 (Flyers & Scraping)
   - Dev B: User Story 2 (Search) + User Story 5 (Auth)
   - Dev C: User Story 3 (Lists) + User Story 4 (History)
3. Reconvene for Polish phase

---

## Notes

- [P] tasks = different files, no dependencies within phase
- [USx] label maps task to specific user story for traceability
- Each user story should be independently completable and testable
- BDD tests must fail before implementing features
- Commit after each task or logical group
- Total tasks: 150 (organized for 6-week timeline)