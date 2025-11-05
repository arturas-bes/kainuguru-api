-- +goose Up
-- +goose StatementBegin
CREATE TABLE flyer_pages (
    id SERIAL PRIMARY KEY,
    flyer_id INTEGER NOT NULL REFERENCES flyers(id) ON DELETE CASCADE,
    page_number INTEGER NOT NULL,
    image_url TEXT,

    -- From clarifications: track extraction failures
    extraction_status VARCHAR(50) DEFAULT 'pending',
    extraction_attempts INTEGER DEFAULT 0,
    extraction_error TEXT,
    needs_manual_review BOOLEAN DEFAULT FALSE,

    -- Store extracted raw data for debugging
    raw_extraction_data JSONB,

    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),

    UNIQUE(flyer_id, page_number)
);

CREATE INDEX idx_flyer_pages_flyer ON flyer_pages(flyer_id);
CREATE INDEX idx_flyer_pages_status ON flyer_pages(extraction_status);
CREATE INDEX idx_flyer_pages_review ON flyer_pages(needs_manual_review)
    WHERE needs_manual_review = TRUE;

-- Add trigger to update updated_at timestamp
CREATE TRIGGER update_flyer_pages_updated_at BEFORE UPDATE ON flyer_pages
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TRIGGER IF EXISTS update_flyer_pages_updated_at ON flyer_pages;
DROP TABLE IF EXISTS flyer_pages;
-- +goose StatementEnd