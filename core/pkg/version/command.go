package version

import (
	"fmt"

	"github.com/spf13/cobra"
)

func Command() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Show version information",
		Long:  "Display version information",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("Pharos version: %s\n", Version)
			if Commit != "" {
				fmt.Printf("Commit: %s\n", Commit)
			}
			if BuildTime != "" {
				fmt.Printf("Build time: %s\n", BuildTime)
			}
		},
	}
}
