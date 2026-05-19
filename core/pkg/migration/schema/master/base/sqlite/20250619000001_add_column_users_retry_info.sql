-- +goose Up
ALTER TABLE users ADD COLUMN retry REAL;
ALTER TABLE users ADD COLUMN retry_at DATETIME;

UPDATE users SET retry = 0 WHERE retry IS NULL;

-- +goose Down
ALTER TABLE users DROP COLUMN retry;
ALTER TABLE users DROP COLUMN retry_at;
