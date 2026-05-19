-- +goose Up
-- +goose StatementBegin
-- ========================================
-- stb_monitoring_hourly_cell 테이블 생성
-- Cell 단위 1시간 집계를 위한 SummingMergeTree 테이블
-- 목적: 셀 단위 품질 추이 및 물리적 상태 조회 (STB 구분 없음)
-- ========================================
CREATE TABLE IF NOT EXISTS catv.stb_monitoring_hourly_cell ON CLUSTER clickhouse_cluster_replicated
(
    `mntr_hour` DateTime,
    `CCELL_NUM` LowCardinality(String) COMMENT '커버리지셀번호',

    -- 메타데이터 (Cell 수준에서 유의미한 정보)
    `STB_MDL_NM` LowCardinality(String),
    `TPO_CD` LowCardinality(String),
    `NET_CL_CD` LowCardinality(String),
    `INFO_CNTR_CD` LowCardinality(String),
    `HPY_CNTR_ORG_ID` LowCardinality(String),
    `SCRBR_SO_NM` LowCardinality(String),
    `L3_EQUIP_ID` LowCardinality(String),
    `L2_EQUIP_ID` LowCardinality(String),
    `STB_VER_CTT` LowCardinality(String),

    -- SNR QAM 집계
    `snr_qam_sum` Float64 DEFAULT 0,
    `snr_qam_count` UInt64 DEFAULT 0,
    `snr_qam_min` SimpleAggregateFunction(min, Nullable(Int32)),
    `snr_qam_max` SimpleAggregateFunction(max, Nullable(Int32)),

    -- SNR 8VSB 집계
    `snr_8vsb_sum` Float64 DEFAULT 0,
    `snr_8vsb_count` UInt64 DEFAULT 0,
    `snr_8vsb_min` SimpleAggregateFunction(min, Nullable(Int32)),
    `snr_8vsb_max` SimpleAggregateFunction(max, Nullable(Int32)),

    -- PWR LVL 집계
    `pwr_lvl_sum` Float64 DEFAULT 0,
    `pwr_lvl_count` UInt64 DEFAULT 0,
    `pwr_lvl_min` SimpleAggregateFunction(min, Nullable(Int32)),
    `pwr_lvl_max` SimpleAggregateFunction(max, Nullable(Int32)),

    -- 신호 미약 팝업 집계
    `popup_cnt_sum` Int64 DEFAULT 0,
    `popup_cnt_max` SimpleAggregateFunction(max, Int32) DEFAULT 0,

    -- 레코드 수 (집계된 신호 발생 건수)
    `record_count` UInt64 DEFAULT 1,

    `insert_time` DateTime DEFAULT now()
)
ENGINE = ReplicatedSummingMergeTree('/clickhouse/tables/{shard}/catv/stb_monitoring_hourly_cell', '{replica}')
PARTITION BY toYYYYMM(mntr_hour)
ORDER BY (mntr_hour, CCELL_NUM, STB_MDL_NM, TPO_CD, NET_CL_CD, INFO_CNTR_CD, HPY_CNTR_ORG_ID, SCRBR_SO_NM, L3_EQUIP_ID, L2_EQUIP_ID, STB_VER_CTT)
TTL mntr_hour + INTERVAL 3 YEAR
COMMENT '셀별 시간당 집계 (STB 구분 없음) - 셀 단위 품질 추이 조회용';
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS catv.stb_monitoring_hourly_cell ON CLUSTER clickhouse_cluster_replicated SYNC;
-- +goose StatementEnd
