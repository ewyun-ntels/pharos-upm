package util

import (
	"fmt"

	"github.com/spf13/cobra"
	"ntels.com/pharos/core/internal/utility"
)

var centrifugePublish = &cobra.Command{
	Use:   "centrifuge_publish endpoint channel data",
	Short: "centrifuge publish",
	PreRunE: func(cmd *cobra.Command, args []string) error {
		if len(args) != 3 {
			return fmt.Errorf("invalid args(%v)", args)
		}

		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		return utility.CentrifugePublish(args[0], args[1], args[2])
	},
}
