-- +goose Up
ALTER TABLE workflow_job DROP COLUMN hosts;

-- +goose Down
ALTER TABLE workflow_job ADD COLUMN hosts JSONB NOT NULL DEFAULT '';

UPDATE workflow_job SET hosts = json('["tarzan-master-0"]') WHERE name = 'sample';
