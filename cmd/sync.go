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

func syncCmd(configFile *string) *cobra.Command {
	config, _ := dao.ReadConfig(*configFile)

	return &cobra.Command{
		Use:   "sync",
		Short: "Clone repositories and add to gitignore",
		Long:  `Clone repositories and add repository to gitignore.`,
		Run: func(cmd *cobra.Command, args []string) {
			runSync(&config)
		},
	}
}

func runSync(config *dao.Config) {
	gitignoreFilename := filepath.Join(filepath.Dir(config.Path), ".gitignore")
	if _, err := os.Stat(gitignoreFilename); os.IsNotExist(err) {
		err := ioutil.WriteFile(gitignoreFilename, []byte(""), 0644)
		core.CheckIfError(err)
	}

	configDir := filepath.Dir(config.Path)
	var projectNames []string
	for _, project := range config.Projects {
		if project.Url == "" {
			continue
		}

		if project.Path == "." {
			continue
		}

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

	urls := config.GetProjectUrls()
	if (len(urls) == 0) {
		fmt.Println("No projects to sync")
	} else {
		config.CloneRepos()
	}
}
