-- Откат seed 007: отвязать и удалить созданные записи

UPDATE qr_codes
SET vehicle_id = NULL, status = 'unregistered', registered_at = NULL
WHERE code LIKE 'S7%';

DELETE FROM vehicles
WHERE user_id IN (
  SELECT display_id FROM users WHERE device_id LIKE 'seed-007-user-%'
);

DELETE FROM users WHERE device_id LIKE 'seed-007-user-%';

DELETE FROM qr_codes WHERE code LIKE 'S7%';
