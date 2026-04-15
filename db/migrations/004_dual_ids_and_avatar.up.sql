-- Add dual IDs (bigint PK + UUID display_id) for users, vehicles, qr_codes.
-- API keeps using UUID (display_id). Bigint id is for internal/admin display.

ALTER TABLE users ADD COLUMN IF NOT EXISTS avatar_url TEXT;

-- If already migrated, exit early.
DO $$
BEGIN
  IF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='users' AND column_name='display_id') THEN
    RETURN;
  END IF;
END $$;

-- Drop FKs that depend on current UUID PKs
ALTER TABLE vehicles DROP CONSTRAINT IF EXISTS vehicles_user_id_fkey;
ALTER TABLE social_profiles DROP CONSTRAINT IF EXISTS social_profiles_user_id_fkey;
ALTER TABLE qr_codes DROP CONSTRAINT IF EXISTS qr_codes_created_by_fkey;
ALTER TABLE telegram_accounts DROP CONSTRAINT IF EXISTS telegram_accounts_user_id_fkey;

ALTER TABLE messages DROP CONSTRAINT IF EXISTS messages_vehicle_id_fkey;
ALTER TABLE qr_codes DROP CONSTRAINT IF EXISTS qr_codes_vehicle_id_fkey;
ALTER TABLE ratings DROP CONSTRAINT IF EXISTS ratings_vehicle_id_fkey;

ALTER TABLE messages DROP CONSTRAINT IF EXISTS messages_qr_code_id_fkey;
ALTER TABLE ratings DROP CONSTRAINT IF EXISTS ratings_qr_code_id_fkey;
ALTER TABLE scan_events DROP CONSTRAINT IF EXISTS scan_events_qr_code_id_fkey;

-- users: uuid id -> display_id, new bigint id PK
ALTER TABLE users RENAME COLUMN id TO display_id;
ALTER TABLE users ADD COLUMN id BIGSERIAL;
ALTER TABLE users DROP CONSTRAINT IF EXISTS users_pkey;
ALTER TABLE users ADD CONSTRAINT users_pkey PRIMARY KEY (id);
ALTER TABLE users ADD CONSTRAINT users_display_id_key UNIQUE (display_id);

-- vehicles: uuid id -> display_id, keep user_id UUID referencing users.display_id
ALTER TABLE vehicles RENAME COLUMN id TO display_id;
ALTER TABLE vehicles ADD COLUMN id BIGSERIAL;
ALTER TABLE vehicles DROP CONSTRAINT IF EXISTS vehicles_pkey;
ALTER TABLE vehicles ADD CONSTRAINT vehicles_pkey PRIMARY KEY (id);
ALTER TABLE vehicles ADD CONSTRAINT vehicles_display_id_key UNIQUE (display_id);

-- qr_codes: uuid id -> display_id, keep vehicle_id UUID referencing vehicles.display_id and created_by UUID referencing users.display_id
ALTER TABLE qr_codes RENAME COLUMN id TO display_id;
ALTER TABLE qr_codes ADD COLUMN id BIGSERIAL;
ALTER TABLE qr_codes DROP CONSTRAINT IF EXISTS qr_codes_pkey;
ALTER TABLE qr_codes ADD CONSTRAINT qr_codes_pkey PRIMARY KEY (id);
ALTER TABLE qr_codes ADD CONSTRAINT qr_codes_display_id_key UNIQUE (display_id);

-- Re-create FKs against UUID display_id columns (API identifiers)
ALTER TABLE vehicles
  ADD CONSTRAINT vehicles_user_id_fkey
  FOREIGN KEY (user_id) REFERENCES users(display_id) ON DELETE CASCADE;

ALTER TABLE social_profiles
  ADD CONSTRAINT social_profiles_user_id_fkey
  FOREIGN KEY (user_id) REFERENCES users(display_id) ON DELETE CASCADE;

ALTER TABLE qr_codes
  ADD CONSTRAINT qr_codes_created_by_fkey
  FOREIGN KEY (created_by) REFERENCES users(display_id);

ALTER TABLE telegram_accounts
  ADD CONSTRAINT telegram_accounts_user_id_fkey
  FOREIGN KEY (user_id) REFERENCES users(display_id) ON DELETE CASCADE;

ALTER TABLE messages
  ADD CONSTRAINT messages_vehicle_id_fkey
  FOREIGN KEY (vehicle_id) REFERENCES vehicles(display_id);

ALTER TABLE qr_codes
  ADD CONSTRAINT qr_codes_vehicle_id_fkey
  FOREIGN KEY (vehicle_id) REFERENCES vehicles(display_id) ON DELETE SET NULL;

ALTER TABLE ratings
  ADD CONSTRAINT ratings_vehicle_id_fkey
  FOREIGN KEY (vehicle_id) REFERENCES vehicles(display_id) ON DELETE CASCADE;

ALTER TABLE messages
  ADD CONSTRAINT messages_qr_code_id_fkey
  FOREIGN KEY (qr_code_id) REFERENCES qr_codes(display_id);

ALTER TABLE ratings
  ADD CONSTRAINT ratings_qr_code_id_fkey
  FOREIGN KEY (qr_code_id) REFERENCES qr_codes(display_id) ON DELETE SET NULL;

ALTER TABLE scan_events
  ADD CONSTRAINT scan_events_qr_code_id_fkey
  FOREIGN KEY (qr_code_id) REFERENCES qr_codes(display_id);

