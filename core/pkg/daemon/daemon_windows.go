//go:build windows

package daemon

import (
	"fmt"

	"ntels.com/pharos/core/pkg/common"
)

const (
	DefaultPidFile = "/var/run/pharos.pid"
	DefaultLogFile = "/var/log/pharos.log"
)

var ErrNotSupportedWindows = fmt.Errorf("daemon mode is not supported on Windows")

// DaemonizeWithConfig is not supported on Windows
func DaemonizeWithConfig(pidFile, logFile string, daemonConfig common.DaemonConfig) error {
	return ErrNotSupportedWindows
}

// RemovePidFile removes the PID file
func RemovePidFile(pidFile string) error {
	return ErrNotSupportedWindows
}

// IsRunning checks if a process with the PID from the PID file is running
func IsRunning(pidFile string) (bool, int, error) {
	return false, 0, ErrNotSupportedWindows
}
