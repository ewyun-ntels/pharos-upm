package udp

import (
	"ntels.com/pharos/core/external/orm"
)

// Diagnostic represents UDP 자가진단전송 (핫키 *106OK 입력 시)
// Note: TCP 자가진단 제어 (원격 요청)는 별도 HTTP 핸들러로 처리
type Diagnostic struct {
	Timestamp      orm.Datetime  `json:"-" db:"timestamp"`
	HostID         string        `json:"hostId" db:"host_id"`
	MacAddr        string        `json:"macAddr" db:"mac_addr"`
	CmMac          string        `json:"cmMac" db:"cm_mac"`
	StbIP          string        `json:"stbIp" db:"stb_ip"`
	CmIP           string        `json:"cmIp" db:"cm_ip"`
	StbModel       string        `json:"stbModel" db:"stb_model"`
	MwVer          string        `json:"mwVer" db:"mw_ver"`
	LocalVer       string        `json:"localVer" db:"local_ver"`
	CloudVer       string        `json:"cloudVer" db:"cloud_ver"`
	LoggingTime    string        `json:"loggingTime" db:"logging_time"`
	LoggingTimeUTC *orm.Datetime `json:"loggingTimeUtc,omitempty" db:"logging_time_utc"`
	ChSid          string        `json:"chSid" db:"ch_sid"`
	ChNum          string        `json:"chNum" db:"ch_num"`
	ChFreq         string        `json:"chFreq" db:"ch_freq"`
	ChMode         string        `json:"chMode" db:"ch_mode"`
	PwrLvl         string        `json:"pwrLvl" db:"pwr_lvl"`
	Snr            string        `json:"snr" db:"snr"`
}

func (d *Diagnostic) convertBatchFormat() (map[string]any, error) {
	timestamp, err := d.Timestamp.Value()
	if err != nil {
		return nil, err
	}

	var loggingTimeUTC any
	if d.LoggingTimeUTC != nil {
		if v, err := d.LoggingTimeUTC.Value(); err != nil {
			return nil, err
		} else {
			loggingTimeUTC = v
		}
	}

	return map[string]any{
		"timestamp":        timestamp,
		"host_id":          d.HostID,
		"mac_addr":         d.MacAddr,
		"cm_mac":           d.CmMac,
		"stb_ip":           d.StbIP,
		"cm_ip":            d.CmIP,
		"stb_model":        d.StbModel,
		"mw_ver":           d.MwVer,
		"local_ver":        d.LocalVer,
		"cloud_ver":        d.CloudVer,
		"logging_time":     d.LoggingTime,
		"logging_time_utc": loggingTimeUTC,
		"ch_sid":           d.ChSid,
		"ch_num":           d.ChNum,
		"ch_freq":          d.ChFreq,
		"ch_mode":          d.ChMode,
		"pwr_lvl":          d.PwrLvl,
		"snr":              d.Snr,
	}, nil
}

func (d *Diagnostic) getTableName() string {
	return "dist_stb_transmission_diagnostic"
}

func (d *Diagnostic) setTimestamp(timestamp orm.Datetime) {
	d.Timestamp = timestamp
}
