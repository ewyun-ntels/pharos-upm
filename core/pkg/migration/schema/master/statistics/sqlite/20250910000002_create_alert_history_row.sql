-- +goose Up
CREATE TABLE IF NOT EXISTS history_alert_row (
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
    version             DATETIME,
    start_timestamp     DATETIME,
    status_change_reason  TEXT,
    status_changed_by     TEXT,
    mask                BOOLEAN
);
CREATE INDEX IF NOT EXISTS idx_history_alert_row_timestamp ON history_alert_row (timestamp);

-- +goose Down
DROP TABLE IF EXISTS history_alert_row;
