-- +goose Up
create table IF NOT EXISTS history_alert_row
(
    timestamp            DateTime,
    id                   String,
    alert_id             Nullable(String),
    name                 Nullable(String),
    alert_type           Nullable(String),
    description          Nullable(String),
    previous_severity    Nullable(String),
    previous_value       Nullable(Float64),
    severity             Nullable(String),
    value                Nullable(Float64),
    labels               Nullable(String),
    status               Nullable(String),
    previous_timestamp   Nullable(DateTime),
    version              DateTime,
    start_timestamp      DateTime,
    status_change_reason Nullable(String),
    status_changed_by    Nullable(String),
    mask                 Bool
) engine = MergeTree ORDER BY timestamp
        TTL timestamp + toIntervalDay(7);

-- +goose Down
DROP TABLE IF EXISTS history_alert_row;
