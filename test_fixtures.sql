-- Test Fixture Data for Search Testing
-- This file creates sample flyers and products for testing search functionality

TRUNCATE TABLE price_history RESTART IDENTITY CASCADE;
TRUNCATE TABLE products RESTART IDENTITY CASCADE;
TRUNCATE TABLE flyers RESTART IDENTITY CASCADE;
TRUNCATE TABLE stores RESTART IDENTITY CASCADE;

-- Ensure stores exist before inserting dependent rows
INSERT INTO stores (code, name, is_active, created_at, updated_at)
VALUES
('maxima', 'Maxima', TRUE, NOW(), NOW()),
('rimi', 'Rimi', TRUE, NOW(), NOW()),
('iki', 'IKI', TRUE, NOW(), NOW()),
('lidl', 'Lidl', TRUE, NOW(), NOW()),
('norfa', 'Norfa', TRUE, NOW(), NOW())
ON CONFLICT (code) DO NOTHING;

-- Create test flyers for different stores
INSERT INTO flyers (store_id, title, valid_from, valid_to, is_archived, created_at) VALUES
((SELECT id FROM stores WHERE code = 'maxima'), 'Maxima Savaitės Pasiūlymai', CURRENT_DATE - INTERVAL '1 day', CURRENT_DATE + INTERVAL '6 days', FALSE, NOW()),
((SELECT id FROM stores WHERE code = 'rimi'), 'Rimi Super Kainos', CURRENT_DATE - INTERVAL '1 day', CURRENT_DATE + INTERVAL '6 days', FALSE, NOW()),
((SELECT id FROM stores WHERE code = 'iki'), 'Iki Mega Nuolaidos', CURRENT_DATE, CURRENT_DATE + INTERVAL '7 days', FALSE, NOW()),
((SELECT id FROM stores WHERE code = 'lidl'), 'Lidl Savaitės Akcijos', CURRENT_DATE - INTERVAL '2 days', CURRENT_DATE + INTERVAL '4 days', FALSE, NOW()),
((SELECT id FROM stores WHERE code = 'norfa'), 'Norfa Kainų Mūšis', CURRENT_DATE, CURRENT_DATE + INTERVAL '5 days', FALSE, NOW())
RETURNING id;

-- Get flyer IDs (they should be 1-5 in order)
-- Now insert products with realistic Lithuanian grocery items

-- Dairy products
INSERT INTO products (flyer_id, store_id, name, normalized_name, brand, category, current_price, original_price, discount_percent, is_on_sale, is_available, valid_from, valid_to) VALUES
-- Maxima
(1, 1, 'Pienas „Žemaitijos" 2,5%, 1L', 'pienas zemaitijos 2,5%, 1l', 'Žemaitijos', 'Pieno produktai', 1.29, 1.59, 18.87, TRUE, TRUE, CURRENT_DATE - INTERVAL '1 day', CURRENT_DATE + INTERVAL '6 days'),
(1, 1, 'Sviestas „Valio", 200g', 'sviestas valio, 200g', 'Valio', 'Pieno produktai', 2.49, 2.99, 16.72, TRUE, TRUE, CURRENT_DATE - INTERVAL '1 day', CURRENT_DATE + INTERVAL '6 days'),
(1, 1, 'Varškė „Kaunas" 9%, 200g', 'varske kaunas 9%, 200g', 'Kaunas', 'Pieno produktai', 0.89, NULL, NULL, FALSE, TRUE, CURRENT_DATE - INTERVAL '1 day', CURRENT_DATE + INTERVAL '6 days'),
(1, 1, 'Grietinė „Rokiškio" 20%, 400g', 'grietine rokiskio 20%, 400g', 'Rokiškio', 'Pieno produktai', 1.49, 1.79, 16.76, TRUE, TRUE, CURRENT_DATE - INTERVAL '1 day', CURRENT_DATE + INTERVAL '6 days'),

