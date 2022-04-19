package exec

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/jedib0t/go-pretty/v6/text"

	"github.com/alajmo/mani/core"
	"github.com/alajmo/mani/core/dao"
	"github.com/alajmo/mani/core/print"
)

func CloneRepos(config *dao.Config, parallel bool) {
	urls := config.GetProjectUrls()
	if len(urls) == 0 {
		fmt.Println("No projects to clone")
		return
	}

	var projects []dao.Project
	for i := range config.ProjectList {
		if config.ProjectList[i].IsSync() == false {
			continue
		}

		if config.ProjectList[i].Url == "" {
			continue
		}

		projectPath, err := core.GetAbsolutePath(config.Path, config.ProjectList[i].Path, config.ProjectList[i].Name)
		core.CheckIfError(err)
		// Project already synced
		if _, err := os.Stat(projectPath); !os.IsNotExist(err) {
			continue
		}

		projects = append(projects, config.ProjectList[i])
	}

	var tasks []dao.Task
	for i := range projects {
		var cmd string
		var cmdArr []string
		var shell string
		var shellProgram string
		if projects[i].Clone != "" {
			shell = "sh -c"
			shellProgram = "sh"
			cmdArr = []string{"-c", projects[i].Clone}
			cmd = projects[i].Clone
		} else {
			projectPath, err := core.GetAbsolutePath(config.Path, projects[i].Path, projects[i].Name)
			core.CheckIfError(err)

			shell = "git"
			shellProgram = "git"
			cmdArr = []string{"clone", "--progress", projects[i].Url, projectPath}
			cmd = strings.Join(cmdArr, " ")
		}

		var task = dao.Task {
			Name: projects[i].Name,

			Shell: shell,
			Cmd: cmd,
			ShellProgram: shellProgram,
			CmdArg: cmdArr,
			SpecData: dao.Spec {
				Parallel: parallel,
			},

			ThemeData: dao.Theme {
				Text: dao.Text {
					Prefix: false,
					Header: true,
					HeaderChar: "*",
					HeaderPrefix: "Project",
				},
			},
		}

		tasks = append(tasks, task)
	}

	target := Exec{Projects: projects, Tasks: tasks, Config: *config}

	clientCh := make(chan Client, len(projects))
	errCh := make(chan error, len(projects))
	err := target.SetCloneClients(clientCh, errCh)
	core.CheckIfError(err)

	target.Text(false)
}

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

func PrintProjectStatus(config *dao.Config) {
	theme, err := config.GetTheme("default")
	core.CheckIfError(err)
	options := print.PrintTableOptions {
		Theme: *theme,
		OmitEmpty: true,
		Output: "table",
		SuppressEmptyColumns: false,
	}

	data := dao.TableOutput {
		Headers: []string{"project", "synced"},
		Rows: []dao.Row {},
	}

	for _, project := range config.ProjectList {
		projectPath, err := core.GetAbsolutePath(config.Path, project.Path, project.Name)
		core.CheckIfError(err)

		if _, err := os.Stat(projectPath); !os.IsNotExist(err) {
			// Project  synced
			data.Rows = append(data.Rows, dao.Row { Columns: []string{ project.Name, text.FgGreen.Sprintf("\u2713") } })
		} else {
			// Project not synced
			data.Rows = append(data.Rows, dao.Row { Columns: []string{ project.Name, text.FgRed.Sprintf("\u2715") } })
		}
	}

	print.PrintTable(data.Rows, options, data.Headers, []string{})
}

func PrintProjectInit(configDir string, projects []dao.Project) {
	theme := dao.Theme {
		Table: dao.DefaultTable,
	}

	options := print.PrintTableOptions {
		Theme: theme,
		OmitEmpty: true,
		Output: "table",
		SuppressEmptyColumns: false,
	}

	data := dao.TableOutput {
		Headers: []string{"project", "path"},
		Rows: []dao.Row {},
	}

	for _, project := range projects {
		data.Rows = append(data.Rows, dao.Row { Columns: []string{ project.Name, project.Path} })
	}

	fmt.Println("\nFollowing projects were added to mani.yaml")
	print.PrintTable(data.Rows, options, data.Headers, []string{})
}

