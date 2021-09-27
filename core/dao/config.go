package dao

import (
	"bufio"
	"container/list"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"gopkg.in/yaml.v3"
	"github.com/theckman/yacspin"
	color "github.com/logrusorgru/aurora"

	"github.com/alajmo/mani/core"
)

var (
	Version = "dev"
	DEFAULT_SHELL = "sh -c"
	ACCEPTABLE_FILE_NAMES = []string{"mani.yaml", "mani.yml", ".mani", ".mani.yaml", ".mani.yml", "Manifile", "Manifile.yaml", "Manifile.yml"}
)

type Config struct {
	Path string

	Import      []string	`yaml:"import"`
	EnvList     []string
	NetworkList []Network
	Shell       string	`yaml:"shell"`
	Projects    []Project	`yaml:"projects"`
	Dirs	    []Dir	`yaml:"dirs"`
	Tasks	    []Task	`yaml:"tasks"`

	Env        yaml.Node    `yaml:"env"`
	Networks   yaml.Node    `yaml:"networks"`

	Theme struct {
		Table string	`yaml:"table"`
		Tree string	`yaml:"tree"`
	}
}

func (c Config) GetEnv() []string {
	var envs []string
	count := len(c.Env.Content)
	for i := 0; i < count; i += 2 {
		env := fmt.Sprintf("%v=%v", c.Env.Content[i].Value, c.Env.Content[i + 1].Value)
		envs = append(envs, env)
	}

	return envs
}

// TODO: Remove since it's not used
func (c *Config) SetEnvList(envList []string) {
	c.EnvList = envList
}

func ReadConfig(cfgName string) (Config, error) {
	var configPath string

	// Try to find config file in current directory and all parents
	if cfgName != "" {
		filename, err := filepath.Abs(cfgName)
		if err != nil {
			return Config{}, err
		}

		configPath = filename
	} else {
		wd, err := os.Getwd()
		if err != nil {
			return Config{}, err
		}

		filename, err := core.FindFileInParentDirs(wd, ACCEPTABLE_FILE_NAMES)
		if err != nil {
			return Config{}, err
		}

		filename, err = filepath.Abs(filename)
		if err != nil {
			return Config{}, err
		}

		configPath = filename
	}

	dat, err := ioutil.ReadFile(configPath)
	if err != nil {
		return Config{}, err
	}

	// Found config, now try to read it
	var config Config
	config.Path = configPath

	err = yaml.Unmarshal(dat, &config)
	if err != nil {
		parseError := &core.FailedToParseFile{ Name: configPath, Msg: err }
		return config, parseError
	}

	// Update the config

	if config.Theme.Table == "" {
		config.Theme.Table = "box"
	}

	if config.Theme.Tree == "" {
		config.Theme.Tree = "line"
	}

	// Set default shell command
	if config.Shell == "" {
		config.Shell = DEFAULT_SHELL
	}

	// Append absolute and relative path for each project
	for i := range config.Projects {
		config.Projects[i].Path, err = core.GetAbsolutePath(configPath, config.Projects[i].Path, config.Projects[i].Name)
		core.CheckIfError(err)

		config.Projects[i].RelPath, err = GetProjectRelPath(configPath, config.Projects[i].Path)
		core.CheckIfError(err)
	}

	// Append absolute and relative path for each dir
	for i := range config.Dirs {
		var abs, err= core.GetAbsolutePath(configPath, config.Dirs[i].Path, "")
		core.CheckIfError(err)

		config.Dirs[i].Name = path.Base(abs)
		config.Dirs[i].Path = abs

		config.Dirs[i].RelPath, err = GetProjectRelPath(configPath, config.Dirs[i].Path)
		core.CheckIfError(err)
	}

	// Import Tasks/Projects/Networks
	tasks := config.Tasks
	projects := config.Projects
	networks := config.SetNetworkList()
	for _, importPath := range config.Import {
	    ts, ps, ns, err := readExternalConfig(importPath)
	    core.CheckIfError(err)

	    tasks = append(tasks, ts...)
	    projects = append(projects, ps...)
	    networks = append(networks, ns...)
	}

	// Set default shell command for all tasks
	for i := range tasks {
		if tasks[i].Shell == "" {
			tasks[i].Shell = DEFAULT_SHELL
		}

		for j := range tasks[i].Commands {
			if tasks[i].Commands[j].Shell == "" {
				tasks[i].Commands[j].Shell = DEFAULT_SHELL
			}
		}
	}

	config.Projects = projects
	config.NetworkList = networks
	config.Tasks = tasks

	return config, nil
}

