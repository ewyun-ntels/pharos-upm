-- +goose Up
ALTER TABLE history_login ADD COLUMN user_agent TEXT;

-- +goose Down
ALTER TABLE history_login DROP COLUMN user_agent;
