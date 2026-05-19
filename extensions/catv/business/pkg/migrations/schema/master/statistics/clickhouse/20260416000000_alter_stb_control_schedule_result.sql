-- +goose Up
-- +goose StatementBegin
ALTER TABLE catv.stb_control_schedule_result ON CLUSTER clickhouse_cluster_replicated RENAME COLUMN IF EXISTS control_time TO control_start_time;
-- +goose StatementEnd

-- +goose StatementBegin
ALTER TABLE catv.stb_control_schedule_result ON CLUSTER clickhouse_cluster_replicated ADD COLUMN IF NOT EXISTS control_end_time Nullable(DateTime) AFTER control_start_time;
-- +goose StatementEnd

-- +goose StatementBegin
ALTER TABLE catv.dist_stb_control_schedule_result ON CLUSTER clickhouse_cluster_replicated RENAME COLUMN IF EXISTS control_time TO control_start_time;
-- +goose StatementEnd

-- +goose StatementBegin
ALTER TABLE catv.dist_stb_control_schedule_result ON CLUSTER clickhouse_cluster_replicated ADD COLUMN IF NOT EXISTS control_end_time Nullable(DateTime) AFTER control_start_time;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE catv.stb_control_schedule_result ON CLUSTER clickhouse_cluster_replicated RENAME COLUMN IF EXISTS control_start_time TO control_time;
-- +goose StatementEnd

-- +goose StatementBegin
ALTER TABLE catv.stb_control_schedule_result ON CLUSTER clickhouse_cluster_replicated DROP COLUMN IF EXISTS control_end_time;
-- +goose StatementEnd

-- +goose StatementBegin
ALTER TABLE catv.dist_stb_control_schedule_result ON CLUSTER clickhouse_cluster_replicated RENAME COLUMN IF EXISTS control_start_time TO control_time;
-- +goose StatementEnd

-- +goose StatementBegin
ALTER TABLE catv.dist_stb_control_schedule_result ON CLUSTER clickhouse_cluster_replicated DROP COLUMN IF EXISTS control_end_time;
-- +goose StatementEnd
