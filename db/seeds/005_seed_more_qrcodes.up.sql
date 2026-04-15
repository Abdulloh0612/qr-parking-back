-- Seed: Generate more QR codes (unregistered, ready to be claimed)
-- Adds 200 codes on top of existing ones.

INSERT INTO qr_codes (display_id, code, vehicle_id, status, created_by, created_at)
SELECT
    gen_random_uuid(),
    UPPER(SUBSTRING(MD5(RANDOM()::TEXT || gs::TEXT), 1, 12)),
    NULL,
    'unregistered',
    (SELECT display_id FROM users WHERE admin_username = 'admin' LIMIT 1),
    NOW() - (RANDOM() * INTERVAL '90 days')
FROM generate_series(1, 200) AS gs
ON CONFLICT (code) DO NOTHING;

