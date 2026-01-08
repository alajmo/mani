package dao

import (
	"bufio"
	"container/list"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strings"

	"gopkg.in/yaml.v3"

	"github.com/alajmo/mani/core"
)

type Project struct {
	Name         string   `yaml:"name"`
	Path         string   `yaml:"path"`
	Desc         string   `yaml:"desc"`
	URL          string   `yaml:"url"`
	Clone        string   `yaml:"clone"`
	Branch       string   `yaml:"branch"`
	SingleBranch *bool    `yaml:"single_branch"`
	Sync         *bool    `yaml:"sync"`
	Tags         []string `yaml:"tags"`
	EnvList      []string `yaml:"-"`
	RemoteList   []Remote `yaml:"-"`

	Env          yaml.Node   `yaml:"env"`
	Remotes      yaml.Node   `yaml:"remotes"`
	Worktrees    yaml.Node   `yaml:"worktrees"`
	WorktreeList []Worktree  `yaml:"-"`
	context      string
	contextLine  int
	RelPath      string
}

type Remote struct {
	Name string
	URL  string
}

type Worktree struct {
	Path   string `yaml:"path"`
	Branch string `yaml:"branch"`
}

func (p *Project) GetContext() string {
	return p.context
}

func (p *Project) GetContextLine() int {
	return p.contextLine
}

func (p Project) IsSingleBranch() bool {
	return p.SingleBranch != nil && *p.SingleBranch
}

func (p Project) IsSync() bool {
	return p.Sync == nil || *p.Sync
}

func (p Project) GetValue(key string, _ int) string {
	switch key {
	case "Project", "project":
		return p.Name
	case "Path", "path":
		return p.Path
	case "RelPath", "relpath":
		return p.RelPath
	case "Desc", "desc", "Description", "description":
		return p.Desc
	case "Url", "url":
		return p.URL
	case "Tag", "tag":
		return strings.Join(p.Tags, ", ")
	case "Worktrees", "worktrees":
		if len(p.WorktreeList) == 0 {
			return ""
		}
		return fmt.Sprintf("%d", len(p.WorktreeList))
	}

	return ""
}

func (c *Config) GetProjectList() ([]Project, []ResourceErrors[Project]) {
	var projects []Project
	count := len(c.Projects.Content)

	projectErrors := []ResourceErrors[Project]{}
	foundErrors := false
	for i := 0; i < count; i += 2 {
		project := &Project{
			context:     c.Path,
			contextLine: c.Projects.Content[i].Line,
		}

		err := c.Projects.Content[i+1].Decode(project)
		if err != nil {
			foundErrors = true
			projectError := ResourceErrors[Project]{Resource: project, Errors: core.StringsToErrors(err.(*yaml.TypeError).Errors)}
			projectErrors = append(projectErrors, projectError)
			continue
		}

		project.Name = c.Projects.Content[i].Value

		// Add absolute and relative path for each project
		project.Path, err = core.GetAbsolutePath(c.Dir, project.Path, project.Name)
		if err != nil {
			foundErrors = true
			projectError := ResourceErrors[Project]{Resource: project, Errors: core.StringsToErrors(err.(*yaml.TypeError).Errors)}
			projectErrors = append(projectErrors, projectError)
			continue
		}

		project.RelPath, err = core.GetRelativePath(c.Dir, project.Path)
		if err != nil {
			foundErrors = true
			projectError := ResourceErrors[Project]{Resource: project, Errors: core.StringsToErrors(err.(*yaml.TypeError).Errors)}
			projectErrors = append(projectErrors, projectError)
			continue
		}

		envList := []string{}
		projectEnvs, err := EvaluateEnv(ParseNodeEnv(project.Env))
		if err != nil {
			foundErrors = true
			projectError := ResourceErrors[Project]{Resource: project, Errors: core.StringsToErrors(err.(*yaml.TypeError).Errors)}
			projectErrors = append(projectErrors, projectError)
			continue
		}
		envList = append(envList, projectEnvs...)
		project.EnvList = envList

		projectRemotes := ParseRemotes(project.Remotes)
		project.RemoteList = projectRemotes

		projectWorktrees := ParseWorktrees(project.Worktrees)
		project.WorktreeList = projectWorktrees

		projects = append(projects, *project)
	}

	if foundErrors {
		return projects, projectErrors
	}

	return projects, nil
}

