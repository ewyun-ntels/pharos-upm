package adapters

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"ntels.com/pharos/shared/types/version"
)

func TestVersionAdapter_ToAPIVersionResponse(t *testing.T) {
	adapter := NewVersionAdapter()
	buildTime := time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)
	updatedAt := time.Date(2024, 1, 2, 12, 0, 0, 0, time.UTC)

	internal := Response{
		Module:    "core",
		Version:   "v1.0.0",
		Commit:    "abc123",
		BuildTime: buildTime,
		UpdatedAt: updatedAt,
	}

	expected := version.Types{
		Module:    "core",
		Version:   "v1.0.0",
		Commit:    "abc123",
		BuildTime: buildTime,
		UpdatedAt: updatedAt,
	}

	result := adapter.ToAPIVersionResponse(internal)
	assert.Equal(t, expected, result)
}
