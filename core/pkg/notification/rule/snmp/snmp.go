package snmp

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strconv"
	"strings"
	"time"

	"github.com/go-viper/mapstructure/v2"
	g "github.com/gosnmp/gosnmp"
	"ntels.com/pharos/core/pkg/common"
	notification_common "ntels.com/pharos/core/pkg/notification/common"
	rule_common "ntels.com/pharos/core/pkg/notification/rule/common"
)

const ValueTypeOctetString = "OctetString"
const ValueTypeCounter32 = "counter32"
const ValueTypeGauge32 = "gauge32"
const ValueTypeCounter64 = "counter64" // default
const ValueTypeUinteger32 = "uinteger32"
const ValueTypeOpaqueFloat = "opaqueFloat"
const ValueTypeOpaqueDouble = "opaqueDouble"
const ValueTypeObjectIdentifier = "objectIdentifier"

// sendTrapFunc allows tests to intercept SendTrap without real network
var sendTrapFunc = func(c *g.GoSNMP, trap g.SnmpTrap) (*g.SnmpPacket, error) { return c.SendTrap(trap) }

// parseEngineIDHex converts various hex formats (e.g., "8000137003", "0x8000137003",
// or "80:00:13:70:03") into the raw bytes expected for SNMPv3 AuthoritativeEngineID.
// Returns nil for empty input.
func parseEngineIDHex(s string) []byte {
	if strings.TrimSpace(s) == "" {
		return nil
	}
	// Remove common separators and optional 0x prefix
	n := strings.TrimSpace(strings.ToLower(s))
	n = strings.TrimPrefix(n, "0x")
	n = strings.ReplaceAll(n, ":", "")
	n = strings.ReplaceAll(n, " ", "")
	// If odd length, it's invalid hex; return nil to let gosnmp handle default
	if len(n)%2 != 0 {
		return nil
	}
	out := make([]byte, 0, len(n)/2)
	for i := 0; i < len(n); i += 2 {
		var b byte
		_, err := fmt.Sscanf(n[i:i+2], "%02x", &b)
		if err != nil {
			return nil
		}
		out = append(out, b)
	}
	return out
}

func getAsnBER(name string) (g.Asn1BER, error) {
	switch name {
	case ValueTypeOctetString:
		return g.OctetString, nil
	case ValueTypeCounter32:
		return g.Counter32, nil
	case ValueTypeGauge32:
		return g.Gauge32, nil
	case ValueTypeCounter64:
		return g.Counter64, nil
	case ValueTypeUinteger32:
		return g.Uinteger32, nil
	case ValueTypeOpaqueFloat:
		return g.OpaqueFloat, nil
	case ValueTypeOpaqueDouble:
		return g.OpaqueDouble, nil
	case ValueTypeObjectIdentifier:
		return g.ObjectIdentifier, nil
	}

	return 0, errors.New("unknown Asn1BER value type")
}

func getAuthProtocol(protocol string) g.SnmpV3AuthProtocol {
	switch strings.ToLower(protocol) {
	case "sha":
		return g.SHA
	case "sha224":
		return g.SHA224
	case "sha256":
		return g.SHA256
	case "sha384":
		return g.SHA384
	case "sha512":
		return g.SHA512
	case "md5":
		return g.MD5
	default:
		return g.NoAuth
	}
}

func getPrivProtocol(protocol string) g.SnmpV3PrivProtocol {
	switch strings.ToLower(protocol) {
	case "des":
		return g.DES
	case "aes":
		return g.AES
	case "aes192":
		return g.AES192
	case "aes256":
		return g.AES256
	default:
		return g.NoPriv
	}
}

