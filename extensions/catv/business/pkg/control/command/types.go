package command

import "math/rand/v2"

// Work type constants for settopbox control operations.
// These constants define all supported control commands that can be sent to settopbox devices.
//
// Categories:
//   - 정보 수집: sysCheck, stb_request_info
//   - 재부팅/재시작: smartReboot, stbRestart, stb_control_restart, stb_control_reset
//   - 단말제어: limitAge, tvLock, skipCh, easyBuying, resetPin, favCh, etc.
//
// Note: These string constants are interned by the Go compiler, providing memory efficiency
// when used repeatedly across the codebase. The validWorkTypes map uses these constants
// as keys, ensuring O(1) lookup performance with minimal memory overhead.
const (
	// 자가진단전송
	WorkTypeSysCheck = "sysCheck"

	// 전송 요청
	WorkTypeStbRequestInfo = "stb_request_info"

	// 스마트리부팅 (v2.7: stbReboot → smartReboot)
	WorkTypeSmartReboot = "smartReboot"

	// STB재시작 (v2.6 추가, v2.7: stbReset → stbRestart)
	WorkTypeStbRestart = "stbRestart"

	// 단말제어
	WorkTypeLimitAge      = "limitAge"
	WorkTypeTvLock        = "tvLock"
	WorkTypeSkipCh        = "skipCh"
	WorkTypeEasyBuying    = "easyBuying"
	WorkTypeResetPin      = "resetPin"
	WorkTypeFavCh         = "favCh"
	WorkTypeZappingAd     = "zappingAd"
	WorkTypeMiniEpg       = "miniEpg"
	WorkTypeMiniEpgAd     = "miniEpgAd"
	WorkTypeTvCaption     = "tvCaption"
	WorkTypeTvImpaired    = "tvImpaired"
	WorkTypeBarkerCh      = "barkerCh"
	WorkTypeVodView       = "vodView"
	WorkTypeVodRelay      = "vodRelay"
	WorkTypeResolution    = "resolution"
	WorkTypeAudioMode     = "audioMode"
	WorkTypeHdmiCec       = "hdmiCec"
	WorkTypeHdcp          = "hdcp"
	WorkTypeHdr           = "hdr"
	WorkTypeMobilePay     = "mobilePay"
	WorkTypeMorningAlarm  = "morningAlarm"
	WorkTypeBootMenu      = "bootMenu"
	WorkTypePmsOn         = "pmsOn"
	WorkTypeOneAdOn       = "oneAdOn"
	WorkTypeAudioLang     = "audioLang"
	WorkTypeStandbyMode   = "standbyMode"
	WorkTypeSavePwr       = "savePwr"
	WorkTypeVoiceGuide    = "voiceGuide"
	WorkTypeChUpDown      = "chUpDown"
	WorkTypeChDca         = "chDca"
	WorkTypeVolume        = "volume"
	WorkTypeShowMenu      = "showMenu"
	WorkTypeLimitContents = "limitContents"
	WorkTypeStbPower      = "stbPower"
	// SNMP 기반 제어 (TCP 기반 smartReboot/stbRestart와 별개)
	WorkTypeStbControlRestart = "stb_control_restart" // SNMP로 처리
	WorkTypeStbControlReset   = "stb_control_reset"   // SNMP로 처리
)

// AllWorkTypes contains all valid work type values.
// This is used to automatically generate the validation map.
var AllWorkTypes = []string{
	WorkTypeSysCheck,
	WorkTypeStbRequestInfo,
	WorkTypeSmartReboot,
	WorkTypeStbRestart,
	WorkTypeLimitAge,
	WorkTypeTvLock,
	WorkTypeSkipCh,
	WorkTypeEasyBuying,
	WorkTypeResetPin,
	WorkTypeFavCh,
	WorkTypeZappingAd,
	WorkTypeMiniEpg,
	WorkTypeMiniEpgAd,
	WorkTypeTvCaption,
	WorkTypeTvImpaired,
	WorkTypeBarkerCh,
	WorkTypeVodView,
	WorkTypeVodRelay,
	WorkTypeResolution,
	WorkTypeAudioMode,
	WorkTypeHdmiCec,
	WorkTypeHdcp,
	WorkTypeHdr,
	WorkTypeMobilePay,
	WorkTypeMorningAlarm,
	WorkTypeBootMenu,
	WorkTypePmsOn,
	WorkTypeOneAdOn,
	WorkTypeAudioLang,
	WorkTypeStandbyMode,
	WorkTypeSavePwr,
	WorkTypeVoiceGuide,
	WorkTypeChUpDown,
	WorkTypeChDca,
	WorkTypeVolume,
	WorkTypeShowMenu,
	WorkTypeLimitContents,
	WorkTypeStbPower,
	WorkTypeStbControlRestart,
	WorkTypeStbControlReset,
}

type StbControlResponse struct {
	Code    string
	Message string
}

func pickRandom(choices []string) string {
	if len(choices) == 0 {
		return ""
	}
	return choices[rand.IntN(len(choices))]
}
