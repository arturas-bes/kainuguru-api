# Tasks: Shopping List Migration Wizard

**Input**: Design documents from `/specs/001-shopping-list-migration/`
**Prerequisites**: plan.md (required), spec.md (required), research.md, data-model.md, contracts/wizard-api.graphql

**Tests**: NOT included (optional - not explicitly requested in specification per constitution note)

**Organization**: Tasks grouped by user story (US1-US6) to enable independent implementation and testing.

## Format: `- [ ] [ID] [P?] [Story?] Description`

- **Checkbox**: `- [ ]` (markdown checkbox, REQUIRED)
- **[ID]**: Sequential number (T001, T002...) in execution order
- **[P]**: Can run in parallel (different files, no blocking dependencies)
- **[Story]**: User story label (US1, US2, etc.) - REQUIRED for story phases only
- **Description**: Clear action with exact file path

**Path Conventions** (Go monolith):
- Models: `internal/models/`
- Services: `internal/services/wizard/`
- Repositories: `internal/repositories/`
- GraphQL: `internal/graphql/schema/`, `internal/graphql/resolvers/`
- Migrations: `migrations/`
- Cache: `internal/cache/`
- Tests: `tests/bdd/`, `tests/integration/`, `tests/unit/`

---

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: Database schema and GraphQL schema initialization

- [X] T001 Create migration file migrations/0XX_add_wizard_tables.sql for offer_snapshots table per data-model.md schema
- [X] T002 Add shopping_list_items.origin enum ('flyer', 'free_text') to migration with DEFAULT 'free_text' for backward compatibility
- [X] T003 [P] Backfill shopping_list_items.origin='flyer' WHERE flyer_product_id IS NOT NULL in migration
- [X] T004 [P] Copy contracts/wizard-api.graphql to internal/graphql/schema/wizard.graphqls for gqlgen
- [X] T005 Run migration and verify tables/indexes created: offer_snapshots, shopping_list_items.origin, all indexes per data-model.md

**Checkpoint**: Database schema ready, GraphQL schema staged for gqlgen

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Core models, repositories, and infrastructure (MUST complete before ANY user story)

**⚠️ CRITICAL**: No user story implementation can begin until this phase is complete

### Database Models

- [X] T006 [P] Create internal/models/offer_snapshot.go with OfferSnapshot struct matching data-model.md (ID, ShoppingListItemID, FlyerProductID, ProductMasterID, StoreID, ProductName, Brand, Price, Unit, SizeValue, SizeUnit, Estimated, ValidFrom, ValidTo, SnapshotReason, CreatedAt)
- [X] T007 [P] Extend internal/models/shopping_list.go with ShoppingListItem.Origin field (enum: 'flyer', 'free_text')
- [X] T008 [P] Create internal/models/wizard_session.go for Redis serialization (ID, UserID, ShoppingListID, ExpiredItems, Suggestions, Decisions, SelectedStores, Status, CreatedAt, ExpiresAt, LastUpdatedAt) per data-model.md

### Repositories

- [X] T009 [P] Create internal/repositories/offer_snapshot_repository.go with Create(), GetByShoppingListItemID(), methods (Bun ORM, immutable: no UPDATE)
- [X] T010 [P] Extend internal/repositories/shopping_list_repository.go with GetExpiredItems(ctx, listID) (join flyer_products.valid_to < NOW())
- [X] T011 Create internal/cache/wizard_cache.go with SaveSession(), GetSession(), DeleteSession() using Redis pattern wizard:session:{id}, 1800s TTL

### Base Services

- [X] T012 Create internal/services/wizard/service.go with struct WizardService{db, redis, searchService, scoringWeights ScoringWeights}
- [X] T013 [P] Create internal/services/wizard/scoring.go with ScoreSuggestion() pure function per research.md section 4 (weights: brand 3.0, store 2.0, size 1.0, price 1.0)
- [X] T014 [P] Create internal/services/wizard/store_selection.go with SelectOptimalStores() per research.md section 5 (greedy algorithm, maxStores constraint)

