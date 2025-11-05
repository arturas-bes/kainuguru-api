-- +goose Up
-- +goose StatementBegin
CREATE TABLE shopping_lists (
    id BIGSERIAL PRIMARY KEY,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,

    -- List details
    name VARCHAR(100) NOT NULL,
    description TEXT,

    -- List settings
    is_default BOOLEAN DEFAULT FALSE,
    is_archived BOOLEAN DEFAULT FALSE,

    -- Sharing settings
    is_public BOOLEAN DEFAULT FALSE,
    share_code VARCHAR(32) UNIQUE, -- For sharing via link

    -- Metadata
    item_count INTEGER DEFAULT 0,
    completed_item_count INTEGER DEFAULT 0,
    estimated_total_price DECIMAL(10,2),

    -- Timestamps
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    last_accessed_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),

    -- Constraints
    CONSTRAINT shopping_lists_name_length CHECK (LENGTH(name) >= 1 AND LENGTH(name) <= 100),
    CONSTRAINT shopping_lists_user_name_unique UNIQUE (user_id, name)
);

-- Indexes for performance
CREATE INDEX idx_shopping_lists_user_id ON shopping_lists(user_id);
CREATE INDEX idx_shopping_lists_user_default ON shopping_lists(user_id, is_default);
CREATE INDEX idx_shopping_lists_share_code ON shopping_lists(share_code) WHERE share_code IS NOT NULL;
CREATE INDEX idx_shopping_lists_updated_at ON shopping_lists(updated_at);

-- Ensure only one default list per user
CREATE UNIQUE INDEX idx_shopping_lists_user_default_unique
ON shopping_lists(user_id)
WHERE is_default = true;

-- Function to update list statistics
CREATE OR REPLACE FUNCTION update_shopping_list_stats()
RETURNS trigger AS $$
BEGIN
    -- Update item counts and estimated total
    UPDATE shopping_lists
    SET
        item_count = (
            SELECT COUNT(*)
            FROM shopping_list_items
            WHERE shopping_list_id = COALESCE(NEW.shopping_list_id, OLD.shopping_list_id)
        ),
        completed_item_count = (
            SELECT COUNT(*)
            FROM shopping_list_items
            WHERE shopping_list_id = COALESCE(NEW.shopping_list_id, OLD.shopping_list_id)
            AND is_checked = true
        ),
        estimated_total_price = (
            SELECT COALESCE(SUM(estimated_price * quantity), 0)
            FROM shopping_list_items
            WHERE shopping_list_id = COALESCE(NEW.shopping_list_id, OLD.shopping_list_id)
        ),
        updated_at = NOW()
    WHERE id = COALESCE(NEW.shopping_list_id, OLD.shopping_list_id);

    RETURN COALESCE(NEW, OLD);
END;
$$ LANGUAGE plpgsql;

-- Function to update updated_at timestamp
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS trigger AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Trigger to auto-update updated_at
CREATE TRIGGER shopping_lists_updated_at_trigger
BEFORE UPDATE ON shopping_lists
FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- Function to ensure only one default list per user
CREATE OR REPLACE FUNCTION ensure_single_default_list()
RETURNS trigger AS $$
BEGIN
    -- If setting this list as default, unset other defaults for this user
    IF NEW.is_default = true THEN
        UPDATE shopping_lists
        SET is_default = false
        WHERE user_id = NEW.user_id
        AND id != NEW.id
        AND is_default = true;
    END IF;

    -- If this is the user's first list, make it default
    IF NOT EXISTS (
        SELECT 1 FROM shopping_lists
        WHERE user_id = NEW.user_id
        AND id != NEW.id
    ) THEN
        NEW.is_default = true;
    END IF;

    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Trigger to manage default list logic
CREATE TRIGGER shopping_lists_default_trigger
BEFORE INSERT OR UPDATE ON shopping_lists
FOR EACH ROW EXECUTE FUNCTION ensure_single_default_list();

-- Function to generate share codes
CREATE OR REPLACE FUNCTION generate_share_code()
RETURNS varchar(32) AS $$
DECLARE
    code varchar(32);
    exists boolean;
BEGIN
    LOOP
        -- Generate random 32-character code
        code := encode(gen_random_bytes(16), 'hex');

        -- Check if it already exists
        SELECT EXISTS(
            SELECT 1 FROM shopping_lists WHERE share_code = code
        ) INTO exists;

        -- If unique, return it
        IF NOT exists THEN
            RETURN code;
        END IF;
    END LOOP;
END;
$$ LANGUAGE plpgsql;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TRIGGER IF EXISTS shopping_lists_default_trigger ON shopping_lists;
DROP TRIGGER IF EXISTS shopping_lists_updated_at_trigger ON shopping_lists;
DROP FUNCTION IF EXISTS ensure_single_default_list();
DROP FUNCTION IF EXISTS update_updated_at_column();
DROP FUNCTION IF EXISTS update_shopping_list_stats();
DROP FUNCTION IF EXISTS generate_share_code();
DROP TABLE IF EXISTS shopping_lists;
-- +goose StatementEnd