func readExternalConfig(importPath string) ([]Task, []Project, []Network, error) {
    dat, err := ioutil.ReadFile(importPath)
    core.CheckIfError(err)

    // Found config, now try to read it
    var config Config
    err = yaml.Unmarshal(dat, &config)
    if err != nil {
	parseError := &core.FailedToParseFile{ Name: importPath, Msg: err }
	core.CheckIfError(parseError)
    }

    // Append absolute and relative path for each project
    for i := range config.Projects {
	    config.Projects[i].Path, err = core.GetAbsolutePath(importPath, config.Projects[i].Path, config.Projects[i].Name)
	    core.CheckIfError(err)

	    config.Projects[i].RelPath, err = GetProjectRelPath(importPath, config.Projects[i].Path)
	    core.CheckIfError(err)
    }

    // Unpack Network to NetworkList
    networks := config.SetNetworkList()

    return config.Tasks, config.Projects, networks, nil
}

// PROJECTS

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
		if (project.Url != "") {
			urls = append(urls, project.Url)
		}
	}

	return urls
}

func (c Config) GetProjectsTree (dirs []string, tags []string) []core.TreeNode {
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
			if (pa.Name == pb.Name) {
				projects = append(projects, pa)
			}
		}
	}

	return projects
}

// TASKS

func (c Config) GetTasksByNames(names []string) []Task {
	if len(names) == 0 {
		return c.Tasks
	}

	var filteredTasks []Task
	var foundTasks []string
	for _, name := range names {
		if core.StringInSlice(name, foundTasks) {
			continue
		}

		for _, project := range c.Tasks {
			if name == project.Name {
				filteredTasks = append(filteredTasks, project)
				foundTasks = append(foundTasks, name)
			}
		}
	}

	return filteredTasks
}

func (c Config) GetTaskNames() []string {
	taskNames := []string{}
	for _, project := range c.Tasks {
		taskNames = append(taskNames, project.Name)
	}

	return taskNames
}

func (c Config) GetTask(task string) (*Task, error) {
	for _, cmd := range c.Tasks {
		if task == cmd.Name {
			return &cmd, nil
		}
	}

	return nil, &core.TaskNotFound{ Name: task }
}

func (c Config) GetTasks() []string {
	var s []string
	for _, cmd := range c.Tasks {
		s = append(s, cmd.Name)
	}

	return s
}

// DIRS

func (c Config) FilterDirs(
    cwdFlag bool,
    allDirsFlag bool,
    dirPathsFlag []string,
    dirsFlag []string,
    tagsFlag []string,
) []Dir {
	var finalDirs []Dir
	if allDirsFlag {
		finalDirs = c.Dirs
	} else {
		var dirPaths []Dir
		if len(dirPathsFlag) > 0 {
			dirPaths = c.GetDirsByPath(dirPathsFlag)
		}

		var tagDirs []Dir
		if len(tagsFlag) > 0 {
			tagDirs = c.GetDirsByTags(tagsFlag)
		}

		var dirs []Dir
		if len(dirsFlag) > 0 {
			dirs = c.GetDirs(dirsFlag)
		}

		var cwdDir Dir
		if cwdFlag {
			cwdDir = c.GetCwdDir()
		}

		finalDirs = GetUnionDirs(dirPaths, tagDirs, dirs, cwdDir)
	}

	return finalDirs
}

