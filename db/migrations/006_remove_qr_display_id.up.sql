BEGIN;

-- ── scan_events: change qr_code_id from UUID (→qr_codes.display_id) to VARCHAR (→qr_codes.code) ──
ALTER TABLE scan_events DROP CONSTRAINT IF EXISTS scan_events_qr_code_id_fkey;
ALTER TABLE scan_events ADD COLUMN qr_code_code VARCHAR(32);
UPDATE scan_events se
    SET qr_code_code = q.code
    FROM qr_codes q
    WHERE q.display_id = se.qr_code_id;
ALTER TABLE scan_events DROP COLUMN qr_code_id;
ALTER TABLE scan_events RENAME COLUMN qr_code_code TO qr_code_id;
ALTER TABLE scan_events ADD CONSTRAINT scan_events_qr_code_id_fkey
    FOREIGN KEY (qr_code_id) REFERENCES qr_codes(code) ON DELETE CASCADE;
CREATE INDEX IF NOT EXISTS idx_scan_qr_code ON scan_events(qr_code_id);

-- ── messages: change qr_code_id from UUID (→qr_codes.display_id) to VARCHAR (→qr_codes.code) ──
ALTER TABLE messages DROP CONSTRAINT IF EXISTS messages_qr_code_id_fkey;
ALTER TABLE messages ADD COLUMN qr_code_code VARCHAR(32);
UPDATE messages m
    SET qr_code_code = q.code
    FROM qr_codes q
    WHERE q.display_id = m.qr_code_id;
ALTER TABLE messages DROP COLUMN qr_code_id;
ALTER TABLE messages RENAME COLUMN qr_code_code TO qr_code_id;
ALTER TABLE messages ADD CONSTRAINT messages_qr_code_id_fkey
    FOREIGN KEY (qr_code_id) REFERENCES qr_codes(code) ON DELETE SET NULL;

-- ── qr_codes: drop display_id ──────────────────────────────────────────────────
ALTER TABLE qr_codes DROP CONSTRAINT IF EXISTS qr_codes_display_id_key;
ALTER TABLE qr_codes DROP COLUMN IF EXISTS display_id;

COMMIT;
