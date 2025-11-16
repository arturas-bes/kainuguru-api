<!--
Sync Impact Report — Constitution v1.2.0
═══════════════════════════════════════════════════════════════════════════════

VERSION CHANGE: 0.0.0 (template) → 1.2.0 (initial ratification)

MODIFIED PRINCIPLES:
  • All principles initialized from template

ADDED SECTIONS:
  • I. Simplicity First — Prefer boring, proven tech in our stack
  • II. Ship Over Perfect — Working > perfect; ship smallest useful slice
  • III. User Value — Must help Lithuanian shoppers plan cheaper trips
  • IV. Pragmatic Choices — Leverage existing services; server-side rules
  • V. Phase Quality Gates — Mandatory deep-analysis, issue resolution, MCP PR
  • Technical Values — Stack constraints (PostgreSQL, GraphQL, logs)
  • Domain Constraints — Lithuanian-first, no fabricated prices, max 1-2 stores
  • Quality Standards — Performance & reliability targets
  • Agent Operating Rules — Non-negotiable wizard guardrails
  • Shopping List Data Lineage — Required fields & traceability
  • Search & Matching Rules — Normalization & ranking weights
  • Minimal Metrics — Prometheus counters for observability

REMOVED SECTIONS:
  • None (first fill)

TEMPLATES REQUIRING UPDATES:
  ✅ .specify/templates/plan-template.md — Constitution Check section verified
  ✅ .specify/templates/spec-template.md — User value & requirements alignment verified
  ✅ .specify/templates/tasks-template.md — Phase gates & quality standards verified
  ✅ .specify/templates/commands/*.md — Agent operating rules propagated

FOLLOW-UP TODOs:
  • None — all placeholders filled with concrete rules from user input
  • AGENTS.md already contains refactoring-specific safety rules (complementary)

═══════════════════════════════════════════════════════════════════════════════
-->

# Kainuguru API Constitution

**Purpose**: Define hard rules for our delivery agent(s) and contributors so we ship a useful MVP fast, without lying to users, and without wrecking the codebase.

## Core Principles

### I. Simplicity First

Prefer boring, proven tech already in our stack. PostgreSQL + Bun, GraphQL (gqlgen), Fiber, Redis for cache, Prometheus for minimal metrics. Any "new toy" requires explicit proof it beats the simplest in-stack option.

**Rationale**: Reduces onboarding friction, leverages team expertise, minimizes operational surface area. New dependencies must justify themselves against existing capabilities.

### II. Ship Over Perfect

Working > perfect. Ship the smallest useful slice, leave TODOs where optimization or refactor is obvious. Revisit after real usage data.

**Rationale**: Real user feedback trumps theoretical perfection. Document known limitations; optimize based on observed bottlenecks, not speculation.

### III. User Value (NON-NEGOTIABLE)

If it doesn't help Lithuanian shoppers plan cheaper, faster trips (search, compare, list, wizard), it doesn't ship. Features must directly support: flyer search, price comparison, shopping list management, or expired item migration.

**Rationale**: Tight scope prevents scope creep. Every feature must trace to a user outcome: save money, save time, or reduce shopping friction.

### IV. Pragmatic Choices

Leverage existing services and patterns from our codebase. Prefer server-side rules over "smart" guesswork. Only add complexity when it is impossible to meet a user outcome otherwise.

**Rationale**: Server-side logic is testable, debuggable, and auditable. Client-side heuristics drift and produce inconsistent results. Complexity requires proof of necessity.

### V. Phase Quality Gates (MANDATORY)

Each phase MUST complete three gates before progression:

1. **Deep Analysis** — Verify principle adherence (simplicity, user value), Lithuanian handling, performance vs standards, TODOs recorded for deferred work
2. **Issue Resolution** — No known functional/security bugs; docs updated (spec & API examples); tests for critical paths pass
3. **MCP Pull Request** — Feature branch with descriptive name; PR via MCP GitHub with scope/risks/user impact, constitution checklist, examples; review/validate/merge

**Rationale**: Prevents technical debt accumulation. Ensures each increment is shippable, documented, and validated before building atop it.

## Technical Values

- **MONOLITH FIRST** — Split later if pain proves it. Premature distribution multiplies complexity.
- **POSTGRESQL PRIMARY** — Core data, full-text search, queues where feasible. Redis only for cache/session.
- **GRAPHQL BY DEFAULT** — Single endpoint with DataLoader to kill N+1 queries. Consistent error handling.
- **STRUCTURED LOGS** — `slog` everywhere; no noisy prints. Logs must be machine-parseable for alerts.
- **FAIL GRACEFULLY** — Return empty arrays and typed errors; do not 500 on partial failures. Surface degraded state to users.

## Domain Constraints

- **LITHUANIAN FIRST** — Diacritics, tokenization, basic lemma/synonym maps for common grocery categories/brands. Search must handle "pieno gaminiai" → "dairy products" equivalence.
- **NO FABRICATED PRICES** — If the flyer doesn't provide a concrete SKU+price (e.g., category "–25%"), we NEVER invent € values. Show discount badge only; no estimation.
- **MAX 1–2 STORES PER PLAN** — Users won't visit 3+ shops. Enforce a store cap. Wizard must respect `maxStores` (default 1; allow 2). Never propose more.

## Quality Standards

- **SEARCH < 500 ms (p95)** for Lithuanian queries over active flyers
- **LISTS PERSIST** across flyer rotations (weekly expiries don't wipe lists)
- **EXTRACTION CONTINUES** — Bad pages don't block the batch; log errors and proceed
- **AUTH SECURE** — JWT + bcrypt; validate inputs; typed error mapping to GraphQL codes
- **TESTS EXIST** — BDD/acceptance for critical flows (search, add to list, wizard suggestions, confirm)

## Agent Operating Rules (NON-NEGOTIABLE)

1. **No hallucinated prices/units** — If SKU price/size is missing, surface discount badge only
2. **Same-brand first** — If it exists anywhere, show it (even if in another shop)
3. **Cap stores** — Respect `maxStores` (default 1; allow 2). Never propose more
4. **Deterministic matching** — Use fixed scoring (brand > store > size > price) with documented tie-breakers
5. **Revalidate on confirm** — Before applying wizard decisions, re-fetch candidates and block if any changed/expired
6. **Idempotent mutations** — All wizard mutations accept optional idempotency key; applying twice must be no-op
7. **Typed errors only** — Map to GraphQL codes (VALIDATION_ERROR, FORBIDDEN, NOT_FOUND, etc.)

### Wizard Guardrails (MVP)

**What the wizard MUST do now:**

- Trigger only for expired flyer-linked items (`origin=flyer`)
- For each item, run two-pass search:
  1. **Strong**: brand + canonical name (+ original store)
  2. **Loose**: canonical name only (allow cross-brand)
- Score candidates:
  - +3 same brand
  - +2 original store
  - +2 user preferred store
  - +1 size within ±20%
  - +1 cheaper than previous
  - Tie-break: lower price wins
- Store selection:
  - Compute coverage per store; pick top store; add a second only if it increases covered items by ≥2 or reduces total by ≥ €X (config)
- Confidence thresholds for auto-apply:
  - CONSERVATIVE: only exact priced SKU matches (≥0.90)
  - BALANCED: strong matches (≥0.75)
  - AGGRESSIVE: broader matches (≥0.55), still never exceeding `maxStores`

**What the wizard MUST NOT do:**

- Do not replace items using category/brand % promos when no concrete SKU price exists (show as advisory)
- Do not propose more than `maxStores`
- Do not hide an available same-brand option

## Shopping List Data Lineage (Required Fields)

Each list item MUST track:

- `origin`: `flyer` | `free_text`
- `product_master_id` (nullable for free text)
- `flyer_product_id` (nullable)
- `store_id` (nullable)

On confirm, write `OfferSnapshot` (store, product/flyer ref, price, estimated flag, valid_to) for historical accuracy and lineage tracing.

## Search & Matching Rules

- **Normalization**: diacritics fold, case fold, stop-word list tuned for LT groceries, small lemma/synonym map (e.g., padažams→sauce, jogurtams→yogurt)
- **Query plan**: prefer active flyers, then historical base if needed
- **Ranking**: fixed weights (above). No ML in MVP. Adjust weights only via config.

## Minimal Metrics (Prometheus)

Required counters:

- `wizard_items_flagged_total`
- `wizard_suggestions_returned_total`
- `wizard_acceptance_rate` (accepted/kept/skipped)
- `wizard_selected_store_count` (histogram)
- `wizard_latency_ms` (p95)
- `search_latency_ms` (p95)
- Alert if coverage = 0 for >15 min after ingestion

No Grafana artwork marathons. One ops dashboard is enough.

## Decision Framework

When evaluating any change, apply in order:

1. **User impact**: Does this help plan a cheaper/shorter trip this week? If not, skip.
2. **Stack fit**: Can PostgreSQL/Redis/GraphQL do it? Use that.
3. **Honesty**: Do we actually have the data to claim this? If not, label it advisory.
4. **Operability**: Can we observe it with logs + minimal Prometheus? If yes, ship.

## Success Metrics (Minimal, not vanity)

- ✓ Flyers ingested, pages parsed
- ✓ Products extracted & searchable
- ✓ Shopping lists CRUD works
- ✓ Auth works (register/login)
- ✓ Wizard can migrate expired items
- ✓ API stays up (no crash loops)

## Governance

### Development Workflow

- **Branching**: `###-short-feature-name`
- **Commits**: clear, scoped; reference issue IDs
- **Pre-merge checks**: `make fmt`, `make lint`, `make generate`, `make test`
- **DataLoader**: mandatory for GraphQL relations to prevent N+1 queries
- **Error handling**: wrap with context, map to GraphQL typed errors

### MCP Pull Request Requirements (Agent Checklist)

Every PR MUST verify:

- [ ] Constitution section(s) touched? Reference + diffs
- [ ] `maxStores` enforced in code path
- [ ] Two-pass search implemented (strong + loose)
- [ ] Same-brand surfaced (if exists)
- [ ] Revalidation before confirm
- [ ] Idempotency on mutating endpoints
- [ ] Tests: expired→wizard, ranking order, store cap, revalidation
- [ ] Metrics counters added/updated
- [ ] No fabricated prices for % promos

### Amendment Process

- Change requires PR with migration notes (if schema touched)
- Verify against Core Principles and Wizard Guardrails
- Semantic versioning:
  - **MAJOR**: Breaking principle/governance changes
  - **MINOR**: New guardrails/agent rules
  - **PATCH**: Clarifications/wording

### Compliance

- Constitution supersedes all other practices
- All PRs/reviews must verify compliance with Core Principles and Agent Operating Rules
- Complexity must be justified via Decision Framework
- Use `AGENTS.md` for refactoring-specific safety rules (complementary to this constitution)

**Version**: 1.2.0 | **Ratified**: 2025-11-15 | **Last Amended**: 2025-11-16
