-- +goose Up
-- +goose StatementBegin
-- ========================================
-- CATV 품질 측정 함수들 (UDF는 자동 복제되므로 ON CLUSTER 불필요)
-- ========================================

-- 1. 데이터 유효성 검증 (결측치 제외)
CREATE FUNCTION IF NOT EXISTS is_catv_valid AS (pwr, snr) ->
  pwr > -25 AND pwr < 25 AND snr > 15 AND snr < 50;
-- +goose StatementEnd

-- +goose StatementBegin
-- 2. PWR 품질 상태 판단
CREATE FUNCTION IF NOT EXISTS pwr_quality AS (pwr) ->
  multiIf(
    pwr <= -25 OR pwr >= 25, 'Invalid',
    pwr <= -16 OR pwr >= 16, 'Criti',
    (pwr >= -15 AND pwr <= -13) OR (pwr >= 13 AND pwr <= 15), 'Warn',
    pwr >= -12 AND pwr <= 12, 'Good',
    'Invalid'
  );
-- +goose StatementEnd

-- +goose StatementBegin
-- 3. SNR 품질 상태 판단 (256QAM)
CREATE FUNCTION IF NOT EXISTS snr_qam_quality AS (snr) ->
  multiIf(
    snr <= 15 OR snr >= 50, 'Invalid',
    snr <= 31, 'Criti',
    snr = 32, 'Warn',
    snr >= 33, 'Good',
    'Invalid'
  );
-- +goose StatementEnd

-- +goose StatementBegin
-- 4. SNR 품질 상태 판단 (8VSB)
CREATE FUNCTION IF NOT EXISTS snr_8vsb_quality AS (snr) ->
  multiIf(
    snr <= 15 OR snr >= 50, 'Invalid',
    snr <= 25, 'Criti',
    snr = 26, 'Warn',
    snr >= 27, 'Good',
    'Invalid'
  );
-- +goose StatementEnd

-- +goose StatementBegin
-- 5. POPUP 품질 상태 판단
CREATE OR REPLACE FUNCTION popup_quality AS (popup_count) ->
    if(popup_count = 0, 'Good', 'Criti');
-- +goose StatementEnd

-- +goose StatementBegin
-- 6. 종합 품질 상태 판단 (최악 상태 기준)
CREATE OR REPLACE FUNCTION catv_quality_status AS (pwr, qam_snr, vsb_snr, popup_cnt) ->
  multiIf(
    pwr IS NULL, 'Invalid',
    popup_cnt > 0, 'Criti',
    pwr_quality(pwr) = 'Criti', 'Criti',
    isNotNull(qam_snr) AND snr_qam_quality(qam_snr) = 'Criti', 'Criti',
    isNotNull(vsb_snr) AND snr_8vsb_quality(vsb_snr) = 'Criti', 'Criti',
    pwr_quality(pwr) = 'Warn', 'Warn',
    isNotNull(qam_snr) AND snr_qam_quality(qam_snr) = 'Warn', 'Warn',
    isNotNull(vsb_snr) AND snr_8vsb_quality(vsb_snr) = 'Warn', 'Warn',
    'Good'
  );
-- +goose StatementEnd

-- +goose StatementBegin
-- 7. Criti 여부 판단 (boolean)
CREATE FUNCTION IF NOT EXISTS is_catv_criti AS (pwr, snr, freq_meth, popup) ->
  popup_quality(popup) = 'Criti' OR
  pwr_quality(pwr) = 'Criti' OR
  (freq_meth = '256QAM' AND snr_qam_quality(snr) = 'Criti') OR
  (freq_meth = '8VSB' AND snr_8vsb_quality(snr) = 'Criti');
-- +goose StatementEnd

-- +goose StatementBegin
-- 8. Warn 여부 판단 (boolean)
CREATE FUNCTION IF NOT EXISTS is_catv_warn AS (pwr, snr, freq_meth, popup) ->
  NOT is_catv_criti(pwr, snr, freq_meth, popup) AND (
    pwr_quality(pwr) = 'Warn' OR
    (freq_meth = '256QAM' AND snr_qam_quality(snr) = 'Warn') OR
    (freq_meth = '8VSB' AND snr_8vsb_quality(snr) = 'Warn')
  );
-- +goose StatementEnd

-- +goose StatementBegin
-- 9. Good 여부 판단 (boolean)
CREATE FUNCTION IF NOT EXISTS is_catv_good AS (pwr, snr, freq_meth, popup) ->
  NOT is_catv_criti(pwr, snr, freq_meth, popup) AND
  NOT is_catv_warn(pwr, snr, freq_meth, popup) AND
  pwr_quality(pwr) = 'Good' AND
  ((freq_meth = '256QAM' AND snr_qam_quality(snr) = 'Good') OR
   (freq_meth = '8VSB' AND snr_8vsb_quality(snr) = 'Good'));
-- +goose StatementEnd

-- +goose StatementBegin
-- 10. 품질 이상 요약 (GROUP BY용 - 이상 항목을 문자열로 반환)
CREATE FUNCTION IF NOT EXISTS catv_quality_summary AS (pwr, snr_qam, snr_8vsb, popup_cnt) ->
  ifNull(
    nullIf(
      arrayStringConcat(
        arrayFilter(x -> x != '', [
          multiIf(pwr_quality(pwr) = 'Criti', 'PWR_CRITI',
                  pwr_quality(pwr) = 'Warn', 'PWR_WARN', ''),
          multiIf(isNotNull(snr_qam) AND snr_qam_quality(snr_qam) = 'Criti', 'QAM_CRITI',
                  isNotNull(snr_qam) AND snr_qam_quality(snr_qam) = 'Warn', 'QAM_WARN', ''),
          multiIf(isNotNull(snr_8vsb) AND snr_8vsb_quality(snr_8vsb) = 'Criti', '8VSB_CRITI',
                  isNotNull(snr_8vsb) AND snr_8vsb_quality(snr_8vsb) = 'Warn', '8VSB_WARN', ''),
          if(assumeNotNull(popup_cnt) > 0, 'POPUP_CRITI', '')
        ]),
        ','
      ),
      ''
    ),
    'GOOD'
  );
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP FUNCTION IF EXISTS catv_quality_summary;
-- +goose StatementEnd

-- +goose StatementBegin
DROP FUNCTION IF EXISTS is_catv_good;
-- +goose StatementEnd

-- +goose StatementBegin
DROP FUNCTION IF EXISTS is_catv_warn;
-- +goose StatementEnd

-- +goose StatementBegin
DROP FUNCTION IF EXISTS is_catv_criti;
-- +goose StatementEnd

-- +goose StatementBegin
DROP FUNCTION IF EXISTS catv_quality_status;
-- +goose StatementEnd

-- +goose StatementBegin
DROP FUNCTION IF EXISTS popup_quality;
-- +goose StatementEnd

-- +goose StatementBegin
DROP FUNCTION IF EXISTS snr_8vsb_quality;
-- +goose StatementEnd

-- +goose StatementBegin
DROP FUNCTION IF EXISTS snr_qam_quality;
-- +goose StatementEnd

-- +goose StatementBegin
DROP FUNCTION IF EXISTS pwr_quality;
-- +goose StatementEnd

-- +goose StatementBegin
DROP FUNCTION IF EXISTS is_catv_valid;
-- +goose StatementEnd
