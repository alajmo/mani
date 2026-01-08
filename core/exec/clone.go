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
	cmd := exec.Command("git", "remote", "add", remote.Name, remote.URL)
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
	cmd := exec.Command("git", "remote", "set-url", remote.Name, remote.URL)
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
	if project.URL == "" {
		return nil
	}

	// Remove remotes found in .git/config but not in RemoteList
	for name, foundURL := range foundRemotes {
		// Ignore origin remote (same as project url)
		if foundURL == project.URL {
			continue
		}

		// Check if this URL exists in project.RemoteList
		urlExists := false
		for _, remote := range project.RemoteList {
			if foundURL == remote.URL {
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

// CheckRemoteBranchExists verifies if a branch exists on the remote
func CheckRemoteBranchExists(repoPath string, branch string) (bool, error) {
	cmd := exec.Command("git", "ls-remote", "--heads", "origin", branch)
	cmd.Dir = repoPath
	output, err := cmd.Output()
	if err != nil {
		return false, err
	}
	return len(output) > 0, nil
}

// CreateWorktree creates a git worktree at the specified path for the given branch.
// If the branch doesn't exist, it creates a new branch.
func CreateWorktree(parentPath string, worktreePath string, branch string, createBranch bool) error {
	var cmd *exec.Cmd
	if createBranch {
		cmd = exec.Command("git", "worktree", "add", "-b", branch, worktreePath)
	} else {
		cmd = exec.Command("git", "worktree", "add", worktreePath, branch)
	}
	cmd.Dir = parentPath
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to create worktree: %s - %s", err, string(output))
	}
	return nil
}

// SyncWorktrees handles worktree creation for a project
func SyncWorktrees(config *dao.Config, project dao.Project) error {
	if len(project.WorktreeList) == 0 {
		return nil
	}

	parentPath, err := core.GetAbsolutePath(config.Dir, project.Path, project.Name)
	if err != nil {
		return err
	}

	// Parent must exist first
	if _, err := os.Stat(parentPath); os.IsNotExist(err) {
		return fmt.Errorf("parent project %s must be cloned before creating worktrees", project.Name)
	}

	for _, wt := range project.WorktreeList {
		// Skip worktrees without a branch
		if wt.Branch == "" {
			fmt.Printf("Warning: skipping worktree for project %s: branch is required\n", project.Name)
			continue
		}

		// Determine worktree path - relative to project path
		// If no path specified, default to <project-dir>/<branch-name>
		var wtPath string
		if wt.Path == "" {
			// Default: subdirectory inside project named after the branch
			// e.g., project at ./dashgrid -> worktree at ./dashgrid/<branch>
			branchName := strings.ReplaceAll(wt.Branch, "/", "-") // Replace slashes in branch names
			wtPath = filepath.Join(parentPath, branchName)
		} else if filepath.IsAbs(wt.Path) {
			wtPath = wt.Path
		} else {
			// Path is relative to project directory
			wtPath = filepath.Join(parentPath, wt.Path)
		}

		// Skip if worktree already exists
		if _, err := os.Stat(wtPath); !os.IsNotExist(err) {
			continue
		}

		// Check if branch exists on remote
		exists, err := CheckRemoteBranchExists(parentPath, wt.Branch)
		if err != nil {
			// If we can't check remote, try to create anyway
			exists = false
		}

		// Create the worktree (create branch if it doesn't exist)
		err = CreateWorktree(parentPath, wtPath, wt.Branch, !exists)
		if err != nil {
			fmt.Printf("Warning: failed to create worktree %s: %v\n", wtPath, err)
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

		if projects[i].URL == "" {
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
				cmdArr = []string{"clone", syncProjects[i].URL, projectPath}
			} else {
				cmdArr = []string{"clone", "--progress", syncProjects[i].URL, projectPath}
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

	// Sync worktrees for all projects that have them
	for i := range projects {
		if len(projects[i].WorktreeList) > 0 {
			err := SyncWorktrees(config, projects[i])
			if err != nil {
				// Log warning but don't fail the entire sync
				fmt.Printf("Warning: failed to sync worktrees for %s: %v\n", projects[i].Name, err)
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
			if project.URL == "" {
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
