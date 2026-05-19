package udp

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"ntels.com/pharos/core/pkg/common"
)

func TestNetworkQualityTransition_JSONMarshalling(t *testing.T) {
	chName := "Channel 101"

	tests := []struct {
		name                     string
		networkQualityTransition NetworkQualityTransition
		wantErr                  bool
	}{
		{
			name: "Valid network quality transition data",
			networkQualityTransition: NetworkQualityTransition{
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
				ChName:      &chName,
				ChQamFreq:   "573000", // v2.5 updated
				ChQamMode:   "QAM256", // v2.5 updated
				// v2.2: ChQamPwrLvl, ChQamSnr 삭제됨
				Ch8vsbFreq:   "579000", // v2.5 updated
				Ch8vsbMode:   "8VSB",   // v2.5 updated
				Ch8vsbPwrLvl: "-11",    // v2.5 updated
				Ch8vsbSnr:    "34",     // v2.5 updated
			},
			wantErr: false,
		},
		{
			name: "Network quality transition with minimal fields",
			networkQualityTransition: NetworkQualityTransition{
				HostID:     "host-002",
				MacAddr:    "11:22:33:44:55:66",
				ChSid:      "CH002",
				ChNum:      "102",
				ChQamFreq:  "573000", // v2.5 updated
				Ch8vsbFreq: "579000", // v2.5 updated
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Marshal
			data, err := json.Marshal(tt.networkQualityTransition)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.NotEmpty(t, data)

			// Unmarshal
			var decoded NetworkQualityTransition
			err = json.Unmarshal(data, &decoded)
			require.NoError(t, err)

			// Verify key fields
			assert.Equal(t, tt.networkQualityTransition.HostID, decoded.HostID)
			assert.Equal(t, tt.networkQualityTransition.MacAddr, decoded.MacAddr)
			assert.Equal(t, tt.networkQualityTransition.ChSid, decoded.ChSid)
		})
	}
}

func TestNetworkQualityTransitionEventHandler_Creation(t *testing.T) {
	handler := NewEventHandler(common.Config{}, TransmissionTypeNetworkQualityTransition)
	assert.NotNil(t, handler)
	assert.Equal(t, TransmissionTypeNetworkQualityTransition, handler.transmissionType)
}

func TestNetworkQualityTransition_QAMFields(t *testing.T) {
	nqt := NetworkQualityTransition{
		HostID:    "test-host",
		MacAddr:   "00:11:22:33:44:55",
		ChSid:     "CH123",
		ChNum:     "456",
		ChQamFreq: "573000", // v2.5 updated
		ChQamMode: "QAM256", // v2.5 updated
		// v2.2: ChQamPwrLvl, ChQamSnr 삭제됨
	}

	assert.Equal(t, "test-host", nqt.HostID)
	assert.Equal(t, "00:11:22:33:44:55", nqt.MacAddr)
	assert.Equal(t, "CH123", nqt.ChSid)
	assert.Equal(t, "456", nqt.ChNum)
	assert.Equal(t, "573000", nqt.ChQamFreq)
	assert.Equal(t, "QAM256", nqt.ChQamMode)
	// v2.2: ChQamPwrLvl, ChQamSnr 삭제됨
}

func TestNetworkQualityTransition_VSBFields(t *testing.T) {
	nqt := NetworkQualityTransition{
		HostID:       "test-host",
		MacAddr:      "00:11:22:33:44:55",
		ChSid:        "CH123",
		ChNum:        "456",
		Ch8vsbFreq:   "579000", // v2.5 updated
		Ch8vsbMode:   "8VSB",   // v2.5 updated
		Ch8vsbPwrLvl: "-11",    // v2.5 updated
		Ch8vsbSnr:    "34",     // v2.5 updated
	}

	assert.Equal(t, "test-host", nqt.HostID)
	assert.Equal(t, "00:11:22:33:44:55", nqt.MacAddr)
	assert.Equal(t, "CH123", nqt.ChSid)
	assert.Equal(t, "456", nqt.ChNum)
	assert.Equal(t, "579000", nqt.Ch8vsbFreq)
	assert.Equal(t, "8VSB", nqt.Ch8vsbMode)
	assert.Equal(t, "-11", nqt.Ch8vsbPwrLvl)
	assert.Equal(t, "34", nqt.Ch8vsbSnr)
}

func BenchmarkNetworkQualityTransition_JSONMarshal(b *testing.B) {
	nqt := NetworkQualityTransition{
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
		ChQamFreq:   "573000", // v2.5 updated
		ChQamMode:   "QAM256", // v2.5 updated
		Ch8vsbFreq:  "579000", // v2.5 updated
		Ch8vsbMode:  "8VSB",   // v2.5 updated
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = json.Marshal(nqt)
	}
}
