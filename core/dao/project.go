package dao

import (
	"bufio"
	"bytes"
	"container/list"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"gopkg.in/yaml.v3"
	color "github.com/logrusorgru/aurora"
	"github.com/theckman/yacspin"

	"github.com/alajmo/mani/core"
)

type Project struct {
	Name  string   `yaml:"name"`
	Path  string   `yaml:"path"`
	Desc  string   `yaml:"desc"`
	Url   string   `yaml:"url"`
	Clone string   `yaml:"clone"`
	Tags  []string `yaml:"tags"`
	EnvList  []string

	Env   yaml.Node `yaml:"env"`
	Context string
	RelPath string
}

func (p Project) GetValue(key string) string {
	switch key {
	case "Name", "name":
		return p.Name
	case "Path", "path":
		return p.Path
	case "RelPath", "relpath":
		return p.RelPath
	case "Desc", "desc", "Description", "description":
		return p.Desc
	case "Url", "url":
		return p.Url
	case "Tags", "tags":
		return strings.Join(p.Tags, ", ")
	}

	return ""
}

func (c *Config) GetProjectList() ([]Project, error) {
	var projects []Project
	count := len(c.Projects.Content)

	var err error
	for i := 0; i < count; i += 2 {
		project := &Project{}
		err = c.Projects.Content[i+1].Decode(project)
		if err != nil {
			return []Project{}, &core.FailedToParseFile{Name: c.Path, Msg: err}
		}

		project.Name = c.Projects.Content[i].Value

		// Add absolute and relative path for each project
		project.Path, err = core.GetAbsolutePath(c.Dir, project.Path, project.Name)
		if err != nil {
			return []Project{}, err
		}

		project.RelPath, err = core.GetRelativePath(c.Dir, project.Path)
		if err != nil {
			return []Project{}, err
		}

		envList, err := core.EvaluateEnv(core.GetEnv(project.Env))
		if err != nil {
			return []Project{}, &core.FailedToParseFile{Name: c.Path, Msg: err}
		}

		project.EnvList = envList

		project.Context = c.Path

		projects = append(projects, *project)
	}

	return projects, nil
}

func (c Config) CloneRepos(parallel bool) {
	// TODO: Refactor
	urls := c.GetProjectUrls()
	if len(urls) == 0 {
		fmt.Println("No projects to sync")
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
			value, found := syncErrors.Load(project.Name)
			if found {
				allProjectsSynced = false

				fmt.Printf("%v %v\n", color.Red("\u2715"), color.Bold(project.Name))
				fmt.Println(value)
			} else {
				fmt.Printf("%v %v\n", color.Green("\u2713"), color.Bold(project.Name))
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
			fmt.Printf("\n%v\n\n", color.Bold(project.Name))
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
			url := core.GetRemoteUrl(path)
			project := Project{Name: name, Path: relPath, Url: url}
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

	if beginElement == nil {
		l.PushBack(maniComment)
		beginElement = l.Back()
	}

	if endElement == nil {
		l.PushBack(maniComment)
	}

	for _, projectName := range projectNames {
		l.InsertAfter(projectName, beginElement)
	}

	err = gitignoreFile.Truncate(0)
	core.CheckIfError(err)

	_, err = gitignoreFile.Seek(0, 0)
	core.CheckIfError(err)

	for e := l.Front(); e != nil; e = e.Next() {
		str := fmt.Sprint(e.Value)
		_, err = gitignoreFile.WriteString(str)
		core.CheckIfError(err)

		_, err = gitignoreFile.WriteString("\n")
		core.CheckIfError(err)
	}

	gitignoreFile.Close()

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
	projectPathsFlag []string,
	projectsFlag []string,
	tagsFlag []string,
) []Project {
	var finalProjects []Project
	if allProjectsFlag {
		finalProjects = c.ProjectList
	} else {
		var projectPaths []Project
		if len(projectPathsFlag) > 0 {
			projectPaths = c.GetProjectsByPath(projectPathsFlag)
		}

		var tagProjects []Project
		if len(tagsFlag) > 0 {
			tagProjects = c.GetProjectsByTags(tagsFlag)
		}

		var projects []Project
		if len(projectsFlag) > 0 {
			projects = c.GetProjectsByName(projectsFlag)
		}

		var cwdProject Project
		if cwdFlag {
			cwdProject = c.GetCwdProject()
		}

		finalProjects = GetUnionProjects(projectPaths, tagProjects, projects, cwdProject)
	}

	return finalProjects
}

func (c Config) GetProject(name string) (*Project, error) {
	for _, project := range c.ProjectList {
		if name == project.Name {
			return &project, nil
		}
	}

	return nil, &core.ProjectNotFound{Name: name}
}

func (c Config) GetProjectsByName(projectNames []string) []Project {
	var matchedProjects []Project

	for _, v := range projectNames {
		for _, p := range c.ProjectList {
			if v == p.Name {
				matchedProjects = append(matchedProjects, p)
			}
		}
	}

	return matchedProjects
}

func (c Config) GetCwdProject() Project {
	cwd, err := os.Getwd()
	core.CheckIfError(err)

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

	return project
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

// Projects must have all dirs to match. For instance, if --tags frontend,backend
// is passed, then a project must have both tags.
func (c Config) GetProjectsByPath(dirs []string) []Project {
	if len(dirs) == 0 {
		return c.ProjectList
	}

	var projects []Project
	for _, project := range c.ProjectList {

		// Variable use to check that all dirs are matched
		var numMatched int = 0
		for _, dir := range dirs {
			if strings.Contains(project.RelPath, dir) {
				numMatched = numMatched + 1
			}
		}

		if numMatched == len(dirs) {
			projects = append(projects, project)
		}
	}

	return projects
}

// Projects must have all tags to match. For instance, if --tags frontend,backend
// is passed, then a project must have both tags.
func (c Config) GetProjectsByTags(tags []string) []Project {
	if len(tags) == 0 {
		return c.ProjectList
	}

	var projects []Project
	for _, project := range c.ProjectList {
		// Variable use to check that all tags are matched
		var numMatched int = 0
		for _, tag := range tags {
			for _, projectTag := range project.Tags {
				if projectTag == tag {
					numMatched = numMatched + 1
				}
			}
		}

		if numMatched == len(tags) {
			projects = append(projects, project)
		}
	}

	return projects
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

func (c Config) GetProjectsTree(dirs []string, tags []string) []core.TreeNode {
	var tree []core.TreeNode
	var projectPaths = []string{}

	dirProjects := c.GetProjectsByPath(dirs)
	tagProjects := c.GetProjectsByTags(tags)
	projects := GetIntersectProjects(dirProjects, tagProjects)

	for _, p := range projects {
		if p.RelPath != "." {
			projectPaths = append(projectPaths, p.RelPath)
		}
	}

	for i := range projectPaths {
		tree = core.AddToTree(tree, strings.Split(projectPaths[i], string(os.PathSeparator)))
	}

	return tree
}

func GetUnionProjects(a []Project, b []Project, c []Project, d Project) []Project {
	prjs := []Project{}

	for _, project := range a {
		if !ProjectInSlice(project.Name, prjs) {
			prjs = append(prjs, project)
		}
	}

	for _, project := range b {
		if !ProjectInSlice(project.Name, prjs) {
			prjs = append(prjs, project)
		}
	}

	for _, project := range c {
		if !ProjectInSlice(project.Name, prjs) {
			prjs = append(prjs, project)
		}
	}

	if d.Name != "" && !ProjectInSlice(d.Name, prjs) {
		prjs = append(prjs, d)
	}

	projects := []Project{}
	projects = append(projects, prjs...)

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
