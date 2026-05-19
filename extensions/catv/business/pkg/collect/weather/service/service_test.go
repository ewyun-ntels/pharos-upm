package service

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"ntels.com/pharos/core/pkg/common"
)

func makeValidConfig() common.Config {
	config := common.Config{}
	config.Catv.Collect.Weather.CronSpec = "0 * * * * *"
	config.Catv.Collect.Weather.Timeout.Connection = "30s"
	config.Catv.Collect.Weather.Timeout.Shut = "10s"
	return config
}

func TestNewService_WithValidConfig(t *testing.T) {
	config := makeValidConfig()

	svc, err := NewService(config)

	require.NoError(t, err)
	require.NotNil(t, svc)
}

func TestNewService_WithInvalidConnectionTimeout(t *testing.T) {
	config := common.Config{}
	config.Catv.Collect.Weather.Timeout.Connection = "not-a-duration"
	config.Catv.Collect.Weather.Timeout.Shut = "10s"

	svc, err := NewService(config)

	assert.Error(t, err)
	assert.Nil(t, svc)
	assert.Contains(t, err.Error(), "invalid FTP connection timeout")
}

func TestNewService_WithInvalidShutTimeout(t *testing.T) {
	config := common.Config{}
	config.Catv.Collect.Weather.Timeout.Connection = "30s"
	config.Catv.Collect.Weather.Timeout.Shut = "bad-value"

	svc, err := NewService(config)

	assert.Error(t, err)
	assert.Nil(t, svc)
	assert.Contains(t, err.Error(), "invalid FTP shutdown timeout")
}

func TestService_StartStop(t *testing.T) {
	config := makeValidConfig()

	svc, err := NewService(config)
	require.NoError(t, err)

	err = svc.Start()
	require.NoError(t, err)

	err = svc.Stop()
	assert.NoError(t, err)
}

func TestService_StartWithInvalidCronSpec(t *testing.T) {
	config := common.Config{}
	config.Catv.Collect.Weather.CronSpec = "INVALID_CRON"
	config.Catv.Collect.Weather.Timeout.Connection = "30s"
	config.Catv.Collect.Weather.Timeout.Shut = "10s"

	svc, err := NewService(config)
	require.NoError(t, err)

	err = svc.Start()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to add FTP handler to cron")
}

func TestService_StopBeforeStart(t *testing.T) {
	config := makeValidConfig()

	svc, err := NewService(config)
	require.NoError(t, err)

	// Start 없이 Stop 호출도 안전해야 함
	err = svc.Stop()
	assert.NoError(t, err)
}

func TestService_HasFileNameConstants(t *testing.T) {
	assert.Equal(t, "3hour.dat", FILE_NAME_3HOUR)
	assert.Equal(t, "land.dat", FILE_NAME_LAND)
	assert.Equal(t, "shko.dat", FILE_NAME_SHKO)
}

func TestService_MultipleStartStop(t *testing.T) {
	config := makeValidConfig()

	svc1, err := NewService(config)
	require.NoError(t, err)

	svc2, err := NewService(config)
	require.NoError(t, err)

	// 두 인스턴스는 독립적이어야 함
	assert.NotEqual(t, svc1, svc2)

	require.NoError(t, svc1.Start())
	require.NoError(t, svc2.Start())

	assert.NoError(t, svc1.Stop())
	assert.NoError(t, svc2.Stop())
}
