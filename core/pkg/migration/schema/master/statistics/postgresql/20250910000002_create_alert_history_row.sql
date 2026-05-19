-- +goose Up
CREATE TABLE IF NOT EXISTS history_alert_row (
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
    version             TIMESTAMP,
    start_timestamp     TIMESTAMP,
    status_change_reason  TEXT NULL,
    status_changed_by     TEXT NULL,
    mask                BOOLEAN
);
CREATE INDEX IF NOT EXISTS idx_history_alert_row_timestamp ON history_alert_row (timestamp);

-- +goose Down
DROP TABLE IF EXISTS history_alert_row;
