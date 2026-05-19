package server

import (
	"bytes"
	"context"
	"log/slog"
	"strings"
	"testing"
)

// contextKey는 context에서 사용할 커스텀 키 타입입니다.
type contextKey string

const traceIDKey contextKey = "trace_id"

// assertContains는 output이 expected의 모든 문자열을 포함하는지 검증합니다.
func assertContains(t *testing.T, output string, expected []string) {
	t.Helper()
	for _, exp := range expected {
		if !strings.Contains(output, exp) {
			t.Errorf("Expected output to contain %q, got: %s", exp, output)
		}
	}
}

func TestSlogNatsAdapter(t *testing.T) {
	var buf bytes.Buffer

	logger := slog.New(slog.NewTextHandler(&buf, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}))

	adapter := NewSlogNatsAdapter(logger)

	tests := []struct {
		name     string
		logFunc  func()
		contains []string
	}{
		{
			name: "Noticef",
			logFunc: func() {
				adapter.Noticef("test notice: %s", "info")
			},
			contains: []string{"level=INFO", "test notice: info", "source=nats"},
		},
		{
			name: "Warnf",
			logFunc: func() {
				adapter.Warnf("test warning: %d", 123)
			},
			contains: []string{"level=WARN", "test warning: 123", "source=nats"},
		},
		{
			name: "Errorf",
			logFunc: func() {
				adapter.Errorf("test error: %v", "failed")
			},
			contains: []string{"level=ERROR", "test error: failed", "source=nats"},
		},
		{
			name: "Debugf",
			logFunc: func() {
				adapter.Debugf("test debug: %t", true)
			},
			contains: []string{"level=DEBUG", "test debug: true", "source=nats"},
		},
		{
			name: "Tracef",
			logFunc: func() {
				adapter.Tracef("test trace: %s", "detail")
			},
			contains: []string{"level=DEBUG", "test trace: detail", "source=nats", "level=trace"},
		},
		{
			name: "MultipleFormatArgs",
			logFunc: func() {
				adapter.Noticef("user=%s, id=%d, active=%t", "john", 123, true)
			},
			contains: []string{"level=INFO", "user=john", "id=123", "active=true", "source=nats"},
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

func TestSlogNatsAdapterFatalf(t *testing.T) {
	var buf bytes.Buffer

	logger := slog.New(slog.NewTextHandler(&buf, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}))

	adapter := NewSlogNatsAdapter(logger)

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
			assertContains(t, output, []string{"level=ERROR", "fatal error", "level=fatal", "source=nats"})
		}
	}()

	adapter.Fatalf("fatal error: %s", "crash")
}

func TestSlogNatsAdapterWithContext(t *testing.T) {
	var buf bytes.Buffer

	logger := slog.New(slog.NewTextHandler(&buf, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}))

	ctx := context.WithValue(context.Background(), traceIDKey, "12345")
	adapter := NewSlogNatsAdapterWithContext(logger, ctx)

	tests := []struct {
		name     string
		logFunc  func()
		contains []string
	}{
		{
			name: "NoticefWithContext",
			logFunc: func() {
				adapter.Noticef("context test: %s", "info")
			},
			contains: []string{"level=INFO", "context test: info", "source=nats"},
		},
		{
			name: "DebugfWithContext",
			logFunc: func() {
				adapter.Debugf("debug with context: %d", 42)
			},
			contains: []string{"level=DEBUG", "debug with context: 42", "source=nats"},
		},
		{
			name: "TracefWithContext",
			logFunc: func() {
				adapter.Tracef("trace with context: %s", "detail")
			},
			contains: []string{"level=DEBUG", "trace with context: detail", "source=nats", "level=trace"},
		},
		{
			name: "WarnfWithContext",
			logFunc: func() {
				adapter.Warnf("warning with context: %s", "alert")
			},
			contains: []string{"level=WARN", "warning with context: alert", "source=nats"},
		},
		{
			name: "ErrorfWithContext",
			logFunc: func() {
				adapter.Errorf("error with context: %s", "failure")
			},
			contains: []string{"level=ERROR", "error with context: failure", "source=nats"},
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

func TestSlogNatsAdapterWithContextFatalf(t *testing.T) {
	var buf bytes.Buffer

	logger := slog.New(slog.NewTextHandler(&buf, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}))

	ctx := context.Background()
	adapter := NewSlogNatsAdapterWithContext(logger, ctx)

	defer func() {
		if r := recover(); r == nil {
			t.Error("Expected Fatalf to panic, but it didn't")
		} else {
			panicMsg, ok := r.(string)
			if !ok || !strings.Contains(panicMsg, "fatal with context: crash") {
				t.Errorf("Expected panic message to contain 'fatal with context: crash', got: %v", r)
			}

			output := buf.String()
			assertContains(t, output, []string{"level=ERROR", "fatal with context", "level=fatal", "source=nats"})
		}
	}()

	adapter.Fatalf("fatal with context: %s", "crash")
}

func TestSlogNatsAdapterWithNilContext(t *testing.T) {
	var buf bytes.Buffer

	logger := slog.New(slog.NewTextHandler(&buf, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}))

	// nil context를 전달하면 자동으로 Background context가 사용되어야 함
	adapter := NewSlogNatsAdapterWithContext(logger, nil)

	adapter.Noticef("test with nil context")

	output := buf.String()
	assertContains(t, output, []string{"level=INFO", "test with nil context", "source=nats"})
}

// Benchmark tests
func BenchmarkSlogNatsAdapter_Noticef(b *testing.B) {
	logger := slog.New(slog.NewTextHandler(&bytes.Buffer{}, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))
	adapter := NewSlogNatsAdapter(logger)

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		adapter.Noticef("benchmark test: %d", i)
	}
}

func BenchmarkSlogNatsAdapter_Debugf(b *testing.B) {
	logger := slog.New(slog.NewTextHandler(&bytes.Buffer{}, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}))
	adapter := NewSlogNatsAdapter(logger)

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		adapter.Debugf("benchmark debug: %d", i)
	}
}

func BenchmarkSlogNatsAdapter_Errorf(b *testing.B) {
	logger := slog.New(slog.NewTextHandler(&bytes.Buffer{}, &slog.HandlerOptions{
		Level: slog.LevelError,
	}))
	adapter := NewSlogNatsAdapter(logger)

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		adapter.Errorf("benchmark error: %d", i)
	}
}

func BenchmarkSlogNatsAdapterWithContext_Noticef(b *testing.B) {
	logger := slog.New(slog.NewTextHandler(&bytes.Buffer{}, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))
	ctx := context.Background()
	adapter := NewSlogNatsAdapterWithContext(logger, ctx)

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		adapter.Noticef("benchmark test: %d", i)
	}
}

func BenchmarkSlogNatsAdapterWithContext_Debugf(b *testing.B) {
	logger := slog.New(slog.NewTextHandler(&bytes.Buffer{}, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}))
	ctx := context.Background()
	adapter := NewSlogNatsAdapterWithContext(logger, ctx)

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		adapter.Debugf("benchmark debug: %d", i)
	}
}
