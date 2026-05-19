package tables

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"ntels.com/pharos/core/external/orm"
)

// TestWeatherTable_InterfaceCompliance는 각 테이블이 WeatherTable 인터페이스를 구현하는지 검증합니다.
func TestWeatherTable_InterfaceCompliance(t *testing.T) {
	// 컴파일 타임 인터페이스 구현 검증
	var _ WeatherTable = (*WeatherForecast3hTable)(nil)
	var _ WeatherTable = (*WeatherForecastDailyTable)(nil)
	var _ WeatherTable = (*WeatherObsTable)(nil)
}

func TestWeatherTable_TableNames(t *testing.T) {
	tests := []struct {
		name     string
		table    WeatherTable
		wantName string
	}{
		{
			name:     "3h forecast table name",
			table:    &WeatherForecast3hTable{},
			wantName: "dist_weather_forecast_3h",
		},
		{
			name:     "daily forecast table name",
			table:    &WeatherForecastDailyTable{},
			wantName: "dist_weather_forecast_daily",
		},
		{
			name:     "obs table name",
			table:    &WeatherObsTable{},
			wantName: "dist_weather_obs",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.wantName, tt.table.TableName())
		})
	}
}

func TestWeatherTable_GetMockData_ReturnsData(t *testing.T) {
	now := time.Now()

	tables := []WeatherTable{
		&WeatherForecast3hTable{},
		&WeatherForecastDailyTable{},
		&WeatherObsTable{},
	}

	for _, table := range tables {
		t.Run(table.TableName(), func(t *testing.T) {
			mockData, err := table.GetMockData(now)
			require.NoError(t, err)
			assert.NotEmpty(t, mockData)
		})
	}
}

func TestWeatherTable_Transform_ReturnsErrorForEmptyData(t *testing.T) {
	tables := []WeatherTable{
		&WeatherForecast3hTable{},
		&WeatherForecastDailyTable{},
		&WeatherObsTable{},
	}

	for _, table := range tables {
		t.Run(table.TableName(), func(t *testing.T) {
			result, err := table.Transform(orm.Datetime{Time: time.Now()}, "")
			assert.Error(t, err)
			assert.Nil(t, result)
		})
	}
}
