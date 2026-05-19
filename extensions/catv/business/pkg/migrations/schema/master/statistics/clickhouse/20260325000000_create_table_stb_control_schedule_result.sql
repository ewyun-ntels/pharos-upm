-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS catv.stb_control_schedule_result ON CLUSTER clickhouse_cluster_replicated
(
	id String DEFAULT concat(schedule_id, '-', toString(request_time)),

	request_time DateTime,
	control_time Nullable(DateTime),

	schedule_id UUID,
	schedule_name String,
	schedule_type String,

	k8s_job_name String,

	stb_mdl_nm String,
	cm_mac_addr String,
	cm_ip_addr String,
	stb_mac_addr String,
	src_ip_addr String,
	work_type String,
	work_value Nullable(String),

    progress_status String,
	result_code String,
	result_message String,

	update_time DateTime DEFAULT now()
)
ENGINE = ReplicatedReplacingMergeTree('/clickhouse/tables/{shard}/catv/stb_control_schedule_result', '{replica}', update_time)
PARTITION BY toYYYYMM(request_time)
ORDER BY (id, schedule_id, k8s_job_name, cm_mac_addr, stb_mac_addr)
TTL request_time + toIntervalYear(3);
-- +goose StatementEnd

-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS catv.dist_stb_control_schedule_result ON CLUSTER clickhouse_cluster_replicated
AS catv.stb_control_schedule_result
ENGINE = Distributed(clickhouse_cluster_replicated, catv, stb_control_schedule_result, sipHash64(id));
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS catv.stb_control_schedule_result ON CLUSTER clickhouse_cluster_replicated SYNC;
-- +goose StatementEnd

-- +goose StatementBegin
DROP TABLE IF EXISTS catv.dist_stb_control_schedule_result ON CLUSTER clickhouse_cluster_replicated SYNC;
-- +goose StatementEnd
