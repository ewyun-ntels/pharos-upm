-- +goose Up
-- +goose StatementBegin
-- ========================================
-- 1. stb_monitoring 테이블 생성 (최종 파싱 데이터 저장용, TTL 3년)
-- ========================================
CREATE TABLE IF NOT EXISTS catv.stb_monitoring ON CLUSTER clickhouse_cluster
(
    `mntr_datetime` DateTime,
    `evt_rcv_datetime` Nullable(DateTime),
    `insert_time` DateTime DEFAULT now(),
    `MNTR_DTHM` Nullable(String) COMMENT '감시일자시분',
    `STB_MAC_ADDR` Nullable(String) COMMENT 'STBMAC주소',
    `UUID_VAL` Nullable(String) COMMENT 'UUID값',
    `AUDIT_ID` Nullable(String) COMMENT '최종변경자ID',
    `EVT_RCV_DTM` Nullable(String) COMMENT '이벤트수신일시',
    `CLCT_SVR_NUM` Nullable(Int64) COMMENT '수집서버번호',
    `SRC_IP_ADDR` Nullable(String) COMMENT '소스IP주소',
    `WEQP_MGMT_NUM` Nullable(String) COMMENT '유선단말관리번호',
    `STB_EQP_MDL_CD` Nullable(String) COMMENT 'STB단말모델코드',
    `PATCH_GRP_CTT` Nullable(String) COMMENT '패치그룹내용',
    `EQP_UPTIME_TMS_INFO` Nullable(String) COMMENT '단말UPTIME시간정보',
    `INFO_CNTR_CD` Nullable(String) COMMENT '정보센터코드',
    `TPO_CD` Nullable(String) COMMENT '국사코드',
    `HPY_CNTR_ORG_ID` Nullable(String) COMMENT '행복센터조직ID',
    `NET_CL_CD` Nullable(String) COMMENT '망구분코드',
    `L3_EQUIP_ID` Nullable(String) COMMENT 'L3장비ID',
    `L3_EQUIP_TID_VAL` Nullable(String) COMMENT 'L3장비TID값',
    `L3_EQUIP_MDL_CD` Nullable(String) COMMENT 'L3장비모델코드',
    `L2_EQUIP_ID` Nullable(String) COMMENT 'L2장비ID',
    `CCELL_NUM` Nullable(String) COMMENT '커버리지셀번호',
    `L2_EQUIP_TID_VAL` Nullable(String) COMMENT 'L2장비TID값',
    `L2_EQUIP_MDL_CD` Nullable(String) COMMENT 'L2장비모델코드',
    `IPTV_SVC_MGMT_NUM` Nullable(Int64) COMMENT 'IPTV서비스관리번호',
    `IPTV_SVC_NUM` Nullable(Int64) COMMENT 'IPTV서비스번호',
    `STB_EVT_VER_INFO` Nullable(String) COMMENT 'STB이벤트버전정보',
    `STB_EVT_SER_NUM` Nullable(Int64) COMMENT 'STB이벤트일련번호',
    `STB_VER_CTT` Nullable(String) COMMENT 'STB버전내용',
    `STB_MDL_NM` Nullable(String) COMMENT 'STB모델명',
    `STB_QLTY_GR_CD` Nullable(String) COMMENT 'STB품질등급코드',
    `PWR_LVL_VAL` Nullable(Int32) COMMENT '파워레벨값',
    `PWR_LVL_STB_QLTY_GR_CD` Nullable(String) COMMENT '파워레벨STB품질등급코드',
    `SNR_VAL` Nullable(Int32) COMMENT 'SNR값',
    `snr_qam` Nullable(Int32) COMMENT 'QAM SNR값',
    `snr_8vsb` Nullable(Int32) COMMENT '8VSB SNR값',
    `SNR_STB_QLTY_GR_CD` Nullable(String) COMMENT 'SNRSTB품질등급코드',
    `SGNL_WEAK_POPUP_CNT` Nullable(Int32) COMMENT '신호미약팝업건수',
    `SGNL_WEAK_POPUP_STB_QLTY_GR_CD` Nullable(String) COMMENT '신호미약팝업STB품질등급코드',
    `SGNL_WEAK_POPUP_OCCR_YN` Nullable(String) COMMENT '신호미약팝업발생여부',
    `TV_CL_CD` Nullable(String) COMMENT 'TV구분코드',
    `SCRBR_SO_NM` Nullable(String) COMMENT '가입자SO명',
    `CHNL_ID` Nullable(String) COMMENT '채널ID',
    `CHNL_NUM` Nullable(String) COMMENT '채널번호',
    `CHNL_NM` Nullable(String) COMMENT '채널명',
    `PGM_CTT` Nullable(String) COMMENT '프로그램내용',
    `CHNL_FREQ_VAL` Nullable(String) COMMENT '채널주파수값',
    `CHNL_FREQ_VAL_NORMALIZED` Nullable(String) COMMENT '정규화된채널주파수값(공백제거)',
    `FREQ_METH_CTT` Nullable(String) COMMENT '주파수방식내용',
    `CATV_STB_PWR_CD` Nullable(String) COMMENT 'CATVSTB전원코드',
    `CM_MAC_ADDR` Nullable(String) COMMENT 'CMMAC주소',
    `CM_IP_ADDR` Nullable(String) COMMENT 'CMIP주소',
    `LOC_UI_VER_CTT` Nullable(String) COMMENT '로컬UI버전내용',
    `CLD_UI_VER_CTT` Nullable(String) COMMENT '클라우드UI버전내용'
)
ENGINE = MergeTree()
PARTITION BY toYYYYMMDD(mntr_datetime)
ORDER BY (mntr_datetime)
TTL mntr_datetime + INTERVAL 1095 DAY
SETTINGS index_granularity = 8192;
-- +goose StatementEnd

-- +goose StatementBegin
-- ========================================
-- 2. stb_monitoring_raw 테이블 생성 (원본 보관용, TTL 90일)
-- Vector를 통해 직접 데이터를 insert함
-- ========================================
CREATE TABLE IF NOT EXISTS catv.stb_monitoring_raw ON CLUSTER clickhouse_cluster
(
    `insert_time` DateTime DEFAULT now(),
    `raw_message` String
)
ENGINE = MergeTree()
PARTITION BY toYYYYMMDD(insert_time)
ORDER BY insert_time
TTL insert_time + INTERVAL 90 DAY
SETTINGS index_granularity = 8192;
-- +goose StatementEnd

-- +goose StatementBegin
-- ========================================
-- 3. stb_monitoring_error 테이블 생성 (파싱 에러 데이터 보관용)
-- ========================================
CREATE TABLE IF NOT EXISTS catv.stb_monitoring_error ON CLUSTER clickhouse_cluster
(
    `insert_time` DateTime DEFAULT now(),
    `raw_message` String,
    `error_reason` String
)
ENGINE = MergeTree()
PARTITION BY toYYYYMMDD(insert_time)
ORDER BY insert_time
TTL insert_time + INTERVAL 90 DAY;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS catv.stb_monitoring_error ON CLUSTER clickhouse_cluster SYNC;
-- +goose StatementEnd

-- +goose StatementBegin
DROP TABLE IF EXISTS catv.stb_monitoring_raw ON CLUSTER clickhouse_cluster SYNC;
-- +goose StatementEnd

-- +goose StatementBegin
DROP TABLE IF EXISTS catv.stb_monitoring ON CLUSTER clickhouse_cluster SYNC;
-- +goose StatementEnd
