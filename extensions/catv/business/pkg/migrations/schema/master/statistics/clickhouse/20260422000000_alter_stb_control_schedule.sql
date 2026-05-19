-- +goose Up
-- +goose StatementBegin
ALTER TABLE catv.stb_control_schedule ON CLUSTER clickhouse_cluster_replicated RENAME COLUMN IF EXISTS settopboxs TO settopboxes;
-- +goose StatementEnd

-- +goose StatementBegin
ALTER TABLE catv.dist_stb_control_schedule ON CLUSTER clickhouse_cluster_replicated RENAME COLUMN IF EXISTS settopboxs TO settopboxes;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE catv.stb_control_schedule ON CLUSTER clickhouse_cluster_replicated RENAME COLUMN IF EXISTS settopboxes TO settopboxs;
-- +goose StatementEnd

-- +goose StatementBegin
ALTER TABLE catv.dist_stb_control_schedule ON CLUSTER clickhouse_cluster_replicated RENAME COLUMN IF EXISTS settopboxes TO settopboxs;
-- +goose StatementEnd
