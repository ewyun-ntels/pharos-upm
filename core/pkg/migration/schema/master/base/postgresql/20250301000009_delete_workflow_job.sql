-- +goose Up
DELETE FROM workflow_job WHERE name LIKE '[alert-rule]%';

-- +goose Down
