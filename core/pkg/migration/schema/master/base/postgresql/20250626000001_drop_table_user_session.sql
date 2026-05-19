-- +goose Up
DROP TABLE IF EXISTS user_session;

-- +goose Down
CREATE TABLE IF NOT EXISTS user_session
(
    token         TEXT PRIMARY KEY NOT NULL,
    refresh_token TEXT,
    username      TEXT,
    expired_at    TIMESTAMP,
    issued_at     TIMESTAMP,
    issuer        TEXT
);
