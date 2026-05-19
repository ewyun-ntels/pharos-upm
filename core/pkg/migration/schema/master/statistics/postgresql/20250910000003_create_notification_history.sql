-- +goose Up
CREATE TABLE IF NOT EXISTS history_notification (
    timestamp   TIMESTAMP,
    id          TEXT,
    alert_id    TEXT NULL,
    name        TEXT NULL,
    alert_type  TEXT NULL,
    description TEXT NULL,
    severity    TEXT NULL,
    value       DOUBLE PRECISION NULL,
    labels      TEXT NULL,
    status      TEXT NULL
);
CREATE INDEX IF NOT EXISTS idx_history_notification_timestamp ON history_notification (timestamp);

-- +goose Down
DROP TABLE IF EXISTS history_notification;
