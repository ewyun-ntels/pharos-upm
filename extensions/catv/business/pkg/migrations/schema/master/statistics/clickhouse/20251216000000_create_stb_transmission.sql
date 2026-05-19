-- +goose Up
-- +goose StatementBegin
-- 원본 데이터 저장 테이블 (파싱 오류 발생 시에도 원본은 보존)
CREATE TABLE IF NOT EXISTS catv.stb_transmission_raw ON CLUSTER clickhouse_cluster_replicated
(
    timestamp DateTime COMMENT '타임스탬프',
    transmission_type String COMMENT '전송 유형',
    raw_data String COMMENT '원본 데이터',
    errors Array(String) COMMENT '파싱 중 발생한 오류 메시지 목록'
)
ENGINE = ReplicatedMergeTree('/clickhouse/tables/{shard}/catv/stb_transmission_raw', '{replica}')
PARTITION BY toYYYYMM(timestamp)
ORDER BY (timestamp, transmission_type)
TTL timestamp + toIntervalYear(3);
-- +goose StatementEnd

-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS catv.dist_stb_transmission_raw ON CLUSTER clickhouse_cluster_replicated
AS catv.stb_transmission_raw
ENGINE = Distributed(clickhouse_cluster_replicated, catv, stb_transmission_raw, sipHash64(transmission_type));
-- +goose StatementEnd

-- +goose StatementBegin
-- 주기전송 (Periodic Transmission) - 10분마다 UDP 전송
CREATE TABLE IF NOT EXISTS catv.stb_transmission_periodic ON CLUSTER clickhouse_cluster_replicated
(
    timestamp           DateTime COMMENT '타임스탬프',
    host_id             String COMMENT 'HOST ID: 10자리 HOST 구분 ID',
    mac_addr            String COMMENT 'STB MAC: 12자리 표준 MAC Address',
    cm_mac              String COMMENT 'CM MAC: 12자리 표준 MAC Address',
    stb_ip              String COMMENT 'STB IP',
    cm_ip               String COMMENT 'CM IP',
    stb_model           String COMMENT 'STB 모델명',
    mw_ver              String COMMENT '서비스 버전: STB 미들웨어 버전',
    local_ver           String COMMENT 'Local UI 버전',
    cloud_ver           String COMMENT 'Cloud UI 버전',
    logging_time        String COMMENT '정보 수집 시간',
    logging_time_utc    Nullable(DateTime) COMMENT '정보 수집 시간 (UTC), 원본에 없는 컬럼',
    sending_time        String COMMENT '정보 전송 시간',
    sending_time_utc    Nullable(DateTime) COMMENT '정보 전송 시간 (UTC), 원본에 없는 컬럼',
    ch_sid              Nullable(String) COMMENT '채널 Source ID',
    ch_num              Nullable(String) COMMENT '채널 번호',
    ch_name             Nullable(String) COMMENT '채널명',
    ch_prg              Nullable(String) COMMENT '프로그램 명',
    ch_freq             Nullable(String) COMMENT '채널 주파수 (MHz)',
    ch_mode             Nullable(String) COMMENT '변조방식 (8VSB/256QAM)',
    pwr_lvl             Nullable(String) COMMENT 'Power Level: 신호 세기',
    snr                 Nullable(String) COMMENT 'SNR: 신호 품질',
    sig_weak            Nullable(String) COMMENT '신호미약 팝업 발생 (Y/N)',
    sig_weak_cnt        Nullable(String) COMMENT '팝업 발생 Count',
    stb_state           Nullable(String) COMMENT 'STB 전원 상태 (watching/standby)',
    running_time        Nullable(String) COMMENT 'STB 구동 시간 정보',
    running_time_sec    Nullable(UInt64) COMMENT 'STB 구동 시간 정보 (초), 원본에 없는 컬럼'
)
ENGINE = ReplicatedMergeTree('/clickhouse/tables/{shard}/catv/stb_transmission_periodic', '{replica}')
PARTITION BY toYYYYMM(timestamp)
ORDER BY (timestamp, host_id)
TTL timestamp + toIntervalYear(3);
-- +goose StatementEnd

-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS catv.dist_stb_transmission_periodic ON CLUSTER clickhouse_cluster_replicated
AS catv.stb_transmission_periodic
ENGINE = Distributed(clickhouse_cluster_replicated, catv, stb_transmission_periodic, sipHash64(host_id));
-- +goose StatementEnd

