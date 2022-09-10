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

func CloneRepos(config *dao.Config, syncProjects []dao.Project, parallel bool) error {
	urls := config.GetProjectUrls()
	if len(urls) == 0 {
		fmt.Println("No projects to clone")
		return nil
	}

	var projects []dao.Project
	for i := range syncProjects {
		if !syncProjects[i].IsSync() {
			continue
		}

		if syncProjects[i].Url == "" {
			continue
		}

		projectPath, err := core.GetAbsolutePath(config.Path, syncProjects[i].Path, syncProjects[i].Name)
		if err != nil {
			return err
		}
		// Project already synced
		if _, err := os.Stat(projectPath); !os.IsNotExist(err) {
			continue
		}

		projects = append(projects, syncProjects[i])
	}

	var tasks []dao.Task
	for i := range projects {
		var cmd string
		var cmdArr []string
		var shell string
		var shellProgram string
		if projects[i].Clone != "" {
			shell = dao.DEFAULT_SHELL
			shellProgram = dao.DEFAULT_SHELL_PROGRAM
			cmdArr = []string{"-c", projects[i].Clone}
			cmd = projects[i].Clone
		} else {
			projectPath, err := core.GetAbsolutePath(config.Path, projects[i].Path, projects[i].Name)
			if err != nil {
				return err
			}

			shell = "git"
			shellProgram = "git"
			if parallel {
				cmdArr = []string{"clone", projects[i].Url, projectPath}
			} else {
				cmdArr = []string{"clone", "--progress", projects[i].Url, projectPath}
			}
			cmd = strings.Join(cmdArr, " ")
		}

		var task = dao.Task{
			Name: projects[i].Name,

			Shell:        shell,
			Cmd:          cmd,
			ShellProgram: shellProgram,
			CmdArg:       cmdArr,
			SpecData: dao.Spec{
				Parallel:     parallel,
				IgnoreErrors: false,
			},

			ThemeData: dao.Theme{
				Text: dao.Text{
					Prefix:       parallel, // we only use prefix when parallel is enabled since we need to see which project returns an error
					Header:       true,
					HeaderChar:   dao.DefaultText.HeaderChar,
					HeaderPrefix: "Project",
					PrefixColors: dao.DefaultText.PrefixColors,
				},
			},
		}

		tasks = append(tasks, task)
	}

	target := Exec{Projects: projects, Tasks: tasks, Config: *config}

	clientCh := make(chan Client, len(projects))
	err := target.SetCloneClients(clientCh)
	if err != nil {
		return err
	}

	target.Text(false)

	return nil
}

func UpdateGitignoreIfExists(config *dao.Config) error {
	// Only add projects to gitignore if a .gitignore file exists in the mani.yaml directory
	gitignoreFilename := filepath.Join(filepath.Dir(config.Path), ".gitignore")
	if _, err := os.Stat(gitignoreFilename); err == nil {
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
			var projectPath string
			projectPath, err = core.GetAbsolutePath(config.Path, project.Path, project.Name)
			if err != nil {
				return err
			}

			if !strings.HasPrefix(projectPath, config.Dir) {
				continue
			}

			if project.Path != "" {
				var relPath string
				relPath, err = filepath.Rel(config.Dir, projectPath)
				if err != nil {
					return err
				}
				projectNames = append(projectNames, relPath)
			} else {
				projectNames = append(projectNames, project.Name)
			}
		}

		err := dao.UpdateProjectsToGitignore(projectNames, gitignoreFilename)
		if err != nil {
			return err
		}
	}

	return nil
}

func (exec *Exec) SetCloneClients(clientCh chan Client) error {
	config := exec.Config
	projects := exec.Projects

	var clients []Client
	for i, project := range projects {
		func(i int, project dao.Project) {
			client := Client{
				Path: config.Dir,
				Name: project.Name,
				Env:  projects[i].EnvList,
			}
			clientCh <- client
			clients = append(clients, client)
		}(i, project)
	}

	close(clientCh)

	exec.Clients = clients

	return nil
}

func PrintProjectStatus(config *dao.Config, projects []dao.Project) error {
	theme, err := config.GetTheme("default")
	if err != nil {
		return err
	}

	options := print.PrintTableOptions{
		Theme:                *theme,
		OmitEmpty:            true,
		Output:               "table",
		SuppressEmptyColumns: false,
	}

	data := dao.TableOutput{
		Headers: []string{"project", "synced"},
		Rows:    []dao.Row{},
	}

	for _, project := range projects {
		projectPath, err := core.GetAbsolutePath(config.Path, project.Path, project.Name)
		if err != nil {
			return err
		}

		if _, err := os.Stat(projectPath); !os.IsNotExist(err) {
			// Project  synced
			data.Rows = append(data.Rows, dao.Row{Columns: []string{project.Name, text.FgGreen.Sprintf("\u2713")}})
		} else {
			// Project not synced
			data.Rows = append(data.Rows, dao.Row{Columns: []string{project.Name, text.FgRed.Sprintf("\u2715")}})
		}
	}

	print.PrintTable(data.Rows, options, data.Headers, []string{})

	return nil
}

func PrintProjectInit(projects []dao.Project) {
	theme := dao.Theme{
		Table: dao.DefaultTable,
	}

	options := print.PrintTableOptions{
		Theme:                theme,
		OmitEmpty:            true,
		Output:               "table",
		SuppressEmptyColumns: false,
	}

	data := dao.TableOutput{
		Headers: []string{"project", "path"},
		Rows:    []dao.Row{},
	}

	for _, project := range projects {
		data.Rows = append(data.Rows, dao.Row{Columns: []string{project.Name, project.Path}})
	}

	fmt.Println("\nFollowing projects were added to mani.yaml")
	print.PrintTable(data.Rows, options, data.Headers, []string{})
}
