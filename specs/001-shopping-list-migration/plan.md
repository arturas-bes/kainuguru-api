````markdown
# Implementation Plan: Shopping List Migration Wizard

**Branch**: `001-shopping-list-migration` | **Date**: 2025-11-16 | **Spec**: [spec.md](./spec.md)
**Input**: Feature specification from `/specs/001-shopping-list-migration/spec.md`

**Note**: This template is filled in by the `/speckit.plan` command. See `.specify/templates/commands/plan.md` for the execution workflow.

## Summary

A wizard-based system that detects expired flyer products in shopping lists and presents ranked alternatives from active flyers using a two-pass search strategy (brand-aware then cross-brand). The system enforces a maximum 2-store constraint, provides confidence scores and explanations, supports both item-by-item and bulk decisions, and creates immutable offer snapshots for historical accuracy. Session state persists for 30 minutes with revalidation before confirming changes.

**Technical Approach**: Extend existing GraphQL API with wizard mutations/queries, leverage SearchService.FuzzySearchProducts for candidate discovery, use deterministic scoring (brand: 3.0, store: 2.0, size: 1.0, price: 1.0), store session state in Redis with 30-min TTL, add offer_snapshots table for immutable historical records.

## Technical Context

**Language/Version**: Go 1.24.0 (toolchain go1.24.9)  
**Primary Dependencies**: gqlgen v0.17.81 (GraphQL), Fiber v2.52.9 (HTTP), Bun v1.2.15 (ORM), go-redis v9.16.0 (cache), zerolog (logging), Viper (config)  
**Storage**: PostgreSQL 15+ (Bun ORM with partitioning), Redis 7+ (session/cache)  
**Testing**: Go testing stdlib, testify v1.11.1 assertions, BDD acceptance tests (tests/bdd/)  
**Target Platform**: Linux server (containerized, production deployment via Docker)  
**Project Type**: Single monolith backend (GraphQL API server)  
**Performance Goals**: 
- Search <500ms p95 for Lithuanian queries over active flyers (constitution requirement)
- Wizard suggestion generation <1 second for typical shopping list (5-10 items)
- Session operations <100ms p95

**Constraints**: 
- No GraphQL schema breaking changes (additive only per constitution)
- Maximum 2 stores per plan (hard limit per constitution)
- Never fabricate prices (constitution: only use actual flyer prices)
- Backward compatible with existing shopping_list_items table
- Must integrate with existing SearchService.FuzzySearchProducts (no duplicate search logic)
- Session data in Redis with 30-minute TTL
- Deterministic ranking (same input must produce same output)

**Scale/Scope**: 
- ~10k users (Lithuanian market MVP)
- ~150k products across flyers (weekly rotation)
- Shopping lists: 5-20 items average
- Active flyers: 10-15 stores weekly
- Expected wizard usage: 20-30% of active users weekly (when flyers expire)

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

### Core Principles Verification

**I. Simplicity First** ✅
- Uses existing stack: PostgreSQL + Bun, GraphQL (gqlgen), Fiber, Redis
- No new dependencies required
- Leverages existing SearchService.FuzzySearchProducts for candidate discovery
- Deterministic scoring with fixed weights (no ML complexity)

**II. Ship Over Perfect** ✅
- MVP scope: wizard for expired items only (not all product substitutions)
- Session management simplified (30-min TTL, no complex state machine)
- Fixed scoring algorithm (defer ML optimization for post-launch)
- Known limitations documented (TODOs for: bulk optimization, preferred store hints, price trend analysis)

**III. User Value** ✅
- **Direct user outcome**: Help Lithuanian shoppers replace expired items → save time
- **Cheaper trips**: Same-brand priority + price proximity scoring → save money
- **Faster trips**: Max 2 stores constraint → reduce shopping friction
- All features trace to core user needs (search, compare, list management, expired item migration)

**IV. Pragmatic Choices** ✅
- Server-side rules for scoring (no client-side heuristics)
- Two-pass search strategy (deterministic, testable)
- Store selection based on coverage algorithm (no guesswork)
- Revalidation before confirm (prevents stale data issues)

**V. Phase Quality Gates** ✅
- Phase 0: Research will document two-pass search patterns, session management in Redis, offer snapshot schema
- Phase 1: Design will produce data-model.md (entities), contracts/ (GraphQL schema), quickstart.md (API examples)
- Phase 2: Tasks will include deep analysis checklist, issue resolution plan, MCP PR workflow
- Each phase ends with validation before progression

### Technical Values Verification

**MONOLITH FIRST** ✅ — Feature stays within existing kainuguru-api monolith (no new services)

