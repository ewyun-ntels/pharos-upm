-- +goose Up
RENAME TABLE history_alert_row TO history_alert_row_old;

CREATE TABLE IF NOT EXISTS history_alert_row
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
    mask                 Bool,
    evaluation_epoch     Nullable(Int64),
    dedup_key            String
) engine = ReplacingMergeTree(version)
    ORDER BY dedup_key
    TTL timestamp + toIntervalDay(7);

INSERT INTO history_alert_row (
    timestamp, id, alert_id, name, alert_type, description,
    previous_severity, previous_value, severity, value, labels, status,
    previous_timestamp, version, start_timestamp, status_change_reason,
    status_changed_by, mask, evaluation_epoch, dedup_key
)
SELECT
    timestamp, id, alert_id, name, alert_type, description,
    previous_severity, previous_value, severity, value, labels, status,
    previous_timestamp, version, start_timestamp, status_change_reason,
    status_changed_by, mask, NULL, concat('legacy:', toString(generateUUIDv4()))
FROM history_alert_row_old;

DROP TABLE history_alert_row_old;

-- +goose Down
RENAME TABLE history_alert_row TO history_alert_row_new;

CREATE TABLE IF NOT EXISTS history_alert_row
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

INSERT INTO history_alert_row (
    timestamp, id, alert_id, name, alert_type, description,
    previous_severity, previous_value, severity, value, labels, status,
    previous_timestamp, version, start_timestamp, status_change_reason,
    status_changed_by, mask
)
SELECT
    timestamp, id, alert_id, name, alert_type, description,
    previous_severity, previous_value, severity, value, labels, status,
    previous_timestamp, version, start_timestamp, status_change_reason,
    status_changed_by, mask
FROM history_alert_row_new FINAL;

DROP TABLE history_alert_row_new;
