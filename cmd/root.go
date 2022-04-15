package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/alajmo/mani/core/dao"
)

const (
	appName      = "mani"
	shortAppDesc = "mani is a tool used to manage repositories"
	longAppDesc  = `mani is a tool used to manage repositories`
)

var (
	config         dao.Config
	configErr      error
	configFilepath string
	userConfigDir  string
	noColor        bool
	rootCmd        = &cobra.Command{
		Use:   appName,
		Short: shortAppDesc,
		Long:  longAppDesc,
	}
)

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	defaultUserConfigDir, _ := os.UserConfigDir()
	defaultUserConfigDir = filepath.Join(defaultUserConfigDir, "mani")

	rootCmd.PersistentFlags().StringVarP(&configFilepath, "config", "c", "", "Config file (by default it checks current and all parent directories for mani.yaml|yml)")
	rootCmd.PersistentFlags().StringVar(&userConfigDir, "user-config-dir", defaultUserConfigDir, "Set user config directory")
	rootCmd.PersistentFlags().BoolVar(&noColor, "no-color", false, "Disable color")

	rootCmd.AddCommand(
		versionCmd(),
		completionCmd(),
		initCmd(),
		execCmd(&config, &configErr),
		runCmd(&config, &configErr),
		listCmd(&config, &configErr),
		describeCmd(&config, &configErr),
		syncCmd(&config, &configErr),
		editCmd(&config, &configErr),
	)
}

func initConfig() {
	config, configErr = dao.ReadConfig(configFilepath, userConfigDir, noColor)
}
