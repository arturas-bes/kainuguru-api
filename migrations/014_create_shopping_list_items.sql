-- +goose Up
-- +goose StatementBegin
CREATE TABLE shopping_list_items (
    id BIGSERIAL PRIMARY KEY,
    shopping_list_id BIGINT NOT NULL REFERENCES shopping_lists(id) ON DELETE CASCADE,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE, -- For attribution

    -- Item details
    description VARCHAR(255) NOT NULL,
    normalized_description VARCHAR(255) NOT NULL, -- Lithuanian chars normalized
    notes TEXT,

    -- Quantity and units
    quantity DECIMAL(10,3) DEFAULT 1 CHECK (quantity > 0 AND quantity <= 999),
    unit VARCHAR(50), -- '1L', '500g', 'vnt', etc
    unit_type VARCHAR(20), -- 'volume', 'weight', 'count'

    -- State
    is_checked BOOLEAN DEFAULT FALSE,
    checked_at TIMESTAMP WITH TIME ZONE,
    checked_by_user_id BIGINT REFERENCES users(id),

    -- Ordering
    sort_order INTEGER DEFAULT 0,

    -- Product linking
    product_master_id BIGINT, -- Link to master product catalog
    linked_product_id BIGINT, -- Link to specific weekly product
    store_id INTEGER, -- Preferred/suggested store
    flyer_id INTEGER, -- If added from flyer

    -- Price tracking
    estimated_price DECIMAL(10,2),
    actual_price DECIMAL(10,2), -- User can update after shopping
    price_source VARCHAR(50), -- 'flyer', 'user_estimate', 'historical'

    -- Categorization
    category VARCHAR(100),
    tags TEXT[], -- Array of tags for organization

    -- Smart suggestions metadata
    suggestion_source VARCHAR(50), -- 'manual', 'flyer', 'previous_items', 'popular'
    matching_confidence DECIMAL(3,2), -- 0.00 to 1.00 for auto-matched items

    -- Product availability
    availability_status VARCHAR(20) DEFAULT 'unknown', -- 'available', 'unavailable', 'unknown'
    availability_checked_at TIMESTAMP WITH TIME ZONE,

    -- Search vector for finding items
    search_vector tsvector,

    -- Timestamps
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),

    -- Constraints
    CONSTRAINT shopping_list_items_description_length CHECK (LENGTH(description) >= 1 AND LENGTH(description) <= 255),
    CONSTRAINT shopping_list_items_quantity_range CHECK (quantity > 0 AND quantity <= 999),
    CONSTRAINT shopping_list_items_confidence_range CHECK (matching_confidence IS NULL OR (matching_confidence >= 0 AND matching_confidence <= 1))
);

-- Indexes for performance
CREATE INDEX idx_shopping_list_items_list_id ON shopping_list_items(shopping_list_id);
CREATE INDEX idx_shopping_list_items_user_id ON shopping_list_items(user_id);
CREATE INDEX idx_shopping_list_items_list_order ON shopping_list_items(shopping_list_id, sort_order);
CREATE INDEX idx_shopping_list_items_checked ON shopping_list_items(shopping_list_id, is_checked);
CREATE INDEX idx_shopping_list_items_product_master ON shopping_list_items(product_master_id) WHERE product_master_id IS NOT NULL;
CREATE INDEX idx_shopping_list_items_store ON shopping_list_items(store_id) WHERE store_id IS NOT NULL;
CREATE INDEX idx_shopping_list_items_category ON shopping_list_items(category) WHERE category IS NOT NULL;
CREATE INDEX idx_shopping_list_items_normalized_desc ON shopping_list_items USING gin(normalized_description gin_trgm_ops);
CREATE INDEX idx_shopping_list_items_search_vector ON shopping_list_items USING gin(search_vector);
CREATE INDEX idx_shopping_list_items_tags ON shopping_list_items USING gin(tags);

-- Function to normalize Lithuanian text
CREATE OR REPLACE FUNCTION normalize_lithuanian_text(input_text TEXT)
RETURNS TEXT AS $$
BEGIN
    IF input_text IS NULL THEN
        RETURN NULL;
    END IF;

    RETURN lower(
        translate(
            input_text,
            'ąčęėįšųūžĄČĘĖĮŠŲŪŽ',
            'aceeisuuzACEEISUUZ'
        )
    );
END;
$$ LANGUAGE plpgsql IMMUTABLE;

