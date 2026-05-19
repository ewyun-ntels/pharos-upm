package command

import (
	"log/slog"
	"time"

	"github.com/spf13/cobra"
	"ntels.com/pharos/core/pkg/common"
	"ntels.com/pharos/core/pkg/daemon"
	common_command "ntels.com/pharos/extensions/catv/business/pkg/common/command"
	control_common "ntels.com/pharos/extensions/catv/business/pkg/control/common"
)

func Command() *cobra.Command {
	var config common.Config

	command := &cobra.Command{
		Use:          control_common.CommandUse,
		Short:        "execute control schedules",
		SilenceUsage: true,
		PreRunE: func(command *cobra.Command, args []string) error {
			if configPath, err := command.Flags().GetStringArray("config"); err != nil {
				slog.Error("failed to get config flag", "error", err)
				return err
			} else if cfg, err := common_command.Load(configPath); err != nil {
				slog.Error("failed to load config", "error", err)
				return err
			} else {
				config = cfg
			}

			isDaemon, err := command.Flags().GetBool("daemon")
			if err != nil {
				return err
			}

			if isDaemon {
				pidFile, _ := command.Flags().GetString("pid-file")
				logFile, _ := command.Flags().GetString("log-file")

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

				slog.Info("Shutting down control-schedule processor...", "duration_seconds", time.Since(startTime).Seconds())
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

			k8sJobName, _ := command.Flags().GetString("k8s-job-name")
			service, err := NewService(config, k8sJobName)
			if err != nil {
				return err
			}

			if err := service.Start(); err != nil {
				return err
			}
			if err := service.Stop(); err != nil {
				return err
			}

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
		"/var/run/control-schedule.pid", "PID file path (default: /var/run/control-schedule.pid)")

	command.PersistentFlags().String(
		"log-file",
		"/var/log/control-schedule.log", "log file path for daemon mode (default: /var/log/control-schedule.log)")

	command.PersistentFlags().String(
		"k8s-job-name",
		"", "k8s job name for the control-schedule processor")

	return command
}
