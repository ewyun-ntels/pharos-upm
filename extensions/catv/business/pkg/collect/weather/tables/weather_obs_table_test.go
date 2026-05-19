package tables

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"ntels.com/pharos/core/external/orm"
)

func TestWeatherObsTable_TableName(t *testing.T) {
	table := WeatherObsTable{}
	assert.Equal(t, "dist_weather_obs", table.TableName())
}

func TestWeatherObsTable_Structure(t *testing.T) {
	now := time.Now()
	table := WeatherObsTable{
		ObsDate:       orm.Datetime{Time: now},
		Location:      "서울",
		Temperature:   15.5,
		WeatherText:   "맑음",
		WeatherIcon:   1,
		WindSpeed:     2.3,
		WindDirection: "북동",
		Rainfall:      func() *float32 { f := float32(0.0); return &f }(),
		Humidity:      45,
	}

	assert.Equal(t, now, table.ObsDate.Time)
	assert.Equal(t, "서울", table.Location)
	assert.Equal(t, float32(15.5), table.Temperature)
	assert.Equal(t, "맑음", table.WeatherText)
	assert.Equal(t, int32(1), table.WeatherIcon)
	assert.Equal(t, float32(2.3), table.WindSpeed)
	assert.Equal(t, "북동", table.WindDirection)
	assert.Equal(t, float32(0.0), *table.Rainfall)
	assert.Equal(t, int32(45), table.Humidity)
}
