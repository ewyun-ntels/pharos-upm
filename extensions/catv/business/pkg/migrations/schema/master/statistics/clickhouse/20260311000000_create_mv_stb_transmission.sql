-- +goose Up
-- +goose StatementBegin
CREATE MATERIALIZED VIEW IF NOT EXISTS catv.mv_stb_transmission_periodic ON CLUSTER clickhouse_cluster_replicated
TO catv.stb_information
AS
SELECT
	stb_model AS STB_MDL_NM,
	cm_mac AS CM_MAC_ADDR,
	cm_ip AS CM_IP_ADDR,
	mac_addr AS STB_MAC_ADDR,
	stb_ip AS SRC_IP_ADDR,
	'testbed' AS CCELL_NUM,
	'testbed' AS L3_EQUIP_TID_VAL,
	'testbed' AS SCRBR_SO_NM,
	NOW() AS UPDATE_TIME
FROM catv.stb_transmission_periodic;
-- +goose StatementEnd

-- +goose StatementBegin
CREATE MATERIALIZED VIEW IF NOT EXISTS catv.mv_stb_transmission_daily ON CLUSTER clickhouse_cluster_replicated
TO catv.stb_information
AS
SELECT
	stb_model AS STB_MDL_NM,
	cm_mac AS CM_MAC_ADDR,
	cm_ip AS CM_IP_ADDR,
	mac_addr AS STB_MAC_ADDR,
	stb_ip AS SRC_IP_ADDR,
	'testbed' AS CCELL_NUM,
	'testbed' AS L3_EQUIP_TID_VAL,
	'testbed' AS SCRBR_SO_NM,
	NOW() AS UPDATE_TIME
FROM catv.stb_transmission_daily;
-- +goose StatementEnd

-- +goose StatementBegin
CREATE MATERIALIZED VIEW IF NOT EXISTS catv.mv_stb_transmission_quality_measurement ON CLUSTER clickhouse_cluster_replicated
TO catv.stb_information
AS
SELECT
	stb_model AS STB_MDL_NM,
	cm_mac AS CM_MAC_ADDR,
	cm_ip AS CM_IP_ADDR,
	mac_addr AS STB_MAC_ADDR,
	stb_ip AS SRC_IP_ADDR,
	'testbed' AS CCELL_NUM,
	'testbed' AS L3_EQUIP_TID_VAL,
	'testbed' AS SCRBR_SO_NM,
	NOW() AS UPDATE_TIME
FROM catv.stb_transmission_quality_measurement;
-- +goose StatementEnd

-- +goose StatementBegin
CREATE MATERIALIZED VIEW IF NOT EXISTS catv.mv_stb_transmission_diagnostic ON CLUSTER clickhouse_cluster_replicated
TO catv.stb_information
AS
SELECT
	stb_model AS STB_MDL_NM,
	cm_mac AS CM_MAC_ADDR,
	cm_ip AS CM_IP_ADDR,
	mac_addr AS STB_MAC_ADDR,
	stb_ip AS SRC_IP_ADDR,
	'testbed' AS CCELL_NUM,
	'testbed' AS L3_EQUIP_TID_VAL,
	'testbed' AS SCRBR_SO_NM,
	NOW() AS UPDATE_TIME
FROM catv.stb_transmission_diagnostic;
-- +goose StatementEnd

-- +goose StatementBegin
CREATE MATERIALIZED VIEW IF NOT EXISTS catv.mv_stb_transmission_network_quality_transition ON CLUSTER clickhouse_cluster_replicated
TO catv.stb_information
AS
SELECT
	stb_model AS STB_MDL_NM,
	cm_mac AS CM_MAC_ADDR,
	cm_ip AS CM_IP_ADDR,
	mac_addr AS STB_MAC_ADDR,
	stb_ip AS SRC_IP_ADDR,
	'testbed' AS CCELL_NUM,
	'testbed' AS L3_EQUIP_TID_VAL,
	'testbed' AS SCRBR_SO_NM,
	NOW() AS UPDATE_TIME
FROM catv.stb_transmission_network_quality_transition;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS catv.mv_stb_transmission_periodic ON CLUSTER clickhouse_cluster_replicated SYNC;
-- +goose StatementEnd

-- +goose StatementBegin
DROP TABLE IF EXISTS catv.mv_stb_transmission_daily ON CLUSTER clickhouse_cluster_replicated SYNC;
-- +goose StatementEnd

-- +goose StatementBegin
DROP TABLE IF EXISTS catv.mv_stb_transmission_quality_measurement ON CLUSTER clickhouse_cluster_replicated SYNC;
-- +goose StatementEnd

-- +goose StatementBegin
DROP TABLE IF EXISTS catv.mv_stb_transmission_diagnostic ON CLUSTER clickhouse_cluster_replicated SYNC;
-- +goose StatementEnd

-- +goose StatementBegin
DROP TABLE IF EXISTS catv.mv_stb_transmission_network_quality_transition ON CLUSTER clickhouse_cluster_replicated SYNC;
-- +goose StatementEnd