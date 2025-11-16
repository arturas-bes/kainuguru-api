-- +goose Up
-- +goose StatementBegin
-- Add is_locked column to shopping_lists table for FR-016 compliance
-- Prevents concurrent wizard sessions on the same list
ALTER TABLE shopping_lists
ADD COLUMN is_locked BOOLEAN NOT NULL DEFAULT false;

-- Add index for quick lookup of locked lists
CREATE INDEX idx_shopping_lists_is_locked ON shopping_lists(is_locked) WHERE is_locked = true;

-- Add comment for documentation
COMMENT ON COLUMN shopping_lists.is_locked IS 'Indicates if list is currently being migrated by wizard session (FR-016)';
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS idx_shopping_lists_is_locked;
ALTER TABLE shopping_lists DROP COLUMN IF EXISTS is_locked;
-- +goose StatementEnd
