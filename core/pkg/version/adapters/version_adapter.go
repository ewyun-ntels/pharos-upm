package adapters

import (
	"time"

	"ntels.com/pharos/shared/types/version"
)

// Response represents internal version response (avoiding import cycle)
type Response struct {
	Module    string    `json:"module"`
	Version   string    `json:"version"`
	Commit    string    `json:"commit"`
	BuildTime time.Time `json:"build_time"`
	UpdatedAt time.Time `json:"updated_at"`
}

// VersionAdapter provides conversion between internal version types and API types
type VersionAdapter interface {
	// ToAPIVersionResponse converts internal Response to shared/types.VersionResponse
	ToAPIVersionResponse(internal Response) version.Types
}

// versionAdapter implements the VersionAdapter interface
type versionAdapter struct{}

// NewVersionAdapter creates a new instance of VersionAdapter
func NewVersionAdapter() VersionAdapter {
	return &versionAdapter{}
}

// ToAPIVersionResponse converts internal Response to shared/types.VersionResponse
func (a *versionAdapter) ToAPIVersionResponse(internal Response) version.Types {

	return version.Types{
		Module:    internal.Module,
		Version:   internal.Version,
		Commit:    internal.Commit,
		BuildTime: internal.BuildTime,
		UpdatedAt: internal.UpdatedAt,
	}
}
