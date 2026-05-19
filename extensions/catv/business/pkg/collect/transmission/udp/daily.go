package udp

import (
	"ntels.com/pharos/core/external/orm"
)

type Daily struct {
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
	LimitAge       *string       `json:"limitAge,omitempty" db:"limit_age"`
	TvLock         *string       `json:"tvLock,omitempty" db:"tv_lock"`
	SkipCh         *string       `json:"skipCh,omitempty" db:"skip_ch"`
	EasyBuying     *string       `json:"easyBuying,omitempty" db:"easy_buying"`
	FavCh          *string       `json:"favCh,omitempty" db:"fav_ch"`
	ZappingAd      *string       `json:"zappingAd,omitempty" db:"zapping_ad"`
	MiniEpg        *string       `json:"miniEpg,omitempty" db:"mini_epg"`
	MiniEpgAd      *string       `json:"miniEpgAd,omitempty" db:"mini_epg_ad"`
	TvCaption      *string       `json:"tvCaption,omitempty" db:"tv_caption"`
	TvImpaired     *string       `json:"tvImpaired,omitempty" db:"tv_impaired"`
	BarkerCh       *string       `json:"barkerCh,omitempty" db:"barker_ch"`
	VodView        *string       `json:"vodView,omitempty" db:"vod_view"`
	VodRelay       *string       `json:"vodRelay,omitempty" db:"vod_relay"`
	Resolution     *string       `json:"resolution,omitempty" db:"resolution"`
	AudioMode      *string       `json:"audioMode,omitempty" db:"audio_mode"`
	HdmiCec        *string       `json:"hdmiCec,omitempty" db:"hdmi_cec"`
	Hdcp           *string       `json:"hdcp,omitempty" db:"hdcp"`
	Hdr            *string       `json:"hdr,omitempty" db:"hdr"`
	MobilePay      *string       `json:"mobilePay,omitempty" db:"mobile_pay"`
	MorningAlarm   *string       `json:"morningAlarm,omitempty" db:"morning_alarm"`
	BootMenu       *string       `json:"bootMenu,omitempty" db:"boot_menu"`
	PmsOn          *string       `json:"pmsOn,omitempty" db:"pms_on"`
	OneAdOn        *string       `json:"oneAdOn,omitempty" db:"one_ad_on"`
	AudioLang      *string       `json:"audioLang,omitempty" db:"audio_lang"`
	StandbyMode    *string       `json:"standbyMode,omitempty" db:"standby_mode"`
	SavePwr        *string       `json:"savePwr,omitempty" db:"save_pwr"`
	VoiceGuide     *string       `json:"voiceGuide,omitempty" db:"voice_guide"`
	RunningTime    *string       `json:"runningTime,omitempty" db:"running_time"`
	RunningTimeSec *uint64       `json:"runningTimeSec,omitempty" db:"running_time_sec"`
	LimitContents  *string       `json:"limitContents,omitempty" db:"limit_contents"`
}

func (d *Daily) convertBatchFormat() (map[string]any, error) {
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
	} else {
		loggingTimeUTC = nil
	}

	var sendingTimeUTC any
	if d.SendingTimeUTC != nil {
		if v, err := d.SendingTimeUTC.Value(); err != nil {
			return nil, err
		} else {
			sendingTimeUTC = v
		}
	} else {
		sendingTimeUTC = nil
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
		"sending_time":     d.SendingTime,
		"sending_time_utc": sendingTimeUTC,
		"limit_age":        d.LimitAge,
		"tv_lock":          d.TvLock,
		"skip_ch":          d.SkipCh,
		"easy_buying":      d.EasyBuying,
		"fav_ch":           d.FavCh,
		"zapping_ad":       d.ZappingAd,
		"mini_epg":         d.MiniEpg,
		"mini_epg_ad":      d.MiniEpgAd,
		"tv_caption":       d.TvCaption,
		"tv_impaired":      d.TvImpaired,
		"barker_ch":        d.BarkerCh,
		"vod_view":         d.VodView,
		"vod_relay":        d.VodRelay,
		"resolution":       d.Resolution,
		"audio_mode":       d.AudioMode,
		"hdmi_cec":         d.HdmiCec,
		"hdcp":             d.Hdcp,
		"hdr":              d.Hdr,
		"mobile_pay":       d.MobilePay,
		"morning_alarm":    d.MorningAlarm,
		"boot_menu":        d.BootMenu,
		"pms_on":           d.PmsOn,
		"one_ad_on":        d.OneAdOn,
		"audio_lang":       d.AudioLang,
		"standby_mode":     d.StandbyMode,
		"save_pwr":         d.SavePwr,
		"voice_guide":      d.VoiceGuide,
		"running_time":     d.RunningTime,
		"running_time_sec": d.RunningTimeSec,
		"limit_contents":   d.LimitContents,
	}, nil
}

func (d *Daily) getTableName() string {
	return "dist_stb_transmission_daily"
}

func (d *Daily) setTimestamp(timestamp orm.Datetime) {
	d.Timestamp = timestamp
}
