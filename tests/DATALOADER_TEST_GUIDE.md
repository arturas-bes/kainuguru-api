# DataLoader Performance Testing Guide

## Overview
This guide helps you verify that the DataLoader implementation is working correctly and reducing N+1 queries.

---

## Prerequisites

1. **Enable Database Query Logging**

Add to your database connection configuration:

```go
// In cmd/api/main.go or wherever you configure the database
db := bun.NewDB(sqldb, pgdialect.New())

// Enable query logging for testing
db.AddQueryHook(bunDebugHook{})

type bunDebugHook struct{}

func (h bunDebugHook) BeforeQuery(ctx context.Context, event *bun.QueryEvent) context.Context {
    return ctx
}

func (h bunDebugHook) AfterQuery(ctx context.Context, event *bun.QueryEvent) {
    fmt.Printf("[SQL] %s\n", event.Query)
}
```

**OR** use environment variable:
```bash
export BUN_DEBUG=1
```

2. **Start the API Server**
```bash
make run
# or
go run cmd/api/main.go
```

3. **Access GraphQL Endpoint**
```
http://localhost:8080/graphql
```

---

## Test 1: Product List with Nested Fields (CRITICAL)

### Purpose
Verify that fetching 20 products with nested Store, Flyer, and ProductMaster relationships doesn't cause N+1 queries.

### Test Query
```graphql
query TestDataLoaderProducts {
  products(first: 20) {
    edges {
      node {
        id
        name
        currentPrice
        currency

        # These should trigger DataLoader batching
        store {
          id
          name
          code
          logoURL
        }

        flyer {
          id
          title
          validFrom
          validTo
        }

        productMaster {
          id
          canonicalName
          brand
        }
      }
    }
    pageInfo {
      hasNextPage
      hasPreviousPage
    }
  }
}
```

### Expected Results

**WITHOUT DataLoader (OLD - BAD):**
```
Query count: ~80-100 queries
- 1 query: SELECT products
- 20 queries: SELECT store WHERE id = ? (for each product)
- 20 queries: SELECT flyer WHERE id = ? (for each product)
- 20 queries: SELECT product_master WHERE id = ? (for each product)
- Additional overhead queries
```

**WITH DataLoader (NEW - GOOD):**
```
Query count: ~4-6 queries
- 1 query: SELECT products LIMIT 20
- 1 query: SELECT stores WHERE id IN (1,2,3...) [BATCHED]
- 1 query: SELECT flyers WHERE id IN (10,11,12...) [BATCHED]
- 1 query: SELECT product_masters WHERE id IN (5,6,7...) [BATCHED]
```

### Verification
1. Run the query
2. Check the server logs for SQL queries
3. Count the number of queries
4. Look for `WHERE id IN (...)` patterns (batched queries)

**Pass Criteria:** ≤ 10 total queries (ideally 4-6)

---

## Test 2: Shopping List Items with Users

### Purpose
Verify that User DataLoader batches user lookups correctly.

### Test Query
```graphql
query TestDataLoaderShoppingList {
  shoppingLists(first: 1) {
    edges {
      node {
        id
        name
        items(first: 20) {
          edges {
            node {
              id
              description
              quantity

              # Should use UserLoader
              user {
                id
                email
                fullName
              }

              # Should use UserLoader (if checked)
              checkedByUser {
                id
                email
              }

              # Should use StoreLoader
              store {
                id
                name
              }

              # Should use ProductMasterLoader
              productMaster {
                id
                canonicalName
              }
            }
          }
        }
      }
    }
  }
}
```

### Expected Results

**WITHOUT DataLoader:**
```
Query count: ~60-80 queries
```

**WITH DataLoader:**
```
Query count: ~5-8 queries
- 1 query: SELECT shopping_lists
- 1 query: SELECT shopping_list_items
- 1 query: SELECT users WHERE id IN (...) [BATCHED]
- 1 query: SELECT stores WHERE id IN (...) [BATCHED]
- 1 query: SELECT product_masters WHERE id IN (...) [BATCHED]
```

**Pass Criteria:** ≤ 10 total queries

---

## Test 3: Flyer with Store

### Purpose
Verify Store DataLoader in flyer queries.

### Test Query
```graphql
query TestDataLoaderFlyers {
  flyers(first: 20) {
    edges {
      node {
        id
        title
        validFrom
        validTo

        # Should use StoreLoader
        store {
          id
          name
          code
          logoURL
        }
      }
    }
  }
}
```

### Expected Results

**WITHOUT DataLoader:**
```
Query count: ~21 queries
- 1 query: SELECT flyers
- 20 queries: SELECT store WHERE id = ?
```

**WITH DataLoader:**
```
Query count: 2 queries
- 1 query: SELECT flyers LIMIT 20
- 1 query: SELECT stores WHERE id IN (...) [BATCHED]
```

**Pass Criteria:** ≤ 3 total queries

---

## Test 4: Complex Nested Query

### Purpose
Stress test with maximum nesting.

### Test Query
```graphql
query TestDataLoaderComplex {
  stores(first: 5) {
    edges {
      node {
        id
        name

        flyers(first: 3) {
          edges {
            node {
              id
              title

              products(first: 10) {
                edges {
                  node {
                    id
                    name

                    productMaster {
                      id
                      canonicalName
                    }
                  }
                }
              }
            }
          }
        }
      }
    }
  }
}
```

