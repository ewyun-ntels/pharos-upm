-- +goose Up
DROP TABLE IF EXISTS groups;

-- +goose Down
CREATE TABLE IF NOT EXISTS groups
(
    name TEXT PRIMARY KEY NOT NULL
);
