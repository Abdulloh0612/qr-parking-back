-- Seed 007: много клиентов, машин и QR с разным временем создания / регистрации
-- Рассчитано на схему после 004–006: qr_codes(id, code, …) без display_id
-- created_by для массовых вставок не задаём (NULL), как и в репозитории

-- A) Незарегистрированные QR: 150 шт., created_at растянуты примерно на 200 дней
INSERT INTO qr_codes (code, status, created_at)
SELECT
  'S7' || LPAD(gs::text, 10, '0'),
  'unregistered',
  NOW()
    - (((gs * 17)::bigint % 200) * INTERVAL '1 day')
    - ((gs * 131) % 86400) * INTERVAL '1 second'
FROM generate_series(1, 150) AS gs
ON CONFLICT (code) DO NOTHING;

-- B) Пользователи: 40 шт., created_at/updated_at за ~100 дней с разным шагом
INSERT INTO users (display_id, phone, first_name, last_name, device_id, created_at, updated_at)
SELECT
  gen_random_uuid(),
  '+99688' || LPAD((200000 + gs)::text, 6, '0'),
  (ARRAY['Алексей', 'Мария', 'Дмитрий', 'Светлана', 'Иван', 'Алтынай', 'Нурлан', 'Элмира'])[1 + (gs % 8)],
  'Клиент ' || gs::text,
  'seed-007-user-' || gs::text,
  (NOW() - INTERVAL '110 days')
    + (gs * INTERVAL '2 days 5 hours')
    + ((gs * 23) % 36) * INTERVAL '1 hour',
  (NOW() - INTERVAL '110 days')
    + (gs * INTERVAL '2 days 5 hours')
    + ((gs * 23) % 36) * INTERVAL '1 hour'
FROM generate_series(1, 40) AS gs
ON CONFLICT (phone) DO NOTHING;

-- C) По одной машине на каждого пользователя seed-007 (разное время относительно пользователя)
INSERT INTO vehicles (display_id, user_id, plate_number, car_model, is_public, reviews_enabled, telegram_enabled, created_at, updated_at)
SELECT
  gen_random_uuid(),
  u.display_id,
  LPAD(((ROW_NUMBER() OVER (ORDER BY u.created_at, u.display_id)) % 89 + 10)::text, 2, '0')
    || 'KG'
    || LPAD((500 + ROW_NUMBER() OVER (ORDER BY u.created_at, u.display_id))::text, 3, '0')
    || 'ZZ',
  (ARRAY[
    'Toyota Camry', 'Kia K5', 'Hyundai Tucson', 'BMW X3', 'Mercedes GLC',
    'Lexus RX', 'Mazda 6', 'VW Tiguan', 'Skoda Octavia', 'Nissan Qashqai'
  ])[1 + ((ROW_NUMBER() OVER (ORDER BY u.created_at, u.display_id)) % 10)],
  TRUE,
  TRUE,
  FALSE,
  u.created_at + ((ROW_NUMBER() OVER (ORDER BY u.created_at, u.display_id)) % 48) * INTERVAL '1 hour',
  u.created_at + ((ROW_NUMBER() OVER (ORDER BY u.created_at, u.display_id)) % 48) * INTERVAL '1 hour'
FROM users u
WHERE u.device_id LIKE 'seed-007-user-%'
  AND NOT EXISTS (SELECT 1 FROM vehicles v WHERE v.user_id = u.display_id);

-- D) Первые 35 по времени машин связываем с QR S7* (active, разный registered_at)
WITH ranked_veh AS (
  SELECT
    v.display_id AS vehicle_id,
    ROW_NUMBER() OVER (ORDER BY v.created_at, v.display_id) AS rn
  FROM vehicles v
  JOIN users u ON u.display_id = v.user_id
  WHERE u.device_id LIKE 'seed-007-user-%'
),
ranked_qr AS (
  SELECT
    q.id,
    ROW_NUMBER() OVER (ORDER BY q.created_at, q.code) AS rn
  FROM qr_codes q
  WHERE q.status = 'unregistered'
    AND q.vehicle_id IS NULL
    AND q.code LIKE 'S7%'
)
UPDATE qr_codes q
SET
  vehicle_id    = rv.vehicle_id,
  status        = 'active',
  registered_at = LEAST(
    NOW(),
    q.created_at
      + INTERVAL '12 hours'
      + (((rv.rn * 97) % 40) * INTERVAL '1 day')
      + (((rv.rn * 13) % 72) * INTERVAL '1 hour')
  )
FROM ranked_veh rv
JOIN ranked_qr rq ON rq.rn = rv.rn
WHERE q.id = rq.id
  AND rv.rn <= 35;
