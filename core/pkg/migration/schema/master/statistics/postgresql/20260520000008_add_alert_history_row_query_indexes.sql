-- +goose Up
CREATE INDEX IF NOT EXISTS ix_history_alert_row_timestamp
ON history_alert_row (timestamp DESC);

CREATE INDEX IF NOT EXISTS ix_history_alert_row_name_timestamp
ON history_alert_row (name, timestamp DESC);

-- +goose Down
DROP INDEX IF EXISTS ix_history_alert_row_name_timestamp;
DROP INDEX IF EXISTS ix_history_alert_row_timestamp;