-- Rimi
(2, 2, 'Pienas „Valio" 3,2%, 1L', 'pienas valio 3,2%, 1l', 'Valio', 'Pieno produktai', 1.35, NULL, NULL, FALSE, TRUE, CURRENT_DATE - INTERVAL '1 day', CURRENT_DATE + INTERVAL '6 days'),
(2, 2, 'Jogurtas „Ehrmann" braškių, 115g', 'jogurtas ehrmann braskiu, 115g', 'Ehrmann', 'Pieno produktai', 0.59, 0.79, 25.32, TRUE, TRUE, CURRENT_DATE - INTERVAL '1 day', CURRENT_DATE + INTERVAL '6 days'),
(2, 2, 'Sūris „Džiugas" 36mėn., 150g', 'suris dziugas 36men., 150g', 'Džiugas', 'Pieno produktai', 3.49, NULL, NULL, FALSE, TRUE, CURRENT_DATE - INTERVAL '1 day', CURRENT_DATE + INTERVAL '6 days'),

-- Bakery products
(1, 1, 'Duona „Rugiena" juoda, 500g', 'duona rugiena juoda, 500g', 'Rugiena', 'Duonos gaminiai', 0.99, NULL, NULL, FALSE, TRUE, CURRENT_DATE - INTERVAL '1 day', CURRENT_DATE + INTERVAL '6 days'),
(2, 2, 'Batonas „Vilniaus" baltyminis, 350g', 'batonas vilniaus baltyminis, 350g', 'Vilniaus', 'Duonos gaminiai', 1.19, 1.39, 14.39, TRUE, TRUE, CURRENT_DATE - INTERVAL '1 day', CURRENT_DATE + INTERVAL '6 days'),
(3, 3, 'Sausainiai „Roshen" sviestiniai, 500g', 'sausainiai roshen sviestiniai, 500g', 'Roshen', 'Duonos gaminiai', 2.29, 2.99, 23.41, TRUE, TRUE, CURRENT_DATE, CURRENT_DATE + INTERVAL '7 days'),

-- Meat products
(1, 1, 'Dešrelės „Kaniava" vakuume, 400g', 'desreles kaniava vakuume, 400g', 'Kaniava', 'Mėsos produktai', 2.99, 3.49, 14.33, TRUE, TRUE, CURRENT_DATE - INTERVAL '1 day', CURRENT_DATE + INTERVAL '6 days'),
(2, 2, 'Kumpis „Krekenavos" virtas, 1kg', 'kumpis krekenavos virtas, 1kg', 'Krekenavos', 'Mėsos produktai', 5.99, NULL, NULL, FALSE, TRUE, CURRENT_DATE - INTERVAL '1 day', CURRENT_DATE + INTERVAL '6 days'),
(4, 4, 'Šoninė „Lidl" rūkyta, 200g', 'sonine lidl rukyta, 200g', 'Lidl', 'Mėsos produktai', 1.99, 2.29, 13.10, TRUE, TRUE, CURRENT_DATE - INTERVAL '2 days', CURRENT_DATE + INTERVAL '4 days'),

-- Fruits and Vegetables
(3, 3, 'Obuoliai „Jonagold", 1kg', 'obuoliai jonagold, 1kg', NULL, 'Vaisiai ir daržovės', 1.49, 1.99, 25.13, TRUE, TRUE, CURRENT_DATE, CURRENT_DATE + INTERVAL '7 days'),
(3, 3, 'Pomidorai vyšniniai, 250g', 'pomidorai vysniniai, 250g', NULL, 'Vaisiai ir daržovės', 1.29, NULL, NULL, FALSE, TRUE, CURRENT_DATE, CURRENT_DATE + INTERVAL '7 days'),
(4, 4, 'Bulvės „Vineta", 2,5kg', 'bulves vineta, 2,5kg', NULL, 'Vaisiai ir daržovės', 1.99, NULL, NULL, FALSE, TRUE, CURRENT_DATE - INTERVAL '2 days', CURRENT_DATE + INTERVAL '4 days'),
(5, 5, 'Morkos, 1kg', 'morkos, 1kg', NULL, 'Vaisiai ir daržovės', 0.89, 1.19, 25.21, TRUE, TRUE, CURRENT_DATE, CURRENT_DATE + INTERVAL '5 days'),

