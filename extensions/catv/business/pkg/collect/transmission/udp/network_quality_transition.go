package udp

import (
	"ntels.com/pharos/core/external/orm"
)

type NetworkQualityTransition struct {
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
	ChName         *string       `json:"chName,omitempty" db:"ch_name"`
	ChQamFreq      string        `json:"chQamFreq" db:"ch_qam_freq"` // v2.5: qamChFreq → chQamFreq
	ChQamMode      string        `json:"chQamMode" db:"ch_qam_mode"` // v2.5: qamChMode → chQamMode
	// v2.2: qamChPwrLvl, qamChSnr 삭제됨
	Ch8vsbFreq   string `json:"ch8vsbFreq" db:"ch_8vsb_freq"`      // v2.5: vsbChFreq → ch8vsbFreq
	Ch8vsbMode   string `json:"ch8vsbMode" db:"ch_8vsb_mode"`      // v2.5: vsbChMode → ch8vsbMode
	Ch8vsbPwrLvl string `json:"ch8vsbPwrLvl" db:"ch_8vsb_pwr_lvl"` // v2.5: vsbChPwrLvl → ch8vsbPwrLvl
	Ch8vsbSnr    string `json:"ch8vsbSnr" db:"ch_8vsb_snr"`        // v2.5: vsbChSnr → ch8vsbSnr
}

func (n *NetworkQualityTransition) convertBatchFormat() (map[string]any, error) {
	timestamp, err := n.Timestamp.Value()
	if err != nil {
		return nil, err
	}

	var loggingTimeUTC any
	if n.LoggingTimeUTC != nil {
		if v, err := n.LoggingTimeUTC.Value(); err != nil {
			return nil, err
		} else {
			loggingTimeUTC = v
		}
	}

	return map[string]any{
		"timestamp":        timestamp,
		"host_id":          n.HostID,
		"mac_addr":         n.MacAddr,
		"cm_mac":           n.CmMac,
		"stb_ip":           n.StbIP,
		"cm_ip":            n.CmIP,
		"stb_model":        n.StbModel,
		"mw_ver":           n.MwVer,
		"local_ver":        n.LocalVer,
		"cloud_ver":        n.CloudVer,
		"logging_time":     n.LoggingTime,
		"logging_time_utc": loggingTimeUTC,
		"ch_sid":           n.ChSid,
		"ch_num":           n.ChNum,
		"ch_name":          n.ChName,
		"ch_qam_freq":      n.ChQamFreq,
		"ch_qam_mode":      n.ChQamMode,
		"ch_8vsb_freq":     n.Ch8vsbFreq,
		"ch_8vsb_mode":     n.Ch8vsbMode,
		"ch_8vsb_pwr_lvl":  n.Ch8vsbPwrLvl,
		"ch_8vsb_snr":      n.Ch8vsbSnr,
	}, nil
}

func (n *NetworkQualityTransition) getTableName() string {
	return "dist_stb_transmission_network_quality_transition"
}

func (n *NetworkQualityTransition) setTimestamp(timestamp orm.Datetime) {
	n.Timestamp = timestamp
}
