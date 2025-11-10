-- +goose Up
-- Login attempts tracking for security and rate limiting
CREATE TABLE login_attempts (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email VARCHAR(255) NOT NULL,
    success BOOLEAN NOT NULL,
    ip_address VARCHAR(45),  -- IPv6 compatible
    user_agent TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW() NOT NULL
);

-- Indexes for performance
CREATE INDEX idx_login_attempts_email ON login_attempts(email);
CREATE INDEX idx_login_attempts_created_at ON login_attempts(created_at);
CREATE INDEX idx_login_attempts_email_created ON login_attempts(email, created_at) WHERE success = FALSE;

-- Auto-cleanup old attempts (keep last 30 days)
-- This can be run as a periodic job
CREATE OR REPLACE FUNCTION cleanup_old_login_attempts() RETURNS void AS $$
BEGIN
    DELETE FROM login_attempts WHERE created_at < NOW() - INTERVAL '30 days';
END;
$$ LANGUAGE plpgsql;

-- +goose Down
DROP FUNCTION IF EXISTS cleanup_old_login_attempts();
DROP TABLE IF EXISTS login_attempts;
