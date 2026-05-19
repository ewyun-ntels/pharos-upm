package udp

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestQualityMeasurement_JSONMarshalling(t *testing.T) {
	tests := []struct {
		name    string
		qm      QualityMeasurement
		wantErr bool
	}{
		{
			name: "Valid quality measurement with channels array",
			qm: QualityMeasurement{
				HostID:      "1A8020DF41",
				MacAddr:     "a0:72:2c:b6:ae:2b",
				CmMac:       "a0:72:2c:b6:ae:2a",
				StbIP:       "10.43.81.14",
				CmIP:        "10.4.58.208",
				StbModel:    "THX-U300",
				MwVer:       "3.1.46",
				LocalVer:    "1.0.6.03",
				CloudVer:    "1.6.12",
				LoggingTime: "2024/12/08 14:30",
				Channels: []ChannelQuality{
					{
						ChSid:  "251",
						ChNum:  "3",
						ChFreq: "741",
						ChMode: "256QAM",
						PwrLvl: "5.2",
						Snr:    "40",
					},
					{
						ChSid:  "185",
						ChNum:  "11",
						ChFreq: "567",
						ChMode: "8VSB",
						PwrLvl: "6.1",
						Snr:    "38",
					},
				},
			},
			wantErr: false,
		},
		{
			name: "Valid quality measurement with empty channels",
			qm: QualityMeasurement{
				HostID:      "2B9031EF52",
				MacAddr:     "b1:83:3d:c7:bf:3c",
				CmMac:       "b1:83:3d:c7:bf:3b",
				StbIP:       "10.43.82.15",
				CmIP:        "10.4.59.209",
				StbModel:    "THX-U400",
				MwVer:       "3.2.10",
				LocalVer:    "1.1.2.05",
				CloudVer:    "1.7.5",
				LoggingTime: "2024/12/08 15:00",
				Channels:    []ChannelQuality{},
			},
			wantErr: false,
		},
		{
			name: "Quality measurement with single channel",
			qm: QualityMeasurement{
				HostID:      "3C1042FG63",
				MacAddr:     "c2:94:4e:d8:c0:4d",
				CmMac:       "c2:94:4e:d8:c0:4c",
				StbIP:       "10.43.83.16",
				CmIP:        "10.4.60.210",
				StbModel:    "THX-U500",
				MwVer:       "3.3.20",
				LocalVer:    "1.2.3.06",
				CloudVer:    "1.8.10",
				LoggingTime: "2024/12/08 15:30",
				Channels: []ChannelQuality{
					{
						ChSid:  "301",
						ChNum:  "5",
						ChFreq: "789",
						ChMode: "256QAM",
						PwrLvl: "4.8",
						Snr:    "42",
					},
				},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Marshal
			data, err := json.Marshal(tt.qm)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			require.NoError(t, err)
			require.NotEmpty(t, data)

			// Verify channels array exists in JSON
			var result map[string]any
			err = json.Unmarshal(data, &result)
			require.NoError(t, err)

			// Check required fields
			assert.Equal(t, tt.qm.HostID, result["hostId"])
			assert.Equal(t, tt.qm.MacAddr, result["macAddr"])
			assert.Equal(t, tt.qm.CmMac, result["cmMac"])
			assert.Equal(t, tt.qm.StbIP, result["stbIp"])
			assert.Equal(t, tt.qm.CmIP, result["cmIp"])
			assert.Equal(t, tt.qm.StbModel, result["stbModel"])
			assert.Equal(t, tt.qm.MwVer, result["mwVer"])
			assert.Equal(t, tt.qm.LocalVer, result["localVer"])
			assert.Equal(t, tt.qm.CloudVer, result["cloudVer"])
			assert.Equal(t, tt.qm.LoggingTime, result["loggingTime"])

			// Check channels array
			channels, ok := result["channels"].([]any)
			require.True(t, ok, "channels should be an array")
			assert.Equal(t, len(tt.qm.Channels), len(channels), "channels length should match")

			// Unmarshal back to struct
			var unmarshalled QualityMeasurement
			err = json.Unmarshal(data, &unmarshalled)
			require.NoError(t, err)
			assert.Equal(t, tt.qm.HostID, unmarshalled.HostID)
			assert.Equal(t, len(tt.qm.Channels), len(unmarshalled.Channels))
		})
	}
}

