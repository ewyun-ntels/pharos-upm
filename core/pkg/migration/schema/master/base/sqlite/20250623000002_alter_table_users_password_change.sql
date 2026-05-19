-- +goose Up
ALTER TABLE users ADD COLUMN password_updated_at DATETIME;
UPDATE users SET password_updated_at = CURRENT_TIMESTAMP;

-- +goose Down
ALTER TABLE users DROP COLUMN password_updated_at;
