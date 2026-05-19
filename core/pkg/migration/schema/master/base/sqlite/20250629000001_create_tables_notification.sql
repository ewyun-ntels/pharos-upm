-- +goose Up
CREATE TABLE IF NOT EXISTS notification_rule
(
    id                UUID PRIMARY KEY NOT NULL,
    notification_type TEXT NOT NULL,
    name              TEXT NOT NULL UNIQUE,
    rule              TEXT NOT NULL,
    timestamp         DATETIME NOT NULL,
    updated_at        DATETIME NOT NULL
);

-- +goose Down
DROP TABLE IF EXISTS notification_rule;
