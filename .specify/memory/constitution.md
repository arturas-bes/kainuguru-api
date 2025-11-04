<!--
Sync Impact Report (2025-11-04):
- Version change: N/A → 1.0.0 (initial constitution)
- Added principles: All core principles established
- Added sections: Technical Values, Constraints, Quality Standards, Decision Framework, Success Metrics
- Templates requiring updates: None (initial setup)
- Follow-up TODOs:
  - TODO(RATIFICATION_DATE): Set to project kickoff date once confirmed
-->

# Kainuguru MVP Constitution

## Core Principles

### I. Simplicity First
Use boring, proven technology. PostgreSQL for everything. No fancy tools.
Every technical decision must prioritize simplicity over sophistication.
Complex solutions require explicit justification and must demonstrate
measurable user value that simpler alternatives cannot provide.

### II. Ship Over Perfect
Working MVP beats perfect architecture. TODOs for optimization.
Features must be delivered functional first, optimized later. Code must
include TODO comments for future improvements rather than delaying shipment
for perfect implementation.

### III. User Value
Every feature must directly help Lithuanian shoppers save money.
No feature gets implemented without clear connection to helping users
find deals, compare prices, or manage their grocery shopping more efficiently.

### IV. Pragmatic Choices
ChatGPT without cost controls initially. Monitor, then optimize.
Choose the fastest path to working functionality. Use existing services
(ChatGPT, PostgreSQL features) over custom implementations. Optimization
comes after validation through real usage.

## Technical Values

- **MONOLITH FIRST** - Start simple, split later if needed
- **POSTGRESQL EVERYTHING** - Database, search, queue in one place
- **GRAPHQL FROM START** - Better developer experience, no REST phase
- **BASIC LOGGING** - Structured logs only, no Grafana/Prometheus
- **FAIL GRACEFULLY** - Return empty arrays, log errors, continue processing

## Constraints

- **NO COMPLEX METRICS** - Ship first, measure later
- **NO OVER-ENGINEERING** - Resist adding "nice to have" features
- **LITHUANIAN FOCUS** - All features must handle Lithuanian language properly

## Quality Standards

- **SEARCH WORKS** - Results in <500ms with Lithuanian text
- **LISTS PERSIST** - Shopping lists survive weekly data rotation
- **EXTRACTION CONTINUES** - Failed pages don't stop processing
- **AUTH SECURE** - Proper JWT + bcrypt, no shortcuts
- **TESTS EXIST** - BDD for critical paths, not 100% coverage

## Decision Framework

When in doubt, ask:
1. Does this help users find grocery deals? If no, skip it.
2. Can PostgreSQL do it? If yes, don't add new infrastructure.
3. Is it simpler to use ChatGPT? If yes, don't build complex logic.
4. Can we monitor it with logs? If yes, don't add metrics tools.

## Success Metrics (Simple)

- ✓ Flyers scraped successfully
- ✓ Products extracted (any success rate)
- ✓ Search returns results
- ✓ Shopping lists work
- ✓ Users can register and login
- ✓ Site doesn't crash

## Post-MVP TODOs

- Add cost monitoring for ChatGPT
- Implement OCR fallback strategies
- Add proper monitoring tools
- Optimize based on real usage
- Scale when we have users

## Governance

- Reference KAINUGURU_MVP_FINAL.md for all decisions
- Use TODO comments liberally for future work
- Test critical paths, not edge cases
- Constitution supersedes all other practices
- Amendments require documentation and migration plan
- All PRs must verify compliance with core principles

**Version**: 1.0.0 | **Ratified**: TODO(RATIFICATION_DATE) | **Last Amended**: 2025-11-04