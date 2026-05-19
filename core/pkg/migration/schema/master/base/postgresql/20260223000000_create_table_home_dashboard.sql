-- +goose Up
CREATE TABLE IF NOT EXISTS home_dashboard
(
    dashboard_id TEXT NOT NULL
);

-- +goose Down
DROP TABLE IF EXISTS home_dashboard;
