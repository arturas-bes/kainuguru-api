-- +goose Up
-- +goose StatementBegin
-- Comprehensive search indexes for optimal product search performance

-- GIN index for full-text search on search_vector (most important for search)
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_products_search_vector_gin
ON products USING gin(search_vector);

-- GIN index for trigram search on normalized names
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_products_normalized_name_gin
ON products USING gin(normalized_name gin_trgm_ops);

-- GIN index for trigram search on brand names
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_products_brand_gin
ON products USING gin(brand gin_trgm_ops)
WHERE brand IS NOT NULL;

-- Composite index for common filter combinations
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_products_search_filters
ON products (store_id, is_available, is_on_sale, current_price)
WHERE is_available = TRUE;

-- Index for current products (within validity period)
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_products_current
ON products (valid_from, valid_to, is_available)
WHERE is_available = TRUE AND valid_to >= NOW();

-- Index for price range queries
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_products_price_range
ON products (current_price, store_id, is_available)
WHERE is_available = TRUE AND current_price > 0;

-- Index for category-based filtering
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_products_category
ON products (category, store_id, current_price)
WHERE category IS NOT NULL AND is_available = TRUE;

-- Composite index for flyer-based queries (used in joins)
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_products_flyer_search
ON products (flyer_id, is_available, current_price)
WHERE is_available = TRUE;

-- Index for sale products specifically
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_products_on_sale
ON products (is_on_sale, discount_percent, current_price, store_id)
WHERE is_on_sale = TRUE AND is_available = TRUE;

-- Partial index for premium/expensive products (for performance on high-value searches)
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_products_premium
ON products (current_price DESC, brand, store_id)
WHERE current_price >= 10.00 AND is_available = TRUE;

-- Index optimized for product name searches (covers exact matches)
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_products_name_lower
ON products (LOWER(name), store_id, current_price)
WHERE is_available = TRUE;

-- Composite index for time-based queries (recent products first)
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_products_recent
ON products (created_at DESC, store_id, is_available)
WHERE is_available = TRUE;

-- Index for unit price comparisons
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_products_unit_comparison
ON products (unit_type, current_price, store_id)
WHERE unit_type IS NOT NULL AND is_available = TRUE;

-- Statistics update to help query planner
ANALYZE products;

-- Create index usage monitoring view (for performance tuning)
CREATE OR REPLACE VIEW product_search_index_usage AS
SELECT
    schemaname,
    tablename,
    indexname,
    idx_tup_read,
    idx_tup_fetch,
    idx_scan,
    ROUND(
        CASE
            WHEN idx_scan > 0
            THEN (idx_tup_read::float / idx_scan)::numeric
            ELSE 0
        END, 2
    ) as avg_tuples_per_scan
FROM pg_stat_user_indexes
WHERE tablename = 'products'
    AND schemaname = current_schema()
ORDER BY idx_scan DESC, idx_tup_read DESC;

-- Create function to analyze search query performance
CREATE OR REPLACE FUNCTION analyze_search_query(query_text TEXT)
RETURNS TABLE(
    query_plan TEXT,
    execution_time_ms NUMERIC,
    index_usage TEXT
) AS $$
DECLARE
    plan_text TEXT;
    start_time TIMESTAMP;
    end_time TIMESTAMP;
