AGENTS.md — Zero‑Risk Refactoring Playbook (Kainuguru API)

Purpose: This file tells any coding agent exactly how to refactor this Go codebase without changing behavior, without hallucinating requirements, and without losing functionality. If the agent can’t meet these constraints, it must stop and report why.

⸻

0) Don’t Be Clever — Be Safe

Hard rules (never violate):
1.	No behavior changes. Same inputs ⇒ same outputs, same DB effects, same logs (level & message text), same GraphQL schema & resolver behavior, same HTTP status codes.
2.	No schema changes. No DB migrations, no field or type renames in protobuf/GraphQL, no JSON shape changes.
3.	No feature work. Refactor only. If you’re “improving” UX or adding validations, stop.
4.	No silent deletions. Delete only dead code proven by grep + references = 0 and a passing test net.
5.	No global context.Background() in request paths. Always propagate the request context.
6.	Tests first. If coverage is missing, write characterization tests that lock current behavior before edits.
7.	go test ./... must pass before and after each change set. No flaky tests; fix or skip with justification.

If any rule conflicts with a proposed change, abort the change and produce a short note in the PR description.

⸻

1) Repository Facts (Use These, Don’t Invent Your Own)
   •	Language: Go.
   •	Primary interface: GraphQL (gqlgen) under cmd/api and internal/graphql/* (resolvers, schema).
   •	DB/ORM: PostgreSQL 15+ via Bun ORM (internal/database). No GORM.
   •	Business logic: internal/services/* (lots of duplicated CRUD).
   •	Data access: Consolidate onto internal/repositories/*; legacy internal/repository/* duplicates exist.
   •	Middleware: internal/middleware/* (auth, request context, etc.).
   •	Helpers: pkg/normalize (Lithuanian text normalization) — don’t break signatures.
   •	Known hotspots: duplicated CRUD across services; oversized resolvers (e.g., query resolver ~800+ LOC); context misuse in handlers/graphql.go.

You must verify file paths before editing using git grep.

⸻

2) Allowed vs. Forbidden Changes

Allowed (safe) refactors:
•	Pure renames (files, funcs, types) with all references updated by the compiler.
•	Extract functions/modules; split large files; remove dead code with zero references.
•	Replace duplicated CRUD with shared helpers or generic repository without changing method contracts.
•	Improve error wrapping (%w), logging structure, and context propagation.
•	Add missing interfaces where they already exist conceptually (DI seams), no runtime behavior change.
•	Add tests (unit/integration/snapshot) and build scripts.

Forbidden:
•	Any change that alters GraphQL schema, resolver signatures, HTTP routes, env var names, configuration defaults, or log messages (text/levels) used by dashboards.
•	Changing DB queries’ semantics, isolation, or pagination defaults.
•	Adding retries/timeouts where there were none (unless guarded behind feature flags and off by default; but prefer not to touch).
•	Introducing new dependencies that change performance characteristics.

⸻

3) Safety Checklists

3.1 Global Pre‑flight
•	go vet ./... and staticcheck ./... are clean or explicitly waived.
•	gofmt/goimports produce no diffs.
•	git grep -n "context.Background()" internal shows no hits in handlers/resolvers (OK in app startup).
•	git grep -n "internal/repository/" → migrate to internal/repositories/ or remove duplicate.

3.2 Handlers & GraphQL
•	Context: never derive from context.Background() in request paths. Use the incoming request context.
•	Resolvers: don’t change argument names, default values, nullability, or error messages; preserve ordering.
•	Pagination/Filtering: keep defaults, include/exclude rules, and sorting semantics identical.

3.3 Services & Repositories
•	Replace copy‑pasted CRUD with a shared helper only if the exported method names and signatures stay the same.
•	Preserve transaction boundaries (begin/commit/rollback points) and isolation assumptions.
•	Keep SQL where‑clauses and preload/join sets identical.

3.4 Logging & Errors
•	Keep log levels and messages stable; add fields is OK, but do not drop existing ones.
•	Wrap errors with %w; never swallow them.

⸻

4) Minimal Workflow for Each Change (One‑Concern Commits)
    1.	Create a branch: git checkout -b refactor/<target>
    2.	Characterize current behavior with tests:
          •	Unit tests on the target package (cover public surface & edge conditions).
          •	For GraphQL, add snapshot tests of queries/mutations that hit the code.
    3.	Refactor one concern (rename/extract/deduplicate/propagate ctx).
    4.	Run: go test ./... and ensure no regression in logs/outputs.
    5.	Self‑review with the Diff Rubric (next section).
    6.	Commit message: refactor(<area>): <action>; no behavior change plus metrics (LOC ±, funcs moved, coverage Δ).

⸻

5) Diff Rubric (Reject if Any Fails)
   •	API invariants: Public signatures unchanged. GraphQL schema unchanged.
   •	Context: No new context.Background() calls on request paths.
   •	Queries: SQL text and parameters equivalent. Pagination & ordering identical.
   •	Errors/logs: Same messages/levels; added context is fine.
   •	Tests: Added/updated characterization tests prove the same outputs.
   •	Imports: No side‑effect import changes.
   •	Performance: Same or fewer allocations; comparable benchmarks if available.

⸻

6) How to Kill CRUD Duplication (Without Risk)

Goal: centralize repeated service CRUD patterns while preserving public APIs.

	1.	Identify repeated methods via grep:

git grep -n "GetByID(ctx context.Context" internal/services
git grep -n "Create(ctx context.Context" internal/services


	2.	Introduce a small, internal helper in internal/repositories/share (or similar) that encapsulates the common Bun patterns (ID lookups, lists with filters, create/update/delete) but does not change existing exported service names.
	3.	Refactor one service at a time by delegating internals to the helper while keeping the same public signatures and error texts.
	4.	Tests: For each service migrated, assert identical error strings, not just types.

⸻

7) Known Hotfixes to Apply Safely

7.1 Context Propagation in GraphQL Handler (Do This First)
•	Current bug: handler uses context.Background() instead of request ctx.
•	Fix (pseudocode; adjust to actual file):

// BEFORE
baseCtx := c.Context()
ctx := context.Background()
ctx = context.WithValue(ctx, middleware.UserContextKey, claims.UserID)

// AFTER
ctx := c.Context() // preserve request context
ctx = context.WithValue(ctx, middleware.UserContextKey, claims.UserID)


	•	Add a unit test that cancels the request context and asserts downstream operations observe cancellation.

7.2 Repository Folder Consolidation
•	Pick one folder: internal/repositories/* is the source of truth.
•	Migrate anything under internal/repository/* or delete duplicates if identical.
•	Update imports; add a short package doc comment describing the layer contract.

⸻

8) Testing Strategy That Prevents Hallucinations
   •	Characterization tests: codify existing behavior before edits — particularly GraphQL response shapes and error texts.
   •	Golden snapshots for GraphQL: store JSON bodies on disk and compare byte‑for‑byte.
   •	DB tests: use a dedicated test database or transactional tests; seed with minimal fixtures.
   •	Concurrency & cancellation: include at least one test that cancels a request and ensures the repo layer stops.
   •	No mocks unless necessary: prefer thin fakes over mocks that can paper over behavior changes.

Useful commands

# run all tests
go test ./...

# race detector
go test -race ./...

# static checks
go vet ./...
staticcheck ./...


⸻

9) PR Template (Use This)

Title: refactor(<area>): <action>; no behavior change

Summary
- WHAT changed:
- WHY (duplication/size/ctx/etc.):
- HOW verified (tests, snapshots):

Safety
- [ ] No GraphQL schema changes
- [ ] No DB schema or query behavior changes
- [ ] No log message/level changes used in dashboards
- [ ] No new context.Background() in request paths

Metrics
- LOC: -XXX / +YYY
- Test coverage Δ: +X%
- Binary size / allocations (if measured): ~same


⸻

10) Agent Operating Mode (Anti‑Hallucination)
    •	Ground yourself: before editing, run git ls-files + git grep to prove the targets exist. If a file/symbol isn’t found, do not invent it.
    •	Echo the diff plan: list the exact files/functions to touch before writing code.
    •	Confine scope: if more than ~200 added lines or >5 files change, split into multiple PRs.
    •	Stop on ambiguity: if a change requires guessing behavior, halt and write a TODO in the PR description instead of changing behavior.

⸻

11) Folder‑Specific Guidance
    •	cmd/api: no behavior changes; keep flags/env vars identical.
    •	internal/graphql/*: keep schema/resolvers unchanged; break up large resolvers via internal helpers only.
    •	internal/services/*: reduce duplication; preserve public APIs and error strings.
    •	internal/repositories/*: single source of truth; add small shared helpers; don’t over‑abstract.
    •	internal/middleware/*: don’t change keys or claim extraction behavior.
    •	pkg/normalize: maintain function names and exact outputs (important for data matching).

⸻

12) Definition of Done for a Refactor PR
    •	All tests pass with -race.
    •	No behavior/regression diffs in GraphQL snapshots.
    •	No new context.Background() in request paths; request cancellation respected.
    •	Duplicated code reduced where targeted; public surface area unchanged.
    •	PR description filled per template with metrics.

⸻

Appendix A — Quick Commands

# Check for duplicated CRUD signatures quickly
git grep -n "func (s .*\) GetByID(ctx context.Context" internal/services

# Find accidental Background usage
git grep -n "context.Background()" internal

# Confirm only one repositories folder remains
ls -la internal/repo*

NON NEGOTIABLE RULES FOR SAFE REFACTORING
Always move by the roadmap REFACTORING_ROADMAP.md and mark tasks as done
Always use analysis available in CODEBASE_ANALYSIS.md and CODE_DUPLICATION_ANALYSIS.md
Always use guidelines REFACTORING_GUIDELINES.md
Always create Pull request after each phase finished using MCP
Test each change manually and automatically
