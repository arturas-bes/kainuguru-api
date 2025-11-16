# Spec-Kit Commands for Shopping List Migration Wizard

## Overview
This document contains the `/specify` and `/plan` commands for using GitHub's spec-kit to implement the Shopping List Migration Wizard feature.

---

## 1. /specify Command

Use this command to generate the feature specification:

```
/specify

Feature Name: Shopping List Migration Wizard

Description:
A wizard-based system that helps users replace expired flyer products in their shopping lists with currently available alternatives from active flyers. The system detects when flyer products expire, presents filtered and ranked alternatives using a two-pass search strategy (brand-aware substitution), and allows users to make decisions item-by-item or in bulk.

User Stories:
1. As a user with a shopping list containing expired flyer items, I want to be notified when products are no longer available so I can update my list with current alternatives
2. As a user, I want to see similar products from current flyers ranked by relevance (same brand, store, size) so I can make quick replacement decisions
3. As a user, I want to limit my shopping to maximum 2 stores to minimize travel time
4. As a user, I want to see same-brand alternatives when available, even if they're in different stores
5. As a user, I want explanations for why products are suggested so I can make informed decisions

Key Requirements:
- Detect expired flyer products automatically when flyers expire
- Use two-pass search: first try same brand + name, then name only for cross-brand substitutes
- Deterministic scoring with fixed weights (brand: 3.0, store: 2.0, size: 1.0, price: 1.0)
- Hard cap at maximum 2 stores per shopping plan
- Always show same-brand option if it exists anywhere
- Create immutable offer snapshots for historical accuracy
- Session management with 30-minute timeout and state persistence
- Support both item-by-item and bulk decision making
- Provide confidence scores and explanations for each suggestion
- Validate data freshness before applying changes (detect if flyers changed mid-session)

Technical Constraints:
- Must integrate with existing SearchService.FuzzySearchProducts (no duplicate search logic)
- Must use existing ProductMaster canonical names and brands
- Must be backward compatible with current shopping_list_items table
- GraphQL schema changes must be additive only (no breaking changes)
- Must preserve existing shopping_list_migration_service functionality
- Maximum 2 stores per plan (hard limit)
- Never fabricate prices - only use actual flyer prices
- Session data in Redis with 30-minute TTL
- Deterministic ranking (same input must produce same output)

Success Criteria:
- Wizard completion rate >70%
- Average decision time <15 seconds per item
- Suggestion acceptance rate >80%
- Same-brand products shown when available 100% of the time
- Store cap never exceeded (max 2 stores)
- Zero data corruption from stale sessions
- P95 suggestion generation <1 second

Out of Scope:
- Machine learning for ranking (use fixed weights initially)
- Automatic enrichment of free-text shopping lists (manual trigger only)
- Category-level promotions (e.g., "20% off all dairy")
- More than 2 stores in a shopping plan
- Historical price prediction
- Cross-user recommendation system
```

---

## 2. /plan Command

Use this command after /specify to generate the implementation plan:

```
/plan

Technical Approach:
We'll implement a wizard-based migration system that integrates with the existing search and product master services. The solution uses a two-pass search strategy for brand-aware substitution, deterministic scoring for consistent results, and session management for state persistence.

Architecture Overview:
1. Data Layer: Extend shopping_list_items with origin tracking, add offer_snapshots table
2. Service Layer: Create WizardService, integrate with existing SearchService
3. API Layer: GraphQL mutations for wizard flow, maintaining backward compatibility
4. Worker Layer: FlyerExpirationWorker to detect and trigger migrations
5. Cache Layer: Redis for session storage with 30-minute TTL

Key Components:

1. Data Model Extensions:
   - shopping_list_items: Add origin, flyer_product_id, store_id, migration tracking
   - offer_snapshots: New table for immutable decision history
   - wizard_sessions: Session state management
   - user_migration_preferences: Store user preferences

2. Two-Pass Search Integration:
   - Pass 1: SearchService.FuzzySearchProducts with brand + name query
   - Pass 2: SearchService.FuzzySearchProducts with name only (if needed)
   - Reuse existing similarity scores from search service

3. Deterministic Scoring Engine:
   - Fixed weights: brand(3.0), original_store(2.0), preferred_store(2.0), size_match(1.0), price_better(1.0)
   - Tie-breaking by price ascending
   - No ML initially, can add later

4. Store Selection Strategy:
   - Greedy algorithm: highest item coverage, lowest total price
   - Hard cap at 2 stores maximum
   - Second store only if: +2 items coverage OR €5+ savings

5. Session Management:
   - Redis-backed sessions with 30-minute timeout
   - Dataset version tracking for staleness detection
   - Idempotency keys to prevent double-apply
   - State recovery on disconnect

6. GraphQL API:
   - startWizard: Initialize session with filters
   - getItemSuggestions: Get ranked alternatives
   - recordDecision: Save user choice
   - completeWizard: Apply all changes atomically
   - Additive schema changes only

Implementation Phases:

Phase 1 - Data Foundation (3 days):
- Database migrations for new fields
- Create offer_snapshots table
- Update Go models
- Create migration scripts

Phase 2 - Search Integration (2 days):
- Implement two-pass search strategy
- Integrate with existing SearchService
- Create deterministic scoring engine
- Add store selection logic

Phase 3 - Wizard Service (3 days):
- Build WizardService with session management
- Implement Redis session storage
- Add dataset version tracking
- Create validation logic

Phase 4 - GraphQL API (2 days):
- Add new types and mutations
- Implement resolvers
- Add idempotency support
- Create error handling

Phase 5 - Worker Integration (2 days):
- Create FlyerExpirationWorker
- Integrate with existing worker system
- Add notification triggers
- Implement batch processing

Phase 6 - Testing & Validation (3 days):
- Integration tests for expired item flow
- Test store cap enforcement
- Verify deterministic ranking
- Test session revalidation
- Performance testing

Rollout Strategy:
1. Deploy database changes (backward compatible)
2. Deploy service code behind feature flag
3. Enable for internal testing (5% traffic)
4. Gradual rollout (25%, 50%, 100%)
5. Monitor metrics and adjust

Risk Mitigation:
- Feature flag for easy rollback
- Shadow mode testing before launch
- Comprehensive integration tests
- Session recovery mechanisms
- Data validation before applying changes
- Audit logging for all decisions

Dependencies:
- Existing SearchService must be stable
- ProductMaster data quality >80% coverage
- Redis cluster availability
- Current flyer ingestion pipeline working

Monitoring:
- wizard_items_flagged_total
- wizard_suggestions_returned
- wizard_acceptance_rate
- wizard_selected_store_count
- wizard_latency_ms (P95 target <1s)
```

