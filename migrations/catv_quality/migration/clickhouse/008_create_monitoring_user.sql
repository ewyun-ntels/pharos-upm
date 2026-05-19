-- +goose Up
-- +goose StatementBegin
-- ========================================
-- 1. 모니터링 전용 롤 생성
-- ========================================
CREATE ROLE IF NOT EXISTS monitoring_role ON CLUSTER clickhouse_cluster;
-- +goose StatementEnd

-- +goose StatementBegin
GRANT SELECT ON catv.* TO monitoring_role ON CLUSTER clickhouse_cluster;
-- +goose StatementEnd

-- +goose StatementBegin
GRANT SELECT ON system.* TO monitoring_role ON CLUSTER clickhouse_cluster;
-- +goose StatementEnd

-- +goose StatementBegin
GRANT REMOTE ON *.* TO monitoring_role ON CLUSTER clickhouse_cluster;
-- +goose StatementEnd

-- +goose StatementBegin
-- ========================================
-- 3. 모니터링 사용자 생성 (비밀번호: catv_mon123!)
-- ========================================
CREATE USER IF NOT EXISTS catv_mon 
IDENTIFIED WITH sha256_password BY 'catv_mon123!' 
ON CLUSTER clickhouse_cluster;
-- +goose StatementEnd

-- +goose StatementBegin
-- ========================================
-- 4. 사용자에게 롤 부여
-- ========================================
GRANT monitoring_role TO catv_mon ON CLUSTER clickhouse_cluster;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP USER IF EXISTS catv_mon ON CLUSTER clickhouse_cluster;
-- +goose StatementEnd

-- +goose StatementBegin
DROP ROLE IF EXISTS monitoring_role ON CLUSTER clickhouse_cluster;
-- +goose StatementEnd
