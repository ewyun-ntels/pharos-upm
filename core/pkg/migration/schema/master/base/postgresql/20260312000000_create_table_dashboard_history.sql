-- +goose Up
CREATE TABLE IF NOT EXISTS dashboard_history (
    id              VARCHAR(255) PRIMARY KEY,
    dashboard_id    VARCHAR(255) NOT NULL,
    dashboard_title TEXT NOT NULL,
    action          VARCHAR(50) NOT NULL,
    changed_by      TEXT NOT NULL,
    changed_at      TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    config_json     TEXT
);
CREATE INDEX IF NOT EXISTS idx_dashboard_history_changed_at ON dashboard_history(changed_at DESC);

-- +goose Down
DROP INDEX IF EXISTS idx_dashboard_history_changed_at;
DROP TABLE IF EXISTS dashboard_history;
