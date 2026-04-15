-- Seed: default admin account
-- Login: admin / Password: admin123
INSERT INTO admins (display_id, username, password_hash)
VALUES (
    gen_random_uuid(),
    'admin',
    '$2a$10$IrDGB9CdxkOpNHEvjyB80.2fzTCrIJNm2iUqwnVTFeIJC1tYBperi'
)
ON CONFLICT (username) DO NOTHING;