**POSTGRESQL PRIMARY** ✅ 
- Core data: shopping_list_items (existing), offer_snapshots (new table)
- Session state: Redis only (ephemeral, 30-min TTL)
- Search: leverages existing SearchService (PostgreSQL full-text search)

**GRAPHQL BY DEFAULT** ✅
- All wizard operations exposed via GraphQL mutations/queries
- DataLoader already implemented for N+1 prevention (existing pattern)
- Single endpoint architecture preserved

**STRUCTURED LOGS** ✅
- Uses existing zerolog infrastructure
- Will add wizard-specific log fields (wizard_session_id, item_count, store_count, acceptance_rate)

**FAIL GRACEFULLY** ✅
- Empty suggestion arrays when no alternatives exist (never 500)
- Typed GraphQL errors (VALIDATION_ERROR, NOT_FOUND, STALE_DATA)
- Partial failure handling (some items succeed, others return errors with explanations)

### Domain Constraints Verification

**LITHUANIAN FIRST** ✅
- Leverages existing pkg/normalize (diacritics, case folding)
- Uses existing SearchService with Lithuanian tokenization
- Canonical names and brands already normalized in ProductMaster

**NO FABRICATED PRICES** ✅
- Only suggest products with concrete flyer_product prices
- Category-only discounts (e.g., "–25%") shown as advisory, never as replacement with invented price
- Offer snapshots store actual prices, flag estimated=false for all wizard suggestions

**MAX 1–2 STORES PER PLAN** ✅
- Hard-coded constraint in store selection algorithm
- Default maxStores=1, allow override to 2 via config
- Algorithm rejects suggestions that would exceed store cap
- Validation at mutation layer (reject if user tries to bypass)

### Quality Standards Verification

**SEARCH <500ms p95** ✅ — Uses existing SearchService (already meets this target per CODEBASE_ANALYSIS.md)

