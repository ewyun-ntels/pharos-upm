-- +goose Up
CREATE TABLE IF NOT EXISTS history_notification (
    timestamp   DATETIME,
    id          TEXT,
    alert_id    TEXT,
    name        TEXT,
    alert_type  TEXT,
    description TEXT,
    severity    TEXT,
    value       REAL,
    labels      TEXT,
    status      TEXT
);
CREATE INDEX IF NOT EXISTS idx_history_notification_timestamp ON history_notification (timestamp);

-- +goose Down
DROP TABLE IF EXISTS history_notification;
