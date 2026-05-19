-- +goose Up
-- +goose StatementBegin
ALTER TABLE catv.stb_transmission_quality_measurement ON CLUSTER clickhouse_cluster_replicated DROP COLUMN IF EXISTS ch_sid;
-- +goose StatementEnd

-- +goose StatementBegin
ALTER TABLE catv.stb_transmission_quality_measurement ON CLUSTER clickhouse_cluster_replicated DROP COLUMN IF EXISTS ch_num;
-- +goose StatementEnd

-- +goose StatementBegin
ALTER TABLE catv.stb_transmission_quality_measurement ON CLUSTER clickhouse_cluster_replicated DROP COLUMN IF EXISTS ch_freq;
-- +goose StatementEnd

-- +goose StatementBegin
ALTER TABLE catv.stb_transmission_quality_measurement ON CLUSTER clickhouse_cluster_replicated DROP COLUMN IF EXISTS ch_mode;
-- +goose StatementEnd

-- +goose StatementBegin
ALTER TABLE catv.stb_transmission_quality_measurement ON CLUSTER clickhouse_cluster_replicated DROP COLUMN IF EXISTS pwr_lvl;
-- +goose StatementEnd

-- +goose StatementBegin
ALTER TABLE catv.stb_transmission_quality_measurement ON CLUSTER clickhouse_cluster_replicated DROP COLUMN IF EXISTS snr;
-- +goose StatementEnd

-- +goose StatementBegin
ALTER TABLE catv.dist_stb_transmission_quality_measurement ON CLUSTER clickhouse_cluster_replicated DROP COLUMN IF EXISTS ch_sid;
-- +goose StatementEnd

-- +goose StatementBegin
ALTER TABLE catv.dist_stb_transmission_quality_measurement ON CLUSTER clickhouse_cluster_replicated DROP COLUMN IF EXISTS ch_num;
-- +goose StatementEnd

-- +goose StatementBegin
ALTER TABLE catv.dist_stb_transmission_quality_measurement ON CLUSTER clickhouse_cluster_replicated DROP COLUMN IF EXISTS ch_freq;
-- +goose StatementEnd

-- +goose StatementBegin
ALTER TABLE catv.dist_stb_transmission_quality_measurement ON CLUSTER clickhouse_cluster_replicated DROP COLUMN IF EXISTS ch_mode;
-- +goose StatementEnd

-- +goose StatementBegin
ALTER TABLE catv.dist_stb_transmission_quality_measurement ON CLUSTER clickhouse_cluster_replicated DROP COLUMN IF EXISTS pwr_lvl;
-- +goose StatementEnd

-- +goose StatementBegin
ALTER TABLE catv.dist_stb_transmission_quality_measurement ON CLUSTER clickhouse_cluster_replicated DROP COLUMN IF EXISTS snr;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE catv.stb_transmission_quality_measurement ON CLUSTER clickhouse_cluster_replicated ADD COLUMN IF NOT EXISTS ch_sid Nullable(String);
-- +goose StatementEnd

-- +goose StatementBegin
ALTER TABLE catv.stb_transmission_quality_measurement ON CLUSTER clickhouse_cluster_replicated ADD COLUMN IF NOT EXISTS ch_num Nullable(String);
-- +goose StatementEnd

-- +goose StatementBegin
ALTER TABLE catv.stb_transmission_quality_measurement ON CLUSTER clickhouse_cluster_replicated ADD COLUMN IF NOT EXISTS ch_freq Nullable(String);
-- +goose StatementEnd

-- +goose StatementBegin
ALTER TABLE catv.stb_transmission_quality_measurement ON CLUSTER clickhouse_cluster_replicated ADD COLUMN IF NOT EXISTS ch_mode Nullable(String);
-- +goose StatementEnd

-- +goose StatementBegin
ALTER TABLE catv.stb_transmission_quality_measurement ON CLUSTER clickhouse_cluster_replicated ADD COLUMN IF NOT EXISTS pwr_lvl Nullable(String);
-- +goose StatementEnd

-- +goose StatementBegin
ALTER TABLE catv.stb_transmission_quality_measurement ON CLUSTER clickhouse_cluster_replicated ADD COLUMN IF NOT EXISTS snr Nullable(String);
-- +goose StatementEnd

-- +goose StatementBegin
ALTER TABLE catv.dist_stb_transmission_quality_measurement ON CLUSTER clickhouse_cluster_replicated ADD COLUMN IF NOT EXISTS ch_sid Nullable(String);
-- +goose StatementEnd

-- +goose StatementBegin
ALTER TABLE catv.dist_stb_transmission_quality_measurement ON CLUSTER clickhouse_cluster_replicated ADD COLUMN IF NOT EXISTS ch_num Nullable(String);
-- +goose StatementEnd

-- +goose StatementBegin
ALTER TABLE catv.dist_stb_transmission_quality_measurement ON CLUSTER clickhouse_cluster_replicated ADD COLUMN IF NOT EXISTS ch_freq Nullable(String);
-- +goose StatementEnd

-- +goose StatementBegin
ALTER TABLE catv.dist_stb_transmission_quality_measurement ON CLUSTER clickhouse_cluster_replicated ADD COLUMN IF NOT EXISTS ch_mode Nullable(String);
-- +goose StatementEnd

-- +goose StatementBegin
ALTER TABLE catv.dist_stb_transmission_quality_measurement ON CLUSTER clickhouse_cluster_replicated ADD COLUMN IF NOT EXISTS pwr_lvl Nullable(String);
-- +goose StatementEnd

-- +goose StatementBegin
ALTER TABLE catv.dist_stb_transmission_quality_measurement ON CLUSTER clickhouse_cluster_replicated ADD COLUMN IF NOT EXISTS snr Nullable(String);
-- +goose StatementEnd
