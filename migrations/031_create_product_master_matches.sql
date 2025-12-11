-- +goose Up
-- +goose StatementBegin
-- Migration: Create product_master_matches table
-- This table stores matching attempts between products and product masters for audit/review purposes

CREATE TABLE IF NOT EXISTS product_master_matches (
    id BIGSERIAL PRIMARY KEY,
    product_id BIGINT NOT NULL, -- No FK due to composite PK on products table
    product_master_id BIGINT NOT NULL REFERENCES product_masters(id) ON DELETE CASCADE,
    confidence DECIMAL(5, 4) NOT NULL CHECK (confidence >= 0 AND confidence <= 1),
    match_type VARCHAR(50) NOT NULL, -- 'exact', 'fuzzy', 'brand', 'category', 'manual'
    match_score DECIMAL(5, 4), -- Additional score for fuzzy matching
    matched_fields JSONB, -- Which fields matched: {name: true, brand: true, etc}
    review_status VARCHAR(50) DEFAULT 'pending', -- 'pending', 'approved', 'rejected'
    reviewed_by UUID REFERENCES users(id) ON DELETE SET NULL,
    reviewed_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    
    UNIQUE(product_id, product_master_id)
);

-- Indexes for efficient querying
CREATE INDEX idx_product_master_matches_product_id ON product_master_matches(product_id);
CREATE INDEX idx_product_master_matches_master_id ON product_master_matches(product_master_id);
CREATE INDEX idx_product_master_matches_confidence ON product_master_matches(confidence DESC);
CREATE INDEX idx_product_master_matches_review_status ON product_master_matches(review_status) WHERE review_status = 'pending';
CREATE INDEX idx_product_master_matches_match_type ON product_master_matches(match_type);

-- Updated at trigger
CREATE TRIGGER update_product_master_matches_updated_at
    BEFORE UPDATE ON product_master_matches
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- Comments
COMMENT ON TABLE product_master_matches IS 'Stores matching attempts between products and product masters for quality control and review';
COMMENT ON COLUMN product_master_matches.confidence IS 'Matching confidence score from 0 to 1';
COMMENT ON COLUMN product_master_matches.match_type IS 'Type of matching algorithm used';
COMMENT ON COLUMN product_master_matches.match_score IS 'Additional scoring metric for fuzzy/similarity matching';
COMMENT ON COLUMN product_master_matches.matched_fields IS 'JSON object indicating which fields contributed to the match';
COMMENT ON COLUMN product_master_matches.review_status IS 'Status of manual review: pending, approved, or rejected';
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS product_master_matches;
-- +goose StatementEnd
