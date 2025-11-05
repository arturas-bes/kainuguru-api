-- +goose Up
-- +goose StatementBegin
-- Enhanced partition function with automatic cleanup
CREATE OR REPLACE FUNCTION manage_product_partitions()
RETURNS void AS $$
DECLARE
    start_date date;
    end_date date;
    partition_name text;
    cleanup_date date;
    old_partition_name text;
BEGIN
    -- Create next week partition if needed
    start_date := date_trunc('week', CURRENT_DATE + INTERVAL '7 days')::date;
    end_date := start_date + INTERVAL '7 days';
    partition_name := 'products_' || to_char(start_date, 'YYYY_WW');

    IF NOT EXISTS (
        SELECT 1 FROM pg_class WHERE relname = partition_name
    ) THEN
        EXECUTE format(
            'CREATE TABLE %I PARTITION OF products FOR VALUES FROM (%L) TO (%L)',
            partition_name, start_date, end_date
        );

        -- Create indexes on new partition
        EXECUTE format('CREATE INDEX ON %I (flyer_id)', partition_name);
        EXECUTE format('CREATE INDEX ON %I (store_id)', partition_name);
        EXECUTE format('CREATE INDEX ON %I (product_master_id)', partition_name);
        EXECUTE format('CREATE INDEX ON %I USING gin(search_vector)', partition_name);
        EXECUTE format('CREATE INDEX ON %I USING gin(normalized_name gin_trgm_ops)', partition_name);

        RAISE NOTICE 'Created partition %', partition_name;
    END IF;

    -- Clean up old partitions (older than 8 weeks)
    cleanup_date := date_trunc('week', CURRENT_DATE - INTERVAL '8 weeks')::date;
    old_partition_name := 'products_' || to_char(cleanup_date, 'YYYY_WW');

    IF EXISTS (
        SELECT 1 FROM pg_class WHERE relname = old_partition_name
    ) THEN
        EXECUTE format('DROP TABLE IF EXISTS %I', old_partition_name);
        RAISE NOTICE 'Dropped old partition %', old_partition_name;
    END IF;
END;
$$ LANGUAGE plpgsql;

-- Function to get current week partition name
CREATE OR REPLACE FUNCTION get_current_week_partition()
RETURNS text AS $$
BEGIN
    RETURN 'products_' || to_char(date_trunc('week', CURRENT_DATE), 'YYYY_WW');
END;
$$ LANGUAGE plpgsql;

-- Function to ensure partition exists for a given date
CREATE OR REPLACE FUNCTION ensure_partition_for_date(target_date DATE)
RETURNS text AS $$
DECLARE
    start_date date;
    end_date date;
    partition_name text;
BEGIN
    start_date := date_trunc('week', target_date)::date;
    end_date := start_date + INTERVAL '7 days';
    partition_name := 'products_' || to_char(start_date, 'YYYY_WW');

    IF NOT EXISTS (
        SELECT 1 FROM pg_class WHERE relname = partition_name
    ) THEN
        EXECUTE format(
            'CREATE TABLE %I PARTITION OF products FOR VALUES FROM (%L) TO (%L)',
            partition_name, start_date, end_date
        );

        -- Create indexes on new partition
        EXECUTE format('CREATE INDEX ON %I (flyer_id)', partition_name);
        EXECUTE format('CREATE INDEX ON %I (store_id)', partition_name);
        EXECUTE format('CREATE INDEX ON %I (product_master_id)', partition_name);
        EXECUTE format('CREATE INDEX ON %I USING gin(search_vector)', partition_name);
        EXECUTE format('CREATE INDEX ON %I USING gin(normalized_name gin_trgm_ops)', partition_name);
    END IF;

    RETURN partition_name;
END;
$$ LANGUAGE plpgsql;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP FUNCTION IF EXISTS ensure_partition_for_date(DATE);
DROP FUNCTION IF EXISTS get_current_week_partition();
DROP FUNCTION IF EXISTS manage_product_partitions();
-- +goose StatementEnd