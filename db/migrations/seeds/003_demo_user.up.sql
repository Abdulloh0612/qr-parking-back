-- Seed: demo user with a registered QR code
DO $$
DECLARE
    v_user_id   UUID := gen_random_uuid();
    v_vehicle_id UUID := gen_random_uuid();
BEGIN
    INSERT INTO users (display_id, phone, first_name, last_name)
    VALUES (v_user_id, '+998900000000', 'Алишер', 'Каримов')
    ON CONFLICT (phone) DO UPDATE SET display_id = EXCLUDED.display_id
    RETURNING display_id INTO v_user_id;

    SELECT display_id INTO v_user_id FROM users WHERE phone = '+998900000000';

    INSERT INTO vehicles (display_id, user_id, plate_number, car_model, is_public, reviews_enabled, telegram_enabled)
    VALUES (v_vehicle_id, v_user_id, '01 A 777 AA', 'Chevrolet Nexia 3', TRUE, TRUE, FALSE)
    ON CONFLICT DO NOTHING;

    INSERT INTO social_profiles (user_id, platform, handle, is_public)
    VALUES
        (v_user_id, 'whatsapp',  '+998900000000', TRUE),
        (v_user_id, 'instagram', 'alisher_demo',   TRUE)
    ON CONFLICT (user_id, platform) DO NOTHING;

    UPDATE qr_codes
    SET vehicle_id    = v_vehicle_id,
        status        = 'active',
        registered_at = NOW()
    WHERE code = 'test-qr-0000000001'
      AND status = 'unregistered';
END $$;