// GetFilteredProjects retrieves a filtered list of projects based on the provided ProjectFlags.
// It processes various filtering criteria and returns the matching projects.
//
// The function follows these steps:
// 1. If a target is specified, loads the target configuration, otherwise sets all to false
// 2. Merges any provided flag values with the target configuration
// 3. Applies all filtering criteria using FilterProjects
func (c Config) GetFilteredProjects(flags *core.ProjectFlags) ([]Project, error) {
	var err error
	var projects []Project

	target := &Target{}
	if flags.Target != "" {
		target, err = c.GetTarget(flags.Target)
		if err != nil {
			return []Project{}, err
		}
	}

	if len(flags.Projects) > 0 {
		target.Projects = flags.Projects
	}

	if len(flags.Paths) > 0 {
		target.Paths = flags.Paths
	}

	if len(flags.Tags) > 0 {
		target.Tags = flags.Tags
	}

	if flags.TagsExpr != "" {
		target.TagsExpr = flags.TagsExpr
	}

	if flags.Cwd {
		target.Cwd = flags.Cwd
	}

	if flags.All {
		target.All = flags.All
	}

	projects, err = c.FilterProjects(
		target.Cwd,
		target.All,
		target.Projects,
		target.Paths,
		target.Tags,
		target.TagsExpr,
	)
	if err != nil {
		return []Project{}, err
	}

	return projects, nil
}

// FilterProjects filters the project list based on various criteria. It supports filtering by:
// - All projects (allProjectsFlag)
// - Current working directory (cwdFlag)
// - Project names (projectsFlag)
// - Project paths (projectPathsFlag)
// - Project tags (tagsFlag)
// - Tag expressions (tagsExprFlag)
//
// Priority handling:
//   - If cwdFlag is true, the function immediately returns only the current working directory
//     project, ignoring all other filters.
//   - For all other combinations of filters, the function collects projects from each filter
//     into separate slices, then finds their intersection. If multiple
//     filters are specified, only projects that match ALL filters will be returned.
func (c Config) FilterProjects(
	cwdFlag bool,
	allProjectsFlag bool,
	projectsFlag []string,
	projectPathsFlag []string,
	tagsFlag []string,
	tagsExprFlag string,
) ([]Project, error) {
	var finalProjects []Project

	var err error
	var inputProjects [][]Project

	if cwdFlag {
		var cwdProjects []Project
		cwdProject, err := c.GetCwdProject()
		cwdProjects = append(cwdProjects, cwdProject)
		if err != nil {
			return []Project{}, err
		}

		return cwdProjects, nil
	}

	if allProjectsFlag {
		inputProjects = append(inputProjects, c.ProjectList)
	}

	if len(projectsFlag) > 0 {
		var projects []Project
		projects, err = c.GetProjectsByName(projectsFlag)
		if err != nil {
			return []Project{}, err
		}
		inputProjects = append(inputProjects, projects)
	}

	if len(projectPathsFlag) > 0 {
		var projectPaths []Project
		projectPaths, err = c.GetProjectsByPath(projectPathsFlag)
		if err != nil {
			return []Project{}, err
		}
		inputProjects = append(inputProjects, projectPaths)
	}

	if len(tagsFlag) > 0 {
		var tagProjects []Project
		tagProjects, err = c.GetProjectsByTags(tagsFlag)
		if err != nil {
			return []Project{}, err
		}
		inputProjects = append(inputProjects, tagProjects)
	}

	if tagsExprFlag != "" {
		var tagExprProjects []Project
		tagExprProjects, err = c.GetProjectsByTagsExpr(tagsExprFlag)
		if err != nil {
			return []Project{}, err
		}
		inputProjects = append(inputProjects, tagExprProjects)
	}

	finalProjects = c.GetIntersectProjects(inputProjects...)

	return finalProjects, nil
}

