package udp

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTransmissionTypeConstants(t *testing.T) {
	// 상수 값이 DB/테이블 파티션 키로 사용되므로 변경 시 데이터 불일치 발생
	assert.Equal(t, "daily", TransmissionTypeDaily)
	assert.Equal(t, "periodic", TransmissionTypePeriodic)
	assert.Equal(t, "diagnostic", TransmissionTypeDiagnostic)
	assert.Equal(t, "quality_measurement", TransmissionTypeQualityMeasurement)
	assert.Equal(t, "network_quality_transition", TransmissionTypeNetworkQualityTransition)
}

func TestTransmissionTypeConstants_Uniqueness(t *testing.T) {
	types := []string{
		TransmissionTypeDaily,
		TransmissionTypePeriodic,
		TransmissionTypeDiagnostic,
		TransmissionTypeQualityMeasurement,
		TransmissionTypeNetworkQualityTransition,
	}

	seen := make(map[string]bool)
	for _, typ := range types {
		assert.NotEmpty(t, typ)
		assert.False(t, seen[typ], "중복 상수 값: %q", typ)
		seen[typ] = true
	}
}
