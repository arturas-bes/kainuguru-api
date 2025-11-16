# SearchService Research Documentation Index

Complete research on the kainuguru-api SearchService implementation. Three comprehensive documents covering all aspects of the search functionality.

## Documents

### 1. [SEARCH_SERVICE_RESEARCH.md](SEARCH_SERVICE_RESEARCH.md) - Comprehensive Research Report
**Length:** 17 KB | **Sections:** 10 major sections

The complete deep-dive into SearchService implementation covering:

1. **FuzzySearchProducts Method Signature & API Contract**
   - Full method signature
   - SearchRequest structure with validation rules
   - SearchResponse structure
   - ProductSearchResult structure

2. **Similarity Scores & Usage**
   - Three-score model (name, brand, combined)
   - Hybrid search scores (FTS + fuzzy)
   - Score range and weighting
   - Application usage patterns

3. **Brand Filtering & Brand-Aware Searching**
   - Current brand support in database
   - Combined name+brand matching
   - Brand faceting implementation
   - Brand filtering limitation (not yet implemented)
   - Recommendations for enhancement

4. **ProductMaster Integration**
   - Current integration status
   - Data model relationships
   - Service-level integration
   - ProductMaster purpose
   - Recommendations for improvement

5. **GraphQL API Integration**
   - Query entry point
   - SearchInput GraphQL type
   - SearchResult GraphQL type
   - Resolver implementation
   - Facets computation (5 dimensions)

6. **SQL Functions (Database Level)**
   - fuzzy_search_products() function
   - hybrid_search_products() function
   - Parameters and return types
   - Algorithm details

7. **Service Configuration & Defaults**
   - Hardcoded values and tuning parameters
   - Service initialization
   - Configuration table

8. **Key Limitations & Gaps**
   - Brand filtering missing
   - ProductMaster not loaded
   - Highlights not populated
   - Match type hardcoded
   - Limited facet active values

9. **Performance Characteristics**
   - Database-level optimizations
   - Index utilization
   - Query performance metrics

10. **Testing Surface**
    - GraphQL integration test examples
    - Test coverage areas

**Best for:** Understanding the complete architecture, SQL implementations, and finding detailed specifications

---

### 2. [SEARCH_SERVICE_ARCHITECTURE.md](SEARCH_SERVICE_ARCHITECTURE.md) - Visual Architecture & Data Flow
**Length:** 20 KB | **Sections:** 10 visual diagrams & flows

Visual representations and step-by-step data flow through the system:

1. **System Architecture Overview**
   - GraphQL → Resolver → Service → Database layer diagram
   - Component relationships

2. **FuzzySearchProducts Data Flow**
   - Complete request-to-response flow with example data
   - Validation, sanitization, database processing
   - Score calculation with real numbers
   - Response building
   - Facet computation
   - Final GraphQL response example

3. **Score Calculation Deep Dive**
   - FuzzySearch score composition (3-score model)
   - Score weighting: 70% name, 30% brand, 20% normalized bonus
   - HybridSearch score composition (FTS vs fuzzy)
   - PreferFuzzy boost mechanism
   - Ranking algorithm

4. **ProductMaster Relationship Diagram**
   - Current state (not integrated)
   - Desired state (if implemented)
   - Benefits of integration

5. **Filtering Architecture**
   - Complete filter processing flow
   - All 7 filter types with database application
   - Always-applied filters

6. **Key Data Structures**
   - GraphQL SearchInput → Service SearchRequest transformation
   - Database result → Service ProductSearchResult
   - Final GraphQL SearchResult

7. **Performance Characteristics**
   - Query speed breakdown
   - Database operations timing
   - Facet computation timing
   - Application processing timing
   - Total request time estimation
   - Optimizations applied

8. **Testing Coverage**
   - Unit tests
   - Integration tests
   - BDD tests with example scenarios

**Best for:** Understanding data flow, visualizing the search pipeline, performance analysis, learning how scores work with real examples

---

### 3. [SEARCH_SERVICE_QUICK_REFERENCE.md](SEARCH_SERVICE_QUICK_REFERENCE.md) - Quick Reference & Code Examples
**Length:** 12 KB | **Sections:** 18 quick-reference sections

Developer-friendly quick lookup guide with code examples:

1. **Method Signatures** - All search service methods
2. **SearchRequest Structure** - Field reference table
3. **Similarity Scores Explained** - Quick explanation of scoring
4. **GraphQL Query Examples** - Copy-paste ready queries
5. **Response Structure** - Go and GraphQL types
6. **Performance Tips** - Optimization guidance
7. **Key Features** - Supported and unsupported features checklist
8. **Database Functions** - SQL function parameters and examples
9. **Scoring Formulas** - Mathematical formulas with examples
10. **Common Patterns** - Code snippets for typical use cases
11. **Pagination** - How to implement pagination
12. **Brand-Aware Search** - Current limitations and workarounds
13. **Error Handling** - Common errors and solutions
14. **Testing Examples** - Unit and integration test code
15. **File Locations** - Where to find relevant code
16. **Related Services** - Dependencies and related components
17. **Future Enhancements** - Roadmap items

**Best for:** Developer reference, quick lookups, copy-paste code examples, testing

