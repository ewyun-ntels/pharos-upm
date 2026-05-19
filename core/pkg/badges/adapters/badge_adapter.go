package adapters

import (
	"maps"
	"time"

	"ntels.com/pharos/core/internal/query_builder"
	"ntels.com/pharos/shared/types/badge"
)

// BadgeSpec represents a badge specification (avoiding import cycle)
type BadgeSpec struct {
	Name       string                  `json:"name"`
	Datasource string                  `json:"datasource"`
	Query      query_builder.QuerySpec `json:"query"`
	Column     string                  `json:"column"`
}

// BadgeValueResponse represents a badge value response (avoiding import cycle)
type BadgeValueResponse struct {
	Name  string `json:"name"`
	Value any    `json:"value"`
}

// BadgeAdapter provides conversion between internal badge types and API types
type BadgeAdapter interface {
	// FromAPIBadgeSpec converts badge.BadgeSpec to internal BadgeSpec
	FromAPIBadgeSpec(api badge.BadgeSpec) BadgeSpec

	// ToAPIBadgeValueResponse converts internal BadgeValueResponse to badge.BadgeValueResponse
	ToAPIBadgeValueResponse(internal BadgeValueResponse, cacheHit bool) badge.BadgeValueResponse

	// FromAPIQuerySpec converts shared/types.QuerySpec to internal query_builder.QuerySpec
	FromAPIQuerySpec(api badge.QuerySpec) query_builder.QuerySpec

	// ToAPIQuerySpec converts internal query_builder.QuerySpec to shared/types.QuerySpec
	ToAPIQuerySpec(internal query_builder.QuerySpec) badge.QuerySpec
}

// badgeAdapter implements the BadgeAdapter interface
type badgeAdapter struct{}

// NewBadgeAdapter creates a new instance of BadgeAdapter
func NewBadgeAdapter() BadgeAdapter {
	return &badgeAdapter{}
}

// FromAPIBadgeSpec converts shared/types.BadgeSpec to internal BadgeSpec
func (a *badgeAdapter) FromAPIBadgeSpec(api badge.BadgeSpec) BadgeSpec {
	return BadgeSpec{
		Name:       api.Name,
		Datasource: api.Datasource,
		Query:      a.FromAPIQuerySpec(api.Query),
		Column:     api.Column,
	}
}

// ToAPIBadgeValueResponse converts internal BadgeValueResponse to shared/types.BadgeValueResponse
func (a *badgeAdapter) ToAPIBadgeValueResponse(internal BadgeValueResponse, cacheHit bool) badge.BadgeValueResponse {
	now := time.Now()

	return badge.BadgeValueResponse{
		Name:        internal.Name,
		Value:       internal.Value,
		LastUpdated: &now,
		CacheHit:    &cacheHit,
	}
}

// FromAPIQuerySpec converts shared/types.QuerySpec to internal query_builder.QuerySpec
func (a *badgeAdapter) FromAPIQuerySpec(api badge.QuerySpec) query_builder.QuerySpec {
	spec := query_builder.QuerySpec{
		Query:     api.Query,
		Variables: make(map[string]any),
	}

	// Convert variables
	if api.Variables != nil {
		maps.Copy(spec.Variables, api.Variables)
	}

	// Handle optional fields
	if api.TimeLabel != nil {
		spec.TimeLabel = *api.TimeLabel
	}

	if api.VariableLabel != nil {
		spec.VariableLabel = *api.VariableLabel
	}

	return spec
}

// ToAPIQuerySpec converts internal query_builder.QuerySpec to shared/types.QuerySpec
func (a *badgeAdapter) ToAPIQuerySpec(internal query_builder.QuerySpec) badge.QuerySpec {
	spec := badge.QuerySpec{
		Query:     internal.Query,
		Variables: make(map[string]any),
	}

	// Convert variables
	if internal.Variables != nil {
		maps.Copy(spec.Variables, internal.Variables)
	}

	// Handle optional fields
	if internal.TimeLabel != "" {
		spec.TimeLabel = &internal.TimeLabel
	}

	if internal.VariableLabel != "" {
		spec.VariableLabel = &internal.VariableLabel
	}

	return spec
}
