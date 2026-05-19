// Code generated from JSON Schema using quicktype. DO NOT EDIT.
// To parse and unparse this JSON data, add this code to your project and do:
//
//    notificationRule, err := UnmarshalNotificationRule(bytes)
//    bytes, err = notificationRule.Marshal()
//
//    notificationType, err := UnmarshalNotificationType(bytes)
//    bytes, err = notificationType.Marshal()
//
//    queryNotificationRule, err := UnmarshalQueryNotificationRule(bytes)
//    bytes, err = queryNotificationRule.Marshal()

package notification

import "time"

import "encoding/json"

func UnmarshalNotificationRule(data []byte) (NotificationRule, error) {
	var r NotificationRule
	err := json.Unmarshal(data, &r)
	return r, err
}

func (r *NotificationRule) Marshal() ([]byte, error) {
	return json.Marshal(r)
}

func UnmarshalNotificationType(data []byte) (NotificationType, error) {
	var r NotificationType
	err := json.Unmarshal(data, &r)
	return r, err
}

func (r *NotificationType) Marshal() ([]byte, error) {
	return json.Marshal(r)
}

func UnmarshalQueryNotificationRule(data []byte) (QueryNotificationRule, error) {
	var r QueryNotificationRule
	err := json.Unmarshal(data, &r)
	return r, err
}

func (r *QueryNotificationRule) Marshal() ([]byte, error) {
	return json.Marshal(r)
}

// Schema for creating new notification rules (POST requests)
type NotificationRule struct {
	// Type of notification rule
	NotificationType NotificationType `json:"notification_type"`
	// Rule configuration object (varies by notification_type)
	Rule map[string]interface{} `json:"rule"`
	// Creation timestamp
	Timestamp *time.Time `json:"timestamp,omitempty"`
	// Last update timestamp
	UpdatedAt *time.Time `json:"updated_at,omitempty"`
}

// Schema for SNMP notification rules that send SNMP traps when alerts are triggered
type QueryNotificationRule struct {
	// Additional PDUs to append to SNMP traps
	AppendPdu []AppendPdu `json:"append_pdu,omitempty"`
	// Description of the notification rule
	Description *string `json:"description,omitempty"`
	// OID for the alert description
	DescriptionOID *string `json:"description_oid,omitempty"`
	// Filter to exclude specific label values
	ExcludeFilter map[string]interface{} `json:"exclude_filter,omitempty"`
	// Unique identifier for the notification rule
	ID *string `json:"id,omitempty"`
	// Name of the notification rule
	Name string `json:"name"`
	// SNMP port number
	Port *int64 `json:"port,omitempty"`
	// SNMPv2 configuration
	Snmpv2Config *Snmpv2Config `json:"snmpv2_config,omitempty"`
	// SNMPv3 configuration
	Snmpv3Config *Snmpv3Config `json:"snmpv3_config,omitempty"`
	// OID for the alert status
	StatusOID *string `json:"status_oid,omitempty"`
	// Target IP address or hostname
	Target string `json:"target"`
	// Transport protocol
	Transport *Transport `json:"transport,omitempty"`
	// Rules for mapping alert labels to SNMP PDUs
	TrapPduRule []TrapPduRule `json:"trap_pdu_rule,omitempty"`
	// OID for the alert value
	ValueOID *string `json:"value_oid,omitempty"`
	// SNMP version
	Version *int64 `json:"version,omitempty"`
}

type AppendPdu struct {
	// SNMP OID
	OID string `json:"oid"`
	// Value to send
	Value interface{} `json:"value"`
	// SNMP value type
	ValueType ValueType `json:"value_type"`
}

// SNMPv2 configuration
type Snmpv2Config struct {
	// SNMP community string
	Community string `json:"community"`
}

// SNMPv3 configuration
type Snmpv3Config struct {
	// Authentication passphrase
	AuthPassphrase *string `json:"auth_passphrase,omitempty"`
	// Authentication protocol
	AuthProtocol *AuthProtocol `json:"auth_protocol,omitempty"`
	// Authoritative engine ID in hex format
	AuthoritativeEngineID *string `json:"authoritative_engine_id,omitempty"`
	// Privacy passphrase
	PrivPassphrase *string `json:"priv_passphrase,omitempty"`
	// Privacy protocol
	PrivProtocol *PrivProtocol `json:"priv_protocol,omitempty"`
	// SNMPv3 username
	Username string `json:"username"`
}

type TrapPduRule struct {
	// Name of the alert label to map
	LabelName string `json:"label_name"`
	// SNMP OID for this label
	OID string `json:"oid"`
	// SNMP value type
	ValueType *ValueType `json:"value_type,omitempty"`
}

// Type of notification rule
type NotificationType string

const (
	Email   NotificationType = "email"
	SNMP    NotificationType = "snmp"
	Slack   NotificationType = "slack"
	Teams   NotificationType = "teams"
	Webhook NotificationType = "webhook"
)

// SNMP value type
type ValueType string

const (
	Counter32        ValueType = "counter32"
	Counter64        ValueType = "counter64"
	Gauge32          ValueType = "gauge32"
	ObjectIdentifier ValueType = "objectIdentifier"
	OctetString      ValueType = "OctetString"
	OpaqueDouble     ValueType = "opaqueDouble"
	OpaqueFloat      ValueType = "opaqueFloat"
	Uinteger32       ValueType = "uinteger32"
)

// Authentication protocol
type AuthProtocol string

const (
	Md5    AuthProtocol = "md5"
	SHA    AuthProtocol = "sha"
	Sha224 AuthProtocol = "sha224"
	Sha256 AuthProtocol = "sha256"
	Sha384 AuthProtocol = "sha384"
	Sha512 AuthProtocol = "sha512"
)

// Privacy protocol
type PrivProtocol string

const (
	AES    PrivProtocol = "aes"
	Aes192 PrivProtocol = "aes192"
	Aes256 PrivProtocol = "aes256"
	DES    PrivProtocol = "des"
)

// Transport protocol
type Transport string

const (
	TCP Transport = "tcp"
	UDP Transport = "udp"
)
