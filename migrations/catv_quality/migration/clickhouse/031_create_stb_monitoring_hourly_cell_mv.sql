-- +goose Up
-- +goose StatementBegin
-- ========================================
-- mv_stb_monitoring_hourly_cell 생성
-- stb_monitoring 테이블에서 Cell별로 1시간 집계
-- ========================================
CREATE MATERIALIZED VIEW IF NOT EXISTS catv.mv_stb_monitoring_hourly_cell ON CLUSTER clickhouse_cluster
TO catv.stb_monitoring_hourly_cell
AS SELECT
    toStartOfHour(mntr_datetime) AS mntr_hour,
    ifNull(CCELL_NUM, 'Unknown') AS CCELL_NUM,

    -- 메타데이터
    ifNull(STB_MDL_NM, 'Unknown') AS STB_MDL_NM,
    ifNull(TPO_CD, 'Unknown') AS TPO_CD,
    ifNull(NET_CL_CD, 'Unknown') AS NET_CL_CD,
    ifNull(INFO_CNTR_CD, 'Unknown') AS INFO_CNTR_CD,
    ifNull(HPY_CNTR_ORG_ID, 'Unknown') AS HPY_CNTR_ORG_ID,
    ifNull(SCRBR_SO_NM, 'Unknown') AS SCRBR_SO_NM,
    ifNull(L3_EQUIP_ID, 'Unknown') AS L3_EQUIP_ID,
    ifNull(L2_EQUIP_ID, 'Unknown') AS L2_EQUIP_ID,
    ifNull(STB_VER_CTT, 'Unknown') AS STB_VER_CTT,

    -- SNR QAM (0 제외)
    sum(if(FREQ_METH_CTT = '256QAM', toFloat64(SNR_VAL), 0)) AS snr_qam_sum,
    countIf(FREQ_METH_CTT = '256QAM' AND SNR_VAL IS NOT NULL) AS snr_qam_count,
    minIf(SNR_VAL, FREQ_METH_CTT = '256QAM' AND SNR_VAL > 0) AS snr_qam_min,
    maxIf(SNR_VAL, FREQ_METH_CTT = '256QAM' AND SNR_VAL > 0) AS snr_qam_max,

    -- SNR 8VSB (0 제외)
    sum(if(FREQ_METH_CTT = '8VSB', toFloat64(SNR_VAL), 0)) AS snr_8vsb_sum,
    countIf(FREQ_METH_CTT = '8VSB' AND SNR_VAL IS NOT NULL) AS snr_8vsb_count,
    minIf(SNR_VAL, FREQ_METH_CTT = '8VSB' AND SNR_VAL > 0) AS snr_8vsb_min,
    maxIf(SNR_VAL, FREQ_METH_CTT = '8VSB' AND SNR_VAL > 0) AS snr_8vsb_max,

    -- PWR LVL (0 포함, NULL 제외)
    sum(toFloat64(ifNull(PWR_LVL_VAL, 0))) AS pwr_lvl_sum,
    count(PWR_LVL_VAL) AS pwr_lvl_count,
    min(PWR_LVL_VAL) AS pwr_lvl_min,
    max(PWR_LVL_VAL) AS pwr_lvl_max,

    -- 신호 미약 팝업
    sum(ifNull(SGNL_WEAK_POPUP_CNT, 0)) AS popup_cnt_sum,
    max(ifNull(SGNL_WEAK_POPUP_CNT, 0)) AS popup_cnt_max,

    -- 레코드 수
    count() AS record_count,

    now() AS insert_time
FROM catv.stb_monitoring
WHERE (SNR_VAL IS NULL OR (SNR_VAL >= 0 AND SNR_VAL <= 100))
  AND (PWR_LVL_VAL IS NULL OR (PWR_LVL_VAL >= -100 AND PWR_LVL_VAL <= 100))
  AND (SGNL_WEAK_POPUP_CNT IS NULL OR SGNL_WEAK_POPUP_CNT >= 0)
GROUP BY mntr_hour, CCELL_NUM, STB_MDL_NM, TPO_CD, NET_CL_CD, INFO_CNTR_CD, HPY_CNTR_ORG_ID, SCRBR_SO_NM, L3_EQUIP_ID, L2_EQUIP_ID, STB_VER_CTT;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP VIEW IF EXISTS catv.mv_stb_monitoring_hourly_cell ON CLUSTER clickhouse_cluster SYNC;
-- +goose StatementEnd
