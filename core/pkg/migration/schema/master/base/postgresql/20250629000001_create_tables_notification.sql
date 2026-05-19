-- +goose Up
CREATE TABLE IF NOT EXISTS notification_rule
(
    id                UUID PRIMARY KEY NOT NULL,
    notification_type TEXT NOT NULL,
    name              TEXT NOT NULL UNIQUE,
    rule              TEXT NOT NULL,
    timestamp         TIMESTAMP NOT NULL,
    updated_at        TIMESTAMP NOT NULL
);

-- +goose Down
DROP TABLE IF EXISTS notification_rule;
