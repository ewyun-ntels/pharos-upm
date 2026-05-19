package command

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"net"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"ntels.com/pharos/core/pkg/common"
	control_common "ntels.com/pharos/extensions/catv/business/pkg/control/common"
	"ntels.com/pharos/extensions/catv/business/pkg/control/common/tables"
)

// newSettopboxConfig returns a minimal config for Settopbox tests.
func newSettopboxConfig() common.Config {
	var cfg common.Config
	cfg.Catv.Control.Timeout.Connection = "100ms"
	cfg.Catv.Control.Timeout.Send = "200ms"
	cfg.Catv.Control.Retry.PortExhaustion.Count = 1
	cfg.Catv.Control.Retry.PortExhaustion.Delay = "1ms"
	cfg.Catv.Control.Retry.Total.Count = 1
	cfg.Catv.Control.Retry.Total.Delay = "1ms"
	return cfg
}

// newSettopboxTLSConfig generates a self-signed TLS config for test listeners.
func newSettopboxTLSConfig(t *testing.T) *tls.Config {
	t.Helper()
	key, err := rsa.GenerateKey(rand.Reader, 1024)
	require.NoError(t, err)
	tmpl := &x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject:      pkix.Name{CommonName: "test"},
		NotBefore:    time.Now().Add(-time.Hour),
		NotAfter:     time.Now().Add(time.Hour),
	}
	certDER, err := x509.CreateCertificate(rand.Reader, tmpl, tmpl, &key.PublicKey, key)
	require.NoError(t, err)
	return &tls.Config{
		Certificates: []tls.Certificate{{
			Certificate: [][]byte{certDER},
			PrivateKey:  key,
		}},
	}
}

// startTLSEchoServer starts a TLS listener that drains reads until EOF then writes response.
func startTLSEchoServer(t *testing.T, response []byte) (host string, port int) {
	t.Helper()
	ln, err := tls.Listen("tcp", "127.0.0.1:0", newSettopboxTLSConfig(t))
	require.NoError(t, err)
	t.Cleanup(func() { ln.Close() })

	go func() {
		conn, err := ln.Accept()
		if err != nil {
			return
		}
		defer conn.Close()
		buf := make([]byte, 1024)
		for {
			if _, err := conn.Read(buf); err != nil {
				break
			}
		}
		_, _ = conn.Write(response)
	}()

	h, p, _ := net.SplitHostPort(ln.Addr().String())
	_, _ = fmt.Sscanf(p, "%d", &port)
	return h, port
}

// --- NewSettopbox ---

func TestNewSettopbox_DefaultTimeouts(t *testing.T) {
	cfg := newSettopboxConfig()
	s := NewSettopbox(cfg, context.Background(), tables.StbControlScheduleResultRaw{})
	assert.Equal(t, 100*time.Millisecond, s.connectionTimeout)
	assert.Equal(t, 200*time.Millisecond, s.sendTimeout)
	assert.Equal(t, time.Millisecond, s.portExhaustionRetryDelay)
	assert.Equal(t, time.Millisecond, s.totalRetryDelay)
	assert.Equal(t, 1, s.portExhaustionRetryCount)
	assert.Equal(t, 1, s.totalRetryCount)
}

func TestNewSettopbox_InvalidDurationsFallback(t *testing.T) {
	var cfg common.Config
	cfg.Catv.Control.Timeout.Connection = "bad"
	cfg.Catv.Control.Timeout.Send = "bad"
	cfg.Catv.Control.Retry.PortExhaustion.Delay = "bad"
	cfg.Catv.Control.Retry.Total.Delay = "bad"
	s := NewSettopbox(cfg, context.Background(), tables.StbControlScheduleResultRaw{})
	assert.Equal(t, 3*time.Second, s.connectionTimeout)
	assert.Equal(t, 5*time.Second, s.sendTimeout)
	assert.Equal(t, 20*time.Second, s.portExhaustionRetryDelay)
	assert.Equal(t, 5*time.Second, s.totalRetryDelay)
}

// --- Control: invalid work_type ---

func TestControl_InvalidWorkType(t *testing.T) {
	cfg := newSettopboxConfig()
	s := NewSettopbox(cfg, context.Background(), tables.StbControlScheduleResultRaw{WorkType: "invalid_type"})
	resp := s.Control(false)
	assert.Equal(t, control_common.ResponseCodeError, resp.Code)
	assert.Contains(t, resp.Message, "unsupported work_type")
}

// --- Control: ctx already cancelled ---

func TestControl_CtxCancelled(t *testing.T) {
	cfg := newSettopboxConfig()
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	s := NewSettopbox(cfg, ctx, tables.StbControlScheduleResultRaw{WorkType: WorkTypeSmartReboot})
	resp := s.Control(false)
	assert.Equal(t, control_common.ResponseCodeError, resp.Code)
	assert.Equal(t, "context cancelled", resp.Message)
}

