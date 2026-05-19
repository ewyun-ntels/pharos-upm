-- +goose Up
CREATE TABLE IF NOT EXISTS alert_rule
(
    id         UUID PRIMARY KEY NOT NULL,
    alert_type TEXT NOT NULL,
    name       TEXT NOT NULL UNIQUE,
    rule       TEXT NOT NULL,
    timestamp  TIMESTAMP NOT NULL,
    updated_at TIMESTAMP NOT NULL
);

CREATE TABLE IF NOT EXISTS alert_status
(
    alert_id          TEXT NOT NULL,
    name              TEXT NOT NULL,
    mask              TEXT,
    alert_type        TEXT,
    description       TEXT,
    severity          TEXT NOT NULL,
    previous_severity TEXT,
    previous_value    REAL,
    value             REAL,
    labels            TEXT,
    timestamp         TIMESTAMP NOT NULL,
    updated_at        TIMESTAMP NOT NULL,
    check_time        TIMESTAMP NOT NULL,
    PRIMARY KEY (alert_id, name)
);

-- +goose Down
DROP TABLE IF EXISTS alert_rule;
DROP TABLE IF EXISTS alert_status;
