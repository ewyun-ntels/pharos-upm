-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS catv.weather_raw ON CLUSTER clickhouse_cluster_replicated
(
    `insert_time` DateTime DEFAULT now(),
    `file` String COMMENT '원본 파일명',
    `raw` String COMMENT '원본 데이터 전체 텍스트',
    `errors` Array(String) COMMENT '파싱 중 발생한 오류 메시지 목록'
)
ENGINE = ReplicatedReplacingMergeTree('/clickhouse/tables/{shard}/catv/weather_raw', '{replica}')
PARTITION BY toYYYYMM(insert_time)
ORDER BY (insert_time, file, raw)
TTL insert_time + toIntervalYear(3);
-- +goose StatementEnd

-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS catv.dist_weather_raw ON CLUSTER clickhouse_cluster_replicated
AS catv.weather_raw
ENGINE = Distributed(clickhouse_cluster_replicated, catv, weather_raw, sipHash64(file));
-- +goose StatementEnd

-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS catv.so_weather_mapping ON CLUSTER clickhouse_cluster_replicated
(
    `so_nm` String COMMENT '가입자SO명 (SCRBR_SO_NM)',
    `weather_location` String COMMENT '날씨 지역명',
    `province` String COMMENT '광역시/도명 (선택적 분석용)'
)
ENGINE = ReplicatedReplacingMergeTree('/clickhouse/tables/{shard}/catv/so_weather_mapping', '{replica}')
ORDER BY (so_nm)
-- +goose StatementEnd

-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS catv.dist_so_weather_mapping ON CLUSTER clickhouse_cluster_replicated
AS catv.so_weather_mapping
ENGINE = Distributed(clickhouse_cluster_replicated, catv, so_weather_mapping, sipHash64(so_nm));
-- +goose StatementEnd

-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS catv.weather_forecast_3h ON CLUSTER clickhouse_cluster_replicated
(
    `insert_time` DateTime DEFAULT now() COMMENT '원본에 없는 컬럼',

    `forecast_hour` DateTime COMMENT '예보시간',
    `location` String COMMENT '지역명',
    `weather_icon` Int32 COMMENT '날씨아이콘',
    `weather_text` String COMMENT '날씨텍스트',
    `temperature` Int32 COMMENT '기온',
    `rain_prob` Int32 COMMENT '강수확률',
    `sunrise` DateTime COMMENT '일출시각',
    `sunset` DateTime COMMENT '일몰시각'
)
ENGINE = ReplicatedReplacingMergeTree('/clickhouse/tables/{shard}/catv/weather_forecast_3h', '{replica}')
PARTITION BY toYYYYMM(forecast_hour)
ORDER BY (forecast_hour, location)
TTL forecast_hour + toIntervalYear(3);
-- +goose StatementEnd

-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS catv.dist_weather_forecast_3h ON CLUSTER clickhouse_cluster_replicated
AS catv.weather_forecast_3h
ENGINE = Distributed(clickhouse_cluster_replicated, catv, weather_forecast_3h, sipHash64(location));
-- +goose StatementEnd

-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS catv.weather_forecast_daily ON CLUSTER clickhouse_cluster_replicated
(
    `insert_time` DateTime DEFAULT now() COMMENT '원본에 없는 컬럼',

    `forecast_date` DateTime COMMENT '예보날짜',
    `location` String COMMENT '지역명',
    `temp_min` Int32 COMMENT '최저기온',
    `temp_max` Int32 COMMENT '최고기온',
    `rain_prob_am` Int32 COMMENT '오전강수확률',
    `rain_prob_pm` Int32 COMMENT '오후강수확률',
    `weather_text` String COMMENT '날씨텍스트',
    `weather_icon` Int32 COMMENT '날씨아이콘',
    `wind_direction` String COMMENT '풍향'
)
ENGINE = ReplicatedReplacingMergeTree('/clickhouse/tables/{shard}/catv/weather_forecast_daily', '{replica}')
PARTITION BY toYYYYMM(forecast_date)
ORDER BY (forecast_date, location)
TTL forecast_date + toIntervalYear(3);
-- +goose StatementEnd

-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS catv.dist_weather_forecast_daily ON CLUSTER clickhouse_cluster_replicated
AS catv.weather_forecast_daily
ENGINE = Distributed(clickhouse_cluster_replicated, catv, weather_forecast_daily, sipHash64(location));
-- +goose StatementEnd

