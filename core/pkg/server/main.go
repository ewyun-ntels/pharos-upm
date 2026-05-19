package server

import (
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/spf13/cobra"
	"ntels.com/pharos/core/pkg/common"
	"ntels.com/pharos/core/pkg/daemon"
	"ntels.com/pharos/core/pkg/message_router"
	message_router_server "ntels.com/pharos/core/pkg/message_router/server"
	"ntels.com/pharos/core/pkg/message_router/subjects"
	"ntels.com/pharos/core/pkg/server/http"
	"ntels.com/pharos/core/pkg/server/http/gin"
	"ntels.com/pharos/core/pkg/server/internal"
	"ntels.com/pharos/core/pkg/server/tcp"
	"ntels.com/pharos/core/pkg/server/udp"
)

func Command() *cobra.Command {
	command := &cobra.Command{
		Use:          "serve",
		Short:        "start server",
		SilenceUsage: true,
		PreRunE: func(command *cobra.Command, args []string) error {
			// Load config first
			if configPath, err := command.Flags().GetStringArray("config"); err != nil {
				return err
			} else if err := load(configPath); err != nil {
				return err
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
			defer unload()

			// Get config path again for RunE scope
			configPath, _ := command.Flags().GetStringArray("config")

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

			return runServerWithRetry(configPath, config)
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
		"", "PID file path (default: /var/run/pharos.pid)")

	command.PersistentFlags().String(
		"log-file",
		"", "log file path for daemon mode (default: /var/log/pharos.log)")

	return command
}

// runServerWithRetry runs the server with graceful error handling
func runServerWithRetry(configPath []string, config common.Config) error {
	// Use first config path for ginServer.Start (which expects string)
	var configPathStr string
	if len(configPath) > 0 {
		configPathStr = configPath[0]
	}

	message_router.Subjects.RemoveAll(func(message_router.Subject) {})
	for _, subject := range []message_router.Subject{
		subjects.GetAgentPingSubject(config),
	} {
		message_router.Subjects.Set(subject.Key, subject)
	}
	for _, subject := range message_router.AddSubjects.GetAll() {
		message_router.Subjects.Set(subject.Key, subject)
	}

	messageRouterServer := message_router_server.NewServer()
	if err := messageRouterServer.Start(configPathStr, config, slog.Default()); err != nil {
		return err
	}
	defer func() {
		if err := messageRouterServer.Stop(); err != nil {
			slog.Error("server stop error", "error", err)
		}
	}()

	http.AddServer("core_http_server", &gin.Server{})

	for name, server := range http.Servers.GetAll() {
		slog.Info("starting http server", "name", name)
		if err := server.Start(configPathStr, config, slog.Default()); err != nil {
			slog.Error("failed to start http server", "name", name, "error", err)
			return err
		}
		defer func(name string, server internal.Server) {
			if err := server.Stop(); err != nil {
				slog.Error("http server stop error", "name", name, "error", err)
			}
		}(name, server)
	}

	for name, server := range tcp.Servers.GetAll() {
		slog.Info("starting tcp server", "name", name)
		if err := server.Start(configPathStr, config, slog.Default()); err != nil {
			slog.Error("failed to start tcp server", "name", name, "error", err)
			return err
		}
		defer func(name string, server internal.Server) {
			if err := server.Stop(); err != nil {
				slog.Error("tcp server stop error", "name", name, "error", err)
			}
		}(name, server)
	}

	for name, server := range udp.Servers.GetAll() {
		slog.Info("starting udp server", "name", name)
		if err := server.Start(configPathStr, config, slog.Default()); err != nil {
			slog.Error("failed to start udp server", "name", name, "error", err)
			return err
		}
		defer func(name string, server internal.Server) {
			if err := server.Stop(); err != nil {
				slog.Error("udp server stop error", "name", name, "error", err)
			}
		}(name, server)
	}

	// Wait for shutdown signals
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)
	<-signals

	slog.Info("server shutdown gracefully")
	return nil
}
