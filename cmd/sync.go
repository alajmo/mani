package cmd

import (
	"fmt"
	core "github.com/alajmo/mani/core"
	"github.com/spf13/cobra"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

func syncCmd(configFile *string) *cobra.Command {
	return &cobra.Command{
		Use:   "sync",
		Short: "Clone repositories and add to gitignore",
		Long:  `Clone repositories and add repository to gitignore.`,
		Run: func(cmd *cobra.Command, args []string) {
			runSync(*configFile)
		},
	}
}

func runSync(configFile string) {
	configPath, config, err := core.ReadConfig(configFile)
	core.CheckIfError(err)

	gitignoreFilename := filepath.Join(filepath.Dir(configPath), ".gitignore")
	if _, err := os.Stat(gitignoreFilename); os.IsNotExist(err) {
		err := ioutil.WriteFile(gitignoreFilename, []byte(""), 0644)
		core.CheckIfError(err)
	}

	configDir := filepath.Dir(configPath)
	var projectNames []string
	for _, project := range config.Projects {
		if project.Url == "" {
			continue
		}

		if project.Path == "." {
			continue
		}

		projectPath, _ := core.GetAbsolutePath(configPath, project.Path, project.Name)
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

	err = core.UpdateProjectsToGitignore(projectNames, gitignoreFilename)
	if err != nil {
		fmt.Println(err)
		return
	}

	urls := core.GetProjectUrls(config.Projects)
	if (len(urls) == 0) {
		fmt.Println("No projects to sync")
	} else {
		core.CloneRepos(configPath, config.Projects)
	}
}