// checkSnmpTypeMatch snmp 의 type 을 확인 float64, string 여부만 판단
//
//	json으로 파싱할경우 무조건 float64만 확인 하기 때문에 float64만 검사 차후 필요시 수정 필요
func checkSnmpTypeMatch(t g.Asn1BER, v any) (any, error) {
	// json 파싱결과는 float으로만 지정됨
	switch t {
	case g.Integer:
		if _, ok := v.(float64); !ok {
			return nil, errors.New("expected number (Integer)")
		}
		return int(v.(float64)), nil
	case g.Counter32, g.Gauge32, g.TimeTicks:
		if _, ok := v.(float64); !ok {
			return nil, errors.New("expected number (uint32-compatible)")
		}
		return uint32(v.(float64)), nil
	case g.Counter64:
		if _, ok := v.(float64); !ok {
			return nil, errors.New("expected number (uint64-compatible)")
		}
		return uint64(v.(float64)), nil
	case g.OctetString, g.ObjectIdentifier, g.IPAddress:
		if _, ok := v.(string); !ok {
			return nil, errors.New("expected string value")
		}
		return v.(string), nil
	case g.Null:
		if v != nil {
			return nil, errors.New("expected nil for Null type")
		}
		return nil, nil
	case g.Opaque:
		// JSON에서는 []byte 불가능하므로 필요시 base64로 다루는 게 일반적
		if _, ok := v.([]byte); !ok {
			return nil, errors.New("expected byte slice (Opaque)")
		}
		return v.([]byte), nil
	default:
		return nil, fmt.Errorf("unsupported SNMP type: %v", t)
	}
}

type AppendPdu struct {
	Oid       string `mapstructure:"oid" json:"oid"`
	ValueType string `mapstructure:"value_type" json:"value_type"`
	Value     any    `mapstructure:"value" json:"value"`
}

func (a *AppendPdu) Validate() error {
	if a.Oid == "" {
		return errors.New("append pdu oid is required")
	}

	if a.ValueType == "" {
		return errors.New("append pdu value type is required")
	}

	ber, err := getAsnBER(a.ValueType)
	if err != nil {
		return errors.Join(err, errors.New("unknown append-pdu value-type"))
	}

	_, err = checkSnmpTypeMatch(ber, a.Value)
	if err != nil {
		return err
	}

	return nil
}

type TrapPduRule struct {
	LabelName string  `mapstructure:"label_name" json:"label_name"`
	Oid       string  `mapstructure:"oid" json:"oid"`
	ValueType *string `mapstructure:"value_type" json:"value_type"`
}

func (s *TrapPduRule) Validate() error {
	if s.Oid == "" {
		return errors.New("trap-pdu-rule oid is empty")
	}

	if s.LabelName == "" {
		return errors.New("tarp-pdu-rule label_name is empty")
	}

	return nil
}

func (s *TrapPduRule) getPdu(v string) (*g.SnmpPDU, error) {
	pdu := &g.SnmpPDU{
		Name: s.Oid,
	}

	if s.ValueType == nil || *s.ValueType == "" {
		if val, err := strconv.ParseUint(v, 10, 64); err == nil {
			pdu.Type = g.Counter64
			pdu.Value = val
		} else {
			pdu.Type = g.OctetString
			pdu.Value = v
		}
		return pdu, nil
	}

	switch *s.ValueType {
	case ValueTypeOctetString:
		pdu.Type = g.OctetString
		pdu.Value = v
	case ValueTypeObjectIdentifier:
		pdu.Type = g.ObjectIdentifier
		pdu.Value = v
	case ValueTypeCounter32:
		val, err := strconv.ParseUint(v, 10, 32)
		if err != nil {
			return nil, err
		}
		pdu.Type = g.Counter32
		pdu.Value = uint32(val)
	case ValueTypeGauge32:
		val, err := strconv.ParseUint(v, 10, 32)
		if err != nil {
			return nil, err
		}
		pdu.Type = g.Gauge32
		pdu.Value = uint32(val)
	case ValueTypeCounter64:
		val, err := strconv.ParseUint(v, 10, 64)
		if err != nil {
			return nil, err
		}
		pdu.Type = g.Counter64
		pdu.Value = val
	case ValueTypeUinteger32:
		val, err := strconv.ParseUint(v, 10, 32)
		if err != nil {
			return nil, err
		}
		pdu.Type = g.Uinteger32
		pdu.Value = uint32(val)
	case ValueTypeOpaqueFloat:
		val, err := strconv.ParseFloat(v, 32)
		if err != nil {
			return nil, err
		}
		pdu.Type = g.OpaqueFloat
		pdu.Value = float32(val)
	case ValueTypeOpaqueDouble:
		val, err := strconv.ParseFloat(v, 64)
		if err != nil {
			return nil, err
		}
		pdu.Type = g.OpaqueDouble
		pdu.Value = val
	default:
		return nil, fmt.Errorf("unknown value type: %s", *s.ValueType)
	}

	return pdu, nil
}

type SNMPv2Config struct {
	Community string `mapstructure:"community,omitempty" json:"community"`
}

