-- +goose Up
CREATE TABLE IF NOT EXISTS groups
(
    name TEXT PRIMARY KEY NOT NULL
);

-- +goose Down
DROP TABLE IF EXISTS groups;
