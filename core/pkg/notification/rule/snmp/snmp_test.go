package snmp

import (
	"encoding/hex"
	"errors"
	"testing"

	g "github.com/gosnmp/gosnmp"
	notification_common "ntels.com/pharos/core/pkg/notification/common"
)

func TestParseEngineIDHex_VariousFormats(t *testing.T) {
	cases := []struct {
		in   string
		want []byte
	}{
		{"8000137003", []byte{0x80, 0x00, 0x13, 0x70, 0x03}},
		{"0x8000137003", []byte{0x80, 0x00, 0x13, 0x70, 0x03}},
		{"80:00:13:70:03", []byte{0x80, 0x00, 0x13, 0x70, 0x03}},
		{"80 00 13 70 03", []byte{0x80, 0x00, 0x13, 0x70, 0x03}},
	}
	for _, c := range cases {
		got := parseEngineIDHex(c.in)
		if hex.EncodeToString(got) != hex.EncodeToString(c.want) {
			t.Fatalf("parseEngineIDHex(%q) = %x, want %x", c.in, got, c.want)
		}
	}
}

func TestParseEngineIDHex_Invalid(t *testing.T) {
	if b := parseEngineIDHex("80:0"); b != nil {
		t.Fatalf("expected nil for odd length, got %x", b)
	}
}

// fake trap capturer
type trapCapture struct {
	called bool
	trap   g.SnmpTrap
}

