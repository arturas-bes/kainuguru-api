# Insomnia Collection Changelog

**Updated:** 2025-11-06 20:00:00 UTC
**Collection Version:** 2.2.0
**Status:** ✅ All endpoints validated against live API

## Summary of Changes

This update brings the Insomnia collection in line with the validated GraphQL API, correcting field names discovered during deep endpoint validation and adding new endpoints.

---

## Major Changes

### 1. ✅ Schema Field Corrections

Fixed incorrect field names that were causing validation errors:

- **Store Type**
  - ❌ Removed: `description` field (doesn't exist in schema)
  - ✅ Validated: All other store fields working

- **Product Type**
  - ❌ Fixed: `thumbnailURL` → `imageURL` (correct field name)
  - ✅ All other product fields validated

- **ProductMaster Type**
  - ❌ Fixed: `name` → `canonicalName` (correct field name)
  - ✅ All matching and statistics fields validated

- **FacetOption Type**
  - ❌ Fixed: `label` → `name` (correct field name in options)
  - ✅ Facet structure validated

### 2. 🆕 New Endpoints Added

#### Price History (Complete Suite)
- **Get Price History** - Full price tracking with pagination
- **Get Current Price** - Latest price for product+store combination
- **Price History with Date Range** - Historical data within date bounds

#### Advanced Search
- **Simple Search** - Basic product search with Lithuanian support
- **Advanced Search with Filters** - Full-featured search with facets
- **Fuzzy Search** - Typo-tolerant search

#### Shopping Lists (Complete CRUD)
- **Get All Shopping Lists** - List with computed fields
- **Get Shopping List by ID** - Individual list with items
- **Get Default Shopping List** - User's default list
- **Create Shopping List** - Create new list
- **Update Shopping List** - Modify existing list
- **Delete Shopping List** - Remove list
- **Set Default Shopping List** - Change default

#### System Endpoints
- **Health Check** - API health validation
- **GraphQL Playground** - Interactive testing

### 3. 📝 Enhanced Descriptions

All endpoints now include:
- Validation status notes (e.g., "Validated: Returns records ordered by recordedAt DESC")
- Schema corrections (e.g., "Validated: use 'imageURL' not 'thumbnailURL'")
- Business logic notes (e.g., "Validated: First list auto-becomes default")
- Performance characteristics
- Authentication requirements

### 4. 🔧 Improved Organization

Reorganized folders for better workflow:
- ⚙️ System - Health checks and utilities
- 🔐 Authentication - User auth flows
- 🏪 Stores & Flyers - Store information
- 🛍️ Products - Product browsing
- 🔍 Search - Advanced search features
- 💰 Price History - Price tracking
- 🛒 Shopping Lists - List management
- 🎯 Product Masters - Canonical products

### 5. 🔑 Environment Updates

Added `refresh_token` variable to environments:
```json
{
  "base_url": "http://localhost:8080",
  "auth_token": "",
  "refresh_token": ""  // NEW
}
```

---

## Complete Endpoint List

### System (2 endpoints)
1. Health Check (GET)
2. GraphQL Playground (GET)

### Authentication (3 endpoints)
1. Register User
2. Login User
3. Get Current User (Me)

### Stores & Flyers (5 endpoints)
1. Get All Stores
2. Get Store by ID
3. Get Store by Code
4. Get Current Flyers
5. Get Flyer Details with Products

### Products (4 endpoints)
1. Get Products with Pagination
2. Get Product by ID (Full Details)
3. Products on Sale
4. (Search moved to Search folder)

### Search (3 endpoints)
1. Simple Search
2. Advanced Search with Filters
3. Fuzzy Search

### Price History (3 endpoints)
1. Get Price History
2. Get Current Price
3. Price History with Date Range

### Shopping Lists (7 endpoints)
1. Get All Shopping Lists
2. Get Shopping List by ID
3. Get Default Shopping List
4. Create Shopping List
5. Update Shopping List
6. Delete Shopping List
7. Set Default Shopping List

### Product Masters (2 endpoints)
1. Get Product Master by ID
2. Get Product Masters with Filters

**Total: 29 validated endpoints**

---

## Field Name Corrections Reference

For developers migrating from old queries:

| Type | Old Field | New Field | Status |
|------|-----------|-----------|--------|
| Store | `description` | N/A | ❌ Field doesn't exist |
| Product | `thumbnailURL` | `imageURL` | ✅ Corrected |
| Product | `discountAmount` | `price.discountAmount` | ✅ Moved to nested field |
| ProductMaster | `name` | `canonicalName` | ✅ Corrected |
| ProductMaster | `brand!` | `brand` (nullable) | ✅ Fixed schema |
| ProductMaster | `category!` | `category` (nullable) | ✅ Fixed schema |
| ProductMasterFilters | `isVerified` | N/A | ❌ Column doesn't exist in DB |
| FacetOption | `label` | `name` | ✅ Corrected |

---

## Validation Notes

### Authentication
- ✅ emailVerified returns `true` immediately (dev mode behavior)
- ✅ Token expiration: 24h access, 7d refresh
- ✅ JWT properly signed with HS256

### Shopping Lists
- ✅ First list auto-becomes default (intentional UX)
- ✅ Computed fields working: `completionPercentage`, `isCompleted`, `canBeShared`
- ✅ Access control validated (can't access other user's lists)

### Price History
- ✅ Records ordered by `recordedAt` DESC (most recent first)
- ✅ Date range filtering working
- ✅ Store filtering working
- ✅ Sale filtering working

### Search
- ✅ Lithuanian full-text search working ("pienas", "duona")
- ✅ Fuzzy matching for typo tolerance
- ✅ Price range filtering validated
- ✅ onSaleOnly filtering validated
- ✅ Empty queries correctly rejected

### Products
- ✅ Price object calculations correct
- ✅ Computed fields: `isCurrentlyOnSale`, `isValid`, `isExpired`
- ✅ Relations: Store, Flyer, ProductMaster all resolve correctly

---

## Example Workflows

### 1. User Registration & Authentication
```
1. Register User
2. Copy accessToken to environment variable
3. Get Current User (Me)
```

### 2. Browse Products
```
1. Get All Stores
2. Get Current Flyers (for selected store)
3. Get Products with Pagination (with filters)
4. Get Product by ID (for details)
```

### 3. Search Products
```
1. Simple Search (basic query)
2. Advanced Search with Filters (refine results)
3. Get Product by ID (view details)
```

### 4. Track Prices
```
1. Get Product Master by ID
2. Get Price History (for that product master)
3. Get Current Price (latest price)
```

### 5. Manage Shopping List
```
1. Login User (get auth token)
2. Get All Shopping Lists
3. Create Shopping List (if none exist)
4. Get Shopping List by ID (with items)
5. Update Shopping List (modify name/description)
```

---

## Testing Notes

### Prerequisites
1. API running on http://localhost:8080
2. Database seeded with test data (`make seed-data`)
3. Redis running

### Testing Authentication
1. Use unique emails for each test (timestamp-based recommended)
2. Store `accessToken` in environment variable for subsequent requests
3. Token valid for 24 hours

### Testing Protected Endpoints
1. Login first to get token
2. Copy token to `auth_token` environment variable
3. All shopping list endpoints require authentication

### Expected Response Times
- Authentication: 200-350ms (bcrypt hashing)
- Simple queries: 1-10ms
- Complex queries: 15-50ms
- Search (first run): 70-125ms
- Search (cached): 1-5ms

---

## Import Instructions

### Insomnia Desktop App
1. Open Insomnia
2. Click "Import" button
3. Select `docs/kainuguru-graphql-insomnia-collection.json`
4. Collection will be imported with all folders and requests

### Environment Setup
1. Select "Development" environment
2. Set `auth_token` after logging in
3. Set `refresh_token` after logging in (for future use)

### Production Environment
1. Switch to "Production" environment
2. Update `base_url` to your production URL
3. Set production tokens

---

## Maintenance

This collection is synchronized with:
- GraphQL Schema: `/internal/graphql/schema/schema.graphql`
- Validation Report: `/tests/ENDPOINT_VALIDATION_REPORT.md`
- Test Suite: `/tests/bdd/steps/*_test.go`

When schema changes:
1. Update collection queries
2. Run validation tests
3. Update this changelog
4. Bump collection version

---

## Support

For issues or questions:
- Schema validation: See `/tests/ENDPOINT_VALIDATION_REPORT.md`
- Test failures: Check `/tests/bdd/steps/` for test code
- API errors: Check logs with `docker-compose logs api`
- Rate limiting: Wait 60 seconds between high-volume tests

---

## Version History

### v2.2.0 (2025-11-06)
- ✅ Fixed ProductMaster filters - removed `isVerified` (column doesn't exist in DB)
- ✅ Updated ProductMaster schema - made `brand` and `category` nullable
- ✅ Replaced `isVerified: true` filter with `minConfidence: 0.5` in queries
- ✅ Fixed ProductMaster endpoints that were failing with database column errors
- ✅ All ProductMaster queries now work correctly

### v2.1.0 (2025-11-06)
- ✅ Removed duplicate `discountAmount` field from Product type
- ✅ Updated "Get Product by ID" query to use `price.discountAmount` only
- ✅ Synchronized with schema optimization changes (Phase 4)

### v2.0.0 (2025-11-06)
- Complete schema validation and field corrections
- Added 12 new endpoints (Price History, Advanced Search, Shopping Lists)
- Reorganized folder structure
- Added validation notes to all descriptions
- Fixed all known field name issues

### v1.0.0 (Previous)
- Initial collection
- Basic CRUD operations
- Authentication flows
- Store and product browsing
