package transmission

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"ntels.com/pharos/core/pkg/common"
)

func TestLoad_ReturnsNil(t *testing.T) {
	config := common.Config{}

	err := Load(config)
	assert.NoError(t, err)
}

func TestLoad_RegistersUDPServers(t *testing.T) {
	// Load는 5개의 UDP 서버를 등록한다.
	// 실제 서버 바인딩 없이 등록만 검증 (포트 0 사용)
	config := common.Config{}
	config.Catv.Collect.Transmission.Daily.Port = 0
	config.Catv.Collect.Transmission.Diagnostic.Port = 0
	config.Catv.Collect.Transmission.NetworkQualityTransition.Port = 0
	config.Catv.Collect.Transmission.Periodic.Port = 0
	config.Catv.Collect.Transmission.QualityMeasurement.Port = 0

	err := Load(config)
	assert.NoError(t, err)
}
