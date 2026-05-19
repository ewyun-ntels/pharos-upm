package udp

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"ntels.com/pharos/core/pkg/common"
)

func TestDiagnostic_JSONMarshalling(t *testing.T) {
	tests := []struct {
		name       string
		diagnostic Diagnostic
		wantErr    bool
	}{
		{
			name: "Valid diagnostic data",
			diagnostic: Diagnostic{
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
				ChSid:       "CH001",
				ChNum:       "101",
				ChFreq:      "573000",
				ChMode:      "QAM256",
				PwrLvl:      "-10",
				Snr:         "35",
			},
			wantErr: false,
		},
		{
			name: "Diagnostic with minimal fields",
			diagnostic: Diagnostic{
				HostID:  "host-002",
				MacAddr: "11:22:33:44:55:66",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Marshal
			data, err := json.Marshal(tt.diagnostic)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.NotEmpty(t, data)

			// Unmarshal
			var decoded Diagnostic
			err = json.Unmarshal(data, &decoded)
			require.NoError(t, err)

			// Verify key fields
			assert.Equal(t, tt.diagnostic.HostID, decoded.HostID)
			assert.Equal(t, tt.diagnostic.MacAddr, decoded.MacAddr)
			assert.Equal(t, tt.diagnostic.ChSid, decoded.ChSid)
		})
	}
}

func TestDiagnosticEventHandler_Creation(t *testing.T) {
	handler := NewEventHandler(common.Config{}, TransmissionTypeDiagnostic)
	assert.NotNil(t, handler)
	assert.Equal(t, TransmissionTypeDiagnostic, handler.transmissionType)
}

func TestDiagnostic_ChannelFields(t *testing.T) {
	diagnostic := Diagnostic{
		HostID:      "test-host",
		MacAddr:     "00:11:22:33:44:55",
		ChSid:       "CH123",
		ChNum:       "456",
		ChFreq:      "573000",
		ChMode:      "QAM256",
		PwrLvl:      "-10",
		Snr:         "35",
		LoggingTime: "2024-01-01T00:00:00Z",
	}

	assert.Equal(t, "test-host", diagnostic.HostID)
	assert.Equal(t, "00:11:22:33:44:55", diagnostic.MacAddr)
	assert.Equal(t, "CH123", diagnostic.ChSid)
	assert.Equal(t, "456", diagnostic.ChNum)
	assert.Equal(t, "573000", diagnostic.ChFreq)
	assert.Equal(t, "QAM256", diagnostic.ChMode)
	assert.Equal(t, "-10", diagnostic.PwrLvl)
	assert.Equal(t, "35", diagnostic.Snr)
	assert.Equal(t, "2024-01-01T00:00:00Z", diagnostic.LoggingTime)
}

func BenchmarkDiagnostic_JSONMarshal(b *testing.B) {
	diagnostic := Diagnostic{
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
		ChSid:       "CH001",
		ChNum:       "101",
		PwrLvl:      "-10",
		Snr:         "35",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = json.Marshal(diagnostic)
	}
}