### GraphQL Integration

- [ ] T015 Run gqlgen generate to create resolvers from internal/graphql/schema/wizard.graphqls
- [ ] T016 Create internal/graphql/resolvers/wizard.resolvers.go stub with all mutation/query signatures (startWizard, decideItem, applyBulkDecisions, confirmWizard, cancelWizard, wizardSession)

**Checkpoint**: Foundation complete - models, repositories, base services ready; user story work can begin in parallel

---

## Phase 3: User Story 1 - Expired Item Detection & Notification (Priority: P1) 🎯 MVP

**Goal**: Detect when flyer products expire and notify users so they can start migration wizard

**Independent Test**: Create shopping list with items from expired flyers, verify system detects expired state and surfaces notification count

**Acceptance Criteria**:
- Shopping list shows count of expired items (e.g., "3 items need updating")
- Badge/indicator visible on login when expired items exist
- Expired items visually distinct from active items in list view

### Implementation Tasks

- [X] T017 [P] [US1] Create internal/services/wizard/expired_detection.go with GetExpiredItemsForList(ctx, listID) using shopping_list_repository.GetExpiredItems()
- [X] T018 [P] [US1] Extend internal/graphql/resolvers/shopping_list.resolvers.go to add ShoppingList.expiredItemCount field resolver (calls GetExpiredItemsForList, returns count)
- [X] T019 [P] [US1] Extend internal/graphql/resolvers/shopping_list.resolvers.go to add ShoppingList.hasActiveWizardSession field resolver (checks Redis for active session)
- [X] T020 [US1] Create worker internal/workers/expire_flyer_items.go that runs daily at midnight to mark items as expired when flyers pass valid_to (uses cron, updates shopping_list_items)
- [X] T021 [US1] Add Prometheus counter wizard_items_flagged_total in internal/monitoring/metrics.go, increment in GetExpiredItemsForList()
- [X] T022 [US1] Add zerolog fields (list_id, expired_count) to expired detection service calls

**Checkpoint**: Users can see expired item counts and are notified to start wizard

---

## Phase 4: User Story 2 - Brand-Aware Product Suggestions (Priority: P1) 🎯 MVP

**Goal**: Generate ranked suggestions using two-pass search (brand+name, then name-only) with deterministic scoring

**Independent Test**: Provide expired "Coca-Cola 12-pack", verify system suggests Coca-Cola alternatives first (pass 1), then other colas (pass 2), ranked by score

**Acceptance Criteria**:
- Same-brand alternatives appear first regardless of store
- Suggestions ranked by score (brand:3.0, store:2.0, size:1.0, price:1.0)
- Confidence scores (0.0-1.0) and explanations provided for each suggestion

### Implementation Tasks

- [X] T023 [P] [US2] Create internal/services/wizard/search.go with TwoPassSearch(ctx, expiredItem) that calls SearchService.FuzzySearchProducts twice (pass 1: brand+name, pass 2: name-only), merges/deduplicates results
- [X] T024 [US2] Implement ScoreSuggestion() in scoring.go per research.md (pure function with brand/store/size/price weights, tie-break on price)
- [X] T025 [US2] Create internal/services/wizard/explanation.go with GenerateExplanation(suggestion, score) returning human-readable text (e.g., "Same brand, similar size, €0.50 cheaper")
- [X] T026 [US2] Implement RankSuggestions(candidates, weights) in scoring.go (sort by TotalScore DESC, PriceCompare ASC, ProductID ASC for determinism)
- [X] T027 [US2] Add unit tests in tests/unit/scoring_test.go with table-driven tests for ScoreSuggestion() determinism (same inputs → same outputs)
- [X] T028 [US2] Implement SelectOptimalStores() in store_selection.go (greedy algorithm, maxStores=2 constraint, minAdditionalItems/minSavingsEUR thresholds per research.md)
- [X] T029 [US2] Add Prometheus histogram wizard_latency_ms and counter wizard_suggestions_returned_total in metrics.go

