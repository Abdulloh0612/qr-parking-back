-- Seed: 5 test users with registered vehicles and active QR codes
-- After migration 004: vehicles.user_id = users.display_id (UUID)
--                      qr_codes.vehicle_id = vehicles.display_id (UUID)

-- User 1: Азамат Исаков — Toyota Camry (with WhatsApp social)
WITH u1 AS (
    INSERT INTO users (display_id, phone, first_name, last_name, device_id)
    VALUES (gen_random_uuid(), '+99670123456', 'Азамат', 'Исаков', 'device-test-001')
    ON CONFLICT (phone) DO UPDATE SET first_name = EXCLUDED.first_name
    RETURNING display_id
),
v1 AS (
    INSERT INTO vehicles (display_id, user_id, plate_number, car_model, is_public, reviews_enabled, telegram_enabled)
    SELECT gen_random_uuid(), u1.display_id, '01KG123ABC', 'Toyota Camry', TRUE, TRUE, FALSE
    FROM u1
    ON CONFLICT DO NOTHING
    RETURNING display_id
),
qr1 AS (
    UPDATE qr_codes SET
        vehicle_id    = v1.display_id,
        status        = 'active',
        registered_at = NOW()
    FROM v1
    WHERE qr_codes.id = (
        SELECT id FROM qr_codes WHERE status = 'unregistered' ORDER BY created_at LIMIT 1
    )
    RETURNING qr_codes.id
)
INSERT INTO social_profiles (id, user_id, platform, handle, is_public)
SELECT gen_random_uuid(), u1.display_id, 'whatsapp', '+99670123456', TRUE FROM u1
ON CONFLICT (user_id, platform) DO NOTHING;

-- User 2: Айгуль Бекова — BMW X5
WITH u2 AS (
    INSERT INTO users (display_id, phone, first_name, last_name, device_id)
    VALUES (gen_random_uuid(), '+99655987654', 'Айгуль', 'Бекова', 'device-test-002')
    ON CONFLICT (phone) DO UPDATE SET first_name = EXCLUDED.first_name
    RETURNING display_id
),
v2 AS (
    INSERT INTO vehicles (display_id, user_id, plate_number, car_model, is_public, reviews_enabled, telegram_enabled)
    SELECT gen_random_uuid(), u2.display_id, '05KG777BBB', 'BMW X5', TRUE, TRUE, FALSE
    FROM u2
    ON CONFLICT DO NOTHING
    RETURNING display_id
)
UPDATE qr_codes SET
    vehicle_id    = v2.display_id,
    status        = 'active',
    registered_at = NOW() - INTERVAL '5 days'
FROM v2
WHERE qr_codes.id = (
    SELECT id FROM qr_codes WHERE status = 'unregistered' ORDER BY created_at LIMIT 1
);

-- User 3: Нурлан Джакыпов — Mercedes GLE
WITH u3 AS (
    INSERT INTO users (display_id, phone, first_name, last_name, device_id)
    VALUES (gen_random_uuid(), '+99670111222', 'Нурлан', 'Джакыпов', 'device-test-003')
    ON CONFLICT (phone) DO UPDATE SET first_name = EXCLUDED.first_name
    RETURNING display_id
),
v3 AS (
    INSERT INTO vehicles (display_id, user_id, plate_number, car_model, is_public, reviews_enabled, telegram_enabled)
    SELECT gen_random_uuid(), u3.display_id, '03KG456DEF', 'Mercedes GLE', TRUE, FALSE, FALSE
    FROM u3
    ON CONFLICT DO NOTHING
    RETURNING display_id
)
UPDATE qr_codes SET
    vehicle_id    = v3.display_id,
    status        = 'active',
    registered_at = NOW() - INTERVAL '10 days'
FROM v3
WHERE qr_codes.id = (
    SELECT id FROM qr_codes WHERE status = 'unregistered' ORDER BY created_at LIMIT 1
);

-- User 4: Жибек Асанова — Hyundai Sonata
WITH u4 AS (
    INSERT INTO users (display_id, phone, first_name, last_name, device_id)
    VALUES (gen_random_uuid(), '+99670333444', 'Жибек', 'Асанова', 'device-test-004')
    ON CONFLICT (phone) DO UPDATE SET first_name = EXCLUDED.first_name
    RETURNING display_id
),
v4 AS (
    INSERT INTO vehicles (display_id, user_id, plate_number, car_model, is_public, reviews_enabled)
    SELECT gen_random_uuid(), u4.display_id, '07KG999GHI', 'Hyundai Sonata', TRUE, TRUE
    FROM u4
    ON CONFLICT DO NOTHING
    RETURNING display_id
)
UPDATE qr_codes SET
    vehicle_id    = v4.display_id,
    status        = 'active',
    registered_at = NOW() - INTERVAL '3 days'
FROM v4
WHERE qr_codes.id = (
    SELECT id FROM qr_codes WHERE status = 'unregistered' ORDER BY created_at LIMIT 1
);

-- User 5: Тимур Садыков — Lexus RX
WITH u5 AS (
    INSERT INTO users (display_id, phone, first_name, last_name, device_id)
    VALUES (gen_random_uuid(), '+99650555666', 'Тимур', 'Садыков', 'device-test-005')
    ON CONFLICT (phone) DO UPDATE SET first_name = EXCLUDED.first_name
    RETURNING display_id
),
v5 AS (
    INSERT INTO vehicles (display_id, user_id, plate_number, car_model, is_public, reviews_enabled)
    SELECT gen_random_uuid(), u5.display_id, '09KG111JKL', 'Lexus RX', TRUE, TRUE
    FROM u5
    ON CONFLICT DO NOTHING
    RETURNING display_id
)
UPDATE qr_codes SET
    vehicle_id    = v5.display_id,
    status        = 'active',
    registered_at = NOW() - INTERVAL '15 days'
FROM v5
WHERE qr_codes.id = (
    SELECT id FROM qr_codes WHERE status = 'unregistered' ORDER BY created_at LIMIT 1
);
