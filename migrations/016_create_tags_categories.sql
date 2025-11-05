-- +goose Up
-- +goose StatementBegin
-- Table for predefined product categories
CREATE TABLE product_categories (
    id SERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL UNIQUE,
    normalized_name VARCHAR(100) NOT NULL,
    parent_id INTEGER REFERENCES product_categories(id),

    -- Display information
    display_name_lt VARCHAR(100) NOT NULL, -- Lithuanian display name
    display_name_en VARCHAR(100), -- English display name
    description TEXT,
    icon_name VARCHAR(50), -- Icon identifier for UI
    color_hex VARCHAR(7), -- Hex color for UI theming

    -- Hierarchy and ordering
    level INTEGER DEFAULT 0, -- 0 = root category, 1 = subcategory, etc.
    sort_order INTEGER DEFAULT 0,

    -- Matching keywords for auto-categorization
    keywords TEXT[], -- Keywords that suggest this category
    excluded_keywords TEXT[], -- Keywords that exclude this category

    -- Usage statistics
    product_count INTEGER DEFAULT 0, -- Number of products in this category
    usage_count INTEGER DEFAULT 0, -- How often used in shopping lists

    -- Status
    is_active BOOLEAN DEFAULT TRUE,

    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),

    -- Constraints
    CONSTRAINT category_name_length CHECK (LENGTH(name) >= 1 AND LENGTH(name) <= 100),
    CONSTRAINT category_level_check CHECK (level >= 0 AND level <= 5)
);

-- Table for product tags (flexible labeling system)
CREATE TABLE product_tags (
    id SERIAL PRIMARY KEY,
    name VARCHAR(50) NOT NULL UNIQUE,
    normalized_name VARCHAR(50) NOT NULL,

    -- Display information
    display_name_lt VARCHAR(50) NOT NULL, -- Lithuanian display name
    display_name_en VARCHAR(50), -- English display name
    description TEXT,
    color_hex VARCHAR(7), -- Hex color for tag display

    -- Tag type and behavior
    tag_type VARCHAR(20) DEFAULT 'general', -- 'general', 'dietary', 'allergen', 'quality', 'seasonal'
    is_system_tag BOOLEAN DEFAULT FALSE, -- System-generated vs user-created

    -- Usage statistics
    usage_count INTEGER DEFAULT 0, -- How often this tag is used

    -- Status
    is_active BOOLEAN DEFAULT TRUE,

    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),

    -- Constraints
    CONSTRAINT tag_name_length CHECK (LENGTH(name) >= 1 AND LENGTH(name) <= 50)
);

-- Junction table for shopping list custom categories
CREATE TABLE shopping_list_categories (
    id BIGSERIAL PRIMARY KEY,
    shopping_list_id BIGINT NOT NULL REFERENCES shopping_lists(id) ON DELETE CASCADE,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,

    -- Category details
    name VARCHAR(100) NOT NULL,
    color_hex VARCHAR(7),
    icon_name VARCHAR(50),
    sort_order INTEGER DEFAULT 0,

    -- Item count in this category for this list
    item_count INTEGER DEFAULT 0,

    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),

    -- Constraints
    CONSTRAINT list_category_name_length CHECK (LENGTH(name) >= 1 AND LENGTH(name) <= 100),
    CONSTRAINT list_category_unique UNIQUE (shopping_list_id, name)
);

-- Junction table for user's personal tags
CREATE TABLE user_tags (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    tag_name VARCHAR(50) NOT NULL,

    -- Custom display properties
    display_name VARCHAR(50),
    color_hex VARCHAR(7),

    -- Usage tracking
    usage_count INTEGER DEFAULT 0,
    last_used_at TIMESTAMP WITH TIME ZONE,

    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),

    -- Constraints
    CONSTRAINT user_tag_name_length CHECK (LENGTH(tag_name) >= 1 AND LENGTH(tag_name) <= 50),
    CONSTRAINT user_tag_unique UNIQUE (user_id, tag_name)
);

-- Indexes for performance
CREATE INDEX idx_product_categories_parent ON product_categories(parent_id);
CREATE INDEX idx_product_categories_level ON product_categories(level);
CREATE INDEX idx_product_categories_active ON product_categories(is_active);
CREATE INDEX idx_product_categories_normalized ON product_categories(normalized_name);
CREATE INDEX idx_product_categories_keywords ON product_categories USING gin(keywords);

CREATE INDEX idx_product_tags_type ON product_tags(tag_type);
CREATE INDEX idx_product_tags_active ON product_tags(is_active);
CREATE INDEX idx_product_tags_system ON product_tags(is_system_tag);
CREATE INDEX idx_product_tags_normalized ON product_tags(normalized_name);
CREATE INDEX idx_product_tags_usage ON product_tags(usage_count DESC);

