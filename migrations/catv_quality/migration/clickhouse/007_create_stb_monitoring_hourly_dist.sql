-- +goose Up
-- +goose StatementBegin
-- ========================================
-- 1. dist_stb_monitoring_hourly 분산 테이블 생성
-- 분산 조회를 위해 Distributed 엔진 사용
-- ========================================
CREATE TABLE IF NOT EXISTS catv.dist_stb_monitoring_hourly ON CLUSTER clickhouse_cluster
AS catv.stb_monitoring_hourly
ENGINE = Distributed(clickhouse_cluster_replicated, catv, stb_monitoring_hourly, rand());
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS catv.dist_stb_monitoring_hourly ON CLUSTER clickhouse_cluster SYNC;
-- +goose StatementEnd
