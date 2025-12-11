-- +goose Up
-- +goose StatementBegin

-- ============================================================================
-- Shopping List Migration Wizard Tables
-- Feature: 001-shopping-list-migration
-- Created: 2025-11-16
-- ============================================================================

-- Add origin enum to shopping_list_items to distinguish flyer vs free-text items
-- Backward compatible: defaults to 'free_text' for existing rows
ALTER TABLE shopping_list_items 
ADD COLUMN IF NOT EXISTS origin VARCHAR(20) DEFAULT 'free_text' 
CHECK (origin IN ('flyer', 'free_text'));

-- Backfill origin='flyer' for items with linked_product_id
UPDATE shopping_list_items 
SET origin = 'flyer' 
WHERE linked_product_id IS NOT NULL AND origin = 'free_text';

-- Add is_locked field for list locking during wizard sessions (FR-016)
-- MOVED TO 035_add_is_locked_to_shopping_lists.sql
-- ALTER TABLE shopping_lists
-- ADD COLUMN IF NOT EXISTS is_locked BOOLEAN DEFAULT FALSE;

-- Index for checking locked lists
-- MOVED TO 035_add_is_locked_to_shopping_lists.sql
-- CREATE INDEX IF NOT EXISTS idx_shopping_lists_locked 
-- ON shopping_lists(id, is_locked) WHERE is_locked = TRUE;

-- ============================================================================
-- Offer Snapshots Table - Immutable historical records of wizard suggestions
-- ============================================================================
CREATE TABLE IF NOT EXISTS offer_snapshots (
    id BIGSERIAL PRIMARY KEY,
    
    -- Shopping list item context
    shopping_list_item_id BIGINT NOT NULL REFERENCES shopping_list_items(id) ON DELETE CASCADE,
    
    -- Product references
    flyer_product_id BIGINT, -- No FK due to partitioned products table
    product_master_id BIGINT REFERENCES product_masters(id) ON DELETE SET NULL,
    store_id INTEGER REFERENCES stores(id) ON DELETE SET NULL,
    
    -- Product snapshot at time of suggestion
    product_name VARCHAR(255) NOT NULL,
    brand VARCHAR(100),
    price DECIMAL(10,2) NOT NULL CHECK (price >= 0),
    unit VARCHAR(50),
    size_value DECIMAL(10,3),
    size_unit VARCHAR(20),
    
    -- Validity period from flyer
    valid_from TIMESTAMP WITH TIME ZONE,
    valid_to TIMESTAMP WITH TIME ZONE,
    
    -- Metadata
    estimated BOOLEAN DEFAULT FALSE NOT NULL, -- FALSE for wizard (always uses actual flyer prices)
    snapshot_reason VARCHAR(50) NOT NULL DEFAULT 'wizard_migration',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW() NOT NULL,
    
    CONSTRAINT chk_snapshot_reason CHECK (
        snapshot_reason IN ('wizard_migration', 'price_history', 'manual_snapshot')
    )
);

-- Indexes for offer_snapshots
CREATE INDEX IF NOT EXISTS idx_offer_snapshots_item 
ON offer_snapshots(shopping_list_item_id);

CREATE INDEX IF NOT EXISTS idx_offer_snapshots_flyer_product 
ON offer_snapshots(flyer_product_id) WHERE flyer_product_id IS NOT NULL;

CREATE INDEX IF NOT EXISTS idx_offer_snapshots_created 
ON offer_snapshots(created_at DESC);

CREATE INDEX IF NOT EXISTS idx_offer_snapshots_store 
ON offer_snapshots(store_id) WHERE store_id IS NOT NULL;

-- ============================================================================
-- Comments for documentation
-- ============================================================================
COMMENT ON TABLE offer_snapshots IS 'Immutable historical records of product offers at time of wizard suggestion or price tracking';
COMMENT ON COLUMN offer_snapshots.estimated IS 'FALSE for wizard suggestions (always uses actual flyer prices per constitution), TRUE for category-level discounts';
COMMENT ON COLUMN offer_snapshots.snapshot_reason IS 'Source context: wizard_migration (from wizard), price_history (tracking), manual_snapshot (user-triggered)';
COMMENT ON COLUMN shopping_list_items.origin IS 'Item source: flyer (from flyer offer) or free_text (manually entered by user)';
-- COMMENT ON COLUMN shopping_lists.is_locked IS 'List lock status during active wizard session (FR-016 compliance)';

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

-- Remove wizard tables in reverse order
DROP TABLE IF EXISTS offer_snapshots CASCADE;

-- Remove added columns
ALTER TABLE shopping_lists DROP COLUMN IF EXISTS is_locked;
ALTER TABLE shopping_list_items DROP COLUMN IF EXISTS origin;

-- +goose StatementEnd
