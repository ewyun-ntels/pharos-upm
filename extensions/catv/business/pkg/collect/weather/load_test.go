package weather

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"ntels.com/pharos/core/pkg/common"
)

func TestLoad_WithValidConfig(t *testing.T) {
	config := common.Config{}
	config.Catv.Collect.Weather.Use = true
	config.Catv.Collect.Weather.CronSpec = "0 * * * * *"
	config.Catv.Collect.Weather.Timeout.Connection = "30s"
	config.Catv.Collect.Weather.Timeout.Shut = "10s"

	err := Load(config)
	assert.NoError(t, err)

	// Unload 호출 후 패닉 없어야 함
	assert.NotPanics(t, func() { Unload() })
}

func TestLoad_WithInvalidConnectionTimeout(t *testing.T) {
	config := common.Config{}
	config.Catv.Collect.Weather.Use = true
	config.Catv.Collect.Weather.CronSpec = "0 * * * * *"
	config.Catv.Collect.Weather.Timeout.Connection = "not-a-duration"
	config.Catv.Collect.Weather.Timeout.Shut = "10s"

	err := Load(config)
	assert.Error(t, err)
}

func TestLoad_WithInvalidCronSpec(t *testing.T) {
	config := common.Config{}
	config.Catv.Collect.Weather.Use = true
	config.Catv.Collect.Weather.CronSpec = "invalid-cron"
	config.Catv.Collect.Weather.Timeout.Connection = "30s"
	config.Catv.Collect.Weather.Timeout.Shut = "10s"

	err := Load(config)
	assert.Error(t, err)
}

func TestLoad_GlobalServiceIsSetAfterLoad(t *testing.T) {
	config := common.Config{}
	config.Catv.Collect.Weather.Use = true
	config.Catv.Collect.Weather.CronSpec = "0 * * * * *"
	config.Catv.Collect.Weather.Timeout.Connection = "30s"
	config.Catv.Collect.Weather.Timeout.Shut = "10s"

	err := Load(config)
	assert.NoError(t, err)
	assert.NotNil(t, GlobalService)

	Unload()
}
