-- +goose Up
-- +goose StatementBegin
-- ========================================
-- dist_stb_monitoring_hourly_stb_freq 분산 테이블 생성
-- ========================================
CREATE TABLE IF NOT EXISTS catv.dist_stb_monitoring_hourly_stb_freq ON CLUSTER clickhouse_cluster
AS catv.stb_monitoring_hourly_stb_freq
ENGINE = Distributed(clickhouse_cluster_replicated, catv, stb_monitoring_hourly_stb_freq, rand());
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS catv.dist_stb_monitoring_hourly_stb_freq ON CLUSTER clickhouse_cluster SYNC;
-- +goose StatementEnd