---

## 3. Spec-Kit Workflow

### Step 1: Initialize Spec-Kit
```bash
# In your project directory
speckit init
```

### Step 2: Generate Specification
```bash
# Copy the /specify command content above and run:
/specify
# This creates spec.md with the feature specification
```

### Step 3: Generate Implementation Plan
```bash
# Copy the /plan command content above and run:
/plan
# This creates plan.md with technical details
```

### Step 4: Generate Tasks
```bash
# After spec and plan are created:
/tasks
# This generates tasks.md with actionable implementation tasks
```

### Step 5: Validate Consistency
```bash
# Check all artifacts are aligned:
/speckit.analyze
```

### Step 6: Begin Implementation
```bash
# Start executing tasks:
/speckit.implement
```

---

## 4. Key Technical Decisions for Spec-Kit

When spec-kit asks for clarification, use these decisions:

### Database Strategy
- **Choice:** Extend existing tables (additive only)
- **Rationale:** Maintains backward compatibility

### Search Strategy
- **Choice:** Two-pass search using existing FuzzySearchProducts
- **Rationale:** Reuses proven search infrastructure

### Scoring Algorithm
- **Choice:** Fixed weights, deterministic
- **Rationale:** Predictable results, can add ML later

### Store Selection
- **Choice:** Greedy algorithm with hard cap of 2
- **Rationale:** Simple, effective, respects user constraints

### Session Storage
- **Choice:** Redis with 30-minute TTL
- **Rationale:** Fast, reliable, automatic cleanup

### API Design
- **Choice:** GraphQL with additive changes only
- **Rationale:** No breaking changes for existing clients

### Testing Strategy
- **Choice:** Integration tests focusing on 5 core scenarios
- **Rationale:** Validates real user flows

### Rollout Strategy
- **Choice:** Feature flag with gradual rollout
- **Rationale:** Safe deployment with easy rollback

---

## 5. Expected Spec-Kit Artifacts

After running the commands, you should have:

1. **spec.md** - Feature specification with requirements
2. **plan.md** - Technical implementation plan
3. **tasks.md** - Detailed task breakdown
4. **constitution.md** - Project principles (if configured)

Each artifact will be cross-referenced and validated for consistency.

---

## 6. Task Prioritization Guide

When spec-kit generates tasks, prioritize them as:

### P0 - Critical Path (Must Have for MVP)
- Database migrations for lineage tracking
- Two-pass search implementation
- Basic wizard session management
- Store cap enforcement
- Core GraphQL mutations

### P1 - Important (Should Have)
- Offer snapshots for history
- Session revalidation
- Idempotency support
- Comprehensive explanations
- Bulk decision support

### P2 - Nice to Have (Could Have)
- User preference learning
- Advanced filtering options
- Performance optimizations
- Extended metrics
- A/B testing framework

---

## 7. Validation Checklist

Before implementation, ensure spec-kit artifacts address:

- [ ] Integration with existing SearchService
- [ ] Backward compatibility requirements
- [ ] Store cap enforcement (max 2)
- [ ] Same-brand preference logic
- [ ] Session timeout handling
- [ ] Data staleness detection
- [ ] Idempotency for mutations
- [ ] Fixed weight scoring
- [ ] Error handling strategy
- [ ] Rollback procedures

---

## Notes for Spec-Kit Usage

1. **Constitution Setup**: If you haven't set up a constitution, spec-kit may prompt you. Focus on principles like:
   - Backward compatibility first
   - Reuse over rebuild
   - Data integrity paramount
   - User trust through transparency

2. **Clarification Questions**: Spec-kit may ask for clarification. Refer to the Shopping List Migration Spec V2.0 for answers.

3. **Task Dependencies**: Ensure spec-kit captures the dependency chain:
   - Database changes → Service implementation → API layer → Testing

4. **Acceptance Criteria**: Each task should have clear acceptance criteria from the specification.

5. **Integration Points**: Emphasize integration with existing services to avoid duplication.