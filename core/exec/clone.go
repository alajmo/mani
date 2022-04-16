package exec

import (
	// "bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	// "sync"

	// "github.com/jedib0t/go-pretty/v6/text"

	"github.com/alajmo/mani/core"
	"github.com/alajmo/mani/core/dao"
)

func CloneRepos(config *dao.Config, parallel bool) {
	urls := config.GetProjectUrls()
	if len(urls) == 0 {
		fmt.Println("No projects to clone")
		return
	}

	var tasks []dao.Task
	projects := config.ProjectList
	for i := range projects {
		if projects[i].Clone != "" {
			tasks = append(tasks, dao.Task{Cmd: projects[i].Clone, Name: projects[i].Name})
		} else {
			projectPath, err := core.GetAbsolutePath(config.Path, projects[i].Path, projects[i].Name)
			if err != nil {
				// syncErrors.Store(project.Name, (&core.FailedToParsePath{Name: projectPath}).Error())
			}

			// fmt.Println(projects[i].Url, projectPath)
			fmt.Println(i)
			tasks = append(tasks, dao.Task{Shell: "sh", Cmd: fmt.Sprintf("git clone %s %s", projects[i].Url, projectPath), Name: "output"})
		}
	}

	target := Exec{Projects: projects, Tasks: tasks, Config: *config}

	clientCh := make(chan Client, len(projects))
	errCh := make(chan error, len(projects))
	err := target.SetCloneClients(clientCh, errCh)
	core.CheckIfError(err)

	target.Text(false)

	// for _, project := range config.ProjectList {
	// 	if project.IsSync() == false {
	// 		continue
	// 	}
	// 	if project.Url != "" {
	// 		wg.Add(1)
	// 		if parallel {
	// 			go CloneRepo(config.Path, project, parallel, &syncErrors, &wg)
	// 		} else {
	// 			CloneRepo(config.Path, project, parallel, &syncErrors, &wg)
	// 			value, found := syncErrors.Load(project.Name)
	// 			if found {
	// 				allProjectsSynced = false
	// 				fmt.Println(value)
	// 			}
	// 		}
	// 	}
	// }
}

// func CloneRepo(
// 	configPath string,
// 	project Project,
// 	parallel bool,
// 	syncErrors *sync.Map,
// 	wg *sync.WaitGroup,
// ) {
// 	defer wg.Done()

// 	projectPath, err := core.GetAbsolutePath(configPath, project.Path, project.Name)
// 	if err != nil {
// 		syncErrors.Store(project.Name, (&core.FailedToParsePath{Name: projectPath}).Error())
// 		return
// 	}

// 	if _, err := os.Stat(projectPath); os.IsNotExist(err) {
// 		if !parallel {
// 			fmt.Printf("\n%v\n\n", text.Bold.Sprintf(project.Name))
// 		}

// 		var cmd *exec.Cmd
// 		if project.Clone == "" {
// 			cmd = exec.Command("git", "clone", project.Url, projectPath)
// 		} else {
// 			cmd = exec.Command("sh", "-config", project.Clone)
// 		}
// 		cmd.Env = os.Environ()

// 		// TODO: Print errors from parallel false

// 		if !parallel {
// 			cmd.Stdout = os.Stdout
// 			cmd.Stderr = os.Stderr

// 			err := cmd.Run()
// 			if err != nil {
// 				syncErrors.Store(project.Name, err.Error())
// 			}
// 		} else {
// 			var errb bytes.Buffer
// 			cmd.Stderr = &errb

// 			err := cmd.Run()
// 			if err != nil {
// 				syncErrors.Store(project.Name, errb.String())
// 			}
// 		}
// 	}
// }

func UpdateGitignoreIfExists(config *dao.Config) {
	// Only add projects to gitignore if a .gitignore file exists in the mani.yaml directory
	gitignoreFilename := filepath.Join(filepath.Dir(config.Path), ".gitignore")
	if _, err := os.Stat(gitignoreFilename); err == nil {
		core.CheckIfError(err)
		// Get relative project names for gitignore file
		var projectNames []string
		for _, project := range config.ProjectList {
			if project.Url == "" {
				continue
			}

			if project.Path == "." {
				continue
			}

			// Project must be below mani config file to be added to gitignore
			projectPath, _ := core.GetAbsolutePath(config.Path, project.Path, project.Name)
			if !strings.HasPrefix(projectPath, config.Dir) {
				continue
			}

			if project.Path != "" {
				relPath, _ := filepath.Rel(config.Dir, projectPath)
				projectNames = append(projectNames, relPath)
			} else {
				projectNames = append(projectNames, project.Name)
			}
		}

		err := dao.UpdateProjectsToGitignore(projectNames, gitignoreFilename)
		core.CheckIfError(err)
	}
}

func (exec *Exec) SetCloneClients(
	clientCh chan Client,
	errCh chan error,
) error {
	config := exec.Config
	projects := exec.Projects

	var clients []Client
	for i, project := range projects {
		func(i int, project dao.Project) {
			client := Client { Path: config.Dir, Name: project.Name }
			clientCh <- client
			clients = append(clients, client)
		}(i, project)
	}

	close(clientCh)
	close(errCh)

	// Return if there's any errors
	for err := range errCh {
		return err
	}

	exec.Clients = clients

	return nil
}