CREATE INDEX idx_shopping_list_categories_list ON shopping_list_categories(shopping_list_id);
CREATE INDEX idx_shopping_list_categories_user ON shopping_list_categories(user_id);
CREATE INDEX idx_shopping_list_categories_order ON shopping_list_categories(shopping_list_id, sort_order);

CREATE INDEX idx_user_tags_user ON user_tags(user_id);
CREATE INDEX idx_user_tags_usage ON user_tags(user_id, usage_count DESC);
CREATE INDEX idx_user_tags_recent ON user_tags(user_id, last_used_at DESC);

-- Function to normalize tag/category names
CREATE OR REPLACE FUNCTION update_normalized_names()
RETURNS trigger AS $$
BEGIN
    NEW.normalized_name := normalize_lithuanian_text(NEW.name);
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Triggers to update normalized names
CREATE TRIGGER product_categories_normalized_trigger
BEFORE INSERT OR UPDATE ON product_categories
FOR EACH ROW EXECUTE FUNCTION update_normalized_names();

CREATE TRIGGER product_tags_normalized_trigger
BEFORE INSERT OR UPDATE ON product_tags
FOR EACH ROW EXECUTE FUNCTION update_normalized_names();

-- Function to update category level based on parent
CREATE OR REPLACE FUNCTION update_category_level()
RETURNS trigger AS $$
BEGIN
    IF NEW.parent_id IS NOT NULL THEN
        SELECT level + 1 INTO NEW.level
        FROM product_categories
        WHERE id = NEW.parent_id;
    ELSE
        NEW.level := 0;
    END IF;

    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Trigger to maintain category hierarchy
CREATE TRIGGER product_categories_level_trigger
BEFORE INSERT OR UPDATE ON product_categories
FOR EACH ROW EXECUTE FUNCTION update_category_level();

-- Function to update category usage counts
CREATE OR REPLACE FUNCTION update_category_usage()
RETURNS trigger AS $$
DECLARE
    cat_name VARCHAR(100);
BEGIN
    -- Get category name from the changed item
    cat_name := COALESCE(NEW.category, OLD.category);

    IF cat_name IS NOT NULL THEN
        -- Update product category usage count
        UPDATE product_categories
        SET
            usage_count = (
                SELECT COUNT(*)
                FROM shopping_list_items
                WHERE category = cat_name
            ),
            updated_at = NOW()
        WHERE name = cat_name;

        -- Update shopping list category item count if it's a custom category
        UPDATE shopping_list_categories
        SET item_count = (
            SELECT COUNT(*)
            FROM shopping_list_items sli
            WHERE sli.shopping_list_id = shopping_list_categories.shopping_list_id
            AND sli.category = shopping_list_categories.name
        )
        WHERE name = cat_name
        AND shopping_list_id = COALESCE(NEW.shopping_list_id, OLD.shopping_list_id);
    END IF;

    RETURN COALESCE(NEW, OLD);
END;
$$ LANGUAGE plpgsql;

-- Trigger to update usage counts when items change category
CREATE TRIGGER shopping_list_items_category_usage_trigger
AFTER INSERT OR UPDATE OR DELETE ON shopping_list_items
FOR EACH ROW EXECUTE FUNCTION update_category_usage();

-- Function to suggest category based on item description
CREATE OR REPLACE FUNCTION suggest_category(item_description TEXT)
RETURNS VARCHAR(100) AS $$
DECLARE
    suggested_category VARCHAR(100);
    normalized_desc TEXT;
BEGIN
    normalized_desc := normalize_lithuanian_text(item_description);

    -- Look for category matches based on keywords
    SELECT name INTO suggested_category
    FROM product_categories
    WHERE is_active = TRUE
    AND EXISTS (
        SELECT 1 FROM unnest(keywords) AS keyword
        WHERE normalized_desc ILIKE '%' || keyword || '%'
    )
    AND NOT EXISTS (
        SELECT 1 FROM unnest(excluded_keywords) AS excluded
        WHERE normalized_desc ILIKE '%' || excluded || '%'
    )
    ORDER BY usage_count DESC, level DESC
    LIMIT 1;

    RETURN suggested_category;
END;
$$ LANGUAGE plpgsql;

-- Function to update user tag usage
CREATE OR REPLACE FUNCTION update_user_tag_usage()
RETURNS trigger AS $$
DECLARE
    tag_name TEXT;
    item_user_id BIGINT;
BEGIN
    item_user_id := COALESCE(NEW.user_id, OLD.user_id);

    -- Update usage for all tags in the item
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