// Dirs must have all paths to match. For instance, if --tags frontend,backend
// is passed, then a dir must have both tags.
func (c Config) GetDirsByPath(drs []string) []Dir {
	if len(drs) == 0 {
		return c.Dirs
	}

	var dirs []Dir
	for _, dir := range c.Dirs {

		// Variable use to check that all dirs are matched
		var numMatched int = 0
		for _, d := range drs {
			if strings.Contains(dir.RelPath, d) {
				numMatched = numMatched + 1
			}
		}

		if numMatched == len(drs) {
			dirs = append(dirs, dir)
		}
	}

	return dirs
}

func (c Config) GetDirs(flagDir []string) []Dir {
	var dirs []Dir

	for _, v := range flagDir {
		for _, d := range c.Dirs {
			if v == d.Name {
				dirs = append(dirs, d)
			}
		}
	}

	return dirs
}

func (c Config) GetCwdDir() Dir {
	cwd, err := os.Getwd()
	core.CheckIfError(err)

	var dir Dir
	parts := strings.Split(cwd, string(os.PathSeparator))

	out:
	for i := len(parts) - 1; i >= 0; i-- {
		p := strings.Join(parts[0:i+1], string(os.PathSeparator))

		for _, pro := range c.Dirs {
			if p == pro.Path {
				dir = pro
				break out
			}
		}
	}

	return dir
}

func GetUnionDirs(a []Dir, b []Dir, c []Dir, d Dir) []Dir {
	drs := []Dir{}

	for _, dir := range a {
		if !DirInSlice(dir.Path, drs) {
			drs = append(drs, dir)
		}
	}

	for _, dir := range b {
		if !DirInSlice(dir.Path, drs) {
			drs = append(drs, dir)
		}
	}

	for _, dir := range c {
		if !DirInSlice(dir.Path, drs) {
			drs = append(drs, dir)
		}
	}

	if d.Name != "" {
		drs = append(drs, d)
	}

	dirs := []Dir{}
	dirs = append(dirs, drs...)

	return dirs
}

func DirInSlice(name string, list []Dir) bool {
	for _, d := range list {
		if d.Name == name {
			return true
		}
	}
	return false
}

