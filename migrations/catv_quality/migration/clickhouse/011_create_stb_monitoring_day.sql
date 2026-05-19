-- +goose Up
-- +goose StatementBegin
-- ========================================
-- 1. stb_monitoring_day 테이블 생성
-- 1일 단위 집계 (CM 단위, hourly와 동일 구조)
-- ========================================
CREATE TABLE IF NOT EXISTS catv.stb_monitoring_day ON CLUSTER clickhouse_cluster_replicated
(
    mntr_day Date,
    STB_MDL_NM LowCardinality(String),
    TPO_CD LowCardinality(String),
    NET_CL_CD LowCardinality(String),
    INFO_CNTR_CD LowCardinality(String),
    HPY_CNTR_ORG_ID LowCardinality(String),
    STB_QLTY_GR_CD LowCardinality(String),
    SCRBR_SO_NM LowCardinality(String),
    L3_EQUIP_ID LowCardinality(String),
    L2_EQUIP_ID LowCardinality(String),
    STB_VER_CTT LowCardinality(String),
    CHNL_FREQ_VAL LowCardinality(String),
    CM_MAC_ADDR LowCardinality(String),
    CM_IP_ADDR LowCardinality(String),
    LOC_UI_VER_CTT LowCardinality(String),
    CLD_UI_VER_CTT LowCardinality(String),
    CCELL_NUM LowCardinality(String) COMMENT '커버리지셀번호',

    -- SNR QAM 집계
    snr_qam_p10 AggregateFunction(quantile(0.1), Nullable(Int32)),
    snr_qam_p50 AggregateFunction(quantile(0.5), Nullable(Int32)),
    snr_qam_p90 AggregateFunction(quantile(0.9), Nullable(Int32)),
    snr_qam_min AggregateFunction(min, Nullable(Int32)),
    snr_qam_max AggregateFunction(max, Nullable(Int32)),
    snr_qam_avg AggregateFunction(avg, Nullable(Int32)),

    -- SNR 8VSB 집계
    snr_8vsb_p10 AggregateFunction(quantile(0.1), Nullable(Int32)),
    snr_8vsb_p50 AggregateFunction(quantile(0.5), Nullable(Int32)),
    snr_8vsb_p90 AggregateFunction(quantile(0.9), Nullable(Int32)),
    snr_8vsb_min AggregateFunction(min, Nullable(Int32)),
    snr_8vsb_max AggregateFunction(max, Nullable(Int32)),
    snr_8vsb_avg AggregateFunction(avg, Nullable(Int32)),

    -- PWR LVL 집계
    pwr_lvl_p10 AggregateFunction(quantile(0.1), Nullable(Int32)),
    pwr_lvl_p50 AggregateFunction(quantile(0.5), Nullable(Int32)),
    pwr_lvl_p90 AggregateFunction(quantile(0.9), Nullable(Int32)),
    pwr_lvl_min AggregateFunction(min, Nullable(Int32)),
    pwr_lvl_max AggregateFunction(max, Nullable(Int32)),
    pwr_lvl_avg AggregateFunction(avg, Nullable(Int32)),

    -- 신호 미약 팝업 집계
    popup_cnt_sum AggregateFunction(sum, Nullable(Int32)),
    popup_cnt_avg AggregateFunction(avg, Nullable(Int32)),
    popup_cnt_max AggregateFunction(max, Nullable(Int32)),

    -- 품질 등급 분포 (상태값 카운트용)
    stb_qlty_gr_count AggregateFunction(count, String),
    pwr_lvl_qlty_gr_count AggregateFunction(count, String),
    snr_qlty_gr_count AggregateFunction(count, String),

    insert_time DateTime DEFAULT now()
)
ENGINE = ReplicatedAggregatingMergeTree('/clickhouse/tables/{shard}/catv/stb_monitoring_day', '{replica}')
PARTITION BY toYYYYMM(mntr_day)
ORDER BY (mntr_day, STB_MDL_NM, TPO_CD, NET_CL_CD, INFO_CNTR_CD, HPY_CNTR_ORG_ID, STB_QLTY_GR_CD, SCRBR_SO_NM, L3_EQUIP_ID, L2_EQUIP_ID, CCELL_NUM, STB_VER_CTT, CHNL_FREQ_VAL, CM_MAC_ADDR, CM_IP_ADDR, LOC_UI_VER_CTT, CLD_UI_VER_CTT)
TTL mntr_day + INTERVAL 5 YEAR
SETTINGS index_granularity = 8192;
-- +goose StatementEnd