func (c Config) GetProject(name string) (*Project, error) {
	for _, project := range c.ProjectList {
		if name == project.Name {
			return &project, nil
		}
	}

	return nil, &core.ProjectNotFound{Name: []string{name}}
}

func (c Config) GetProjectsByName(projectNames []string) ([]Project, error) {
	var matchedProjects []Project

	foundProjectNames := make(map[string]bool)
	for _, p := range projectNames {
		foundProjectNames[p] = false
	}

	for _, v := range projectNames {
		for _, p := range c.ProjectList {
			if v == p.Name {
				foundProjectNames[p.Name] = true
				matchedProjects = append(matchedProjects, p)
			}
		}
	}

	nonExistingProjects := []string{}
	for k, v := range foundProjectNames {
		if !v {
			nonExistingProjects = append(nonExistingProjects, k)
		}
	}

	if len(nonExistingProjects) > 0 {
		return []Project{}, &core.ProjectNotFound{Name: nonExistingProjects}
	}

	return matchedProjects, nil
}

// GetProjectsByPath Projects must have all dirs to match.
// If user provides a path which does not exist, then return error containing
// all the paths it didn't find.
// Supports glob patterns:
// - '*' matches any sequence of non-separator characters
// - '**' matches any sequence of characters including separators
func (c Config) GetProjectsByPath(dirs []string) ([]Project, error) {
	if len(dirs) == 0 {
		return c.ProjectList, nil
	}

	foundDirs := make(map[string]bool)
	for _, dir := range dirs {
		foundDirs[dir] = false
	}

	projects := []Project{}
	for _, project := range c.ProjectList {
		// Variable use to check that all dirs are matched
		var numMatched = 0
		for _, dir := range dirs {

			matchPath := func(dir string, path string) (bool, error) {
				// Handle glob pattern
				if strings.Contains(dir, "*") {
					// Handle '**' glob pattern
					if strings.Contains(dir, "**") {
						// Convert the glob pattern to a regex pattern
						regexPattern := strings.ReplaceAll(dir, "**/", "<glob>")
						regexPattern = strings.ReplaceAll(regexPattern, "*", "[^/]*")
						regexPattern = strings.ReplaceAll(regexPattern, "?", ".")
						regexPattern = strings.ReplaceAll(regexPattern, "<glob>", "(.*/)*")
						regexPattern = "^" + regexPattern + "$"

						matched, err := regexp.MatchString(regexPattern, path)

						if err != nil {
							return false, err
						}

						if matched {
							return true, nil
						}
					}

					// Handle standard glob pattern
					matched, err := filepath.Match(dir, path)

					if err != nil {
						return false, err
					}

					if matched {
						return true, nil
					}
				}

				// Try matching as a partial path
				if strings.Contains(path, dir) {
					return true, nil
				}

				return false, nil
			}

			matched, err := matchPath(dir, project.RelPath)

			if err != nil {
				return []Project{}, err
			}

			if matched {
				foundDirs[dir] = true
				numMatched++
			}
		}

		if numMatched == len(dirs) {
			projects = append(projects, project)
		}
	}

	nonExistingDirs := []string{}
	for k, v := range foundDirs {
		if !v {
			nonExistingDirs = append(nonExistingDirs, k)
		}
	}

	if len(nonExistingDirs) > 0 {
		return []Project{}, &core.DirNotFound{Dirs: nonExistingDirs}
	}

	return projects, nil
}

