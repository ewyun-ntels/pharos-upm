-- +goose Up
CREATE TABLE IF NOT EXISTS plugin_datasource
(
    id          TEXT NOT NULL,
    type        TEXT NOT NULL,
    name        TEXT NOT NULL,
    data        JSONB,
    secure_data JSONB,

    PRIMARY KEY (name, type)
);

-- +goose Down
DROP TABLE IF EXISTS plugin_datasource;