-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS catv.weather_obs ON CLUSTER clickhouse_cluster_replicated
(
    `insert_time` DateTime DEFAULT now() COMMENT '원본에 없는 컬럼',

    `obs_date` DateTime COMMENT '관측날짜 (날짜 + 발표시각)',
    `source` String COMMENT '데이터 출처 (SHKO, AWS)',
    `location` String COMMENT '지역',
    `temperature` Float32 COMMENT '기온',
    `weather_text` String COMMENT '날씨텍스트',
    `weather_icon` Int32 COMMENT '날씨아이콘',
    `wind_speed` Float32 COMMENT '풍속',
    `wind_direction` String COMMENT '풍향',
    `rainfall` Nullable(Float32) COMMENT '강수량',
    `humidity` Int32 COMMENT '습도'
)
ENGINE = ReplicatedReplacingMergeTree('/clickhouse/tables/{shard}/catv/weather_obs', '{replica}')
PARTITION BY toYYYYMM(obs_date)
ORDER BY (obs_date, source, location)
TTL obs_date + toIntervalYear(3);
-- +goose StatementEnd

-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS catv.dist_weather_obs ON CLUSTER clickhouse_cluster_replicated
AS catv.weather_obs
ENGINE = Distributed(clickhouse_cluster_replicated, catv, weather_obs, sipHash64(location));
-- +goose StatementEnd

-- +goose StatementBegin
CREATE VIEW IF NOT EXISTS catv.v_weather_unified  ON CLUSTER clickhouse_cluster_replicated
AS
SELECT
    datetime,
    location,
    -- obs_* (shko.dat/shko2.dat - 실시간 관측)
    obs_source,
    max(obs_temperature) AS obs_temperature,
    max(obs_weather_text) AS obs_weather_text,
    max(obs_weather_icon) AS obs_weather_icon,
    max(obs_wind_speed) AS obs_wind_speed,
    max(obs_wind_direction) AS obs_wind_direction,
    max(obs_rainfall) AS obs_rainfall,
    max(obs_humidity) AS obs_humidity,
    -- forecast_3h_* (3hour.dat - 3시간 예보)
    max(forecast_3h_weather_icon) AS forecast_3h_weather_icon,
    max(forecast_3h_weather_text) AS forecast_3h_weather_text,
    max(forecast_3h_temperature) AS forecast_3h_temperature,
    max(forecast_3h_rain_prob) AS forecast_3h_rain_prob,
    max(forecast_3h_sunrise) AS forecast_3h_sunrise,
    max(forecast_3h_sunset) AS forecast_3h_sunset,
    -- forecast_daily_* (land.dat - 일별 예보)
    max(forecast_daily_temp_min) AS forecast_daily_temp_min,
    max(forecast_daily_temp_max) AS forecast_daily_temp_max,
    max(forecast_daily_rain_prob_am) AS forecast_daily_rain_prob_am,
    max(forecast_daily_rain_prob_pm) AS forecast_daily_rain_prob_pm,
    max(forecast_daily_weather_text) AS forecast_daily_weather_text,
    max(forecast_daily_weather_icon) AS forecast_daily_weather_icon,
    max(forecast_daily_wind_direction) AS forecast_daily_wind_direction
