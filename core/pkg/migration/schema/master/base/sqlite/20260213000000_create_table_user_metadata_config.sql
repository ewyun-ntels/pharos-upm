-- +goose Up
-- User metadata configuration history (JSON Schema for user additional fields)
-- Stores entire UserMetadataConfig (schema + uiSchema) as atomic configuration
-- Latest record is the active configuration
CREATE TABLE IF NOT EXISTS user_metadata_config_history (
    id TEXT PRIMARY KEY NOT NULL,        -- UUID
    config_json TEXT NOT NULL,           -- Complete UserMetadataConfig as JSON
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_by TEXT,                     -- Username who made this change
    description TEXT                     -- Optional change description
);

-- Index for history queries (sorted by creation time)
CREATE INDEX IF NOT EXISTS idx_user_config_created ON user_metadata_config_history(created_at DESC);

-- +goose Down
DROP TABLE IF EXISTS user_metadata_config_history;