-- +goose StatementBegin
-- ========================================
-- 2. mv_stb_monitoring_day Materialized View 생성
-- stb_monitoring → stb_monitoring_day (원본에서 직접 집계)
-- ========================================
CREATE MATERIALIZED VIEW IF NOT EXISTS catv.mv_stb_monitoring_day ON CLUSTER clickhouse_cluster
TO catv.stb_monitoring_day
AS
SELECT
    toDate(mntr_datetime) AS mntr_day,
    STB_MDL_NM,
    TPO_CD,
    NET_CL_CD,
    INFO_CNTR_CD,
    HPY_CNTR_ORG_ID,
    STB_QLTY_GR_CD,
    SCRBR_SO_NM,
    L3_EQUIP_ID,
    L2_EQUIP_ID,
    STB_VER_CTT,
    replace(CHNL_FREQ_VAL, ' ', '') AS CHNL_FREQ_VAL,
    CM_MAC_ADDR,
    CM_IP_ADDR,
    LOC_UI_VER_CTT,
    CLD_UI_VER_CTT,
    CCELL_NUM,

    -- SNR QAM 집계
    quantileState(0.1)(if(FREQ_METH_CTT = '256QAM', SNR_VAL, NULL)) AS snr_qam_p10,
    quantileState(0.5)(if(FREQ_METH_CTT = '256QAM', SNR_VAL, NULL)) AS snr_qam_p50,
    quantileState(0.9)(if(FREQ_METH_CTT = '256QAM', SNR_VAL, NULL)) AS snr_qam_p90,
    minState(if(FREQ_METH_CTT = '256QAM', SNR_VAL, NULL)) AS snr_qam_min,
    maxState(if(FREQ_METH_CTT = '256QAM', SNR_VAL, NULL)) AS snr_qam_max,
    avgState(if(FREQ_METH_CTT = '256QAM', SNR_VAL, NULL)) AS snr_qam_avg,

    -- SNR 8VSB 집계
    quantileState(0.1)(if(FREQ_METH_CTT = '8VSB', SNR_VAL, NULL)) AS snr_8vsb_p10,
    quantileState(0.5)(if(FREQ_METH_CTT = '8VSB', SNR_VAL, NULL)) AS snr_8vsb_p50,
    quantileState(0.9)(if(FREQ_METH_CTT = '8VSB', SNR_VAL, NULL)) AS snr_8vsb_p90,
    minState(if(FREQ_METH_CTT = '8VSB', SNR_VAL, NULL)) AS snr_8vsb_min,
    maxState(if(FREQ_METH_CTT = '8VSB', SNR_VAL, NULL)) AS snr_8vsb_max,
    avgState(if(FREQ_METH_CTT = '8VSB', SNR_VAL, NULL)) AS snr_8vsb_avg,

    -- PWR LVL 집계
    quantileState(0.1)(PWR_LVL_VAL) AS pwr_lvl_p10,
    quantileState(0.5)(PWR_LVL_VAL) AS pwr_lvl_p50,
    quantileState(0.9)(PWR_LVL_VAL) AS pwr_lvl_p90,
    minState(PWR_LVL_VAL) AS pwr_lvl_min,
    maxState(PWR_LVL_VAL) AS pwr_lvl_max,
    avgState(PWR_LVL_VAL) AS pwr_lvl_avg,

    -- 신호 미약 팝업 집계
    sumState(SGNL_WEAK_POPUP_CNT) AS popup_cnt_sum,
    avgState(SGNL_WEAK_POPUP_CNT) AS popup_cnt_avg,
    maxState(SGNL_WEAK_POPUP_CNT) AS popup_cnt_max,

    -- 품질 등급 분포
    countState(STB_QLTY_GR_CD) AS stb_qlty_gr_count,
    countState(PWR_LVL_STB_QLTY_GR_CD) AS pwr_lvl_qlty_gr_count,
    countState(SNR_STB_QLTY_GR_CD) AS snr_qlty_gr_count

FROM catv.stb_monitoring
WHERE (SNR_VAL IS NULL OR (SNR_VAL >= 0 AND SNR_VAL <= 100))
  AND (PWR_LVL_VAL IS NULL OR (PWR_LVL_VAL >= -100 AND PWR_LVL_VAL <= 100))
  AND (SGNL_WEAK_POPUP_CNT IS NULL OR SGNL_WEAK_POPUP_CNT >= 0)

GROUP BY
    mntr_day,
    STB_MDL_NM,
    TPO_CD,
    NET_CL_CD,
    INFO_CNTR_CD,
    HPY_CNTR_ORG_ID,
    STB_QLTY_GR_CD,
    SCRBR_SO_NM,
    L3_EQUIP_ID,
    L2_EQUIP_ID,
    STB_VER_CTT,
    CHNL_FREQ_VAL,
    CM_MAC_ADDR,
    CM_IP_ADDR,
    LOC_UI_VER_CTT,
    CLD_UI_VER_CTT,
    CCELL_NUM;
-- +goose StatementEnd

-- +goose StatementBegin
-- ========================================
-- 3. dist_stb_monitoring_day Distributed 테이블 생성
-- ========================================
CREATE TABLE IF NOT EXISTS catv.dist_stb_monitoring_day ON CLUSTER clickhouse_cluster
AS catv.stb_monitoring_day
ENGINE = Distributed(
    'clickhouse_cluster_replicated',
    'catv',
    'stb_monitoring_day',
    rand()
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS catv.dist_stb_monitoring_day ON CLUSTER clickhouse_cluster SYNC;
-- +goose StatementEnd

-- +goose StatementBegin
DROP TABLE IF EXISTS catv.mv_stb_monitoring_day ON CLUSTER clickhouse_cluster SYNC;
-- +goose StatementEnd

-- +goose StatementBegin
DROP TABLE IF EXISTS catv.stb_monitoring_day ON CLUSTER clickhouse_cluster_replicated SYNC;
-- +goose StatementEnd
