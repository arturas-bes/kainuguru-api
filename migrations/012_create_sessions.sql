-- +goose Up
-- User session management for JWT and authentication
CREATE TABLE user_sessions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token_hash VARCHAR(255) UNIQUE NOT NULL,
    expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
    ip_address INET,
    user_agent TEXT,
    is_active BOOLEAN DEFAULT TRUE,
    refresh_token_hash VARCHAR(255),
    refresh_expires_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    last_used_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),

    -- Session metadata
    device_type VARCHAR(50), -- 'web', 'mobile', 'api'
    browser_info JSONB DEFAULT '{}',
    location_info JSONB DEFAULT '{}',

    CONSTRAINT sessions_expiry_check CHECK (expires_at > created_at),
    CONSTRAINT sessions_refresh_expiry_check CHECK (
        refresh_expires_at IS NULL OR refresh_expires_at > expires_at
    ),
    CONSTRAINT sessions_device_type_check CHECK (
        device_type IN ('web', 'mobile', 'api', 'unknown')
    )
);

-- Indexes for performance and cleanup
CREATE INDEX idx_sessions_user ON user_sessions(user_id);
CREATE INDEX idx_sessions_token ON user_sessions(token_hash);
CREATE INDEX idx_sessions_refresh_token ON user_sessions(refresh_token_hash) WHERE refresh_token_hash IS NOT NULL;
CREATE INDEX idx_sessions_expires ON user_sessions(expires_at);
CREATE INDEX idx_sessions_active ON user_sessions(user_id, is_active) WHERE is_active = TRUE;
CREATE INDEX idx_sessions_cleanup ON user_sessions(expires_at, is_active) WHERE is_active = TRUE;

-- Partial index for active sessions
CREATE INDEX idx_sessions_active_user ON user_sessions(user_id, last_used_at)
    WHERE is_active = TRUE AND expires_at > NOW();

-- +goose Down
DROP TABLE IF EXISTS user_sessions;