-- Beverages
(1, 1, 'Sultys „Cido" obuolių, 1L', 'sultys cido obuoliu, 1l', 'Cido', 'Gėrimai', 1.79, NULL, NULL, FALSE, TRUE, CURRENT_DATE - INTERVAL '1 day', CURRENT_DATE + INTERVAL '6 days'),
(2, 2, 'Vanduo „Vytautas" gazuotas, 1L', 'vanduo vytautas gazuotas, 1l', 'Vytautas', 'Gėrimai', 0.69, 0.89, 22.47, TRUE, TRUE, CURRENT_DATE - INTERVAL '1 day', CURRENT_DATE + INTERVAL '6 days'),
(3, 3, 'Kava „Paulig" maltos, 500g', 'kava paulig maltos, 500g', 'Paulig', 'Gėrimai', 4.99, 6.49, 23.11, TRUE, TRUE, CURRENT_DATE, CURRENT_DATE + INTERVAL '7 days'),
(4, 4, 'Arbata „Lipton" juodoji, 100pak.', 'arbata lipton juodoji, 100pak.', 'Lipton', 'Gėrimai', 2.49, NULL, NULL, FALSE, TRUE, CURRENT_DATE - INTERVAL '2 days', CURRENT_DATE + INTERVAL '4 days'),

-- Snacks and Sweets
(2, 2, 'Šokoladas „Milka" pieninis, 100g', 'sokoladas milka pieninis, 100g', 'Milka', 'Saldumynai', 1.29, 1.49, 13.42, TRUE, TRUE, CURRENT_DATE - INTERVAL '1 day', CURRENT_DATE + INTERVAL '6 days'),
(3, 3, 'Traškučiai „Lays" klasikiniai, 150g', 'traskuciai lays klasikiniai, 150g', 'Lays', 'Užkandžiai', 1.99, 2.49, 20.08, TRUE, TRUE, CURRENT_DATE, CURRENT_DATE + INTERVAL '7 days'),
(4, 4, 'Ledai „Magnum" klasikiniai, 4pak.', 'ledai magnum klasikiniai, 4pak.', 'Magnum', 'Saldumynai', 3.99, 4.99, 20.04, TRUE, TRUE, CURRENT_DATE - INTERVAL '2 days', CURRENT_DATE + INTERVAL '4 days'),

-- Household items
(5, 5, 'Skalbimo milteliai „Persil", 2kg', 'skalbimo milteliai persil, 2kg', 'Persil', 'Buitinė chemija', 8.99, 10.99, 18.20, TRUE, TRUE, CURRENT_DATE, CURRENT_DATE + INTERVAL '5 days'),
(5, 5, 'Tualetinis popierius „Zewa", 8rul.', 'tualetinis popierius zewa, 8rul.', 'Zewa', 'Buitinė chemija', 4.49, NULL, NULL, FALSE, TRUE, CURRENT_DATE, CURRENT_DATE + INTERVAL '5 days'),

-- More products to test pagination
(1, 1, 'Pienas „Romuva" 1,5%, 1L', 'pienas romuva 1,5%, 1l', 'Romuva', 'Pieno produktai', 1.19, NULL, NULL, FALSE, TRUE, CURRENT_DATE - INTERVAL '1 day', CURRENT_DATE + INTERVAL '6 days'),
(2, 2, 'Pienas „Baltaragio" 2,5%, 1L', 'pienas baltaragio 2,5%, 1l', 'Baltaragio', 'Pieno produktai', 1.25, 1.45, 13.79, TRUE, TRUE, CURRENT_DATE - INTERVAL '1 day', CURRENT_DATE + INTERVAL '6 days'),
(3, 3, 'Pienas ekologiškas „Organic", 1L', 'pienas ekologiskas organic, 1l', 'Organic', 'Pieno produktai', 2.49, NULL, NULL, FALSE, TRUE, CURRENT_DATE, CURRENT_DATE + INTERVAL '7 days'),
(4, 4, 'Pienas „Mano pienas" 3,5%, 1L', 'pienas mano pienas 3,5%, 1l', 'Mano pienas', 'Pieno produktai', 1.39, NULL, NULL, FALSE, TRUE, CURRENT_DATE - INTERVAL '2 days', CURRENT_DATE + INTERVAL '4 days');

-- Verify the data
SELECT 'Inserted products:' as info, COUNT(*) as count FROM products;
SELECT 'Products by category:' as info, category, COUNT(*) as count FROM products GROUP BY category ORDER BY count DESC;
SELECT 'Products by store:' as info, s.name as store, COUNT(p.id) as count FROM products p JOIN stores s ON p.store_id = s.id GROUP BY s.name ORDER BY count DESC;
SELECT 'Products on sale:' as info, COUNT(*) as count FROM products WHERE is_on_sale = TRUE;
