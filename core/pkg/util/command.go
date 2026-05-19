package util

import "github.com/spf13/cobra"

func Command() *cobra.Command {
	command := &cobra.Command{
		Use:          "util",
		Short:        "utilities",
		SilenceUsage: true,
	}

	command.AddCommand(
		postgresqlMetricCollect,
		certification,
		centrifugePublish,
	)

	return command
}
