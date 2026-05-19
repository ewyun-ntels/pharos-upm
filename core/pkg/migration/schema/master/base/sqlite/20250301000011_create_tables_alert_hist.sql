-- +goose Up
CREATE TABLE IF NOT EXISTS alert_hist
(
    timestamp         DATETIME NOT NULL,
    updated_at        DATETIME NOT NULL,
    alert_id          TEXT NOT NULL,
    name              TEXT NOT NULL,
    alert_type        TEXT,
    description       TEXT,
    previous_severity TEXT NOT NULL,
    previous_value    REAL,
    severity          TEXT NOT NULL,
    value             REAL,
    labels            TEXT
);
CREATE INDEX IF NOT EXISTS idx_alert_hist_timestamp_alert_id ON alert_hist(timestamp, name);

-- +goose Down
DROP TABLE IF EXISTS alert_hist;
