-- +goose Up
-- +goose StatementBegin

-- Fix fuzzy_search_products to cast similarity() results to double precision
DROP FUNCTION IF EXISTS fuzzy_search_products(TEXT, FLOAT, INTEGER, INTEGER, INTEGER[], TEXT, DECIMAL, DECIMAL, BOOLEAN);

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
    product_id BIGINT,
    name TEXT,
    brand TEXT,
    category TEXT,
    current_price DECIMAL,
    store_id INTEGER,
    flyer_id INTEGER,
    name_similarity DOUBLE PRECISION,
    brand_similarity DOUBLE PRECISION,
    combined_similarity DOUBLE PRECISION
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
        similarity(p.name, search_query)::DOUBLE PRECISION as name_similarity,
        COALESCE(similarity(p.brand, search_query), 0)::DOUBLE PRECISION as brand_similarity,
        (
            similarity(p.name, search_query) * 0.7 +
            COALESCE(similarity(p.brand, search_query), 0) * 0.3 +
            similarity(p.normalized_name, normalized_query) * 0.2
        )::DOUBLE PRECISION as combined_similarity
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
        AND (NOT on_sale_only OR p.is_on_sale = TRUE)
    ORDER BY
        combined_similarity DESC,
        p.current_price ASC,
        p.name ASC
    LIMIT limit_count OFFSET offset_count;
END;
$$ LANGUAGE plpgsql;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP FUNCTION IF EXISTS fuzzy_search_products(TEXT, FLOAT, INTEGER, INTEGER, INTEGER[], TEXT, DECIMAL, DECIMAL, BOOLEAN);
-- +goose StatementEnd
