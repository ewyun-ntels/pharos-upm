package main

import (
	"log/slog"
	"os"

	"github.com/spf13/cobra"
	"ntels.com/pharos/core/pkg/command"
	"ntels.com/pharos/core/pkg/daemon"
	"ntels.com/pharos/core/pkg/server"
	"ntels.com/pharos/core/pkg/util"
	"ntels.com/pharos/core/pkg/version"
	// Extensions는 SITE_MODE에 따라 import_*.go 파일에서 자동으로 import됩니다
	// 새로운 extension 추가 시: go generate
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	slog.SetDefault(logger)

	var rootCmd = &cobra.Command{Use: "pharos"}

	// command가 필요할 경우 하단에 추가
	rootCmd.AddCommand(server.Command())
	rootCmd.AddCommand(daemon.Command())
	rootCmd.AddCommand(util.Command())
	rootCmd.AddCommand(version.Command())

	for name, c := range command.GetCommands() {
		if c == nil {
			slog.Warn("Skipped registering nil command", "name", name)
			continue
		}

		slog.Info("Register command", "name", name)
		rootCmd.AddCommand(c)
	}

	err := rootCmd.Execute()
	if err != nil {
		slog.Error("Error executing root command", "error", err)
		return
	}
}
