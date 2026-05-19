package udp

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"ntels.com/pharos/core/external/orm"
)

func TestRaw_JSONMarshalling(t *testing.T) {
	now := orm.Datetime{Time: time.Date(2024, 6, 15, 0, 0, 0, 0, time.UTC)}

	tests := []struct {
		name string
		raw  Raw
	}{
		{
			name: "에러 없음",
			raw: Raw{
				Timestamp:        now,
				TransmissionType: TransmissionTypePeriodic,
				RawData:          `{"hostId":"host-001"}`,
				Errors:           []string{},
			},
		},
		{
			name: "에러 있음",
			raw: Raw{
				Timestamp:        now,
				TransmissionType: TransmissionTypeDaily,
				RawData:          "not-json",
				Errors:           []string{"parse error", "insert error"},
			},
		},
		{
			name: "errors nil",
			raw: Raw{
				Timestamp:        now,
				TransmissionType: TransmissionTypeDiagnostic,
				RawData:          "data",
				Errors:           nil,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := json.Marshal(tt.raw)
			require.NoError(t, err)

			var decoded Raw
			require.NoError(t, json.Unmarshal(data, &decoded))

			assert.Equal(t, tt.raw.TransmissionType, decoded.TransmissionType)
			assert.Equal(t, tt.raw.RawData, decoded.RawData)
			assert.Equal(t, len(tt.raw.Errors), len(decoded.Errors))
			for i, e := range tt.raw.Errors {
				assert.Equal(t, e, decoded.Errors[i])
			}
		})
	}
}

func TestRaw_JSONTags(t *testing.T) {
	now := orm.Datetime{Time: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)}
	raw := Raw{
		Timestamp:        now,
		TransmissionType: TransmissionTypeQualityMeasurement,
		RawData:          "test-data",
		Errors:           []string{"err1"},
	}

	data, err := json.Marshal(raw)
	require.NoError(t, err)

	var m map[string]any
	require.NoError(t, json.Unmarshal(data, &m))

	// JSON 태그 키 확인
	assert.Contains(t, m, "transmission_type")
	assert.Contains(t, m, "raw_data")
	assert.Contains(t, m, "errors")
	assert.Equal(t, "quality_measurement", m["transmission_type"])
	assert.Equal(t, "test-data", m["raw_data"])
}

func TestRaw_Errors_AppendBehavior(t *testing.T) {
	raw := Raw{
		TransmissionType: TransmissionTypeNetworkQualityTransition,
		RawData:          "raw",
	}

	assert.Equal(t, TransmissionTypeNetworkQualityTransition, raw.TransmissionType)
	assert.Equal(t, "raw", raw.RawData)

	assert.Nil(t, raw.Errors)

	raw.Errors = append(raw.Errors, "first error")
	assert.Len(t, raw.Errors, 1)

	raw.Errors = append(raw.Errors, "second error")
	assert.Len(t, raw.Errors, 2)
	assert.Equal(t, "first error", raw.Errors[0])
	assert.Equal(t, "second error", raw.Errors[1])
}
