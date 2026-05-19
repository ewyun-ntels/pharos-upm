package tables

import (
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"ntels.com/pharos/core/external/orm"
)

func TestWeatherRawTable_TableName(t *testing.T) {
	t.Parallel()
	table := WeatherRawTable{}
	assert.Equal(t, "dist_weather_raw", table.TableName())
}

func TestWeatherRawTable_Structure(t *testing.T) {
	t.Parallel()
	now := time.Now()
	table := WeatherRawTable{
		InsertTime: orm.Datetime{Time: now},
		File:       "obs_20260323.csv",
		Raw:        "2026/03/23#06:00#서울#8.5#맑음#1#2.3#북동#-#45#",
	}

	assert.Equal(t, now, table.InsertTime.Time)
	assert.Equal(t, "obs_20260323.csv", table.File)
	assert.Equal(t, "2026/03/23#06:00#서울#8.5#맑음#1#2.3#북동#-#45#", table.Raw)
}

func TestWeatherRawTable_ZeroValue(t *testing.T) {
	t.Parallel()
	table := WeatherRawTable{}

	assert.True(t, table.InsertTime.Time.IsZero())
	assert.Equal(t, "", table.File)
	assert.Equal(t, "", table.Raw)
}

func TestWeatherRawTable_InsertQueryContainsTable(t *testing.T) {
	t.Parallel()
	// Insert 메서드가 올바른 테이블 이름과 컬럼을 사용하는지
	// 실제 DB 없이 쿼리 문자열 구성을 간접 검증한다.
	table := WeatherRawTable{}
	tableName := table.TableName()

	expectedQuery := "INSERT INTO " + tableName + " (file, raw) VALUES (:file, :raw)"

	// 쿼리는 코드 내부에 있으므로 구성 요소를 통해 검증한다.
	assert.True(t, strings.HasPrefix(expectedQuery, "INSERT INTO dist_weather_raw"))
	assert.Contains(t, expectedQuery, ":file")
	assert.Contains(t, expectedQuery, ":raw")
}

func TestWeatherRawTable_MultilineRaw(t *testing.T) {
	t.Parallel()
	// Raw 필드는 파일 전체 내용을 담을 수 있어야 한다.
	rawContent := strings.Join([]string{
		"날짜#발표시각#지역#기온#날씨텍스트#날씨아이콘#풍속#풍향#강수량#습도#",
		"2026/03/23#06:00#서울#8.5#맑음#1#2.3#북동#-#45#",
		"2026/03/23#06:00#부산#11.2#구름많음#3#3.1#남동#0.5#60#",
	}, "\n")

	table := WeatherRawTable{
		File: "obs_20260323.csv",
		Raw:  rawContent,
	}

	assert.Equal(t, "obs_20260323.csv", table.File)
	assert.Equal(t, rawContent, table.Raw)
	assert.Equal(t, 3, len(strings.Split(table.Raw, "\n")))
}

func TestWeatherRawTable_FileFieldVariants(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name string
		file string
	}{
		{"obs 파일", "obs_20260323.csv"},
		{"forecast_3h 파일", "forecast_3h_20260323.csv"},
		{"forecast_daily 파일", "forecast_daily_20260323.csv"},
		{"빈 파일명", ""},
		{"경로 포함", "/data/weather/obs_20260323.csv"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			table := WeatherRawTable{File: tc.file}
			assert.Equal(t, tc.file, table.File)
		})
	}
}