**Checkpoint**: Two-pass search generates deterministic, ranked suggestions with explanations

---

## Phase 5: User Story 3 - Store Limitation Management (Priority: P2)

**Goal**: Enforce maximum 2-store constraint in store selection algorithm

**Independent Test**: Generate suggestions from 5 different stores, verify SelectOptimalStores() returns ≤2 stores maximizing coverage

**Acceptance Criteria**:
- Store selection never exceeds 2 stores (hard limit)
- Algorithm prioritizes top store by coverage, adds 2nd only if beneficial (≥2 items or ≥€5 savings)
- Validation prevents users from bypassing constraint

### Implementation Tasks

- [ ] T030 [P] [US3] Add maxStores validation in store_selection.go SelectOptimalStores() (reject if result would exceed config.MaxStores)
- [ ] T031 [P] [US3] Add coverage calculation in store_selection.go (count items per store, pick top 2 by coverage)
- [ ] T032 [US3] Add savings calculation in store_selection.go (compare prices, compute €savings for 2nd store justification)
- [ ] T033 [US3] Add unit tests in tests/unit/store_selection_test.go verifying maxStores constraint never violated
- [ ] T034 [US3] Add Prometheus histogram wizard_selected_store_count in metrics.go (track store count distribution)

**Checkpoint**: Store selection respects 2-store limit and optimizes coverage

---

## Phase 6: User Story 4 - Item-by-Item Decision Making (Priority: P2)

**Goal**: Allow users to review and decide on each expired item (REPLACE, KEEP, REMOVE)

**Independent Test**: Start wizard with 5 expired items, verify user can make different decisions per item, session state updates correctly

**Acceptance Criteria**:
- Each item presented individually with ranked suggestions
- User can REPLACE (select suggestion), KEEP (manual search later), or REMOVE from list
- Explanations visible for each suggestion (why suggested)

### Implementation Tasks

- [X] T035 [P] [US4] Implement startWizard mutation resolver in wizard.resolvers.go (create WizardSession, call GetExpiredItemsForList, TwoPassSearch for each, SelectOptimalStores, save to Redis, return session)
- [ ] T036 [P] [US4] Implement decideItem mutation resolver in wizard.resolvers.go (load session from Redis, validate decision, update Decisions map, save session, return updated session)
- [ ] T037 [US4] Add idempotency key handling in decideItem (check Redis wizard:idempotency:{key}, store result with 24h TTL per data-model.md)
- [ ] T038 [US4] Implement wizardSession query resolver in wizard.resolvers.go (load from Redis, map to GraphQL type)
- [ ] T039 [US4] Add session expiration check in all resolvers (if ExpiresAt < NOW(), return EXPIRED status, delete from Redis)
- [ ] T040 [US4] Add Prometheus counters wizard_acceptance_rate (track REPLACE/KEEP/REMOVE counts) in metrics.go

**Checkpoint**: Users can make granular decisions per item with session state persistence

---

## Phase 7: User Story 5 - Bulk Decision Making (Priority: P3)

**Goal**: Allow users to accept all top suggestions at once for faster processing

**Independent Test**: Create wizard session with 10 expired items, apply "Replace All", verify all items updated with top suggestions and ≤2 stores

**Acceptance Criteria**:
- applyBulkDecisions mutation accepts all top-ranked suggestions
- If bulk would exceed 2 stores, automatically limit to top 2 stores by item count
- Return updated session with all decisions applied

### Implementation Tasks

- [ ] T041 [P] [US5] Implement applyBulkDecisions mutation resolver in wizard.resolvers.go (iterate Suggestions, select top for each item, validate maxStores, update Decisions, save session)
- [ ] T042 [US5] Add bulk store validation in applyBulkDecisions (if >2 stores, re-run SelectOptimalStores to pick best 2, update decisions accordingly)
- [ ] T043 [US5] Add idempotency key handling in applyBulkDecisions (same pattern as decideItem)
- [ ] T044 [US5] Add unit tests in tests/unit/wizard_service_test.go for bulk decision logic (verify store cap enforcement)

