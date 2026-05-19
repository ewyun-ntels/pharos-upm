-- +goose Up
ALTER TABLE dashboard ADD COLUMN annotations TEXT NOT NULL DEFAULT '[]';

-- +goose Down
ALTER TABLE dashboard DROP COLUMN annotations;
