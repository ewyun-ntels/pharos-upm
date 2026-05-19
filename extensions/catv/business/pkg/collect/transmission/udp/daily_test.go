package udp

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"ntels.com/pharos/core/pkg/common"
)

func TestDaily_JSONMarshalling(t *testing.T) {
	tests := []struct {
		name    string
		daily   Daily
		wantErr bool
	}{
		{
			name: "Valid daily data",
			daily: Daily{
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
			name: "Daily data with minimal fields",
			daily: Daily{
				HostID:  "host-002",
				MacAddr: "11:22:33:44:55:66",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Marshal
			data, err := json.Marshal(tt.daily)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.NotEmpty(t, data)

			// Unmarshal
			var decoded Daily
			err = json.Unmarshal(data, &decoded)
			require.NoError(t, err)

			// Verify key fields
			assert.Equal(t, tt.daily.HostID, decoded.HostID)
			assert.Equal(t, tt.daily.MacAddr, decoded.MacAddr)
			assert.Equal(t, tt.daily.CmMac, decoded.CmMac)
		})
	}
}

func TestDaily_AllFields(t *testing.T) {
	limitAge := "18"
	tvLock := "enabled"
	resolution := "1920x1080"
	hdmiCec := "on"

	daily := Daily{
		HostID:      "test-host",
		MacAddr:     "00:11:22:33:44:55",
		CmMac:       "AA:BB:CC:DD:EE:FF",
		StbIP:       "10.0.0.1",
		CmIP:        "10.0.0.2",
		StbModel:    "Model-X",
		MwVer:       "1.0",
		LocalVer:    "2.0",
		CloudVer:    "3.0",
		LoggingTime: "2024-01-01T00:00:00Z",
		SendingTime: "2024-01-01T00:01:00Z",
		LimitAge:    &limitAge,
		TvLock:      &tvLock,
		Resolution:  &resolution,
		HdmiCec:     &hdmiCec,
	}

	// Validate all fields are accessible
	assert.Equal(t, "test-host", daily.HostID)
	assert.Equal(t, "00:11:22:33:44:55", daily.MacAddr)
	assert.Equal(t, "AA:BB:CC:DD:EE:FF", daily.CmMac)
	assert.Equal(t, "10.0.0.1", daily.StbIP)
	assert.Equal(t, "10.0.0.2", daily.CmIP)
	assert.Equal(t, "Model-X", daily.StbModel)
	assert.Equal(t, "1.0", daily.MwVer)
	assert.Equal(t, "2.0", daily.LocalVer)
	assert.Equal(t, "3.0", daily.CloudVer)
	assert.Equal(t, "2024-01-01T00:00:00Z", daily.LoggingTime)
	assert.Equal(t, "2024-01-01T00:01:00Z", daily.SendingTime)
	assert.NotNil(t, daily.LimitAge)
	assert.Equal(t, "18", *daily.LimitAge)
	assert.NotNil(t, daily.TvLock)
	assert.Equal(t, "enabled", *daily.TvLock)
	assert.NotNil(t, daily.Resolution)
	assert.Equal(t, "1920x1080", *daily.Resolution)
	assert.NotNil(t, daily.HdmiCec)
	assert.Equal(t, "on", *daily.HdmiCec)
}

func TestDailyEventHandler_Creation(t *testing.T) {
	handler := NewEventHandler(common.Config{}, TransmissionTypeDaily)
	assert.NotNil(t, handler)
	assert.Equal(t, TransmissionTypeDaily, handler.transmissionType)
}

func TestDaily_DBTags(t *testing.T) {
	// Verify struct tags are correctly defined
	daily := Daily{}

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

	err = json.Unmarshal(data, &daily)
	require.NoError(t, err)

	assert.Equal(t, "test-host", daily.HostID)
	assert.Equal(t, "00:11:22:33:44:55", daily.MacAddr)
	assert.Equal(t, "STB-100", daily.StbModel)
}

func TestDaily_EmptyValues(t *testing.T) {
	data := []byte(`{}`)

	var daily Daily
	err := json.Unmarshal(data, &daily)
	require.NoError(t, err)

	// All fields should be zero values
	assert.Empty(t, daily.HostID)
	assert.Empty(t, daily.MacAddr)
	assert.Empty(t, daily.StbModel)
}

func TestDaily_LargePayload(t *testing.T) {
	runningTime := "12345678"
	audioLang := "Korean-한국어-with-many-characters"

	daily := Daily{
		HostID:      "host-with-very-long-id-that-exceeds-normal-length",
		MacAddr:     "FF:FF:FF:FF:FF:FF",
		CmMac:       "EE:EE:EE:EE:EE:EE",
		StbIP:       "255.255.255.255",
		CmIP:        "255.255.255.254",
		StbModel:    "This is a very long model name that should still work",
		MwVer:       "1.2.3.4.5.6.7.8.9.10",
		LocalVer:    "Local version with many characters and special symbols !@#$%",
		CloudVer:    "Cloud version equally long with unicode 한글 테스트",
		LoggingTime: "2024-12-31T23:59:59.999999Z",
		SendingTime: "2024-12-31T23:59:59.999999Z",
		RunningTime: &runningTime,
		AudioLang:   &audioLang,
	}

	data, err := json.Marshal(daily)
	require.NoError(t, err)
	assert.True(t, len(data) > 200, "Expected large payload")

	var decoded Daily
	err = json.Unmarshal(data, &decoded)
	require.NoError(t, err)
	assert.Equal(t, daily.HostID, decoded.HostID)
	assert.Equal(t, daily.CloudVer, decoded.CloudVer)
}

func BenchmarkDaily_JSONMarshal(b *testing.B) {
	daily := Daily{
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
		_, _ = json.Marshal(daily)
	}
}

func BenchmarkDaily_JSONUnmarshal(b *testing.B) {
	data := []byte(`{"hostId":"host-001","macAddr":"00:11:22:33:44:55","cmMac":"AA:BB:CC:DD:EE:FF","stbIp":"192.168.1.100","cmIp":"192.168.1.101","stbModel":"STB-X100","mwVer":"1.0.0","localVer":"2.0.0","cloudVer":"3.0.0","loggingTime":"2024-01-15T10:30:00Z"}`)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var daily Daily
		_ = json.Unmarshal(data, &daily)
	}
}
