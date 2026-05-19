-- +goose Up
CREATE TABLE IF NOT EXISTS users
(
    username      TEXT PRIMARY KEY NOT NULL,
    password_hash TEXT
);

-- +goose Down
DROP TABLE IF EXISTS users;
