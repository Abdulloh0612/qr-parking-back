-- Seed: Messages, ratings, and scan events for registered vehicles
-- After migration 004: qr_codes.id = bigint, qr_codes.display_id = UUID
-- messages.qr_code_id, ratings.qr_code_id, scan_events.qr_code_id all reference qr_codes.display_id (UUID)

-- Messages to vehicle owners
INSERT INTO messages (id, qr_code_id, vehicle_id, content, sender_name, sender_phone, is_delivered, created_at)
SELECT
    gen_random_uuid(),
    q.display_id,
    q.vehicle_id,
    msg.content,
    msg.sender_name,
    msg.sender_phone,
    (RANDOM() > 0.3),
    NOW() - (RANDOM() * INTERVAL '20 days')
FROM qr_codes q
CROSS JOIN (VALUES
    ('Вы заблокировали мой выезд, пожалуйста уберите машину', 'Аскар',   '+99670999888'),
    ('Ваша машина мешает, срочно нужно выехать!',              'Фатима',  '+99655777666'),
    ('Добрый день, пожалуйста переставьте авто',               'Данияр',  '+99670444333'),
    ('Уберите машину пожалуйста',                              NULL,      NULL)
) AS msg(content, sender_name, sender_phone)
WHERE q.status = 'active'
  AND q.vehicle_id IS NOT NULL;

-- Ratings for vehicles with reviews_enabled
INSERT INTO ratings (id, vehicle_id, qr_code_id, rating, comment, created_at)
SELECT
    gen_random_uuid(),
    q.vehicle_id,
    q.display_id,
    (FLOOR(RANDOM() * 3) + 3)::SMALLINT,
    CASE WHEN RANDOM() > 0.5 THEN 'Отличный водитель, быстро среагировал' ELSE NULL END,
    NOW() - (RANDOM() * INTERVAL '30 days')
FROM qr_codes q
JOIN vehicles v ON v.display_id = q.vehicle_id AND v.reviews_enabled = TRUE
WHERE q.status = 'active';

-- Additional 5-star ratings
INSERT INTO ratings (id, vehicle_id, qr_code_id, rating, comment, created_at)
SELECT
    gen_random_uuid(),
    q.vehicle_id,
    q.display_id,
    5,
    'Очень вежливый и отзывчивый, убрал машину моментально!',
    NOW() - (RANDOM() * INTERVAL '10 days')
FROM qr_codes q
JOIN vehicles v ON v.display_id = q.vehicle_id AND v.reviews_enabled = TRUE
WHERE q.status = 'active';

-- Scan events (multiple per QR code)
INSERT INTO scan_events (id, qr_code_id, ip_address, user_agent, scanned_at)
SELECT
    gen_random_uuid(),
    q.display_id,
    ('192.168.' || FLOOR(RANDOM() * 255) || '.' || FLOOR(RANDOM() * 255))::INET,
    ua.agent,
    NOW() - (RANDOM() * INTERVAL '30 days')
FROM qr_codes q
CROSS JOIN (VALUES
    ('Mozilla/5.0 (iPhone; CPU iPhone OS 17_0 like Mac OS X) AppleWebKit/605.1.15'),
    ('Mozilla/5.0 (Linux; Android 13; SM-S908B) AppleWebKit/537.36 Chrome/112.0'),
    ('Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 Chrome/121.0'),
    ('Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 Safari/537.36'),
    ('Mozilla/5.0 (Linux; Android 12; Pixel 6) AppleWebKit/537.36 Chrome/110.0')
) AS ua(agent)
WHERE q.status = 'active';
