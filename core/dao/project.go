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

	color "github.com/logrusorgru/aurora"
	"github.com/theckman/yacspin"

	"github.com/alajmo/mani/core"
)

type Project struct {
	Name        string   `yaml:"name"`
	Path        string   `yaml:"path"`
	Description string   `yaml:"description"`
	Url         string   `yaml:"url"`
	Clone       string   `yaml:"clone"`
	Tags        []string `yaml:"tags"`

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
	case "Description", "description":
		return p.Description
	case "Url", "url":
		return p.Url
	case "Tags", "tags":
		return strings.Join(p.Tags, ", ")
	}

	return ""
}

func CloneRepo(
	configPath string,
	project Project,
	serial bool,
	syncErrors map[string]string,
	wg *sync.WaitGroup,
) {
	defer wg.Done()
	projectPath, err := core.GetAbsolutePath(configPath, project.Path, project.Name)
	if err != nil {
		syncErrors[project.Name] = (&core.FailedToParsePath{Name: projectPath}).Error()
		return
	}

	if _, err := os.Stat(projectPath); os.IsNotExist(err) {
		if serial {
			fmt.Printf("\n%v\n\n", color.Bold(project.Name))
		}

		var cmd *exec.Cmd
		if project.Clone == "" {
			cmd = exec.Command("git", "clone", project.Url, projectPath)
		} else {
			cmd = exec.Command("sh", "-c", project.Clone)
		}
		cmd.Env = os.Environ()

		if serial {
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr

			err := cmd.Run()
			if err != nil {
				syncErrors[project.Name] = err.Error()
			} else {
				syncErrors[project.Name] = ""
			}
		} else {
			var errb bytes.Buffer
			cmd.Stderr = &errb

			err := cmd.Run()
			if err != nil {
				syncErrors[project.Name] = errb.String()
			} else {
				syncErrors[project.Name] = ""
			}
		}
	}

	return
}

func GetProjectRelPath(configPath string, path string) (string, error) {
	baseDir := filepath.Dir(configPath)
	relPath, err := filepath.Rel(baseDir, path)

	return relPath, err
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

	const maniComment = "# mani-projects #"
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

func (c Config) CloneRepos(serial bool) {
	urls := c.GetProjectUrls()
	if len(urls) == 0 {
		fmt.Println("No projects to sync")
		return
	}

	var cfg yacspin.Config
	cfg = yacspin.Config{
		Frequency:       100 * time.Millisecond,
		CharSet:         yacspin.CharSets[9],
		SuffixAutoColon: false,
		Message:         " Cloning",
	}

	spinner, err := yacspin.New(cfg)

	if !serial {
		err = spinner.Start()
		core.CheckIfError(err)
	}

	syncErrors := make(map[string]string)
	var wg sync.WaitGroup
	allProjectsSynced := true
	for _, project := range c.Projects {
		if project.Url != "" {
			wg.Add(1)

			if serial {
				CloneRepo(c.Path, project, serial, syncErrors, &wg)
				if syncErrors[project.Name] != "" {
					allProjectsSynced = false
					fmt.Println(syncErrors[project.Name])
				}
			} else {
				go CloneRepo(c.Path, project, serial, syncErrors, &wg)
			}
		}
	}

	wg.Wait()

	if !serial {
		err = spinner.Stop()
		core.CheckIfError(err)
	}

	if !serial {
		for _, project := range c.Projects {
			if syncErrors[project.Name] != "" {
				allProjectsSynced = false

				fmt.Printf("%v %v\n", color.Red("\u2715"), color.Bold(project.Name))
				fmt.Println(syncErrors[project.Name])
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

func (c Config) FilterProjects(
	cwdFlag bool,
	allProjectsFlag bool,
	projectPathsFlag []string,
	projectsFlag []string,
	tagsFlag []string,
) []Project {
	var finalProjects []Project
	if allProjectsFlag {
		finalProjects = c.Projects
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
			projects = c.GetProjects(projectsFlag)
		}

		var cwdProject Project
		if cwdFlag {
			cwdProject = c.GetCwdProject()
		}

		finalProjects = GetUnionProjects(projectPaths, tagProjects, projects, cwdProject)
	}

	return finalProjects
}

func (c Config) GetProjects(flagProjects []string) []Project {
	var matchedProjects []Project

	for _, v := range flagProjects {
		for _, p := range c.Projects {
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

		for _, pro := range c.Projects {
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
func (c Config) GetProjectDirs() []string {
	dirs := []string{}
	for _, project := range c.Projects {

		ps := strings.Split(filepath.Dir(project.RelPath), string(os.PathSeparator))
		for i := 1; i <= len(ps); i++ {
			p := filepath.Join(ps[0:i]...)

			if p != "." && !core.StringInSlice(p, dirs) {
				dirs = append(dirs, p)
			}
		}
	}

	return dirs
}

func (c Config) GetProjectsByName(names []string) []Project {
	if len(names) == 0 {
		return c.Projects
	}

	var filteredProjects []Project
	var foundProjectNames []string
	for _, name := range names {
		if core.StringInSlice(name, foundProjectNames) {
			continue
		}

		for _, project := range c.Projects {
			if name == project.Name {
				filteredProjects = append(filteredProjects, project)
				foundProjectNames = append(foundProjectNames, name)
			}
		}
	}

	return filteredProjects
}

// Projects must have all dirs to match. For instance, if --tags frontend,backend
// is passed, then a project must have both tags.
func (c Config) GetProjectsByPath(dirs []string) []Project {
	if len(dirs) == 0 {
		return c.Projects
	}

	var projects []Project
	for _, project := range c.Projects {

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
		return c.Projects
	}

	var projects []Project
	for _, project := range c.Projects {
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
	projectNames := []string{}
	for _, project := range c.Projects {
		projectNames = append(projectNames, project.Name)
	}

	return projectNames
}

func (c Config) GetProjectUrls() []string {
	urls := []string{}
	for _, project := range c.Projects {
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

	if d.Name != "" {
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
