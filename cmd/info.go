package cmd

import (
	"github.com/spf13/cobra"

	"github.com/alajmo/mani/core/dao"
	"github.com/alajmo/mani/core/print"
)

func infoCmd(configFile *string) *cobra.Command {
	config, _ := dao.ReadConfig(*configFile)

	cmd := cobra.Command{
		Use:   "info",
		Short: "Print configuration file path",
		Long:  "Print configuration file path.",
		Run: func(cmd *cobra.Command, args []string) {
			runInfo(&config)
		},
	}

	return &cmd
}

func runInfo(config *dao.Config) {
	print.PrintInfo(config)
}
