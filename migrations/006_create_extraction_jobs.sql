-- +goose Up
-- +goose StatementBegin
CREATE TABLE extraction_jobs (
    id BIGSERIAL PRIMARY KEY,
    job_type VARCHAR(50) NOT NULL, -- 'scrape_flyer', 'extract_page', 'match_products'
    status VARCHAR(50) DEFAULT 'pending',
    priority INTEGER DEFAULT 5,

    -- Job payload
    payload JSONB NOT NULL,

    -- Processing metadata
    attempts INTEGER DEFAULT 0,
    max_attempts INTEGER DEFAULT 3,
    worker_id VARCHAR(100),
    started_at TIMESTAMP WITH TIME ZONE,
    completed_at TIMESTAMP WITH TIME ZONE,

    -- Error tracking
    error_message TEXT,
    error_count INTEGER DEFAULT 0,

    -- Scheduling
    scheduled_for TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    expires_at TIMESTAMP WITH TIME ZONE,

    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX idx_jobs_status_scheduled ON extraction_jobs(status, scheduled_for)
    WHERE status = 'pending';
CREATE INDEX idx_jobs_type ON extraction_jobs(job_type);
CREATE INDEX idx_jobs_expires ON extraction_jobs(expires_at)
    WHERE expires_at IS NOT NULL;
CREATE INDEX idx_jobs_worker ON extraction_jobs(worker_id)
    WHERE worker_id IS NOT NULL;

-- Add trigger to update updated_at timestamp
CREATE TRIGGER update_extraction_jobs_updated_at BEFORE UPDATE ON extraction_jobs
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- Function to get next job with SKIP LOCKED pattern
CREATE OR REPLACE FUNCTION get_next_job(job_types TEXT[], worker_id_param TEXT DEFAULT NULL)
RETURNS SETOF extraction_jobs AS $$
BEGIN
    RETURN QUERY
    UPDATE extraction_jobs
    SET
        status = 'processing',
        worker_id = COALESCE(worker_id_param, 'worker_' || extract(epoch from now())),
        started_at = NOW(),
        attempts = attempts + 1,
        updated_at = NOW()
    WHERE id = (
        SELECT id
        FROM extraction_jobs
        WHERE
            status = 'pending'
            AND (job_types IS NULL OR job_type = ANY(job_types))
            AND scheduled_for <= NOW()
            AND (expires_at IS NULL OR expires_at > NOW())
            AND attempts < max_attempts
        ORDER BY priority DESC, created_at ASC
        FOR UPDATE SKIP LOCKED
        LIMIT 1
    )
    RETURNING *;
END;
$$ LANGUAGE plpgsql;

-- Function to mark job as completed
CREATE OR REPLACE FUNCTION complete_job(job_id BIGINT)
RETURNS BOOLEAN AS $$
BEGIN
    UPDATE extraction_jobs
    SET
        status = 'completed',
        completed_at = NOW(),
        updated_at = NOW()
    WHERE id = job_id;

    RETURN FOUND;
END;
$$ LANGUAGE plpgsql;

-- Function to mark job as failed
CREATE OR REPLACE FUNCTION fail_job(job_id BIGINT, error_msg TEXT)
RETURNS BOOLEAN AS $$
DECLARE
    current_attempts INTEGER;
    max_attempts_allowed INTEGER;
BEGIN
    SELECT attempts, max_attempts INTO current_attempts, max_attempts_allowed
    FROM extraction_jobs WHERE id = job_id;

    IF current_attempts >= max_attempts_allowed THEN
        -- Mark as permanently failed
        UPDATE extraction_jobs
        SET
            status = 'failed',
            error_message = error_msg,
            error_count = error_count + 1,
            completed_at = NOW(),
            updated_at = NOW()
        WHERE id = job_id;
    ELSE
        -- Reset to pending for retry
        UPDATE extraction_jobs
        SET
            status = 'pending',
            worker_id = NULL,
            started_at = NULL,
            error_message = error_msg,
            error_count = error_count + 1,
            scheduled_for = NOW() + INTERVAL '5 minutes',
            updated_at = NOW()
        WHERE id = job_id;
    END IF;

    RETURN FOUND;
END;
$$ LANGUAGE plpgsql;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP FUNCTION IF EXISTS fail_job(BIGINT, TEXT);
DROP FUNCTION IF EXISTS complete_job(BIGINT);
DROP FUNCTION IF EXISTS get_next_job(TEXT[], TEXT);
DROP TRIGGER IF EXISTS update_extraction_jobs_updated_at ON extraction_jobs;
DROP TABLE IF EXISTS extraction_jobs;
-- +goose StatementEnd