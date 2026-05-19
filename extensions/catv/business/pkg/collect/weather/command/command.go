package command

import (
	"log/slog"
	"time"

	"github.com/spf13/cobra"
	"ntels.com/pharos/core/pkg/common"
	"ntels.com/pharos/core/pkg/daemon"
	"ntels.com/pharos/extensions/catv/business/pkg/collect/weather/service"
	common_command "ntels.com/pharos/extensions/catv/business/pkg/common/command"
)

var GlobalService *service.Service

func Command() *cobra.Command {
	var config common.Config

	command := &cobra.Command{
		Use:          "catv-collect-weather",
		Short:        "CATV Collect Weather Data Processor",
		SilenceUsage: true,
		PreRunE: func(command *cobra.Command, args []string) error {
			// Load config first
			if configPath, err := command.Flags().GetStringArray("config"); err != nil {
				return err
			} else if cfg, err := common_command.Load(configPath); err != nil {
				return err
			} else {
				config = cfg
			}

			// Check if daemon mode is requested AFTER config is loaded
			isDaemon, err := command.Flags().GetBool("daemon")
			if err != nil {
				return err
			}

			if isDaemon {
				pidFile, _ := command.Flags().GetString("pid-file")
				logFile, _ := command.Flags().GetString("log-file")

				// Initialize daemonization with config support
				if err := daemon.DaemonizeWithConfig(pidFile, logFile, config.Daemon); err != nil {
					return err
				}
			}

			return nil
		},
		RunE: func(command *cobra.Command, args []string) error {
			startTime := time.Now()

			defer func() {
				common_command.Unload()

				slog.Info("Total CATV weather job processor runtime", "duration_seconds", time.Since(startTime).Seconds())
			}()

			// Get config path again for RunE scope
			_, _ = command.Flags().GetStringArray("config")

			// Check if running as daemon for cleanup
			isDaemon, _ := command.Flags().GetBool("daemon")
			if isDaemon {
				pidFile, _ := command.Flags().GetString("pid-file")
				defer func() {
					if err := daemon.RemovePidFile(pidFile); err != nil {
						slog.Error("failed to remove PID file", "error", err)
					}
				}()

				// Add panic recovery for daemon mode
				defer func() {
					if r := recover(); r != nil {
						slog.Error("daemon panic recovered", "panic", r)
						// Don't restart automatically - let systemd handle it
					}
				}()
			}

			service, err := service.NewService(config)
			if err != nil {
				return err
			}
			GlobalService = service

			if err := GlobalService.Start(); err != nil {
				return err
			}
			if err := GlobalService.Stop(); err != nil {
				return err
			}

			slog.Info("Shutting down CATV weather job processor...")

			return nil
		},
	}

	command.PersistentFlags().StringArrayP(
		"config",
		"c",
		[]string{"./config/config.toml"}, "configuration file path")

	command.PersistentFlags().BoolP(
		"daemon",
		"d",
		false, "run as daemon (background process)")

	command.PersistentFlags().String(
		"pid-file",
		"/var/run/catv-weather.pid", "PID file path (default: /var/run/catv-weather.pid)")

	command.PersistentFlags().String(
		"log-file",
		"/var/log/catv-weather.log", "log file path for daemon mode (default: /var/log/catv-weather.log)")

	return command
}
