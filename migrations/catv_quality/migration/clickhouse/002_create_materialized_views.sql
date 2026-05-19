-- +goose Up
-- +goose StatementBegin
-- ========================================
-- 1. mv_stb_monitoring 생성
-- stb_monitoring_raw 테이블에 데이터가 들어오면 트리거되어 파싱 후 stb_monitoring에 저장함
-- Vector를 통해 stb_monitoring_raw에 직접 insert된 데이터를 처리함
-- ========================================
CREATE MATERIALIZED VIEW IF NOT EXISTS catv.mv_stb_monitoring ON CLUSTER clickhouse_cluster
TO catv.stb_monitoring
AS SELECT
    parseDateTime(JSONExtract(raw_message, 'MNTR_DTHM', 'String'), '%Y%m%d%H%i', 'Asia/Seoul') AS mntr_datetime,
    parseDateTime(JSONExtract(raw_message, 'EVT_RCV_DTM', 'String'), '%Y%m%d%H%i%s', 'Asia/Seoul') AS evt_rcv_datetime,
    insert_time AS insert_time,
    JSONExtract(raw_message, 'MNTR_DTHM', 'Nullable(String)') AS MNTR_DTHM,
    JSONExtract(raw_message, 'STB_MAC_ADDR', 'Nullable(String)') AS STB_MAC_ADDR,
    JSONExtract(raw_message, 'UUID_VAL', 'Nullable(String)') AS UUID_VAL,
    JSONExtract(raw_message, 'AUDIT_ID', 'Nullable(String)') AS AUDIT_ID,
    JSONExtract(raw_message, 'EVT_RCV_DTM', 'Nullable(String)') AS EVT_RCV_DTM,
    JSONExtract(raw_message, 'CLCT_SVR_NUM', 'Nullable(Int64)') AS CLCT_SVR_NUM,
    JSONExtract(raw_message, 'SRC_IP_ADDR', 'Nullable(String)') AS SRC_IP_ADDR,
    JSONExtract(raw_message, 'WEQP_MGMT_NUM', 'Nullable(String)') AS WEQP_MGMT_NUM,
    JSONExtract(raw_message, 'STB_EQP_MDL_CD', 'Nullable(String)') AS STB_EQP_MDL_CD,
    JSONExtract(raw_message, 'PATCH_GRP_CTT', 'Nullable(String)') AS PATCH_GRP_CTT,
    JSONExtract(raw_message, 'EQP_UPTIME_TMS_INFO', 'Nullable(String)') AS EQP_UPTIME_TMS_INFO,
    JSONExtract(raw_message, 'INFO_CNTR_CD', 'Nullable(String)') AS INFO_CNTR_CD,
    JSONExtract(raw_message, 'TPO_CD', 'Nullable(String)') AS TPO_CD,
    JSONExtract(raw_message, 'HPY_CNTR_ORG_ID', 'Nullable(String)') AS HPY_CNTR_ORG_ID,
    JSONExtract(raw_message, 'NET_CL_CD', 'Nullable(String)') AS NET_CL_CD,
    JSONExtract(raw_message, 'L3_EQUIP_ID', 'Nullable(String)') AS L3_EQUIP_ID,
    JSONExtract(raw_message, 'L3_EQUIP_TID_VAL', 'Nullable(String)') AS L3_EQUIP_TID_VAL,
    JSONExtract(raw_message, 'L3_EQUIP_MDL_CD', 'Nullable(String)') AS L3_EQUIP_MDL_CD,
    JSONExtract(raw_message, 'L2_EQUIP_ID', 'Nullable(String)') AS L2_EQUIP_ID,
    JSONExtract(raw_message, 'CCELL_NUM', 'Nullable(String)') AS CCELL_NUM,
    JSONExtract(raw_message, 'L2_EQUIP_TID_VAL', 'Nullable(String)') AS L2_EQUIP_TID_VAL,
    JSONExtract(raw_message, 'L2_EQUIP_MDL_CD', 'Nullable(String)') AS L2_EQUIP_MDL_CD,
    JSONExtract(raw_message, 'IPTV_SVC_MGMT_NUM', 'Nullable(Int64)') AS IPTV_SVC_MGMT_NUM,
    JSONExtract(raw_message, 'IPTV_SVC_NUM', 'Nullable(Int64)') AS IPTV_SVC_NUM,
    JSONExtract(raw_message, 'STB_EVT_VER_INFO', 'Nullable(String)') AS STB_EVT_VER_INFO,
    JSONExtract(raw_message, 'STB_EVT_SER_NUM', 'Nullable(Int64)') AS STB_EVT_SER_NUM,
    JSONExtract(raw_message, 'STB_VER_CTT', 'Nullable(String)') AS STB_VER_CTT,
    JSONExtract(raw_message, 'STB_MDL_NM', 'Nullable(String)') AS STB_MDL_NM,
    JSONExtract(raw_message, 'STB_QLTY_GR_CD', 'Nullable(String)') AS STB_QLTY_GR_CD,
    JSONExtract(raw_message, 'PWR_LVL_VAL', 'Nullable(Int32)') AS PWR_LVL_VAL,
    JSONExtract(raw_message, 'PWR_LVL_STB_QLTY_GR_CD', 'Nullable(String)') AS PWR_LVL_STB_QLTY_GR_CD,
    JSONExtract(raw_message, 'SNR_VAL', 'Nullable(Int32)') AS SNR_VAL,
    if(JSONExtract(raw_message, 'FREQ_METH_CTT', 'Nullable(String)') = '256QAM',
       JSONExtract(raw_message, 'SNR_VAL', 'Nullable(Int32)'), NULL) AS snr_qam,
    if(JSONExtract(raw_message, 'FREQ_METH_CTT', 'Nullable(String)') = '8VSB',
       JSONExtract(raw_message, 'SNR_VAL', 'Nullable(Int32)'), NULL) AS snr_8vsb,
    JSONExtract(raw_message, 'SNR_STB_QLTY_GR_CD', 'Nullable(String)') AS SNR_STB_QLTY_GR_CD,
    JSONExtract(raw_message, 'SGNL_WEAK_POPUP_CNT', 'Nullable(Int32)') AS SGNL_WEAK_POPUP_CNT,
    JSONExtract(raw_message, 'SGNL_WEAK_POPUP_STB_QLTY_GR_CD', 'Nullable(String)') AS SGNL_WEAK_POPUP_STB_QLTY_GR_CD,
    JSONExtract(raw_message, 'SGNL_WEAK_POPUP_OCCR_YN', 'Nullable(String)') AS SGNL_WEAK_POPUP_OCCR_YN,
    JSONExtract(raw_message, 'TV_CL_CD', 'Nullable(String)') AS TV_CL_CD,
    JSONExtract(raw_message, 'SCRBR_SO_NM', 'Nullable(String)') AS SCRBR_SO_NM,
    JSONExtract(raw_message, 'CHNL_ID', 'Nullable(String)') AS CHNL_ID,
    JSONExtract(raw_message, 'CHNL_NUM', 'Nullable(String)') AS CHNL_NUM,
    JSONExtract(raw_message, 'CHNL_NM', 'Nullable(String)') AS CHNL_NM,
    JSONExtract(raw_message, 'PGM_CTT', 'Nullable(String)') AS PGM_CTT,
    JSONExtract(raw_message, 'CHNL_FREQ_VAL', 'Nullable(String)') AS CHNL_FREQ_VAL,
    replaceAll(JSONExtract(raw_message, 'CHNL_FREQ_VAL', 'Nullable(String)'), ' ', '') AS CHNL_FREQ_VAL_NORMALIZED,
    JSONExtract(raw_message, 'FREQ_METH_CTT', 'Nullable(String)') AS FREQ_METH_CTT,
    JSONExtract(raw_message, 'CATV_STB_PWR_CD', 'Nullable(String)') AS CATV_STB_PWR_CD,
    JSONExtract(raw_message, 'CM_MAC_ADDR', 'Nullable(String)') AS CM_MAC_ADDR,
    JSONExtract(raw_message, 'CM_IP_ADDR', 'Nullable(String)') AS CM_IP_ADDR,
    JSONExtract(raw_message, 'LOC_UI_VER_CTT', 'Nullable(String)') AS LOC_UI_VER_CTT,
    JSONExtract(raw_message, 'CLD_UI_VER_CTT', 'Nullable(String)') AS CLD_UI_VER_CTT
