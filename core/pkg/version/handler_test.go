package version

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"ntels.com/pharos/core/pkg/common"
	"ntels.com/pharos/shared/types/version"
)

func TestGetHandler_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Mock config (no database access needed for this test)
	config := common.Config{}

	// Create test context
	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)

	// Call handler
	handler := GetHandler(config)
	handler(c)

	// Verify response
	require.Equal(t, http.StatusOK, rec.Code)

	var result version.Types
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &result))

	// Verify basic fields are present
	assert.Equal(t, "core", result.Module)
	assert.Equal(t, "develop", result.Version) // Default version from versionresponse.go
	// Commit might be empty in development builds
	// BuildTime might be zero if not set during build
}

func TestGetHandler_ResponseStructure(t *testing.T) {
	gin.SetMode(gin.TestMode)

	config := common.Config{}
	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)

	handler := GetHandler(config)
	handler(c)

	require.Equal(t, http.StatusOK, rec.Code)

	// Verify response can be unmarshaled into shared/types/version.Types
	var result version.Types
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &result))

	// Verify response structure - module should always be present
	assert.Equal(t, "core", result.Module)
	assert.Equal(t, "develop", result.Version)
	// Commit might be empty in development builds
	// BuildTime and UpdatedAt are time.Time fields, so they exist even if zero
}
