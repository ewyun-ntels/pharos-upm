-- +goose Up
CREATE TABLE IF NOT EXISTS history_alert (
    timestamp           DATETIME,
    id                  TEXT,
    alert_id            TEXT,
    name                TEXT,
    alert_type          TEXT,
    description         TEXT,
    previous_severity   TEXT,
    previous_value      REAL,
    severity            TEXT,
    value               REAL,
    labels              TEXT,
    status              TEXT,
    previous_timestamp  DATETIME,
    version             DATETIME
);
CREATE INDEX IF NOT EXISTS idx_history_alert_timestamp ON history_alert (timestamp);
CREATE UNIQUE INDEX IF NOT EXISTS ux_history_alert_id ON history_alert (id);

-- +goose Down
DROP TABLE IF EXISTS history_alert;
DROP INDEX IF EXISTS ux_history_alert_id;
