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

type WeatherObsTable struct {
	InsertTime orm.Datetime `json:"insert_time" db:"insert_time"`

	ObsDate       orm.Datetime `json:"obs_date" db:"obs_date"`
	Source        string       `json:"source" db:"source"`
	Location      string       `json:"location" db:"location"`
	Temperature   float32      `json:"temperature" db:"temperature"`
	WeatherText   string       `json:"weather_text" db:"weather_text"`
	WeatherIcon   int32        `json:"weather_icon" db:"weather_icon"`
	WindSpeed     float32      `json:"wind_speed" db:"wind_speed"`
	WindDirection string       `json:"wind_direction" db:"wind_direction"`
	Rainfall      *float32     `json:"rainfall" db:"rainfall"`
	Humidity      int32        `json:"humidity" db:"humidity"`
}

func (w WeatherObsTable) TableName() string {
	return "dist_weather_obs"
}

func (w WeatherObsTable) Insert(config common.Config) error {
	var query = "INSERT INTO " + w.TableName() +
		" (insert_time, obs_date, source, location, temperature, weather_text, weather_icon, wind_speed, wind_direction, rainfall, humidity)" +
		" VALUES (:insert_time, :obs_date, :source, :location, :temperature, :weather_text, :weather_icon, :wind_speed, :wind_direction, :rainfall, :humidity)"

	handler := func(db *sqlx.DB) error {
		_, err := db.NamedExec(query, w)
		return err
	}

	return orm.StatisticsHandler(config.Catv.Database.Driver, &config.Catv.Database, handler)
}

func (w WeatherObsTable) Transform(insertTime orm.Datetime, data string) ([]WeatherTable, error) {
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

		// # 구분자로 분리: 날짜#발표시각#지역#기온#날씨텍스트#날씨아이콘#풍속#풍향#강수량#습도#
		fields := strings.Split(line, "#")
		if len(fields) < 10 {
			continue // 필드가 부족하면 스킵
		}

		dateStr := fields[0]
		timeStr := fields[1]
		location := fields[2]
		temperatureStr := fields[3]
		weatherText := fields[4]
		weatherIconStr := fields[5]
		windSpeedStr := fields[6]
		windDirection := fields[7]
		rainfallStr := fields[8]
		humidityStr := fields[9]

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

		// 시각 파싱 (HH:MM)
		timeParts := strings.Split(timeStr, ":")
		if len(timeParts) != 2 {
			continue
		}
		hour, err1 := strconv.Atoi(timeParts[0])
		minute, err2 := strconv.Atoi(timeParts[1])
		if err1 != nil || err2 != nil {
			continue
		}

		// 관측날짜 (날짜 + 발표시각)
		obsDateTime := time.Date(year, time.Month(month), day, hour, minute, 0, 0, timeLocation)

		// 기온 파싱
		temperature, err := strconv.ParseFloat(temperatureStr, 32)
		if err != nil {
			continue
		}

		// 날씨아이콘 파싱
		weatherIcon, err := strconv.ParseInt(weatherIconStr, 10, 32)
		if err != nil {
			continue
		}

		// 풍속 파싱
		windSpeed, err := strconv.ParseFloat(windSpeedStr, 32)
		if err != nil {
			continue
		}

		// 강수량 파싱 ('-'는 NULL 처리)
		var rainfall *float32
		if rainfallStr != "-" {
			rainfallVal, err := strconv.ParseFloat(rainfallStr, 32)
			if err == nil {
				rainfallFloat32 := float32(rainfallVal)
				rainfall = &rainfallFloat32
			}
		}

		// 습도 파싱
		humidity, err := strconv.ParseInt(humidityStr, 10, 32)
		if err != nil {
			continue
		}

		// WeatherObsTable 구조체 생성
		table := WeatherObsTable{
			InsertTime: insertTime,

			ObsDate:       orm.Datetime{Time: obsDateTime.UTC()},
			Source:        w.Source,
			Location:      location,
			Temperature:   float32(temperature),
			WeatherText:   weatherText,
			WeatherIcon:   int32(weatherIcon),
			WindSpeed:     float32(windSpeed),
			WindDirection: windDirection,
			Rainfall:      rainfall,
			Humidity:      int32(humidity),
		}

		results = append(results, table)
	}

	return results, nil
}