func (c Config) GetDirNames() []string {
	names := []string{}
	for _, dir := range c.Dirs {
		names = append(names, dir.Name)
	}

	return names
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
func (c Config) GetDirPaths() []string {
	dirs := []string{}
	for _, dir := range c.Dirs {

		ps := strings.Split(filepath.Dir(dir.RelPath), string(os.PathSeparator))
		for i := 1; i <= len(ps); i++ {
			p := filepath.Join(ps[0:i]...)

			if p != "." && !core.StringInSlice(p, dirs) {
				dirs = append(dirs, p)
			}
		}
	}

	return dirs
}

func GetIntersectDirs(a []Dir, b []Dir) []Dir {
	dirs := []Dir{}

	for _, pa := range a {
		for _, pb := range b {
			if (pa.Name == pb.Name) {
				dirs = append(dirs, pa)
			}
		}
	}

	return dirs
}

func (c Config) GetDirsByName(names []string) []Dir {
	if len(names) == 0 {
		return c.Dirs
	}

	var filtered []Dir
	var found []string
	for _, name := range names {
		if core.StringInSlice(name, found) {
			continue
		}

		for _, dir := range c.Dirs {
			if name == dir.Name {
				filtered = append(filtered, dir)
				found = append(found, name)
			}
		}
	}

	return filtered
}

// Dirs must have all tags to match. For instance, if --tags frontend,backend
// is passed, then a dir must have both tags.
func (c Config) GetDirsByTags(tags []string) []Dir {
	if len(tags) == 0 {
		return c.Dirs
	}

	var dirs []Dir
	for _, dir := range c.Dirs {
		// Variable use to check that all tags are matched
		var numMatched int = 0
		for _, tag := range tags {
			for _, dirTag := range dir.Tags {
				if dirTag == tag {
					numMatched = numMatched + 1
				}
			}
		}

		if numMatched == len(tags) {
			dirs = append(dirs, dir)
		}
	}

	return dirs
}

func (c Config) GetDirsTree (drs []string, tags []string) []core.TreeNode {
	var tree []core.TreeNode
	var paths = []string{}

	dirPaths := c.GetDirsByPath(drs)
	dirTags := c.GetDirsByTags(tags)
	dirs := GetIntersectDirs(dirPaths, dirTags)

	for _, p := range dirs {
		if p.RelPath != "." {
			paths = append(paths, p.RelPath)
		}
	}

	for i := range paths {
		tree = core.AddToTree(tree, strings.Split(paths[i], string(os.PathSeparator)))
	}

	return tree
}

// TAGS

func (c Config) GetTagsByProject(projectNames []string) []string {
	tags := []string{}
	for _, project := range c.Projects {
		if core.StringInSlice(project.Name, projectNames) {
			tags = append(tags, project.Tags...)
		}
	}

	return tags
}

func (c Config) GetTags() []string {
	tags := []string{}
	for _, project := range c.Projects {
		for _, tag := range project.Tags {
			if !core.StringInSlice(tag, tags) {
				tags = append(tags, tag)
			}
		}
	}

	for _, dir := range c.Dirs {
		for _, tag := range dir.Tags {
			if !core.StringInSlice(tag, tags) {
				tags = append(tags, tag)
			}
		}
	}

	return tags
}

// Open mani config in editor
func (c Config) EditConfig() {
	editor := os.Getenv("EDITOR")
	cmd := exec.Command(editor, c.Path)
	cmd.Env = os.Environ()
    cmd.Stdin = os.Stdin
    cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
    err := cmd.Run()
	core.CheckIfError(err)
}

// Open mani config in editor and optionally go to line matching the task name
func (c Config) EditTask(taskName string) {
	dat, err := ioutil.ReadFile(c.Path)
	core.CheckIfError(err)

	type ConfigTmp struct {
		Tasks	   yaml.Node
	}

	var configTmp ConfigTmp
	err = yaml.Unmarshal([]byte(dat), &configTmp)
	core.CheckIfError(err)

	lineNr := 0
	if taskName == "" {
		lineNr = configTmp.Tasks.Line - 1
	} else {
		out:
		for _, task := range configTmp.Tasks.Content {
			for _, node := range task.Content {
				if node.Value == taskName {
					lineNr = node.Line
					break out
				}
			}
		}
	}

	editor := os.Getenv("EDITOR")
	var args []string
	switch editor {
	case "vim":
		args = []string{fmt.Sprintf("+%v", lineNr), c.Path}
	case "vi":
		args = []string{fmt.Sprintf("+%v", lineNr), c.Path}
	case "emacs":
		args = []string{fmt.Sprintf("+%v", lineNr), c.Path}
	case "nano":
		args = []string{fmt.Sprintf("+%v", lineNr), c.Path}
	case "code": // visual studio code
		args = []string{"--goto", fmt.Sprintf("%s:%v", c.Path, lineNr)}
	case "idea": // Intellij
		args = []string{"--line", fmt.Sprintf("%v", lineNr), c.Path}
	case "subl": // Sublime
		args = []string{fmt.Sprintf("%s:%v", c.Path, lineNr)}
	case "atom":
		args = []string{fmt.Sprintf("%s:%v", c.Path, lineNr)}
	case "notepad-plus-plus":
		args = []string{"-n", fmt.Sprintf("%v", lineNr), c.Path}
	default:
		args = []string{c.Path}
	}

	cmd := exec.Command(editor, args...)
	cmd.Env = os.Environ()
    cmd.Stdin = os.Stdin
    cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
    err = cmd.Run()
	core.CheckIfError(err)
}

// Open mani config in editor and optionally go to line matching the project name
func (c Config) EditProject(projectName string) {
	dat, err := ioutil.ReadFile(c.Path)
	core.CheckIfError(err)

	type ConfigTmp struct {
		Projects	   yaml.Node
	}

	var configTmp ConfigTmp
	err = yaml.Unmarshal([]byte(dat), &configTmp)
	core.CheckIfError(err)

	lineNr := 0
	if projectName == "" {
		lineNr = configTmp.Projects.Line - 1
	} else {
		out:
		for _, project := range configTmp.Projects.Content {
			for _, node := range project.Content {
				if node.Value == projectName {
					lineNr = node.Line
					break out
				}
			}
		}
	}

	editor := os.Getenv("EDITOR")
	var args []string
	switch editor {
	case "vim":
		args = []string{fmt.Sprintf("+%v", lineNr), c.Path}
	case "vi":
		args = []string{fmt.Sprintf("+%v", lineNr), c.Path}
	case "emacs":
		args = []string{fmt.Sprintf("+%v", lineNr), c.Path}
	case "nano":
		args = []string{fmt.Sprintf("+%v", lineNr), c.Path}
	case "code": // visual studio code
		args = []string{"--goto", fmt.Sprintf("%s:%v", c.Path, lineNr)}
	case "idea": // Intellij
		args = []string{"--line", fmt.Sprintf("%v", lineNr), c.Path}
	case "subl": // Sublime
		args = []string{fmt.Sprintf("%s:%v", c.Path, lineNr)}
	case "atom":
		args = []string{fmt.Sprintf("%s:%v", c.Path, lineNr)}
	case "notepad-plus-plus":
		args = []string{"-n", fmt.Sprintf("%v", lineNr), c.Path}
	default:
		args = []string{c.Path}
	}

	cmd := exec.Command(editor, args...)
	cmd.Env = os.Environ()
    cmd.Stdin = os.Stdin
    cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
    err = cmd.Run()
	core.CheckIfError(err)
}

// Open mani config in editor and optionally go to line matching the dir name
func (c Config) EditDir(name string) {
	dat, err := ioutil.ReadFile(c.Path)
	core.CheckIfError(err)

	type ConfigTmp struct {
		Dirs	yaml.Node
	}

	var configTmp ConfigTmp
	err = yaml.Unmarshal([]byte(dat), &configTmp)
	core.CheckIfError(err)

	lineNr := 0
	if name == "" {
		lineNr = configTmp.Dirs.Line - 1
	} else {
		out:
		for _, dir := range configTmp.Dirs.Content {
			for _, node := range dir.Content {
				if node.Value == name {
					lineNr = node.Line
					break out
				}
			}
		}
	}

	editor := os.Getenv("EDITOR")
	var args []string
	switch editor {
	case "vim":
		args = []string{fmt.Sprintf("+%v", lineNr), c.Path}
	case "vi":
		args = []string{fmt.Sprintf("+%v", lineNr), c.Path}
	case "emacs":
		args = []string{fmt.Sprintf("+%v", lineNr), c.Path}
	case "nano":
		args = []string{fmt.Sprintf("+%v", lineNr), c.Path}
	case "code": // visual studio code
		args = []string{"--goto", fmt.Sprintf("%s:%v", c.Path, lineNr)}
	case "idea": // Intellij
		args = []string{"--line", fmt.Sprintf("%v", lineNr), c.Path}
	case "subl": // Sublime
		args = []string{fmt.Sprintf("%s:%v", c.Path, lineNr)}
	case "atom":
		args = []string{fmt.Sprintf("%s:%v", c.Path, lineNr)}
	case "notepad-plus-plus":
		args = []string{"-n", fmt.Sprintf("%v", lineNr), c.Path}
	default:
		args = []string{c.Path}
	}

	cmd := exec.Command(editor, args...)
	cmd.Env = os.Environ()
    cmd.Stdin = os.Stdin
    cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
    err = cmd.Run()
	core.CheckIfError(err)
}

func UpdateProjectsToGitignore(projectNames []string, gitignoreFilename string) error {
	l := list.New()
	gitignoreFile, err := os.OpenFile(gitignoreFilename, os.O_RDWR, 0644)

	if err != nil {
		return &core.FailedToOpenFile{ Name: gitignoreFilename }
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
	if (len(urls) == 0) {
		fmt.Println("No projects to sync")
		return
	}

	var cfg yacspin.Config
	cfg = yacspin.Config {
		Frequency:       100 * time.Millisecond,
		CharSet:         yacspin.CharSets[9],
		SuffixAutoColon: false,
		Message: " Cloning",
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

func GetClosestConfigFile() (string, error) {
	wd, _ := os.Getwd()
	filename, err := core.FindFileInParentDirs(wd, ACCEPTABLE_FILE_NAMES)
	return filename, err
}
