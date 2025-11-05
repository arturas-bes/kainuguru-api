-- +goose Up
-- +goose StatementBegin
-- Enhanced Lithuanian full-text search configuration for product search

-- Create Lithuanian text search configuration
CREATE TEXT SEARCH CONFIGURATION lithuanian_search (COPY = simple);

-- Configure Lithuanian text search dictionary mappings
ALTER TEXT SEARCH CONFIGURATION lithuanian_search
    ALTER MAPPING FOR asciiword, asciihword, hword_asciipart, word, hword, hword_part
    WITH lithuanian_ispell, lithuanian_stem, simple;

-- Create custom Lithuanian stop words to improve search relevance
CREATE TEXT SEARCH DICTIONARY lithuanian_stop (
    TEMPLATE = simple,
    STOPWORDS = lithuanian
);

-- Add Lithuanian stop words list
INSERT INTO pg_ts_config_map (cfgname, maptokentype, mapseqno, mapdict)
SELECT 'lithuanian_search', 12, 1, 'lithuanian_stop'::regdictionary
WHERE NOT EXISTS (
    SELECT 1 FROM pg_ts_config_map
    WHERE cfgname = 'lithuanian_search' AND maptokentype = 12
);

-- Create function to normalize Lithuanian text for search
CREATE OR REPLACE FUNCTION normalize_lithuanian_text(input_text TEXT)
RETURNS TEXT AS $$
BEGIN
    IF input_text IS NULL OR input_text = '' THEN
        RETURN '';
    END IF;

    -- Convert to lowercase
    input_text := LOWER(input_text);

    -- Normalize Lithuanian diacritics for better search matching
    input_text := TRANSLATE(input_text, 'ąčęėįšųūž', 'aceeisuuz');

    -- Remove common punctuation and extra spaces
    input_text := REGEXP_REPLACE(input_text, '[,.()/-]', ' ', 'g');
    input_text := REGEXP_REPLACE(input_text, '\s+', ' ', 'g');
    input_text := TRIM(input_text);

    RETURN input_text;
END;
$$ LANGUAGE plpgsql IMMUTABLE;

-- Create function to generate search vectors with Lithuanian support
CREATE OR REPLACE FUNCTION generate_product_search_vector(
    product_name TEXT,
    brand TEXT DEFAULT NULL,
    category TEXT DEFAULT NULL,
    description TEXT DEFAULT NULL
)
RETURNS tsvector AS $$
DECLARE
    search_text TEXT;
    normalized_text TEXT;
BEGIN
    -- Combine all searchable fields
    search_text := COALESCE(product_name, '');

    IF brand IS NOT NULL AND brand != '' THEN
        search_text := search_text || ' ' || brand;
    END IF;

    IF category IS NOT NULL AND category != '' THEN
        search_text := search_text || ' ' || category;
    END IF;

    IF description IS NOT NULL AND description != '' THEN
        search_text := search_text || ' ' || description;
    END IF;

    -- Normalize the text
    normalized_text := normalize_lithuanian_text(search_text);

    -- Generate search vector with Lithuanian configuration and weight different fields
    RETURN
        setweight(to_tsvector('lithuanian_search', COALESCE(product_name, '')), 'A') ||
        setweight(to_tsvector('lithuanian_search', COALESCE(brand, '')), 'B') ||
        setweight(to_tsvector('lithuanian_search', COALESCE(category, '')), 'C') ||
        setweight(to_tsvector('lithuanian_search', COALESCE(description, '')), 'D') ||
        setweight(to_tsvector('simple', normalized_text), 'B');
END;
$$ LANGUAGE plpgsql IMMUTABLE;

-- Update existing products to use the new search vector function
UPDATE products SET
    search_vector = generate_product_search_vector(name, brand, category, description),
    updated_at = NOW()
WHERE search_vector IS NULL OR search_vector = '';

-- Create trigger to automatically update search vectors
CREATE OR REPLACE FUNCTION update_product_search_vector()
RETURNS TRIGGER AS $$
BEGIN
    NEW.search_vector := generate_product_search_vector(
        NEW.name,
        NEW.brand,
        NEW.category,
        NEW.description
    );
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Drop existing trigger if it exists and create new one
DROP TRIGGER IF EXISTS product_search_vector_trigger ON products;
CREATE TRIGGER product_search_vector_trigger
    BEFORE INSERT OR UPDATE OF name, brand, category, description
    ON products
    FOR EACH ROW
    EXECUTE FUNCTION update_product_search_vector();

