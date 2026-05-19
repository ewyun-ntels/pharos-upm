-- +goose Up
CREATE TABLE IF NOT EXISTS history_notification
(
    timestamp          DateTime,
    id                 String,
    alert_id            Nullable(String),
    name               Nullable(String),
    alert_type          Nullable(String),
    description        Nullable(String),
    severity           Nullable(String),
    value              Nullable(Float64),
    labels             Nullable(String),
    status             Nullable(String)
) ENGINE = MergeTree()
    ORDER BY (timestamp)
    TTL timestamp + INTERVAL 7 DAY;

-- +goose Down
DROP TABLE IF EXISTS history_notification;
