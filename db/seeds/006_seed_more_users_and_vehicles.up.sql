-- Seed: More test users with vehicles + socials + linked QR codes
-- After migration 004: all FK references use display_id (UUID), not id (bigint)

WITH data(idx, phone, first_name, last_name, device_id, plate_number, car_model, whatsapp, instagram, reviews_enabled, telegram_enabled) AS (
  VALUES
    (1,  '+996700100101', 'Алишер',  'Абдиев',    'device-seed-101', '01KG101AAA', 'Toyota Prius',        '+996700100101', 'alisher_prius', TRUE,  TRUE),
    (2,  '+996700100102', 'Эрмек',   'Садыков',   'device-seed-102', '02KG102BBB', 'Honda Fit',           '+996700100102', 'ermek_fit',     TRUE,  FALSE),
    (3,  '+996700100103', 'Айжан',   'Токтосун',  'device-seed-103', '03KG103CCC', 'Kia K5',              NULL,            'aizhan_k5',     TRUE,  TRUE),
    (4,  '+996700100104', 'Нурбек',  'Ибраев',    'device-seed-104', '04KG104DDD', 'Hyundai Elantra',     '+996700100104', NULL,            FALSE, FALSE),
    (5,  '+996700100105', 'Диана',   'Осмонова',  'device-seed-105', '05KG105EEE', 'Lexus RX',            '+996700100105', 'diana_lex',     TRUE,  TRUE),
    (6,  '+996700100106', 'Талант',  'Жумабаев',  'device-seed-106', '06KG106FFF', 'BMW 5',               NULL,            NULL,            TRUE,  FALSE),
    (7,  '+996700100107', 'Бакыт',   'Калиев',    'device-seed-107', '07KG107GGG', 'Mercedes E',          '+996700100107', 'bakyt_e',       TRUE,  FALSE),
    (8,  '+996700100108', 'Алина',   'Кожокова',  'device-seed-108', '08KG108HHH', 'Mazda 6',             '+996700100108', 'alina_mazda',   TRUE,  TRUE),
    (9,  '+996700100109', 'Самат',   'Рысбеков',  'device-seed-109', '09KG109III', 'Toyota Camry',        '+996700100109', 'samat_camry',   TRUE,  TRUE),
    (10, '+996700100110', 'Гульнара','Исмаилова', 'device-seed-110', '10KG110JJJ', 'Nissan X-Trail',      NULL,            'gulnara_x',     TRUE,  FALSE),
    (11, '+996700100111', 'Ильяс',   'Орозбеков', 'device-seed-111', '11KG111KKK', 'Toyota Land Cruiser', '+996700100111', NULL,            TRUE,  TRUE),
    (12, '+996700100112', 'Каныкей', 'Асанова',   'device-seed-112', '12KG112LLL', 'Hyundai Tucson',      '+996700100112', 'kanykei_tuc',   TRUE,  FALSE)
),
ins_users AS (
  INSERT INTO users (display_id, phone, first_name, last_name, device_id)
  SELECT gen_random_uuid(), d.phone, d.first_name, d.last_name, d.device_id
  FROM data d
  ON CONFLICT (phone) DO UPDATE
    SET first_name = EXCLUDED.first_name,
        last_name  = EXCLUDED.last_name,
        device_id  = EXCLUDED.device_id
  RETURNING display_id, phone
),
ins_social_whatsapp AS (
  INSERT INTO social_profiles (id, user_id, platform, handle, is_public)
  SELECT gen_random_uuid(), u.display_id, 'whatsapp', d.whatsapp, TRUE
  FROM ins_users u
  JOIN data d ON d.phone = u.phone
  WHERE d.whatsapp IS NOT NULL
  ON CONFLICT (user_id, platform) DO UPDATE
    SET handle    = EXCLUDED.handle,
        is_public = EXCLUDED.is_public
  RETURNING id
),
ins_social_instagram AS (
  INSERT INTO social_profiles (id, user_id, platform, handle, is_public)
  SELECT gen_random_uuid(), u.display_id, 'instagram', d.instagram, TRUE
  FROM ins_users u
  JOIN data d ON d.phone = u.phone
  WHERE d.instagram IS NOT NULL
  ON CONFLICT (user_id, platform) DO UPDATE
    SET handle    = EXCLUDED.handle,
        is_public = EXCLUDED.is_public
  RETURNING id
),
ins_vehicles AS (
  INSERT INTO vehicles (display_id, user_id, plate_number, car_model, is_public, reviews_enabled, telegram_enabled)
  SELECT gen_random_uuid(), u.display_id, d.plate_number, d.car_model, TRUE, d.reviews_enabled, d.telegram_enabled
  FROM ins_users u
  JOIN data d ON d.phone = u.phone
  ON CONFLICT DO NOTHING
  RETURNING display_id
),
ranked_vehicles AS (
  SELECT v.display_id AS vehicle_id,
         row_number() OVER (ORDER BY v.display_id) AS rn
  FROM ins_vehicles v
),
ranked_qr AS (
  SELECT q.id AS qr_id,
         row_number() OVER (ORDER BY q.created_at, q.code) AS rn
  FROM qr_codes q
  WHERE q.status = 'unregistered' AND q.vehicle_id IS NULL
),
assign AS (
  UPDATE qr_codes q
  SET vehicle_id    = rv.vehicle_id,
      status        = 'active',
      registered_at = NOW() - (RANDOM() * INTERVAL '40 days')
  FROM ranked_vehicles rv
  JOIN ranked_qr rq ON rq.rn = rv.rn
  WHERE q.id = rq.qr_id
  RETURNING q.vehicle_id
)
SELECT 1;
