-- +goose Up
ALTER TABLE workflow_job DROP COLUMN serve_type;

INSERT INTO workflow_job(name, hosts, schedule, active, task_names, task_datas, updated_at) VALUES
('sample', json('["tarzan-master-0"]'), '*/1 * * * * *', false, json('["sample-task-01", "sample-task-02"]'),'{"sample-task-01":{"field-01":"value-01"}, "sample-task-02":"value"}', '2025-04-03T00:00:00Z');

-- +goose Down
DELETE FROM workflow_job WHERE name = 'sample';

ALTER TABLE workflow_job ADD COLUMN serve_type TEXT NOT NULL;
