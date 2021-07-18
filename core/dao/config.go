package dao

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"io/ioutil"
	"bufio"
	"container/list"
	"strings"

	"gopkg.in/yaml.v3"

	"github.com/alajmo/mani/core"
)

var (
	Version = "dev"
	DEFAULT_SHELL = "sh -c"
	ACCEPTABLE_FILE_NAMES = []string{"mani.yaml", "mani.yml", ".mani", ".mani.yaml", ".mani.yml"}
)

type Config struct {
	Path string

	Env		 map[string]string  `yaml:"env"`
	Shell    string				`yaml:"shell"`
	Projects []Project			`yaml:"projects"`
	Commands []Command			`yaml:"commands"`
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

	// Set default shell command
	if config.Shell == "" {
		config.Shell = DEFAULT_SHELL
	}

	// Set default shell command for all commands
	for i := range config.Commands {
		if config.Commands[i].Shell == "" {
			config.Commands[i].Shell = DEFAULT_SHELL
		}
	}

	// Append absolute and relative path for each project
	for i := range config.Projects {
		config.Projects[i].Path, err = GetAbsolutePath(configPath, config.Projects[i].Path, config.Projects[i].Name)
		core.CheckIfError(err)

		config.Projects[i].RelPath, err = GetProjectRelPath(configPath, config.Projects[i].Path)
		core.CheckIfError(err)
	}

	return config, nil
}

func (c Config) EvaluateEnv() ([]string, error) {
	var envs []string

	for k, v := range c.Env {
		if strings.HasPrefix(v, "$(") && strings.HasSuffix(v, ")") {
			v = strings.TrimPrefix(v, "$(")
			v = strings.TrimSuffix(v, ")")

			out, err := exec.Command("sh", "-c", v).Output()
			if err != nil {
				return envs, &core.ConfigEnvFailed { Name: k, Err: err }
			}

			envs = append(envs, fmt.Sprintf("%v=%v", k, string(out)))
		} else {
			envs = append(envs, fmt.Sprintf("%v=%v", k, v))
		}
	}

	return envs, nil
}

// PROJECTS

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

func (c Config) FilterProjects(
	cwdFlag bool,
	allProjectsFlag bool,
	tagsFlag []string,
	projectsFlag []string,
) []Project {
	var finalProjects []Project
	if allProjectsFlag {
		finalProjects = c.Projects
	} else {
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

		finalProjects = GetUnionProjects(tagProjects, projects, cwdProject)
	}

	return finalProjects
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

// Projects must have all tags to match. For instance, if --tags frontend,backend
// is passed, then a project must have both tags.
func (c Config) GetProjectsByTags(tags []string) []Project {
	if len(tags) == 0 {
		return c.Projects
	}

	var filteredProjects []Project
	for _, project := range c.Projects {
		var foundTags int = 0
		for _, tag := range tags {
			for _, projectTag := range project.Tags {
				if projectTag == tag {
					foundTags = foundTags + 1
				}
			}
		}

		if foundTags == len(tags) {
			filteredProjects = append(filteredProjects, project)
		}
	}

	return filteredProjects
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

func GetUnionProjects(a []Project, b []Project, c Project) []Project {
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

	if c.Name != "" {
		prjs = append(prjs, c)
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

// COMMANDS

func (c Config) GetCommandsByNames(names []string) []Command {
	if len(names) == 0 {
		return c.Commands
	}

	var filteredCommands []Command
	var foundCommands []string
	for _, name := range names {
		if core.StringInSlice(name, foundCommands) {
			continue
		}

		for _, project := range c.Commands {
			if name == project.Name {
				filteredCommands = append(filteredCommands, project)
				foundCommands = append(foundCommands, name)
			}
		}
	}

	return filteredCommands
}

func (c Config) GetCommandNames() []string {
	commandNames := []string{}
	for _, project := range c.Commands {
		commandNames = append(commandNames, project.Name)
	}

	return commandNames
}

func (c Config) GetCommand(command string) (*Command, error) {
	for _, cmd := range c.Commands {
		if command == cmd.Name {
			return &cmd, nil
		}
	}

	return nil, &core.CommandNotFound{ Name: command }
}

func (c Config) GetCommands() []string {
	var s []string
	for _, cmd := range c.Commands {
		s = append(s, cmd.Name)
	}

	return s
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

	return tags
}

func (c Config) EditFile() {
	cmd := exec.Command("sh", "-c", fmt.Sprintf("$EDITOR %s", c.Path))
	cmd.Env = os.Environ()
    cmd.Stdin = os.Stdin
    cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
    err := cmd.Run()
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

func (c Config) CloneRepos() {
	urls := c.GetProjectUrls()
	if (len(urls) == 0) {
		fmt.Println("No projects to sync")
		return
	}

	allProjectsSynced := true
	for _, project := range c.Projects {
		if project.Url != "" {
			err := CloneRepo(c.Path, project)

			if err != nil {
				allProjectsSynced = false
				fmt.Println(err)
			}
		}
	}

	if allProjectsSynced {
		fmt.Println("All projects synced")
	}
}

func GetClosestConfigFile() (string, error) {
	wd, _ := os.Getwd()
	filename, err := core.FindFileInParentDirs(wd, ACCEPTABLE_FILE_NAMES)
	return filename, err
}
