-- +goose Up
CREATE TABLE IF NOT EXISTS workflow_execution
(
    id          UUID PRIMARY KEY NOT NULL,
    job         TEXT NOT NULL,
    start_ts    TIMESTAMP,
    modified_ts TIMESTAMP,
    state       TEXT,
    tasks       JSONB
);

-- +goose Down
DROP TABLE IF EXISTS workflow_execution;
