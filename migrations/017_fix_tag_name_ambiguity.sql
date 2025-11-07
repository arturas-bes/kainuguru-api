-- +goose Up
-- +goose StatementBegin
-- Fix ambiguous column reference in update_user_tag_usage() trigger
-- Issue: variable name 'tag_name' conflicts with column name 'user_tags.tag_name'

-- Drop the existing trigger first
DROP TRIGGER IF EXISTS shopping_list_items_tag_usage_trigger ON shopping_list_items;

-- Replace the function with a fixed version that uses 'tag_value' instead of 'tag_name'
CREATE OR REPLACE FUNCTION update_user_tag_usage()
RETURNS trigger AS $$
DECLARE
    tag_value TEXT;  -- Renamed from 'tag_name' to avoid ambiguity
    item_user_id UUID;
BEGIN
    item_user_id := COALESCE(NEW.user_id, OLD.user_id);

    -- Update usage for all tags in the item
    IF NEW.tags IS NOT NULL THEN
        FOREACH tag_value IN ARRAY NEW.tags
        LOOP
            INSERT INTO user_tags (user_id, tag_name, usage_count, last_used_at)
            VALUES (item_user_id, tag_value, 1, NOW())
            ON CONFLICT (user_id, tag_name)
            DO UPDATE SET
                usage_count = user_tags.usage_count + 1,
                last_used_at = NOW();
        END LOOP;
    END IF;

    RETURN COALESCE(NEW, OLD);
END;
$$ LANGUAGE plpgsql;

-- Recreate the trigger
CREATE TRIGGER shopping_list_items_tag_usage_trigger
AFTER INSERT OR UPDATE ON shopping_list_items
FOR EACH ROW EXECUTE FUNCTION update_user_tag_usage();
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
-- Revert to the original function with the ambiguous name
DROP TRIGGER IF EXISTS shopping_list_items_tag_usage_trigger ON shopping_list_items;

CREATE OR REPLACE FUNCTION update_user_tag_usage()
RETURNS trigger AS $$
DECLARE
    tag_name TEXT;
    item_user_id UUID;
BEGIN
    item_user_id := COALESCE(NEW.user_id, OLD.user_id);

    IF NEW.tags IS NOT NULL THEN
        FOREACH tag_name IN ARRAY NEW.tags
        LOOP
            INSERT INTO user_tags (user_id, tag_name, usage_count, last_used_at)
            VALUES (item_user_id, tag_name, 1, NOW())
            ON CONFLICT (user_id, tag_name)
            DO UPDATE SET
                usage_count = user_tags.usage_count + 1,
                last_used_at = NOW();
        END LOOP;
    END IF;

    RETURN COALESCE(NEW, OLD);
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER shopping_list_items_tag_usage_trigger
AFTER INSERT OR UPDATE ON shopping_list_items
FOR EACH ROW EXECUTE FUNCTION update_user_tag_usage();
-- +goose StatementEnd
