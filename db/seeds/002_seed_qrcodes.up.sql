-- Seed: Generate 20 test QR codes (unregistered, ready to be claimed)
INSERT INTO qr_codes (display_id, code, vehicle_id, status, created_by, created_at)
SELECT
    gen_random_uuid(),
    UPPER(SUBSTRING(MD5(RANDOM()::TEXT || generate_series::TEXT), 1, 12)),
    NULL,
    'unregistered',
    (SELECT display_id FROM users WHERE admin_username = 'admin' LIMIT 1),
    NOW() - (RANDOM() * INTERVAL '30 days')
FROM generate_series(1, 20)
ON CONFLICT (code) DO NOTHING;
