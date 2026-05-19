-- +goose Up
INSERT INTO workflow_job(name, schedule, active, task_names, task_datas, updated_at) VALUES
('sqlite-backup', '0 0 2 * * *', true, json('["sqlite-backup"]'),json('{}'), '2025-07-10T00:00:00Z');

INSERT INTO workflow_job(name, schedule, active, task_names, task_datas, updated_at) VALUES
('sqlite-vacuum', '0 0 3 * * *', true, json('["sqlite-vacuum"]'),json('{}'), '2025-07-10T00:00:00Z');

-- +goose Down
DELETE FROM workflow_job WHERE name='sqlite-backup';
DELETE FROM workflow_job WHERE name='sqlite-vacuum';
