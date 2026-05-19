-- +goose Up
-- +goose StatementBegin
-- ========================================
-- 1. stb_monitoring_cell_hourly 테이블 생성
-- 셀×채널 단위 1시간 집계 (CM_MAC_ADDR 제거로 추가 압축)
-- ========================================
CREATE TABLE IF NOT EXISTS catv.stb_monitoring_cell_hourly ON CLUSTER clickhouse_cluster_replicated
(
    -- 시간 (파티션 키)
    mntr_hour DateTime,

    -- 위치 정보 (낮은 cardinality)
    TPO_CD LowCardinality(String),              -- 국사코드
    INFO_CNTR_CD LowCardinality(String),        -- 정보센터코드
    HPY_CNTR_ORG_ID LowCardinality(String),     -- 행복센터조직ID
    NET_CL_CD LowCardinality(String),           -- 망구분코드

    -- 네트워크 장비 정보
    L3_EQUIP_ID LowCardinality(String),         -- L3장비ID
    L2_EQUIP_ID LowCardinality(String),         -- L2장비ID
    CCELL_NUM LowCardinality(String) COMMENT '커버리지셀번호',

    -- 품질 및 기타 정보
    STB_QLTY_GR_CD LowCardinality(String),      -- STB품질등급코드
    STB_MDL_NM LowCardinality(String),          -- STB모델명
    SCRBR_SO_NM LowCardinality(String),         -- 가입자SO명
    CHNL_FREQ_VAL LowCardinality(String),       -- 채널주파수값

    -- SNR QAM 통계 (Aggregate Functions)
    snr_qam_p10 AggregateFunction(quantile(0.1), Nullable(Int32)),
    snr_qam_p50 AggregateFunction(quantile(0.5), Nullable(Int32)),
    snr_qam_p90 AggregateFunction(quantile(0.9), Nullable(Int32)),
    snr_qam_min AggregateFunction(min, Nullable(Int32)),
    snr_qam_max AggregateFunction(max, Nullable(Int32)),
    snr_qam_avg AggregateFunction(avg, Nullable(Int32)),

    -- SNR 8VSB 통계
    snr_8vsb_p10 AggregateFunction(quantile(0.1), Nullable(Int32)),
    snr_8vsb_p50 AggregateFunction(quantile(0.5), Nullable(Int32)),
    snr_8vsb_p90 AggregateFunction(quantile(0.9), Nullable(Int32)),
    snr_8vsb_min AggregateFunction(min, Nullable(Int32)),
    snr_8vsb_max AggregateFunction(max, Nullable(Int32)),
    snr_8vsb_avg AggregateFunction(avg, Nullable(Int32)),

    -- Power Level 통계
    pwr_lvl_p10 AggregateFunction(quantile(0.1), Nullable(Int32)),
    pwr_lvl_p50 AggregateFunction(quantile(0.5), Nullable(Int32)),
    pwr_lvl_p90 AggregateFunction(quantile(0.9), Nullable(Int32)),
    pwr_lvl_min AggregateFunction(min, Nullable(Int32)),
    pwr_lvl_max AggregateFunction(max, Nullable(Int32)),
    pwr_lvl_avg AggregateFunction(avg, Nullable(Int32)),

    -- Popup 통계
    popup_cnt_sum AggregateFunction(sum, Nullable(Int32)),
    popup_cnt_avg AggregateFunction(avg, Nullable(Int32)),
    popup_cnt_max AggregateFunction(max, Nullable(Int32)),

    -- 품질등급별 카운트
    stb_qlty_gr_count AggregateFunction(count, String),
    pwr_lvl_qlty_gr_count AggregateFunction(count, String),
    snr_qlty_gr_count AggregateFunction(count, String),

    -- CM 수 (셀 내 장비 수)
    cm_count AggregateFunction(uniq, Nullable(String)),

    -- 삽입 시간
    insert_time DateTime DEFAULT now()
)
ENGINE = ReplicatedAggregatingMergeTree('/clickhouse/tables/{shard}/catv/stb_monitoring_cell_hourly', '{replica}')
PARTITION BY toYYYYMM(mntr_hour)
ORDER BY (mntr_hour, TPO_CD, INFO_CNTR_CD, HPY_CNTR_ORG_ID, NET_CL_CD,
          L3_EQUIP_ID, L2_EQUIP_ID, CCELL_NUM, CHNL_FREQ_VAL,
          STB_QLTY_GR_CD, STB_MDL_NM, SCRBR_SO_NM)
TTL mntr_hour + INTERVAL 3 YEAR
SETTINGS index_granularity = 8192;
-- +goose StatementEnd

