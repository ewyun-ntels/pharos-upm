-- +goose Up
UPDATE users
SET extra = '{"role:super_admin":true}'
WHERE extra = '{"super_admin": true}';
-- +goose Down
UPDATE users
SET extra = '{"super_admin": true}'
WHERE extra = '{"role:super_admin":true}';