// GetProjectsByTags Projects must have all tags to match. For instance, if --tags frontend,backend
// is passed, then a project must have both tags.
// We only return error if the flags provided do not exist in the mani config.
func (c Config) GetProjectsByTags(tags []string) ([]Project, error) {
	if len(tags) == 0 {
		return c.ProjectList, nil
	}

	foundTags := make(map[string]bool)
	for _, tag := range tags {
		foundTags[tag] = false
	}

	// Find projects matching the flag
	var projects []Project
	for _, project := range c.ProjectList {
		// Variable use to check that all tags are matched
		var numMatched = 0
		for _, tag := range tags {
			for _, projectTag := range project.Tags {
				if projectTag == tag {
					foundTags[tag] = true
					numMatched = numMatched + 1
				}
			}
		}

		if numMatched == len(tags) {
			projects = append(projects, project)
		}
	}

	nonExistingTags := []string{}
	for k, v := range foundTags {
		if !v {
			nonExistingTags = append(nonExistingTags, k)
		}
	}

	if len(nonExistingTags) > 0 {
		return []Project{}, &core.TagNotFound{Tags: nonExistingTags}
	}

	return projects, nil
}

// GetProjectsByTagsExpr Projects must have all tags to match. For instance, if --tags frontend,backend
// is passed, then a project must have both tags.
// We only return error if the tags provided do not exist.
func (c Config) GetProjectsByTagsExpr(tagsExpr string) ([]Project, error) {
	if tagsExpr == "" {
		return c.ProjectList, nil
	}

	var projects []Project
	for _, project := range c.ProjectList {
		matches, err := evaluateExpression(&project, tagsExpr)
		if err != nil {
			return c.ProjectList, &core.TagExprInvalid{Expression: err.Error()}
		}
		if matches {
			projects = append(projects, project)
		}
	}

	return projects, nil
}

func (c Config) GetCwdProject() (Project, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return Project{}, err
	}

	var project Project
	parts := strings.Split(cwd, string(os.PathSeparator))

out:
	for i := len(parts) - 1; i >= 0; i-- {
		p := strings.Join(parts[0:i+1], string(os.PathSeparator))

		for _, pro := range c.ProjectList {
			if p == pro.Path {
				project = pro
				break out
			}
		}
	}

	return project, nil
}

/**
 * GetProjectPaths For each project path, get all the enumerations of dirnames.
 * Example:
 * Input:
 *   - /frontend/tools/project-a
 *   - /frontend/tools/project-b
 *   - /frontend/tools/node/project-c
 *   - /backend/project-d
 * Output:
 *   - /frontend
 *   - /frontend/tools
 *   - /frontend/tools/node
 *   - /backend
 */
func (c Config) GetProjectPaths() []string {
	dirs := []string{}
	for _, project := range c.ProjectList {
		// Ignore projects outside of mani.yaml directory
		if strings.Contains(project.Path, c.Dir) {
			ps := strings.Split(filepath.Dir(project.RelPath), string(os.PathSeparator))
			for i := 1; i <= len(ps); i++ {
				p := filepath.Join(ps[0:i]...)

				if p != "." && !slices.Contains(dirs, p) {
					dirs = append(dirs, p)
				}
			}
		}
	}

	return dirs
}

func (c Config) GetProjectNames() []string {
	names := []string{}
	for _, project := range c.ProjectList {
		names = append(names, project.Name)
	}

	return names
}

func (c Config) GetProjectUrls() []string {
	urls := []string{}
	for _, project := range c.ProjectList {
		if project.URL != "" {
			urls = append(urls, project.URL)
		}
	}

	return urls
}

func (c Config) GetProjectsTree(dirs []string, tags []string) ([]TreeNode, error) {
	dirProjects, err := c.GetProjectsByPath(dirs)
	if err != nil {
		return []TreeNode{}, err
	}

	tagProjects, err := c.GetProjectsByTags(tags)
	if err != nil {
		return []TreeNode{}, err
	}

	projects := c.GetIntersectProjects(dirProjects, tagProjects)

	var projectPaths = []TNode{}
	for _, p := range projects {
		node := TNode{Name: p.Name, Path: p.RelPath}
		projectPaths = append(projectPaths, node)
	}

	var tree []TreeNode
	for i := range projectPaths {
		tree = AddToTree(tree, projectPaths[i])
	}

	return tree, nil
}

