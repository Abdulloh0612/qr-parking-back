CREATE EXTENSION IF NOT EXISTS "pgcrypto";

CREATE TABLE users (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    phone       VARCHAR(20) UNIQUE NOT NULL,
    first_name  VARCHAR(100) NOT NULL,
    last_name   VARCHAR(100) NOT NULL,
    is_admin    BOOLEAN DEFAULT FALSE,
    created_at  TIMESTAMPTZ DEFAULT NOW(),
    updated_at  TIMESTAMPTZ DEFAULT NOW()
);

CREATE TABLE vehicles (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id         UUID REFERENCES users(id) ON DELETE CASCADE,
    plate_number    VARCHAR(20) NOT NULL,
    car_model       VARCHAR(100) NOT NULL,
    is_public       BOOLEAN DEFAULT FALSE,
    created_at      TIMESTAMPTZ DEFAULT NOW(),
    updated_at      TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_vehicles_user ON vehicles(user_id);

CREATE TABLE social_profiles (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id     UUID REFERENCES users(id) ON DELETE CASCADE,
    platform    VARCHAR(30) NOT NULL,
    handle      VARCHAR(255) NOT NULL,
    is_public   BOOLEAN DEFAULT TRUE,
    created_at  TIMESTAMPTZ DEFAULT NOW()
);

CREATE UNIQUE INDEX idx_social_user_platform ON social_profiles(user_id, platform);

CREATE TABLE qr_codes (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    code          VARCHAR(32) UNIQUE NOT NULL,
    vehicle_id    UUID REFERENCES vehicles(id) ON DELETE SET NULL,
    status        VARCHAR(20) DEFAULT 'unregistered',
    created_by    UUID REFERENCES users(id),
    registered_at TIMESTAMPTZ,
    created_at    TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_qr_code ON qr_codes(code);
CREATE INDEX idx_qr_status ON qr_codes(status);

CREATE TABLE scan_events (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    qr_code_id  UUID REFERENCES qr_codes(id),
    ip_address  INET,
    user_agent  TEXT,
    scanned_at  TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_scan_qr ON scan_events(qr_code_id);
CREATE INDEX idx_scan_time ON scan_events(scanned_at DESC);

CREATE TABLE messages (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    qr_code_id      UUID REFERENCES qr_codes(id),
    vehicle_id      UUID REFERENCES vehicles(id),
    content         TEXT NOT NULL,
    sender_name     VARCHAR(100),
    sender_phone    VARCHAR(20),
    is_delivered    BOOLEAN DEFAULT FALSE,
    delivered_at    TIMESTAMPTZ,
    created_at      TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_messages_vehicle ON messages(vehicle_id);

CREATE TABLE telegram_accounts (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id     UUID REFERENCES users(id) ON DELETE CASCADE UNIQUE,
    tg_user_id  BIGINT UNIQUE NOT NULL,
    tg_username VARCHAR(100),
    linked_at   TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_tg_user ON telegram_accounts(user_id);
