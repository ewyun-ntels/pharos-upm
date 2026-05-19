package collect

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"ntels.com/pharos/core/pkg/common"
)

func TestLoad_ReturnsNoErrorWithValidConfig(t *testing.T) {
	config := common.Config{}
	config.Catv.Collect.Weather.Use = true
	config.Catv.Collect.Weather.CronSpec = "0 * * * * *"
	config.Catv.Collect.Weather.Timeout.Connection = "30s"
	config.Catv.Collect.Weather.Timeout.Shut = "10s"

	err := Load(config)
	// weather.Load는 서비스를 Start하므로 cron spec이 유효하면 에러 없음
	assert.NoError(t, err)

	// 정상적으로 Unload도 가능해야 함
	assert.NotPanics(t, func() { Unload() })
}

func TestLoad_ReturnsErrorWithInvalidWeatherTimeout(t *testing.T) {
	config := common.Config{}
	config.Catv.Collect.Weather.Use = true
	config.Catv.Collect.Weather.CronSpec = "0 * * * * *"
	config.Catv.Collect.Weather.Timeout.Connection = "invalid-timeout"
	config.Catv.Collect.Weather.Timeout.Shut = "10s"

	err := Load(config)
	assert.Error(t, err)
}