**Checkpoint**: Bulk operations work with automatic store limitation

---

## Phase 8: User Story 6 - Session Persistence & Recovery (Priority: P3)

**Goal**: Save wizard progress in Redis (30-min TTL) and allow resumption if interrupted

**Independent Test**: Start wizard, close app, reopen within 30 minutes, verify session restored; wait >30 min, verify session expired

**Acceptance Criteria**:
- Session persists in Redis for 30 minutes
- User can resume where they left off if within TTL
- If flyer data changed, detect staleness and prompt restart
- Sessions >30 minutes are deleted (EXPIRED status)

### Implementation Tasks

- [ ] T045 [P] [US6] Add session TTL extension in wizard_cache.go ExtendSessionTTL(sessionID) (reset Redis key to 1800s on any update)
- [ ] T046 [P] [US6] Add staleness detection in wizardSession query resolver and confirmWizard (store datasetVersion in session, compare flyer_products.updated_at max on session load and confirm, return STALE_DATA error if changed since session start)
- [ ] T047 [US6] Implement revalidation in confirmWizard before applying (re-fetch all selected flyerProductIDs, verify valid_to still future, prices unchanged)
- [ ] T048 [US6] Add session cleanup worker internal/workers/cleanup_expired_sessions.go (runs hourly, deletes Redis keys wizard:session:* where expires_at < NOW())

**Checkpoint**: Sessions persist and resume reliably with staleness protection

---

## Phase 9: Final - Confirm Wizard & Apply Changes

**Goal**: Implement confirmWizard mutation that persists all decisions to PostgreSQL

**Purpose**: This is the culminating action that applies all wizard decisions permanently

### Implementation Tasks

- [ ] T049 Create internal/services/wizard/confirm.go with ConfirmWizard(ctx, sessionID) method
- [ ] T050 Implement confirmWizard mutation resolver in wizard.resolvers.go (load session, validate Status=IN_PROGRESS, call ConfirmWizard service)
- [ ] T051 In confirm.go ConfirmWizard(), start Bun transaction for atomicity
- [ ] T052 [P] In transaction: for each REPLACE decision, create OfferSnapshot with snapshot_reason='wizard_migration', estimated=false per data-model.md
- [ ] T053 [P] In transaction: for each REPLACE decision, update shopping_list_item.flyer_product_id to new product, set origin='flyer'
- [ ] T054 [P] In transaction: for each REMOVE decision, DELETE shopping_list_item
- [ ] T055 In transaction: for KEEP decisions, no changes (item remains expired, user handles manually)
- [ ] T056 After transaction commit, update session Status=COMPLETED, delete from Redis
- [ ] T057 Add idempotency key handling in confirmWizard (check wizard:idempotency:{key}, store session_id with 24h TTL)
- [ ] T058 Add revalidation logic: re-fetch all selected flyer_product_ids, verify valid_to >= NOW(), prices match session suggestions (return STALE_DATA error if changed)
- [ ] T059 Add rollback on revalidation failure (keep session IN_PROGRESS, allow user to review stale items)

**Checkpoint**: Wizard decisions permanently applied to shopping list with full ACID guarantees

---

## Phase 10: Error Handling & Observability

**Purpose**: GraphQL error mapping, logging, and metrics

- [ ] T060 [P] Create internal/graphql/errors/wizard_errors.go with typed errors (ValidationError, StaleDataError, NotFoundError, ExpiredSessionError) mapping to GraphQL codes
- [ ] T061 [P] Add error wrapping in all wizard service methods using pkg/errors (or fmt.Errorf with %w)
- [ ] T062 [P] Map service errors to GraphQL errors in all resolvers (e.g., ErrSessionExpired → ExpiredSessionError)
- [ ] T063 Add structured logging to wizard service methods: session_id, list_id, user_id, item_count, store_count, decision_action fields
- [ ] T064 [P] Add all Prometheus metrics to internal/monitoring/metrics.go: wizard_items_flagged_total, wizard_suggestions_returned_total, wizard_acceptance_rate (histogram), wizard_selected_store_count (histogram), wizard_latency_ms (histogram)
- [ ] T065 Add metrics instrumentation to service methods (defer recordDuration, increment counters at decision points)

