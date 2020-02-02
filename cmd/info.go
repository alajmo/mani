package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	color "github.com/logrusorgru/aurora"
	core "github.com/samiralajmovic/mani/core"
)

func infoCmd() *cobra.Command {
	cmd := cobra.Command{
		Use:   "info",
		Short: "Print configuration file",
		Long: "Print configuration file",
		Run: func(cmd *cobra.Command, args []string) {
			printInfo()
		},
	}

	return &cmd
}


func printInfo() {
	filename, err := core.GetClosestConfigFile()

	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println(color.Blue("Configuration: "), filename)
}
