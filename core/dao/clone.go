package dao

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/theckman/yacspin"
	"github.com/jedib0t/go-pretty/v6/text"

	"github.com/alajmo/mani/core"
)

func (c Config) SyncProjects(configDir string, parallelFlag bool) {
	// Get relative project names for gitignore file
	var projectNames []string
	for _, project := range c.ProjectList {
		if project.Url == "" {
			continue
		}

		if project.Path == "." {
			continue
		}

		// Project must be below mani config file to be added to gitignore
		projectPath, _ := core.GetAbsolutePath(c.Path, project.Path, project.Name)
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

	// Only add projects to gitignore if a .gitignore file exists in the mani.yaml directory
	gitignoreFilename := filepath.Join(filepath.Dir(c.Path), ".gitignore")
	if _, err := os.Stat(gitignoreFilename); err == nil {
		core.CheckIfError(err)

		err := UpdateProjectsToGitignore(projectNames, gitignoreFilename)
		core.CheckIfError(err)
	}

	c.CloneRepos(parallelFlag)
}

func (c Config) CloneRepos(parallel bool) {
	// TODO: Refactor
	urls := c.GetProjectUrls()
	if len(urls) == 0 {
		fmt.Println("No projects to clone")
		return
	}

	cfg := yacspin.Config{
		Frequency:       100 * time.Millisecond,
		CharSet:         yacspin.CharSets[9],
		SuffixAutoColon: false,
		Message:         " Cloning",
	}

	spinner, err := yacspin.New(cfg)
	core.CheckIfError(err)

	if parallel {
		err = spinner.Start()
		core.CheckIfError(err)
	}

	syncErrors := sync.Map{}
	var wg sync.WaitGroup
	allProjectsSynced := true
	for _, project := range c.ProjectList {
		if project.IsSync() == false {
			continue
		}

		if project.Url != "" {
			wg.Add(1)

			if parallel {
				go CloneRepo(c.Path, project, parallel, &syncErrors, &wg)
			} else {
				CloneRepo(c.Path, project, parallel, &syncErrors, &wg)

				value, found := syncErrors.Load(project.Name)
				if found {
					allProjectsSynced = false
					fmt.Println(value)
				}
			}
		}
	}

	wg.Wait()

	if parallel {
		err = spinner.Stop()
		core.CheckIfError(err)

		for _, project := range c.ProjectList {
			if project.IsSync() == false {
				continue
			}

			value, found := syncErrors.Load(project.Name)
			if found {
				allProjectsSynced = false

				fmt.Printf("%v %v\n", text.FgRed.Sprintf("\u2715"), text.Bold.Sprintf(project.Name))
				fmt.Println(value)
			} else {
				fmt.Printf("%v %v\n", text.FgGreen.Sprintf("\u2713"), text.Bold.Sprintf(project.Name))
			}
		}
	}

	if allProjectsSynced {
		fmt.Println("\nAll projects synced")
	} else {
		fmt.Println("\nFailed to clone all projects")
	}
}

func CloneRepo(
	configPath string,
	project Project,
	parallel bool,
	syncErrors *sync.Map,
	wg *sync.WaitGroup,
) {
	// TODO: Refactor

	defer wg.Done()
	projectPath, err := core.GetAbsolutePath(configPath, project.Path, project.Name)
	if err != nil {
		syncErrors.Store(project.Name, (&core.FailedToParsePath{Name: projectPath}).Error())
		return
	}

	if _, err := os.Stat(projectPath); os.IsNotExist(err) {
		if !parallel {
			fmt.Printf("\n%v\n\n", text.Bold.Sprintf(project.Name))
		}

		var cmd *exec.Cmd
		if project.Clone == "" {
			cmd = exec.Command("git", "clone", project.Url, projectPath)
		} else {
			cmd = exec.Command("sh", "-c", project.Clone)
		}
		cmd.Env = os.Environ()

		if !parallel {
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr

			err := cmd.Run()
			if err != nil {
				syncErrors.Store(project.Name, err.Error())
			}
		} else {
			var errb bytes.Buffer
			cmd.Stderr = &errb

			err := cmd.Run()
			if err != nil {
				syncErrors.Store(project.Name, errb.String())
			}
		}
	}
}

