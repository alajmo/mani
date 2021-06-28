package cmd

import (
	"github.com/alajmo/mani/core"
	"github.com/spf13/cobra"
)

func infoCmd(configFile *string) *cobra.Command {
	cmd := cobra.Command{
		Use:   "info",
		Short: "Print configuration file path",
		Long:  "Print configuration file path.",
		Run: func(cmd *cobra.Command, args []string) {
			runInfo(configFile)
		},
	}

	return &cmd
}

func runInfo(configFile *string) {
	configPath, config, _ := core.ReadConfig(*configFile)
	core.PrintInfo(configPath, config)
}
