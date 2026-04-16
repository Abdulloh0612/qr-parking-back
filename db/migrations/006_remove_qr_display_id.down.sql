BEGIN;

-- Restore display_id on qr_codes
ALTER TABLE qr_codes ADD COLUMN IF NOT EXISTS display_id UUID DEFAULT gen_random_uuid() UNIQUE;

-- messages: revert qr_code_id back to UUID referencing qr_codes.display_id
ALTER TABLE messages DROP CONSTRAINT IF EXISTS messages_qr_code_id_fkey;
ALTER TABLE messages ADD COLUMN qr_code_uuid UUID;
UPDATE messages m
    SET qr_code_uuid = q.display_id
    FROM qr_codes q
    WHERE q.code = m.qr_code_id;
ALTER TABLE messages DROP COLUMN qr_code_id;
ALTER TABLE messages RENAME COLUMN qr_code_uuid TO qr_code_id;
ALTER TABLE messages ADD CONSTRAINT messages_qr_code_id_fkey
    FOREIGN KEY (qr_code_id) REFERENCES qr_codes(display_id);

-- scan_events: revert qr_code_id back to UUID referencing qr_codes.display_id
ALTER TABLE scan_events DROP CONSTRAINT IF EXISTS scan_events_qr_code_id_fkey;
ALTER TABLE scan_events DROP INDEX IF EXISTS idx_scan_qr_code;
ALTER TABLE scan_events ADD COLUMN qr_code_uuid UUID;
UPDATE scan_events se
    SET qr_code_uuid = q.display_id
    FROM qr_codes q
    WHERE q.code = se.qr_code_id;
ALTER TABLE scan_events DROP COLUMN qr_code_id;
ALTER TABLE scan_events RENAME COLUMN qr_code_uuid TO qr_code_id;
ALTER TABLE scan_events ADD CONSTRAINT scan_events_qr_code_id_fkey
    FOREIGN KEY (qr_code_id) REFERENCES qr_codes(display_id);

COMMIT;
