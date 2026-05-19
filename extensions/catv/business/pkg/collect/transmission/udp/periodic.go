package udp

import (
	"ntels.com/pharos/core/external/orm"
)

type Periodic struct {
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
	SendingTime    string        `json:"sendingTime" db:"sending_time"`
	SendingTimeUTC *orm.Datetime `json:"sendingTimeUtc,omitempty" db:"sending_time_utc"`
	ChSid          *string       `json:"chSid,omitempty" db:"ch_sid"`
	ChNum          *string       `json:"chNum,omitempty" db:"ch_num"`
	ChName         *string       `json:"chName,omitempty" db:"ch_name"`
	ChPrg          *string       `json:"chPrg,omitempty" db:"ch_prg"`
	ChFreq         *string       `json:"chFreq,omitempty" db:"ch_freq"`
	ChMode         *string       `json:"chMode,omitempty" db:"ch_mode"`
	PwrLvl         *string       `json:"pwrLvl,omitempty" db:"pwr_lvl"`
	Snr            *string       `json:"snr,omitempty" db:"snr"`
	SigWeak        *string       `json:"sigWeak,omitempty" db:"sig_weak"`
	SigWeakCnt     *string       `json:"sigWeakCnt,omitempty" db:"sig_weak_cnt"`
	StbState       *string       `json:"stbState,omitempty" db:"stb_state"`
	RunningTime    *string       `json:"runningTime,omitempty" db:"running_time"`
	RunningTimeSec *uint64       `json:"runningTimeSec,omitempty" db:"running_time_sec"`
}

func (p *Periodic) convertBatchFormat() (map[string]any, error) {
	timestamp, err := p.Timestamp.Value()
	if err != nil {
		return nil, err
	}
	var loggingTimeUTC any
	if p.LoggingTimeUTC != nil {
		if v, err := p.LoggingTimeUTC.Value(); err != nil {
			return nil, err
		} else {
			loggingTimeUTC = v
		}
	}

	var sendingTimeUTC any
	if p.SendingTimeUTC != nil {
		if v, err := p.SendingTimeUTC.Value(); err != nil {
			return nil, err
		} else {
			sendingTimeUTC = v
		}
	}

	return map[string]any{
		"timestamp":        timestamp,
		"host_id":          p.HostID,
		"mac_addr":         p.MacAddr,
		"cm_mac":           p.CmMac,
		"stb_ip":           p.StbIP,
		"cm_ip":            p.CmIP,
		"stb_model":        p.StbModel,
		"mw_ver":           p.MwVer,
		"local_ver":        p.LocalVer,
		"cloud_ver":        p.CloudVer,
		"logging_time":     p.LoggingTime,
		"logging_time_utc": loggingTimeUTC,
		"sending_time":     p.SendingTime,
		"sending_time_utc": sendingTimeUTC,
		"ch_sid":           p.ChSid,
		"ch_num":           p.ChNum,
		"ch_name":          p.ChName,
		"ch_prg":           p.ChPrg,
		"ch_freq":          p.ChFreq,
		"ch_mode":          p.ChMode,
		"pwr_lvl":          p.PwrLvl,
		"snr":              p.Snr,
		"sig_weak":         p.SigWeak,
		"sig_weak_cnt":     p.SigWeakCnt,
		"stb_state":        p.StbState,
		"running_time":     p.RunningTime,
		"running_time_sec": p.RunningTimeSec,
	}, nil
}

func (p *Periodic) getTableName() string {
	return "dist_stb_transmission_periodic"
}

func (p *Periodic) setTimestamp(timestamp orm.Datetime) {
	p.Timestamp = timestamp
}
