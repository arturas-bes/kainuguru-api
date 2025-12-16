-- +goose Up
-- +goose StatementBegin

-- Fix the search vector trigger to normalize Lithuanian characters BEFORE creating tsvector
-- This ensures that both the indexed data and search queries use ASCII equivalents
-- and can match properly (e.g., "vištiena" stored as "vistiena" will match search "vistiena")

CREATE OR REPLACE FUNCTION update_product_search_vector()
RETURNS TRIGGER AS $$
BEGIN
    -- Normalize Lithuanian characters to ASCII before creating tsvector
    -- This ensures "vištiena" -> "vistiena" in the index, matching search queries
    NEW.search_vector := to_tsvector('lithuanian',
        sanitize_for_tsvector(normalize_lithuanian_text(COALESCE(NEW.name, ''))) || ' ' ||
        sanitize_for_tsvector(normalize_lithuanian_text(COALESCE(NEW.brand, ''))) || ' ' ||
        sanitize_for_tsvector(normalize_lithuanian_text(COALESCE(NEW.description, ''))) || ' ' ||
        sanitize_for_tsvector(normalize_lithuanian_text(COALESCE(NEW.category, ''))) || ' ' ||
        sanitize_for_tsvector(normalize_lithuanian_text(COALESCE(NEW.subcategory, '')))
    );
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Rebuild search vectors for all existing products
-- This triggers the updated function for each product
UPDATE products SET updated_at = updated_at;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

-- Restore original trigger without normalization
CREATE OR REPLACE FUNCTION update_product_search_vector()
RETURNS TRIGGER AS $$
BEGIN
    NEW.search_vector := to_tsvector('lithuanian',
        sanitize_for_tsvector(COALESCE(NEW.name, '')) || ' ' ||
        sanitize_for_tsvector(COALESCE(NEW.brand, '')) || ' ' ||
        sanitize_for_tsvector(COALESCE(NEW.description, '')) || ' ' ||
        sanitize_for_tsvector(COALESCE(NEW.category, '')) || ' ' ||
        sanitize_for_tsvector(COALESCE(NEW.subcategory, ''))
    );
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Rebuild search vectors
UPDATE products SET updated_at = updated_at;

-- +goose StatementEnd
