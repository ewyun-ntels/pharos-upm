package tables

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"ntels.com/pharos/core/external/orm"
)

func TestWeatherForecastDailyTable_TableName(t *testing.T) {
	table := WeatherForecastDailyTable{}
	assert.Equal(t, "dist_weather_forecast_daily", table.TableName())
}

func TestWeatherForecastDailyTable_Structure(t *testing.T) {
	now := time.Now()
	table := WeatherForecastDailyTable{
		ForecastDate:  orm.Datetime{Time: now},
		Location:      "서울",
		TempMin:       8,
		TempMax:       20,
		RainProbAm:    10,
		RainProbPm:    20,
		WeatherText:   "대체로 맑음",
		WeatherIcon:   3,
		WindDirection: "북동",
	}

	assert.Equal(t, now, table.ForecastDate.Time)
	assert.Equal(t, "서울", table.Location)
	assert.Equal(t, int32(8), table.TempMin)
	assert.Equal(t, int32(20), table.TempMax)
	assert.Equal(t, int32(10), table.RainProbAm)
	assert.Equal(t, int32(20), table.RainProbPm)
	assert.Equal(t, "대체로 맑음", table.WeatherText)
	assert.Equal(t, int32(3), table.WeatherIcon)
	assert.Equal(t, "북동", table.WindDirection)
}