BEGIN
    -- Enable timing
    start_time := clock_timestamp();

    -- Execute EXPLAIN ANALYZE for the search query
    EXECUTE format('EXPLAIN (ANALYZE, BUFFERS, FORMAT TEXT)
                   SELECT * FROM search_products_ranked(%L, 10)', query_text)
    INTO plan_text;

    end_time := clock_timestamp();

    RETURN QUERY SELECT
        plan_text,
        EXTRACT(MILLISECONDS FROM (end_time - start_time)),
        'Check pg_stat_user_indexes for detailed usage'::TEXT;
END;
$$ LANGUAGE plpgsql;

-- Create materialized view for search suggestions (updated periodically)
CREATE MATERIALIZED VIEW IF NOT EXISTS popular_product_searches AS
SELECT
    normalized_name,
    brand,
    category,
    COUNT(*) as frequency,
    MIN(current_price) as min_price,
    MAX(current_price) as max_price,
    ARRAY_AGG(DISTINCT store_id) as available_stores
FROM products p
INNER JOIN flyers f ON f.id = p.flyer_id
WHERE
    f.is_archived = FALSE
    AND f.valid_from <= NOW()
    AND f.valid_to >= NOW()
    AND p.is_available = TRUE
    AND LENGTH(p.normalized_name) > 2
GROUP BY normalized_name, brand, category
HAVING COUNT(*) >= 2  -- Only include products available in multiple instances
ORDER BY frequency DESC, normalized_name;

-- Create unique index on the materialized view for fast lookups
CREATE UNIQUE INDEX IF NOT EXISTS idx_popular_searches_unique
ON popular_product_searches (normalized_name, COALESCE(brand, ''), COALESCE(category, ''));

-- Create GIN index for suggestion search
CREATE INDEX IF NOT EXISTS idx_popular_searches_gin
ON popular_product_searches USING gin(normalized_name gin_trgm_ops);

-- Function to refresh search suggestions (call this periodically)
CREATE OR REPLACE FUNCTION refresh_search_suggestions()
RETURNS VOID AS $$
BEGIN
    REFRESH MATERIALIZED VIEW CONCURRENTLY popular_product_searches;
    ANALYZE popular_product_searches;
END;
$$ LANGUAGE plpgsql;

-- Create function to get search suggestions
CREATE OR REPLACE FUNCTION get_search_suggestions(
    partial_query TEXT,
    limit_count INTEGER DEFAULT 10
)
RETURNS TABLE(
    suggestion TEXT,
    frequency BIGINT,
    min_price NUMERIC,
    max_price NUMERIC,
    store_count INTEGER
) AS $$
BEGIN
    RETURN QUERY
    SELECT
        pps.normalized_name as suggestion,
        pps.frequency,
        pps.min_price,
        pps.max_price,
        array_length(pps.available_stores, 1) as store_count
    FROM popular_product_searches pps
    WHERE
        similarity(pps.normalized_name, normalize_lithuanian_text(partial_query)) > 0.3
        OR pps.normalized_name ILIKE ('%' || normalize_lithuanian_text(partial_query) || '%')
    ORDER BY
        similarity(pps.normalized_name, normalize_lithuanian_text(partial_query)) DESC,
        pps.frequency DESC
    LIMIT limit_count;
END;
$$ LANGUAGE plpgsql;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
-- Remove search indexes and related objects

DROP FUNCTION IF EXISTS get_search_suggestions(TEXT, INTEGER);
DROP FUNCTION IF EXISTS refresh_search_suggestions();
DROP MATERIALIZED VIEW IF EXISTS popular_product_searches;
DROP FUNCTION IF EXISTS analyze_search_query(TEXT);
DROP VIEW IF EXISTS product_search_index_usage;

-- Drop indexes (these will be dropped concurrently to avoid blocking)
DROP INDEX CONCURRENTLY IF EXISTS idx_products_search_vector_gin;
DROP INDEX CONCURRENTLY IF EXISTS idx_products_normalized_name_gin;
DROP INDEX CONCURRENTLY IF EXISTS idx_products_brand_gin;
DROP INDEX CONCURRENTLY IF EXISTS idx_products_search_filters;
DROP INDEX CONCURRENTLY IF EXISTS idx_products_current;
DROP INDEX CONCURRENTLY IF EXISTS idx_products_price_range;
DROP INDEX CONCURRENTLY IF EXISTS idx_products_category;
DROP INDEX CONCURRENTLY IF EXISTS idx_products_flyer_search;
DROP INDEX CONCURRENTLY IF EXISTS idx_products_on_sale;
DROP INDEX CONCURRENTLY IF EXISTS idx_products_premium;
DROP INDEX CONCURRENTLY IF EXISTS idx_products_name_lower;
DROP INDEX CONCURRENTLY IF EXISTS idx_products_recent;
DROP INDEX CONCURRENTLY IF EXISTS idx_products_unit_comparison;
DROP INDEX CONCURRENTLY IF EXISTS idx_popular_searches_unique;
DROP INDEX CONCURRENTLY IF EXISTS idx_popular_searches_gin;

-- +goose StatementEnd