-- +goose StatementBegin
-- ========================================
-- 2. mv_stb_monitoring_cell_hourly Materialized View 생성
-- ========================================
CREATE MATERIALIZED VIEW IF NOT EXISTS catv.mv_stb_monitoring_cell_hourly ON CLUSTER clickhouse_cluster
TO catv.stb_monitoring_cell_hourly
AS
SELECT
    -- 시간: 1시간 단위로 truncate
    toStartOfHour(mntr_datetime) AS mntr_hour,

    -- 위치 정보
    TPO_CD,
    INFO_CNTR_CD,
    HPY_CNTR_ORG_ID,
    NET_CL_CD,

    -- 네트워크 장비 정보
    L3_EQUIP_ID,
    L2_EQUIP_ID,
    CCELL_NUM,

    -- 품질 및 기타 정보
    STB_QLTY_GR_CD,
    STB_MDL_NM,
    SCRBR_SO_NM,
    replace(CHNL_FREQ_VAL, ' ', '') AS CHNL_FREQ_VAL,

    -- SNR QAM 통계 (FREQ_METH_CTT = '256QAM'인 경우만)
    quantileState(0.1)(if(FREQ_METH_CTT = '256QAM', SNR_VAL, NULL)) AS snr_qam_p10,
    quantileState(0.5)(if(FREQ_METH_CTT = '256QAM', SNR_VAL, NULL)) AS snr_qam_p50,
    quantileState(0.9)(if(FREQ_METH_CTT = '256QAM', SNR_VAL, NULL)) AS snr_qam_p90,
    minState(if(FREQ_METH_CTT = '256QAM', SNR_VAL, NULL)) AS snr_qam_min,
    maxState(if(FREQ_METH_CTT = '256QAM', SNR_VAL, NULL)) AS snr_qam_max,
    avgState(if(FREQ_METH_CTT = '256QAM', SNR_VAL, NULL)) AS snr_qam_avg,

    -- SNR 8VSB 통계 (FREQ_METH_CTT = '8VSB'인 경우만)
    quantileState(0.1)(if(FREQ_METH_CTT = '8VSB', SNR_VAL, NULL)) AS snr_8vsb_p10,
    quantileState(0.5)(if(FREQ_METH_CTT = '8VSB', SNR_VAL, NULL)) AS snr_8vsb_p50,
    quantileState(0.9)(if(FREQ_METH_CTT = '8VSB', SNR_VAL, NULL)) AS snr_8vsb_p90,
    minState(if(FREQ_METH_CTT = '8VSB', SNR_VAL, NULL)) AS snr_8vsb_min,
    maxState(if(FREQ_METH_CTT = '8VSB', SNR_VAL, NULL)) AS snr_8vsb_max,
    avgState(if(FREQ_METH_CTT = '8VSB', SNR_VAL, NULL)) AS snr_8vsb_avg,

    -- Power Level 통계
    quantileState(0.1)(PWR_LVL_VAL) AS pwr_lvl_p10,
    quantileState(0.5)(PWR_LVL_VAL) AS pwr_lvl_p50,
    quantileState(0.9)(PWR_LVL_VAL) AS pwr_lvl_p90,
    minState(PWR_LVL_VAL) AS pwr_lvl_min,
    maxState(PWR_LVL_VAL) AS pwr_lvl_max,
    avgState(PWR_LVL_VAL) AS pwr_lvl_avg,

    -- Popup 통계
    sumState(SGNL_WEAK_POPUP_CNT) AS popup_cnt_sum,
    avgState(SGNL_WEAK_POPUP_CNT) AS popup_cnt_avg,
    maxState(SGNL_WEAK_POPUP_CNT) AS popup_cnt_max,

    -- 품질등급별 카운트
    countState(STB_QLTY_GR_CD) AS stb_qlty_gr_count,
    countState(PWR_LVL_STB_QLTY_GR_CD) AS pwr_lvl_qlty_gr_count,
    countState(SNR_STB_QLTY_GR_CD) AS snr_qlty_gr_count,

    -- CM 수 (셀 내 unique CM 장비 수)
    uniqState(CM_MAC_ADDR) AS cm_count

FROM catv.stb_monitoring
WHERE (SNR_VAL IS NULL OR (SNR_VAL >= 0 AND SNR_VAL <= 100))
  AND (PWR_LVL_VAL IS NULL OR (PWR_LVL_VAL >= -100 AND PWR_LVL_VAL <= 100))
  AND (SGNL_WEAK_POPUP_CNT IS NULL OR SGNL_WEAK_POPUP_CNT >= 0)

GROUP BY
    mntr_hour,
    TPO_CD,
    INFO_CNTR_CD,
    HPY_CNTR_ORG_ID,
    NET_CL_CD,
    L3_EQUIP_ID,
    L2_EQUIP_ID,
    CCELL_NUM,
    CHNL_FREQ_VAL,
    STB_QLTY_GR_CD,
    STB_MDL_NM,
    SCRBR_SO_NM;
-- +goose StatementEnd

-- +goose StatementBegin
-- ========================================
-- 3. dist_stb_monitoring_cell_hourly Distributed 테이블 생성
-- ========================================
CREATE TABLE IF NOT EXISTS catv.dist_stb_monitoring_cell_hourly ON CLUSTER clickhouse_cluster
AS catv.stb_monitoring_cell_hourly
ENGINE = Distributed(
    'clickhouse_cluster_replicated',  -- 클러스터 이름
    'catv',                           -- 데이터베이스
    'stb_monitoring_cell_hourly',     -- 로컬 테이블
    rand()                            -- Sharding key (랜덤 분산)
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS catv.dist_stb_monitoring_cell_hourly ON CLUSTER clickhouse_cluster SYNC;
-- +goose StatementEnd

-- +goose StatementBegin
DROP TABLE IF EXISTS catv.mv_stb_monitoring_cell_hourly ON CLUSTER clickhouse_cluster SYNC;
-- +goose StatementEnd

-- +goose StatementBegin
DROP TABLE IF EXISTS catv.stb_monitoring_cell_hourly ON CLUSTER clickhouse_cluster_replicated SYNC;
-- +goose StatementEnd
