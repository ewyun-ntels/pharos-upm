-- +goose Up
-- +goose StatementBegin
-- ========================================
-- stb_monitoring_hourly_stb 테이블 생성
-- STB 단위 1시간 집계를 위한 SummingMergeTree 테이블
-- 목적: 빠른 전체 품질 추이 조회 (주파수 구분 없음)
-- ========================================
CREATE TABLE IF NOT EXISTS catv.stb_monitoring_hourly_stb ON CLUSTER clickhouse_cluster_replicated
(
    `mntr_hour` DateTime,
    `CM_MAC_ADDR` LowCardinality(String),

    -- 메타데이터 (STB 정보)
    `STB_MDL_NM` LowCardinality(String),
    `TPO_CD` LowCardinality(String),
    `NET_CL_CD` LowCardinality(String),
    `INFO_CNTR_CD` LowCardinality(String),
    `HPY_CNTR_ORG_ID` LowCardinality(String),
    `SCRBR_SO_NM` LowCardinality(String),
    `L3_EQUIP_ID` LowCardinality(String),
    `L2_EQUIP_ID` LowCardinality(String),
    `STB_VER_CTT` LowCardinality(String),
    `CM_IP_ADDR` LowCardinality(String),
    `LOC_UI_VER_CTT` LowCardinality(String),
    `CLD_UI_VER_CTT` LowCardinality(String),
    `CCELL_NUM` LowCardinality(String) COMMENT '커버리지셀번호',

    -- SNR QAM 집계 (sum과 count로 저장, 평균은 sum/count로 계산)
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

    -- 레코드 수 (STB당 몇 개 레코드가 집계되었는지)
    `record_count` UInt64 DEFAULT 1,

    `insert_time` DateTime DEFAULT now()
)
ENGINE = ReplicatedSummingMergeTree('/clickhouse/tables/{shard}/catv/stb_monitoring_hourly_stb', '{replica}')
PARTITION BY toYYYYMM(mntr_hour)
ORDER BY (mntr_hour, CM_MAC_ADDR, STB_MDL_NM, TPO_CD, NET_CL_CD, INFO_CNTR_CD, HPY_CNTR_ORG_ID, SCRBR_SO_NM, L3_EQUIP_ID, L2_EQUIP_ID, STB_VER_CTT, CCELL_NUM)
TTL mntr_hour + INTERVAL 3 YEAR
COMMENT 'STB별 시간당 집계 (주파수 구분 없음) - 전체 품질 추이 조회용';
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS catv.stb_monitoring_hourly_stb ON CLUSTER clickhouse_cluster_replicated SYNC;
-- +goose StatementEnd