// --- Control: connection refused (no listener) ---

func TestControl_ConnectionRefused(t *testing.T) {
	cfg := newSettopboxConfig()
	cfg.Catv.Control.Port = 19998 // nothing listening
	s := NewSettopbox(cfg, context.Background(), tables.StbControlScheduleResultRaw{
		WorkType:  WorkTypeSmartReboot,
		SrcIpAddr: "127.0.0.1",
	})
	resp := s.Control(false)
	assert.Equal(t, control_common.ResponseCodeError, resp.Code)
	assert.Contains(t, resp.Message, "failed to connect to")
}

// --- Control: simulation mode ---

func TestControl_SimulationMode_ValidCode(t *testing.T) {
	cfg := newSettopboxConfig()
	cfg.Catv.Control.Simulation = true
	validCodes := map[string]bool{
		control_common.ResponseCodeSuccess: true,
		control_common.ResponseCodeFailure: true,
		control_common.ResponseCodeError:   true,
	}
	for range 10 {
		s := NewSettopbox(cfg, context.Background(), tables.StbControlScheduleResultRaw{WorkType: WorkTypeSmartReboot})
		resp := s.Control(false)
		assert.True(t, validCodes[resp.Code], "unexpected code: %s", resp.Code)
	}
}

// --- Control: real TLS server, JSON response ---

func TestControl_Success_JSONResponse(t *testing.T) {
	host, port := startTLSEchoServer(t, []byte(`{"Code":"0","Message":"ok"}`))
	cfg := newSettopboxConfig()
	cfg.Catv.Control.Port = port
	s := NewSettopbox(cfg, context.Background(), tables.StbControlScheduleResultRaw{
		WorkType:  WorkTypeSmartReboot,
		SrcIpAddr: host,
	})
	resp := s.Control(false)
	assert.Equal(t, "0", resp.Code)
	assert.Equal(t, "ok", resp.Message)
}

// --- Control: real TLS server, stb_request_info returns raw ---

func TestControl_StbRequestInfo_ReturnsRaw(t *testing.T) {
	payload := `{"hostId":"abc","macAddr":"11:22:33:44:55:66"}`
	host, port := startTLSEchoServer(t, []byte(payload))
	cfg := newSettopboxConfig()
	cfg.Catv.Control.Port = port
	s := NewSettopbox(cfg, context.Background(), tables.StbControlScheduleResultRaw{
		WorkType:  WorkTypeStbRequestInfo,
		SrcIpAddr: host,
	})
	resp := s.Control(false)
	assert.Equal(t, control_common.ResponseCodeSuccess, resp.Code)
	assert.Equal(t, payload, resp.Message)
}

// --- Control: retry=false skips retry on retryable error ---

func TestControl_NoRetry_ReturnsImmediately(t *testing.T) {
	cfg := newSettopboxConfig()
	cfg.Catv.Control.Retry.Total.Count = 3
	cfg.Catv.Control.Port = 19997 // nothing listening → timeout error
	s := NewSettopbox(cfg, context.Background(), tables.StbControlScheduleResultRaw{
		WorkType:  WorkTypeSmartReboot,
		SrcIpAddr: "127.0.0.1",
	})
	start := time.Now()
	resp := s.Control(false) // retry=false: should not loop
	elapsed := time.Since(start)
	assert.Equal(t, control_common.ResponseCodeError, resp.Code)
	// With retry=false and connection refused (not retryable), must be fast
	assert.Less(t, elapsed, 2*time.Second)
}

// --- parseResponse ---

func TestParseResponse_EmptyResponse(t *testing.T) {
	s := NewSettopbox(newSettopboxConfig(), context.Background(), tables.StbControlScheduleResultRaw{})
	resp := s.parseResponse(WorkTypeSmartReboot, []byte{})
	assert.Equal(t, control_common.ResponseCodeError, resp.Code)
	assert.Contains(t, resp.Message, "no response received")
}

func TestParseResponse_SysCheck_RawString(t *testing.T) {
	s := NewSettopbox(newSettopboxConfig(), context.Background(), tables.StbControlScheduleResultRaw{})
	resp := s.parseResponse(WorkTypeSysCheck, []byte("raw-data"))
	assert.Equal(t, control_common.ResponseCodeError, resp.Code)
	assert.Contains(t, resp.Message, "failed to unmarshal")
}

func TestParseResponse_StbRequestInfo_RawString(t *testing.T) {
	s := NewSettopbox(newSettopboxConfig(), context.Background(), tables.StbControlScheduleResultRaw{})
	resp := s.parseResponse(WorkTypeStbRequestInfo, []byte("info-data"))
	assert.Equal(t, control_common.ResponseCodeError, resp.Code)
	assert.Contains(t, resp.Message, "failed to unmarshal")
}

