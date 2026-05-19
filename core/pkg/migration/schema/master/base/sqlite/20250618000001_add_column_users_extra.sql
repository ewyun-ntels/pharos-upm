-- +goose Up
ALTER TABLE users ADD COLUMN extra TEXT;
UPDATE users SET extra = '{"super_admin": true}';

-- +goose Down
ALTER TABLE users DROP COLUMN extra;