func TestSNMP_Validate(t *testing.T) {
	// invalid transport
	s := &SNMP{Transport: "quic"}
	if err := s.Validate(); err == nil {
		t.Fatalf("expected invalid transport error")
	}
	// duplicate label names
	s = &SNMP{TrapPduRule: []TrapPduRule{{LabelName: "a", Oid: "1.2"}, {LabelName: "a", Oid: "1.3"}}}
	if err := s.Validate(); err == nil {
		t.Fatalf("expected duplicate label error")
	}
	// duplicate OIDs
	s = &SNMP{TrapPduRule: []TrapPduRule{{LabelName: "a", Oid: "1.2"}, {LabelName: "b", Oid: "1.2"}}}
	if err := s.Validate(); err == nil {
		t.Fatalf("expected duplicate OID error")
	}
	// invalid trap rule
	s = &SNMP{TrapPduRule: []TrapPduRule{{LabelName: "", Oid: ""}}}
	if err := s.Validate(); err == nil {
		t.Fatalf("expected trap rule validation error")
	}
	// invalid append pdu
	s = &SNMP{AppendPdu: []AppendPdu{{Oid: "1.2.3", ValueType: "unknown", Value: 1}}}
	if err := s.Validate(); err == nil {
		t.Fatalf("expected append pdu error")
	}
	// valid basic
	s = &SNMP{Name: "UnitTest", TrapPduRule: []TrapPduRule{{LabelName: "a", Oid: "1.2"}}}
	if err := s.Validate(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestSNMP_Send_BuildsTrapAndRespectsFilters(t *testing.T) {
	// capture hook
	var capture trapCapture
	old := sendTrapFunc
	sendTrapFunc = func(c *g.GoSNMP, trap g.SnmpTrap) (*g.SnmpPacket, error) {
		capture.called = true
		capture.trap = trap
		return nil, nil
	}
	defer func() { sendTrapFunc = old }()

	s := &SNMP{
		TrapPduRule: []TrapPduRule{
			{LabelName: "a", Oid: "1.3.6.1.4.1.1"},
			{LabelName: "b", Oid: "1.3.6.1.4.1.2"},
		},
		ExcludeFilter: map[string]any{"b": "X"},
	}
	// build map for Send()
	s.trapPduRuleMap = map[string]TrapPduRule{"a": s.TrapPduRule[0], "b": s.TrapPduRule[1]}

	descOid := "1.3.6.1.4.1.9"
	valOid := "1.3.6.1.4.1.10"
	stOid := "1.3.6.1.4.1.11"
	s.DescriptionOid = &descOid
	s.ValueOid = &valOid
	s.StatusOid = &stOid

	s.AppendPdu = []AppendPdu{{Oid: "1.3.6.1.6.3.1.1.4.1.0", ValueType: ValueTypeObjectIdentifier, Value: "1.3.6.1.4.1.9608.7.1.2.1"}}

	val := &notification_common.AlertValue{
		Description: "desc",
		Value:       3.14,
		Status:      "alerting",
		Labels:      map[string]string{"a": "hello", "b": "X", "c": "ignored"},
	}

	// Send should succeed and build expected variables
	if err := s.Send(val); err != nil {
		t.Fatalf("Send error: %v", err)
	}
	if !capture.called {
		t.Fatalf("expected SendTrap to be called")
	}
	// expected variables: label 'a' + description + value + status + 1 append = 5
	if len(capture.trap.Variables) != 5 {
		t.Fatalf("unexpected var count: %d", len(capture.trap.Variables))
	}
	// ensure OIDs present
	names := map[string]bool{}
	for _, v := range capture.trap.Variables {
		names[v.Name] = true
	}
	for _, want := range []string{"1.3.6.1.4.1.1", descOid, valOid, stOid, "1.3.6.1.6.3.1.1.4.1.0"} {
		if !names[want] {
			t.Fatalf("missing OID %s", want)
		}
	}
}

func TestSNMP_Send_ErrorOnPduRuleParse(t *testing.T) {
	// cause getPdu to fail by forcing numeric type with non-numeric value
	old := sendTrapFunc
	sendTrapFunc = func(c *g.GoSNMP, trap g.SnmpTrap) (*g.SnmpPacket, error) { return nil, nil }
	defer func() { sendTrapFunc = old }()

	vt := ValueTypeCounter32
	s := &SNMP{
		TrapPduRule: []TrapPduRule{{LabelName: "num", Oid: "1.2.3", ValueType: &vt}},
	}
	s.trapPduRuleMap = map[string]TrapPduRule{"num": s.TrapPduRule[0]}

	val := &notification_common.AlertValue{Labels: map[string]string{"num": "NaN"}}
	if err := s.Send(val); err == nil {
		t.Fatalf("expected error from getPdu parse failure")
	}

	// nil value error
	if err := s.Send(nil); err == nil {
		t.Fatalf("expected error for nil value")
	}
}

func TestGetAsnBER(t *testing.T) {
	cases := []struct {
		in   string
		want g.Asn1BER
		ok   bool
	}{
		{ValueTypeOctetString, g.OctetString, true},
		{ValueTypeCounter32, g.Counter32, true},
		{ValueTypeGauge32, g.Gauge32, true},
		{ValueTypeCounter64, g.Counter64, true},
		{ValueTypeUinteger32, g.Uinteger32, true},
		{ValueTypeOpaqueFloat, g.OpaqueFloat, true},
		{ValueTypeOpaqueDouble, g.OpaqueDouble, true},
		{ValueTypeObjectIdentifier, g.ObjectIdentifier, true},
		{"unknown", 0, false},
	}
	for _, c := range cases {
		got, err := getAsnBER(c.in)
		if c.ok {
			if err != nil || got != c.want {
				t.Fatalf("getAsnBER(%q) got=%v err=%v", c.in, got, err)
			}
		} else if err == nil {
			t.Fatalf("getAsnBER(%q) expected error", c.in)
		}
	}
}

func TestGetAuthPrivProtocol(t *testing.T) {
	if getAuthProtocol("sha256") != g.SHA256 {
		t.Fatal("auth proto map failed")
	}
	if getAuthProtocol("md5") != g.MD5 {
		t.Fatal("auth proto map failed")
	}
	if getAuthProtocol("unknown") != g.NoAuth {
		t.Fatal("default noauth failed")
	}

	if getPrivProtocol("aes") != g.AES {
		t.Fatal("priv proto map failed")
	}
	if getPrivProtocol("des") != g.DES {
		t.Fatal("priv proto map failed")
	}
	if getPrivProtocol("unknown") != g.NoPriv {
		t.Fatal("default nopriv failed")
	}
}

func TestCheckSnmpTypeMatch_Success(t *testing.T) {
	// Integer
	v, err := checkSnmpTypeMatch(g.Integer, float64(10))
	if err != nil || v.(int) != 10 {
		t.Fatalf("Integer mismatch: %v %v", v, err)
	}
	// Counter32/Gauge32/TimeTicks as uint32
	for _, typ := range []g.Asn1BER{g.Counter32, g.Gauge32, g.TimeTicks} {
		vv, err := checkSnmpTypeMatch(typ, float64(123))
		if err != nil {
			t.Fatalf("%v err: %v", typ, err)
		}
		if _, ok := vv.(uint32); !ok {
			t.Fatalf("%v not uint32", typ)
		}
	}
	// Counter64
	vv, err := checkSnmpTypeMatch(g.Counter64, float64(123))
	if err != nil {
		t.Fatalf("Counter64 err: %v", err)
	}
	if _, ok := vv.(uint64); !ok {
		t.Fatalf("Counter64 not uint64")
	}
	// String-like
	for _, typ := range []g.Asn1BER{g.OctetString, g.ObjectIdentifier, g.IPAddress} {
		vv, err := checkSnmpTypeMatch(typ, "abc")
		if err != nil || vv.(string) != "abc" {
			t.Fatalf("string types mismatch: %v %v", vv, err)
		}
	}
	// Null expects nil
	vv, err = checkSnmpTypeMatch(g.Null, nil)
	if err != nil || vv != nil {
		t.Fatalf("Null mismatch: %v %v", vv, err)
	}
}

func TestCheckSnmpTypeMatch_Errors(t *testing.T) {
	// wrong type for integer
	if _, err := checkSnmpTypeMatch(g.Integer, "x"); err == nil {
		t.Fatal("expected error for Integer type")
	}
	// wrong type for Counter32
	if _, err := checkSnmpTypeMatch(g.Counter32, "x"); err == nil {
		t.Fatal("expected error for Counter32 type")
	}
	// wrong type for Counter64
	if _, err := checkSnmpTypeMatch(g.Counter64, "x"); err == nil {
		t.Fatal("expected error for Counter64 type")
	}
	// wrong type for OctetString
	if _, err := checkSnmpTypeMatch(g.OctetString, 1.2); err == nil {
		t.Fatal("expected error for OctetString type")
	}
	// Null with non-nil
	if _, err := checkSnmpTypeMatch(g.Null, 0); err == nil {
		t.Fatal("expected error for Null non-nil")
	}
	// Opaque requires []byte - we pass string to force error
	if _, err := checkSnmpTypeMatch(g.Opaque, "bytes"); err == nil {
		t.Fatal("expected error for Opaque non-bytes")
	}
	// Unsupported type
	if _, err := checkSnmpTypeMatch(g.NsapAddress, 1); err == nil {
		t.Fatal("expected unsupported type error")
	}
}

func TestTrapPduRuleValidate(t *testing.T) {
	// missing fields
	r := TrapPduRule{}
	if err := r.Validate(); err == nil {
		t.Fatal("expected error for empty rule")
	}
	r = TrapPduRule{LabelName: "code"}
	if err := r.Validate(); err == nil {
		t.Fatal("expected error for empty oid")
	}
	r = TrapPduRule{Oid: "1.2.3"}
	if err := r.Validate(); err == nil {
		t.Fatal("expected error for empty label")
	}
	// valid
	r = TrapPduRule{LabelName: "code", Oid: "1.2.3"}
	if err := r.Validate(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestTrapPduRule_getPdu_DefaultAndTyped(t *testing.T) {
	// default (nil ValueType) numeric -> Counter64, string -> OctetString
	r := TrapPduRule{LabelName: "k", Oid: "1.2.3"}
	pdu, err := r.getPdu("123")
	if err != nil || pdu.Type != g.Counter64 {
		t.Fatalf("default numeric type unexpected: %v %v", pdu, err)
	}
	pdu, err = r.getPdu("abc")
	if err != nil || pdu.Type != g.OctetString {
		t.Fatalf("default string type unexpected: %v %v", pdu, err)
	}
	// explicit types
	for _, tc := range []struct {
		vt   string
		in   string
		want g.Asn1BER
	}{
		{ValueTypeOctetString, "x", g.OctetString},
		{ValueTypeObjectIdentifier, "1.2", g.ObjectIdentifier},
		{ValueTypeCounter32, "10", g.Counter32},
		{ValueTypeGauge32, "11", g.Gauge32},
		{ValueTypeCounter64, "12", g.Counter64},
		{ValueTypeUinteger32, "13", g.Uinteger32},
		{ValueTypeOpaqueFloat, "1.5", g.OpaqueFloat},
		{ValueTypeOpaqueDouble, "2.5", g.OpaqueDouble},
	} {
		vt := tc.vt
		r := TrapPduRule{LabelName: "k", Oid: "1.2.3", ValueType: &vt}
		pdu, err := r.getPdu(tc.in)
		if err != nil || pdu.Type != tc.want {
			t.Fatalf("getPdu %s -> %v err=%v", tc.vt, pdu, err)
		}
	}
	// parse error for numeric types
	vt := ValueTypeCounter32
	r = TrapPduRule{LabelName: "k", Oid: "1.2.3", ValueType: &vt}
	if _, err := r.getPdu("notnum"); err == nil {
		t.Fatal("expected parse error")
	}
}

func TestAppendPduValidate(t *testing.T) {
	// missing oid
	a := AppendPdu{ValueType: ValueTypeOctetString, Value: "x"}
	if err := a.Validate(); err == nil {
		t.Fatal("expected oid error")
	}
	// missing value type
	a = AppendPdu{Oid: "1.2.3"}
	if err := a.Validate(); err == nil {
		t.Fatal("expected value_type error")
	}
	// unknown value type
	a = AppendPdu{Oid: "1.2.3", ValueType: "unknown", Value: "x"}
	if err := a.Validate(); err == nil {
		t.Fatal("expected unknown type error")
	}
	// mismatch type: expect number for Counter64, got string
	a = AppendPdu{Oid: "1.2.3", ValueType: ValueTypeCounter64, Value: "x"}
	if err := a.Validate(); err == nil {
		t.Fatal("expected mismatch error")
	}
	// valid cases
	a = AppendPdu{Oid: "1.2.3", ValueType: ValueTypeOctetString, Value: "ok"}
	if err := a.Validate(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	a = AppendPdu{Oid: "1.2.3", ValueType: ValueTypeCounter32, Value: float64(1)}
	if err := a.Validate(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestParseEngineIDHex_EmptyAndBad(t *testing.T) {
	if b := parseEngineIDHex(""); b != nil {
		t.Fatal("empty should be nil")
	}
	if b := parseEngineIDHex("abc"); b != nil {
		t.Fatal("odd hex should be nil")
	}
}

// compile-time check to ensure imported errors isn't optimized out
var _ = errors.New
