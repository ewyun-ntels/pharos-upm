-- +goose Up
ALTER TABLE dashboard
    ADD COLUMN IF NOT EXISTS folder_id VARCHAR(255);

-- +goose Down
ALTER TABLE dashboard DROP COLUMN IF EXISTS folder_id;