FROM (
    SELECT
        obs_date AS datetime,
        location,
        source AS obs_source,
        temperature AS obs_temperature,
        weather_text AS obs_weather_text,
        weather_icon AS obs_weather_icon,
        wind_speed AS obs_wind_speed,
        wind_direction AS obs_wind_direction,
        rainfall AS obs_rainfall,
        humidity AS obs_humidity,
        NULL AS forecast_3h_weather_icon,
        NULL AS forecast_3h_weather_text,
        NULL AS forecast_3h_temperature,
        NULL AS forecast_3h_rain_prob,
        NULL AS forecast_3h_sunrise,
        NULL AS forecast_3h_sunset,
        NULL AS forecast_daily_temp_min,
        NULL AS forecast_daily_temp_max,
        NULL AS forecast_daily_rain_prob_am,
        NULL AS forecast_daily_rain_prob_pm,
        NULL AS forecast_daily_weather_text,
        NULL AS forecast_daily_weather_icon,
        NULL AS forecast_daily_wind_direction
    FROM catv.dist_weather_obs FINAL
    
    UNION ALL
    
    SELECT
        forecast_hour AS datetime,
        location,
        NULL AS obs_source,
        NULL AS obs_temperature,
        NULL AS obs_weather_text,
        NULL AS obs_weather_icon,
        NULL AS obs_wind_speed,
        NULL AS obs_wind_direction,
        NULL AS obs_rainfall,
        NULL AS obs_humidity,
        weather_icon AS forecast_3h_weather_icon,
        weather_text AS forecast_3h_weather_text,
        temperature AS forecast_3h_temperature,
        rain_prob AS forecast_3h_rain_prob,
        sunrise AS forecast_3h_sunrise,
        sunset AS forecast_3h_sunset,
        NULL AS forecast_daily_temp_min,
        NULL AS forecast_daily_temp_max,
        NULL AS forecast_daily_rain_prob_am,
        NULL AS forecast_daily_rain_prob_pm,
        NULL AS forecast_daily_weather_text,
        NULL AS forecast_daily_weather_icon,
        NULL AS forecast_daily_wind_direction
    FROM catv.dist_weather_forecast_3h FINAL
    
    UNION ALL
    
    SELECT
        forecast_date AS datetime,
        location,
        NULL AS obs_source,
        NULL AS obs_temperature,
        NULL AS obs_weather_text,
        NULL AS obs_weather_icon,
        NULL AS obs_wind_speed,
        NULL AS obs_wind_direction,
        NULL AS obs_rainfall,
        NULL AS obs_humidity,
        NULL AS forecast_3h_weather_icon,
        NULL AS forecast_3h_weather_text,
        NULL AS forecast_3h_temperature,
        NULL AS forecast_3h_rain_prob,
        NULL AS forecast_3h_sunrise,
        NULL AS forecast_3h_sunset,
        temp_min AS forecast_daily_temp_min,
        temp_max AS forecast_daily_temp_max,
        rain_prob_am AS forecast_daily_rain_prob_am,
        rain_prob_pm AS forecast_daily_rain_prob_pm,
        weather_text AS forecast_daily_weather_text,
        weather_icon AS forecast_daily_weather_icon,
        wind_direction AS forecast_daily_wind_direction
    FROM catv.dist_weather_forecast_daily FINAL
)
GROUP BY datetime, location, obs_source;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS catv.weather_raw ON CLUSTER clickhouse_cluster_replicated SYNC;
-- +goose StatementEnd

-- +goose StatementBegin
DROP TABLE IF EXISTS catv.dist_weather_raw ON CLUSTER clickhouse_cluster_replicated SYNC;
-- +goose StatementEnd

-- +goose StatementBegin
DROP TABLE IF EXISTS catv.so_weather_mapping ON CLUSTER clickhouse_cluster_replicated SYNC;
-- +goose StatementEnd

-- +goose StatementBegin
DROP TABLE IF EXISTS catv.dist_so_weather_mapping ON CLUSTER clickhouse_cluster_replicated SYNC;
-- +goose StatementEnd

-- +goose StatementBegin
DROP TABLE IF EXISTS catv.weather_forecast_3h ON CLUSTER clickhouse_cluster_replicated SYNC;
-- +goose StatementEnd

-- +goose StatementBegin
DROP TABLE IF EXISTS catv.dist_weather_forecast_3h ON CLUSTER clickhouse_cluster_replicated SYNC;
-- +goose StatementEnd

-- +goose StatementBegin
DROP TABLE IF EXISTS catv.weather_forecast_daily ON CLUSTER clickhouse_cluster_replicated SYNC;
-- +goose StatementEnd

-- +goose StatementBegin
DROP TABLE IF EXISTS catv.dist_weather_forecast_daily ON CLUSTER clickhouse_cluster_replicated SYNC;
-- +goose StatementEnd

-- +goose StatementBegin
DROP TABLE IF EXISTS catv.weather_obs ON CLUSTER clickhouse_cluster_replicated SYNC;
-- +goose StatementEnd

-- +goose StatementBegin
DROP TABLE IF EXISTS catv.dist_weather_obs ON CLUSTER clickhouse_cluster_replicated SYNC;
-- +goose StatementEnd

-- +goose StatementBegin
DROP TABLE IF EXISTS catv.v_weather_unified ON CLUSTER clickhouse_cluster_replicated SYNC;
-- +goose StatementEnd