**Checkpoint**: Errors are typed and observable, metrics captured for all wizard operations

---

## Phase 11: Cancel Wizard

**Purpose**: Allow users to abandon wizard session without applying changes

- [ ] T066 Implement cancelWizard mutation resolver in wizard.resolvers.go (load session, update Status=CANCELLED, delete from Redis, return true)
- [ ] T067 Add idempotency key handling in cancelWizard (same pattern)

**Checkpoint**: Users can exit wizard without changes

---

## Phase 12: Integration & Polish

**Purpose**: Cross-cutting concerns and final integration

- [ ] T068 [P] Add DataLoader for Store, Product, ProductMaster to prevent N+1 queries in wizardSession resolver (use existing dataloader pattern from codebase)
- [ ] T069 [P] Add rate limiting to startWizard mutation (max 5 sessions per user per hour) to prevent abuse
- [ ] T070 Update internal/graphql/schema/shopping_list.graphqls to extend ShoppingList type with expiredItemCount and hasActiveWizardSession fields
- [ ] T071 [P] Create quickstart example in specs/001-shopping-list-migration/quickstart.md showing cURL/GraphQL calls for full wizard flow
- [ ] T072 Add comment documentation to all public wizard service methods (godoc format)
- [ ] T073 Run gqlgen generate final time to regenerate resolvers with all changes
- [ ] T074 Run go fmt ./... and go vet ./... on all wizard code
- [ ] T075 Verify all wizard files follow internal/ package structure conventions

**Checkpoint**: Wizard fully integrated, documented, and polished

---

## Phase 13: List Locking & Tests (Constitution Compliance)

**Purpose**: Implement FR-016 list locking and add BDD tests per constitution Phase Quality Gates

- [ ] T076 [P] Add shopping_lists.is_locked BOOLEAN DEFAULT false to migration file (for FR-016 compliance)
- [ ] T077 [P] Implement locking logic in startWizard mutation: SET shopping_lists.is_locked=true, reject if already locked with "migration in progress" error
- [ ] T078 Implement unlock logic in confirmWizard and cancelWizard mutations: SET shopping_lists.is_locked=false after transaction complete
- [ ] T079 Add ShoppingList.isLocked field resolver in shopping_list.resolvers.go (returns is_locked value for "migration in progress" indicator)
- [ ] T080 Create tests/bdd/wizard_expired_detection_test.go with scenarios: user with expired items sees notification, badge count matches expired count
- [ ] T081 Create tests/bdd/wizard_suggestions_test.go with scenarios: same-brand appears first, suggestions ranked by score, confidence in 0.0-1.0 range
- [ ] T082 Create tests/bdd/wizard_decisions_test.go with scenarios: REPLACE/SKIP/REMOVE actions persist, session state updates correctly
- [ ] T083 Create tests/bdd/wizard_confirm_test.go with scenarios: confirm applies all decisions atomically, revalidation blocks on stale data, unlock occurs
- [ ] T084 Create tests/bdd/wizard_session_test.go with scenarios: session expires after 30 min, staleness detected on resume, cancel unlocks list

**Checkpoint**: List locking enforced (FR-016) and BDD tests exist (constitution requirement)

---

## Dependencies & Execution Order

### User Story Completion Order (for MVP)

**Phase 1**: US1 (Expired Detection) - foundational, enables wizard trigger
**Phase 2**: US2 (Suggestions) - core wizard value proposition, depends on US1
**Phase 3**: US3 (Store Limit) - enhances US2, independent otherwise
**Phase 4**: US4 (Decisions) - requires US1+US2 complete
**Phase 5**: US5 (Bulk) - requires US4 complete (extends decision logic)
**Phase 6**: US6 (Persistence) - enhances US4, can be developed in parallel

