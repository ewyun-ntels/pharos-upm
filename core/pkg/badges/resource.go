package badges

import (
	"ntels.com/pharos/core/internal/query_builder"
)

// BadgeSpec is the payload for creating/updating a badge definition.
type BadgeSpec struct {
	Name       string                  `json:"name"`
	Datasource string                  `json:"datasource"`
	Query      query_builder.QuerySpec `json:"query"`
	Column     string                  `json:"column"` // the column name to extract from query result
}

// BadgeValueResponse is the GET response containing the extracted value.
type BadgeValueResponse struct {
	Name  string `json:"name"`
	Value any    `json:"value"`
}
