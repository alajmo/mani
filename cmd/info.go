package cmd

import (
	"fmt"
	color "github.com/logrusorgru/aurora"
	core "github.com/samiralajmovic/mani/core"
	"github.com/spf13/cobra"
	"path/filepath"
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
	var configPath string
	if *configFile != "" {
		configPath = *configFile
	} else {
		lala, err := core.GetClosestConfigFile()
		configPath = lala

		if err != nil {
			fmt.Println(err)
			return
		}
	}

	absConfigPath, err := filepath.Abs(configPath)
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println(color.Blue("Configuration: "), absConfigPath)
}
