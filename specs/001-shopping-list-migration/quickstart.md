# Quick Start Guide: Shopping List Migration Wizard

**Feature**: Shopping List Migration Wizard
**Date**: 2025-11-15
**Purpose**: Test and verify the wizard implementation

## Prerequisites

- Docker and docker-compose running
- PostgreSQL with test data
- Redis server running
- GraphQL playground accessible

## Setup Test Environment

### 1. Apply Database Migrations

```bash
# Run migration to add wizard tables
go run cmd/migrate/main.go up

# Verify tables created
psql -U kainuguru -d kainuguru_db -c "\dt wizard_*"
psql -U kainuguru -d kainuguru_db -c "\dt offer_*"
```

### 2. Start Services

```bash
# Start all services
docker-compose up -d

# Verify Redis is running
redis-cli ping

# Check GraphQL endpoint
curl http://localhost:8080/graphql
```

### 3. Create Test Data

```sql
-- Insert test shopping list with flyer products
INSERT INTO shopping_lists (user_id, name)
VALUES (1, 'Test Migration List');

-- Add items linked to expiring flyer products
INSERT INTO shopping_list_items
(shopping_list_id, name, linked_product_id, store_id, flyer_id, quantity)
VALUES
(1, 'Coca-Cola 2L', 12345, 5, 100, 2),
(1, 'Žemaitijos Milk 1L', 12346, 5, 100, 1),
(1, 'Gardėsis Bread', 12347, 5, 100, 1);

-- Mark flyer as expired
UPDATE flyers SET valid_until = NOW() - INTERVAL '1 day'
WHERE id = 100;
```

## Testing the Wizard Flow

### Step 1: Detect Expired Items

```graphql
query CheckExpiredItems {
  hasExpiredItems(shoppingListId: "1") {
    hasExpiredItems
    expiredCount
    items {
      id
      productName
      brand
      originalPrice
      expiryDate
    }
    suggestedAction
  }
}
```

**Expected Response**:
```json
{
  "data": {
    "hasExpiredItems": {
      "hasExpiredItems": true,
      "expiredCount": 3,
      "items": [...],
      "suggestedAction": "Start migration wizard to update expired items"
    }
  }
}
```

### Step 2: Start Wizard Session

```graphql
mutation StartMigration {
  startWizard(input: {
    shoppingListId: "1"
    autoMode: false
    maxStores: 2
  }) {
    id
    status
    expiredItems {
      id
      productName
      brand
      originalPrice
    }
    progress {
      totalItems
      percentComplete
    }
    expiresAt
  }
}
```

**Save the session ID** from response for next steps.

### Step 3: Get Suggestions for First Item

```graphql
query GetSuggestions {
  getItemSuggestions(input: {
    sessionId: "YOUR_SESSION_ID"
    itemId: "FIRST_ITEM_ID"
    maxResults: 5
    minConfidence: 0.5
  }) {
    id
    product {
      name
      brand
      price
      store {
        name
      }
    }
    score
    confidence
    explanation
    scoreBreakdown {
      brandScore
      storeScore
      sizeScore
      priceScore
      totalScore
    }
  }
}
```

### Step 4: Record Decision

```graphql
mutation RecordItemDecision {
  recordDecision(input: {
    sessionId: "YOUR_SESSION_ID"
    itemId: "FIRST_ITEM_ID"
    decision: REPLACE
    suggestionId: "SELECTED_SUGGESTION_ID"
  }) {
    id
    progress {
      currentItem
      totalItems
      itemsMigrated
      percentComplete
    }
    selectedStores {
      store {
        name
      }
      itemCount
      totalPrice
    }
  }
}
```

### Step 5: Test Bulk Accept

```graphql
mutation BulkAccept {
  bulkAcceptSuggestions(input: {
    sessionId: "YOUR_SESSION_ID"
    itemIds: ["ITEM_2_ID", "ITEM_3_ID"]
    minConfidence: 0.8
  }) {
    id
    progress {
      itemsMigrated
      itemsSkipped
      percentComplete
    }
  }
}
```

### Step 6: Complete Wizard

