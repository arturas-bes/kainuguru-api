# Shopping List Migration Wizard - Quickstart Guide

## Overview

The Shopping List Migration Wizard helps users update their shopping lists when flyer products expire. It provides intelligent, brand-aware suggestions and guides users through item-by-item decisions.

## Prerequisites

- Authentication token (JWT)
- Shopping list with expired items
- GraphQL endpoint: `https://api.kainuguru.com/graphql`

## Complete Wizard Flow

### Step 1: Check for Expired Items

```graphql
query CheckExpiredItems {
  myShoppingLists {
    id
    name
    expiredItemCount
    hasActiveWizardSession
    isLocked
  }
}
```

### Step 2: Start Wizard Session

```graphql
mutation StartWizard {
  startWizard(input: { shoppingListID: "123" }) {
    id
    status
    expiresAt
    selectedStores {
      id
      name
    }
    expiredItems {
      id
      itemID
      itemName
      brand
      suggestions {
        id
        productID
        productName
        brand
        price
        confidence
        explanation
      }
    }
  }
}
```

### Step 3: Make Decisions

#### Replace with Suggestion
```graphql
mutation AcceptSuggestion {
  recordDecision(input: {
    sessionID: "550e8400-e29b-41d4-a716-446655440000"
    itemID: "456"
    decision: REPLACE
    suggestionID: "789"
  }) {
    id
    status
  }
}
```

#### Skip Item
```graphql
mutation SkipItem {
  recordDecision(input: {
    sessionID: "550e8400-e29b-41d4-a716-446655440000"
    itemID: "456"
    decision: SKIP
  }) {
    id
    status
  }
}
```

#### Remove Item
```graphql
mutation RemoveItem {
  recordDecision(input: {
    sessionID: "550e8400-e29b-41d4-a716-446655440000"
    itemID: "456"
    decision: REMOVE
  }) {
    id
    status
  }
}
```

### Step 4: Confirm and Apply

```graphql
mutation ConfirmWizard {
  confirmWizard(input: {
    sessionID: "550e8400-e29b-41d4-a716-446655440000"
  }) {
    success
    result {
      itemsUpdated
      itemsDeleted
      storeCount
      totalEstimatedPrice
    }
  }
}
```

### Alternative: Cancel Wizard

```graphql
mutation CancelWizard {
  cancelWizard(input: {
    sessionID: "550e8400-e29b-41d4-a716-446655440000"
  }) {
    success
  }
}
```

## cURL Examples

### Start Wizard
```bash
curl -X POST https://api.kainuguru.com/graphql \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -d '{
    "query": "mutation StartWizard($input: StartWizardInput!) { startWizard(input: $input) { id status } }",
    "variables": {
      "input": {
        "shoppingListID": "123"
      }
    }
  }'
```

### Record Decision
```bash
curl -X POST https://api.kainuguru.com/graphql \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -d '{
    "query": "mutation RecordDecision($input: RecordDecisionInput!) { recordDecision(input: $input) { id } }",
    "variables": {
      "input": {
        "sessionID": "550e8400-e29b-41d4-a716-446655440000",
        "itemID": "456",
        "decision": "REPLACE",
        "suggestionID": "789"
      }
    }
  }'
```

### Confirm Wizard
```bash
curl -X POST https://api.kainuguru.com/graphql \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -d '{
    "query": "mutation ConfirmWizard($input: ConfirmWizardInput!) { confirmWizard(input: $input) { success } }",
    "variables": {
      "input": {
        "sessionID": "550e8400-e29b-41d4-a716-446655440000"
      }
    }
  }'
```

## Key Features

- **Session Expiration:** 30 minutes TTL
- **Store Limit:** Maximum 2 stores per session
- **Rate Limit:** 5 sessions per user per hour
- **List Locking:** Prevents concurrent wizard sessions
- **Idempotency:** Safe retry for all mutations
- **Revalidation:** Staleness detection before applying changes
- **ACID Transactions:** Atomic application of all decisions

## Error Handling

### Rate Limit Exceeded
```json
{
  "errors": [
    {
      "message": "rate limit exceeded: maximum 5 wizard sessions per hour"
    }
  ]
}
```

### List Already Locked
```json
{
  "errors": [
    {
      "message": "shopping list is already being migrated by another active wizard session"
    }
  ]
}
```

### Session Expired
```json
{
  "errors": [
    {
      "message": "session has expired"
    }
  ]
}
```

### Stale Data
```json
{
  "errors": [
    {
      "message": "revalidation failed: 2 products are stale or expired"
    }
  ]
}
```

## Constitution Compliance

- **FR-002:** Same-brand alternatives prioritized
- **FR-003:** Maximum 2 stores enforced
- **FR-004:** Deterministic scoring
- **FR-005:** Human-readable explanations
- **FR-016:** List locking prevents concurrent sessions
- **FR-020:** Origin tracking for wizard-selected products

## Metrics

Prometheus metrics tracked:
- `wizard_items_flagged_total{reason}`
- `wizard_suggestions_returned{has_same_brand}`
- `wizard_acceptance_rate_total{decision}`
- `wizard_selected_store_count{session_status}`
- `wizard_latency_ms{operation}`
- `wizard_sessions_total{status}`
- `wizard_revalidation_errors_total{error_type}`

## Technical Notes

- **Session Storage:** Redis with 30min TTL
- **Idempotency:** 24h cache
- **Transactions:** PostgreSQL ACID guarantees
- **N+1 Prevention:** DataLoaders for relations
- **Rate Limiting:** Redis sorted sets

## Support

For details see:
- `/docs/SHOPPING_LIST_MIGRATION_SPEC_V2.md`
- `/docs/api.md`
