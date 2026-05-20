-- +goose Up
ALTER TABLE history_alert_row ADD COLUMN evaluation_epoch BIGINT NULL;

CREATE UNIQUE INDEX IF NOT EXISTS ux_history_alert_row_auto_epoch
ON history_alert_row (id, status, severity, evaluation_epoch)
WHERE evaluation_epoch IS NOT NULL AND status_change_reason = 'auto';

-- +goose Down
DROP INDEX IF EXISTS ux_history_alert_row_auto_epoch;
ALTER TABLE history_alert_row DROP COLUMN evaluation_epoch;
