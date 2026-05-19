package types

import (
	"fmt"
	"strings"

	"github.com/gosnmp/gosnmp"
	"github.com/pkg/errors"
	"gopkg.in/yaml.v3"
)

type SNMP struct {
	Version   int    `yaml:"version,omitempty"`
	Community string `yaml:"community,omitempty"`

	SnmpVersion gosnmp.SnmpVersion
}

func (s *SNMP) Init() error {
	if s == nil {
		return errors.New("nil pointer")
	}

	switch s.Version {
	case 1:
		s.SnmpVersion = gosnmp.Version1
	case 2:
		s.SnmpVersion = gosnmp.Version2c
	case 3:
		s.SnmpVersion = gosnmp.Version3
	default:
		return errors.New(fmt.Sprintf("SNMP version must be 1, 2 or 3. Got: %d", s.Version))
	}
	return nil
}

type SnmpObject struct {
	Oid  string `yaml:"oid,omitempty"`
	Name string `yaml:"name,omitempty"`
}

type SnmpTable struct {
	Oid     string           `yaml:"oid,omitempty"`
	Name    string           `yaml:"name,omitempty"`
	Index   []SnmpTableIndex `yaml:"index,omitempty"`
	Items   map[int]string   `yaml:"items,omitempty"`
	Context []string         `yaml:"context,omitempty"`

	IndexLength   int
	ExternalIndex []IndexPosition
}

type plainSnmpTable SnmpTable

func (r *SnmpTable) UnmarshalYAML(value *yaml.Node) error {
	plain := (*plainSnmpTable)(r)
	err := value.Decode(plain)
	if err != nil {
		return err
	}

	r.IndexLength = 0
	for i := range r.Index {
		if r.Index[i].Name == "entry" {
			r.IndexLength += 1
			continue
		}

		if r.Index[i].Size <= 0 {
			r.Index[i].Size = 1
		}
		pos := IndexPosition{
			Name:  r.Index[i].Name,
			Start: r.IndexLength,
			End:   r.IndexLength + r.Index[i].Size,
		}

		r.IndexLength += r.Index[i].Size
		find := false
		for _, name := range r.Items {
			if r.Index[i].Name == name {
				find = true
				break
			}
		}

		if !find {
			r.ExternalIndex = append(r.ExternalIndex, pos)
		}
	}

	if r.IndexLength == 0 {
		return errors.New("invalid table config")
	}

	return nil
}

type SnmpTableIndex struct {
	Name string `yaml:"name,omitempty"`
	Size int    `yaml:"size,omitempty"`
}

type IndexPosition struct {
	Name  string
	Start int
	End   int
}

type SnmpResult map[string]interface{}

func SnmpTrimDot(dataUnit gosnmp.SnmpPDU) gosnmp.SnmpPDU {
	if dataUnit.Type == gosnmp.ObjectIdentifier {
		oid, ok := dataUnit.Value.(string)
		if ok {
			dataUnit.Value = strings.TrimPrefix(oid, ".")
		}
	}
	dataUnit.Name = strings.TrimPrefix(dataUnit.Name, ".")
	return dataUnit
}

func SnmpGetValue(dataUnit gosnmp.SnmpPDU, raw bool) interface{} {
	if raw {
		return dataUnit.Value
	}

	switch dataUnit.Type {
	case gosnmp.OctetString:
		return string(dataUnit.Value.([]byte))
	}

	return dataUnit.Value
}

func pduTypeToString(t gosnmp.PDUType) string {
	switch t {
	case gosnmp.Sequence:
		return "Sequence"
	case gosnmp.GetRequest:
		return "GetRequest"
	case gosnmp.GetNextRequest:
		return "GetNextRequest"
	case gosnmp.GetResponse:
		return "GetResponse"
	case gosnmp.SetRequest:
		return "SetRequest"
	case gosnmp.Trap:
		return "Trap"
	case gosnmp.GetBulkRequest:
		return "GetBulkRequest"
	case gosnmp.InformRequest:
		return "InformRequest"
	case gosnmp.SNMPv2Trap:
		return "SNMPv2Trap"
	case gosnmp.Report:
		return "Report"
	default:
		return "Unknown"
	}
}
