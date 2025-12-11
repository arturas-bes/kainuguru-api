-- +goose Up
-- +goose StatementBegin

-- Add matching indexes for product_masters
CREATE INDEX IF NOT EXISTS idx_product_masters_normalized_name_trgm
ON product_masters USING gin (normalized_name gin_trgm_ops);

CREATE INDEX IF NOT EXISTS idx_product_masters_brand
ON product_masters (brand) WHERE brand IS NOT NULL;

CREATE INDEX IF NOT EXISTS idx_product_masters_status_confidence
ON product_masters (status, confidence_score);

CREATE INDEX IF NOT EXISTS idx_product_masters_category
ON product_masters (category) WHERE category IS NOT NULL;

-- Add confidence_score and match tracking columns if they don't exist
ALTER TABLE product_masters
ADD COLUMN IF NOT EXISTS confidence_score DECIMAL(3,2) DEFAULT 0.50,
ADD COLUMN IF NOT EXISTS last_seen_date TIMESTAMP WITH TIME ZONE,
ADD COLUMN IF NOT EXISTS match_count INTEGER DEFAULT 0;

-- Create product_master_matches table for audit trail
-- MOVED TO 031_create_product_master_matches.sql
-- CREATE TABLE IF NOT EXISTS product_master_matches (
--     id BIGSERIAL PRIMARY KEY,
--     product_id BIGINT NOT NULL,
--     master_id BIGINT NOT NULL REFERENCES product_masters(id),
--     confidence_score DECIMAL(3,2) NOT NULL,
--     match_method VARCHAR(50) NOT NULL,
--     matched_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
--     matched_by VARCHAR(100),
--     is_verified BOOLEAN DEFAULT FALSE
-- );

-- Add indexes for product_master_matches
-- MOVED TO 031_create_product_master_matches.sql
-- CREATE INDEX IF NOT EXISTS idx_pmm_product ON product_master_matches(product_id);
-- CREATE INDEX IF NOT EXISTS idx_pmm_master ON product_master_matches(master_id);
-- CREATE INDEX IF NOT EXISTS idx_pmm_matched_at ON product_master_matches(matched_at);
-- CREATE INDEX IF NOT EXISTS idx_pmm_method ON product_master_matches(match_method);

-- Add index for products.product_master_id if not exists
CREATE INDEX IF NOT EXISTS idx_products_product_master_id ON products(product_master_id);

-- Update existing product_masters to have reasonable confidence scores
UPDATE product_masters
SET confidence_score = CASE
    WHEN match_count >= 10 THEN 0.9
    WHEN match_count >= 5 THEN 0.7
    WHEN match_count >= 2 THEN 0.6
    ELSE 0.5
END
WHERE confidence_score = 0 OR confidence_score IS NULL;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DROP TABLE IF EXISTS product_master_matches;

DROP INDEX IF EXISTS idx_product_masters_normalized_name_trgm;
DROP INDEX IF EXISTS idx_product_masters_brand;
DROP INDEX IF EXISTS idx_product_masters_status_confidence;
DROP INDEX IF EXISTS idx_product_masters_category;
DROP INDEX IF EXISTS idx_products_product_master_id;

ALTER TABLE product_masters
DROP COLUMN IF EXISTS confidence_score,
DROP COLUMN IF EXISTS last_seen_date,
DROP COLUMN IF EXISTS match_count;

-- +goose StatementEnd
