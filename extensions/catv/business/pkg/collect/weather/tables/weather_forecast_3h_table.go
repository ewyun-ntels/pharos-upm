package tables

import (
	"fmt"
	"math/rand/v2"
	"strconv"
	"strings"
	"time"

	"github.com/jmoiron/sqlx"
	"ntels.com/pharos/core/external/orm"
	"ntels.com/pharos/core/pkg/common"
	"ntels.com/pharos/extensions/catv/business/pkg/collect/weather/tables/mock"
)

type WeatherForecast3hTable struct {
	InsertTime orm.Datetime `json:"insert_time" db:"insert_time"`

	ForecastHour orm.Datetime `json:"forecast_hour" db:"forecast_hour"`
	Location     string       `json:"location" db:"location"`
	WeatherIcon  int32        `json:"weather_icon" db:"weather_icon"`
	WeatherText  string       `json:"weather_text" db:"weather_text"`
	Temperature  int32        `json:"temperature" db:"temperature"`
	RainProb     int32        `json:"rain_prob" db:"rain_prob"`
	Sunrise      orm.Datetime `json:"sunrise" db:"sunrise"`
	Sunset       orm.Datetime `json:"sunset" db:"sunset"`
}

func (w WeatherForecast3hTable) TableName() string {
	return "dist_weather_forecast_3h"
}

func (w WeatherForecast3hTable) Insert(config common.Config) error {
	var query = "INSERT INTO " + w.TableName() +
		" (insert_time, forecast_hour, location, weather_icon, weather_text, temperature, rain_prob, sunrise, sunset)" +
		" VALUES (:insert_time, :forecast_hour, :location, :weather_icon, :weather_text, :temperature, :rain_prob, :sunrise, :sunset)"

	handler := func(db *sqlx.DB) error {
		_, err := db.NamedExec(query, w)
		return err
	}

	return orm.StatisticsHandler(config.Catv.Database.Driver, &config.Catv.Database, handler)
}

func (w WeatherForecast3hTable) Transform(insertTime orm.Datetime, data string) ([]WeatherTable, error) {
	timeLocation, err := time.LoadLocation("Asia/Seoul")
	if err != nil {
		return nil, fmt.Errorf("failed to load location: %w", err)
	}
	var now = time.Now().In(timeLocation)

	var results []WeatherTable

	lines := strings.Split(strings.TrimSpace(data), "\n")
	if len(lines) < 2 {
		return nil, fmt.Errorf("invalid data format: expected at least 2 lines (header + data)")
	}

	// 첫 줄은 헤더이므로 스킵
	for i := 1; i < len(lines); i++ {
		line := strings.TrimSpace(lines[i])
		if line == "" {
			continue
		}

		// # 구분자로 분리: 지역명#예보시간#날씨아이콘#날씨텍스트#기온#강수확률#일출시각#일몰시각#
		fields := strings.Split(line, "#")
		if len(fields) < 8 {
			continue // 필드가 부족하면 스킵
		}

		location := fields[0]
		forecastHourStr := fields[1]
		weatherIconStr := fields[2]
		weatherText := fields[3]
		temperatureStr := fields[4]
		rainProbStr := fields[5]
		sunriseStr := fields[6]
		sunsetStr := fields[7]

		// 예보시간 파싱 (HH) -> DateTime (현재 날짜 + HH시)
		forecastHour, err := strconv.Atoi(forecastHourStr)
		if err != nil {
			continue
		}

		// 날씨아이콘 파싱
		weatherIcon, err := strconv.ParseInt(weatherIconStr, 10, 32)
		if err != nil {
			continue
		}

		// 기온 파싱
		temperature, err := strconv.ParseInt(temperatureStr, 10, 32)
		if err != nil {
			continue
		}

		// 강수확률 파싱
		rainProb, err := strconv.ParseInt(rainProbStr, 10, 32)
		if err != nil {
			continue
		}

		// 일출시각 파싱 (HH:MM) -> DateTime
		sunriseParts := strings.Split(sunriseStr, ":")
		var sunrise time.Time
		if len(sunriseParts) == 2 {
			sunriseHour, err1 := strconv.Atoi(sunriseParts[0])
			sunriseMin, err2 := strconv.Atoi(sunriseParts[1])
			if err1 == nil && err2 == nil {
				sunrise = time.Date(now.Year(), now.Month(), now.Day(), sunriseHour, sunriseMin, 0, 0, timeLocation)
			}
		}

		// 일몰시각 파싱 (HH:MM) -> DateTime
		sunsetParts := strings.Split(sunsetStr, ":")
		var sunset time.Time
		if len(sunsetParts) == 2 {
			sunsetHour, err1 := strconv.Atoi(sunsetParts[0])
			sunsetMin, err2 := strconv.Atoi(sunsetParts[1])
			if err1 == nil && err2 == nil {
				sunset = time.Date(now.Year(), now.Month(), now.Day(), sunsetHour, sunsetMin, 0, 0, timeLocation)
			}
		}

		// 3시간 데이터를 1시간 단위로 확장 (예: 09시 → 09, 10, 11시)
		for hourOffset := range 3 {
			forecastDateTime := time.Date(now.Year(), now.Month(), now.Day(), forecastHour+hourOffset, 0, 0, 0, timeLocation)

			table := WeatherForecast3hTable{
				InsertTime: insertTime,

				ForecastHour: orm.Datetime{Time: forecastDateTime.UTC()},
				Location:     location,
				WeatherIcon:  int32(weatherIcon),
				WeatherText:  weatherText,
				Temperature:  int32(temperature),
				RainProb:     int32(rainProb),
				Sunrise:      orm.Datetime{Time: sunrise.UTC()},
				Sunset:       orm.Datetime{Time: sunset.UTC()},
			}

			results = append(results, table)
		}
	}

	return results, nil
}

