-- Best-effort rollback for 004 (drops bigint id PK and renames display_id back to id).
-- NOTE: This is destructive if data was created after migration.

DO $$
BEGIN
  -- qr_codes
  IF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='qr_codes' AND column_name='display_id') THEN
    ALTER TABLE qr_codes DROP CONSTRAINT IF EXISTS qr_codes_pkey;
    ALTER TABLE qr_codes DROP CONSTRAINT IF EXISTS qr_codes_display_id_key;
    ALTER TABLE qr_codes ADD CONSTRAINT qr_codes_pkey PRIMARY KEY (display_id);
    ALTER TABLE qr_codes DROP COLUMN IF EXISTS id;
    ALTER TABLE qr_codes RENAME COLUMN display_id TO id;
  END IF;

  -- vehicles
  IF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='vehicles' AND column_name='display_id') THEN
    ALTER TABLE vehicles DROP CONSTRAINT IF EXISTS vehicles_pkey;
    ALTER TABLE vehicles DROP CONSTRAINT IF EXISTS vehicles_display_id_key;
    ALTER TABLE vehicles ADD CONSTRAINT vehicles_pkey PRIMARY KEY (display_id);
    ALTER TABLE vehicles DROP COLUMN IF EXISTS id;
    ALTER TABLE vehicles RENAME COLUMN display_id TO id;
  END IF;

  -- users
  IF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='users' AND column_name='display_id') THEN
    ALTER TABLE users DROP CONSTRAINT IF EXISTS users_pkey;
    ALTER TABLE users DROP CONSTRAINT IF EXISTS users_display_id_key;
    ALTER TABLE users ADD CONSTRAINT users_pkey PRIMARY KEY (display_id);
    ALTER TABLE users DROP COLUMN IF EXISTS id;
    ALTER TABLE users RENAME COLUMN display_id TO id;
  END IF;
END $$;

ALTER TABLE users DROP COLUMN IF EXISTS avatar_url;

