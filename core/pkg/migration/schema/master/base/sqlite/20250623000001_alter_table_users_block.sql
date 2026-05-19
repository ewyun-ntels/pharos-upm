-- +goose Up
ALTER TABLE users ADD COLUMN blocked BOOLEAN;
UPDATE users SET blocked = false;

-- +goose Down
ALTER TABLE users DROP COLUMN blocked;
