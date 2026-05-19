-- +goose Up
ALTER TABLE dashboard ADD COLUMN folder_id TEXT;

-- +goose Down
ALTER TABLE dashboard DROP COLUMN folder_id;
