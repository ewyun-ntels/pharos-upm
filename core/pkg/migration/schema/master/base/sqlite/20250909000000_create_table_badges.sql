-- +goose Up
CREATE TABLE IF NOT EXISTS badges (
    name TEXT PRIMARY KEY,
    datasource TEXT NOT NULL,
    query_json TEXT NOT NULL,
    column_name TEXT NOT NULL,
    updated_at TIMESTAMP NOT NULL
);

-- +goose Down
DROP TABLE IF EXISTS badges;
