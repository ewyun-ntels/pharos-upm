-- +goose Up
-- +goose StatementBegin
-- ========================================
-- 1. stb_monitoring_hourly 테이블 생성
-- 1시간 단위 집계를 위한 ReplicatedAggregatingMergeTree 테이블
-- ========================================
CREATE TABLE IF NOT EXISTS catv.stb_monitoring_hourly ON CLUSTER clickhouse_cluster_replicated
(
    `mntr_hour` DateTime,
    `STB_MDL_NM` LowCardinality(String),
    `TPO_CD` LowCardinality(String),
    `NET_CL_CD` LowCardinality(String),
    `INFO_CNTR_CD` LowCardinality(String),
    `HPY_CNTR_ORG_ID` LowCardinality(String),
    `STB_QLTY_GR_CD` LowCardinality(String),
    `SCRBR_SO_NM` LowCardinality(String),
    `L3_EQUIP_ID` LowCardinality(String),
    `L2_EQUIP_ID` LowCardinality(String),
    `STB_VER_CTT` LowCardinality(String),
    `CHNL_FREQ_VAL` LowCardinality(String),
    `CM_MAC_ADDR` LowCardinality(String),
    `CM_IP_ADDR` LowCardinality(String),
    `LOC_UI_VER_CTT` LowCardinality(String),
    `CLD_UI_VER_CTT` LowCardinality(String),
    `CCELL_NUM` LowCardinality(String) COMMENT '커버리지셀번호',
    
    -- SNR QAM 집계
    `snr_qam_p10` AggregateFunction(quantile(0.1), Nullable(Int32)),
    `snr_qam_p50` AggregateFunction(quantile(0.5), Nullable(Int32)),
    `snr_qam_p90` AggregateFunction(quantile(0.9), Nullable(Int32)),
    `snr_qam_min` AggregateFunction(min, Nullable(Int32)),
    `snr_qam_max` AggregateFunction(max, Nullable(Int32)),
    `snr_qam_avg` AggregateFunction(avg, Nullable(Int32)),
    
    -- SNR 8VSB 집계
    `snr_8vsb_p10` AggregateFunction(quantile(0.1), Nullable(Int32)),
    `snr_8vsb_p50` AggregateFunction(quantile(0.5), Nullable(Int32)),
    `snr_8vsb_p90` AggregateFunction(quantile(0.9), Nullable(Int32)),
    `snr_8vsb_min` AggregateFunction(min, Nullable(Int32)),
    `snr_8vsb_max` AggregateFunction(max, Nullable(Int32)),
    `snr_8vsb_avg` AggregateFunction(avg, Nullable(Int32)),

    -- PWR LVL 집계
    `pwr_lvl_p10` AggregateFunction(quantile(0.1), Nullable(Int32)),
    `pwr_lvl_p50` AggregateFunction(quantile(0.5), Nullable(Int32)),
    `pwr_lvl_p90` AggregateFunction(quantile(0.9), Nullable(Int32)),
    `pwr_lvl_min` AggregateFunction(min, Nullable(Int32)),
    `pwr_lvl_max` AggregateFunction(max, Nullable(Int32)),
    `pwr_lvl_avg` AggregateFunction(avg, Nullable(Int32)),

    -- 신호 미약 팝업 집계
    `popup_cnt_sum` AggregateFunction(sum, Nullable(Int32)),
    `popup_cnt_avg` AggregateFunction(avg, Nullable(Int32)),
    `popup_cnt_max` AggregateFunction(max, Nullable(Int32)),
    
    -- 품질 등급 분포 (상태값 카운트용)
    `stb_qlty_gr_count` AggregateFunction(count, String),
    `pwr_lvl_qlty_gr_count` AggregateFunction(count, String),
    `snr_qlty_gr_count` AggregateFunction(count, String),
    
    `insert_time` DateTime DEFAULT now()
)
ENGINE = ReplicatedAggregatingMergeTree('/clickhouse/tables/{shard}/catv/stb_monitoring_hourly', '{replica}')
PARTITION BY toYYYYMM(mntr_hour)
ORDER BY (mntr_hour, STB_MDL_NM, TPO_CD, NET_CL_CD, INFO_CNTR_CD, HPY_CNTR_ORG_ID, STB_QLTY_GR_CD, SCRBR_SO_NM, L3_EQUIP_ID, L2_EQUIP_ID, CCELL_NUM, STB_VER_CTT, CHNL_FREQ_VAL, CM_MAC_ADDR, CM_IP_ADDR, LOC_UI_VER_CTT, CLD_UI_VER_CTT)
TTL mntr_hour + INTERVAL 3 YEAR;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS catv.stb_monitoring_hourly ON CLUSTER clickhouse_cluster_replicated SYNC;
-- +goose StatementEnd
