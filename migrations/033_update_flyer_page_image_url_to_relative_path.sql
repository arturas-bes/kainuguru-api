-- +goose Up
-- +goose StatementBegin
-- Migration: Update flyer_pages.image_url to store relative paths instead of full URLs
-- Date: 2025-11-09
-- Description: Strip the base URL from image_url to store only the relative path

-- Update existing records to remove the base URL prefix
UPDATE flyer_pages
SET image_url = REPLACE(
    REPLACE(
        REPLACE(image_url, 'http://localhost:8080/', ''),
        'https://localhost:8080/', ''
    ),
    'http://localhost:8080', ''
)
WHERE image_url IS NOT NULL 
  AND (
    image_url LIKE 'http://localhost:8080/%' 
    OR image_url LIKE 'https://localhost:8080/%'
    OR image_url = 'http://localhost:8080'
  );

-- Also handle production URLs if any exist
UPDATE flyer_pages
SET image_url = REGEXP_REPLACE(image_url, '^https?://[^/]+/', '')
WHERE image_url IS NOT NULL 
  AND image_url ~ '^https?://';

-- Add comment to table
COMMENT ON COLUMN flyer_pages.image_url IS 'Relative path to flyer page image (e.g., flyers/iki/2025-11-03-iki-iki-kaininis-leidinys-nr-45/page-1.jpg). Base URL is configured via FLYER_BASE_URL env variable.';
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
-- No easy way to revert without knowing the original base URL for each record
-- +goose StatementEnd
