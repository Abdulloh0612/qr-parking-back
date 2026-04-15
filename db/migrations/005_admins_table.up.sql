CREATE TABLE IF NOT EXISTS admins (
    id           BIGSERIAL PRIMARY KEY,
    display_id   UUID        NOT NULL DEFAULT gen_random_uuid() UNIQUE,
    username     VARCHAR(64) NOT NULL UNIQUE,
    password_hash TEXT       NOT NULL,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at   TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Migrate existing admins from users table
INSERT INTO admins (display_id, username, password_hash, created_at, updated_at)
SELECT display_id, admin_username, password_hash, created_at, updated_at
FROM users
WHERE is_admin = TRUE
  AND admin_username IS NOT NULL
  AND password_hash IS NOT NULL
ON CONFLICT (username) DO NOTHING;
