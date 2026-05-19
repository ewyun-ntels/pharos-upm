-- +goose Up
-- +goose StatementBegin
DROP TABLE IF EXISTS catv.stb_information_temp ON CLUSTER clickhouse_cluster_replicated SYNC;
-- +goose StatementEnd

-- +goose StatementBegin
CREATE TABLE catv.stb_information_temp ON CLUSTER clickhouse_cluster_replicated
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
ENGINE = ReplicatedReplacingMergeTree('/clickhouse/tables/{shard}/catv/stb_information_v2', '{replica}', UPDATE_TIME)
PARTITION BY toYYYYMMDD(UPDATE_TIME)
ORDER BY CM_MAC_ADDR
TTL UPDATE_TIME + toIntervalDay(30);
-- +goose StatementEnd

-- +goose StatementBegin
EXCHANGE TABLES catv.stb_information AND catv.stb_information_temp ON CLUSTER clickhouse_cluster_replicated;
-- +goose StatementEnd

-- +goose StatementBegin
INSERT INTO catv.stb_information SELECT * FROM catv.stb_information_temp;
-- +goose StatementEnd

-- +goose StatementBegin
DROP TABLE IF EXISTS catv.stb_information_temp ON CLUSTER clickhouse_cluster_replicated SYNC;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS catv.stb_information_temp ON CLUSTER clickhouse_cluster_replicated SYNC;
-- +goose StatementEnd

-- +goose StatementBegin
CREATE TABLE catv.stb_information_temp ON CLUSTER clickhouse_cluster_replicated
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
ENGINE = ReplicatedReplacingMergeTree('/clickhouse/tables/{shard}/catv/stb_information_v1', '{replica}')
PARTITION BY toYYYYMMDD(UPDATE_TIME)
ORDER BY (CM_MAC_ADDR, UPDATE_TIME)
TTL UPDATE_TIME + toIntervalDay(30);
-- +goose StatementEnd

-- +goose StatementBegin
EXCHANGE TABLES catv.stb_information AND catv.stb_information_temp ON CLUSTER clickhouse_cluster_replicated;
-- +goose StatementEnd

-- +goose StatementBegin
INSERT INTO catv.stb_information SELECT * FROM catv.stb_information_temp;
-- +goose StatementEnd

-- +goose StatementBegin
DROP TABLE IF EXISTS catv.stb_information_temp ON CLUSTER clickhouse_cluster_replicated SYNC;
-- +goose StatementEnd