-- Create function for Lithuanian-aware similarity search
CREATE OR REPLACE FUNCTION lithuanian_similarity(text1 TEXT, text2 TEXT)
RETURNS FLOAT AS $$
BEGIN
    RETURN similarity(
        normalize_lithuanian_text(text1),
        normalize_lithuanian_text(text2)
    );
END;
$$ LANGUAGE plpgsql IMMUTABLE;

-- Create search function for products with ranking
CREATE OR REPLACE FUNCTION search_products_ranked(
    search_query TEXT,
    limit_count INTEGER DEFAULT 50,
    offset_count INTEGER DEFAULT 0,
    store_ids INTEGER[] DEFAULT NULL,
    min_price DECIMAL DEFAULT NULL,
    max_price DECIMAL DEFAULT NULL,
    on_sale_only BOOLEAN DEFAULT FALSE
)
RETURNS TABLE(
    product_id INTEGER,
    name TEXT,
    brand TEXT,
    current_price DECIMAL,
    store_id INTEGER,
    flyer_id INTEGER,
    search_rank REAL,
    similarity_rank REAL
) AS $$
DECLARE
    query_vector tsvector;
    normalized_query TEXT;
BEGIN
    -- Normalize the search query
    normalized_query := normalize_lithuanian_text(search_query);
    query_vector := to_tsvector('lithuanian_search', search_query);

    RETURN QUERY
    SELECT
        p.id,
        p.name,
        p.brand,
        p.current_price,
        p.store_id,
        p.flyer_id,
        ts_rank_cd(p.search_vector, query_vector) AS search_rank,
        GREATEST(
            lithuanian_similarity(p.name, search_query),
            COALESCE(lithuanian_similarity(p.brand, search_query), 0)
        ) AS similarity_rank
    FROM products p
    INNER JOIN flyers f ON f.id = p.flyer_id
    WHERE
        -- Text search conditions
        (
            p.search_vector @@ query_vector OR
            lithuanian_similarity(p.name, search_query) > 0.3 OR
            COALESCE(lithuanian_similarity(p.brand, search_query), 0) > 0.3
        )
        -- Flyer must be current and not archived
        AND f.is_archived = FALSE
        AND f.valid_from <= NOW()
        AND f.valid_to >= NOW()
        -- Product must be available
        AND p.is_available = TRUE
        -- Optional filters
        AND (store_ids IS NULL OR p.store_id = ANY(store_ids))
        AND (min_price IS NULL OR p.current_price >= min_price)
        AND (max_price IS NULL OR p.current_price <= max_price)
        AND (on_sale_only = FALSE OR p.is_on_sale = TRUE)
    ORDER BY
        -- Rank by relevance (FTS rank + similarity)
        (ts_rank_cd(p.search_vector, query_vector) + similarity_rank) DESC,
        -- Secondary sort by price (cheaper first)
        p.current_price ASC,
        -- Tertiary sort by name
        p.name ASC
    LIMIT limit_count OFFSET offset_count;
END;
$$ LANGUAGE plpgsql;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
-- Remove Lithuanian FTS configuration

DROP FUNCTION IF EXISTS search_products_ranked(TEXT, INTEGER, INTEGER, INTEGER[], DECIMAL, DECIMAL, BOOLEAN);
DROP FUNCTION IF EXISTS lithuanian_similarity(TEXT, TEXT);
DROP TRIGGER IF EXISTS product_search_vector_trigger ON products;
DROP FUNCTION IF EXISTS update_product_search_vector();
DROP FUNCTION IF EXISTS generate_product_search_vector(TEXT, TEXT, TEXT, TEXT);
DROP FUNCTION IF EXISTS normalize_lithuanian_text(TEXT);
DROP TEXT SEARCH DICTIONARY IF EXISTS lithuanian_stop;
DROP TEXT SEARCH CONFIGURATION IF EXISTS lithuanian_search;

-- Reset search_vector to simple configuration
UPDATE products SET
    search_vector = to_tsvector('simple', COALESCE(name, '') || ' ' || COALESCE(brand, ''))
WHERE search_vector IS NOT NULL;

-- +goose StatementEnd