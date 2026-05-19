-- +goose Up
CREATE TABLE IF NOT EXISTS dashboard_history (
    id              TEXT PRIMARY KEY,
    dashboard_id    TEXT NOT NULL,
    dashboard_title TEXT NOT NULL,
    action          TEXT NOT NULL,
    changed_by      TEXT NOT NULL,
    changed_at      TEXT NOT NULL DEFAULT (datetime('now')),
    config_json     TEXT
);
CREATE INDEX IF NOT EXISTS idx_dashboard_history_changed_at ON dashboard_history(changed_at DESC);

-- +goose Down
DROP INDEX IF EXISTS idx_dashboard_history_changed_at;
DROP TABLE IF EXISTS dashboard_history;
