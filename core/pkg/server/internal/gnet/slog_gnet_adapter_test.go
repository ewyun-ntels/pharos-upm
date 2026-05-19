package gnet

import (
	"bytes"
	"log/slog"
	"strings"
	"testing"
)

// assertContains는 output이 expected의 모든 문자열을 포함하는지 검증합니다.
func assertContains(t *testing.T, output string, expected []string) {
	t.Helper()
	for _, exp := range expected {
		if !strings.Contains(output, exp) {
			t.Errorf("Expected output to contain %q, got: %s", exp, output)
		}
	}
}

func TestSlogGnetAdapter(t *testing.T) {
	var buf bytes.Buffer

	logger := slog.New(slog.NewTextHandler(&buf, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}))

	adapter := NewSlogGnetAdapter(logger)

	tests := []struct {
		name     string
		logFunc  func()
		contains []string
	}{
		{
			name: "Infof",
			logFunc: func() {
				adapter.Infof("test info: %s", "message")
			},
			contains: []string{"level=INFO", "test info: message", "source=gnet"},
		},
		{
			name: "Debugf",
			logFunc: func() {
				adapter.Debugf("test debug: %d", 42)
			},
			contains: []string{"level=DEBUG", "test debug: 42", "source=gnet"},
		},
		{
			name: "Warnf",
			logFunc: func() {
				adapter.Warnf("test warning: %t", true)
			},
			contains: []string{"level=WARN", "test warning: true", "source=gnet"},
		},
		{
			name: "Errorf",
			logFunc: func() {
				adapter.Errorf("test error: %v", "failed")
			},
			contains: []string{"level=ERROR", "test error: failed", "source=gnet"},
		},
		{
			name: "MultipleFormatArgs",
			logFunc: func() {
				adapter.Infof("server=%s, port=%d, active=%t", "localhost", 8080, true)
			},
			contains: []string{"level=INFO", "server=localhost", "port=8080", "active=true", "source=gnet"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buf.Reset()
			tt.logFunc()

			output := buf.String()
			assertContains(t, output, tt.contains)
		})
	}
}

func TestSlogGnetAdapterFatalf(t *testing.T) {
	var buf bytes.Buffer

	logger := slog.New(slog.NewTextHandler(&buf, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}))

	adapter := NewSlogGnetAdapter(logger)

	defer func() {
		if r := recover(); r == nil {
			t.Error("Expected Fatalf to panic, but it didn't")
		} else {
			// panic 메시지 검증
			panicMsg, ok := r.(string)
			if !ok || !strings.Contains(panicMsg, "fatal error: crash") {
				t.Errorf("Expected panic message to contain 'fatal error: crash', got: %v", r)
			}

			// 로그 출력 검증
			output := buf.String()
			assertContains(t, output, []string{"level=ERROR", "fatal error", "level=fatal", "source=gnet"})
		}
	}()

	adapter.Fatalf("fatal error: %s", "crash")
}

// Benchmark tests
func BenchmarkSlogGnetAdapter_Infof(b *testing.B) {
	logger := slog.New(slog.NewTextHandler(&bytes.Buffer{}, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))
	adapter := NewSlogGnetAdapter(logger)

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		adapter.Infof("benchmark test: %d", i)
	}
}

func BenchmarkSlogGnetAdapter_Debugf(b *testing.B) {
	logger := slog.New(slog.NewTextHandler(&bytes.Buffer{}, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}))
	adapter := NewSlogGnetAdapter(logger)

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		adapter.Debugf("benchmark debug: %d", i)
	}
}

func BenchmarkSlogGnetAdapter_Errorf(b *testing.B) {
	logger := slog.New(slog.NewTextHandler(&bytes.Buffer{}, &slog.HandlerOptions{
		Level: slog.LevelError,
	}))
	adapter := NewSlogGnetAdapter(logger)

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		adapter.Errorf("benchmark error: %d", i)
	}
}
