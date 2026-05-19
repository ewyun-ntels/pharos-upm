-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS catv.stb_control_schedule ON CLUSTER clickhouse_cluster_replicated
(
    id                   UUID DEFAULT generateUUIDv7(),
    name                 String,

    schedule_type        String,
    schedule_spec_once   Nullable(DateTime),
    schedule_spec_repeat Nullable(String),

    sos                  Array(String),
    l3s                  Array(String),
    cells                Array(String),
    settopboxs           Array(String),

    work_type            String,
    work_value           Nullable(String),

    area_type            String,
    area_ids             Array(String),

    create_time          DateTime DEFAULT now(), -- MATERIALIZED 필드는 SELECT *로 조회되지 않아서 사용하지 않음
    update_time          DateTime DEFAULT now(),
    is_deleted           UInt8
)
ENGINE = ReplicatedReplacingMergeTree('/clickhouse/tables/{shard}/catv/stb_control_schedule', '{replica}', update_time, is_deleted)
PARTITION BY toYYYYMM(update_time)
ORDER BY name;
-- +goose StatementEnd

-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS catv.dist_stb_control_schedule ON CLUSTER clickhouse_cluster_replicated
AS catv.stb_control_schedule
ENGINE = Distributed(clickhouse_cluster_replicated, catv, stb_control_schedule, sipHash64(name));
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS catv.stb_control_schedule ON CLUSTER clickhouse_cluster_replicated SYNC;
-- +goose StatementEnd

-- +goose StatementBegin
DROP TABLE IF EXISTS catv.dist_stb_control_schedule ON CLUSTER clickhouse_cluster_replicated SYNC;
-- +goose StatementEnd
