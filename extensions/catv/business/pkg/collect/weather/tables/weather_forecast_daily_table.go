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

type WeatherForecastDailyTable struct {
	InsertTime orm.Datetime `json:"insert_time" db:"insert_time"`

	ForecastDate  orm.Datetime `json:"forecast_date" db:"forecast_date"`
	Location      string       `json:"location" db:"location"`
	TempMin       int32        `json:"temp_min" db:"temp_min"`
	TempMax       int32        `json:"temp_max" db:"temp_max"`
	RainProbAm    int32        `json:"rain_prob_am" db:"rain_prob_am"`
	RainProbPm    int32        `json:"rain_prob_pm" db:"rain_prob_pm"`
	WeatherText   string       `json:"weather_text" db:"weather_text"`
	WeatherIcon   int32        `json:"weather_icon" db:"weather_icon"`
	WindDirection string       `json:"wind_direction" db:"wind_direction"`
}

func (w WeatherForecastDailyTable) TableName() string {
	return "dist_weather_forecast_daily"
}

func (w WeatherForecastDailyTable) Insert(config common.Config) error {
	var query = "INSERT INTO " + w.TableName() +
		" (insert_time, forecast_date, location, temp_min, temp_max, rain_prob_am, rain_prob_pm, weather_text, weather_icon, wind_direction)" +
		" VALUES (:insert_time, :forecast_date, :location, :temp_min, :temp_max, :rain_prob_am, :rain_prob_pm, :weather_text, :weather_icon, :wind_direction)"

	handler := func(db *sqlx.DB) error {
		_, err := db.NamedExec(query, w)
		return err
	}

	return orm.StatisticsHandler(config.Catv.Database.Driver, &config.Catv.Database, handler)
}

func (w WeatherForecastDailyTable) Transform(insertTime orm.Datetime, data string) ([]WeatherTable, error) {
	timeLocation, err := time.LoadLocation("Asia/Seoul")
	if err != nil {
		return nil, fmt.Errorf("failed to load location: %w", err)
	}

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

		// # 구분자로 분리: 날짜#지역명#최저기온#최고기온#오전강수확률#오후강수확률#날씨텍스트#날씨아이콘#풍향#
		fields := strings.Split(line, "#")
		if len(fields) < 9 {
			continue // 필드가 부족하면 스킵
		}

		dateStr := fields[0]
		location := fields[1]
		tempMinStr := fields[2]
		tempMaxStr := fields[3]
		rainProbAmStr := fields[4]
		rainProbPmStr := fields[5]
		weatherText := fields[6]
		weatherIconStr := fields[7]
		windDirection := fields[8]

		// 날짜 파싱 (YYYY/MM/DD)
		dateParts := strings.Split(dateStr, "/")
		if len(dateParts) != 3 {
			continue
		}
		year, err1 := strconv.Atoi(dateParts[0])
		month, err2 := strconv.Atoi(dateParts[1])
		day, err3 := strconv.Atoi(dateParts[2])
		if err1 != nil || err2 != nil || err3 != nil {
			continue
		}

		// 최저기온 파싱 ('-'는 스킵)
		var tempMin int32
		if tempMinStr != "-" {
			tempMinVal, err := strconv.ParseInt(tempMinStr, 10, 32)
			if err != nil {
				continue
			}
			tempMin = int32(tempMinVal)
		}

		// 최고기온 파싱 ('-'는 스킵)
		var tempMax int32
		if tempMaxStr != "-" {
			tempMaxVal, err := strconv.ParseInt(tempMaxStr, 10, 32)
			if err != nil {
				continue
			}
			tempMax = int32(tempMaxVal)
		}

		// 오전 강수확률 파싱 ('-'는 0으로 처리)
		var rainProbAm int32
		if rainProbAmStr != "-" {
			rainProbAmVal, err := strconv.ParseInt(rainProbAmStr, 10, 32)
			if err != nil {
				continue
			}
			rainProbAm = int32(rainProbAmVal)
		}

		// 오후 강수확률 파싱 ('-'는 0으로 처리)
		var rainProbPm int32
		if rainProbPmStr != "-" {
			rainProbPmVal, err := strconv.ParseInt(rainProbPmStr, 10, 32)
			if err != nil {
				continue
			}
			rainProbPm = int32(rainProbPmVal)
		}

		// 날씨아이콘 파싱 ('-'는 스킵)
		var weatherIcon int32
		if weatherIconStr != "-" {
			weatherIconVal, err := strconv.ParseInt(weatherIconStr, 10, 32)
			if err != nil {
				continue
			}
			weatherIcon = int32(weatherIconVal)
		}

		// 일별 데이터를 1시간 단위로 확장 (00:00부터 23:00까지 24개)
		for hour := range 24 {
			forecastDateTime := time.Date(year, time.Month(month), day, hour, 0, 0, 0, timeLocation)

			table := WeatherForecastDailyTable{
				InsertTime: insertTime,

				ForecastDate:  orm.Datetime{Time: forecastDateTime.UTC()},
				Location:      location,
				TempMin:       tempMin,
				TempMax:       tempMax,
				RainProbAm:    rainProbAm,
				RainProbPm:    rainProbPm,
				WeatherText:   weatherText,
				WeatherIcon:   weatherIcon,
				WindDirection: windDirection,
			}

			results = append(results, table)
		}
	}

	return results, nil
}

