-- +goose Up
CREATE TABLE IF NOT EXISTS history_requests (
    id                TEXT,
    timestamp         TIMESTAMP,
    "user"            TEXT,
    method            TEXT,
    path              TEXT,
    query             TEXT,
    request_headers   TEXT,
    response_headers  TEXT,
    status_code       INTEGER,
    duration          BIGINT,
    request_body      TEXT,
    response_body     TEXT,
    error             TEXT,
    middleware_trace  TEXT,
    route_trace       TEXT,
    created_at        TIMESTAMP DEFAULT now()
);
CREATE INDEX IF NOT EXISTS idx_history_requests_timestamp ON history_requests (timestamp);

-- +goose Down
DROP TABLE IF EXISTS history_requests;
