-- +goose Up
-- +goose StatementBegin
-- Migration: 025_performance_indexes.sql
-- Purpose: Add indexes to optimize DataLoader batch queries and improve GraphQL query performance
-- Related: DataLoader implementation and N+1 query optimization
-- Date: 2025-11-06
-- Note: CONCURRENTLY removed as it cannot be used in transactions

-- =============================================================================
-- PRODUCTS TABLE INDEXES
-- =============================================================================

-- Composite index for product queries filtered by store and validity
CREATE INDEX IF NOT EXISTS idx_products_store_valid
ON products(store_id, valid_from, valid_to)
WHERE is_available = true;

-- Composite index for products on sale
CREATE INDEX IF NOT EXISTS idx_products_flyer_sale
ON products(flyer_id, is_on_sale, valid_from, valid_to)
WHERE is_available = true;

-- Index for product master relationships
CREATE INDEX IF NOT EXISTS idx_products_product_master
ON products(product_master_id)
WHERE product_master_id IS NOT NULL;

-- Index for price range queries
CREATE INDEX IF NOT EXISTS idx_products_price_range
ON products(current_price)
WHERE is_available = true AND current_price IS NOT NULL;

-- =============================================================================
-- FLYERS TABLE INDEXES
-- =============================================================================

-- Composite index for active flyers by store
CREATE INDEX IF NOT EXISTS idx_flyers_store_active
ON flyers(store_id, valid_from, valid_to)
WHERE is_archived = false;

-- Index for flyer validity queries
CREATE INDEX IF NOT EXISTS idx_flyers_validity_dates
ON flyers(valid_from, valid_to)
WHERE is_archived = false;

-- =============================================================================
-- PRICE HISTORY INDEXES
-- =============================================================================

-- Composite index for price history queries by product master and time
CREATE INDEX IF NOT EXISTS idx_price_history_product_time
ON price_history(product_master_id, recorded_at DESC);

-- Composite index for price history by store and time
CREATE INDEX IF NOT EXISTS idx_price_history_store_time
ON price_history(store_id, product_master_id, recorded_at DESC);

-- Index for querying sale prices
CREATE INDEX IF NOT EXISTS idx_price_history_sales
ON price_history(product_master_id, recorded_at DESC)
WHERE is_on_sale = true;

-- =============================================================================
-- SHOPPING LIST INDEXES
-- =============================================================================

-- Index for user's shopping lists
CREATE INDEX IF NOT EXISTS idx_shopping_lists_user
ON shopping_lists(user_id, created_at DESC);

-- Index for default shopping lists
CREATE INDEX IF NOT EXISTS idx_shopping_lists_user_default
ON shopping_lists(user_id)
WHERE is_default = true;

-- Composite index for shopping list items
CREATE INDEX IF NOT EXISTS idx_shopping_list_items_list
ON shopping_list_items(shopping_list_id, is_checked, sort_order);

-- Index for unchecked items
CREATE INDEX IF NOT EXISTS idx_shopping_list_items_unchecked
ON shopping_list_items(shopping_list_id, sort_order)
WHERE is_checked = false;

-- =============================================================================
-- ANALYZE TABLES
-- =============================================================================
ANALYZE products;
ANALYZE flyers;
ANALYZE price_history;
ANALYZE shopping_lists;
ANALYZE shopping_list_items;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS idx_products_store_valid;
DROP INDEX IF EXISTS idx_products_flyer_sale;
DROP INDEX IF EXISTS idx_products_product_master;
DROP INDEX IF EXISTS idx_products_price_range;
DROP INDEX IF EXISTS idx_flyers_store_active;
DROP INDEX IF EXISTS idx_flyers_validity_dates;
DROP INDEX IF EXISTS idx_price_history_product_time;
DROP INDEX IF EXISTS idx_price_history_store_time;
DROP INDEX IF EXISTS idx_price_history_sales;
DROP INDEX IF EXISTS idx_shopping_lists_user;
DROP INDEX IF EXISTS idx_shopping_lists_user_default;
DROP INDEX IF EXISTS idx_shopping_list_items_list;
DROP INDEX IF EXISTS idx_shopping_list_items_unchecked;
-- +goose StatementEnd
