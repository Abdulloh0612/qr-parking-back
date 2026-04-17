-- Roles: admin (default), super_admin — управление учётками admins только у super_admin.
ALTER TABLE admins
  ADD COLUMN IF NOT EXISTS role VARCHAR(32) NOT NULL DEFAULT 'admin'
  CHECK (role IN ('admin', 'super_admin'));

-- Существующий seeded-логин «admin» — суперпользователь.
UPDATE admins SET role = 'super_admin' WHERE username = 'admin';

-- На случай, если записи с таким логином нет: хотя бы один super_admin.
UPDATE admins SET role = 'super_admin'
WHERE id = (SELECT id FROM admins ORDER BY id ASC LIMIT 1)
  AND NOT EXISTS (SELECT 1 FROM admins WHERE role = 'super_admin');
