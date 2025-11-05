-- +goose Up
-- +goose StatementBegin
-- Enable trigram extension for fuzzy searching
CREATE EXTENSION IF NOT EXISTS pg_trgm;

CREATE TABLE products (
    id BIGSERIAL,
    flyer_id INTEGER NOT NULL,
    flyer_page_id INTEGER REFERENCES flyer_pages(id),
    store_id INTEGER NOT NULL,

    -- Product information
    name VARCHAR(255) NOT NULL,
    normalized_name VARCHAR(255) NOT NULL, -- Lithuanian chars normalized
    brand VARCHAR(100),
    description TEXT,

    -- Pricing
    price_current DECIMAL(10,2),
    price_original DECIMAL(10,2),
    discount_percentage DECIMAL(5,2),

    -- Units and quantity
    unit VARCHAR(50), -- '1L', '500g', 'vnt', etc
    quantity DECIMAL(10,3),

    -- Categorization
    category VARCHAR(100),
    tags TEXT[], -- Array of tags for search

    -- Matching to master catalog
    product_master_id INTEGER,
    matching_confidence DECIMAL(3,2), -- 0.00 to 1.00

    -- Dates for partitioning
    valid_from DATE NOT NULL,
    valid_to DATE NOT NULL,

    -- Search vectors
    search_vector tsvector,

    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),

    PRIMARY KEY (id, valid_from)
) PARTITION BY RANGE (valid_from);

-- Create function to auto-generate weekly partitions
CREATE OR REPLACE FUNCTION create_weekly_partition()
RETURNS void AS $$
DECLARE
    start_date date;
    end_date date;
    partition_name text;
BEGIN
    -- Calculate current week boundaries
    start_date := date_trunc('week', CURRENT_DATE)::date;
    end_date := start_date + INTERVAL '7 days';
    partition_name := 'products_' || to_char(start_date, 'YYYY_WW');

    -- Create partition if not exists
    IF NOT EXISTS (
        SELECT 1 FROM pg_class WHERE relname = partition_name
    ) THEN
        EXECUTE format(
            'CREATE TABLE %I PARTITION OF products FOR VALUES FROM (%L) TO (%L)',
            partition_name, start_date, end_date
        );

        -- Create indexes on new partition
        EXECUTE format('CREATE INDEX ON %I (flyer_id)', partition_name);
        EXECUTE format('CREATE INDEX ON %I (store_id)', partition_name);
        EXECUTE format('CREATE INDEX ON %I (product_master_id)', partition_name);
        EXECUTE format('CREATE INDEX ON %I USING gin(search_vector)', partition_name);
        EXECUTE format('CREATE INDEX ON %I USING gin(normalized_name gin_trgm_ops)', partition_name);
    END IF;
END;
$$ LANGUAGE plpgsql;

-- Trigger to update search vector
CREATE OR REPLACE FUNCTION update_product_search_vector()
RETURNS trigger AS $$
BEGIN
    NEW.search_vector := to_tsvector('lithuanian',
        COALESCE(NEW.name, '') || ' ' ||
        COALESCE(NEW.brand, '') || ' ' ||
        COALESCE(NEW.description, '')
    );
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER products_search_vector_trigger
BEFORE INSERT OR UPDATE ON products
FOR EACH ROW EXECUTE FUNCTION update_product_search_vector();

-- Create current week partition
SELECT create_weekly_partition();

-- Create next week partition (for edge cases)
DO $$
DECLARE
    start_date date;
    end_date date;
    partition_name text;
BEGIN
    start_date := date_trunc('week', CURRENT_DATE + INTERVAL '7 days')::date;
    end_date := start_date + INTERVAL '7 days';
    partition_name := 'products_' || to_char(start_date, 'YYYY_WW');

    IF NOT EXISTS (
        SELECT 1 FROM pg_class WHERE relname = partition_name
    ) THEN
        EXECUTE format(
            'CREATE TABLE %I PARTITION OF products FOR VALUES FROM (%L) TO (%L)',
            partition_name, start_date, end_date
        );

        -- Create indexes on new partition
        EXECUTE format('CREATE INDEX ON %I (flyer_id)', partition_name);
        EXECUTE format('CREATE INDEX ON %I (store_id)', partition_name);
        EXECUTE format('CREATE INDEX ON %I (product_master_id)', partition_name);
        EXECUTE format('CREATE INDEX ON %I USING gin(search_vector)', partition_name);
        EXECUTE format('CREATE INDEX ON %I USING gin(normalized_name gin_trgm_ops)', partition_name);
    END IF;
END $$;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TRIGGER IF EXISTS products_search_vector_trigger ON products;
DROP FUNCTION IF EXISTS update_product_search_vector();
DROP FUNCTION IF EXISTS create_weekly_partition();
DROP TABLE IF EXISTS products;
DROP EXTENSION IF EXISTS pg_trgm;
-- +goose StatementEnd