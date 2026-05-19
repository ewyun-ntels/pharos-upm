-- +goose Up
ALTER TABLE history_login ADD COLUMN user_agent Nullable(String);

-- +goose Down
ALTER TABLE history_login DROP COLUMN user_agent;
