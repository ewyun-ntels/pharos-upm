package common

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestResponseCodeConstants(t *testing.T) {
	assert.Equal(t, "1", ResponseCodeSuccess)
	assert.Equal(t, "0", ResponseCodeFailure)
	assert.Equal(t, "-1", ResponseCodeError)
}

func TestResponseCodeConstants_Distinct(t *testing.T) {
	codes := []string{ResponseCodeSuccess, ResponseCodeFailure, ResponseCodeError}
	seen := make(map[string]bool)
	for _, c := range codes {
		assert.False(t, seen[c], "duplicate response code: %s", c)
		seen[c] = true
	}
}

func TestStbControlResponse_Fields(t *testing.T) {
	r := StbControlResponse{
		Code:    ResponseCodeSuccess,
		Message: "ok",
	}
	assert.Equal(t, ResponseCodeSuccess, r.Code)
	assert.Equal(t, "ok", r.Message)
}

func TestStbControlResponse_ZeroValue(t *testing.T) {
	var r StbControlResponse
	assert.Empty(t, r.Code)
	assert.Empty(t, r.Message)
}
