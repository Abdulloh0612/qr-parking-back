-- Seed: 10 unregistered QR codes ready to be claimed
INSERT INTO qr_codes (display_id, code, status)
VALUES
    (gen_random_uuid(), 'test-qr-0000000001', 'unregistered'),
    (gen_random_uuid(), 'test-qr-0000000002', 'unregistered'),
    (gen_random_uuid(), 'test-qr-0000000003', 'unregistered'),
    (gen_random_uuid(), 'test-qr-0000000004', 'unregistered'),
    (gen_random_uuid(), 'test-qr-0000000005', 'unregistered'),
    (gen_random_uuid(), 'test-qr-0000000006', 'unregistered'),
    (gen_random_uuid(), 'test-qr-0000000007', 'unregistered'),
    (gen_random_uuid(), 'test-qr-0000000008', 'unregistered'),
    (gen_random_uuid(), 'test-qr-0000000009', 'unregistered'),
    (gen_random_uuid(), 'test-qr-0000000010', 'unregistered')
ON CONFLICT (code) DO NOTHING;