-- +goose StatementBegin
-- 일일전송 (Daily Transmission) - 일일 UDP 전송
CREATE TABLE IF NOT EXISTS catv.stb_transmission_daily ON CLUSTER clickhouse_cluster_replicated
(
    timestamp           DateTime COMMENT '타임스탬프',
    host_id             String COMMENT 'HOST ID: 10자리 HOST 구분 ID',
    mac_addr            String COMMENT 'STB MAC: 12자리 표준 MAC Address',
    cm_mac              String COMMENT 'CM MAC: 12자리 표준 MAC Address',
    stb_ip              String COMMENT 'STB IP',
    cm_ip               String COMMENT 'CM IP',
    stb_model           String COMMENT 'STB 모델명',
    mw_ver              String COMMENT '서비스 버전: STB 미들웨어 버전',
    local_ver           String COMMENT 'Local UI 버전',
    cloud_ver           String COMMENT 'Cloud UI 버전',
    logging_time        String COMMENT '정보 수집 시간',
    logging_time_utc    Nullable(DateTime) COMMENT '정보 수집 시간 (UTC), 원본에 없는 컬럼',
    sending_time        String COMMENT '정보 전송 시간',
    sending_time_utc    Nullable(DateTime) COMMENT '정보 전송 시간 (UTC), 원본에 없는 컬럼',
    limit_age           Nullable(String) COMMENT '시청 연령 제한',
    tv_lock             Nullable(String) COMMENT 'TV 잠금',
    skip_ch             Nullable(String) COMMENT '차단 채널',
    easy_buying         Nullable(String) COMMENT '간편 구매',
    fav_ch              Nullable(String) COMMENT '선호 채널',
    zapping_ad          Nullable(String) COMMENT '채널 전환 홍보',
    mini_epg            Nullable(String) COMMENT '채널 가이드 표시',
    mini_epg_ad         Nullable(String) COMMENT '채널 가이드 배너 노출',
    tv_caption          Nullable(String) COMMENT '자막 방송',
    tv_impaired         Nullable(String) COMMENT '화면 해설 방송',
    barker_ch           Nullable(String) COMMENT 'TV 시작 채널',
    vod_view            Nullable(String) COMMENT 'VOD 확인 방식',
    vod_relay           Nullable(String) COMMENT '회차 이어보기',
    resolution          Nullable(String) COMMENT '화면 비율',
    audio_mode          Nullable(String) COMMENT '오디오 출력',
    hdmi_cec            Nullable(String) COMMENT '전원 동기화: HDMI-CEC',
    hdcp                Nullable(String) COMMENT 'HDCP 설정',
    hdr                 Nullable(String) COMMENT 'HDR 기능',
    mobile_pay          Nullable(String) COMMENT '결제 방식 추가',
    morning_alarm       Nullable(String) COMMENT '모닝 알람',
    boot_menu           Nullable(String) COMMENT '홈 메뉴 노출',
    pms_on              Nullable(String) COMMENT '실시간 혜택 정보 제공',
    one_ad_on           Nullable(String) COMMENT '맞춤형 광고 보기',
    audio_lang          Nullable(String) COMMENT '음성 언어',
    standby_mode        Nullable(String) COMMENT '대기모드 전환',
    save_pwr            Nullable(String) COMMENT '저전력 모드',
    voice_guide         Nullable(String) COMMENT '음성 안내',
    running_time        Nullable(String) COMMENT 'STB 구동 시간 정보',
    running_time_sec    Nullable(UInt64) COMMENT 'STB 구동 시간 정보 (초), 원본에 없는 컬럼',
    limit_contents      Nullable(String) COMMENT '성인 콘텐츠 표시'
)
ENGINE = ReplicatedMergeTree('/clickhouse/tables/{shard}/catv/stb_transmission_daily', '{replica}')
PARTITION BY toYYYYMM(timestamp)
ORDER BY (timestamp, host_id)
TTL timestamp + toIntervalYear(3);
-- +goose StatementEnd

-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS catv.dist_stb_transmission_daily ON CLUSTER clickhouse_cluster_replicated
AS catv.stb_transmission_daily
ENGINE = Distributed(clickhouse_cluster_replicated, catv, stb_transmission_daily, sipHash64(host_id));
-- +goose StatementEnd

