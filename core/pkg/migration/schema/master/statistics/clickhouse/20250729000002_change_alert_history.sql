-- +goose Up
-- table 엔진 변경
CREATE TABLE IF NOT EXISTS history_alert
(
    timestamp          DateTime,
    id                 String,
    alert_id           Nullable(String),
    name               Nullable(String),
    alert_type         Nullable(String),
    description        Nullable(String),
    previous_severity  Nullable(String),
    previous_value     Nullable(Float64),
    severity           Nullable(String),
    value              Nullable(Float64),
    labels             Nullable(String),
    status             Nullable(String),
    previous_timestamp Nullable(TIMESTAMP),
    version          Datetime
) engine = ReplacingMergeTree(version)
    ORDER BY id
    TTL timestamp + toIntervalDay(7);

-- +goose Down
DROP TABLE IF EXISTS history_alert;