// IsGitWorktree checks if the given path is a git worktree (not the main repo).
// Returns: (isWorktree bool, parentRepoPath string, branch string, err error)
//
// A worktree's .git is a FILE (not directory) containing:
// "gitdir: /path/to/main-repo/.git/worktrees/worktree-name"
func IsGitWorktree(path string) (bool, string, string, error) {
	gitPath := filepath.Join(path, ".git")
	info, err := os.Stat(gitPath)
	if err != nil {
		return false, "", "", err
	}

	// If .git is a directory, it's a regular git repo (or the main worktree)
	if info.IsDir() {
		return false, "", "", nil
	}

	// .git is a file - read its content
	content, err := os.ReadFile(gitPath)
	if err != nil {
		return false, "", "", err
	}

	// Parse "gitdir: <path>"
	contentStr := strings.TrimSpace(string(content))
	if !strings.HasPrefix(contentStr, "gitdir: ") {
		return false, "", "", nil
	}

	gitDir := strings.TrimPrefix(contentStr, "gitdir: ")

	// Make gitDir absolute if it's relative
	if !filepath.IsAbs(gitDir) {
		gitDir = filepath.Join(path, gitDir)
		gitDir = filepath.Clean(gitDir)
	}

	// Extract parent repo path from gitdir
	// Pattern: <parent-repo>/.git/worktrees/<name>
	if strings.Contains(gitDir, string(filepath.Separator)+".git"+string(filepath.Separator)+"worktrees"+string(filepath.Separator)) {
		// Find the .git directory position
		idx := strings.Index(gitDir, string(filepath.Separator)+".git"+string(filepath.Separator)+"worktrees"+string(filepath.Separator))
		if idx >= 0 {
			parentRepoPath := gitDir[:idx]

			// Get current branch using git command
			branch, _ := core.GetWorktreeBranch(path)

			return true, parentRepoPath, branch, nil
		}
	}

	return false, "", "", nil
}

func FindVCSystems(rootPath string) ([]Project, error) {
	projects := []Project{}
	// Map to track worktrees: parentAbsPath -> []Worktree
	worktreeMap := make(map[string][]Worktree)
	// Map to track project absolute paths for matching with worktree parents
	projectAbsPaths := make(map[string]*Project)

	err := filepath.Walk(rootPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Is file
		if !info.IsDir() {
			return nil
		}

		if path == rootPath {
			return nil
		}

		// Is Directory and Has a Git Dir inside, add to projects and SkipDir
		gitDir := filepath.Join(path, ".git")
		if _, err := os.Stat(gitDir); !os.IsNotExist(err) {
			name := filepath.Base(path)
			relPath, _ := filepath.Rel(rootPath, path)

			// Check if this is a worktree
			isWorktree, parentRepoPath, branch, _ := IsGitWorktree(path)

			if isWorktree {
				// This is a worktree - add to worktree map
				wt := Worktree{
					Path:   relPath,
					Branch: branch,
				}
				worktreeMap[parentRepoPath] = append(worktreeMap[parentRepoPath], wt)
			} else {
				// This is a regular repository
				var project Project
				url, rErr := core.GetRemoteURL(path)
				if rErr != nil {
					project = Project{Name: name, Path: relPath}
				} else {
					project = Project{Name: name, Path: relPath, URL: url}
				}

				projects = append(projects, project)
				// Store reference by absolute path for worktree matching
				projectAbsPaths[path] = &projects[len(projects)-1]
			}

			return filepath.SkipDir
		}

		return nil
	})

	if err != nil {
		return projects, err
	}

	// Attach worktrees to their parent projects
	for parentPath, worktrees := range worktreeMap {
		if project, ok := projectAbsPaths[parentPath]; ok {
			project.WorktreeList = worktrees
		}
	}

	return projects, nil
}

