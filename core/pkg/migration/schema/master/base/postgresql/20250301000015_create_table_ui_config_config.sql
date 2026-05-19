-- +goose Up
CREATE TABLE IF NOT EXISTS ui_config_config
(
    config TEXT
);

-- +goose Down
DROP TABLE IF EXISTS ui_config_config;
