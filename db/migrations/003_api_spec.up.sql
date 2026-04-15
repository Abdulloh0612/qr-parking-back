ALTER TABLE users ADD COLUMN IF NOT EXISTS device_id VARCHAR(128) UNIQUE;

ALTER TABLE vehicles ADD COLUMN IF NOT EXISTS photo_url TEXT;
ALTER TABLE vehicles ADD COLUMN IF NOT EXISTS reviews_enabled BOOLEAN DEFAULT TRUE;
ALTER TABLE vehicles ADD COLUMN IF NOT EXISTS telegram_enabled BOOLEAN DEFAULT FALSE;

CREATE TABLE IF NOT EXISTS ratings (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    vehicle_id  UUID NOT NULL REFERENCES vehicles(id) ON DELETE CASCADE,
    qr_code_id  UUID REFERENCES qr_codes(id) ON DELETE SET NULL,
    rating      SMALLINT NOT NULL CHECK (rating >= 1 AND rating <= 5),
    comment     VARCHAR(200),
    created_at  TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_ratings_vehicle ON ratings(vehicle_id);