-- +goose StatementBegin
-- 품질계측전송 (Quality Measurement) - Sleep 모드 진입 시 UDP 전송
CREATE TABLE IF NOT EXISTS catv.stb_transmission_quality_measurement ON CLUSTER clickhouse_cluster_replicated
(
    timestamp           DateTime COMMENT '타임스탬프',
    host_id             String COMMENT 'HOST ID: 10자리 HOST 구분 ID',
    mac_addr            String COMMENT 'STB MAC: 12자리 표준 MAC Address',
    cm_mac              String COMMENT 'CM MAC: 12자리 표준 MAC Address',
    stb_ip              String COMMENT 'STB IP',
    cm_ip               String COMMENT 'CM IP',
    stb_model           String COMMENT 'STB 모델명',
    mw_ver              String COMMENT '서비스 버전: STB 미들웨어 버전',
    local_ver           String COMMENT 'Local UI 버전',
    cloud_ver           String COMMENT 'Cloud UI 버전',
    logging_time        String COMMENT '정보 수집 시간',
    logging_time_utc    Nullable(DateTime) COMMENT '정보 수집 시간 (UTC), 원본에 없는 컬럼',
    channels            String COMMENT '채널 품질 수집 목록 (JSON array)',
    ch_sid              Nullable(String) COMMENT '채널 sid',
    ch_num              Nullable(String) COMMENT '채널 number',
    ch_freq             Nullable(String) COMMENT 'frequency',
    ch_mode             Nullable(String) COMMENT 'modulator',
    pwr_lvl             Nullable(String) COMMENT 'powerLevel',
    snr                 Nullable(String) COMMENT 'snr'
)
ENGINE = ReplicatedMergeTree('/clickhouse/tables/{shard}/catv/stb_transmission_quality_measurement', '{replica}')
PARTITION BY toYYYYMM(timestamp)
ORDER BY (timestamp, host_id)
TTL timestamp + toIntervalYear(3);
-- +goose StatementEnd

-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS catv.dist_stb_transmission_quality_measurement ON CLUSTER clickhouse_cluster_replicated
AS catv.stb_transmission_quality_measurement
ENGINE = Distributed(clickhouse_cluster_replicated, catv, stb_transmission_quality_measurement, sipHash64(host_id));
-- +goose StatementEnd

-- +goose StatementBegin
-- 자가진단전송 (Self-Diagnosis) - UDP 전송 (핫키 *106OK 입력 시)
-- Note: TCP 자가진단 제어 (원격 요청)는 별도 HTTP 핸들러로 처리
CREATE TABLE IF NOT EXISTS catv.stb_transmission_diagnostic ON CLUSTER clickhouse_cluster_replicated
(
    timestamp           DateTime COMMENT '타임스탬프',
    host_id             String COMMENT 'HOST ID: 10자리 HOST 구분 ID',
    mac_addr            String COMMENT 'STB MAC: 12자리 표준 MAC Address',
    cm_mac              String COMMENT 'CM MAC: 12자리 표준 MAC Address',
    stb_ip              String COMMENT 'STB IP',
    cm_ip               String COMMENT 'CM IP',
    stb_model           String COMMENT 'STB 모델명',
    mw_ver              String COMMENT '서비스 버전: STB 미들웨어 버전',
    local_ver           String COMMENT 'Local UI 버전',
    cloud_ver           String COMMENT 'Cloud UI 버전',
    logging_time        String COMMENT '정보 수집 시간',
    logging_time_utc    Nullable(DateTime) COMMENT '정보 수집 시간 (UTC), 원본에 없는 컬럼',
    ch_sid              String COMMENT '채널 sid',
    ch_num              String COMMENT '채널 number',
    ch_freq             String COMMENT 'frequency',
    ch_mode             String COMMENT 'modulator',
    pwr_lvl             String COMMENT 'powerLevel',
    snr                 String COMMENT 'snr'
)
ENGINE = ReplicatedMergeTree('/clickhouse/tables/{shard}/catv/stb_transmission_diagnostic', '{replica}')
ORDER BY (timestamp, host_id)
TTL timestamp + toIntervalYear(3);
-- +goose StatementEnd

-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS catv.dist_stb_transmission_diagnostic ON CLUSTER clickhouse_cluster_replicated
AS catv.stb_transmission_diagnostic
ENGINE = Distributed(clickhouse_cluster_replicated, catv, stb_transmission_diagnostic, sipHash64(host_id));
-- +goose StatementEnd