type SNMPv3Config struct {
	//AuthoritativeEngineID string `mapstructure:"authoritative_engine_id,omitempty" json:"authoritative_engine_id"`
	AuthoritativeEngineIDHex string `mapstructure:"authoritative_engine_id,omitempty" json:"authoritative_engine_id"`
	Username                 string `mapstructure:"username,omitempty" json:"username"`
	AuthProtocol             string `mapstructure:"auth_protocol,omitempty" json:"auth_protocol"`
	AuthPassphrase           string `mapstructure:"auth_passphrase,omitempty" json:"auth_passphrase"`
	PrivProtocol             string `mapstructure:"priv_protocol,omitempty" json:"priv_protocol"`
	PrivPassphrase           string `mapstructure:"priv_passphrase,omitempty" json:"priv_passphrase"`
}

type SNMP struct {
	Id             string         `mapstructure:"id,omitempty" json:"id"`
	Name           string         `mapstructure:"name" json:"name"`
	Description    string         `mapstructure:"description" json:"description,omitempty"`
	ValueOid       *string        `mapstructure:"value_oid" json:"value_oid,omitempty"`
	DescriptionOid *string        `mapstructure:"description_oid,omitempty" json:"description_oid,omitempty"`
	StatusOid      *string        `mapstructure:"status_oid,omitempty" json:"status_oid,omitempty"`
	ExcludeFilter  map[string]any `mapstructure:"exclude_filter" json:"exclude_filter,omitempty"`
	TrapPduRule    []TrapPduRule  `mapstructure:"trap_pdu_rule,omitempty" json:"trap_pdu_rule,omitempty"`
	Port           uint16         `mapstructure:"port,omitempty" json:"port,omitempty"`
	Target         string         `mapstructure:"target,omitempty" json:"target,omitempty"`
	Version        uint8          `mapstructure:"version,omitempty" json:"version,omitempty"`
	Transport      string         `mapstructure:"transport,omitempty" json:"transport,omitempty"`
	SNMPv2Config   *SNMPv2Config  `mapstructure:"snmpv2_config,omitempty" json:"snmpv2_config,omitempty"`
	SNMPv3Config   *SNMPv3Config  `mapstructure:"snmpv3_config,omitempty" json:"snmpv3_config,omitempty"`
	AppendPdu      []AppendPdu    `mapstructure:"append_pdu,omitempty" json:"append_pdu,omitempty"`
	trapPduRuleMap map[string]TrapPduRule
	startTimestamp time.Time
	connector      *g.GoSNMP
	config         common.Config
}

func (s *SNMP) GetId() string {
	return s.Id
}

func (s *SNMP) SetId(id string) {
	s.Id = id
}

func (s *SNMP) GetName() string {
	return s.Name
}

func (s *SNMP) Validate() error {
	if s.Name == "" {
		return errors.New("name is required")
	}

	if s.Transport != "udp" && s.Transport != "tcp" && s.Transport != "" {
		return fmt.Errorf("invalid transport type: %s", s.Transport)
	}

	labelNames := make(map[string]struct{})
	oids := make(map[string]struct{})

	for _, rule := range s.TrapPduRule {
		if _, exists := labelNames[rule.LabelName]; exists {
			return fmt.Errorf("duplicate label name: %s", rule.LabelName)
		}
		labelNames[rule.LabelName] = struct{}{}

		if _, exists := oids[rule.Oid]; exists {
			return fmt.Errorf("duplicate OID: %s", rule.Oid)
		}
		oids[rule.Oid] = struct{}{}

		if err := rule.Validate(); err != nil {
			return err
		}
	}

	for _, rule := range s.AppendPdu {
		if err := rule.Validate(); err != nil {
			return err
		}
	}

	return nil
}

