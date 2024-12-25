package exec

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/alajmo/mani/core"
	"github.com/alajmo/mani/core/dao"
	"github.com/alajmo/mani/core/print"
	"github.com/gookit/color"
)

func getRemotes(project dao.Project) (map[string]string, error) {
	cmd := exec.Command("git", "remote", "-v")
	cmd.Dir = project.Path
	output, err := cmd.CombinedOutput()

	if err != nil {
		return nil, err
	}

	outputStr := string(output)
	lines := strings.Split(outputStr, "\n")

	remotes := make(map[string]string)
	for _, line := range lines {
		if line == "" {
			continue
		}
		parts := strings.Fields(line)
		if len(parts) < 3 {
			return nil, fmt.Errorf("unexpected line: %s", line)
		}
		remotes[parts[0]] = parts[1]
	}

	return remotes, nil
}

func addRemote(project dao.Project, remote dao.Remote) error {
	cmd := exec.Command("git", "remote", "add", remote.Name, remote.Url)
	cmd.Dir = project.Path
	_, err := cmd.CombinedOutput()

	if err != nil {
		return err
	}

	return nil
}

func removeRemote(project dao.Project, name string) error {
	cmd := exec.Command("git", "remote", "remove", name)
	cmd.Dir = project.Path
	_, err := cmd.CombinedOutput()
	if err != nil {
		return err
	}
	return nil
}

func updateRemote(project dao.Project, remote dao.Remote) error {
	cmd := exec.Command("git", "remote", "set-url", remote.Name, remote.Url)
	cmd.Dir = project.Path
	_, err := cmd.CombinedOutput()

	if err != nil {
		return err
	}

	return nil
}

