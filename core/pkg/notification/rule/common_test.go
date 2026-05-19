package rule

import (
	"encoding/json"
	"testing"

	"ntels.com/pharos/core/pkg/common"
	notifcommon "ntels.com/pharos/core/pkg/notification/common"
	"ntels.com/pharos/core/pkg/notification/rule/snmp"
)

func Test_GetRuleTemplate(t *testing.T) {
	// Known type returns SNMP implementation
	r, err := GetRuleTemplate(notifcommon.TypeSNMP)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, ok := r.(*snmp.SNMP); !ok {
		t.Fatalf("expected *snmp.SNMP, got %T", r)
	}

	// Unknown type returns error
	if _, err := GetRuleTemplate("unknown"); err == nil {
		t.Fatalf("expected error for unknown notification type")
	}
}

func Test_GetRule_ErrorFromLoad(t *testing.T) {
	// Provide minimal data with unsupported version (0) so SNMP.Load returns error
	data := map[string]any{
		"name":    "test-snmp",
		"version": float64(0),
	}
	_, err := GetRule(notifcommon.TypeSNMP, common.Config{}, data)
	if err == nil {
		t.Fatalf("expected error due to unsupported snmp version")
	}
}

func Test_GetRuleFromStr_ErrorFromLoad(t *testing.T) {
	m := map[string]any{
		"name":    "test-snmp",
		"version": 0,
	}
	b, _ := json.Marshal(m)
	_, err := GetRuleFromStr(notifcommon.TypeSNMP, common.Config{}, b)
	if err == nil {
		t.Fatalf("expected error due to unsupported snmp version via JSON path")
	}
}
