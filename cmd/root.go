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
	shortAppDesc = "mani is a tool used to manage multiple repositories"
	longAppDesc  = `mani is a tool used to manage multiple repositories`
)

var (
	config     dao.Config
	configErr  error
	configFile string
	userConfigDir string
	rootCmd    = &cobra.Command{
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
	rootCmd.PersistentFlags().StringVarP(&configFile, "config", "c", "", "config file (by default it checks current and all parent directories for mani.yaml|yml)")

	defaultUserConfigDir, _ := os.UserConfigDir()
	defaultUserConfigDir = filepath.Join(defaultUserConfigDir, "mani")

	rootCmd.PersistentFlags().StringVar(&userConfigDir, "user-config-dir", defaultUserConfigDir, "Set user config directory")

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
	config, configErr = dao.ReadConfig(configFile, userConfigDir)
}
