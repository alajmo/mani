package cmd

import (
	"fmt"
	color "github.com/logrusorgru/aurora"
	"github.com/alajmo/mani/core"
	"github.com/spf13/cobra"
)

func infoCmd(configFile *string) *cobra.Command {
	cmd := cobra.Command{
		Use:   "info",
		Short: "Print configuration file path",
		Long:  "Print configuration file path.",
		Run: func(cmd *cobra.Command, args []string) {
			printInfo(configFile)
		},
	}

	return &cmd
}

func printInfo(configFile *string) {
	configPath, _, err := core.ReadConfig(*configFile)
	core.CheckIfError(err)

	fmt.Println(color.Blue("Configuration: "), configPath)
}
