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

## Next planned steps (Phase 2 remaining scope)
1. Continue peeling the remaining product master flows (e.g., master creation from products, advanced matching orchestration) off direct Bun usage by extending the new repository with targeted helpers.
2. Roll the base repository across the existing store/flyer/shopping list implementations to remove duplicated query plumbing once the remaining high-volume services are characterized.
3. Extract the GraphQL pagination helper (CODE_DUPLICATION_ANALYSIS §Pagination) and roll it through `internal/graphql/resolvers/query.go` plus related mutation resolvers, guarded by golden snapshot tests to ensure connection shapes/logging stay identical.
4. Expand middleware coverage by exercising both `Required` and `Optional` paths of `internal/middleware/auth.go` and ensure server wiring calls `NewAuthMiddleware` directly; only after the request path is fully characterized should we resume the roadmap’s error-package consolidation work.
