//go:build !windows

package daemon

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"syscall"

	"github.com/sevlyar/go-daemon"
	"ntels.com/pharos/core/pkg/common"
)

const (
	DefaultPidFile = "/var/run/pharos.pid"
	DefaultLogFile = "/var/log/pharos.log"
)

// HealthChecker interface for health monitoring
type HealthChecker interface {
	HealthCheck(ctx context.Context) error
}

// ProcessManager manages daemon lifecycle with health monitoring
// Currently unused but kept for future health monitoring features
type ProcessManager struct {
	// These fields are reserved for future health monitoring implementation
	_ string
}

// DaemonizeWithConfig runs the current process as a daemon with config support
func DaemonizeWithConfig(pidFile, logFile string, daemonConfig common.DaemonConfig) error {
	// Priority: CLI flags > Config > Defaults
	if pidFile == "" && daemonConfig.PidFile != "" {
		pidFile = daemonConfig.PidFile
	}
	if pidFile == "" {
		pidFile = DefaultPidFile
	}

	if logFile == "" && daemonConfig.LogFile != "" {
		logFile = daemonConfig.LogFile
	}
	if logFile == "" {
		logFile = DefaultLogFile
	}

	return Daemonize(pidFile, logFile)
}

// Daemonize runs the current process as a daemon using sevlyar/go-daemon
func Daemonize(pidFile, logFile string) error {
	if pidFile == "" {
		pidFile = DefaultPidFile
	}
	if logFile == "" {
		logFile = DefaultLogFile
	}

	// Test write permissions before daemonizing
	if err := testWritePermission(pidFile); err != nil {
		return fmt.Errorf("insufficient permissions for PID file %s: %w", pidFile, err)
	}
	if err := testWritePermission(logFile); err != nil {
		return fmt.Errorf("insufficient permissions for log file %s: %w", logFile, err)
	}

	// Create daemon context
	cntxt := &daemon.Context{
		PidFileName: pidFile,
		PidFilePerm: 0644,
		LogFileName: logFile,
		LogFilePerm: 0640,
		WorkDir:     "./", // Keep current working directory for relative paths
		Umask:       027,
		Args:        os.Args,
	}

	// Reborn creates child process
	child, err := cntxt.Reborn()
	if err != nil {
		return fmt.Errorf("failed to create daemon: %w", err)
	}

	if child != nil {
		// Parent process - child was successfully created
		slog.Info("daemon process started", "pid", child.Pid)
		return nil
	}

	// Child process - we are now the daemon
	// The daemon will continue execution after this function returns
	defer func() {
		if err := cntxt.Release(); err != nil {
			slog.Error("failed to release daemon context", "error", err)
		}
	}()

	slog.Info("daemon process initialized", "pid", os.Getpid())
	return nil
}

// testWritePermission tests if we can write to the given file path
func testWritePermission(filePath string) error {
	dir := filepath.Dir(filePath)

	// Check if directory exists and is writable
	if info, err := os.Stat(dir); err != nil {
		if os.IsNotExist(err) {
			// Try to create the directory
			if err := os.MkdirAll(dir, 0755); err != nil {
				return fmt.Errorf("cannot create directory %s: %w", dir, err)
			}
		} else {
			return fmt.Errorf("cannot access directory %s: %w", dir, err)
		}
	} else if !info.IsDir() {
		return fmt.Errorf("%s is not a directory", dir)
	}

	// Test write permission by creating a temporary file
	tempFile := filepath.Join(dir, ".pharos_write_test")
	if err := os.WriteFile(tempFile, []byte("test"), 0644); err != nil {
		return fmt.Errorf("cannot write to directory %s: %w", dir, err)
	}

	// Clean up test file
	_ = os.Remove(tempFile)
	return nil
}

// RemovePidFile removes the PID file
func RemovePidFile(pidFile string) error {
	if pidFile == "" {
		pidFile = DefaultPidFile
	}

	if err := os.Remove(pidFile); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to remove PID file: %w", err)
	}

	return nil
}

// IsRunning checks if a process with the PID from the PID file is running
func IsRunning(pidFile string) (bool, int, error) {
	if pidFile == "" {
		pidFile = DefaultPidFile
	}

	if daemon.WasReborn() {
		// We are in daemon process
		return true, os.Getpid(), nil
	}

	// Check if daemon is running by trying to read PID file
	if _, err := os.Stat(pidFile); os.IsNotExist(err) {
		return false, 0, nil
	}

	// Read PID and check if process exists
	data, err := os.ReadFile(pidFile)
	if err != nil {
		return false, 0, fmt.Errorf("failed to read PID file: %w", err)
	}

	// Parse PID - go-daemon stores PID as string
	pidStr := string(data)
	if len(pidStr) > 0 && pidStr[len(pidStr)-1] == '\n' {
		pidStr = pidStr[:len(pidStr)-1]
	}

	var pid int
	if _, err := fmt.Sscanf(pidStr, "%d", &pid); err != nil {
		return false, 0, fmt.Errorf("invalid PID in file: %w", err)
	}

	// Check if process exists
	process, err := os.FindProcess(pid)
	if err != nil {
		return false, pid, nil
	}

	// Send signal 0 to check if process exists (Unix only)
	err = process.Signal(syscall.Signal(0))
	if err != nil {
		return false, pid, nil
	}

	return true, pid, nil
}
