-- +goose Up
-- +goose StatementBegin
-- ========================================
-- dist_stb_monitoring_day_stb_perf 분산 테이블 생성
-- clickhouse_cluster (샤드 2개) 사용으로 쿼리 성능 약 2배 향상
-- ========================================
CREATE TABLE IF NOT EXISTS catv.dist_stb_monitoring_day_stb_perf ON CLUSTER clickhouse_cluster
AS catv.stb_monitoring_day_stb_perf
ENGINE = Distributed(clickhouse_cluster, catv, stb_monitoring_day_stb_perf, rand())
COMMENT 'STB별 일별 집계 v2 분산 테이블 (샤드 2개)';
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS catv.dist_stb_monitoring_day_stb_perf ON CLUSTER clickhouse_cluster SYNC;
-- +goose StatementEnd
