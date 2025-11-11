# Refactoring Status – Go toolchain 1.25.4 baseline

## Toolchain
- Go 1.25.4 installed locally; project builds/tests clean when invoking `GOROOT=/usr/local/go go test ./...`.
- Repo itself defaults to `toolchain auto`, so CI should either pin GOROOT or use the modern toolchain cache.

## Request Context (AGENTS instructions)
- Only behavior-preserving refactors allowed.
- Request: move every service to a shared repository pattern (single `internal/repositories/`), eliminate duplicated CRUD, and keep filters in neutral packages.

## Done so far
1. **Toolchain stabilized** (run instructions above).
2. **Base repository extracted** under `internal/repositories/base` with tests ensuring GetByID/GetAll semantics and limit/pagination behavior.
3. **Store service & repository** fully migrated to shared helper (uses `base.Repository[models.Store]`). All CRUD + count now go through the base repo; service only adds business logic (timestamps, validations).
4. **Flyer service** migrated: `internal/repositories/flyer_repository.go` uses shared helper + custom filters; Flyer service delegates everything to repository.
5. **Flyer page service** migrated: repository handles pagination/filters, service only does control logic.
6. **Product service** migrated: introduced `product_repository.go` using base helper and special queries (current/valid/on-sale). Service now depends on repo interface as planned.
7. **ServiceFactory** now expects repository providers for stores/flyers/flyer pages/products. All entrypoints (API server, CLI tools) new up a `RepositoryFactory` and pass it into `NewServiceFactory`.
8. **Shopping list service migrated**: introduced `internal/shoppinglist` filters + repository contract, filled in repository implementation (categories/access/share helpers), and rewired `ShoppingListService` to depend solely on that interface. Added unit tests covering default list handling, share settings, archiving, and access validation. All command binaries now import `internal/bootstrap` so repository factories register before services spin up.
9. **Shopping list item service migrated**: created the `internal/shoppinglistitem` package (filters + repository contract), wired it through the repository factory/bootstrap, and refactored `ShoppingListItemService` to delegate all persistence to the repository. Added regression tests exercising create defaults, single item updates, bulk operations, and access validation.

## Next planned steps (pending)
1. Price history, extraction job, product master services still use direct Bun queries. Next focus:
   - Define neutral packages for their filter/contracts.
   - Port repositories to the shared helpers and add targeted characterization tests.
2. GraphQL resolvers around shopping list items still have duplicated filter conversion logic; consider moving to helpers once remaining services are migrated.