func (w WeatherForecastDailyTable) ConvertBatchFormat() (map[string]any, error) {
	insertTime, err := w.InsertTime.Value()
	if err != nil {
		return nil, fmt.Errorf("failed to convert insert time: %w", err)
	}

	forecastDate, err := w.ForecastDate.Value()
	if err != nil {
		return nil, fmt.Errorf("failed to convert forecast date: %w", err)
	}

	return map[string]any{
		"insert_time": insertTime,

		"forecast_date":  forecastDate,
		"location":       w.Location,
		"temp_min":       w.TempMin,
		"temp_max":       w.TempMax,
		"rain_prob_am":   w.RainProbAm,
		"rain_prob_pm":   w.RainProbPm,
		"weather_text":   w.WeatherText,
		"weather_icon":   w.WeatherIcon,
		"wind_direction": w.WindDirection,
	}, nil
}

func (w WeatherForecastDailyTable) GetMockData(t time.Time) (string, error) {
	var data strings.Builder
	data.WriteString("날짜#지역명#최저기온#최고기온#오전강수확률#오후강수확률#날씨텍스트#날씨아이콘#풍향#" + CRLF)

	var getMockData = func(location string) string {
		var weatherTexts = []string{"맑음", "흐림", "구름많음", "구름많고 눈", "눈/비", "비/눈", "흐리고 비", "흐리고 눈", "구름많고 비", "흐린후갬", "맑다가 흐림"}
		var windDirections = []string{"-", "남동-남", "북서-북", "동-남동", "서-북서", "남서-남", "북-북동", "남-남서", "서북-서", "북동-동", "동북-북", "남남-남동"}

		// 날씨 선택
		weatherText := weatherTexts[rand.IntN(len(weatherTexts))]

		// 현실적인 일 최저/최고 온도 생성 (지역/월 고려)
		tempMin, tempMax := mock.GetDailyTempRange(location, t.Month())

		// 현실적인 강수확률 생성 (오전/오후, 월/날씨 고려)
		rainProbAm := mock.GetRealisticRainProb(t.Month(), weatherText)
		rainProbPm := mock.GetRealisticRainProb(t.Month(), weatherText)
		// 오후 강수확률이 일반적으로 더 높음
		if rainProbPm < rainProbAm {
			rainProbPm = min(rainProbAm+rand.Int32N(10), 100)
		}

		// 날씨에 맞는 아이콘
		weatherIcon := mock.GetWeatherIconByText(weatherText)

		//2026/01/14#백령도#-5#5#20#30#흐림#4#남동-남#
		return fmt.Sprintf("%s#%s#%d#%d#%d#%d#%s#%d#%s",
			t.Format("2006/01/02"),
			location,
			tempMin,
			tempMax,
			rainProbAm,
			rainProbPm,
			weatherText,
			weatherIcon,
			windDirections[rand.IntN(len(windDirections))],
		)
	}

	timeLocation, err := time.LoadLocation("Asia/Seoul")
	if err != nil {
		return "", fmt.Errorf("failed to load location: %w", err)
	}
	t = t.In(timeLocation)

	for _, location := range []string{
		"백령도", "속초", "청천", "강화", "동두천", "인천", "수원", "파주", "이천", "오산",
		"평택", "제천", "보은", "천안", "홍성", "서산", "태안", "문경", "안동", "상주",
		"영주", "울진", "청주", "대전", "전주", "추풍령", "포항", "경주", "대구", "울산",
		"창원", "광주", "목포", "여수", "흑산도", "완도", "고흥", "강릉", "동해", "삼척",
		"태백", "정선", "북강릉", "북춘천", "양평", "춘천", "철원", "대관령", "인제", "홍천",
		"원주", "제주", "고산", "성산", "서귀포", "진주", "통영", "사천", "거제", "남해",
		"북창원", "양산", "밀양", "산청", "거창", "합천",
	} {
		data.WriteString(getMockData(location) + CRLF)
	}

	result, err := UTF8ToEUC_KR(data.String())
	if err != nil {
		return "", fmt.Errorf("failed to convert data encoding to EUC-KR: %w", err)
	}

	return string(result), nil
}
