-- +goose Up
-- Role metadata configuration history (Kustomize-style overlay)
-- Stores entire []RoleMetadata array as atomic configuration
-- Latest record is the active configuration
CREATE TABLE IF NOT EXISTS role_metadata_config_history (
    id TEXT PRIMARY KEY NOT NULL,        -- UUID
    config_json TEXT NOT NULL,           -- Complete []RoleMetadata array as JSON
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_by TEXT,                     -- Username who made this change
    description TEXT                     -- Optional change description
);

-- Index for history queries (sorted by creation time)
CREATE INDEX IF NOT EXISTS idx_config_created ON role_metadata_config_history(created_at DESC);

-- +goose Down
DROP TABLE IF EXISTS role_metadata_config_history;
