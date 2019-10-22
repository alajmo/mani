package cmd

import (
	"bufio"
	"fmt"
	color "github.com/logrusorgru/aurora"
	core "github.com/samiralajmovic/loop/core"
	"github.com/spf13/cobra"
	"os"
	"path/filepath"
)

func syncCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "sync",
		Short: "Sync repos",
		Long:  "Sync repos",
		Run:   runSync,
	}
}

func runSync(cmd *cobra.Command, args []string) {
	// Check gitignore file exists
	gitignoreFilename, _ := filepath.Abs(".gitignore")
	if _, err := os.Stat(gitignoreFilename); os.IsNotExist(err) {
		fmt.Println("fatal: missing", filepath.Base(gitignoreFilename))
		return
	}

	// Clone Repos if not cloned
	configFilename, _ := filepath.Abs("mani.yaml")
	if _, err := os.Stat(configFilename); os.IsNotExist(err) {
		fmt.Println("fatal: not a mani repository (or any of the parent directories):", filepath.Base(configFilename))
	} else {
		config := core.ReadConfig()

		// Get Projects
		projects := make(map[string]bool)
		for _, project := range config.Projects {
			projects[project.Name] = false
		}

		// Add Projects to gitignore
		gitignoreFile, _ := os.Open(gitignoreFilename)
		scanner := bufio.NewScanner(gitignoreFile)
		for scanner.Scan() {
			line := scanner.Text()
			if _, ok := projects[line]; ok {
				projects[line] = true
			}
		}

		gitignoreFile, _ = os.OpenFile(gitignoreFilename, os.O_APPEND|os.O_WRONLY, 0644)
		for project, found := range projects {
			if !found {
				gitignoreFile.WriteString(project)
				gitignoreFile.WriteString("\n")
				fmt.Println(color.Green("\u2713"), "added project", color.Bold(project), "to .gitignore")
			}
		}
		gitignoreFile.Close()

		// Clone Repos
		for _, project := range config.Projects {
			// Check projects exist in gitignore, if not, add them
			// AddStringToFile(project.Name, ".gitignore")

			if project.Url != "" {
				core.CloneRepo(project.Name, project.Url, project.Path)
			}
		}
	}
}
