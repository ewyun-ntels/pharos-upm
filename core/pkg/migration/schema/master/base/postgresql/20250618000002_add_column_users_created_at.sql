-- +goose Up
ALTER TABLE users ADD COLUMN created_at TIMESTAMP;
UPDATE users SET created_at = CURRENT_TIMESTAMP;

-- +goose Down
ALTER TABLE users DROP COLUMN created_at;
