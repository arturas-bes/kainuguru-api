# Backend Adjustments for Frontend Implementation

> Required and recommended backend changes to support the Kainuguru frontend

**Version:** 1.0
**Date:** 2025-12-10
**Priority Legend:** P0 = Blocker, P1 = Important, P2 = Nice-to-have

---

## Table of Contents

1. [Executive Summary](#1-executive-summary)
2. [Required Changes (P0)](#2-required-changes-p0)
3. [Recommended Improvements (P1)](#3-recommended-improvements-p1)
4. [Future Enhancements (P2)](#4-future-enhancements-p2)
5. [Schema Cleanup](#5-schema-cleanup)
6. [Performance Optimizations](#6-performance-optimizations)
7. [CORS & Security](#7-cors--security)
8. [API Documentation](#8-api-documentation)

---

## 1. Executive Summary

The current backend GraphQL API is **well-suited for frontend development**. Most features required for MVP are already implemented. This document outlines adjustments needed to optimize the frontend-backend integration.

### Status Overview

| Category | Status | Action Required |
|----------|--------|-----------------|
| Core Queries | ✅ Ready | None |
| Auth Mutations | ✅ Ready | Minor token handling |
| Shopping Lists | ✅ Ready | None |
| Wizard API | ✅ Ready | None |
| Search | ✅ Ready | None |
| Price Alerts | ✅ Ready | None |
| DateTime handling | ⚠️ Inconsistent | Standardization recommended |
| Image URLs | ⚠️ Relative paths | Base URL configuration needed |
| CORS | ❓ Unknown | Configuration required |

---

## 2. Required Changes (P0)

These changes are **blockers** for frontend development.

### 2.1 CORS Configuration

**Issue:** Frontend will run on a different origin (e.g., `https://kainuguru.lt`) than the API (`https://api.kainuguru.lt`).

**Required:**

```go
// cmd/api/server/server.go or middleware

func corsMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        origin := c.Request.Header.Get("Origin")

        allowedOrigins := []string{
            "http://localhost:3000",           // Development
            "https://kainuguru.lt",            // Production
            "https://www.kainuguru.lt",        // Production www
            "https://staging.kainuguru.lt",    // Staging
        }

        for _, allowed := range allowedOrigins {
            if origin == allowed {
                c.Header("Access-Control-Allow-Origin", origin)
                break
            }
        }

        c.Header("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
        c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization")
        c.Header("Access-Control-Allow-Credentials", "true")
        c.Header("Access-Control-Max-Age", "86400")

        if c.Request.Method == "OPTIONS" {
            c.AbortWithStatus(204)
            return
        }

        c.Next()
    }
}
```

**Environment Variables:**
```env
CORS_ALLOWED_ORIGINS=http://localhost:3000,https://kainuguru.lt
```

---

### 2.2 Image URL Base Path

**Issue:** `FlyerPage.imageURL` and `Product.imageURL` contain relative paths. Frontend needs absolute URLs.

**Current:**
```json
{
  "imageURL": "/flyers/maxima/2024-01/page-1.jpg"
}
```

**Option A: Backend returns absolute URLs (Recommended)**

Add environment variable and resolver logic:

```go
// config
type Config struct {
    CDNBaseURL string `env:"CDN_BASE_URL" default:"https://cdn.kainuguru.lt"`
}

// resolver
func (r *flyerPageResolver) ImageURL(ctx context.Context, obj *model.FlyerPage) (*string, error) {
    if obj.ImageURL == nil {
        return nil, nil
    }

    fullURL := r.config.CDNBaseURL + *obj.ImageURL
    return &fullURL, nil
}
```

**Option B: Frontend handles (fallback)**

Frontend can prepend base URL, but this adds coupling:

```typescript
const getImageUrl = (path: string | null) =>
  path ? `${process.env.NEXT_PUBLIC_CDN_URL}${path}` : null;
```

**Recommendation:** Option A - Backend should return full URLs.

---

### 2.3 Auth Token in Response Headers

**Issue:** Current `AuthPayload` returns tokens in response body. For security, `refreshToken` should optionally be in HTTP-only cookie.

**Current Schema:**
```graphql
type AuthPayload {
  user: User!
  accessToken: String!
  refreshToken: String!  # Exposed in body
  expiresAt: String!
  tokenType: String!
}
```

**Recommended Change:**

Keep current behavior but **also** set HTTP-only cookie:

```go
func (r *mutationResolver) Login(ctx context.Context, input model.LoginInput) (*model.AuthPayload, error) {
    // ... existing login logic ...

    // Set refresh token as HTTP-only cookie
    ginCtx := ctx.Value("GinContext").(*gin.Context)
    ginCtx.SetCookie(
        "refresh_token",
        payload.RefreshToken,
        int(time.Hour * 24 * 7 / time.Second), // 7 days
        "/",
        "", // Domain
        true,  // Secure (HTTPS only)
        true,  // HTTP-only
    )

    return payload, nil
}
```

Frontend can then use either approach:
- Body token (for mobile apps, SPAs that need explicit control)
- Cookie (for SSR, more secure by default)

---

## 3. Recommended Improvements (P1)

These changes improve developer experience and consistency.

### 3.1 DateTime Scalar Consistency

**Issue:** Schema mixes `DateTime` scalar and `String` for date fields.

**Current state:**

```graphql
# Uses DateTime (wizard.graphql)
type WizardSession {
  startedAt: DateTime!
  expiresAt: DateTime!
}

# Uses String (schema.graphql)
type Product {
  validFrom: String!    # Should be DateTime
  validTo: String!      # Should be DateTime
  createdAt: String!    # Should be DateTime
}
```

**Recommended:**

Standardize all date fields to `DateTime`:

```graphql
# schema.graphql
scalar DateTime

type Product {
  validFrom: DateTime!
  validTo: DateTime!
  createdAt: DateTime!
  updatedAt: DateTime!
}

type Flyer {
  validFrom: DateTime!
  validTo: DateTime!
  createdAt: DateTime!
  updatedAt: DateTime!
}

# ... same for all types with date fields
```

**Backend change:**

Ensure DateTime scalar marshals/unmarshals consistently:

```go
// internal/graphql/scalar/datetime.go
func MarshalDateTime(t time.Time) graphql.Marshaler {
    return graphql.WriterFunc(func(w io.Writer) {
        io.WriteString(w, strconv.Quote(t.Format(time.RFC3339)))
    })
}

func UnmarshalDateTime(v interface{}) (time.Time, error) {
    switch v := v.(type) {
    case string:
        return time.Parse(time.RFC3339, v)
    default:
        return time.Time{}, fmt.Errorf("invalid datetime")
    }
}
```

**Impact:** Frontend can remove date parsing utilities; GraphQL codegen will type dates correctly.

---

### 3.2 Add `homeSummary` Query

**Issue:** Frontend home page requires multiple queries. A combined query improves performance.

**Current:** Frontend must call:
- `currentFlyers`
- `productsOnSale`
- `stores` (for store list)

**Recommended:** Add aggregated query:

```graphql
type HomeSummary {
  currentFlyers: [Flyer!]!
  hotDeals: [Product!]!
  stores: [Store!]!
  totalProducts: Int!
  totalFlyers: Int!
}

extend type Query {
  homeSummary(storeIDs: [Int!], dealLimit: Int = 12): HomeSummary!
}
```

**Resolver:**

```go
func (r *queryResolver) HomeSummary(ctx context.Context, storeIDs []int, dealLimit *int) (*model.HomeSummary, error) {
    limit := 12
    if dealLimit != nil {
        limit = *dealLimit
    }

    // Parallel fetch with errgroup
    g, ctx := errgroup.WithContext(ctx)

    var flyers []*model.Flyer
    var deals []*model.Product
    var stores []*model.Store

    g.Go(func() error {
        var err error
        flyers, err = r.flyerRepo.CurrentFlyers(ctx, storeIDs, 10)
        return err
    })

    g.Go(func() error {
        var err error
        deals, err = r.productRepo.OnSale(ctx, storeIDs, limit)
        return err
    })

    g.Go(func() error {
        var err error
        stores, err = r.storeRepo.Active(ctx)
        return err
    })

    if err := g.Wait(); err != nil {
        return nil, err
    }

    return &model.HomeSummary{
        CurrentFlyers: flyers,
        HotDeals:      deals,
        Stores:        stores,
    }, nil
}
```

**Frontend benefit:** Single query for home page, better caching.

---

### 3.3 Add `shoppingListDetail` Query

**Issue:** Shopping list detail page needs list + items + wizard status. Currently requires multiple queries or complex nested selection.

**Recommended:**

```graphql
type ShoppingListDetail {
  list: ShoppingList!
  items: [ShoppingListItem!]!
  itemsByCategory: [CategoryGroup!]!
  wizardAvailable: Boolean!
  expiredItemCount: Int!
}

type CategoryGroup {
  category: String!
  items: [ShoppingListItem!]!
}

extend type Query {
  shoppingListDetail(id: Int!): ShoppingListDetail
}
```

---

### 3.4 Search Suggestions Endpoint

**Issue:** Autocomplete needs fast suggestions. Current `searchProducts` is too heavy.

**Recommended:**

```graphql
type SearchSuggestion {
  text: String!
  type: SuggestionType!
  count: Int
}

enum SuggestionType {
  PRODUCT
  CATEGORY
  BRAND
  RECENT
}

extend type Query {
  searchSuggestions(
    query: String!
    limit: Int = 5
  ): [SearchSuggestion!]!
}
```

**Resolver:** Quick prefix search on product names, categories, brands.

---

## 4. Future Enhancements (P2)

Lower priority improvements for later phases.

### 4.1 User Preferences Storage

For Phase 2+, store user preferences in backend:

```graphql
type UserPreferences {
  preferredStoreIds: [Int!]!
  preferredLanguage: String!
  emailNotifications: Boolean!
  pushNotifications: Boolean!
}

input UpdateUserPreferencesInput {
  preferredStoreIds: [Int!]
  preferredLanguage: String
  emailNotifications: Boolean
  pushNotifications: Boolean
}

extend type Query {
  myPreferences: UserPreferences!
}

extend type Mutation {
  updateMyPreferences(input: UpdateUserPreferencesInput!): UserPreferences!
}
```

### 4.2 Real-time Price Alert Notifications

For Phase 3, add subscriptions:

```graphql
extend type Subscription {
  priceAlertTriggered(userId: ID!): PriceAlertNotification!
}

type PriceAlertNotification {
  alert: PriceAlert!
  triggeredPrice: Float!
  product: Product!
  timestamp: DateTime!
}
```

### 4.3 Analytics Events

For tracking user behavior:

```graphql
input AnalyticsEventInput {
  eventType: String!
  entityType: String
  entityId: String
  metadata: String  # JSON string
}

extend type Mutation {
  trackEvent(input: AnalyticsEventInput!): Boolean!
}
```

---

## 5. Schema Cleanup

### 5.1 Remove Unused Fields

Review and remove if not used:

```graphql
# Product - potentially unused
type Product {
  extractionConfidence: Float!    # Internal use only?
  extractionMethod: String!       # Internal use only?
  requiresReview: Boolean!        # Admin only?
}

# Consider moving to separate AdminProduct type
```

### 5.2 Consolidate Connection Types

All connection types follow same pattern - consider using generics or code generation.

### 5.3 Input Validation

Add explicit validation in resolvers:

```go
func (r *mutationResolver) CreatePriceAlert(ctx context.Context, input model.CreatePriceAlertInput) (*model.PriceAlert, error) {
    // Validation
    if input.TargetPrice <= 0 {
        return nil, gqlerror.Errorf("targetPrice must be positive")
    }

    if input.AlertType == model.AlertTypePercentageDrop && input.DropPercent == nil {
        return nil, gqlerror.Errorf("dropPercent required for PERCENTAGE_DROP alerts")
    }

    // ... rest of logic
}
```

---

## 6. Performance Optimizations

### 6.1 DataLoader Implementation

Ensure DataLoaders are used for nested resolvers:

```go
// Batch load stores for products
func (r *productResolver) Store(ctx context.Context, obj *model.Product) (*model.Store, error) {
    return r.loaders.StoreLoader.Load(ctx, obj.StoreID)
}
```

**Required DataLoaders:**
- StoreLoader (by ID)
- FlyerLoader (by ID)
- ProductMasterLoader (by ID)
- UserLoader (by ID)

### 6.2 Query Complexity Limits

Add complexity analysis to prevent expensive queries:

```go
// gqlgen.yml
complexity:
  Query:
    products: 10
    searchProducts: 20
    flyers: 10
  Product:
    priceHistory: 5
    productMaster: 2
```

### 6.3 Response Caching

Add cache headers for public queries:

```go
func cacheMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        // Check if query is cacheable (GET or specific POST patterns)
        if isCacheableQuery(c) {
            c.Header("Cache-Control", "public, max-age=60")
        }
        c.Next()
    }
}
```

**Cacheable queries:**
- `stores` - 5 min
- `currentFlyers` - 5 min
- `productsOnSale` - 1 min
- `searchProducts` - 30 sec

---

## 7. CORS & Security

### 7.1 Rate Limiting

Add rate limiting for mutations:

```go
// Per-user limits
var rateLimits = map[string]rateLimit{
    "login":              {maxRequests: 5, window: time.Minute},
    "register":           {maxRequests: 3, window: time.Minute},
    "createShoppingList": {maxRequests: 10, window: time.Minute},
    "createPriceAlert":   {maxRequests: 20, window: time.Hour},
}
```

### 7.2 Query Depth Limiting

Prevent deeply nested queries:

```go
// Maximum depth of 10 levels
srv := handler.NewDefaultServer(generated.NewExecutableSchema(cfg))
srv.Use(extension.FixedComplexityLimit(100))
```

### 7.3 Auth Token Validation

Ensure all protected queries check auth:

```go
func (r *queryResolver) ShoppingLists(ctx context.Context, ...) (*model.ShoppingListConnection, error) {
    user := auth.UserFromContext(ctx)
    if user == nil {
        return nil, gqlerror.Errorf("authentication required")
    }

    // Only return user's own lists
    return r.listRepo.FindByUserID(ctx, user.ID, filters, pagination)
}
```

---

## 8. API Documentation

### 8.1 GraphQL Playground

Ensure playground is available in development:

```go
if config.Environment == "development" {
    router.GET("/graphql", playgroundHandler())
}
```

### 8.2 Schema Comments

Add descriptions to all types and fields:

```graphql
"""
Product represents a grocery item extracted from a flyer.
Products are linked to their source flyer and store.
"""
type Product {
  """Unique identifier for the product"""
  id: Int!

  """Display name of the product"""
  name: String!

  """Brand name if identified (e.g., "Dvaro", "Rokiškio")"""
  brand: String

  """Current pricing information"""
  price: ProductPrice!

  """Whether this product is currently on sale"""
  isOnSale: Boolean!
}
```

### 8.3 Error Codes

Standardize error codes:

```go
const (
    ErrCodeAuthentication = "AUTHENTICATION_REQUIRED"
    ErrCodeAuthorization  = "AUTHORIZATION_DENIED"
    ErrCodeNotFound       = "RESOURCE_NOT_FOUND"
    ErrCodeValidation     = "VALIDATION_ERROR"
    ErrCodeRateLimit      = "RATE_LIMIT_EXCEEDED"
    ErrCodeInternal       = "INTERNAL_ERROR"
)

func NewAuthError() *gqlerror.Error {
    return &gqlerror.Error{
        Message: "Authentication required",
        Extensions: map[string]interface{}{
            "code": ErrCodeAuthentication,
        },
    }
}
```

Frontend can then handle errors consistently:

```typescript
if (error.extensions?.code === 'AUTHENTICATION_REQUIRED') {
  router.push('/login');
}
```

---

## Summary: Priority Actions

### Immediate (Before Frontend MVP)

| Item | Priority | Effort | Owner |
|------|----------|--------|-------|
| CORS configuration | P0 | 2h | Backend |
| Image URL base path | P0 | 2h | Backend |
| Auth cookie option | P0 | 4h | Backend |

### Before Phase 2

| Item | Priority | Effort |
|------|----------|--------|
| DateTime standardization | P1 | 4h |
| `homeSummary` query | P1 | 4h |
| DataLoader audit | P1 | 4h |
| Rate limiting | P1 | 4h |

### Phase 3+

| Item | Priority | Effort |
|------|----------|--------|
| User preferences | P2 | 8h |
| Real-time subscriptions | P2 | 16h |
| Analytics events | P2 | 8h |

---

**End of Document**

Version: 1.0
Last Updated: 2025-12-10