-- Function to update search vector and normalized description
CREATE OR REPLACE FUNCTION update_shopping_list_item_search()
RETURNS trigger AS $$
BEGIN
    -- Update normalized description
    NEW.normalized_description := normalize_lithuanian_text(NEW.description);

    -- Update search vector
    NEW.search_vector := to_tsvector('lithuanian',
        COALESCE(NEW.description, '') || ' ' ||
        COALESCE(NEW.notes, '') || ' ' ||
        COALESCE(NEW.category, '') || ' ' ||
        COALESCE(array_to_string(NEW.tags, ' '), '')
    );

    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Trigger to update search data
CREATE TRIGGER shopping_list_items_search_trigger
BEFORE INSERT OR UPDATE ON shopping_list_items
FOR EACH ROW EXECUTE FUNCTION update_shopping_list_item_search();

-- Function to auto-set sort order for new items
CREATE OR REPLACE FUNCTION set_default_sort_order()
RETURNS trigger AS $$
BEGIN
    -- If no sort_order specified, put at end of list
    IF NEW.sort_order = 0 OR NEW.sort_order IS NULL THEN
        SELECT COALESCE(MAX(sort_order), 0) + 1
        INTO NEW.sort_order
        FROM shopping_list_items
        WHERE shopping_list_id = NEW.shopping_list_id;
    END IF;

    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Trigger to auto-set sort order
CREATE TRIGGER shopping_list_items_sort_order_trigger
BEFORE INSERT ON shopping_list_items
FOR EACH ROW EXECUTE FUNCTION set_default_sort_order();

-- Function to update checked status and timestamp
CREATE OR REPLACE FUNCTION update_item_checked_status()
RETURNS trigger AS $$
BEGIN
    -- If checked status changed
    IF OLD.is_checked != NEW.is_checked THEN
        IF NEW.is_checked = true THEN
            NEW.checked_at = NOW();
            NEW.checked_by_user_id = NEW.user_id; -- Will be set by application
        ELSE
            NEW.checked_at = NULL;
            NEW.checked_by_user_id = NULL;
        END IF;
    END IF;

    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Trigger to manage checked status
CREATE TRIGGER shopping_list_items_checked_trigger
BEFORE UPDATE ON shopping_list_items
FOR EACH ROW EXECUTE FUNCTION update_item_checked_status();

-- Trigger to update list statistics when items change
CREATE TRIGGER shopping_list_items_stats_trigger
AFTER INSERT OR UPDATE OR DELETE ON shopping_list_items
FOR EACH ROW EXECUTE FUNCTION update_shopping_list_stats();

-- Function to prevent duplicate items in same list
CREATE OR REPLACE FUNCTION prevent_duplicate_items()
RETURNS trigger AS $$
BEGIN
    -- Check for existing item with same normalized description
    IF EXISTS (
        SELECT 1 FROM shopping_list_items
        WHERE shopping_list_id = NEW.shopping_list_id
        AND normalized_description = normalize_lithuanian_text(NEW.description)
        AND id != COALESCE(NEW.id, 0) -- Exclude self for updates
    ) THEN
        RAISE EXCEPTION 'Item with description "%" already exists in this shopping list', NEW.description
            USING ERRCODE = 'unique_violation',
                  HINT = 'Consider increasing the quantity of the existing item instead';
    END IF;

    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Trigger to prevent duplicates (can be disabled for bulk operations)
CREATE TRIGGER shopping_list_items_duplicate_trigger
BEFORE INSERT OR UPDATE ON shopping_list_items
FOR EACH ROW EXECUTE FUNCTION prevent_duplicate_items();
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TRIGGER IF EXISTS shopping_list_items_duplicate_trigger ON shopping_list_items;
DROP TRIGGER IF EXISTS shopping_list_items_stats_trigger ON shopping_list_items;
DROP TRIGGER IF EXISTS shopping_list_items_checked_trigger ON shopping_list_items;
DROP TRIGGER IF EXISTS shopping_list_items_sort_order_trigger ON shopping_list_items;
DROP TRIGGER IF EXISTS shopping_list_items_search_trigger ON shopping_list_items;
DROP FUNCTION IF EXISTS prevent_duplicate_items();
DROP FUNCTION IF EXISTS update_item_checked_status();
DROP FUNCTION IF EXISTS set_default_sort_order();
DROP FUNCTION IF EXISTS update_shopping_list_item_search();
DROP FUNCTION IF EXISTS normalize_lithuanian_text(TEXT);
DROP TABLE IF EXISTS shopping_list_items;
-- +goose StatementEnd