FROM catv.stb_monitoring_raw
WHERE isNotNull(parseDateTimeOrNull(JSONExtract(raw_message, 'MNTR_DTHM', 'String'), '%Y%m%d%H%i'));
-- +goose StatementEnd

-- +goose StatementBegin
-- ========================================
-- 2. mv_stb_monitoring_error 생성
-- MNTR_DTHM 파싱에 실패한 데이터를 stb_monitoring_error 테이블에 저장함
-- ========================================
CREATE MATERIALIZED VIEW IF NOT EXISTS catv.mv_stb_monitoring_error ON CLUSTER clickhouse_cluster
TO catv.stb_monitoring_error
AS SELECT
    now() AS insert_time,
    raw_message,
    'Invalid MNTR_DTHM format or missing' AS error_reason
FROM catv.stb_monitoring_raw
WHERE isNull(parseDateTimeOrNull(JSONExtract(raw_message, 'MNTR_DTHM', 'String'), '%Y%m%d%H%i'));
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP VIEW IF EXISTS catv.mv_stb_monitoring ON CLUSTER clickhouse_cluster SYNC;
-- +goose StatementEnd

-- +goose StatementBegin
DROP VIEW IF EXISTS catv.mv_stb_monitoring_error ON CLUSTER clickhouse_cluster SYNC;
-- +goose StatementEnd
