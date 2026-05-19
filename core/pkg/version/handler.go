package version

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"ntels.com/pharos/core/pkg/common"
	"ntels.com/pharos/core/pkg/version/adapters"
)

// parseTimeBestEffort tries multiple layouts and returns UTC time truncated to seconds.
func parseTimeBestEffort(s string) (time.Time, bool) {
	layouts := []string{time.RFC3339, time.RFC3339Nano, time.DateTime}
	for _, layout := range layouts {
		if t, err := time.Parse(layout, s); err == nil {
			return t.UTC().Truncate(time.Second), true
		}
	}
	return time.Time{}, false
}

func GetHandler(config common.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Prepare internal response with build defaults
		internalResp := adapters.Response{
			Module:  "core",
			Version: Version,
			Commit:  Commit,
		}
		if bt, ok := parseTimeBestEffort(BuildTime); ok {
			internalResp.BuildTime = bt
		}

		// Fetch via model; fall back on not found
		m := Model{Config: &config}
		if row, found, _ := m.GetCoreVersionRow(); found {
			if row.Module != "" {
				internalResp.Module = row.Module
			}
			if row.Version != "" {
				internalResp.Version = row.Version
			}
			if row.Commit.Valid {
				internalResp.Commit = row.Commit.String
			}
			if row.BuildTime.Valid {
				if bt, ok := parseTimeBestEffort(row.BuildTime.String); ok {
					internalResp.BuildTime = bt
				}
			}
			if row.UpdatedAt.Valid {
				internalResp.UpdatedAt = row.UpdatedAt.Time.UTC().Truncate(time.Second)
			}
		}

		// Convert to API response using adapter
		adapter := adapters.NewVersionAdapter()
		apiResp := adapter.ToAPIVersionResponse(internalResp)

		c.JSON(http.StatusOK, apiResp)
	}
}
