-- +goose Up
CREATE TABLE IF NOT EXISTS workflow_job
(
    name       TEXT PRIMARY KEY NOT NULL,
    serve_type TEXT NOT NULL,
    hosts      JSONB NOT NULL,
    schedule   TEXT NOT NULL,
    active     BOOLEAN NOT NULL,
    task_names JSONB NOT NULL,
    task_datas JSONB NOT NULL,
    updated_at DATETIME NOT NULL
);

-- +goose Down
DROP TABLE IF EXISTS workflow_job;
