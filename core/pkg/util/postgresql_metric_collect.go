package util

import (
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/rapidloop/pgmetrics/collector"
	"github.com/spf13/cobra"
)

var postgresqlMetricCollect = &cobra.Command{
	Use:   "postgresql_metric_collect host port user password [databases]",
	Short: "postgresql metric collect",
	PreRunE: func(cmd *cobra.Command, args []string) error {
		if len(args) < 4 {
			return fmt.Errorf("invalid args(%v)", args)
		} else if _, err := strconv.ParseUint(args[1], 10, 16); err != nil {
			return fmt.Errorf("invalid port(%s)", args[1])
		}

		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		port, _ := strconv.ParseUint(args[1], 10, 16)

		config := collector.DefaultCollectConfig()
		config.Host = args[0]
		config.Port = uint16(port)
		config.User = args[2]
		config.Password = args[3]

		databases := []string{}
		for i := 4; i < len(args); i++ {
			databases = append(databases, args[i])
		}

		model := collector.Collect(config, databases)
		if output, err := json.Marshal(model); err != nil {
			return err
		} else {
			fmt.Println(string(output))
		}

		return nil
	},
}