### Expected Results

**Calculation:**
- 5 stores
- 3 flyers per store = 15 flyers
- 10 products per flyer = 150 products
- Each product may have a product master

**WITHOUT DataLoader:**
```
Query count: ~200+ queries
- Massive N+1 problem
```

**WITH DataLoader:**
```
Query count: ~5-10 queries
- 1 query: SELECT stores
- 1 query: SELECT flyers WHERE store_id IN (...)
- 1 query: SELECT products WHERE flyer_id IN (...)
- 1 query: SELECT product_masters WHERE id IN (...)
```

**Pass Criteria:** ≤ 15 total queries

---

## Performance Benchmarking

### Using cURL with Timing

```bash
# Test without DataLoader (simulate by removing middleware)
time curl -X POST http://localhost:8080/graphql \
  -H "Content-Type: application/json" \
  -d '{
    "query": "query { products(first: 20) { edges { node { id name store { name } flyer { title } productMaster { canonicalName } } } } }"
  }'

# Expected: 500-800ms (before optimization)
# Expected: 50-100ms (after optimization)
```

### Load Testing with Apache Bench

```bash
# Install Apache Bench
# macOS: brew install httpd
# Linux: sudo apt-get install apache2-utils

# Create query file
cat > query.json << 'EOF'
{
  "query": "query { products(first: 20) { edges { node { id name store { name } flyer { title } productMaster { canonicalName } } } } }"
}
EOF

# Run 100 requests with 10 concurrent
ab -n 100 -c 10 -p query.json -T application/json http://localhost:8080/graphql

# Look for:
# - Requests per second (should be higher with DataLoader)
# - Time per request (should be lower with DataLoader)
# - 95th percentile response time
```

---

## Debugging Tips

### 1. Check DataLoader is in Context

Add temporary logging in resolvers:
```go
func (r *productResolver) Store(ctx context.Context, obj *models.Product) (*models.Store, error) {
    loaders := dataloaders.FromContext(ctx)
    if loaders == nil {
        panic("DataLoaders not found in context!")
    }
    return loaders.StoreLoader.Load(ctx, obj.StoreID)()
}
```

### 2. Verify Batch Sizes

Add logging to dataloader.go:
```go
func batchStoreLoader(service services.StoreService) dataloader.BatchFunc[int, *models.Store] {
    return func(ctx context.Context, keys []int) []*dataloader.Result[*models.Store] {
        fmt.Printf("[DataLoader] Batching %d store IDs: %v\n", len(keys), keys)
        // ... rest of implementation
    }
}
```

### 3. PostgreSQL Query Logging

Enable PostgreSQL query logging:
```sql
ALTER DATABASE kainuguru SET log_statement = 'all';
ALTER DATABASE kainuguru SET log_duration = 'on';
```

Then check logs:
```bash
tail -f /usr/local/var/postgres/server.log
# or
tail -f /var/log/postgresql/postgresql-15-main.log
```

### 4. Count Queries with SQL

```sql
-- Reset statistics
SELECT pg_stat_reset();

-- Run your GraphQL query

-- Check query count
SELECT
    calls,
    query
FROM pg_stat_statements
WHERE query NOT LIKE '%pg_stat%'
ORDER BY calls DESC
LIMIT 20;
```

---

## Success Criteria Checklist

- [ ] ✅ Build compiles without errors
- [ ] ✅ Test 1: Product list uses ≤10 queries
- [ ] ✅ Test 2: Shopping list uses ≤10 queries
- [ ] ✅ Test 3: Flyer list uses ≤3 queries
- [ ] ✅ Test 4: Complex nested uses ≤15 queries
- [ ] ✅ Response time improved by >50%
- [ ] ✅ No panic/errors in logs
- [ ] ✅ DataLoader batching visible in logs

---

## Common Issues & Solutions

### Issue: "DataLoaders not found in context"
**Solution:** Ensure middleware is registered in graphql.go handler

### Issue: Still seeing N+1 queries
**Solution:**
1. Check resolver is using `dataloaders.FromContext(ctx)`
2. Verify service has `GetByIDs()` method
3. Check DataLoader is created in NewLoaders()

### Issue: Panic on nil pointer
**Solution:** Check for nil values before loading:
```go
if obj.ProductMasterID == nil {
    return nil, nil
}
```

### Issue: Duplicate queries still appearing
**Solution:** Check batch window timing (default 10ms) - may need adjustment for slow networks

---

## Next Steps After Verification

Once DataLoader is verified working:

1. **Add Metrics**
   - Track query counts
   - Monitor response times
   - Set up alerts for regressions

2. **Production Deployment**
   - Enable in staging first
   - Monitor error rates
   - Gradual rollout

3. **Documentation**
   - Update API docs
   - Add performance notes
   - Create runbook

4. **Further Optimizations**
   - Add database indexes
   - Implement DateTime scalar
   - Fix ID type consistency

---

## Support

If you encounter issues:
1. Check server logs for errors
2. Verify database queries in logs
3. Review DataLoader configuration
4. Check GraphQL handler middleware

For performance questions, refer to:
- GRAPHQL_OPTIMIZATION_PROGRESS.md
- Original analysis report
