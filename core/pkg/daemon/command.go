package daemon

import (
	"fmt"
	"log/slog"
	"os"
	"syscall"

	"github.com/spf13/cobra"
)

// Command returns the daemon management command
func Command() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "daemon",
		Short: "Daemon management commands",
	}

	cmd.AddCommand(statusCommand())
	cmd.AddCommand(stopCommand())

	return cmd
}

// statusCommand returns the daemon status command
func statusCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "status",
		Short: "Check daemon status",
		RunE: func(cmd *cobra.Command, args []string) error {
			pidFile, _ := cmd.Flags().GetString("pid-file")

			running, pid, err := IsRunning(pidFile)
			if err != nil {
				return fmt.Errorf("failed to check daemon status: %w", err)
			}

			if running {
				slog.Info("daemon is running", "pid", pid)
				fmt.Printf("Daemon is running (PID: %d)\n", pid)
			} else {
				if pid > 0 {
					fmt.Printf("Daemon is not running (stale PID file with PID: %d)\n", pid)
				} else {
					fmt.Println("Daemon is not running")
				}
			}

			return nil
		},
	}

	cmd.Flags().String("pid-file", "", "PID file path (default: /var/run/pharos.pid)")

	return cmd
}

// stopCommand returns the daemon stop command
func stopCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "stop",
		Short: "Stop daemon",
		RunE: func(cmd *cobra.Command, args []string) error {
			pidFile, _ := cmd.Flags().GetString("pid-file")

			running, pid, err := IsRunning(pidFile)
			if err != nil {
				return fmt.Errorf("failed to check daemon status: %w", err)
			}

			if !running {
				fmt.Println("Daemon is not running")
				return nil
			}

			// Send SIGTERM to the daemon process
			process, err := os.FindProcess(pid)
			if err != nil {
				return fmt.Errorf("failed to find process: %w", err)
			}

			if err := process.Signal(syscall.SIGTERM); err != nil {
				return fmt.Errorf("failed to send SIGTERM to daemon: %w", err)
			}

			slog.Info("sent SIGTERM to daemon", "pid", pid)
			fmt.Printf("Sent SIGTERM to daemon (PID: %d)\n", pid)

			return nil
		},
	}

	cmd.Flags().String("pid-file", "", "PID file path (default: /var/run/pharos.pid)")

	return cmd
}
