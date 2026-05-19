package slogrotate

import (
	"bufio"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"ntels.com/pharos/core/pkg/common"
)

// captureOutput redirects os.Stdout to a pipe and returns a function to stop capture and the captured string.
func captureOutput() (func() string, error) {
	r, w, err := os.Pipe()
	if err != nil {
		return nil, err
	}
	orig := os.Stdout
	os.Stdout = w

	stop := func() string {
		_ = w.Close()
		os.Stdout = orig
		var b strings.Builder
		s := bufio.NewScanner(r)
		for s.Scan() {
			b.WriteString(s.Text())
		}
		_ = r.Close()
		return b.String()
	}
	return stop, nil
}

func TestSetupGlobal_NoConsole_WritesToFileNotStdout(t *testing.T) {
	tdir := t.TempDir()

	cfg := &common.Config{}
	cfg.Logger.FileName = "tarzan-test.log"
	cfg.Logger.LogPath = tdir
	cfg.Logger.IsConsole = false
	cfg.Logger.IsJson = false

	rot := New(cfg)
	logger, err := rot.SetupGlobal(&slog.HandlerOptions{Level: slog.LevelInfo}, cfg.Logger.IsJson, cfg.Logger.IsConsole)
	if err != nil || logger == nil {
		t.Fatalf("failed to setup global logger: %v", err)
	}

	stop, err := captureOutput()
	if err != nil {
		t.Fatalf("failed to capture stdout: %v", err)
	}

	slog.Info("hello-file-only", "k", "v")
	// Ensure write flush/rotation done
	time.Sleep(50 * time.Millisecond)

	captured := stop()
	if len(strings.TrimSpace(captured)) != 0 {
		t.Fatalf("expected no stdout output, got: %q", captured)
	}

	// Check log file content
	logFile := filepath.Join(tdir, "tarzan-test.log")
	f, err := os.Open(logFile)
	if err != nil {
		t.Fatalf("expected log file to exist: %v", err)
	}
	defer func() { _ = f.Close() }()
	content, _ := io.ReadAll(f)
	if !strings.Contains(string(content), "hello-file-only") {
		t.Fatalf("expected log file to contain message, got: %q", string(content))
	}
}
