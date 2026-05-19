-- +goose Up
CREATE TABLE IF NOT EXISTS history_requests
(
    id               String,
    timestamp        TIMESTAMP,
    `user`           String,
    method           String,
    path             String,
    query            String,
    request_headers  String,
    response_headers String,
    status_code      Int32,
    duration         Int64,
    request_body     String,
    response_body    String,
    error            String,
    middleware_trace String,
    route_trace      String,
    created_at       DateTime DEFAULT now()
    ) ENGINE = MergeTree()
    PARTITION BY toYYYYMMDD(timestamp)
    ORDER BY (id, timestamp)
    TTL timestamp + INTERVAL 7 DAY;

-- +goose Down
DROP TABLE IF EXISTS history_requests;
