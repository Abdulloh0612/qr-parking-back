-- Seed: Admin user
-- Password: admin123 (bcrypt hash)
-- To regenerate: go run cmd/hashpass/main.go admin123
INSERT INTO admins (username, password_hash)
VALUES (
    'admin',
    '$2a$10$RI0Fb1CSWAjKugLr4FuYuuTQh0q3rtXn37NXPyIFPb83bRMK2QPiO'  -- admin123
)
ON CONFLICT (username) DO NOTHING;
