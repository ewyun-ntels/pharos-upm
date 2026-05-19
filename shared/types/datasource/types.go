// Code generated from JSON Schema using quicktype. DO NOT EDIT.
// To parse and unparse this JSON data, add this code to your project and do:
//
//    types, err := UnmarshalTypes(bytes)
//    bytes, err = types.Marshal()

package datasource

import "encoding/json"

func UnmarshalTypes(data []byte) (Types, error) {
	var r Types
	err := json.Unmarshal(data, &r)
	return r, err
}

func (r *Types) Marshal() ([]byte, error) {
	return json.Marshal(r)
}

// Standard response schema for all datasource queries (database, external APIs, etc.)
type Types struct {
	// Result rows
	Data []map[string]interface{} `json:"data"`
	// Column metadata
	Meta []map[string]string `json:"meta"`
	// Number of rows returned
	Rows int64 `json:"rows"`
	// Query executed (SQL, API call, etc.)
	SQL        *string     `json:"sql,omitempty"`
	Statistics *Statistics `json:"statistics,omitempty"`
}

type Statistics struct {
	// Query execution time in seconds
	Elapsed *float64 `json:"elapsed,omitempty"`
}
