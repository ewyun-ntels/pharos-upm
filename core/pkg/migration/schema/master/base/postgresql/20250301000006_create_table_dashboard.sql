-- +goose Up
CREATE TABLE IF NOT EXISTS dashboard
(
    id     TEXT PRIMARY KEY NOT NULL,
    config JSONB NOT NULL
);

-- +goose Down
DROP TABLE IF EXISTS dashboard;
