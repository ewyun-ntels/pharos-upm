// Code generated from JSON Schema using quicktype. DO NOT EDIT.
// To parse and unparse this JSON data, add this code to your project and do:
//
//    types, err := UnmarshalTypes(bytes)
//    bytes, err = types.Marshal()

package version

import "time"

import "encoding/json"

func UnmarshalTypes(data []byte) (Types, error) {
	var r Types
	err := json.Unmarshal(data, &r)
	return r, err
}

func (r *Types) Marshal() ([]byte, error) {
	return json.Marshal(r)
}

// Schema for version information API response
type Types struct {
	// Build timestamp
	BuildTime time.Time `json:"build_time"`
	// Git commit hash
	Commit string `json:"commit"`
	// Module name
	Module string `json:"module"`
	// Last update timestamp
	UpdatedAt time.Time `json:"updated_at"`
	// Version string
	Version string `json:"version"`
}
