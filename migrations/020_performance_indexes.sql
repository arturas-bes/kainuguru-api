-- Migration: 020_performance_indexes.sql
-- Purpose: Add indexes to optimize DataLoader batch queries and improve GraphQL query performance
-- Related: DataLoader implementation and N+1 query optimization
-- Date: 2025-11-06

-- =============================================================================
-- PRODUCTS TABLE INDEXES
-- =============================================================================

-- Composite index for product queries filtered by store and validity
-- Used by: Product queries with store filter, DataLoader batch queries
-- Performance impact: Critical for nested product queries
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_products_store_valid
ON products(store_id, valid_from, valid_to)
WHERE is_available = true;

-- Composite index for products by flyer and sale status
-- Used by: Flyer products queries, on-sale product filters
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_products_flyer_sale
ON products(flyer_id, is_on_sale)
WHERE is_available = true;

-- Index for product master lookups (DataLoader)
-- Used by: Product.productMaster resolver batching
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_products_product_master
ON products(product_master_id)
WHERE product_master_id IS NOT NULL;

-- Index for flyer page lookups (DataLoader)
-- Used by: Product.flyerPage resolver batching
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_products_flyer_page
ON products(flyer_page_id)
WHERE flyer_page_id IS NOT NULL;

-- Covering index for product search with common fields
-- Used by: Product list queries that need basic info without joining
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_products_search_covering
ON products(id, store_id, flyer_id, name, normalized_name, current_price, is_on_sale)
WHERE is_available = true;

-- =============================================================================
-- FLYERS TABLE INDEXES
-- =============================================================================

-- Composite index for current/valid flyer queries
-- Used by: currentFlyers, validFlyers queries
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_flyers_store_valid
ON flyers(store_id, valid_from, valid_to)
WHERE is_archived = false;

-- Index for flyers by status and store (DataLoader support)
-- Used by: Store.flyers resolver with status filter
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_flyers_store_status
ON flyers(store_id, status)
WHERE is_archived = false;

-- Composite index for date-based flyer queries
-- Used by: Flyer queries with date range filters
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_flyers_dates
ON flyers(valid_from, valid_to)
WHERE is_archived = false;

-- =============================================================================
-- FLYER PAGES TABLE INDEXES
-- =============================================================================

-- Index for flyer page lookups by flyer (DataLoader)
-- Used by: Flyer.flyerPages resolver batching
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_flyer_pages_flyer
ON flyer_pages(flyer_id, page_number);

-- Index for flyer pages by status (processing queries)
-- Used by: Processing job queries
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_flyer_pages_status
ON flyer_pages(status)
WHERE status IN ('PENDING', 'PROCESSING');

-- =============================================================================
-- PRICE HISTORY TABLE INDEXES
-- =============================================================================

-- Composite index for price history lookups (DataLoader + queries)
-- Used by: PriceHistory queries by product master and store
-- Performance impact: Critical for price trend queries
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_price_history_lookup
ON price_history(product_master_id, store_id, recorded_at DESC)
WHERE is_active = true;

-- Index for current price queries
-- Used by: Getting latest price for a product
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_price_history_current
ON price_history(product_master_id, store_id, valid_to DESC)
WHERE is_active = true AND is_available = true;

-- Index for sale products in price history
-- Used by: Finding on-sale products across price history
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_price_history_sales
ON price_history(store_id, is_on_sale, valid_from, valid_to)
WHERE is_active = true AND is_on_sale = true;

-- Index for price history by flyer
-- Used by: Linking price history to specific flyers
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_price_history_flyer
ON price_history(flyer_id, recorded_at DESC)
WHERE flyer_id IS NOT NULL AND is_active = true;

-- =============================================================================
-- SHOPPING LISTS TABLE INDEXES
-- =============================================================================

-- Index for user's shopping lists (DataLoader)
-- Used by: User.shoppingLists resolver
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_shopping_lists_user
ON shopping_lists(user_id, created_at DESC)
WHERE is_archived = false;

-- Index for default shopping list lookup
-- Used by: myDefaultShoppingList query
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_shopping_lists_default
ON shopping_lists(user_id, is_default)
WHERE is_default = true AND is_archived = false;

-- Index for shared shopping lists
-- Used by: sharedShoppingList query by share code
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_shopping_lists_share
ON shopping_lists(share_code)
WHERE share_code IS NOT NULL AND is_public = true;

-- =============================================================================
-- SHOPPING LIST ITEMS TABLE INDEXES
-- =============================================================================

-- Composite index for shopping list items (primary use case)
-- Used by: ShoppingList.items resolver with sorting
-- Performance impact: Critical for shopping list queries
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_shopping_list_items_list_order
ON shopping_list_items(shopping_list_id, is_checked, sort_order);

