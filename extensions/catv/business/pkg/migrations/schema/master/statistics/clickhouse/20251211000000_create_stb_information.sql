-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS catv.stb_information ON CLUSTER clickhouse_cluster_replicated
(
	STB_MDL_NM       String,
	CM_MAC_ADDR      String,
	CM_IP_ADDR       String,
	STB_MAC_ADDR     String,
	SRC_IP_ADDR      String,
	CCELL_NUM        String,
	L3_EQUIP_TID_VAL String,
	SCRBR_SO_NM      String,
	UPDATE_TIME      DateTime
)
ENGINE = ReplicatedReplacingMergeTree('/clickhouse/tables/{shard}/catv/stb_information', '{replica}')
PARTITION BY toYYYYMMDD(UPDATE_TIME)
ORDER BY (CM_MAC_ADDR, UPDATE_TIME)
TTL UPDATE_TIME + toIntervalDay(30);
-- +goose StatementEnd

-- +goose StatementBegin
CREATE MATERIALIZED VIEW IF NOT EXISTS catv.mv_stb_information ON CLUSTER clickhouse_cluster_replicated
TO catv.stb_information
AS
SELECT
	STB_MDL_NM,
	CM_MAC_ADDR,
	CM_IP_ADDR,
	STB_MAC_ADDR,
	SRC_IP_ADDR,
	CCELL_NUM,
	L3_EQUIP_TID_VAL,
	SCRBR_SO_NM,
	NOW() AS UPDATE_TIME
FROM catv.stb_monitoring;
-- +goose StatementEnd

-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS catv.dist_stb_information ON CLUSTER clickhouse_cluster_replicated
AS catv.stb_information
ENGINE = Distributed(clickhouse_cluster_replicated, catv, stb_information, sipHash64(CM_MAC_ADDR));
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS catv.stb_information ON CLUSTER clickhouse_cluster_replicated SYNC;
-- +goose StatementEnd

-- +goose StatementBegin
DROP TABLE IF EXISTS catv.mv_stb_information ON CLUSTER clickhouse_cluster_replicated SYNC;
-- +goose StatementEnd

-- +goose StatementBegin
DROP TABLE IF EXISTS catv.dist_stb_information ON CLUSTER clickhouse_cluster_replicated SYNC;
-- +goose StatementEnd