func syncRemotes(project dao.Project) error {
	foundRemotes, err := getRemotes(project)
	if err != nil {
		return err
	}

	// Add remotes found in RemoteList but not in .git/config
	for _, remote := range project.RemoteList {
		_, found := foundRemotes[remote.Name]
		if found {
			err := updateRemote(project, remote)
			if err != nil {
				return err
			}
		} else {
			err := addRemote(project, remote)
			if err != nil {
				return err
			}
		}
	}

	// Don't remove remotes if project url is empty
	if project.Url == "" {
		return nil
	}

	// Remove remotes found in .git/config but not in RemoteList
	for name, foundUrl := range foundRemotes {
		// Ignore origin remote (same as project url)
		if foundUrl == project.Url {
			continue
		}

		// Check if this URL exists in project.RemoteList
		urlExists := false
		for _, remote := range project.RemoteList {
			if foundUrl == remote.Url {
				urlExists = true
				break
			}
		}

		// If URL is not in RemoteList, remove the remote
		if !urlExists {
			err := removeRemote(project, name)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func CloneRepos(config *dao.Config, projects []dao.Project, syncFlags core.SyncFlags) error {
	urls := config.GetProjectUrls()
	if len(urls) == 0 {
		fmt.Println("No projects to clone")
		return nil
	}

	var syncProjects []dao.Project
	for i := range projects {
		if !syncFlags.IgnoreSyncState && !projects[i].IsSync() {
			continue
		}

		if projects[i].Url == "" {
			continue
		}

		projectPath, err := core.GetAbsolutePath(config.Path, projects[i].Path, projects[i].Name)
		if err != nil {
			return err
		}

		// Project already synced
		if _, err := os.Stat(projectPath); !os.IsNotExist(err) {
			continue
		}

		syncProjects = append(syncProjects, projects[i])
	}

	var tasks []dao.Task
	for i := range syncProjects {
		var cmd string
		var cmdArr []string
		var shell string
		var shellProgram string

		if syncProjects[i].Clone != "" {
			shell = dao.DEFAULT_SHELL
			shellProgram = dao.DEFAULT_SHELL_PROGRAM
			cmdArr = []string{"-c", syncProjects[i].Clone}
			cmd = syncProjects[i].Clone
		} else {
			projectPath, err := core.GetAbsolutePath(config.Path, syncProjects[i].Path, syncProjects[i].Name)
			if err != nil {
				return err
			}

			shell = "git"
			shellProgram = "git"
			if syncFlags.Parallel {
				cmdArr = []string{"clone", syncProjects[i].Url, projectPath}
			} else {
				cmdArr = []string{"clone", "--progress", syncProjects[i].Url, projectPath}
			}

			if syncProjects[i].Branch != "" {
				cmdArr = append(cmdArr, "--branch", syncProjects[i].Branch)
			}

			if syncProjects[i].IsSingleBranch() {
				cmdArr = append(cmdArr, "--single-branch")
			}

			cmd = strings.Join(cmdArr, " ")
		}

		if len(syncProjects) > 0 {
			var task = dao.Task{
				Name: syncProjects[i].Name,

				Shell:        shell,
				Cmd:          cmd,
				ShellProgram: shellProgram,
				CmdArg:       cmdArr,
				SpecData: dao.Spec{
					Parallel:     syncFlags.Parallel,
					Forks:        syncFlags.Forks,
					IgnoreErrors: false,
				},

				ThemeData: dao.Theme{
					Color: core.Ptr(true),
					Stream: dao.Stream{
						Prefix:       syncFlags.Parallel, // we only use prefix when parallel is enabled since we need to see which project returns an error
						Header:       true,
						HeaderChar:   dao.DefaultStream.HeaderChar,
						HeaderPrefix: "Project",
						PrefixColors: dao.DefaultStream.PrefixColors,
					},
				},
			}

			tasks = append(tasks, task)
		}
	}

	if len(syncProjects) > 0 {
		target := Exec{Projects: syncProjects, Tasks: tasks, Config: *config}
		clientCh := make(chan Client, len(syncProjects))
		err := target.SetCloneClients(clientCh)
		if err != nil {
			return err
		}
		target.Text(false, os.Stdout, os.Stderr)
	}

	// User has opt-in to Sync remotes
	if *config.SyncRemotes {
		for i := range projects {
			// Project must have a Remote List defined
			if len(projects[i].RemoteList) > 0 {
				err := syncRemotes(projects[i])
				if err != nil {
					return err
				}
			}
		}
	}

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
	theme := dao.Theme{
		Color: core.Ptr(true),
		Table: dao.DefaultTable,
	}
	theme.Table.Border.Rows = core.Ptr(false)
	theme.Table.Header.Format = core.Ptr("t")

	options := print.PrintTableOptions{
		Theme:            theme,
		Output:           "table",
		Color:            *theme.Color,
		AutoWrap:         true,
		OmitEmptyRows:    false,
		OmitEmptyColumns: false,
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
			// Project synced
			data.Rows = append(data.Rows, dao.Row{Columns: []string{project.Name, color.FgGreen.Sprintf("\u2713")}})
		} else {
			// Project not synced
			data.Rows = append(data.Rows, dao.Row{Columns: []string{project.Name, color.FgRed.Sprintf("\u2715")}})
		}
	}

	fmt.Println()
	print.PrintTable(data.Rows, options, data.Headers, []string{}, os.Stdout)
	fmt.Println()

	return nil
}

func PrintProjectInit(projects []dao.Project) {
	theme := dao.Theme{
		Table: dao.DefaultTable,
		Color: core.Ptr(true),
	}
	theme.Table.Border.Rows = core.Ptr(false)
	theme.Table.Header.Format = core.Ptr("t")

	options := print.PrintTableOptions{
		Theme:            theme,
		Output:           "table",
		Color:            true,
		AutoWrap:         true,
		OmitEmptyRows:    true,
		OmitEmptyColumns: false,
	}

	data := dao.TableOutput{
		Headers: []string{"project", "path"},
		Rows:    []dao.Row{},
	}

	for _, project := range projects {
		data.Rows = append(data.Rows, dao.Row{Columns: []string{project.Name, project.Path}})
	}

	fmt.Println("\nFollowing projects were added to mani.yaml")
	fmt.Println()
	print.PrintTable(data.Rows, options, data.Headers, []string{}, os.Stdout)
}
