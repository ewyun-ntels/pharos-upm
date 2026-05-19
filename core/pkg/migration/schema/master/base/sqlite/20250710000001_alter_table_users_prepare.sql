-- +goose Up
ALTER TABLE users ADD COLUMN prepare TEXT;

-- +goose Down
ALTER TABLE users DROP COLUMN prepare;

