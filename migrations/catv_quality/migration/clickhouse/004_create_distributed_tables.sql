-- +goose Up
-- +goose StatementBegin
-- ========================================
-- 1. dist_stb_monitoring 분산 테이블 생성
-- ========================================
CREATE TABLE IF NOT EXISTS catv.dist_stb_monitoring ON CLUSTER clickhouse_cluster
AS catv.stb_monitoring
ENGINE = Distributed(clickhouse_cluster, catv, stb_monitoring, rand())
SETTINGS bytes_to_throw_insert = 5368709120;
-- +goose StatementEnd

-- +goose StatementBegin
-- ========================================
-- 2. dist_stb_monitoring_raw 분산 테이블 생성
-- ========================================
CREATE TABLE IF NOT EXISTS catv.dist_stb_monitoring_raw ON CLUSTER clickhouse_cluster
AS catv.stb_monitoring_raw
ENGINE = Distributed(clickhouse_cluster, catv, stb_monitoring_raw, rand())
SETTINGS bytes_to_throw_insert = 5368709120;
-- +goose StatementEnd

-- +goose StatementBegin
-- ========================================
-- 3. dist_stb_monitoring_error 분산 테이블 생성
-- ========================================
CREATE TABLE IF NOT EXISTS catv.dist_stb_monitoring_error ON CLUSTER clickhouse_cluster
AS catv.stb_monitoring_error
ENGINE = Distributed(clickhouse_cluster, catv, stb_monitoring_error, rand())
SETTINGS bytes_to_throw_insert = 5368709120;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS catv.dist_stb_monitoring_error ON CLUSTER clickhouse_cluster SYNC;
-- +goose StatementEnd

-- +goose StatementBegin
DROP TABLE IF EXISTS catv.dist_stb_monitoring_raw ON CLUSTER clickhouse_cluster SYNC;
-- +goose StatementEnd

-- +goose StatementBegin
DROP TABLE IF EXISTS catv.dist_stb_monitoring ON CLUSTER clickhouse_cluster SYNC;
-- +goose StatementEnd
