-- +goose Up
CREATE TABLE IF NOT EXISTS user_session
(
    token         TEXT PRIMARY KEY NOT NULL,
    refresh_token TEXT,
    username      TEXT,
    expired_at    DATETIME,
    issued_at     DATETIME,
    issuer        TEXT
);

-- +goose Down
DROP TABLE IF EXISTS user_session;