func UpdateProjectsToGitignore(projectNames []string, gitignoreFilename string) (err error) {
	l := list.New()
	gitignoreFile, err := os.OpenFile(gitignoreFilename, os.O_RDWR, 0644)
	if err != nil {
		return &core.FailedToOpenFile{Name: gitignoreFilename}
	}
	defer func() {
		closeErr := gitignoreFile.Close()
		if err == nil {
			err = closeErr
		}
	}()

	scanner := bufio.NewScanner(gitignoreFile)
	for scanner.Scan() {
		line := scanner.Text()
		l.PushBack(line)
	}

	const maniComment = "# mani #"
	var insideComment = false
	var beginElement *list.Element
	var endElement *list.Element
	var next *list.Element

	// Remove all projects inside # mani #
	for e := l.Front(); e != nil; e = next {
		next = e.Next()

		if e.Value == maniComment && !insideComment {
			insideComment = true
			beginElement = e
			continue
		}

		if e.Value == maniComment {
			endElement = e
			break
		}

		if insideComment {
			l.Remove(e)
		}
	}

	// If missing start # mani #
	if beginElement == nil {
		l.PushBack(maniComment)
		beginElement = l.Back()
	}

	// If missing ending # mani #
	if endElement == nil {
		l.PushBack(maniComment)
	}

	// Insert projects within # mani # section
	for _, projectName := range projectNames {
		l.InsertAfter(projectName, beginElement)
	}

	err = gitignoreFile.Truncate(0)
	if err != nil {
		return err
	}

	_, err = gitignoreFile.Seek(0, 0)
	if err != nil {
		return err
	}

	// Write to gitignore file
	for e := l.Front(); e != nil; e = e.Next() {
		str := fmt.Sprint(e.Value)
		_, err = gitignoreFile.WriteString(str)
		if err != nil {
			return err
		}

		_, err = gitignoreFile.WriteString("\n")
		if err != nil {
			return err
		}
	}

	return nil
}

// ParseRemotes List of remotes (key: value)
func ParseRemotes(node yaml.Node) []Remote {
	var remotes []Remote
	count := len(node.Content)

	for i := 0; i < count; i += 2 {
		remote := Remote{
			Name: node.Content[i].Value,
			URL:  node.Content[i+1].Value,
		}

		remotes = append(remotes, remote)
	}

	return remotes
}

// ParseWorktrees parses worktree definitions from YAML
func ParseWorktrees(node yaml.Node) []Worktree {
	var worktrees []Worktree

	for _, content := range node.Content {
		var wt Worktree
		if err := content.Decode(&wt); err == nil {
			worktrees = append(worktrees, wt)
		}
	}

	return worktrees
}

func (c Config) GetIntersectProjects(ps ...[]Project) []Project {
	counts := make(map[string]int, len(c.ProjectList))
	for _, projects := range ps {
		for _, project := range projects {
			counts[project.Name] += 1
		}
	}

	var projects []Project
	for _, p := range c.ProjectList {
		if counts[p.Name] == len(ps) && len(ps) > 0 {
			projects = append(projects, p)
		}
	}

	return projects
}

// TREE

type TNode struct {
	Name string
	Path string
}

type TreeNode struct {
	Path        string
	ProjectName string
	Children    []TreeNode
}

// AddToTree recursively builds a tree structure from path components
// root: The current level of tree nodes
// node: Node containing path and name information to be added
func AddToTree(root []TreeNode, node TNode) []TreeNode {
	// Return if path is empty or starts with separator
	items := strings.Split(node.Path, string(os.PathSeparator))
	if len(items) == 0 || items[0] == "" {
		return root
	}

	if len(items) > 0 {
		var i int
		// Search for existing node with same path at current level
		for i = 0; i < len(root); i++ {
			if root[i].Path == items[0] { // already in tree
				break
			}
		}

		// If node doesn't exist at current level, create new node
		if i == len(root) {
			root = append(root, TreeNode{
				Path:        items[0],
				ProjectName: "",
				Children:    []TreeNode{},
			})
		}

		// If this is the last component in the path (leaf node/file)
		if len(items) == 1 {
			root[i].ProjectName = node.Name // Set name for projects only
		} else {
			root[i].ProjectName = ""
			str := strings.Join(items[1:], string(os.PathSeparator))
			n := TNode{Name: node.Name, Path: str}
			root[i].Children = AddToTree(root[i].Children, n)
		}
	}

	return root
}