```graphql
mutation CompleteWizard {
  completeWizard(input: {
    sessionId: "YOUR_SESSION_ID"
    applyChanges: true
    savePreferences: true
  }) {
    success
    summary {
      totalItems
      itemsMigrated
      itemsSkipped
      totalSavings
      storesUsed {
        name
      }
      averageConfidence
    }
    errors {
      code
      message
    }
  }
}
```

## Testing Edge Cases

### Test 1: Store Limit Enforcement

```graphql
# Try to select items from 3+ stores
# Should get error when exceeding 2 stores
```

### Test 2: Session Timeout

```bash
# Wait 31 minutes after starting session
# Try to continue - should get SessionExpiredError
```

### Test 3: Data Staleness Detection

```sql
-- Change flyer data mid-session
UPDATE flyers SET dataset_version = dataset_version + 1;
```

```graphql
# Try to complete wizard
# Should get StaleDataError
```

### Test 4: Same-Brand Priority

```graphql
# Verify same-brand products appear first
# Even from different stores
```

## Performance Testing

### Load Test Script

```bash
#!/bin/bash
# Create 100 concurrent wizard sessions

for i in {1..100}; do
  curl -X POST http://localhost:8080/graphql \
    -H "Content-Type: application/json" \
    -d '{"query":"mutation { startWizard(input: { shoppingListId: \"'$i'\" }) { id } }"}' &
done

wait
echo "Load test complete"
```

### Measure Latencies

```graphql
# Run with timing
time {
  getItemSuggestions(input: {
    sessionId: "..."
    itemId: "..."
  }) {
    # Should complete in <1 second (P95)
  }
}
```

## Monitoring & Debugging

### Check Redis Sessions

```bash
# List all wizard sessions
redis-cli keys "wizard:session:*"

# Inspect session data
redis-cli hgetall "wizard:session:UUID"

# Check TTL
redis-cli ttl "wizard:session:UUID"
```

### Database Queries

```sql
-- Active sessions
SELECT * FROM wizard_sessions
WHERE status = 'active'
ORDER BY started_at DESC;

-- Migration history
SELECT * FROM offer_snapshots
WHERE session_id = 'UUID'
ORDER BY created_at;

-- User statistics
SELECT
  COUNT(*) as sessions,
  AVG(completion_rate) as avg_completion,
  SUM(items_migrated) as total_migrated
FROM wizard_sessions
WHERE user_id = 1;
```

### GraphQL Introspection

```graphql
# Verify schema additions
{
  __type(name: "WizardSession") {
    fields {
      name
      type {
        name
      }
    }
  }
}
```

## Troubleshooting

### Common Issues

1. **Session Not Found**
   - Check Redis connection
   - Verify session hasn't expired
   - Check session ID format

2. **No Suggestions Returned**
   - Verify SearchService is running
   - Check if products exist in current flyers
   - Review search similarity thresholds

3. **Store Limit Exceeded**
   - Verify store selection logic
   - Check selectedStores in session

4. **Data Version Mismatch**
   - Resync dataset_version
   - Start new session

## Success Criteria Validation

- [ ] Wizard completion rate >70%
- [ ] Average decision time <15 seconds
- [ ] Suggestion acceptance rate >80%
- [ ] Same-brand shown when available
- [ ] Never exceed 2 stores
- [ ] P95 latency <1 second

## Rollback Procedure

If issues arise:

```bash
# 1. Disable feature flag
curl -X POST http://localhost:8080/admin/features \
  -d '{"wizard_enabled": false}'

# 2. Clear Redis sessions
redis-cli --scan --pattern "wizard:*" | xargs redis-cli del

# 3. Rollback database if needed
go run cmd/migrate/main.go down

# 4. Restore original shopping lists
psql -U kainuguru -d kainuguru_db < backup.sql
```

## Next Steps

After successful testing:

1. Enable for 5% of users
2. Monitor metrics for 24 hours
3. Gradually increase to 25%, 50%, 100%
4. Collect user feedback
5. Iterate on scoring algorithm

## Support

For issues or questions:
- Check logs: `docker-compose logs api`
- Review metrics: `http://localhost:8080/metrics`
- Database state: `psql -U kainuguru -d kainuguru_db`