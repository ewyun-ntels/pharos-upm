package udp

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"ntels.com/pharos/core/pkg/common"
)

func TestPeriodic_JSONMarshalling(t *testing.T) {
	tests := []struct {
		name     string
		periodic Periodic
		wantErr  bool
	}{
		{
			name: "Valid periodic data",
			periodic: Periodic{
				HostID:      "host-001",
				MacAddr:     "00:11:22:33:44:55",
				CmMac:       "00:AA:BB:CC:DD:EE",
				StbIP:       "192.168.1.100",
				CmIP:        "192.168.1.101",
				StbModel:    "STB-X100",
				MwVer:       "1.0.0",
				LocalVer:    "2.0.0",
				CloudVer:    "3.0.0",
				LoggingTime: "2024-01-15T10:30:00Z",
			},
			wantErr: false,
		},
		{
			name: "Periodic with minimal fields",
			periodic: Periodic{
				HostID:  "host-002",
				MacAddr: "11:22:33:44:55:66",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Marshal
			data, err := json.Marshal(tt.periodic)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.NotEmpty(t, data)

			// Unmarshal
			var decoded Periodic
			err = json.Unmarshal(data, &decoded)
			require.NoError(t, err)

			// Verify key fields
			assert.Equal(t, tt.periodic.HostID, decoded.HostID)
			assert.Equal(t, tt.periodic.MacAddr, decoded.MacAddr)
			assert.Equal(t, tt.periodic.CmMac, decoded.CmMac)
		})
	}
}

func TestPeriodicEventHandler_Creation(t *testing.T) {
	handler := NewEventHandler(common.Config{}, TransmissionTypePeriodic)
	assert.NotNil(t, handler)
	assert.Equal(t, TransmissionTypePeriodic, handler.transmissionType)
}

func TestPeriodic_DBTags(t *testing.T) {
	periodic := Periodic{}

	data, err := json.Marshal(map[string]string{
		"hostId":      "test-host",
		"macAddr":     "00:11:22:33:44:55",
		"cmMac":       "AA:BB:CC:DD:EE:FF",
		"stbIp":       "192.168.1.1",
		"cmIp":        "192.168.1.2",
		"stbModel":    "STB-100",
		"mwVer":       "1.0",
		"localVer":    "2.0",
		"cloudVer":    "3.0",
		"loggingTime": "2024-01-01T00:00:00Z",
	})
	require.NoError(t, err)

	err = json.Unmarshal(data, &periodic)
	require.NoError(t, err)

	assert.Equal(t, "test-host", periodic.HostID)
	assert.Equal(t, "00:11:22:33:44:55", periodic.MacAddr)
	assert.Equal(t, "STB-100", periodic.StbModel)
}

func BenchmarkPeriodic_JSONMarshal(b *testing.B) {
	periodic := Periodic{
		HostID:      "host-001",
		MacAddr:     "00:11:22:33:44:55",
		CmMac:       "AA:BB:CC:DD:EE:FF",
		StbIP:       "192.168.1.100",
		CmIP:        "192.168.1.101",
		StbModel:    "STB-X100",
		MwVer:       "1.0.0",
		LocalVer:    "2.0.0",
		CloudVer:    "3.0.0",
		LoggingTime: "2024-01-15T10:30:00Z",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = json.Marshal(periodic)
	}
}

func BenchmarkPeriodic_JSONUnmarshal(b *testing.B) {
	data := []byte(`{"hostId":"host-001","macAddr":"00:11:22:33:44:55","cmMac":"AA:BB:CC:DD:EE:FF","stbIp":"192.168.1.100","cmIp":"192.168.1.101","stbModel":"STB-X100","mwVer":"1.0.0","localVer":"2.0.0","cloudVer":"3.0.0","loggingTime":"2024-01-15T10:30:00Z"}`)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var periodic Periodic
		_ = json.Unmarshal(data, &periodic)
	}
}
