-- +goose Up
ALTER TABLE plugin_datasource RENAME TO plugin_datasource_old;

CREATE TABLE IF NOT EXISTS plugin_datasource
(
    name        TEXT PRIMARY KEY NOT NULL,
    type        TEXT NOT NULL,
    data        JSONB,
    secure_data JSONB
);

INSERT INTO plugin_datasource SELECT name, type, data, secure_data FROM plugin_datasource_old;

DROP TABLE IF EXISTS plugin_datasource_old;

-- +goose Down
ALTER TABLE plugin_datasource RENAME TO plugin_datasource_old;

CREATE TABLE IF NOT EXISTS plugin_datasource
(
    id          TEXT NOT NULL,
    type        TEXT NOT NULL,
    name        TEXT NOT NULL,
    data        JSONB,
    secure_data JSONB,

    PRIMARY KEY (name, type)
);

INSERT INTO plugin_datasource SELECT concat(type, name), type, name, data, secure_data FROM plugin_datasource_old;

DROP TABLE IF EXISTS plugin_datasource_old;
