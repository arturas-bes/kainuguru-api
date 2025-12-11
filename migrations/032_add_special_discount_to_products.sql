-- +goose Up
-- +goose StatementBegin
-- Migration: Add special_discount column to products table
-- This column stores special discount types like "1+1", "3 už 2", etc.

-- Add special_discount column to products table
ALTER TABLE products ADD COLUMN IF NOT EXISTS special_discount TEXT;

-- Add comment
COMMENT ON COLUMN products.special_discount IS 'Special discount type (e.g., "1+1", "3 už 2 €", "Antra prekė -50%")';

-- Create index for queries filtering by special discounts
CREATE INDEX IF NOT EXISTS idx_products_special_discount ON products(special_discount) WHERE special_discount IS NOT NULL;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE products DROP COLUMN IF EXISTS special_discount;
-- +goose StatementEnd
