-- +goose Up
-- +goose StatementBegin
CREATE TABLE stores (
    id SERIAL PRIMARY KEY,
    code VARCHAR(50) UNIQUE NOT NULL, -- 'iki', 'maxima', 'rimi', etc
    name VARCHAR(100) NOT NULL,
    logo_url TEXT,
    website_url TEXT,
    flyer_source_url TEXT,

    -- Location data for future store-specific features
    locations JSONB DEFAULT '[]', -- Array of {lat, lng, address, city}

    -- Scraping configuration
    scraper_config JSONB DEFAULT '{}',
    scrape_schedule VARCHAR(50) DEFAULT 'weekly',
    last_scraped_at TIMESTAMP WITH TIME ZONE,

    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX idx_stores_code ON stores(code);
CREATE INDEX idx_stores_active ON stores(is_active);

-- Insert initial Lithuanian grocery stores
INSERT INTO stores (code, name, logo_url, website_url, flyer_source_url, is_active) VALUES
('iki', 'IKI', 'https://www.iki.lt/themes/iki/logo.png', 'https://www.iki.lt', 'https://www.iki.lt/akcijos', true),
('maxima', 'Maxima', 'https://www.maxima.lt/themes/maxima/logo.png', 'https://www.maxima.lt', 'https://www.maxima.lt/akcijos', true),
('rimi', 'Rimi', 'https://www.rimi.lt/themes/rimi/logo.png', 'https://www.rimi.lt', 'https://www.rimi.lt/akcijos', true),
('lidl', 'Lidl', 'https://www.lidl.lt/themes/lidl/logo.png', 'https://www.lidl.lt', 'https://www.lidl.lt/akcijos', true),
('norfa', 'Norfa', 'https://www.norfa.lt/themes/norfa/logo.png', 'https://www.norfa.lt', 'https://www.norfa.lt/akcijos', true);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS stores;
-- +goose StatementEnd