UPDATE qr_codes SET vehicle_id = NULL, status = 'unregistered', registered_at = NULL
WHERE code = 'test-qr-0000000001';

DELETE FROM users WHERE phone = '+998900000000';