func (w WeatherForecast3hTable) ConvertBatchFormat() (map[string]any, error) {
	insertTime, err := w.InsertTime.Value()
	if err != nil {
		return nil, fmt.Errorf("failed to convert insert time: %w", err)
	}

	forecastHour, err := w.ForecastHour.Value()
	if err != nil {
		return nil, fmt.Errorf("failed to convert forecast hour: %w", err)
	}

	sunrise, err := w.Sunrise.Value()
	if err != nil {
		return nil, fmt.Errorf("failed to convert sunrise: %w", err)
	}

	sunset, err := w.Sunset.Value()
	if err != nil {
		return nil, fmt.Errorf("failed to convert sunset: %w", err)
	}

	return map[string]any{
		"insert_time": insertTime,

		"forecast_hour": forecastHour,
		"location":      w.Location,
		"weather_icon":  w.WeatherIcon,
		"weather_text":  w.WeatherText,
		"temperature":   w.Temperature,
		"rain_prob":     w.RainProb,
		"sunrise":       sunrise,
		"sunset":        sunset,
	}, nil
}

func (w WeatherForecast3hTable) GetMockData(t time.Time) (string, error) {
	var data strings.Builder
	data.WriteString("지역명#예보시간#날씨아이콘#날씨텍스트#기온#강수확률#일출시각#일몰시각#" + CRLF)

	var getMockData = func(t time.Time, location string) string {
		var timeLocation = t.Location()
		var weatherTexts = []string{"맑음", "흐림", "구름많음", "구름조금"}

		// 예보 시각
		forecastTime := time.Date(t.Year(), t.Month(), t.Day(), t.Hour(), 0, 0, 0, timeLocation)

		// 날씨 선택
		weatherText := weatherTexts[rand.IntN(len(weatherTexts))]

		// 현실적인 온도 생성 (지역/월/시간 고려)
		temperature := int32(mock.GetRealisticTemperature(location, forecastTime))

		// 현실적인 강수확률 생성 (월/날씨 고려)
		rainProb := mock.GetRealisticRainProb(t.Month(), weatherText)

		// 날씨에 맞는 아이콘
		weatherIcon := mock.GetWeatherIconByText(weatherText)

		// 일출/일몰 시각 (월별 차이)
		var sunriseHour, sunsetHour int
		switch t.Month() {
		case time.June, time.July:
			sunriseHour, sunsetHour = 5, 20 // 여름: 일출 빠름, 일몰 늦음
		case time.December, time.January:
			sunriseHour, sunsetHour = 7, 17 // 겨울: 일출 늦음, 일몰 빠름
		default:
			sunriseHour, sunsetHour = 6, 18 // 봄/가을
		}

		return fmt.Sprintf("%s#%02d#%02d#%s#%d#%d#%02d:00#%02d:00#", location, forecastTime.Hour(), weatherIcon, weatherText, temperature, rainProb, sunriseHour, sunsetHour)
	}

	loc, err := time.LoadLocation("Asia/Seoul")
	if err != nil {
		return "", fmt.Errorf("failed to load location: %w", err)
	}
	t = t.In(loc)

	for _, location := range []string{
		"백령도", "속초", "청천", "강화", "동두천", "인천", "수원", "파주", "이천", "오산",
		"평택", "제천", "보은", "천안", "홍성", "서산", "태안", "문경", "안동", "상주",
		"영주", "울진", "청주", "대전", "전주", "추풍령", "포항", "경주", "대구", "울산",
		"창원", "광주", "목포", "여수", "흑산도", "완도", "고흥", "강릉", "동해", "삼척",
		"태백", "정선", "북강릉", "북춘천", "양평", "춘천", "철원", "대관령", "인제", "홍천",
		"원주",
	} {
		data.WriteString(getMockData(t, location) + CRLF)
	}

	result, err := UTF8ToEUC_KR(data.String())
	if err != nil {
		return "", fmt.Errorf("failed to convert data encoding to EUC-KR: %w", err)
	}

	return string(result), nil
}
