-- +goose Up
CREATE TABLE IF NOT EXISTS dashboard_folder (
    id          VARCHAR(255) PRIMARY KEY,
    name        TEXT NOT NULL,
    created_at  TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

-- +goose Down
DROP TABLE IF EXISTS dashboard_folder;
