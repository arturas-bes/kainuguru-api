-- +goose Up
-- +goose StatementBegin

-- Add all missing columns that exist in the Product Go model but not in the database
-- This aligns the database schema with the GraphQL schema and Go model

-- Unit and packaging information
ALTER TABLE products ADD COLUMN IF NOT EXISTS unit_size VARCHAR(50);
ALTER TABLE products ADD COLUMN IF NOT EXISTS unit_type VARCHAR(50);
ALTER TABLE products ADD COLUMN IF NOT EXISTS unit_price VARCHAR(100);
ALTER TABLE products ADD COLUMN IF NOT EXISTS package_size VARCHAR(50);
ALTER TABLE products ADD COLUMN IF NOT EXISTS weight VARCHAR(50);
ALTER TABLE products ADD COLUMN IF NOT EXISTS volume VARCHAR(50);

-- Visual and position data
ALTER TABLE products ADD COLUMN IF NOT EXISTS image_url TEXT;
ALTER TABLE products ADD COLUMN IF NOT EXISTS bounding_box JSONB;
ALTER TABLE products ADD COLUMN IF NOT EXISTS page_position JSONB;

-- Sale date tracking
ALTER TABLE products ADD COLUMN IF NOT EXISTS sale_start_date TIMESTAMP WITH TIME ZONE;
ALTER TABLE products ADD COLUMN IF NOT EXISTS sale_end_date TIMESTAMP WITH TIME ZONE;

-- Stock and availability
ALTER TABLE products ADD COLUMN IF NOT EXISTS stock_level VARCHAR(50);

-- AI extraction metadata
ALTER TABLE products ADD COLUMN IF NOT EXISTS extraction_confidence DECIMAL(3,2) DEFAULT 0.0;
ALTER TABLE products ADD COLUMN IF NOT EXISTS extraction_method VARCHAR(50) DEFAULT 'ocr';
ALTER TABLE products ADD COLUMN IF NOT EXISTS requires_review BOOLEAN DEFAULT FALSE;

-- Timestamp tracking
ALTER TABLE products ADD COLUMN IF NOT EXISTS updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW();

-- Create trigger to auto-update updated_at timestamp
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS update_products_updated_at ON products;
CREATE TRIGGER update_products_updated_at
BEFORE UPDATE ON products
FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

-- Drop trigger
DROP TRIGGER IF EXISTS update_products_updated_at ON products;

-- Remove added columns (keep function as other tables may use it)
ALTER TABLE products DROP COLUMN IF EXISTS updated_at;
ALTER TABLE products DROP COLUMN IF EXISTS requires_review;
ALTER TABLE products DROP COLUMN IF EXISTS extraction_method;
ALTER TABLE products DROP COLUMN IF EXISTS extraction_confidence;
ALTER TABLE products DROP COLUMN IF EXISTS stock_level;
ALTER TABLE products DROP COLUMN IF EXISTS sale_end_date;
ALTER TABLE products DROP COLUMN IF EXISTS sale_start_date;
ALTER TABLE products DROP COLUMN IF EXISTS page_position;
ALTER TABLE products DROP COLUMN IF EXISTS bounding_box;
ALTER TABLE products DROP COLUMN IF EXISTS image_url;
ALTER TABLE products DROP COLUMN IF EXISTS volume;
ALTER TABLE products DROP COLUMN IF EXISTS weight;
ALTER TABLE products DROP COLUMN IF EXISTS package_size;
ALTER TABLE products DROP COLUMN IF EXISTS unit_price;
ALTER TABLE products DROP COLUMN IF EXISTS unit_type;
ALTER TABLE products DROP COLUMN IF EXISTS unit_size;

-- +goose StatementEnd