### Parallel Execution Opportunities

**Can run in parallel** (marked with [P] in tasks):
- T002 (migration fields) + T003 (backfill) + T004 (GraphQL schema copy)
- T006 (OfferSnapshot model) + T007 (ShoppingListItem extension) + T008 (WizardSession model)
- T009 (repository) + T010 (repository) + T011 (cache)
- T013 (scoring) + T014 (store selection) can be developed independently
- Phase 3 (US1) + Phase 4 (US2) can start after Phase 2 complete
- T023 (search) + T025 (explanation) + T027 (tests) can run in parallel
- All error handling (T060-T062) + metrics (T064-T065) can run in parallel

### Critical Path (Sequential Dependencies)

1. Phase 1: Setup (T001-T005) → Phase 2: Foundational (T006-T016)
2. Phase 2 complete → Phase 3 (US1) can start
3. Phase 2 complete → Phase 4 (US2) can start
4. Phase 3 + Phase 4 complete → Phase 6 (US4 - decisions) can start
5. Phase 6 complete → Phase 9 (confirm) can start
6. Phase 9 complete → Integration tests

**MVP Scope**: US1 (detection) + US2 (suggestions) + US4 (decisions) + Phase 9 (confirm)
- Delivers core wizard functionality
- Deferrable: US5 (bulk - nice to have), US6 (persistence - UX enhancement), US3 (already enforced in US2)

---

## Implementation Strategy

**Week 1**: Phases 1-2 (Setup + Foundation) - 16 tasks
**Week 2**: Phases 3-4 (US1 + US2) - 19 tasks  
**Week 3**: Phases 5-9 (US3-US6 + Confirm) - 27 tasks
**Week 4**: Phases 10-12 (Polish + Integration) - 13 tasks

**Total Tasks**: 75
**Parallelizable**: 32 tasks (43%)
**Critical Path**: ~43 tasks sequential

**MVP Delivery** (Phases 1-4 + Phase 9 only): ~40 tasks, 2-2.5 weeks

---

## Validation Checklist

After completing all tasks, verify:

- [ ] All migrations applied successfully: `go run cmd/migrator/main.go status`
- [ ] All tests pass: `go test ./internal/services/wizard/... ./internal/graphql/resolvers/...`
- [ ] GraphQL schema valid: `gqlgen generate` produces no errors
- [ ] All resolvers implemented: no "unimplemented" panics in wizard.resolvers.go
- [ ] Metrics exposed: `curl localhost:9090/metrics | grep wizard_`
- [ ] Redis keys have TTL: `redis-cli TTL wizard:session:test-id` returns 1800
- [ ] Constitution compliance:
  - maxStores never exceeded (check store_selection.go logic)
  - Idempotency keys on all mutations (check resolvers)
  - No fabricated prices (offer_snapshots.estimated=false check)
  - Same-brand first (verify TwoPassSearch order)
  - Revalidation before confirm (check confirmWizard logic)
  - Typed errors only (check error mapping)
  - List locking enforced (check startWizard/confirmWizard/cancelWizard)
  - BDD tests exist for critical flows (check tests/bdd/wizard_*_test.go)
- [ ] No `internal/repository/` paths (should be `internal/repositories/`)
- [ ] All wizard code uses context propagation (no context.Background() in request paths)

---

## Success Metrics (Track Post-Launch)

- wizard_items_flagged_total: Track volume of expired items
- wizard_acceptance_rate: Target ≥80% suggestions accepted
- wizard_latency_ms (p95): Target <1000ms for typical list
- wizard_selected_store_count: Verify 1-2 stores dominant distribution
- User completion rate: Target ≥70% of started wizards reach confirmWizard

---

**Total Tasks**: 84 (75 original + 9 constitution compliance)
**Estimated Effort**: 3.5-4.5 weeks (1 developer)
**MVP Subset**: 45 tasks (US1+US2+US4+Confirm+Locking+Core Tests = 2.5-3 weeks)