func TestParseResponse_SysCheck_ValidJSON(t *testing.T) {
	s := NewSettopbox(newSettopboxConfig(), context.Background(), tables.StbControlScheduleResultRaw{})
	resp := s.parseResponse(WorkTypeSysCheck, []byte(`{"status":"ok","version":"1.0"}`))
	assert.Equal(t, control_common.ResponseCodeSuccess, resp.Code)
	var result map[string]string
	require.NoError(t, json.Unmarshal([]byte(resp.Message), &result))
	assert.Equal(t, "ok", result["status"])
	assert.Equal(t, "1.0", result["version"])
}

func TestParseResponse_StbRequestInfo_ValidJSON(t *testing.T) {
	s := NewSettopbox(newSettopboxConfig(), context.Background(), tables.StbControlScheduleResultRaw{})
	resp := s.parseResponse(WorkTypeStbRequestInfo, []byte(`{"loggingTime":"2024/06/15 09:00","runningTime":"1day 0hour 0min 0sec"}`))
	assert.Equal(t, control_common.ResponseCodeSuccess, resp.Code)
	var result map[string]string
	require.NoError(t, json.Unmarshal([]byte(resp.Message), &result))
	// KST → UTC 변환 필드 추가됨
	assert.NotEmpty(t, result["loggingTimeUtc"])
	// runningTimeSec 변환 필드 추가됨
	assert.Equal(t, "86400", result["runningTimeSec"])
}

func TestParseResponse_ValidJSON(t *testing.T) {
	s := NewSettopbox(newSettopboxConfig(), context.Background(), tables.StbControlScheduleResultRaw{})
	resp := s.parseResponse(WorkTypeSmartReboot, []byte(`{"Code":"0","Message":"ok"}`))
	assert.Equal(t, "0", resp.Code)
	assert.Equal(t, "ok", resp.Message)
}

func TestParseResponse_InvalidJSON(t *testing.T) {
	s := NewSettopbox(newSettopboxConfig(), context.Background(), tables.StbControlScheduleResultRaw{})
	resp := s.parseResponse(WorkTypeSmartReboot, []byte("not-json"))
	assert.Equal(t, control_common.ResponseCodeError, resp.Code)
	assert.Contains(t, resp.Message, "failed to unmarshal")
}

// --- isPortExhaustionError ---

func TestIsPortExhaustionError_Settopbox(t *testing.T) {
	s := NewSettopbox(newSettopboxConfig(), context.Background(), tables.StbControlScheduleResultRaw{})
	assert.False(t, s.isPortExhaustionError(nil))
	assert.True(t, s.isPortExhaustionError(errors.New("cannot assign requested address")))
	assert.True(t, s.isPortExhaustionError(errors.New("address already in use")))
	assert.True(t, s.isPortExhaustionError(errors.New("too many open files")))
	assert.False(t, s.isPortExhaustionError(errors.New("connection refused")))
}

// --- isRetryableError ---

func TestIsRetryableError_Settopbox(t *testing.T) {
	s := NewSettopbox(newSettopboxConfig(), context.Background(), tables.StbControlScheduleResultRaw{})
	tests := []struct {
		msg  string
		want bool
	}{
		{"", false},
		{"connection refused", false}, // NOT retryable: device is off
		{"i/o timeout", true},
		{"TIMEOUT", true},
		{"network unreachable", true},
		{"connection reset by peer", true},
		{"broken pipe", true},
		{"no route to host", true},
		{"some other error", false},
	}
	for _, tt := range tests {
		t.Run(tt.msg, func(t *testing.T) {
			assert.Equal(t, tt.want, s.isRetryableError(tt.msg))
		})
	}
}

// --- getConnectionAddress ---

func TestGetConnectionAddress(t *testing.T) {
	cfg := newSettopboxConfig()
	cfg.Catv.Control.Port = 8801
	s := NewSettopbox(cfg, context.Background(), tables.StbControlScheduleResultRaw{SrcIpAddr: "10.0.0.1"})
	assert.Equal(t, "10.0.0.1:8801", s.getConnectionAddress())
}

// --- pickRandom ---

func TestPickRandom_Settopbox_Empty(t *testing.T) {
	assert.Equal(t, "", pickRandom(nil))
	assert.Equal(t, "", pickRandom([]string{}))
}

func TestPickRandom_Settopbox_Single(t *testing.T) {
	assert.Equal(t, "only", pickRandom([]string{"only"}))
}

func TestPickRandom_Settopbox_ReturnsFromSlice(t *testing.T) {
	choices := []string{"a", "b", "c"}
	set := map[string]bool{"a": true, "b": true, "c": true}
	for range 20 {
		assert.True(t, set[pickRandom(choices)])
	}
}
