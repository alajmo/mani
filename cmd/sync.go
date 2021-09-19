package cmd

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"github.com/alajmo/mani/core"
	"github.com/alajmo/mani/core/dao"
)

func syncCmd(config *dao.Config, configErr *error) *cobra.Command {
	var serial bool

	cmd := cobra.Command{
		Use:   "sync",
		Short: "Clone repositories and add them to gitignore",
		Long:  `Clone repositories and add them to gitignore.
In-case you need to enter credentials before cloning, run the command with the serial flag.`,
		Example: `  # Clone repositories one at a time
  mani sync

  # Clone repositories serial
  mani sync --serial`,
		Run: func(cmd *cobra.Command, args []string) {
			core.CheckIfError(*configErr)
			runSync(config, serial)
		},
	}

	cmd.Flags().BoolVarP(&serial, "serial", "s", false, "Clone projects one at a time")

	return &cmd
}

func runSync(config *dao.Config, serialFlag bool) {
	gitignoreFilename := filepath.Join(filepath.Dir(config.Path), ".gitignore")
	if _, err := os.Stat(gitignoreFilename); os.IsNotExist(err) {
		err := ioutil.WriteFile(gitignoreFilename, []byte(""), 0644)
		core.CheckIfError(err)
	}

	configDir := filepath.Dir(config.Path)

	// Get relative project names for gitignore file
	var projectNames []string
	for _, project := range config.Projects {
		if project.Url == "" {
			continue
		}

		if project.Path == "." {
			continue
		}

		// Project must be below mani config file
		projectPath, _ := dao.GetAbsolutePath(config.Path, project.Path, project.Name)
		if !strings.HasPrefix(projectPath, configDir) {
			continue
		}

		if project.Path != "" {
			relPath, _ := filepath.Rel(configDir, projectPath)
			projectNames = append(projectNames, relPath)
		} else {
			projectNames = append(projectNames, project.Name)
		}
	}

	err := dao.UpdateProjectsToGitignore(projectNames, gitignoreFilename)
	if err != nil {
		fmt.Println(err)
		return
	}

	config.CloneRepos(serialFlag)
}
