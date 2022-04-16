package exec

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	// "time"

	// "github.com/theckman/yacspin"
	"github.com/jedib0t/go-pretty/v6/text"

	"github.com/alajmo/mani/core"
	"github.com/alajmo/mani/core/dao"
)

func (c Exec) SyncProjects(configDir string, parallelFlag bool) {
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

	task := Task{Cmd: cmd, Name: "output"}
	target := exec.Exec{Projects: projects, Task: task, Config: *config}

	clientCh := make(chan Client, len(projects))
	errCh := make(chan error, len(projects))
	err := exec.SetClients(clientCh, errCh)
	core.CheckIfError(err)

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

}

func CloneRepo(
	configPath string,
	project Project,
	parallel bool,
	syncErrors *sync.Map,
	wg *sync.WaitGroup,
) {
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

		// TODO: Print errors from parallel false

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
