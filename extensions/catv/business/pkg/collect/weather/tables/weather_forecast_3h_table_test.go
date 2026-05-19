package tables

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"ntels.com/pharos/core/external/orm"
)

func TestWeatherForecast3hTable_TableName(t *testing.T) {
	table := WeatherForecast3hTable{}
	assert.Equal(t, "dist_weather_forecast_3h", table.TableName())
}

func TestWeatherForecast3hTable_Structure(t *testing.T) {
	now := time.Now()
	sunrise := time.Date(2026, 3, 5, 6, 30, 0, 0, time.UTC)
	sunset := time.Date(2026, 3, 5, 18, 30, 0, 0, time.UTC)

	table := WeatherForecast3hTable{
		ForecastHour: orm.Datetime{Time: now},
		Location:     "서울",
		WeatherIcon:  5,
		WeatherText:  "맑음",
		Temperature:  15,
		RainProb:     10,
		Sunrise:      orm.Datetime{Time: sunrise},
		Sunset:       orm.Datetime{Time: sunset},
	}

	assert.Equal(t, now, table.ForecastHour.Time)
	assert.Equal(t, "서울", table.Location)
	assert.Equal(t, int32(5), table.WeatherIcon)
	assert.Equal(t, "맑음", table.WeatherText)
	assert.Equal(t, int32(15), table.Temperature)
	assert.Equal(t, int32(10), table.RainProb)
	assert.Equal(t, sunrise, table.Sunrise.Time)
	assert.Equal(t, sunset, table.Sunset.Time)
}
