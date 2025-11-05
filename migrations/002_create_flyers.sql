-- +goose Up
-- +goose StatementBegin
CREATE TABLE flyers (
    id SERIAL PRIMARY KEY,
    store_id INTEGER NOT NULL REFERENCES stores(id),
    title VARCHAR(255),
    valid_from DATE NOT NULL,
    valid_to DATE NOT NULL,
    page_count INTEGER,
    source_url TEXT,

    -- Archival status from clarifications
    is_archived BOOLEAN DEFAULT FALSE,
    archived_at TIMESTAMP WITH TIME ZONE,

    -- Processing metadata
    status VARCHAR(50) DEFAULT 'pending', -- pending, processing, completed, failed
    extraction_started_at TIMESTAMP WITH TIME ZONE,
    extraction_completed_at TIMESTAMP WITH TIME ZONE,
    products_extracted INTEGER DEFAULT 0,

    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),

    CONSTRAINT flyers_date_check CHECK (valid_to >= valid_from)
);

CREATE INDEX idx_flyers_store ON flyers(store_id);
CREATE INDEX idx_flyers_validity ON flyers(valid_from, valid_to);
CREATE INDEX idx_flyers_status ON flyers(status);
CREATE INDEX idx_flyers_current ON flyers(valid_from, valid_to)
    WHERE is_archived = FALSE;

-- Add trigger to update updated_at timestamp
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ language 'plpgsql';

CREATE TRIGGER update_flyers_updated_at BEFORE UPDATE ON flyers
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TRIGGER IF EXISTS update_flyers_updated_at ON flyers;
DROP FUNCTION IF EXISTS update_updated_at_column();
DROP TABLE IF EXISTS flyers;
-- +goose StatementEnd