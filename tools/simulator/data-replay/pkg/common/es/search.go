package es

import "encoding/json"

type SearchRange struct {
	Gte *string `json:"gte,omitempty"`
	Lte *string `json:"lte,omitempty"`
	Gt  *string `json:"gt,omitempty"`
	Lt  *string `json:"lt,omitempty"`
}

type SearchExists struct {
	Field string `json:"field"`
}

type SearchQueryString struct {
	AnalyzeWildcard bool   `json:"analyze_wildcard"`
	Query           string `json:"query"`
}

type SearchFilter struct {
	Range       map[string]SearchRange `json:"range,omitempty"`
	Exists      *SearchExists          `json:"exists,omitempty"`
	QueryString *SearchQueryString     `json:"query_string,omitempty"`
	Bool        *SearchBool            `json:"bool,omitempty"`
}

type SearchBool struct {
	Filter []SearchFilter `json:"filter,omitempty"`
	Should []SearchFilter `json:"should,omitempty"`
}

type SearchQuery struct {
	Bool SearchBool `json:"bool"`
}

type SearchSort struct {
	Order string `json:"order"`
}

type Search struct {
	Size  *int                    `json:"size,omitempty"`
	Query SearchQuery             `json:"query,omitempty"`
	Sort  []map[string]SearchSort `json:"sort,omitempty"`
	Aggs  *json.RawMessage        `json:"aggs,omitempty"`
}
