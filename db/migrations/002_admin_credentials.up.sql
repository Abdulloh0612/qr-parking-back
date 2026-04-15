ALTER TABLE users ADD COLUMN IF NOT EXISTS admin_username VARCHAR(64) UNIQUE;
ALTER TABLE users ADD COLUMN IF NOT EXISTS password_hash TEXT;

COMMENT ON COLUMN users.admin_username IS 'Логин для входа в админ-панель (только при is_admin)';
COMMENT ON COLUMN users.password_hash IS 'bcrypt-хеш пароля администратора';