-- Trigger to track user tag usage
CREATE TRIGGER shopping_list_items_tag_usage_trigger
AFTER INSERT OR UPDATE ON shopping_list_items
FOR EACH ROW EXECUTE FUNCTION update_user_tag_usage();

-- Insert default categories
INSERT INTO product_categories (name, display_name_lt, display_name_en, icon_name, color_hex, keywords, level) VALUES
('dairy', 'Pieno produktai', 'Dairy Products', 'milk', '#F8F9FA', ARRAY['pienas', 'jogurtas', 'grietinė', 'varškė', 'sūris', 'sviestas'], 0),
('meat', 'Mėsa ir žuvis', 'Meat & Fish', 'meat', '#FF6B6B', ARRAY['mėsa', 'kiauliena', 'jautiena', 'vištiena', 'žuvis', 'dešros'], 0),
('fruits', 'Vaisiai', 'Fruits', 'apple', '#4ECDC4', ARRAY['obuoliai', 'bananai', 'apelsinai', 'citrina', 'braškės', 'vynuogės'], 0),
('vegetables', 'Daržovės', 'Vegetables', 'carrot', '#45B7D1', ARRAY['pomidorai', 'agurkai', 'svogūnai', 'morkos', 'kopūstai', 'bulvės'], 0),
('bread', 'Duona ir konditerija', 'Bread & Bakery', 'bread', '#FFA07A', ARRAY['duona', 'bandelės', 'tortai', 'sausainiai', 'pyragai'], 0),
('beverages', 'Gėrimai', 'Beverages', 'glass', '#98D8C8', ARRAY['vanduo', 'sultys', 'gazuoti', 'alus', 'vynas', 'kava', 'arbata'], 0),
('household', 'Namų ūkio prekės', 'Household Items', 'home', '#A8E6CF', ARRAY['plovikliai', 'muilas', 'popierius', 'valymo', 'higienos'], 0),
('frozen', 'Šaldyti produktai', 'Frozen Foods', 'snowflake', '#B8E6F5', ARRAY['šaldyti', 'ledai', 'šaldyta'], 0),
('snacks', 'Užkandžiai', 'Snacks', 'cookie', '#FFE5B4', ARRAY['traškučiai', 'riešutai', 'šokoladas', 'saldainiai'], 0),
('health', 'Sveikata ir grožis', 'Health & Beauty', 'heart', '#FFB6C1', ARRAY['vaistai', 'vitaminai', 'kosmetika', 'šampūnas'], 0);

-- Insert default tags
INSERT INTO product_tags (name, display_name_lt, display_name_en, tag_type, is_system_tag, color_hex) VALUES
('organic', 'Ekologiškas', 'Organic', 'quality', true, '#4CAF50'),
('gluten_free', 'Be glitimo', 'Gluten Free', 'dietary', true, '#FF9800'),
('lactose_free', 'Be laktozės', 'Lactose Free', 'dietary', true, '#2196F3'),
('vegetarian', 'Vegetariškas', 'Vegetarian', 'dietary', true, '#8BC34A'),
('vegan', 'Veganiška', 'Vegan', 'dietary', true, '#009688'),
('local', 'Vietinis', 'Local', 'quality', true, '#795548'),
('seasonal', 'Sezoniška', 'Seasonal', 'seasonal', true, '#FF5722'),
('discount', 'Akcija', 'On Sale', 'general', true, '#F44336'),
('premium', 'Premium', 'Premium', 'quality', true, '#9C27B0'),
('fresh', 'Šviežias', 'Fresh', 'quality', true, '#4CAF50');
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TRIGGER IF EXISTS shopping_list_items_tag_usage_trigger ON shopping_list_items;
DROP TRIGGER IF EXISTS shopping_list_items_category_usage_trigger ON shopping_list_items;
DROP TRIGGER IF EXISTS product_categories_level_trigger ON product_categories;
DROP TRIGGER IF EXISTS product_tags_normalized_trigger ON product_tags;
DROP TRIGGER IF EXISTS product_categories_normalized_trigger ON product_categories;
DROP FUNCTION IF EXISTS update_user_tag_usage();
DROP FUNCTION IF EXISTS suggest_category(TEXT);
DROP FUNCTION IF EXISTS update_category_usage();
DROP FUNCTION IF EXISTS update_category_level();
DROP FUNCTION IF EXISTS update_normalized_names();
DROP TABLE IF EXISTS user_tags;
DROP TABLE IF EXISTS shopping_list_categories;
DROP TABLE IF EXISTS product_tags;
DROP TABLE IF EXISTS product_categories;
-- +goose StatementEnd