// Code generated from JSON Schema using quicktype. DO NOT EDIT.
// To parse and unparse this JSON data, add this code to your project and do:
//
//    badgeSpec, err := UnmarshalBadgeSpec(bytes)
//    bytes, err = badgeSpec.Marshal()
//
//    badgeValueResponse, err := UnmarshalBadgeValueResponse(bytes)
//    bytes, err = badgeValueResponse.Marshal()

package badge

import "time"

import "encoding/json"

func UnmarshalBadgeSpec(data []byte) (BadgeSpec, error) {
	var r BadgeSpec
	err := json.Unmarshal(data, &r)
	return r, err
}

func (r *BadgeSpec) Marshal() ([]byte, error) {
	return json.Marshal(r)
}

func UnmarshalBadgeValueResponse(data []byte) (BadgeValueResponse, error) {
	var r BadgeValueResponse
	err := json.Unmarshal(data, &r)
	return r, err
}

func (r *BadgeValueResponse) Marshal() ([]byte, error) {
	return json.Marshal(r)
}

// Schema for badge definition requests
type BadgeSpec struct {
	// Cache time-to-live in seconds
	CacheTTL *int64 `json:"cache_ttl,omitempty"`
	// Column name to extract from query result
	Column string `json:"column"`
	// Name of the datasource to query
	Datasource string `json:"datasource"`
	// Unique name of the badge
	Name  string    `json:"name"`
	Query QuerySpec `json:"query"`
}

// Query specification with variables and metadata
type QuerySpec struct {
	// Template query string
	Query string `json:"query"`
	// Label for time column
	TimeLabel *string `json:"time_label,omitempty"`
	// Label for variable column
	VariableLabel *string `json:"variable_label,omitempty"`
	// Variables for query template
	Variables map[string]interface{} `json:"variables,omitempty"`
}

// Schema for badge value API responses
type BadgeValueResponse struct {
	// Whether the value was served from cache
	CacheHit *bool `json:"cache_hit,omitempty"`
	// Timestamp when value was last updated
	LastUpdated *time.Time `json:"last_updated,omitempty"`
	// Name of the badge
	Name string `json:"name"`
	// Extracted value from query result
	Value interface{} `json:"value"`
}