-- Index for user's items across all lists
-- Used by: User item queries, item suggestions
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_shopping_list_items_user
ON shopping_list_items(user_id, created_at DESC);

-- Index for product master linkage (DataLoader)
-- Used by: ShoppingListItem.productMaster resolver batching
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_shopping_list_items_product_master
ON shopping_list_items(product_master_id)
WHERE product_master_id IS NOT NULL;

-- Index for linked products (DataLoader)
-- Used by: ShoppingListItem.linkedProduct resolver batching
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_shopping_list_items_linked_product
ON shopping_list_items(linked_product_id)
WHERE linked_product_id IS NOT NULL;

-- Index for store linkage (DataLoader)
-- Used by: ShoppingListItem.store resolver batching
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_shopping_list_items_store
ON shopping_list_items(store_id)
WHERE store_id IS NOT NULL;

-- Index for category-based queries
-- Used by: Item filtering by category
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_shopping_list_items_category
ON shopping_list_items(shopping_list_id, category)
WHERE category IS NOT NULL;

-- =============================================================================
-- PRODUCT MASTERS TABLE INDEXES
-- =============================================================================

-- Index for product master by normalized name (search)
-- Used by: Product matching, search suggestions
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_product_masters_normalized
ON product_masters(normalized_name)
WHERE status = 'ACTIVE';

-- Index for product masters by brand and category
-- Used by: Product master queries with filters
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_product_masters_brand_category
ON product_masters(brand, category)
WHERE status = 'ACTIVE';

-- Index for verified product masters
-- Used by: Queries for verified products only
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_product_masters_verified
ON product_masters(status, is_verified)
WHERE status = 'ACTIVE' AND is_verified = true;

-- =============================================================================
-- USERS TABLE INDEXES
-- =============================================================================

-- Index for email lookups (authentication)
-- Used by: Login, registration, email verification
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_users_email
ON users(email)
WHERE is_active = true;

-- Index for user sessions lookup
-- Used by: Session validation, user authentication
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_sessions_user
ON user_sessions(user_id, expires_at DESC)
WHERE is_valid = true;

-- Index for session token hash lookup
-- Used by: Token validation
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_sessions_token
ON user_sessions(access_token_hash)
WHERE is_valid = true;

-- =============================================================================
-- LOGIN ATTEMPTS TABLE INDEXES (Security)
-- =============================================================================

-- Index for rate limiting checks
-- Used by: Login attempt tracking, rate limiting
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_login_attempts_email_time
ON login_attempts(email, attempted_at DESC);

-- =============================================================================
-- EXTRACTION JOBS TABLE INDEXES (Background Processing)
-- =============================================================================

-- Index for job queue queries
-- Used by: Job processing, worker assignment
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_extraction_jobs_queue
ON extraction_jobs(status, priority DESC, scheduled_at ASC)
WHERE status IN ('PENDING', 'PROCESSING');

-- Index for worker assignment
-- Used by: Finding jobs assigned to specific workers
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_extraction_jobs_worker
ON extraction_jobs(worker_id, status)
WHERE worker_id IS NOT NULL;

-- =============================================================================
-- PERFORMANCE NOTES
-- =============================================================================

-- CONCURRENTLY: Indexes are created without locking the table
-- This allows the migration to run in production without downtime

-- WHERE clauses: Partial indexes reduce index size and improve performance
-- Only relevant rows are indexed (e.g., is_active = true, is_archived = false)

-- Index order: Column order matters for composite indexes
-- Most selective column first, then sort columns

-- Covering indexes: Include frequently accessed columns to avoid table lookups
-- Example: idx_products_search_covering includes common fields

-- =============================================================================
-- MONITORING
-- =============================================================================

-- Check index usage after deployment:
-- SELECT schemaname, tablename, indexname, idx_scan, idx_tup_read, idx_tup_fetch
-- FROM pg_stat_user_indexes
-- WHERE schemaname = 'public'
-- ORDER BY idx_scan DESC;

-- Check index size:
-- SELECT schemaname, tablename, indexname, pg_size_pretty(pg_relation_size(indexrelid))
-- FROM pg_stat_user_indexes
-- WHERE schemaname = 'public'
-- ORDER BY pg_relation_size(indexrelid) DESC;

-- =============================================================================
-- ROLLBACK
-- =============================================================================

-- To remove these indexes if needed:
-- DROP INDEX CONCURRENTLY IF EXISTS idx_products_store_valid;
-- DROP INDEX CONCURRENTLY IF EXISTS idx_products_flyer_sale;
-- ... (repeat for all indexes)
