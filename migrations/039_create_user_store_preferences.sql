-- +goose Up
-- User store preferences for preferred/nearby stores feature
CREATE TABLE user_store_preferences (
    id SERIAL PRIMARY KEY,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    store_id INT NOT NULL REFERENCES stores(id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE(user_id, store_id)
);

-- Index for efficient lookups by user
CREATE INDEX idx_user_store_prefs_user ON user_store_preferences(user_id);

-- Index for reverse lookups (find users who prefer a store)
CREATE INDEX idx_user_store_prefs_store ON user_store_preferences(store_id);

-- +goose Down
DROP TABLE IF EXISTS user_store_preferences;
