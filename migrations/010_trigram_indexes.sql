-- +goose Up
-- +goose StatementBegin
-- Advanced trigram search configuration for fuzzy matching

-- Ensure pg_trgm extension is available (already created in 004_create_products.sql but ensure it's there)
CREATE EXTENSION IF NOT EXISTS pg_trgm;

-- Configure trigram similarity threshold for Lithuanian language
-- Lower threshold for better fuzzy matching with Lithuanian diacritics
SET pg_trgm.similarity_threshold = 0.3;

-- Create specialized trigram indexes for different search patterns

-- Main trigram index for product names (most important for fuzzy search)
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_products_name_trigram
ON products USING gin(name gin_trgm_ops);

-- Trigram index for normalized names (handles diacritics)
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_products_normalized_trigram
ON products USING gin(normalized_name gin_trgm_ops);

-- Trigram index for brand search with fuzzy matching
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_products_brand_trigram
ON products USING gin(brand gin_trgm_ops)
WHERE brand IS NOT NULL AND brand != '';

-- Composite trigram index for name + brand combined search
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_products_name_brand_trigram
ON products USING gin((name || ' ' || COALESCE(brand, '')) gin_trgm_ops);

-- Trigram index for category search
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_products_category_trigram
ON products USING gin(category gin_trgm_ops)
WHERE category IS NOT NULL AND category != '';

-- Advanced similarity search function with Lithuanian character handling
CREATE OR REPLACE FUNCTION fuzzy_search_products(
    search_query TEXT,
    similarity_threshold FLOAT DEFAULT 0.3,
    limit_count INTEGER DEFAULT 50,
    offset_count INTEGER DEFAULT 0,
    store_ids INTEGER[] DEFAULT NULL,
    category_filter TEXT DEFAULT NULL,
    min_price DECIMAL DEFAULT NULL,
    max_price DECIMAL DEFAULT NULL
)
RETURNS TABLE(
    product_id INTEGER,
    name TEXT,
    brand TEXT,
    category TEXT,
    current_price DECIMAL,
    store_id INTEGER,
    flyer_id INTEGER,
    name_similarity FLOAT,
    brand_similarity FLOAT,
    combined_similarity FLOAT
) AS $$
DECLARE
    normalized_query TEXT;
BEGIN
    -- Normalize the search query for Lithuanian characters
    normalized_query := normalize_lithuanian_text(search_query);

    RETURN QUERY
    SELECT
        p.id,
        p.name,
        p.brand,
        p.category,
        p.current_price,
        p.store_id,
        p.flyer_id,
        similarity(p.name, search_query) as name_similarity,
        COALESCE(similarity(p.brand, search_query), 0) as brand_similarity,
        -- Combined similarity: prioritize name matches, boost with brand matches
        (
            similarity(p.name, search_query) * 0.7 +
            COALESCE(similarity(p.brand, search_query), 0) * 0.3 +
            -- Add bonus for normalized matching (handles diacritics)
            similarity(p.normalized_name, normalized_query) * 0.2
        ) as combined_similarity
    FROM products p
    INNER JOIN flyers f ON f.id = p.flyer_id
    WHERE
        -- Primary fuzzy matching conditions
        (
            similarity(p.name, search_query) >= similarity_threshold OR
            similarity(p.normalized_name, normalized_query) >= similarity_threshold OR
            (p.brand IS NOT NULL AND similarity(p.brand, search_query) >= similarity_threshold) OR
            -- Additional fuzzy matching for combined name+brand
            similarity(p.name || ' ' || COALESCE(p.brand, ''), search_query) >= similarity_threshold
        )
        -- Ensure flyer is current and available
        AND f.is_archived = FALSE
        AND f.valid_from <= NOW()
        AND f.valid_to >= NOW()
        AND p.is_available = TRUE
        -- Apply optional filters
        AND (store_ids IS NULL OR p.store_id = ANY(store_ids))
        AND (category_filter IS NULL OR p.category ILIKE ('%' || category_filter || '%'))
        AND (min_price IS NULL OR p.current_price >= min_price)
        AND (max_price IS NULL OR p.current_price <= max_price)
    ORDER BY
        combined_similarity DESC,
        p.current_price ASC,
        p.name ASC
    LIMIT limit_count OFFSET offset_count;
END;
$$ LANGUAGE plpgsql;

-- Enhanced product search function combining FTS and trigram search
CREATE OR REPLACE FUNCTION hybrid_search_products(
    search_query TEXT,
    limit_count INTEGER DEFAULT 50,
    offset_count INTEGER DEFAULT 0,
    store_ids INTEGER[] DEFAULT NULL,
    min_price DECIMAL DEFAULT NULL,
    max_price DECIMAL DEFAULT NULL,
    prefer_fuzzy BOOLEAN DEFAULT FALSE
)
RETURNS TABLE(
    product_id INTEGER,
    name TEXT,
    brand TEXT,
    current_price DECIMAL,
    store_id INTEGER,
    flyer_id INTEGER,
    search_score FLOAT,
    match_type TEXT
) AS $$
DECLARE
    query_vector tsvector;
    normalized_query TEXT;
BEGIN
    query_vector := to_tsvector('lithuanian_search', search_query);
    normalized_query := normalize_lithuanian_text(search_query);

    RETURN QUERY
    WITH fts_results AS (
        SELECT
            p.id,
            p.name,
            p.brand,
            p.current_price,
            p.store_id,
            p.flyer_id,
            ts_rank_cd(p.search_vector, query_vector) * 2.0 as score,
            'fts'::TEXT as match_type
        FROM products p
        INNER JOIN flyers f ON f.id = p.flyer_id
        WHERE
            p.search_vector @@ query_vector
            AND f.is_archived = FALSE
            AND f.valid_from <= NOW()
            AND f.valid_to >= NOW()
            AND p.is_available = TRUE
            AND (store_ids IS NULL OR p.store_id = ANY(store_ids))
            AND (min_price IS NULL OR p.current_price >= min_price)
            AND (max_price IS NULL OR p.current_price <= max_price)
    ),
    trigram_results AS (
        SELECT
            p.id,
            p.name,
            p.brand,
            p.current_price,
            p.store_id,
            p.flyer_id,
            (
                similarity(p.name, search_query) * 0.6 +
                COALESCE(similarity(p.brand, search_query), 0) * 0.2 +
                similarity(p.normalized_name, normalized_query) * 0.2
            ) as score,
            'fuzzy'::TEXT as match_type
        FROM products p
        INNER JOIN flyers f ON f.id = p.flyer_id
        WHERE
            (
                similarity(p.name, search_query) >= 0.3 OR
                similarity(p.normalized_name, normalized_query) >= 0.3 OR
                (p.brand IS NOT NULL AND similarity(p.brand, search_query) >= 0.3)
            )
            AND f.is_archived = FALSE
            AND f.valid_from <= NOW()
            AND f.valid_to >= NOW()
            AND p.is_available = TRUE
            AND (store_ids IS NULL OR p.store_id = ANY(store_ids))
            AND (min_price IS NULL OR p.current_price >= min_price)
            AND (max_price IS NULL OR p.current_price <= max_price)
            -- Exclude results already found by FTS (avoid duplicates)
            AND NOT EXISTS (SELECT 1 FROM fts_results WHERE fts_results.id = p.id)
    ),
    combined_results AS (
        SELECT * FROM fts_results
        UNION ALL
        SELECT * FROM trigram_results
    )
    SELECT
        cr.id,
        cr.name,
        cr.brand,
        cr.current_price,
        cr.store_id,
        cr.flyer_id,
        CASE
            WHEN prefer_fuzzy THEN cr.score * (CASE WHEN cr.match_type = 'fuzzy' THEN 1.2 ELSE 1.0 END)
            ELSE cr.score
        END as search_score,
        cr.match_type
    FROM combined_results cr
    ORDER BY search_score DESC, cr.current_price ASC, cr.name ASC
    LIMIT limit_count OFFSET offset_count;
END;
$$ LANGUAGE plpgsql;

-- Function to find similar products (useful for "customers also viewed")
CREATE OR REPLACE FUNCTION find_similar_products(
    product_id INTEGER,
    limit_count INTEGER DEFAULT 10
)
RETURNS TABLE(
    similar_product_id INTEGER,
    name TEXT,
    brand TEXT,
    current_price DECIMAL,
    store_id INTEGER,
    similarity_score FLOAT
) AS $$
DECLARE
    target_product RECORD;
BEGIN
    -- Get the target product details
    SELECT p.name, p.brand, p.category, p.normalized_name
    INTO target_product
    FROM products p
    WHERE p.id = product_id;

    IF NOT FOUND THEN
        RETURN;
    END IF;

    RETURN QUERY
    SELECT
        p.id,
        p.name,
        p.brand,
        p.current_price,
        p.store_id,
        (
            -- Similar name
            similarity(p.name, target_product.name) * 0.4 +
            -- Similar brand
            CASE
                WHEN p.brand IS NOT NULL AND target_product.brand IS NOT NULL
                THEN similarity(p.brand, target_product.brand) * 0.3
                ELSE 0
            END +
            -- Same category bonus
            CASE
                WHEN p.category = target_product.category THEN 0.3
                ELSE 0
            END
        ) as similarity_score
    FROM products p
    INNER JOIN flyers f ON f.id = p.flyer_id
    WHERE
        p.id != product_id
        AND f.is_archived = FALSE
        AND f.valid_from <= NOW()
        AND f.valid_to >= NOW()
        AND p.is_available = TRUE
        AND (
            similarity(p.name, target_product.name) >= 0.3 OR
            similarity(p.normalized_name, target_product.normalized_name) >= 0.3 OR
            p.category = target_product.category
        )
    ORDER BY similarity_score DESC, p.current_price ASC
    LIMIT limit_count;
END;
$$ LANGUAGE plpgsql;

-- Create function to suggest query corrections for typos
CREATE OR REPLACE FUNCTION suggest_query_corrections(
    search_query TEXT,
    limit_count INTEGER DEFAULT 5
)
RETURNS TABLE(
    suggestion TEXT,
    confidence FLOAT
) AS $$
BEGIN
    RETURN QUERY
    SELECT DISTINCT
        p.name as suggestion,
        similarity(p.name, search_query) as confidence
    FROM products p
    INNER JOIN flyers f ON f.id = p.flyer_id
    WHERE
        similarity(p.name, search_query) BETWEEN 0.25 AND 0.6  -- Potential typos
        AND f.is_archived = FALSE
        AND f.valid_from <= NOW()
        AND f.valid_to >= NOW()
        AND p.is_available = TRUE
        AND LENGTH(p.name) > 3  -- Avoid suggesting very short words
    ORDER BY confidence DESC, suggestion ASC
    LIMIT limit_count;
END;
$$ LANGUAGE plpgsql;

-- Update statistics for better query planning
ANALYZE products;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
-- Remove trigram-specific functions and indexes

DROP FUNCTION IF EXISTS suggest_query_corrections(TEXT, INTEGER);
DROP FUNCTION IF EXISTS find_similar_products(INTEGER, INTEGER);
DROP FUNCTION IF EXISTS hybrid_search_products(TEXT, INTEGER, INTEGER, INTEGER[], DECIMAL, DECIMAL, BOOLEAN);
DROP FUNCTION IF EXISTS fuzzy_search_products(TEXT, FLOAT, INTEGER, INTEGER, INTEGER[], TEXT, DECIMAL, DECIMAL);

-- Drop trigram indexes
DROP INDEX CONCURRENTLY IF EXISTS idx_products_name_trigram;
DROP INDEX CONCURRENTLY IF EXISTS idx_products_normalized_trigram;
DROP INDEX CONCURRENTLY IF EXISTS idx_products_brand_trigram;
DROP INDEX CONCURRENTLY IF EXISTS idx_products_name_brand_trigram;
DROP INDEX CONCURRENTLY IF EXISTS idx_products_category_trigram;

-- Reset similarity threshold to default
RESET pg_trgm.similarity_threshold;

-- +goose StatementEnd