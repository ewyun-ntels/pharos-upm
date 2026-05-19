-- +goose Up
ALTER TABLE alert_status ADD COLUMN id TEXT;
ALTER TABLE alert_status ADD COLUMN status TEXT;
DELETE FROM alert_status WHERE 1=1;
-- UPDATE alert_status SET status = 'alerting';

-- +goose Down
ALTER TABLE alert_status DROP COLUMN status;
ALTER TABLE alert_status DROP COLUMN id;

SELECT * FROM alert_status;