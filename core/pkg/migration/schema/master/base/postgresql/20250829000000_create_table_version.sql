-- +goose Up
CREATE TABLE IF NOT EXISTS core_app_version (
  module TEXT PRIMARY KEY,
  version TEXT NOT NULL,
  "commit" TEXT,
  build_time TEXT,
  updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- +goose Down
DROP TABLE IF EXISTS core_app_version;
