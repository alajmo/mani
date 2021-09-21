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
	configDir := filepath.Dir(config.Path)

	syncDirs(configDir, config, serialFlag)
	syncProjects(configDir, config, serialFlag)
}

func syncDirs(configDir string, config *dao.Config, serialFlag bool) {
	for _, dir := range config.Dirs {
		fmt.Println(dir.Path)

		if _, err := os.Stat(dir.Path); os.IsNotExist(err) {
			os.MkdirAll(dir.Path, os.ModePerm)
		}
	}
}

func syncProjects(configDir string, config *dao.Config, serialFlag bool) {
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
		projectPath, _ := core.GetAbsolutePath(config.Path, project.Path, project.Name)
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

	if len(projectNames) > 0 {
		gitignoreFilename := filepath.Join(filepath.Dir(config.Path), ".gitignore")
		if _, err := os.Stat(gitignoreFilename); os.IsNotExist(err) {
			err := ioutil.WriteFile(gitignoreFilename, []byte(""), 0644)
			core.CheckIfError(err)
		}

		err := dao.UpdateProjectsToGitignore(projectNames, gitignoreFilename)
		if err != nil {
			fmt.Println(err)
			return
		}

		config.CloneRepos(serialFlag)
	}
}
