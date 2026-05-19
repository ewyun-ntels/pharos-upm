package schedules

import (
	"bytes"
	"io"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSerFromBody_ImmediatelyOK(t *testing.T) {
	body := `{"name":"test","schedule_type":"immediately","area_type":"all","work_type":"stb_request_info","sos":[],"l3s":[],"cells":[],"settopboxes":[],"area_ids":[]}`
	var req ScheduleRequest
	err := req.serFromBody(io.NopCloser(strings.NewReader(body)))
	require.NoError(t, err)
	assert.Equal(t, "test", req.Name)
	assert.Equal(t, "immediately", req.ScheduleType)
	assert.Equal(t, "all", req.AreaType)
	assert.Equal(t, "stb_request_info", req.WorkType)
}

func TestSerFromBody_WithWorkValue(t *testing.T) {
	body := `{"name":"n","schedule_type":"immediately","area_type":"all","work_type":"limitAge","work_value":"18","sos":[],"l3s":[],"cells":[],"settopboxes":[],"area_ids":[]}`
	var req ScheduleRequest
	require.NoError(t, req.serFromBody(io.NopCloser(strings.NewReader(body))))
	require.NotNil(t, req.WorkValue)
	assert.Equal(t, "18", *req.WorkValue)
}

func TestSerFromBody_UnknownFieldReturnsError(t *testing.T) {
	body := `{"name":"n","unknown_field":"x"}`
	var req ScheduleRequest
	err := req.serFromBody(io.NopCloser(strings.NewReader(body)))
	assert.Error(t, err)
}

func TestSerFromBody_InvalidJSON(t *testing.T) {
	body := `{not json}`
	var req ScheduleRequest
	err := req.serFromBody(io.NopCloser(strings.NewReader(body)))
	assert.Error(t, err)
}

func TestSerFromBody_EmptyBody(t *testing.T) {
	var req ScheduleRequest
	err := req.serFromBody(io.NopCloser(strings.NewReader("")))
	assert.Error(t, err)
}

func TestSerFromBody_RespectsMaxBodySize(t *testing.T) {
	// MaxRequestBodySize+1 바이트 전송 시 LimitReader가 잤라내어 JSON 파싱 에러
	large := strings.Repeat("a", MaxRequestBodySize+1)
	body := `{"name":"` + large + `"}`
	var req ScheduleRequest
	err := req.serFromBody(io.NopCloser(bytes.NewReader([]byte(body))))
	assert.Error(t, err)
}

func TestSerFromBody_AreaIDs(t *testing.T) {
	body := `{"name":"n","schedule_type":"immediately","area_type":"so","area_ids":["SO-01","SO-02"],"work_type":"stb_request_info","sos":[],"l3s":[],"cells":[],"settopboxes":[]}`
	var req ScheduleRequest
	require.NoError(t, req.serFromBody(io.NopCloser(strings.NewReader(body))))
	assert.Equal(t, []string{"SO-01", "SO-02"}, req.AreaIDs)
	assert.Equal(t, "so", req.AreaType)
}
