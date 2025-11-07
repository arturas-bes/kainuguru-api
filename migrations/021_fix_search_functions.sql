-- +goose Up
-- +goose StatementBegin
-- Fix search functionality issues

-- 1. Create popular_product_searches table for search suggestions
CREATE TABLE IF NOT EXISTS popular_product_searches (
    id BIGSERIAL PRIMARY KEY,
    search_term VARCHAR(255) NOT NULL,
    normalized_term VARCHAR(255) NOT NULL,
    search_count INTEGER DEFAULT 1,
    result_count INTEGER DEFAULT 0,
    avg_click_position DECIMAL(5,2),
    last_searched_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Create indexes for efficient lookups
CREATE INDEX idx_popular_searches_term ON popular_product_searches(search_term);
CREATE INDEX idx_popular_searches_normalized ON popular_product_searches(normalized_term);
CREATE INDEX idx_popular_searches_count ON popular_product_searches(search_count DESC);
CREATE INDEX idx_popular_searches_last_searched ON popular_product_searches(last_searched_at DESC);

-- Create trigram index for fuzzy matching on search terms
CREATE INDEX idx_popular_searches_term_trigram
ON popular_product_searches USING gin(search_term gin_trgm_ops);

-- 2. Fix hybrid_search_products function - change 'lithuanian_search' to 'lithuanian'
--    and add missing category_filter and on_sale_only parameters
DROP FUNCTION IF EXISTS hybrid_search_products(TEXT, INTEGER, INTEGER, INTEGER[], DECIMAL, DECIMAL, BOOLEAN);

CREATE OR REPLACE FUNCTION hybrid_search_products(
    search_query TEXT,
    limit_count INTEGER DEFAULT 50,
    offset_count INTEGER DEFAULT 0,
    store_ids INTEGER[] DEFAULT NULL,
    min_price DECIMAL DEFAULT NULL,
    max_price DECIMAL DEFAULT NULL,
    prefer_fuzzy BOOLEAN DEFAULT FALSE,
    category_filter TEXT DEFAULT NULL,
    on_sale_only BOOLEAN DEFAULT FALSE
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
    -- Fix: Use 'lithuanian' instead of 'lithuanian_search'
    query_vector := to_tsvector('lithuanian', search_query);
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
            AND (category_filter IS NULL OR p.category ILIKE ('%' || category_filter || '%'))
            AND (NOT on_sale_only OR p.is_on_sale = TRUE)
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
            AND (category_filter IS NULL OR p.category ILIKE ('%' || category_filter || '%'))
            AND (NOT on_sale_only OR p.is_on_sale = TRUE)
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

-- 3. Fix fuzzy_search_products function - add on_sale_only parameter
DROP FUNCTION IF EXISTS fuzzy_search_products(TEXT, FLOAT, INTEGER, INTEGER, INTEGER[], TEXT, DECIMAL, DECIMAL);

CREATE OR REPLACE FUNCTION fuzzy_search_products(
    search_query TEXT,
    similarity_threshold FLOAT DEFAULT 0.3,
    limit_count INTEGER DEFAULT 50,
    offset_count INTEGER DEFAULT 0,
    store_ids INTEGER[] DEFAULT NULL,
    category_filter TEXT DEFAULT NULL,
    min_price DECIMAL DEFAULT NULL,
    max_price DECIMAL DEFAULT NULL,
    on_sale_only BOOLEAN DEFAULT FALSE
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
        AND (NOT on_sale_only OR p.is_on_sale = TRUE)
    ORDER BY
        combined_similarity DESC,
        p.current_price ASC,
        p.name ASC
    LIMIT limit_count OFFSET offset_count;
END;
$$ LANGUAGE plpgsql;

-- 4. Create get_search_suggestions function
CREATE OR REPLACE FUNCTION get_search_suggestions(
    partial_query TEXT,
    limit_count INTEGER DEFAULT 10
)
RETURNS TABLE(
    suggestion TEXT,
    frequency BIGINT,
    min_price DECIMAL,
    max_price DECIMAL,
    store_count INTEGER
) AS $$
DECLARE
    normalized_query TEXT;
BEGIN
    normalized_query := normalize_lithuanian_text(partial_query);

    RETURN QUERY
    WITH product_suggestions AS (
        -- Get suggestions from actual products
        SELECT DISTINCT
            p.name as suggestion_text,
            COUNT(*) OVER (PARTITION BY p.name) as freq,
            MIN(p.current_price) OVER (PARTITION BY p.name) as min_pr,
            MAX(p.current_price) OVER (PARTITION BY p.name) as max_pr,
            COUNT(DISTINCT p.store_id) OVER (PARTITION BY p.name) as store_cnt
        FROM products p
        INNER JOIN flyers f ON f.id = p.flyer_id
        WHERE
            (
                p.name ILIKE (partial_query || '%') OR
                p.normalized_name ILIKE (normalized_query || '%') OR
                similarity(p.name, partial_query) >= 0.3
            )
            AND f.is_archived = FALSE
            AND f.valid_from <= NOW()
            AND f.valid_to >= NOW()
            AND p.is_available = TRUE
            AND LENGTH(p.name) > 2
        LIMIT 50
    ),
    popular_suggestions AS (
        -- Get suggestions from popular searches
        SELECT
            ps.search_term as suggestion_text,
            ps.search_count as freq,
            0::DECIMAL as min_pr,
            0::DECIMAL as max_pr,
            0 as store_cnt
        FROM popular_product_searches ps
        WHERE
            (
                ps.search_term ILIKE (partial_query || '%') OR
                ps.normalized_term ILIKE (normalized_query || '%') OR
                similarity(ps.search_term, partial_query) >= 0.3
            )
        LIMIT 20
    ),
    combined_suggestions AS (
        SELECT * FROM product_suggestions
        UNION
        SELECT * FROM popular_suggestions
    )
    SELECT
        cs.suggestion_text,
        cs.freq::BIGINT,
        cs.min_pr,
        cs.max_pr,
        cs.store_cnt
    FROM combined_suggestions cs
    ORDER BY cs.freq DESC, cs.suggestion_text ASC
    LIMIT limit_count;
END;
$$ LANGUAGE plpgsql;

-- 5. Create refresh_search_suggestions function
CREATE OR REPLACE FUNCTION refresh_search_suggestions()
RETURNS void AS $$
BEGIN
    -- This function can be used to rebuild the popular_product_searches table
    -- from search analytics or other sources. For now, it's a placeholder.
    -- In production, this would analyze search logs and update the table.

    -- Clean up old entries (older than 90 days)
    DELETE FROM popular_product_searches
    WHERE last_searched_at < NOW() - INTERVAL '90 days';

    -- Update normalized terms for all entries
    UPDATE popular_product_searches
    SET normalized_term = normalize_lithuanian_text(search_term)
    WHERE normalized_term IS NULL OR normalized_term = '';

    -- Vacuum the table to reclaim space
    -- Note: VACUUM cannot run inside a function, so we just update stats
    ANALYZE popular_product_searches;
END;
$$ LANGUAGE plpgsql;

-- 6. Create function to track search queries
CREATE OR REPLACE FUNCTION track_search_query(
    query_text TEXT,
    result_cnt INTEGER DEFAULT 0
)
RETURNS void AS $$
DECLARE
    normalized_term TEXT;
BEGIN
    normalized_term := normalize_lithuanian_text(query_text);

    -- Insert or update search term
    INSERT INTO popular_product_searches (
        search_term,
        normalized_term,
        search_count,
        result_count,
        last_searched_at
    ) VALUES (
        query_text,
        normalized_term,
        1,
        result_cnt,
        NOW()
    )
    ON CONFLICT (search_term) DO UPDATE SET
        search_count = popular_product_searches.search_count + 1,
        result_count = (popular_product_searches.result_count + EXCLUDED.result_count) / 2,
        last_searched_at = NOW(),
        updated_at = NOW();
END;
$$ LANGUAGE plpgsql;

-- Add unique constraint on search_term
ALTER TABLE popular_product_searches
ADD CONSTRAINT unique_search_term UNIQUE (search_term);

-- Update statistics
ANALYZE popular_product_searches;
ANALYZE products;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

-- Drop new functions
DROP FUNCTION IF EXISTS track_search_query(TEXT, INTEGER);
DROP FUNCTION IF EXISTS refresh_search_suggestions();
DROP FUNCTION IF EXISTS get_search_suggestions(TEXT, INTEGER);

-- Restore old function signatures
DROP FUNCTION IF EXISTS fuzzy_search_products(TEXT, FLOAT, INTEGER, INTEGER, INTEGER[], TEXT, DECIMAL, DECIMAL, BOOLEAN);

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
        (
            similarity(p.name, search_query) * 0.7 +
            COALESCE(similarity(p.brand, search_query), 0) * 0.3 +
            similarity(p.normalized_name, normalized_query) * 0.2
        ) as combined_similarity
    FROM products p
    INNER JOIN flyers f ON f.id = p.flyer_id
    WHERE
        (
            similarity(p.name, search_query) >= similarity_threshold OR
            similarity(p.normalized_name, normalized_query) >= similarity_threshold OR
            (p.brand IS NOT NULL AND similarity(p.brand, search_query) >= similarity_threshold) OR
            similarity(p.name || ' ' || COALESCE(p.brand, ''), search_query) >= similarity_threshold
        )
        AND f.is_archived = FALSE
        AND f.valid_from <= NOW()
        AND f.valid_to >= NOW()
        AND p.is_available = TRUE
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

DROP FUNCTION IF EXISTS hybrid_search_products(TEXT, INTEGER, INTEGER, INTEGER[], DECIMAL, DECIMAL, BOOLEAN, TEXT, BOOLEAN);

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

-- Drop table
DROP TABLE IF EXISTS popular_product_searches;

-- +goose StatementEnd
