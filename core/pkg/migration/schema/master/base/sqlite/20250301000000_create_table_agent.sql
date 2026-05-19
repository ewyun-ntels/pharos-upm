-- +goose Up
-- TODO index처리는 일단 보류 (추후 성능 이슈 발생시 고려)
CREATE TABLE IF NOT EXISTS agent
(
    host           text      not null primary key,
    url            text      not null,
    health         text      not null,
    version        text      not null,
    server_version text      not null,
    is_in_recovery boolean   not null,
    uptime         timestamp not null,
    created_at     timestamp not null,
    updated_at     timestamp not null
);

-- +goose Down
DROP TABLE IF EXISTS agent;
