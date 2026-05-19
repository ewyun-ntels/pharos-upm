-- +goose Up
UPDATE casbin_policy_dashboard SET v3 = 'user';

-- +goose Down
UPDATE casbin_policy_dashboard SET v3 = '';