---

## Key Findings Summary

### FuzzySearchProducts Method
- **Signature:** `FuzzySearchProducts(ctx context.Context, req *SearchRequest) (*SearchResponse, error)`
- **Returns:** Three similarity scores (name, brand, combined) + full product objects
- **Algorithm:** PostgreSQL trigram similarity matching with Lithuanian normalization

### Similarity Scores
- **Three scores returned:**
  1. `name_similarity`: 0.0-1.0 (product name vs query)
  2. `brand_similarity`: 0.0-1.0 (brand vs query)
  3. `combined_similarity`: 0.3-1.0+ (weighted combination, used for ranking)
- **Weighting:** 70% name + 30% brand + 20% normalized bonus
- **Threshold:** 0.3 minimum (30% match required)
- **Score usage:** Results sorted by combined_similarity DESC

### Brand Awareness
- **Current Status:**
  - Brand IS included in similarity calculations (30% weight)
  - Brand faceting WORKS (shows top 20 brands)
  - Brand filtering NOT IMPLEMENTED (can't filter by brand yet)
- **Comment in code:** "Brand filtering not implemented in SearchRequest yet"
- **Recommendation:** Add brand filter to SearchInput and SearchRequest

### ProductMaster Integration
- **Current Status:** NOT INTEGRATED
- **Available:** Via separate `productMaster` query only
- **Missing:** Not loaded in search results relations
- **Gap:** Would require adding `.Relation("ProductMaster")` to load
- **Benefit if added:** Deduplication, canonical data, shopping list linking

### GraphQL API Surface
- **Query:** `searchProducts(input: SearchInput!): SearchResult!`
- **Filters:** q, storeIDs, minPrice, maxPrice, onSaleOnly, category, tags, first
- **Not supported:** cursor (after), brand filter
- **Response includes:** 5 facet dimensions + pagination + suggestions

### Database Functions
- **fuzzy_search_products()** - Pure fuzzy matching with 3 similarity scores
- **hybrid_search_products()** - FTS + fuzzy with match type indication
- **Both:** Support price, store, category, tag, sale filters + pagination

---

## How to Use These Documents

### For Understanding Architecture
1. Start with **ARCHITECTURE.md** - Read system overview and data flow
2. Review score calculation diagrams
3. Understand performance characteristics

### For Implementation
1. Check **QUICK_REFERENCE.md** - Method signatures and examples
2. Use code snippets for patterns
3. Reference **RESEARCH.md** - For detailed specifications

### For Development
1. Use **QUICK_REFERENCE.md** - For daily reference
2. Consult **RESEARCH.md** - For deep details
3. Review **ARCHITECTURE.md** - For debugging data flow

### For Enhancement Planning
1. Read "Key Limitations & Gaps" in **RESEARCH.md**
2. Review "Future Enhancements" in **QUICK_REFERENCE.md**
3. Understand ProductMaster integration in **RESEARCH.md** Section 4

---

## Quick Facts

| Aspect | Detail |
|--------|--------|
| Service File | `internal/services/search/service.go` |
| GraphQL Resolver | `internal/graphql/resolvers/query.go:265-376` |
| DB Functions | `migrations/029_add_tags_to_search_functions.sql` |
| Tests | `tests/bdd/steps/search_test.go` |
| Similarity Threshold | 0.3 (30% minimum match) |
| Max Stores Filterable | 50 |
| Max Query Length | 255 characters |
| Default Limit | 50 results |
| Top Brands Shown | 20 (facet limit) |
| Typical Query Time | 150-550ms |
| Name Weight in Score | 70% |
| Brand Weight in Score | 30% |
| Normalized Bonus | 20% (Lithuanian diacritics) |
| FTS Boost | 2.0x |
| PreferFuzzy Boost | 1.2x |
| Brand Filtering | NOT IMPLEMENTED |
| ProductMaster Loaded | NO |
| Highlights Populated | NO |

---

## Document Statistics

- **Total Documentation:** 49 KB
- **Code Examples:** 20+
- **SQL Functions Documented:** 2
- **Diagrams:** 8 ASCII architecture diagrams
- **Data Flows:** 2 complete flows with examples
- **GraphQL Examples:** 3 queries
- **Test Examples:** 2 complete test functions
- **Performance Metrics:** 10+ timing estimates
- **Limitations Identified:** 6 major gaps
- **File References:** 8 key source files

---

## File Locations (Absolute Paths)

- `/Users/arturas/Dev/kainuguru_all/kainuguru-api/docs/SEARCH_SERVICE_RESEARCH.md`
- `/Users/arturas/Dev/kainuguru_all/kainuguru-api/docs/SEARCH_SERVICE_ARCHITECTURE.md`
- `/Users/arturas/Dev/kainuguru_all/kainuguru-api/docs/SEARCH_SERVICE_QUICK_REFERENCE.md`
- `/Users/arturas/Dev/kainuguru_all/kainuguru-api/docs/SEARCH_SERVICE_INDEX.md` (this file)

---

**Research Date:** November 15, 2025  
**Codebase:** kainuguru-api  
**Branch:** 001-system-validation  
**Status:** Complete Research

