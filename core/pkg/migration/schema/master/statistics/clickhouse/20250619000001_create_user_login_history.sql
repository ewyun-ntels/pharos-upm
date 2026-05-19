-- +goose Up
DROP TABLE IF EXISTS history_login;
CREATE TABLE IF NOT EXISTS history_login
(
    session_id String,
    user_id    String,
    client_ip  Nullable(String),
    started_at DateTime,
    ended_at   Nullable(DateTime),
    end_reason Nullable(String),
    version    DateTime DEFAULT now()
) ENGINE = ReplacingMergeTree(version)
ORDER BY session_id;

-- +goose Down
DROP TABLE IF EXISTS history_login;