-- +goose StatementBegin
-- 망품질전환전송 (Network Quality Switching) - UDP 전송
CREATE TABLE IF NOT EXISTS catv.stb_transmission_network_quality_transition ON CLUSTER clickhouse_cluster_replicated
(
    timestamp           DateTime COMMENT '타임스탬프',
    host_id             String COMMENT 'HOST ID: 10자리 HOST 구분 ID',
    mac_addr            String COMMENT 'STB MAC: 12자리 표준 MAC Address',
    cm_mac              String COMMENT 'CM MAC: 12자리 표준 MAC Address',
    stb_ip              String COMMENT 'STB IP',
    cm_ip               String COMMENT 'CM IP',
    stb_model           String COMMENT 'STB 모델명',
    mw_ver              String COMMENT '서비스 버전: STB 미들웨어 버전',
    local_ver           String COMMENT 'Local UI 버전',
    cloud_ver           String COMMENT 'Cloud UI 버전',
    logging_time        String COMMENT '정보 수집 시간',
    logging_time_utc    Nullable(DateTime) COMMENT '정보 수집 시간 (UTC), 원본에 없는 컬럼',
    ch_sid              String COMMENT '채널 sid',
    ch_num              String COMMENT '채널 number',
    ch_name             Nullable(String) COMMENT '채널명',
    ch_qam_freq         String COMMENT 'QAM frequency (v2.5: qamChFreq → chQamFreq)',
    ch_qam_mode         String COMMENT 'QAM modulator (v2.5: qamChMode → chQamMode)',
    -- v2.2: ch_qam_pwr_lvl, ch_qam_snr 삭제됨
    ch_8vsb_freq        String COMMENT '8VSB frequency (v2.5: vsbChFreq → ch8vsbFreq)',
    ch_8vsb_mode        String COMMENT '8VSB modulator (v2.5: vsbChMode → ch8vsbMode)',
    ch_8vsb_pwr_lvl     String COMMENT '8VSB powerLevel (v2.5: vsbChPwrLvl → ch8vsbPwrLvl)',
    ch_8vsb_snr         String COMMENT '8VSB snr (v2.5: vsbChSnr → ch8vsbSnr)'
)
ENGINE = ReplicatedMergeTree('/clickhouse/tables/{shard}/catv/stb_transmission_network_quality_transition', '{replica}')
ORDER BY (timestamp, host_id)
TTL timestamp + toIntervalYear(3);
-- +goose StatementEnd

-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS catv.dist_stb_transmission_network_quality_transition ON CLUSTER clickhouse_cluster_replicated
AS catv.stb_transmission_network_quality_transition
ENGINE = Distributed(clickhouse_cluster_replicated, catv, stb_transmission_network_quality_transition, sipHash64(host_id));
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS catv.stb_transmission_raw ON CLUSTER clickhouse_cluster_replicated SYNC;
-- +goose StatementEnd

-- +goose StatementBegin
DROP TABLE IF EXISTS catv.dist_stb_transmission_raw ON CLUSTER clickhouse_cluster_replicated SYNC;
-- +goose StatementEnd

-- +goose StatementBegin
DROP TABLE IF EXISTS catv.stb_transmission_periodic ON CLUSTER clickhouse_cluster_replicated SYNC;
-- +goose StatementEnd

-- +goose StatementBegin
DROP TABLE IF EXISTS catv.dist_stb_transmission_periodic ON CLUSTER clickhouse_cluster_replicated SYNC;
-- +goose StatementEnd

-- +goose StatementBegin
DROP TABLE IF EXISTS catv.stb_transmission_daily ON CLUSTER clickhouse_cluster_replicated SYNC;
-- +goose StatementEnd

-- +goose StatementBegin
DROP TABLE IF EXISTS catv.dist_stb_transmission_daily ON CLUSTER clickhouse_cluster_replicated SYNC;
-- +goose StatementEnd

-- +goose StatementBegin
DROP TABLE IF EXISTS catv.stb_transmission_quality_measurement ON CLUSTER clickhouse_cluster_replicated SYNC;
-- +goose StatementEnd

-- +goose StatementBegin
DROP TABLE IF EXISTS catv.dist_stb_transmission_quality_measurement ON CLUSTER clickhouse_cluster_replicated SYNC;
-- +goose StatementEnd

-- +goose StatementBegin
DROP TABLE IF EXISTS catv.stb_transmission_diagnostic ON CLUSTER clickhouse_cluster_replicated SYNC;
-- +goose StatementEnd

-- +goose StatementBegin
DROP TABLE IF EXISTS catv.dist_stb_transmission_diagnostic ON CLUSTER clickhouse_cluster_replicated SYNC;
-- +goose StatementEnd

-- +goose StatementBegin
DROP TABLE IF EXISTS catv.stb_transmission_network_quality_transition ON CLUSTER clickhouse_cluster_replicated SYNC;
-- +goose StatementEnd

-- +goose StatementBegin
DROP TABLE IF EXISTS catv.dist_stb_transmission_network_quality_transition ON CLUSTER clickhouse_cluster_replicated SYNC;
-- +goose StatementEnd
