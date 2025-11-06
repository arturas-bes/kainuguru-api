-- +goose Up
-- +goose StatementBegin
-- Update products table schema to match the Go model and GraphQL schema

-- Rename existing columns
ALTER TABLE products RENAME COLUMN price_current TO current_price;
ALTER TABLE products RENAME COLUMN price_original TO original_price;
ALTER TABLE products RENAME COLUMN discount_percentage TO discount_percent;

-- Add missing product specification columns
ALTER TABLE products ADD COLUMN IF NOT EXISTS subcategory VARCHAR(100);
ALTER TABLE products ADD COLUMN IF NOT EXISTS unit_size VARCHAR(50);
ALTER TABLE products ADD COLUMN IF NOT EXISTS unit_type VARCHAR(50);
ALTER TABLE products ADD COLUMN IF NOT EXISTS unit_price VARCHAR(50);
ALTER TABLE products ADD COLUMN IF NOT EXISTS package_size VARCHAR(50);
ALTER TABLE products ADD COLUMN IF NOT EXISTS weight VARCHAR(50);
ALTER TABLE products ADD COLUMN IF NOT EXISTS volume VARCHAR(50);

-- Add image and location columns
ALTER TABLE products ADD COLUMN IF NOT EXISTS image_url TEXT;
ALTER TABLE products ADD COLUMN IF NOT EXISTS bounding_box JSONB;
ALTER TABLE products ADD COLUMN IF NOT EXISTS page_position JSONB;

-- Add availability and promotion columns
ALTER TABLE products ADD COLUMN IF NOT EXISTS is_on_sale BOOLEAN DEFAULT false;
ALTER TABLE products ADD COLUMN IF NOT EXISTS sale_start_date TIMESTAMP WITH TIME ZONE;
ALTER TABLE products ADD COLUMN IF NOT EXISTS sale_end_date TIMESTAMP WITH TIME ZONE;
ALTER TABLE products ADD COLUMN IF NOT EXISTS is_available BOOLEAN DEFAULT true;
ALTER TABLE products ADD COLUMN IF NOT EXISTS stock_level VARCHAR(50);

-- Add extraction metadata columns
ALTER TABLE products ADD COLUMN IF NOT EXISTS extraction_confidence DECIMAL(3,2) DEFAULT 0.0;
ALTER TABLE products ADD COLUMN IF NOT EXISTS extraction_method VARCHAR(50) DEFAULT 'ocr';
ALTER TABLE products ADD COLUMN IF NOT EXISTS requires_review BOOLEAN DEFAULT false;

-- Add updated_at timestamp
ALTER TABLE products ADD COLUMN IF NOT EXISTS updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW();

-- Note: Cannot change valid_from as it's part of the partition key
-- Keep valid_from as DATE, change valid_to only
ALTER TABLE products ALTER COLUMN valid_to TYPE TIMESTAMP WITH TIME ZONE;

-- Drop old unit and quantity columns (replaced by more specific columns)
ALTER TABLE products DROP COLUMN IF EXISTS unit;
ALTER TABLE products DROP COLUMN IF EXISTS quantity;

-- Drop tags column (replaced by category/subcategory)
ALTER TABLE products DROP COLUMN IF EXISTS tags;

-- Rename matching_confidence to use extraction_confidence instead
ALTER TABLE products DROP COLUMN IF EXISTS matching_confidence;

-- Update trigger function to include new fields in search vector
CREATE OR REPLACE FUNCTION update_product_search_vector()
RETURNS trigger AS $$
BEGIN
    NEW.search_vector := to_tsvector('lithuanian',
        COALESCE(NEW.name, '') || ' ' ||
        COALESCE(NEW.brand, '') || ' ' ||
        COALESCE(NEW.description, '') || ' ' ||
        COALESCE(NEW.category, '') || ' ' ||
        COALESCE(NEW.subcategory, '')
    );
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Create trigger to update updated_at timestamp (function may already exist, that's ok)
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Drop trigger if it exists, then create it
DROP TRIGGER IF EXISTS update_products_updated_at ON products;
CREATE TRIGGER update_products_updated_at
BEFORE UPDATE ON products
FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
-- Revert products table schema changes

-- Drop triggers (keep function as it's used by other tables)
DROP TRIGGER IF EXISTS update_products_updated_at ON products;

-- Revert search vector function
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

-- Restore old columns
ALTER TABLE products ADD COLUMN IF NOT EXISTS unit VARCHAR(50);
ALTER TABLE products ADD COLUMN IF NOT EXISTS quantity DECIMAL(10,3);
ALTER TABLE products ADD COLUMN IF NOT EXISTS tags TEXT[];
ALTER TABLE products ADD COLUMN IF NOT EXISTS matching_confidence DECIMAL(3,2);

-- Change valid_to back to date (valid_from stays as is - partition key)
ALTER TABLE products ALTER COLUMN valid_to TYPE DATE;

-- Remove new columns
ALTER TABLE products DROP COLUMN IF EXISTS updated_at;
ALTER TABLE products DROP COLUMN IF EXISTS requires_review;
ALTER TABLE products DROP COLUMN IF EXISTS extraction_method;
ALTER TABLE products DROP COLUMN IF EXISTS extraction_confidence;
ALTER TABLE products DROP COLUMN IF EXISTS stock_level;
ALTER TABLE products DROP COLUMN IF EXISTS is_available;
ALTER TABLE products DROP COLUMN IF EXISTS sale_end_date;
ALTER TABLE products DROP COLUMN IF EXISTS sale_start_date;
ALTER TABLE products DROP COLUMN IF EXISTS is_on_sale;
ALTER TABLE products DROP COLUMN IF EXISTS page_position;
ALTER TABLE products DROP COLUMN IF EXISTS bounding_box;
ALTER TABLE products DROP COLUMN IF EXISTS image_url;
ALTER TABLE products DROP COLUMN IF EXISTS volume;
ALTER TABLE products DROP COLUMN IF EXISTS weight;
ALTER TABLE products DROP COLUMN IF EXISTS package_size;
ALTER TABLE products DROP COLUMN IF EXISTS unit_price;
ALTER TABLE products DROP COLUMN IF EXISTS unit_type;
ALTER TABLE products DROP COLUMN IF EXISTS unit_size;
ALTER TABLE products DROP COLUMN IF EXISTS subcategory;

-- Revert column renames
ALTER TABLE products RENAME COLUMN discount_percent TO discount_percentage;
ALTER TABLE products RENAME COLUMN original_price TO price_original;
ALTER TABLE products RENAME COLUMN current_price TO price_current;

-- +goose StatementEnd
