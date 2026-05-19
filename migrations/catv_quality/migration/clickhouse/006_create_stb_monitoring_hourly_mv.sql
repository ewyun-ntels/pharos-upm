-- +goose Up
-- +goose StatementBegin
-- ========================================
-- 1. mv_stb_monitoring_hourly 생성
-- stb_monitoring 테이블에 데이터가 들어오면 트리거되어 1시간 단위로 집계하여 저장함
-- ========================================
CREATE MATERIALIZED VIEW IF NOT EXISTS catv.mv_stb_monitoring_hourly ON CLUSTER clickhouse_cluster
TO catv.stb_monitoring_hourly
AS SELECT
    toStartOfHour(mntr_datetime) AS mntr_hour,
    ifNull(STB_MDL_NM, 'Unknown') AS STB_MDL_NM,
    ifNull(TPO_CD, 'Unknown') AS TPO_CD,
    ifNull(NET_CL_CD, 'Unknown') AS NET_CL_CD,
    ifNull(INFO_CNTR_CD, 'Unknown') AS INFO_CNTR_CD,
    ifNull(HPY_CNTR_ORG_ID, 'Unknown') AS HPY_CNTR_ORG_ID,
    ifNull(STB_QLTY_GR_CD, 'Unknown') AS STB_QLTY_GR_CD,
    ifNull(SCRBR_SO_NM, 'Unknown') AS SCRBR_SO_NM,
    ifNull(L3_EQUIP_ID, 'Unknown') AS L3_EQUIP_ID,
    ifNull(L2_EQUIP_ID, 'Unknown') AS L2_EQUIP_ID,
    ifNull(STB_VER_CTT, 'Unknown') AS STB_VER_CTT,
    replaceAll(ifNull(CHNL_FREQ_VAL, 'Unknown'), ' ', '') AS CHNL_FREQ_VAL,
    ifNull(CM_MAC_ADDR, 'Unknown') AS CM_MAC_ADDR,
    ifNull(CM_IP_ADDR, 'Unknown') AS CM_IP_ADDR,
    ifNull(LOC_UI_VER_CTT, 'Unknown') AS LOC_UI_VER_CTT,
    ifNull(CLD_UI_VER_CTT, 'Unknown') AS CLD_UI_VER_CTT,
    ifNull(CCELL_NUM, 'Unknown') AS CCELL_NUM,
    
    -- SNR QAM 집계 (FREQ_METH_CTT가 256QAM인 경우만)
    quantileState(0.1)(if(FREQ_METH_CTT = '256QAM', SNR_VAL, NULL)) AS snr_qam_p10,
    quantileState(0.5)(if(FREQ_METH_CTT = '256QAM', SNR_VAL, NULL)) AS snr_qam_p50,
    quantileState(0.9)(if(FREQ_METH_CTT = '256QAM', SNR_VAL, NULL)) AS snr_qam_p90,
    minState(if(FREQ_METH_CTT = '256QAM', SNR_VAL, NULL)) AS snr_qam_min,
    maxState(if(FREQ_METH_CTT = '256QAM', SNR_VAL, NULL)) AS snr_qam_max,
    avgState(if(FREQ_METH_CTT = '256QAM', SNR_VAL, NULL)) AS snr_qam_avg,
    
    -- SNR 8VSB 집계 (FREQ_METH_CTT가 8VSB인 경우만)
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
    countState(SNR_STB_QLTY_GR_CD) AS snr_qlty_gr_count,
    
    now() AS insert_time
FROM catv.stb_monitoring
WHERE (SNR_VAL IS NULL OR (SNR_VAL >= 0 AND SNR_VAL <= 100))
  AND (PWR_LVL_VAL IS NULL OR (PWR_LVL_VAL >= -100 AND PWR_LVL_VAL <= 100))
  AND (SGNL_WEAK_POPUP_CNT IS NULL OR SGNL_WEAK_POPUP_CNT >= 0)
GROUP BY mntr_hour, STB_MDL_NM, TPO_CD, NET_CL_CD, INFO_CNTR_CD, HPY_CNTR_ORG_ID, STB_QLTY_GR_CD, SCRBR_SO_NM, L3_EQUIP_ID, L2_EQUIP_ID, CCELL_NUM, STB_VER_CTT, CHNL_FREQ_VAL, CM_MAC_ADDR, CM_IP_ADDR, LOC_UI_VER_CTT, CLD_UI_VER_CTT;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP VIEW IF EXISTS catv.mv_stb_monitoring_hourly ON CLUSTER clickhouse_cluster SYNC;
-- +goose StatementEnd