func (w WeatherObsTable) ConvertBatchFormat() (map[string]any, error) {
	insertTime, err := w.InsertTime.Value()
	if err != nil {
		return nil, fmt.Errorf("failed to convert insert time: %w", err)
	}

	obsDate, err := w.ObsDate.Value()
	if err != nil {
		return nil, fmt.Errorf("failed to convert obs date: %w", err)
	}

	var rainfall any
	if w.Rainfall != nil {
		rainfall = *w.Rainfall
	} else {
		rainfall = nil
	}

	return map[string]any{
		"insert_time": insertTime,

		"obs_date":       obsDate,
		"source":         w.Source,
		"location":       w.Location,
		"temperature":    w.Temperature,
		"weather_text":   w.WeatherText,
		"weather_icon":   w.WeatherIcon,
		"wind_speed":     w.WindSpeed,
		"wind_direction": w.WindDirection,
		"rainfall":       rainfall,
		"humidity":       w.Humidity,
	}, nil
}

func (w WeatherObsTable) GetMockData(t time.Time) (string, error) {
	var data strings.Builder
	data.WriteString("날짜#발표시각#지역#기온#날씨텍스트#날씨아이콘#풍속#풍향#강수량#습도#" + CRLF)

	var getMockData = func(location string) string {
		var weatherTexts = []string{"맑음", "흐림", "구름많음", "구름조금"}
		var windDirections = []string{
			"서", "북북동", "남", "북동", "남남서", "북서", "남동", "북북서", "동북동", "남서",
			"동", "북", "서남서", "서북서", "남남동", "동남동", "북동북", "남서남", "서서남", "동동남",
			"북서북", "남동남", "서서북", "북북동",
		}

		// 관측 시각 (현재 시각)
		obsTime := time.Now().Truncate(60 * time.Minute)

		// 날씨 선택
		weatherText := weatherTexts[rand.IntN(len(weatherTexts))]

		// 현실적인 온도 생성 (지역/월/시간 고려)
		temperature := mock.GetRealisticTemperature(location, obsTime)

		// 현실적인 풍속 생성 (지역 타입/월 고려)
		windSpeed := mock.GetRealisticWindSpeed(location, obsTime.Month())

		// 현실적인 습도 생성 (지역/월/날씨 고려)
		humidity := mock.GetRealisticHumidity(location, obsTime.Month(), weatherText)

		// 현실적인 강수량 생성 (날씨/월 고려)
		rainfall := mock.GetRealisticRainfall(weatherText, obsTime.Month())

		// 날씨에 맞는 아이콘
		weatherIcon := mock.GetWeatherIconByText(weatherText)

		return fmt.Sprintf("%s#%s#%s#%.1f#%s#%d#%.1f#%s#%s#%d",
			obsTime.Format("2006/01/02"),
			obsTime.Format("15:04"),
			location,
			temperature,
			weatherText,
			weatherIcon,
			windSpeed,
			windDirections[rand.IntN(len(windDirections))],
			func() string {
				switch rainfall {
				case nil:
					return "-"
				default:
					return fmt.Sprintf("%.1f", *rainfall)
				}
			}(),
			humidity,
		)
	}

	for _, location := range []string{
		"속초", "철원", "동두천", "파주", "대관령", "춘천", "백령도", "북강릉", "북춘천", "서울",
		"인천", "원주", "울릉도", "수원", "영월", "충주", "서산", "울진", "청주", "대전",
		"추풍령", "안동", "상주", "포항", "군산", "대구", "전주", "울산", "창원", "광주",
		"부산", "통영", "목포", "여수", "흑산도", "완도", "고흥", "진도", "강릉", "거제",
		"남해", "거창", "동해", "양평", "이천", "인제", "홍천", "태백", "정선군", "제천",
		"보은", "천안", "보령", "부여", "금산", "세종", "부안", "임실", "정읍", "남원",
		"장수", "고창군", "영광군", "김해시", "순창군", "양산시", "보성군", "강진군", "장흥군", "해남군",
		"고흥군", "의령군", "함안군", "창녕군", "고성군", "남해군", "하동군", "산청군", "함양군", "거창군",
		"합천군", "밀양시", "산청군", "의성군", "구미시", "영천시",
	} {
		data.WriteString(getMockData(location) + CRLF)
	}

	result, err := UTF8ToEUC_KR(data.String())
	if err != nil {
		return "", fmt.Errorf("failed to convert data encoding to EUC-KR: %w", err)
	}

	return string(result), nil
}
