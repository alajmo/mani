package dao

import (
	"bufio"
	"container/list"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"

	"github.com/alajmo/mani/core"
)

type Project struct {
	Name    string   `yaml:"name"`
	Path    string   `yaml:"path"`
	Desc    string   `yaml:"desc"`
	Url     string   `yaml:"url"`
	Clone   string   `yaml:"clone"`
	Tags    []string `yaml:"tags"`
	Sync    *bool    `yaml:"sync"`
	EnvList []string `yaml:"-"`

	Env         yaml.Node `yaml:"env"`
	context     string
	contextLine int
	RelPath     string
}

func (p *Project) GetContext() string {
	return p.context
}

func (p *Project) GetContextLine() int {
	return p.contextLine
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
		return p.Url
	case "Tag", "tag":
		return strings.Join(p.Tags, ", ")
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

		envList := []string{
			fmt.Sprintf("MANI_PROJECT_PATH=%s", project.Path),
			fmt.Sprintf("MANI_PROJECT_URL=%s", project.Url),
		}

		projectEnvs, err := EvaluateEnv(ParseNodeEnv(project.Env))
		if err != nil {
			foundErrors = true
			projectError := ResourceErrors[Project]{Resource: project, Errors: core.StringsToErrors(err.(*yaml.TypeError).Errors)}
			projectErrors = append(projectErrors, projectError)
			continue
		}

		envList = append(envList, projectEnvs...)

		project.EnvList = envList

		projects = append(projects, *project)
	}

	if foundErrors {
		return projects, projectErrors
	}

	return projects, nil
}

func FindVCSystems(rootPath string) ([]Project, error) {
	projects := []Project{}
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

			var project Project
			url, rErr := core.GetRemoteUrl(path)
			if rErr != nil {
				project = Project{Name: name, Path: relPath}
			} else {
				project = Project{Name: name, Path: relPath, Url: url}
			}

			projects = append(projects, project)

			return filepath.SkipDir
		}

		return nil
	})

	return projects, err
}

func UpdateProjectsToGitignore(projectNames []string, gitignoreFilename string) error {
	l := list.New()
	gitignoreFile, err := os.OpenFile(gitignoreFilename, os.O_RDWR, 0644)
	if err != nil {
		return &core.FailedToOpenFile{Name: gitignoreFilename}
	}
	defer gitignoreFile.Close()

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

func ProjectInSlice(name string, list []Project) bool {
	for _, p := range list {
		if p.Name == name {
			return true
		}
	}
	return false
}

func (c Config) FilterProjects(
	cwdFlag bool,
	allProjectsFlag bool,
	projectsFlag []string,
	projectPathsFlag []string,
	tagsFlag []string,
) ([]Project, error) {
	var finalProjects []Project
	if allProjectsFlag {
		finalProjects = c.ProjectList
	} else {
		var err error

		var projectPaths []Project
		if len(projectPathsFlag) > 0 {
			projectPaths, err = c.GetProjectsByPath(projectPathsFlag)
			if err != nil {
				return []Project{}, err
			}
		}

		var projects []Project
		if len(projectsFlag) > 0 {
			projects, err = c.GetProjectsByName(projectsFlag)
			if err != nil {
				return []Project{}, err
			}
		}

		var tagProjects []Project
		if len(tagsFlag) > 0 {
			tagProjects, err = c.GetProjectsByTags(tagsFlag)
			if err != nil {
				return []Project{}, err
			}
		}

		var cwdProjects []Project
		if cwdFlag {
			cwdProject, err := c.GetCwdProject()
			cwdProjects = append(cwdProjects, cwdProject)
			if err != nil {
				return []Project{}, err
			}
		}

		finalProjects = GetUnionProjects(cwdProjects, projectPaths, tagProjects, projects)
	}

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

// Projects must have all dirs to match.
// If user provides a path which does not exist, then return error containing
// all the paths it didn't find.
func (c Config) GetProjectsByPath(dirs []string) ([]Project, error) {
	if len(dirs) == 0 {
		return c.ProjectList, nil
	}

	foundDirs := make(map[string]bool)
	for _, dir := range dirs {
		foundDirs[dir] = false
	}

	var projects []Project
	for _, project := range c.ProjectList {
		// Variable use to check that all dirs are matched
		var numMatched int = 0
		for _, dir := range dirs {
			if strings.Contains(project.RelPath, dir) {
				foundDirs[dir] = true
				numMatched = numMatched + 1
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

// Projects must have all tags to match. For instance, if --tags frontend,backend
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
		var numMatched int = 0
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
 * For each project path, get all the enumerations of dirnames.
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

				if p != "." && !core.StringInSlice(p, dirs) {
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
		if project.Url != "" {
			urls = append(urls, project.Url)
		}
	}

	return urls
}

func (c Config) GetProjectsTree(dirs []string, tags []string) ([]TreeNode, error) {
	var tree []TreeNode
	var projectPaths = []string{}

	dirProjects, err := c.GetProjectsByPath(dirs)
	if err != nil {
		return []TreeNode{}, err
	}

	tagProjects, err := c.GetProjectsByTags(tags)
	if err != nil {
		return []TreeNode{}, err
	}

	projects := GetIntersectProjects(dirProjects, tagProjects)

	for _, p := range projects {
		if p.RelPath != "." {
			projectPaths = append(projectPaths, p.RelPath)
		}
	}

	for i := range projectPaths {
		tree = AddToTree(tree, strings.Split(projectPaths[i], string(os.PathSeparator)))
	}

	return tree, nil
}

func GetUnionProjects(ps ...[]Project) []Project {
	projects := []Project{}
	for _, part := range ps {
		for _, project := range part {
			if !ProjectInSlice(project.Name, projects) {
				projects = append(projects, project)
			}
		}
	}

	return projects
}

func GetIntersectProjects(a []Project, b []Project) []Project {
	projects := []Project{}

	for _, pa := range a {
		for _, pb := range b {
			if pa.Name == pb.Name {
				projects = append(projects, pa)
			}
		}
	}

	return projects
}
