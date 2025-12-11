-- +goose Up
-- +goose StatementBegin

-- Drop existing functions to recreate with new signature
DROP FUNCTION IF EXISTS fuzzy_search_products(TEXT, FLOAT, INTEGER, INTEGER, INTEGER[], TEXT, DECIMAL, DECIMAL, BOOLEAN, TEXT[]);
DROP FUNCTION IF EXISTS hybrid_search_products(TEXT, INTEGER, INTEGER, INTEGER[], DECIMAL, DECIMAL, BOOLEAN, TEXT, BOOLEAN, TEXT[]);

-- Recreate fuzzy_search_products with flyer_ids parameter
CREATE OR REPLACE FUNCTION fuzzy_search_products(
    search_query TEXT,
    similarity_threshold FLOAT DEFAULT 0.3,
    limit_count INTEGER DEFAULT 50,
    offset_count INTEGER DEFAULT 0,
    store_ids INTEGER[] DEFAULT NULL,
    category_filter TEXT DEFAULT NULL,
    min_price DECIMAL DEFAULT NULL,
    max_price DECIMAL DEFAULT NULL,
    on_sale_only BOOLEAN DEFAULT FALSE,
    tag_filters TEXT[] DEFAULT NULL,
    flyer_ids INTEGER[] DEFAULT NULL
)
RETURNS TABLE(
    product_id BIGINT,
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
        p.name::TEXT,
        p.brand::TEXT,
        p.category::TEXT,
        p.current_price,
        p.store_id,
        p.flyer_id,
        similarity(p.name, search_query)::FLOAT as name_similarity,
        COALESCE(similarity(p.brand, search_query), 0)::FLOAT as brand_similarity,
        (
            similarity(p.name, search_query) * 0.7 +
            COALESCE(similarity(p.brand, search_query), 0) * 0.3 +
            similarity(p.normalized_name, normalized_query) * 0.2
        )::FLOAT as combined_similarity
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
        AND (flyer_ids IS NULL OR p.flyer_id = ANY(flyer_ids))
        AND (min_price IS NULL OR p.current_price >= min_price)
        AND (max_price IS NULL OR p.current_price <= max_price)
        AND (category_filter IS NULL OR p.category ILIKE ('%' || category_filter || '%'))
        AND (NOT on_sale_only OR p.is_on_sale = TRUE)
        AND (tag_filters IS NULL OR p.tags && tag_filters)
    ORDER BY combined_similarity DESC, p.current_price ASC
    LIMIT limit_count OFFSET offset_count;
END;
$$ LANGUAGE plpgsql;

-- Recreate hybrid_search_products with flyer_ids parameter
CREATE OR REPLACE FUNCTION hybrid_search_products(
    search_query TEXT,
    limit_count INTEGER DEFAULT 50,
    offset_count INTEGER DEFAULT 0,
    store_ids INTEGER[] DEFAULT NULL,
    min_price DECIMAL DEFAULT NULL,
    max_price DECIMAL DEFAULT NULL,
    prefer_fuzzy BOOLEAN DEFAULT FALSE,
    category_filter TEXT DEFAULT NULL,
    on_sale_only BOOLEAN DEFAULT FALSE,
    tag_filters TEXT[] DEFAULT NULL,
    flyer_ids INTEGER[] DEFAULT NULL
)
RETURNS TABLE(
    product_id BIGINT,
    name TEXT,
    brand TEXT,
    current_price DECIMAL,
    store_id INTEGER,
    flyer_id INTEGER,
    search_score FLOAT,
    match_type TEXT
) AS $$
DECLARE
    query_tsquery tsquery;
    normalized_query TEXT;
BEGIN
    -- Convert search query to tsquery properly
    query_tsquery := plainto_tsquery('lithuanian', search_query);
    normalized_query := normalize_lithuanian_text(search_query);

    RETURN QUERY
    WITH fts_results AS (
        SELECT
            p.id,
            p.name::TEXT,
            p.brand::TEXT,
            p.current_price,
            p.store_id,
            p.flyer_id,
            ts_rank_cd(p.search_vector, query_tsquery) * 2.0 as score,
            'fts'::TEXT as match_type
        FROM products p
        INNER JOIN flyers f ON f.id = p.flyer_id
        WHERE
            p.search_vector @@ query_tsquery
            AND f.is_archived = FALSE
            AND f.valid_from <= NOW()
            AND f.valid_to >= NOW()
            AND p.is_available = TRUE
            AND (store_ids IS NULL OR p.store_id = ANY(store_ids))
            AND (flyer_ids IS NULL OR p.flyer_id = ANY(flyer_ids))
            AND (min_price IS NULL OR p.current_price >= min_price)
            AND (max_price IS NULL OR p.current_price <= max_price)
            AND (category_filter IS NULL OR p.category ILIKE ('%' || category_filter || '%'))
            AND (NOT on_sale_only OR p.is_on_sale = TRUE)
            AND (tag_filters IS NULL OR p.tags && tag_filters)
    ),
    trigram_results AS (
        SELECT
            p.id,
            p.name::TEXT,
            p.brand::TEXT,
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
            AND (flyer_ids IS NULL OR p.flyer_id = ANY(flyer_ids))
            AND (min_price IS NULL OR p.current_price >= min_price)
            AND (max_price IS NULL OR p.current_price <= max_price)
            AND (category_filter IS NULL OR p.category ILIKE ('%' || category_filter || '%'))
            AND (NOT on_sale_only OR p.is_on_sale = TRUE)
            AND (tag_filters IS NULL OR p.tags && tag_filters)
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

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

-- Restore original functions without flyer_ids parameter
DROP FUNCTION IF EXISTS fuzzy_search_products(TEXT, FLOAT, INTEGER, INTEGER, INTEGER[], TEXT, DECIMAL, DECIMAL, BOOLEAN, TEXT[], INTEGER[]);
DROP FUNCTION IF EXISTS hybrid_search_products(TEXT, INTEGER, INTEGER, INTEGER[], DECIMAL, DECIMAL, BOOLEAN, TEXT, BOOLEAN, TEXT[], INTEGER[]);

-- Recreate fuzzy_search_products without flyer_ids
CREATE OR REPLACE FUNCTION fuzzy_search_products(
    search_query TEXT,
    similarity_threshold FLOAT DEFAULT 0.3,
    limit_count INTEGER DEFAULT 50,
    offset_count INTEGER DEFAULT 0,
    store_ids INTEGER[] DEFAULT NULL,
    category_filter TEXT DEFAULT NULL,
    min_price DECIMAL DEFAULT NULL,
    max_price DECIMAL DEFAULT NULL,
    on_sale_only BOOLEAN DEFAULT FALSE,
    tag_filters TEXT[] DEFAULT NULL
)
RETURNS TABLE(
    product_id BIGINT,
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
        p.name::TEXT,
        p.brand::TEXT,
        p.category::TEXT,
        p.current_price,
        p.store_id,
        p.flyer_id,
        similarity(p.name, search_query)::FLOAT as name_similarity,
        COALESCE(similarity(p.brand, search_query), 0)::FLOAT as brand_similarity,
        (
            similarity(p.name, search_query) * 0.7 +
            COALESCE(similarity(p.brand, search_query), 0) * 0.3 +
            similarity(p.normalized_name, normalized_query) * 0.2
        )::FLOAT as combined_similarity
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
        AND (min_price IS NULL OR p.current_price >= min_price)
        AND (max_price IS NULL OR p.current_price <= max_price)
        AND (category_filter IS NULL OR p.category ILIKE ('%' || category_filter || '%'))
        AND (NOT on_sale_only OR p.is_on_sale = TRUE)
        AND (tag_filters IS NULL OR p.tags && tag_filters)
    ORDER BY combined_similarity DESC, p.current_price ASC
    LIMIT limit_count OFFSET offset_count;
END;
$$ LANGUAGE plpgsql;

-- Recreate hybrid_search_products without flyer_ids
CREATE OR REPLACE FUNCTION hybrid_search_products(
    search_query TEXT,
    limit_count INTEGER DEFAULT 50,
    offset_count INTEGER DEFAULT 0,
    store_ids INTEGER[] DEFAULT NULL,
    min_price DECIMAL DEFAULT NULL,
    max_price DECIMAL DEFAULT NULL,
    prefer_fuzzy BOOLEAN DEFAULT FALSE,
    category_filter TEXT DEFAULT NULL,
    on_sale_only BOOLEAN DEFAULT FALSE,
    tag_filters TEXT[] DEFAULT NULL
)
RETURNS TABLE(
    product_id BIGINT,
    name TEXT,
    brand TEXT,
    current_price DECIMAL,
    store_id INTEGER,
    flyer_id INTEGER,
    search_score FLOAT,
    match_type TEXT
) AS $$
DECLARE
    query_tsquery tsquery;
    normalized_query TEXT;
BEGIN
    query_tsquery := plainto_tsquery('lithuanian', search_query);
    normalized_query := normalize_lithuanian_text(search_query);

    RETURN QUERY
    WITH fts_results AS (
        SELECT
            p.id,
            p.name::TEXT,
            p.brand::TEXT,
            p.current_price,
            p.store_id,
            p.flyer_id,
            ts_rank_cd(p.search_vector, query_tsquery) * 2.0 as score,
            'fts'::TEXT as match_type
        FROM products p
        INNER JOIN flyers f ON f.id = p.flyer_id
        WHERE
            p.search_vector @@ query_tsquery
            AND f.is_archived = FALSE
            AND f.valid_from <= NOW()
            AND f.valid_to >= NOW()
            AND p.is_available = TRUE
            AND (store_ids IS NULL OR p.store_id = ANY(store_ids))
            AND (min_price IS NULL OR p.current_price >= min_price)
            AND (max_price IS NULL OR p.current_price <= max_price)
            AND (category_filter IS NULL OR p.category ILIKE ('%' || category_filter || '%'))
            AND (NOT on_sale_only OR p.is_on_sale = TRUE)
            AND (tag_filters IS NULL OR p.tags && tag_filters)
    ),
    trigram_results AS (
        SELECT
            p.id,
            p.name::TEXT,
            p.brand::TEXT,
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
            AND (tag_filters IS NULL OR p.tags && tag_filters)
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

-- +goose StatementEnd