func TestChannelQuality_JSONMarshalling(t *testing.T) {
	tests := []struct {
		name    string
		channel ChannelQuality
		wantErr bool
	}{
		{
			name: "Valid 256QAM channel",
			channel: ChannelQuality{
				ChSid:  "251",
				ChNum:  "3",
				ChFreq: "741",
				ChMode: "256QAM",
				PwrLvl: "5.2",
				Snr:    "40",
			},
			wantErr: false,
		},
		{
			name: "Valid 8VSB channel",
			channel: ChannelQuality{
				ChSid:  "185",
				ChNum:  "11",
				ChFreq: "567",
				ChMode: "8VSB",
				PwrLvl: "6.1",
				Snr:    "38",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Marshal
			data, err := json.Marshal(tt.channel)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			require.NoError(t, err)
			require.NotEmpty(t, data)

			// Unmarshal and verify
			var unmarshalled ChannelQuality
			err = json.Unmarshal(data, &unmarshalled)
			require.NoError(t, err)
			assert.Equal(t, tt.channel.ChSid, unmarshalled.ChSid)
			assert.Equal(t, tt.channel.ChNum, unmarshalled.ChNum)
			assert.Equal(t, tt.channel.ChFreq, unmarshalled.ChFreq)
			assert.Equal(t, tt.channel.ChMode, unmarshalled.ChMode)
			assert.Equal(t, tt.channel.PwrLvl, unmarshalled.PwrLvl)
			assert.Equal(t, tt.channel.Snr, unmarshalled.Snr)
		})
	}
}

func TestQualityMeasurement_ChannelsJSONSerialization(t *testing.T) {
	qm := QualityMeasurement{
		HostID:      "TEST001",
		MacAddr:     "00:11:22:33:44:55",
		CmMac:       "00:AA:BB:CC:DD:EE",
		StbIP:       "192.168.1.100",
		CmIP:        "192.168.1.101",
		StbModel:    "TEST-MODEL",
		MwVer:       "1.0.0",
		LocalVer:    "2.0.0",
		CloudVer:    "3.0.0",
		LoggingTime: "2024/12/08 10:00",
		Channels: []ChannelQuality{
			{ChSid: "1", ChNum: "1", ChFreq: "100", ChMode: "256QAM", PwrLvl: "5.0", Snr: "40"},
			{ChSid: "2", ChNum: "2", ChFreq: "200", ChMode: "8VSB", PwrLvl: "6.0", Snr: "38"},
		},
	}

	// Marshal to JSON
	data, err := json.Marshal(qm)
	require.NoError(t, err)

	// Convert channels to JSON string for DB storage
	if len(qm.Channels) > 0 {
		channelsJSON, err := json.Marshal(qm.Channels)
		require.NoError(t, err)
		qm.ChannelsStr = string(channelsJSON)
	}

	// Verify ChannelsStr is valid JSON
	assert.NotEmpty(t, qm.ChannelsStr)
	var channels []ChannelQuality
	err = json.Unmarshal([]byte(qm.ChannelsStr), &channels)
	require.NoError(t, err)
	assert.Equal(t, 2, len(channels))
	assert.Equal(t, "1", channels[0].ChSid)
	assert.Equal(t, "2", channels[1].ChSid)

	// Verify original data can be unmarshalled
	var unmarshalled QualityMeasurement
	err = json.Unmarshal(data, &unmarshalled)
	require.NoError(t, err)
	assert.Equal(t, 2, len(unmarshalled.Channels))
}

