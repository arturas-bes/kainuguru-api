<!--
Sync Impact Report (2025-11-05):
- Version change: 1.0.0 → 1.1.0 (minor bump - new governance procedures added)
- Modified principles: V. Phase Quality Gates (new principle added)
- Added sections: Phase Quality Gates governance, MCP Pull Request Requirements
- Governance section: Enhanced with mandatory phase analysis and PR workflows
- Templates requiring updates:
  ✅ plan-template.md (Constitution Check section already supports new gates)
  ✅ tasks-template.md (checkpoint structure aligns with phase requirements)
  ✅ spec-template.md (no changes needed - requirements already support phased delivery)
- Follow-up TODOs: None (all requirements integrated)
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

### V. Phase Quality Gates
After each implementation phase, ALL changes MUST be deeply analyzed and
issues MUST be fixed before proceeding to the next phase. Every phase
MUST conclude with a pull request created and merged via MCP GitHub
integration. No phase progression without comprehensive validation and
proper version control workflow.

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

### Development Workflow

- Reference KAINUGURU_MVP_FINAL.md for all decisions
- Use TODO comments liberally for future work
- Test critical paths, not edge cases
- Constitution supersedes all other practices

### Phase Quality Gates (MANDATORY)

Every implementation phase MUST follow this workflow:

1. **Deep Analysis**: Thoroughly analyze ALL changes made during the phase
   - Review code quality and adherence to principles
   - Identify and document any technical debt introduced
   - Verify all TODOs are appropriately documented
   - Check compliance with Lithuanian language requirements
   - Validate performance against quality standards

2. **Issue Resolution**: Fix ALL identified issues before phase completion
   - Address code quality problems
   - Resolve security vulnerabilities
   - Fix functionality bugs
   - Update documentation gaps
   - No phase can be considered complete with known issues

3. **Pull Request Workflow**: MANDATORY for every phase completion
   - Create feature branch with descriptive name (###-feature-name format)
   - Commit all phase changes with clear commit messages
   - Create pull request via MCP GitHub integration
   - Include comprehensive description of phase changes
   - Merge pull request only after review and validation
   - Delete feature branch after successful merge

### Amendment Process

- Amendments require documentation and migration plan
- All PRs must verify compliance with core principles
- Version increments follow semantic versioning:
  - MAJOR: Backward incompatible governance/principle changes
  - MINOR: New principles or expanded governance procedures
  - PATCH: Clarifications, wording fixes, non-semantic refinements

**Version**: 1.1.0 | **Ratified**: TODO(RATIFICATION_DATE) | **Last Amended**: 2025-11-05