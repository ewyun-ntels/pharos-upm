package udp

import (
	"encoding/json"

	"ntels.com/pharos/core/external/orm"
)

type ChannelQuality struct {
	ChSid  string `json:"chSid"`
	ChNum  string `json:"chNum"`
	ChFreq string `json:"chFreq"`
	ChMode string `json:"chMode"`
	PwrLvl string `json:"pwrLvl"`
	Snr    string `json:"snr"`
}

type QualityMeasurement struct {
	Timestamp      orm.Datetime     `json:"-" db:"timestamp"`
	HostID         string           `json:"hostId" db:"host_id"`
	MacAddr        string           `json:"macAddr" db:"mac_addr"`
	CmMac          string           `json:"cmMac" db:"cm_mac"`
	StbIP          string           `json:"stbIp" db:"stb_ip"`
	CmIP           string           `json:"cmIp" db:"cm_ip"`
	StbModel       string           `json:"stbModel" db:"stb_model"`
	MwVer          string           `json:"mwVer" db:"mw_ver"`
	LocalVer       string           `json:"localVer" db:"local_ver"`
	CloudVer       string           `json:"cloudVer" db:"cloud_ver"`
	LoggingTime    string           `json:"loggingTime" db:"logging_time"`
	LoggingTimeUTC *orm.Datetime    `json:"loggingTimeUtc,omitempty" db:"logging_time_utc"`
	Channels       []ChannelQuality `json:"channels" db:"-"` // JSON array
	ChannelsStr    string           `json:"-" db:"channels"` // For DB storage
}

func (q *QualityMeasurement) convertBatchFormat() (map[string]any, error) {
	timestamp, err := q.Timestamp.Value()
	if err != nil {
		return nil, err
	}

	var loggingTimeUTC any
	if q.LoggingTimeUTC != nil {
		if v, err := q.LoggingTimeUTC.Value(); err != nil {
			return nil, err
		} else {
			loggingTimeUTC = v
		}
	}

	var channelsStr string
	if len(q.Channels) > 0 {
		channelsBytes, err := json.Marshal(q.Channels)
		if err != nil {
			return nil, err
		}
		channelsStr = string(channelsBytes)
	}

	return map[string]any{
		"timestamp":        timestamp,
		"host_id":          q.HostID,
		"mac_addr":         q.MacAddr,
		"cm_mac":           q.CmMac,
		"stb_ip":           q.StbIP,
		"cm_ip":            q.CmIP,
		"stb_model":        q.StbModel,
		"mw_ver":           q.MwVer,
		"local_ver":        q.LocalVer,
		"cloud_ver":        q.CloudVer,
		"logging_time":     q.LoggingTime,
		"logging_time_utc": loggingTimeUTC,
		"channels":         channelsStr, // Store JSON string in DB
	}, nil
}

func (q *QualityMeasurement) getTableName() string {
	return "dist_stb_transmission_quality_measurement"
}

func (q *QualityMeasurement) setTimestamp(timestamp orm.Datetime) {
	q.Timestamp = timestamp
}