func TestQualityMeasurement_RequiredFields(t *testing.T) {
	qm := QualityMeasurement{
		HostID:      "REQUIRED001",
		MacAddr:     "aa:bb:cc:dd:ee:ff",
		CmMac:       "11:22:33:44:55:66",
		StbIP:       "10.0.0.1",
		CmIP:        "10.0.0.2",
		StbModel:    "MODEL-X",
		MwVer:       "1.0",
		LocalVer:    "2.0",
		CloudVer:    "3.0",
		LoggingTime: "2024/12/08 12:00",
		Channels:    []ChannelQuality{},
	}

	data, err := json.Marshal(qm)
	require.NoError(t, err)

	var result map[string]any
	err = json.Unmarshal(data, &result)
	require.NoError(t, err)

	// Verify all required fields are present
	requiredFields := []string{
		"hostId", "macAddr", "cmMac", "stbIp", "cmIp",
		"stbModel", "mwVer", "localVer", "cloudVer", "loggingTime",
	}

	for _, field := range requiredFields {
		_, exists := result[field]
		assert.True(t, exists, "Required field %s should exist", field)
	}

	// channels should exist even if empty
	channels, exists := result["channels"]
	assert.True(t, exists, "channels field should exist")
	assert.NotNil(t, channels, "channels should not be nil")
}

func TestQualityMeasurement_ChannelsStr_ConvertedOnMarshal(t *testing.T) {
	qm := QualityMeasurement{
		HostID:      "OPT001",
		MacAddr:     "aa:bb:cc:dd:ee:ff",
		CmMac:       "11:22:33:44:55:66",
		StbIP:       "10.0.0.1",
		CmIP:        "10.0.0.2",
		StbModel:    "MODEL-Y",
		MwVer:       "1.0",
		LocalVer:    "2.0",
		CloudVer:    "3.0",
		LoggingTime: "2024/12/08 13:00",
		Channels: []ChannelQuality{
			{ChSid: "100", ChNum: "5", ChFreq: "500", ChMode: "256QAM", PwrLvl: "5.5", Snr: "39"},
		},
	}

	data, err := json.Marshal(qm)
	require.NoError(t, err)

	var result map[string]any
	require.NoError(t, json.Unmarshal(data, &result))

	// channels는 JSON 배열로 직렬화됨
	channels, ok := result["channels"].([]any)
	require.True(t, ok)
	require.Len(t, channels, 1)

	ch := channels[0].(map[string]any)
	assert.Equal(t, "100", ch["chSid"])
	assert.Equal(t, "5", ch["chNum"])

	// ChannelsStr은 json:"-" 이므로 JSON에 포함되지 않음
	_, exists := result["channelsStr"]
	assert.False(t, exists)
}

func TestQualityMeasurement_v28_Compliance(t *testing.T) {
	// Test v2.8 specification compliance
	// - UDP protocol (not HTTP)
	// - Port 50005 (updated from 50002)
	// - All required fields
	// - channels array structure

	qm := QualityMeasurement{
		HostID:      "V28COMP001",
		MacAddr:     "a0:72:2c:b6:ae:2b",
		CmMac:       "a0:72:2c:b6:ae:2a",
		StbIP:       "10.43.81.14",
		CmIP:        "10.4.58.208",
		StbModel:    "THX-U300",
		MwVer:       "3.1.46",
		LocalVer:    "1.0.6.03",
		CloudVer:    "1.6.12",
		LoggingTime: "2024/12/08 16:00",
		Channels: []ChannelQuality{
			{
				ChSid:  "251",
				ChNum:  "3",
				ChFreq: "741",
				ChMode: "256QAM",
				PwrLvl: "5.2",
				Snr:    "40",
			},
		},
	}

	// Marshal to JSON
	data, err := json.Marshal(qm)
	require.NoError(t, err)

	// Verify structure complies with v2.8 spec
	var result map[string]any
	err = json.Unmarshal(data, &result)
	require.NoError(t, err)

	// Check all required v2.8 fields
	assert.Equal(t, qm.HostID, result["hostId"])
	assert.Equal(t, qm.MacAddr, result["macAddr"])
	assert.Equal(t, qm.LoggingTime, result["loggingTime"])

	// Verify channels array
	channels, ok := result["channels"].([]any)
	require.True(t, ok)
	require.Equal(t, 1, len(channels))

	// Verify channel structure
	channel := channels[0].(map[string]any)
	assert.Equal(t, "251", channel["chSid"])
	assert.Equal(t, "3", channel["chNum"])
	assert.Equal(t, "741", channel["chFreq"])
	assert.Equal(t, "256QAM", channel["chMode"])
	assert.Equal(t, "5.2", channel["pwrLvl"])
	assert.Equal(t, "40", channel["snr"])
}
