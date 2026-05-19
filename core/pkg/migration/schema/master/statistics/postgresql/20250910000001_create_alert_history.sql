-- +goose Up
CREATE TABLE IF NOT EXISTS history_alert (
    timestamp           TIMESTAMP,
    id                  TEXT,
    alert_id            TEXT NULL,
    name                TEXT NULL,
    alert_type          TEXT NULL,
    description         TEXT NULL,
    previous_severity   TEXT NULL,
    previous_value      DOUBLE PRECISION NULL,
    severity            TEXT NULL,
    value               DOUBLE PRECISION NULL,
    labels              TEXT NULL,
    status              TEXT NULL,
    previous_timestamp  TIMESTAMP NULL,
    version             TIMESTAMP
);
CREATE INDEX IF NOT EXISTS idx_history_alert_timestamp ON history_alert (timestamp);
CREATE UNIQUE INDEX IF NOT EXISTS ux_history_alert_id ON history_alert (id);

-- +goose Down
DROP TABLE IF EXISTS history_alert;
DROP INDEX IF EXISTS ux_history_alert_id;
