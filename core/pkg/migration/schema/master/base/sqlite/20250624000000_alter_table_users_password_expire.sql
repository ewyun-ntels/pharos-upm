-- +goose Up
ALTER TABLE users ADD COLUMN password_expired_at DATETIME;
UPDATE users SET password_expired_at = null;

-- +goose Down
ALTER TABLE users DROP COLUMN password_expired_at;