**LISTS PERSIST** ✅ — Shopping lists remain intact; wizard creates new snapshots (doesn't delete expired items until user confirms)

**EXTRACTION CONTINUES** ✅ — N/A for wizard (applies to scraper/enrichment pipelines)

**AUTH SECURE** ✅ — Uses existing JWT + middleware (internal/middleware/auth.go), wizard mutations require authentication

**TESTS EXIST** ✅ — Will add BDD acceptance tests for wizard flows (expired→wizard, suggestions, confirm, skip)

### Agent Operating Rules Verification

**1. No hallucinated prices** ✅ — Only suggest products from flyer_products table with non-null price

**2. Same-brand first** ✅ — Two-pass search: pass 1 = brand+name, pass 2 = name only (cross-brand)

**3. Cap stores** ✅ — maxStores enforced in store selection algorithm and mutation validation

**4. Deterministic matching** ✅ — Fixed scoring weights (brand: 3.0, store: 2.0, size: 1.0, price: 1.0), tie-break on price

**5. Revalidate on confirm** ✅ — Re-fetch candidates before applying wizard decisions, block if any changed/expired

**6. Idempotent mutations** ✅ — Wizard mutations accept optional idempotency_key (stored in Redis with applied state)

**7. Typed errors only** ✅ — Map to GraphQL codes (VALIDATION_ERROR, FORBIDDEN, NOT_FOUND, STALE_DATA, INTERNAL_ERROR)

### Decision Framework Application

**1. User impact** ✅ — Directly helps plan cheaper/faster trips this week (expired items → current alternatives)

**2. Stack fit** ✅ — PostgreSQL for data, Redis for session, GraphQL for API, existing SearchService for discovery

**3. Honesty** ✅ — Only suggest products with actual flyer prices; advisory-only for category discounts

**4. Operability** ✅ — Logs + Prometheus counters (wizard_items_flagged_total, wizard_acceptance_rate, wizard_latency_ms)

### GATE RESULT: ✅ PASS — All constitution requirements met

**Complexity Justification**: None required (no violations detected)

## Project Structure

### Documentation (this feature)

```text
specs/001-shopping-list-migration/
├── spec.md              # Feature specification (user stories, requirements)
├── plan.md              # This file (/speckit.plan command output)
├── research.md          # Phase 0 output (/speckit.plan command)
├── data-model.md        # Phase 1 output (/speckit.plan command)
├── quickstart.md        # Phase 1 output (/speckit.plan command)
├── contracts/           # Phase 1 output (/speckit.plan command)
│   ├── schema.graphql   # GraphQL schema extensions (wizard mutations/queries)
│   └── types.graphql    # GraphQL types (WizardSession, Suggestion, etc.)
└── tasks.md             # Phase 2 output (/speckit.tasks command - NOT created by /speckit.plan)
```

### Source Code (repository root)

**Structure Decision**: Single monolith Go project (Option 1). This feature extends the existing kainuguru-api codebase without introducing new services or breaking the monolith. All wizard functionality integrates into the current layered architecture (GraphQL → Services → Repositories → Database).

```text
kainuguru-api/ (repository root)
├── cmd/
│   └── api/                          # Main GraphQL API server (existing)
│       └── main.go                   # Wizard GraphQL endpoints added via gqlgen
│
├── internal/
│   ├── models/                       # Data models (extend existing)
│   │   ├── shopping_list.go          # Add wizard-specific fields (origin, snapshot_id)
│   │   ├── offer_snapshot.go         # NEW: immutable historical offer records
│   │   └── wizard_session.go         # NEW: session state model (for Redis serialization)
│   │
│   ├── services/                     # Business logic layer
│   │   ├── shopping_list/            # Shopping list service (existing, extend)
│   │   │   ├── service.go            # Add GetExpiredItems()
│   │   │   └── migration.go          # REFACTOR: extract wizard logic here
│   │   ├── wizard/                   # NEW: wizard service
│   │   │   ├── service.go            # Core wizard orchestration
│   │   │   ├── search.go             # Two-pass search strategy
│   │   │   ├── scoring.go            # Deterministic scoring algorithm
│   │   │   ├── store_selection.go    # Store coverage optimization
│   │   │   └── session.go            # Session management (Redis operations)
│   │   └── search/                   # Search service (existing, use as-is)
│   │       └── service.go            # FuzzySearchProducts (leverage existing)
│   │
│   ├── repositories/                 # Data access layer
│   │   ├── shopping_list_repository.go   # Extend: GetExpiredItems, UpdateOrigin
│   │   └── offer_snapshot_repository.go  # NEW: CRUD for offer_snapshots
│   │
│   ├── graphql/                      # GraphQL layer
│   │   ├── schema/                   # Schema definitions
│   │   │   ├── schema.graphql        # Main schema (existing)
│   │   │   └── wizard.graphql        # NEW: wizard mutations/queries
│   │   ├── resolvers/                # Resolver implementations
│   │   │   ├── wizard.resolvers.go   # NEW: wizard mutation/query resolvers
│   │   │   └── shopping_list.resolvers.go  # Extend: add expiredItemCount field
│   │   └── generated/                # gqlgen generated code (auto-updated)
│   │
│   ├── cache/                        # Redis caching layer (existing)
│   │   └── wizard_cache.go           # NEW: wizard session cache operations
│   │
│   └── middleware/                   # HTTP middleware (existing)
│       └── auth.go                   # Auth middleware (wizard mutations use existing JWT)
│
├── migrations/                       # Database migrations
│   └── 0XX_add_wizard_tables.sql     # NEW: offer_snapshots table + shopping_list_items.origin
│
├── tests/
│   ├── bdd/                          # BDD acceptance tests
│   │   └── wizard_test.go            # NEW: end-to-end wizard scenarios
│   ├── integration/                  # Integration tests
│   │   └── wizard_service_test.go    # NEW: wizard service integration tests
│   └── unit/                         # Unit tests
│       └── scoring_test.go           # NEW: deterministic scoring algorithm tests
│
└── pkg/                              # Shared packages
    └── normalize/                    # Lithuanian text normalization (existing, use as-is)
```

**Key Integration Points**:
- **GraphQL Schema**: Additive only (new types: `WizardSession`, `Suggestion`, `OfferSnapshot`; new mutations/queries under `Mutation`/`Query`)
- **SearchService**: Reuse `FuzzySearchProducts` for candidate discovery (no duplication)
- **Shopping List Service**: Extend with wizard-specific methods (don't break existing CRUD)
- **Redis**: Store wizard sessions with 30-min TTL (key pattern: `wizard:session:{session_id}`)
- **PostgreSQL**: New `offer_snapshots` table; extend `shopping_list_items` with `origin` enum

## Complexity Tracking

> **Fill ONLY if Constitution Check has violations that must be justified**

No violations detected. Complexity tracking not required.

---

## Phase 1 Complete: Post-Design Constitution Re-Check

**Date**: 2025-11-16

### Artifacts Generated

✅ **research.md** (Phase 0) — 289 lines, all technical decisions documented
✅ **data-model.md** — Entity specifications, relationships, validation rules
✅ **contracts/wizard-api.graphql** — 397 lines, GraphQL schema extensions
✅ **quickstart.md** — API usage examples and integration guide
✅ **.github/agents/copilot-instructions.md** — Agent context updated with feature tech stack

### Constitution Re-Verification (Post-Design)

**Core Principles**:
- ✅ **I. Simplicity First** — Design leverages existing SearchService, PostgreSQL, Redis; no new dependencies
- ✅ **II. Ship Over Perfect** — MVP scope maintained (wizard for expired items only); ML scoring deferred
- ✅ **III. User Value** — All designed entities trace to user outcomes (replace expired → cheaper/faster trips)
- ✅ **IV. Pragmatic Choices** — Server-side deterministic scoring; two-pass search; revalidation before confirm
- ✅ **V. Phase Quality Gates** — Phase 0 research complete; Phase 1 design complete; ready for Phase 2 tasks

**Technical Values**:
- ✅ **MONOLITH FIRST** — Feature contained within kainuguru-api (no new services)
- ✅ **POSTGRESQL PRIMARY** — offer_snapshots table + shopping_list_items.origin; Redis only for session (ephemeral)
- ✅ **GRAPHQL BY DEFAULT** — All operations via GraphQL mutations/queries in wizard-api.graphql
- ✅ **STRUCTURED LOGS** — Design includes wizard-specific log fields (session_id, acceptance_rate)
- ✅ **FAIL GRACEFULLY** — GraphQL schema includes typed errors (VALIDATION_ERROR, STALE_DATA, NOT_FOUND)

**Domain Constraints**:
- ✅ **LITHUANIAN FIRST** — Uses existing pkg/normalize; SearchService has Lithuanian tokenization
- ✅ **NO FABRICATED PRICES** — offer_snapshots.estimated=false for all wizard suggestions (constitution compliance)
- ✅ **MAX 1–2 STORES** — Store selection algorithm enforces maxStores constraint (design in research.md)

**Quality Standards**:
- ✅ **SEARCH <500ms p95** — Reuses existing SearchService (already meets target)
- ✅ **LISTS PERSIST** — Wizard creates snapshots; expired items remain until user confirms (no premature deletion)
- ✅ **AUTH SECURE** — Wizard mutations require JWT authentication (existing middleware)
- ✅ **TESTS EXIST** — Will add BDD tests in Phase 2 (tests/bdd/wizard_test.go per project structure)

**Agent Operating Rules**:
- ✅ **1. No hallucinated prices** — offer_snapshots.estimated=false enforced by DB constraint
- ✅ **2. Same-brand first** — Two-pass search (Pass 1: brand+name, Pass 2: name only) in research.md
- ✅ **3. Cap stores** — maxStores validated in store selection algorithm (research.md section 5)
- ✅ **4. Deterministic matching** — Fixed scoring weights (brand:3.0, store:2.0, size:1.0, price:1.0) in research.md section 4
- ✅ **5. Revalidate on confirm** — Designed in wizard service confirmWizard mutation (re-fetch candidates)
- ✅ **6. Idempotent mutations** — Idempotency keys in data-model.md (Redis pattern: wizard:idempotency:{key})
- ✅ **7. Typed errors only** — GraphQL schema includes error types (wizard-api.graphql)

### Design Integrity Checks

**GraphQL Schema** (wizard-api.graphql):
- ✅ All mutations accept idempotencyKey parameter
- ✅ WizardSession includes status, expiresAt fields
- ✅ Suggestion type includes isSameBrand, priceComparison fields
- ✅ Error types defined (ValidationError, StaleDataError, NotFoundError)
- ✅ Additive only (extends existing types, no breaking changes)

**Data Model** (data-model.md):
- ✅ shopping_list_items.origin enum (backward compatible: default 'free_text')
- ✅ offer_snapshots table (immutable: no UPDATE operations)
- ✅ WizardSession (Redis-only, 30-min TTL)
- ✅ Foreign keys with appropriate CASCADE/SET NULL policies

**Project Structure** (plan.md):
- ✅ Single monolith (kainuguru-api)
- ✅ Wizard service in internal/services/wizard/ (6 files)
- ✅ Repositories extended (shopping_list_repository.go, offer_snapshot_repository.go)
- ✅ Tests planned (bdd/, integration/, unit/)
- ✅ Migration script (0XX_add_wizard_tables.sql)

### GATE RESULT: ✅ PASS — Design aligns with constitution

**Changes Since Initial Check**: None required (design validated constitution compliance)

**Recommendation**: Proceed to Phase 2 (/speckit.tasks) to generate implementation tasks.

---

## Next Steps

**Command**: `/speckit.tasks` to generate tasks.md with implementation breakdown by user story

**Expected Outcome**: Task list organized by priority (P1, P2, P3) with:
- Foundational tasks (auth, search integration, session management)
- User Story 1 tasks (P1: expired item detection & notification)
- User Story 2 tasks (P1: brand-aware suggestions)
- User Story 3 tasks (P2: store limitation)
- User Story 4 tasks (P2: item-by-item decisions)
- User Story 5 tasks (P3: bulk decisions)
- User Story 6 tasks (P3: session persistence)

**Phase 2 Deliverable**: Executable task list with file paths, dependencies, test requirements, and completion criteria.
