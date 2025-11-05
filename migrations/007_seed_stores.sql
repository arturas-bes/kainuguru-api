-- +goose Up
-- +goose StatementBegin
-- Additional store configurations and detailed location data for Lithuanian stores

-- Update existing stores with more detailed configuration
UPDATE stores SET
    scraper_config = '{
        "user_agent": "KainuguruBot/1.0 (Lithuanian grocery price aggregator)",
        "request_delay": 2000,
        "max_retries": 3,
        "respect_robots_txt": true,
        "flyer_selector": ".flyer-container",
        "pagination_selector": ".page-nav"
    }',
    locations = '[
        {"city": "Vilnius", "lat": 54.6872, "lng": 25.2797, "address": "Konstitucijos pr. 7"},
        {"city": "Kaunas", "lat": 54.8985, "lng": 23.9036, "address": "Laisvės al. 53"},
        {"city": "Klaipėda", "lat": 55.7033, "lng": 21.1443, "address": "Taikos pr. 139"}
    ]'
WHERE code = 'iki';

UPDATE stores SET
    scraper_config = '{
        "user_agent": "KainuguruBot/1.0 (Lithuanian grocery price aggregator)",
        "request_delay": 1500,
        "max_retries": 3,
        "respect_robots_txt": true,
        "flyer_selector": ".maxima-flyer",
        "api_endpoint": "https://www.maxima.lt/api/flyers"
    }',
    locations = '[
        {"city": "Vilnius", "lat": 54.6872, "lng": 25.2797, "address": "Verkių g. 29"},
        {"city": "Kaunas", "lat": 54.8985, "lng": 23.9036, "address": "Savanorių pr. 255"},
        {"city": "Klaipėda", "lat": 55.7033, "lng": 21.1443, "address": "Sausio 15-osios g. 13"},
        {"city": "Šiauliai", "lat": 55.9349, "lng": 23.3131, "address": "Vilniaus g. 213"}
    ]'
WHERE code = 'maxima';

UPDATE stores SET
    scraper_config = '{
        "user_agent": "KainuguruBot/1.0 (Lithuanian grocery price aggregator)",
        "request_delay": 2500,
        "max_retries": 3,
        "respect_robots_txt": true,
        "flyer_selector": ".rimi-weekly-flyer",
        "requires_js": true
    }',
    locations = '[
        {"city": "Vilnius", "lat": 54.6872, "lng": 25.2797, "address": "Ozo g. 25"},
        {"city": "Kaunas", "lat": 54.8985, "lng": 23.9036, "address": "Islandijos pl. 32"},
        {"city": "Klaipėda", "lat": 55.7033, "lng": 21.1443, "address": "Šilutės pl. 35"}
    ]'
WHERE code = 'rimi';

UPDATE stores SET
    scraper_config = '{
        "user_agent": "KainuguruBot/1.0 (Lithuanian grocery price aggregator)",
        "request_delay": 1000,
        "max_retries": 3,
        "respect_robots_txt": true,
        "flyer_selector": ".lidl-leaflet",
        "weekly_schedule": "monday"
    }',
    locations = '[
        {"city": "Vilnius", "lat": 54.6872, "lng": 25.2797, "address": "Rinktinės g. 56"},
        {"city": "Kaunas", "lat": 54.8985, "lng": 23.9036, "address": "Raudondvario pl. 101"},
        {"city": "Klaipėda", "lat": 55.7033, "lng": 21.1443, "address": "Sausio 15-osios g. 25"}
    ]'
WHERE code = 'lidl';

UPDATE stores SET
    scraper_config = '{
        "user_agent": "KainuguruBot/1.0 (Lithuanian grocery price aggregator)",
        "request_delay": 3000,
        "max_retries": 2,
        "respect_robots_txt": true,
        "flyer_selector": ".norfa-offers",
        "regional_focus": "rural"
    }',
    locations = '[
        {"city": "Panevėžys", "lat": 55.7353, "lng": 24.3569, "address": "Respublikos g. 3"},
        {"city": "Alytus", "lat": 54.3963, "lng": 24.0458, "address": "Pulko g. 15"},
        {"city": "Marijampolė", "lat": 54.5593, "lng": 23.3542, "address": "Vilniaus g. 229"}
    ]'
WHERE code = 'norfa';

-- Insert additional stores
INSERT INTO stores (code, name, logo_url, website_url, flyer_source_url, scraper_config, locations, is_active) VALUES
('barbora', 'Barbora', 'https://www.barbora.lt/static/img/logo.png', 'https://www.barbora.lt', 'https://www.barbora.lt/akcijos',
    '{
        "user_agent": "KainuguruBot/1.0 (Lithuanian grocery price aggregator)",
        "request_delay": 2000,
        "max_retries": 3,
        "respect_robots_txt": true,
        "api_endpoint": "https://www.barbora.lt/api/promotions",
        "requires_auth": false,
        "type": "online_only"
    }',
    '[
        {"city": "Vilnius", "lat": 54.6872, "lng": 25.2797, "address": "Online delivery service"},
        {"city": "Kaunas", "lat": 54.8985, "lng": 23.9036, "address": "Online delivery service"},
        {"city": "Klaipėda", "lat": 55.7033, "lng": 21.1443, "address": "Online delivery service"}
    ]',
    true
),
('elta', 'Elta', 'https://www.elta.lt/static/img/logo.png', 'https://www.elta.lt', 'https://www.elta.lt/akcijos',
    '{
        "user_agent": "KainuguruBot/1.0 (Lithuanian grocery price aggregator)",
        "request_delay": 2500,
        "max_retries": 2,
        "respect_robots_txt": true,
        "flyer_selector": ".elta-akcijos",
        "regional_focus": "small_towns"
    }',
    '[
        {"city": "Utena", "lat": 55.4975, "lng": 25.5997, "address": "Utenio g. 2"},
        {"city": "Telšiai", "lat": 55.9814, "lng": 22.2476, "address": "Respublikos g. 34"}
    ]',
    true
);

-- Create store popularity ranking based on Lithuanian market share
UPDATE stores SET scraper_config = scraper_config || '{"market_share": 35, "priority": 1}' WHERE code = 'maxima';
UPDATE stores SET scraper_config = scraper_config || '{"market_share": 25, "priority": 2}' WHERE code = 'rimi';
UPDATE stores SET scraper_config = scraper_config || '{"market_share": 20, "priority": 3}' WHERE code = 'iki';
UPDATE stores SET scraper_config = scraper_config || '{"market_share": 12, "priority": 4}' WHERE code = 'lidl';
UPDATE stores SET scraper_config = scraper_config || '{"market_share": 5, "priority": 5}' WHERE code = 'norfa';
UPDATE stores SET scraper_config = scraper_config || '{"market_share": 2, "priority": 6}' WHERE code = 'barbora';
UPDATE stores SET scraper_config = scraper_config || '{"market_share": 1, "priority": 7}' WHERE code = 'elta';
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DELETE FROM stores WHERE code IN ('barbora', 'elta');

UPDATE stores SET
    scraper_config = '{}',
    locations = '[]'
WHERE code IN ('iki', 'maxima', 'rimi', 'lidl', 'norfa');
-- +goose StatementEnd