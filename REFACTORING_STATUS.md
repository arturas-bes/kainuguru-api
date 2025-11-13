# Refactoring Status – Go toolchain 1.25.4 baseline

## Toolchain
- Go 1.25.4 installed locally; project builds/tests clean when invoking `GOROOT=/usr/local/go go test ./...`.
- Repo itself defaults to `toolchain auto`, so CI should either pin GOROOT or use the modern toolchain cache.

## Request Context (AGENTS instructions)
- Only behavior-preserving refactors allowed.
- Request: move every service to a shared repository pattern (single `internal/repositories/`), eliminate duplicated CRUD, and keep filters in neutral packages.

## Done so far
1. **Toolchain stabilized** (run instructions above).
2. **GraphQL handler propagates request context**: `internal/handlers/graphql.go` now derives from `c.Context()` and cancellation is asserted in `internal/handlers/graphql_test.go`, satisfying AGENTS §7.1.
3. **Repository tree consolidated + DI hook**: only `internal/repositories/*` remains and `_ "github.com/kainuguru/kainuguru-api/internal/bootstrap"` wires store/flyer/shopping list/shopping list item/extraction job factories into `internal/services/repository_registry.go` before any service is constructed.
4. **Store service + repository split**: the new `internal/store` contract fronts `internal/repositories/store_repository.go`, and `internal/services/store_service.go` + `_test.go` prove the service is a thin delegator (filters/counting/timestamps stay intact).
5. **Flyer service migrated**: `internal/flyer` exposes filters + repository interface, `internal/repositories/flyer_repository.go` centralizes Bun queries, and `internal/services/flyer_service_test.go` locks delegation/error paths.
6. **Shopping list + shopping list item services migrated**: the domain packages (`internal/shoppinglist`, `internal/shoppinglistitem`) own filters/contracts, repositories live under `internal/repositories/`, and the services/tests enforce defaults, share-code rules, and bulk item semantics.
7. **Extraction job stack refactored**: `internal/extractionjob` defines filters/interfaces, `internal/repositories/extraction_job_repository.go` implements queue-safe Bun operations with sqlite-backed tests, and `internal/services/extraction_job_service.go` gains DI + logger injection with table-driven tests.
8. **Generic base repository added**: `internal/repositories/base` now exposes typed CRUD helpers + sqlite-backed characterization tests so future services can stop re-implementing Bun boilerplate.
9. **Flyer page domain refactor**: introduced `internal/flyerpage` (filters + repository contract), wired a Bun-backed repository built on the new base helper, updated the service to delegate via DI, and added delegation/timestamp tests to lock behavior.
10. **Price history service migrated**: new `internal/pricehistory` package defines filters + repository contract, `internal/repositories/price_history_repository.go` centralizes the Bun queries, and `internal/services/price_history_service.go` now delegates via DI with unit tests covering delegation/error propagation.
11. **Product service migrated**: introduced `internal/product` (filters + repository contract), implemented `internal/repositories/product_repository.go` on top of the base helper, refactored `internal/services/product_service.go` to depend on the repository/DI seam, and added delegation/normalization tests.
12. **Product master repository introduced**: `internal/productmaster` now defines the contract, `internal/repositories/product_master_repository.go` handles CRUD/filtering, the transactional MatchProduct flow, master verification/deactivation/duplicate handling, and the statistics queries. `internal/services/product_master_service.go` + `_test.go` now delegate those responsibilities.
13. **Product master worker delegates to the repository**: added sqlite-backed characterization tests for unmatched products/review flags/master stats, extended `productmaster.Repository` with targeted helpers, and refactored `internal/workers/product_master_worker.go` to stop touching `*bun.DB` directly so background matching now shares the same data-access layer.
14. **Store/flyer/shopping list repositories share the base helper**: added sqlite-backed characterization tests for each repository, switched the remaining Bun CRUD paths to `internal/repositories/base`, and extracted filter/pagination helpers so service logic keeps identical ordering/error semantics.
15. **GraphQL pagination helper extracted**: introduced typed pagination args + unit tests in `internal/graphql/resolvers/helpers.go`, updated every resolver that previously duplicated the `limit/offset` dance (stores/flyers/products/shopping lists/items/price history) to use the helper, and kept the connection builders untouched to avoid schema changes.
16. **Auth middleware characterized**: added Fiber-based tests in `internal/middleware/auth_test.go` that lock required vs optional flows and assert context propagation so future changes to `NewAuthMiddleware` can't silently regress authentication behavior.
17. **GraphQL endpoint runs through shared auth middleware**: the Fiber `/graphql` route now wraps the optional `NewAuthMiddleware` configured with real JWT/session services, so request contexts already contain user/session IDs before resolver execution.
18. **User + session repositories deduplicated**: both `internal/repositories/user_repository.go` and `session_repository.go` now sit on top of `base.Repository`, complete with sqlite-backed characterization tests to lock filtering/pagination + maintenance behaviors before the refactor.
19. **GraphQL pagination snapshots captured**: `internal/graphql/resolvers/pagination_snapshot_test.go` + `testdata/*.json` record golden connection payloads for stores/flyers/products/shopping lists/items/price history.
20. **GraphQL handler relies solely on middleware context**: the legacy token parsing fallback in `internal/handlers/graphql.go` was removed so `/graphql` now trusts the shared middleware for auth state.
21. **Extraction job + shopping list item repositories now use the base helper**: both repos delegate CRUD to `internal/repositories/base` with sqlite-backed tests covering filter/order semantics, leaving only domain-specific locking logic outside the helper.
22. **Shared error-handling package complete**: `pkg/errors` now has comprehensive unit tests (14 test cases covering all error types, wrapping, type checking, and status code mapping), usage examples demonstrating migration patterns, and complete documentation (README.md) with service-layer patterns, migration strategy, and best practices. The package was already implemented with typed errors (validation/authentication/authorization/notfound/conflict/ratelimit/external/internal), HTTP status mapping, error wrapping with `%w` semantics, and GraphQL compatibility—ready for service migration.

## Next planned steps (Phase 2 remaining scope)
1. Wire the new GraphQL pagination snapshots into CI/regression gates (or at least document the `-update_graphql_snapshots` workflow) so connection changes require explicit approval.
2. Push toward the Week 3 goals (70% coverage + large-file splits) by prioritizing service unit tests and resolver extraction work.
3. Begin migrating critical services to use `pkg/errors` (starting with product_master, shopping_list_item, auth services per migration strategy).
