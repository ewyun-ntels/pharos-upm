package command

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestWorkTypeConstants(t *testing.T) {
	assert.Equal(t, "sysCheck", WorkTypeSysCheck)
	assert.Equal(t, "stb_request_info", WorkTypeStbRequestInfo)
	assert.Equal(t, "smartReboot", WorkTypeSmartReboot)
	assert.Equal(t, "stbRestart", WorkTypeStbRestart)
	assert.Equal(t, "limitAge", WorkTypeLimitAge)
	assert.Equal(t, "stb_control_restart", WorkTypeStbControlRestart)
	assert.Equal(t, "stb_control_reset", WorkTypeStbControlReset)
}

func TestAllWorkTypesContainsAllConstants(t *testing.T) {
	set := make(map[string]bool, len(AllWorkTypes))
	for _, wt := range AllWorkTypes {
		set[wt] = true
	}

	expected := []string{
		WorkTypeSysCheck, WorkTypeStbRequestInfo, WorkTypeSmartReboot, WorkTypeStbRestart,
		WorkTypeLimitAge, WorkTypeTvLock, WorkTypeSkipCh, WorkTypeEasyBuying,
		WorkTypeResetPin, WorkTypeFavCh, WorkTypeZappingAd, WorkTypeMiniEpg,
		WorkTypeMiniEpgAd, WorkTypeTvCaption, WorkTypeTvImpaired, WorkTypeBarkerCh,
		WorkTypeVodView, WorkTypeVodRelay, WorkTypeResolution, WorkTypeAudioMode,
		WorkTypeHdmiCec, WorkTypeHdcp, WorkTypeHdr, WorkTypeMobilePay,
		WorkTypeMorningAlarm, WorkTypeBootMenu, WorkTypePmsOn, WorkTypeOneAdOn,
		WorkTypeAudioLang, WorkTypeStandbyMode, WorkTypeSavePwr, WorkTypeVoiceGuide,
		WorkTypeChUpDown, WorkTypeChDca, WorkTypeVolume, WorkTypeShowMenu,
		WorkTypeLimitContents, WorkTypeStbPower, WorkTypeStbControlRestart, WorkTypeStbControlReset,
	}
	for _, wt := range expected {
		assert.Truef(t, set[wt], "AllWorkTypes missing: %s", wt)
	}
}

func TestAllWorkTypesNoDuplicates(t *testing.T) {
	seen := make(map[string]int)
	for i, wt := range AllWorkTypes {
		assert.NotContainsf(t, seen, wt, "duplicate work type %q at index %d", wt, i)
		seen[wt] = i
	}
}

func TestStbControlResponseFields(t *testing.T) {
	resp := StbControlResponse{Code: "0", Message: "ok"}
	assert.Equal(t, "0", resp.Code)
	assert.Equal(t, "ok", resp.Message)
}
