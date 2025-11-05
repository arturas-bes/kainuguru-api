-- +goose Up
-- +goose StatementBegin
CREATE TABLE product_masters (
    id BIGSERIAL PRIMARY KEY,

    -- Master product information
    name VARCHAR(255) NOT NULL,
    normalized_name VARCHAR(255) NOT NULL, -- Lithuanian chars normalized
    brand VARCHAR(100),
    description TEXT,

    -- Categorization
    category VARCHAR(100),
    subcategory VARCHAR(100),
    tags TEXT[], -- Array of tags for search and organization

    -- Standard units and packaging
    standard_unit VARCHAR(50), -- '1L', '500g', 'vnt', etc
    unit_type VARCHAR(20), -- 'volume', 'weight', 'count'
    standard_size DECIMAL(10,3), -- Numerical value of standard size
    packaging_variants TEXT[], -- Common size variants: ['500ml', '1L', '2L']

    -- Product identifiers and matching
    barcode VARCHAR(50), -- EAN/UPC barcode if available
    manufacturer_code VARCHAR(100),
    alternative_names TEXT[], -- Other names this product might be known by

    -- Statistics and quality metrics
    match_count INTEGER DEFAULT 0, -- How many products matched to this master
    confidence_score DECIMAL(3,2) DEFAULT 0, -- Overall confidence in master data (0-1)
    last_seen_date DATE, -- When this product was last seen in flyers

    -- Price tracking
    avg_price DECIMAL(10,2), -- Average price across all stores
    min_price DECIMAL(10,2), -- Minimum price seen
    max_price DECIMAL(10,2), -- Maximum price seen
    price_trend VARCHAR(20) DEFAULT 'stable', -- 'increasing', 'decreasing', 'stable'
    last_price_update TIMESTAMP WITH TIME ZONE,

    -- Availability and popularity
    availability_score DECIMAL(3,2) DEFAULT 0, -- How often available (0-1)
    popularity_score DECIMAL(3,2) DEFAULT 0, -- How often added to lists (0-1)
    seasonal_availability JSONB, -- Seasonal availability patterns

    -- Quality and user preferences
    user_rating DECIMAL(2,1), -- Average user rating (1-5)
    review_count INTEGER DEFAULT 0,
    nutritional_info JSONB, -- Nutritional information if available
    allergens TEXT[], -- Known allergens

    -- Search and matching
    search_vector tsvector,
    match_keywords TEXT[], -- Keywords that help match products to this master

    -- Status and lifecycle
    status VARCHAR(20) DEFAULT 'active', -- 'active', 'inactive', 'merged', 'deleted'
    merged_into_id BIGINT REFERENCES product_masters(id), -- If merged into another master

    -- Timestamps
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),

    -- Constraints
    CONSTRAINT product_masters_name_length CHECK (LENGTH(name) >= 1 AND LENGTH(name) <= 255),
    CONSTRAINT product_masters_confidence_range CHECK (confidence_score >= 0 AND confidence_score <= 1),
    CONSTRAINT product_masters_availability_range CHECK (availability_score >= 0 AND availability_score <= 1),
    CONSTRAINT product_masters_popularity_range CHECK (popularity_score >= 0 AND popularity_score <= 1),
    CONSTRAINT product_masters_rating_range CHECK (user_rating IS NULL OR (user_rating >= 1 AND user_rating <= 5))
);

-- Indexes for performance
CREATE INDEX idx_product_masters_normalized_name ON product_masters(normalized_name);
CREATE INDEX idx_product_masters_brand ON product_masters(brand) WHERE brand IS NOT NULL;
CREATE INDEX idx_product_masters_category ON product_masters(category) WHERE category IS NOT NULL;
CREATE INDEX idx_product_masters_barcode ON product_masters(barcode) WHERE barcode IS NOT NULL;
CREATE INDEX idx_product_masters_status ON product_masters(status);
CREATE INDEX idx_product_masters_popularity ON product_masters(popularity_score DESC);
CREATE INDEX idx_product_masters_availability ON product_masters(availability_score DESC);
CREATE INDEX idx_product_masters_avg_price ON product_masters(avg_price) WHERE avg_price IS NOT NULL;
CREATE INDEX idx_product_masters_search_vector ON product_masters USING gin(search_vector);
CREATE INDEX idx_product_masters_tags ON product_masters USING gin(tags);
CREATE INDEX idx_product_masters_match_keywords ON product_masters USING gin(match_keywords);
CREATE INDEX idx_product_masters_normalized_trgm ON product_masters USING gin(normalized_name gin_trgm_ops);
CREATE INDEX idx_product_masters_last_seen ON product_masters(last_seen_date DESC);

-- Unique constraint to prevent duplicate masters
CREATE UNIQUE INDEX idx_product_masters_unique_normalized
ON product_masters(normalized_name, brand, standard_unit)
WHERE status = 'active';

-- Function to update search vector and normalized name
CREATE OR REPLACE FUNCTION update_product_master_search()
RETURNS trigger AS $$
BEGIN
    -- Update normalized name
    NEW.normalized_name := normalize_lithuanian_text(NEW.name);

    -- Update search vector
    NEW.search_vector := to_tsvector('lithuanian',
        COALESCE(NEW.name, '') || ' ' ||
        COALESCE(NEW.brand, '') || ' ' ||
        COALESCE(NEW.description, '') || ' ' ||
        COALESCE(NEW.category, '') || ' ' ||
        COALESCE(NEW.subcategory, '') || ' ' ||
        COALESCE(array_to_string(NEW.tags, ' '), '') || ' ' ||
        COALESCE(array_to_string(NEW.alternative_names, ' '), '') || ' ' ||
        COALESCE(array_to_string(NEW.match_keywords, ' '), '')
    );

    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Trigger to update search data