func (s *SNMP) Load(config common.Config, data map[string]any) error {
	s.config = config

	err := mapstructure.Decode(data, s)
	if err != nil {
		slog.Error("rule load error", "error", err)
		return err
	}

	if s.Transport == "" {
		s.Transport = "udp"
	}

	s.trapPduRuleMap = make(map[string]TrapPduRule)
	for _, rule := range s.TrapPduRule {
		s.trapPduRuleMap[rule.LabelName] = rule
	}

	switch g.SnmpVersion(s.Version) {
	case g.Version2c:
		if s.SNMPv2Config == nil {
			return errors.New("SNMPv2Config is nil")
		}

		s.connector = &g.GoSNMP{
			Port:               s.Port,
			Target:             s.Target,
			Transport:          s.Transport,
			Community:          s.SNMPv2Config.Community,
			Version:            g.Version2c,
			Timeout:            time.Duration(1) * time.Second,
			Retries:            3,
			ExponentialTimeout: true,
			MaxOids:            g.MaxOids,
		}
	case g.Version3:
		if s.SNMPv3Config == nil {
			return errors.New("SNMPv3Config is nil")
		}

		s.connector = &g.GoSNMP{
			Port:               s.Port,
			Target:             s.Target,
			Transport:          s.Transport,
			Version:            g.SnmpVersion(s.Version),
			Timeout:            time.Duration(1) * time.Second,
			Retries:            3,
			ExponentialTimeout: true,
			MaxOids:            g.MaxOids,
			SecurityModel:      g.UserSecurityModel,
			MsgFlags:           g.AuthPriv, // AuthPriv 모드 사용 (인증 및 암호화)
			SecurityParameters: &g.UsmSecurityParameters{
				UserName:                 s.SNMPv3Config.Username,
				AuthoritativeEngineID:    string(parseEngineIDHex(s.SNMPv3Config.AuthoritativeEngineIDHex)),
				AuthenticationProtocol:   getAuthProtocol(s.SNMPv3Config.AuthProtocol),
				AuthenticationPassphrase: s.SNMPv3Config.AuthPassphrase,
				PrivacyProtocol:          getPrivProtocol(s.SNMPv3Config.PrivProtocol),
				PrivacyPassphrase:        s.SNMPv3Config.PrivPassphrase,
			},
		}
	default:
		return fmt.Errorf("unsupported snmp version: %d", g.SnmpVersion(s.Version))
	}

	err = s.connector.Connect()
	if err != nil {
		slog.Error("snmp connect failed", "error", err)
		return err
	}
	s.startTimestamp = time.Now()

	return nil
}

func (s *SNMP) Send(value *notification_common.AlertValue) error {
	if value == nil {
		return errors.New("value is nil")
	}

	var pdu []g.SnmpPDU
	for k, v := range value.Labels {
		if s.ExcludeFilter[k] == v {
			continue
		}

		if pduRule, exists := s.trapPduRuleMap[k]; exists {
			pduValue, err := pduRule.getPdu(v)
			if err != nil {
				slog.Error("pdu get error", "error", err)
				return err
			}
			pdu = append(pdu, *pduValue)
		}
	}

	if s.DescriptionOid != nil {
		pdu = append(pdu, g.SnmpPDU{
			Name:  *s.DescriptionOid,
			Type:  g.OctetString,
			Value: value.Description,
		})
	}

	if s.ValueOid != nil {
		pdu = append(pdu, g.SnmpPDU{
			Name:  *s.ValueOid,
			Type:  g.OpaqueFloat,
			Value: value.Value,
		})
	}

	if s.StatusOid != nil {
		pdu = append(pdu, g.SnmpPDU{
			Name:  *s.StatusOid,
			Type:  g.OctetString,
			Value: value.Status,
		})
	}

	for _, appendPdu := range s.AppendPdu {
		ber, err := getAsnBER(appendPdu.ValueType)
		if err != nil {
			slog.Error("pdu append error", "error", err)
			continue
		}

		match, err := checkSnmpTypeMatch(ber, appendPdu.Value)
		if err != nil {
			slog.Error("pdu append error(value type mismatch)", "error", err)
			continue
		}
		pdu = append(pdu, g.SnmpPDU{
			Name:  appendPdu.Oid,
			Type:  ber,
			Value: match,
		})
	}

	trap := g.SnmpTrap{
		Variables: pdu,
	}

	_, err := sendTrapFunc(s.connector, trap)
	if err != nil {
		return err
	}

	h := rule_common.History{
		Config: s.config,
	}
	err = h.Send([]notification_common.AlertValue{*value})
	if err != nil {
		slog.Error("alert event history send failed", "error", err)
	}

	return nil
}

func (s *SNMP) Run(_ context.Context) error {
	return nil
}

func (s *SNMP) Destroy() error {
	if s.connector != nil {
		_ = s.connector.Conn.Close()
	}

	return nil
}
