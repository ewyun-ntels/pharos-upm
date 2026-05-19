-- +goose Up
DROP TABLE IF EXISTS history_login;
CREATE TABLE IF NOT EXISTS history_login (
    session_id TEXT NOT NULL,
    user_id    TEXT NOT NULL,
    client_ip  TEXT,
    started_at DATETIME NOT NULL,
    ended_at   DATETIME,
    end_reason TEXT
);
CREATE UNIQUE INDEX IF NOT EXISTS ux_history_login_session_id ON history_login (session_id);

-- +goose Down
DROP INDEX IF EXISTS ux_history_login_session_id;
DROP TABLE IF EXISTS history_login;