CREATE TRIGGER product_masters_search_trigger
BEFORE INSERT OR UPDATE ON product_masters
FOR EACH ROW EXECUTE FUNCTION update_product_master_search();

-- Function to update statistics from products table
CREATE OR REPLACE FUNCTION update_product_master_stats()
RETURNS trigger AS $$
DECLARE
    master_id BIGINT;
    stats_record RECORD;
BEGIN
    -- Get the product_master_id from the changed record
    master_id := COALESCE(NEW.product_master_id, OLD.product_master_id);

    -- Skip if no master linked
    IF master_id IS NULL THEN
        RETURN COALESCE(NEW, OLD);
    END IF;

    -- Calculate statistics from linked products
    SELECT
        COUNT(*) as match_count,
        AVG(price_current) as avg_price,
        MIN(price_current) as min_price,
        MAX(price_current) as max_price,
        MAX(valid_to) as last_seen
    INTO stats_record
    FROM products
    WHERE product_master_id = master_id
    AND price_current IS NOT NULL
    AND valid_to >= CURRENT_DATE - INTERVAL '30 days'; -- Only recent data

    -- Update the product master
    UPDATE product_masters
    SET
        match_count = stats_record.match_count,
        avg_price = stats_record.avg_price,
        min_price = stats_record.min_price,
        max_price = stats_record.max_price,
        last_seen_date = stats_record.last_seen,
        last_price_update = CASE
            WHEN stats_record.avg_price IS NOT NULL THEN NOW()
            ELSE last_price_update
        END,
        updated_at = NOW()
    WHERE id = master_id;

    RETURN COALESCE(NEW, OLD);
END;
$$ LANGUAGE plpgsql;

-- Trigger to update master stats when products change
-- Note: This will be added after products table is updated with the trigger
-- CREATE TRIGGER products_master_stats_trigger
-- AFTER INSERT OR UPDATE OR DELETE ON products
-- FOR EACH ROW EXECUTE FUNCTION update_product_master_stats();

-- Function to calculate popularity and availability scores
CREATE OR REPLACE FUNCTION update_product_master_scores()
RETURNS void AS $$
DECLARE
    master_record RECORD;
    shopping_list_usage INTEGER;
    availability_weeks INTEGER;
    total_weeks INTEGER;
BEGIN
    FOR master_record IN
        SELECT id FROM product_masters WHERE status = 'active'
    LOOP
        -- Calculate shopping list usage (popularity)
        SELECT COUNT(DISTINCT shopping_list_id)
        INTO shopping_list_usage
        FROM shopping_list_items
        WHERE product_master_id = master_record.id;

        -- Calculate availability (how many weeks product was available in last 12 weeks)
        SELECT
            COUNT(DISTINCT date_trunc('week', valid_from)) as availability_weeks,
            12 as total_weeks
        INTO availability_weeks, total_weeks
        FROM products
        WHERE product_master_id = master_record.id
        AND valid_from >= CURRENT_DATE - INTERVAL '12 weeks';

        -- Update scores
        UPDATE product_masters
        SET
            popularity_score = LEAST(1.0, shopping_list_usage / 100.0), -- Scale to 0-1
            availability_score = CASE
                WHEN total_weeks > 0 THEN availability_weeks::DECIMAL / total_weeks
                ELSE 0
            END,
            updated_at = NOW()
        WHERE id = master_record.id;
    END LOOP;
END;
$$ LANGUAGE plpgsql;

-- Function to merge duplicate product masters
CREATE OR REPLACE FUNCTION merge_product_masters(
    source_id BIGINT,
    target_id BIGINT
) RETURNS void AS $$
BEGIN
    -- Validate that both masters exist and are different
    IF source_id = target_id THEN
        RAISE EXCEPTION 'Cannot merge product master into itself';
    END IF;

    IF NOT EXISTS (SELECT 1 FROM product_masters WHERE id = source_id AND status = 'active') THEN
        RAISE EXCEPTION 'Source product master % does not exist or is not active', source_id;
    END IF;

    IF NOT EXISTS (SELECT 1 FROM product_masters WHERE id = target_id AND status = 'active') THEN
        RAISE EXCEPTION 'Target product master % does not exist or is not active', target_id;
    END IF;

    -- Update all references to point to target
    UPDATE products SET product_master_id = target_id WHERE product_master_id = source_id;
    UPDATE shopping_list_items SET product_master_id = target_id WHERE product_master_id = source_id;

    -- Mark source as merged
    UPDATE product_masters
    SET
        status = 'merged',
        merged_into_id = target_id,
        updated_at = NOW()
    WHERE id = source_id;

    -- Update target statistics
    PERFORM update_product_master_stats();
END;
$$ LANGUAGE plpgsql;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP FUNCTION IF EXISTS merge_product_masters(BIGINT, BIGINT);
DROP FUNCTION IF EXISTS update_product_master_scores();
DROP FUNCTION IF EXISTS update_product_master_stats();
DROP TRIGGER IF EXISTS product_masters_search_trigger ON product_masters;
DROP FUNCTION IF EXISTS update_product_master_search();
DROP TABLE IF EXISTS product_masters;
-- +goose StatementEnd