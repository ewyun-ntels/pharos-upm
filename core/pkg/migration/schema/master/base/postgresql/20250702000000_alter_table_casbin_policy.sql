-- +goose Up
CREATE TABLE IF NOT EXISTS casbin_policy(
    p_type VARCHAR(32)  DEFAULT '' NOT NULL,
    v0     VARCHAR(255) DEFAULT '' NOT NULL,
    v1     VARCHAR(255) DEFAULT '' NOT NULL,
    v2     VARCHAR(255) DEFAULT '' NOT NULL,
    v3     VARCHAR(255) DEFAULT '' NOT NULL,
    v4     VARCHAR(255) DEFAULT '' NOT NULL,
    v5     VARCHAR(255) DEFAULT '' NOT NULL
);
CREATE INDEX IF NOT EXISTS idx_casbin_policy ON casbin_policy (p_type,v0,v1);

ALTER TABLE casbin_policy RENAME TO casbin_policy_dashboard;

-- +goose Down
ALTER TABLE casbin_policy_dashboard RENAME TO casbin_policy;
