package exec

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/alajmo/mani/core"
	"github.com/alajmo/mani/core/dao"
	"github.com/alajmo/mani/core/print"
	"github.com/gookit/color"
	"golang.org/x/term"
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
		return &core.FailedToCreateWorktree{Path: worktreePath, Err: err, Output: string(output)}
	}
	return nil
}

// GetWorktrees returns a map of existing worktrees (path -> branch)
func GetWorktrees(parentPath string) (map[string]string, error) {
	return core.GetWorktreeList(parentPath)
}

// RemoveWorktree removes a git worktree (keeps the branch)
func RemoveWorktree(parentPath string, worktreePath string) error {
	cmd := exec.Command("git", "worktree", "remove", worktreePath)
	cmd.Dir = parentPath
	output, err := cmd.CombinedOutput()
	if err != nil {
		return &core.FailedToRemoveWorktree{Path: worktreePath, Err: err, Output: string(output)}
	}
	return nil
}

// SyncWorktrees handles worktree creation and optionally removal for a project
func SyncWorktrees(config *dao.Config, project dao.Project, removeOrphans bool) error {
	parentPath, err := core.GetAbsolutePath(config.Path, project.Path, project.Name)
	if err != nil {
		return err
	}

	// Parent must exist first (skip if not cloned yet)
	if _, err := os.Stat(parentPath); os.IsNotExist(err) {
		return nil
	}

	// Prune stale worktree references (e.g. from manually deleted directories)
	pruneCmd := exec.Command("git", "worktree", "prune")
	pruneCmd.Dir = parentPath
	_ = pruneCmd.Run()

	// Build map of expected worktree paths from config
	expectedPaths := make(map[string]bool)
	for _, wt := range project.WorktreeList {
		var wtPath string
		if filepath.IsAbs(wt.Path) {
			wtPath = filepath.Clean(wt.Path)
		} else {
			wtPath = filepath.Join(parentPath, wt.Path)
		}
		expectedPaths[wtPath] = true

		// Create worktree if it doesn't exist
		if _, err := os.Stat(wtPath); os.IsNotExist(err) {
			// Try checking out existing branch first (local or remote-tracking)
			err = CreateWorktree(parentPath, wtPath, wt.Branch, false)
			if err != nil {
				// Branch doesn't exist anywhere — create it
				err = CreateWorktree(parentPath, wtPath, wt.Branch, true)
			}
			if err != nil {
				return err
			}
		}
	}

	// Remove worktrees not in config (only if enabled)
	if removeOrphans {
		existingWorktrees, err := GetWorktrees(parentPath)
		if err != nil {
			return err
		}

		for wtPath := range existingWorktrees {
			if !expectedPaths[wtPath] {
				err := RemoveWorktree(parentPath, wtPath)
				if err != nil {
					return err
				}
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

	// Sync worktrees: create if defined, remove orphans if enabled
	for i := range projects {
		if len(projects[i].WorktreeList) > 0 || *config.RemoveOrphanedWorktrees {
			err := SyncWorktrees(config, projects[i], *config.RemoveOrphanedWorktrees)
			if err != nil {
				return err
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
		var gitignoreEntries []string
		for _, project := range config.ProjectList {
			if project.URL == "" {
				continue
			}

			// Project must be below mani config file to be added to gitignore
			var projectPath string
			projectPath, err = core.GetAbsolutePath(config.Path, project.Path, project.Name)
			if err != nil {
				return err
			}

			// Skip the root project (it is the mani directory itself)
			if projectPath == config.Dir {
				continue
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
				gitignoreEntries = append(gitignoreEntries, relPath)
			} else {
				gitignoreEntries = append(gitignoreEntries, project.Name)
			}

			// Add worktrees to gitignore as well
			for _, wt := range project.WorktreeList {
				var wtAbsPath string
				if filepath.IsAbs(wt.Path) {
					wtAbsPath = filepath.Clean(wt.Path)
				} else {
					wtAbsPath = filepath.Join(projectPath, wt.Path)
				}

				// Worktree must be below mani config file to be added to gitignore
				if !strings.HasPrefix(wtAbsPath, config.Dir) {
					continue
				}

				wtRelPath, err := filepath.Rel(config.Dir, wtAbsPath)
				if err != nil {
					continue
				}
				gitignoreEntries = append(gitignoreEntries, wtRelPath)
			}
		}

		err := dao.UpdateProjectsToGitignore(gitignoreEntries, gitignoreFilename)
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
	if len(projects) == 0 {
		return
	}

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

// RemoveOrphanedProjects removes git repositories below the mani config directory
// that are no longer defined as projects or worktrees in the mani configuration.
// It asks for confirmation first, and removes nothing without a terminal to ask on.
func RemoveOrphanedProjects(config *dao.Config) error {
	orphanedPaths, err := findOrphanedProjects(config.Dir, config.ProjectList)
	if err != nil {
		return err
	}

	if len(orphanedPaths) == 0 {
		fmt.Println("No orphaned project directories found.")
		return nil
	}

	fmt.Printf("\n%s Found %d orphaned project %s:\n\n",
		color.FgYellow.Sprint("⚠"), len(orphanedPaths), pluralizeDirectory(len(orphanedPaths)))
	for _, relPath := range orphanedPaths {
		fmt.Printf("  %s %s\n", color.FgRed.Sprint("✗"), relPath)
	}

	fmt.Printf("\n%s These directories contain git repositories that are no longer defined in your mani.yaml configuration.\n",
		color.FgYellow.Sprint("⚠"))

	if !term.IsTerminal(int(os.Stdin.Fd())) {
		fmt.Println("Skipping removal: confirming it needs an interactive terminal.")
		return nil
	}

	fmt.Printf("%s This action will permanently delete these directories and all their contents.\n",
		color.FgRed.Sprint("!"))
	fmt.Print("\nAre you sure you want to delete these directories? [y/N]: ")
	response, err := bufio.NewReader(os.Stdin).ReadString('\n')
	if err != nil && !errors.Is(err, io.EOF) {
		return fmt.Errorf("failed to read user input: %w", err)
	}

	response = strings.TrimSpace(strings.ToLower(response))
	if response != "y" && response != "yes" {
		fmt.Println("Operation cancelled.")
		return nil
	}

	fmt.Println("\nRemoving orphaned project directories...")
	for _, relPath := range orphanedPaths {
		fmt.Printf("Removing %s... ", relPath)
		path := filepath.Join(config.Dir, relPath)
		err = os.RemoveAll(path)
		if err != nil {
			fmt.Printf("%s (error: %s)\n", color.FgRed.Sprint("failed"), err)
			return fmt.Errorf("failed to remove directory %s: %w", path, err)
		}
		fmt.Printf("%s\n", color.FgGreen.Sprint("done"))
	}

	fmt.Printf("\n%s Successfully removed %d orphaned project %s.\n",
		color.FgGreen.Sprint("✓"), len(orphanedPaths), pluralizeDirectory(len(orphanedPaths)))

	return nil
}

// findOrphanedProjects returns the git clones below configDir, relative to it,
// that neither are nor contain a project or worktree of the configuration.
//
// To find projects stored in sub-directories, it descends into directories that
// aren't git repositories, down to the depth of the deepest project. It never
// looks inside a repository (submodules, vendored checkouts) or a hidden
// directory, and only reports directories with a .git directory: a .git file
// marks a worktree or submodule, whose removal is left to git and to
// remove_orphaned_worktrees.
//
// Directories are compared by file identity rather than by path, so a project
// configured through a symlink or with different letter case (on
// case-insensitive file systems) is never mistaken for an orphan.
func findOrphanedProjects(configDir string, projects []dao.Project) ([]string, error) {
	root, err := filepath.EvalSymlinks(configDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read config directory: %w", err)
	}
	rootInfo, err := os.Stat(root)
	if err != nil {
		return nil, fmt.Errorf("failed to read config directory: %w", err)
	}

	// Keep every project and worktree directory below the config directory,
	// and every directory between it and the config directory
	var kept []os.FileInfo
	maxDepth := 0
	keep := func(path string) {
		var chain []os.FileInfo
		for dir, depth := filepath.Clean(path), 0; ; dir, depth = filepath.Dir(dir), depth+1 {
			info, err := os.Stat(dir)
			if err == nil {
				if os.SameFile(info, rootInfo) {
					kept = append(kept, chain...)
					maxDepth = max(maxDepth, depth)
					return
				}
				chain = append(chain, info)
			}
			if filepath.Dir(dir) == dir {
				return // Not below the config directory
			}
		}
	}

	for _, project := range projects {
		projectPath, err := core.GetAbsolutePath(configDir, project.Path, project.Name)
		if err != nil {
			return nil, err
		}
		keep(projectPath)

		for _, wt := range project.WorktreeList {
			if filepath.IsAbs(wt.Path) {
				keep(wt.Path)
			} else {
				keep(filepath.Join(projectPath, wt.Path))
			}
		}
	}

	if maxDepth == 0 {
		return nil, nil
	}

	isKept := func(info os.FileInfo) bool {
		for _, k := range kept {
			if os.SameFile(info, k) {
				return true
			}
		}
		return false
	}

	var orphanedPaths []string
	err = filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			if path == root {
				return fmt.Errorf("failed to read config directory: %w", err)
			}
			// Skip directories we can't read rather than aborting the sync
			if d != nil && d.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		// The config directory itself is never an orphan, even when it is a git repository
		if !d.IsDir() || path == root {
			return nil
		}

		if strings.HasPrefix(d.Name(), ".") {
			return filepath.SkipDir
		}

		relPath, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}

		gitInfo, err := os.Lstat(filepath.Join(path, ".git"))
		if err != nil {
			// Not a repository: look for projects inside it, down to the deepest project
			if strings.Count(relPath, string(filepath.Separator))+1 < maxDepth {
				return nil
			}
			return filepath.SkipDir
		}

		if gitInfo.IsDir() {
			info, err := d.Info()
			if err == nil && !isKept(info) {
				orphanedPaths = append(orphanedPaths, relPath)
			}
		}

		return filepath.SkipDir
	})
	if err != nil {
		return nil, err
	}

	return orphanedPaths, nil
}

func pluralizeDirectory(count int) string {
	if count == 1 {
		return "directory"
	}
	return "directories"
}
