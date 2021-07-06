package dao

import (
	"github.com/alajmo/mani/core"
	"fmt"
	"os"
	"os/exec"
	// "os/user"
	"path/filepath"
	// "github.com/theckman/yacspin"
	// "time"
	"gopkg.in/yaml.v3"
	"io/ioutil"
	// color "github.com/logrusorgru/aurora"
	// "bufio"
	// "container/list"
	"strings"
)

var (
	Version = "dev"
	DEFAULT_SHELL = "sh -c"
	ACCEPTABLE_FILE_NAMES = []string{"mani.yaml", "mani.yml", ".mani", ".mani.yaml", ".mani.yml"}
)

type Config struct {
	Shell    string    `yaml:"shell"`
	Projects []Project `yaml:"projects"`
	Commands []Command `yaml:"commands"`
}

func ReadConfig(cfgName string) (string, Config, error) {
	var configPath string

	if cfgName != "" {
		filename, err := filepath.Abs(cfgName)
		if err != nil {
			return "", Config{}, err
		}
		configPath = filename
	} else {
		wd, err := os.Getwd()
		if err != nil {
			return "", Config{}, err
		}

		filename, err := core.FindFileInParentDirs(wd, ACCEPTABLE_FILE_NAMES)
		if err != nil {
			return "", Config{}, err
		}

		filename, err = filepath.Abs(filename)
		if err != nil {
			return "", Config{}, err
		}

		configPath = filename
	}

	dat, err := ioutil.ReadFile(configPath)

	if err != nil {
		return "", Config{}, err
	}

	var config Config
	err = yaml.Unmarshal(dat, &config)
	if err != nil {
		parseError := &core.FailedToParseFile{configPath, err}
		return "", config, parseError
	}

	// Default shell command
	if config.Shell == "" {
		config.Shell = DEFAULT_SHELL
	}

	for i := range config.Commands {
		if config.Commands[i].Shell == "" {
			config.Commands[i].Shell = DEFAULT_SHELL
		}
	}

	for i := range config.Projects {
		config.Projects[i].Path, err = GetAbsolutePath(configPath, config.Projects[i].Path, config.Projects[i].Name)
		core.CheckIfError(err)

		config.Projects[i].RelPath, err = GetProjectRelPath(configPath, config.Projects[i].Path)
		core.CheckIfError(err)
	}

	return configPath, config, nil
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

func (c Config) FilterProjectOnName(names []string) []Project {
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

func (c Config) GetProjectsByTag(tags []string, projects []Project) []Project {
	var matchedProjects []Project

	for _, v := range tags {
		for _, p := range projects {
			for _, t := range p.Tags {
				if t == v {
					matchedProjects = append(matchedProjects, p)
				}
			}
		}
	}

	return matchedProjects
}

func (c Config) FilterProjects(
	config Config,
	cwdFlag bool,
	allProjectsFlag bool,
	tagsFlag []string,
	projectsFlag []string,
) []Project {
	var finalProjects []Project
	if allProjectsFlag {
		finalProjects = config.Projects
	} else {
		var tagProjects []Project
		if len(tagsFlag) > 0 {
			tagProjects = GetProjectsByTag(tagsFlag, config.Projects)
		}

		var projects []Project
		if len(projectsFlag) > 0 {
			projects = GetProjects(projectsFlag, config.Projects)
		}

		var cwdProject Project
		if cwdFlag {
			cwdProject = GetCwdProject(config.Projects)
		}

		finalProjects = GetUnionProjects(tagProjects, projects, cwdProject)
	}

	return finalProjects
}

func GetUnionProjects(a []Project, b []Project, c Project) []Project {
	m := []Project{}

	for _, project := range a {
		if !ProjectInSlice(project.Name, m) {
			m = append(m, project)
		}
	}

	for _, project := range b {
		if !ProjectInSlice(project.Name, m) {
			m = append(m, project)
		}
	}

	if c.Name != "" {
		m = append(m, c)
	}

	projects := []Project{}
	projects = append(projects, m...)

	return projects
}

// Projects must have all tags to match.
func (c Config) FilterProjectOnTag(tags []string) []Project {
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

// COMMANDS

func (c Config) FilterCommandOnName(names []string) []Command {
	if len(names) == 0 {
		return c.Commands
	}

	var filteredCommands []Command
	var foundCommands []string
	for _, name := range names {
		if StringInSlice(name, foundCommands) {
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

	return nil, &core.CommandNotFound{command}
}

func GetCommands(commands []Command) []string {
	var s []string
	for _, cmd := range commands {
		s = append(s, cmd.Name)
	}

	return s
}

// TAGS

func (c Config) FilterTagOnProject(projectNames []string) []string {
	tags := []string{}
	for _, project := range c.Projects {
		if StringInSlice(project.Name, projectNames) {
			tags = append(tags, project.Tags...)
		}
	}

	return tags
}

func (c Config) GetTags() []string {
	tags := []string{}
	for _, project := range c.Projects {
		for _, tag := range project.Tags {
			if !StringInSlice(tag, tags) {
				tags = append(tags, tag)
			}
		}
	}

	return tags
}

func EditFile(configPath string) {
	cmd := exec.Command("sh", "-c", fmt.Sprintf("$EDITOR %s", configPath))
	cmd.Env = os.Environ()
    cmd.Stdin = os.Stdin
    cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
    err := cmd.Run()
	CheckIfError(err)
}

func UpdateProjectsToGitignore(projectNames []string, gitignoreFilename string) error {
	l := list.New()
	gitignoreFile, err := os.OpenFile(gitignoreFilename, os.O_RDWR, 0644)

	if err != nil {
		return &FailedToOpenFile{gitignoreFilename}
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
	CheckIfError(err)

	_, err = gitignoreFile.Seek(0, 0)
	CheckIfError(err)

	for e := l.Front(); e != nil; e = e.Next() {
		str := fmt.Sprint(e.Value)
		_, err = gitignoreFile.WriteString(str)
		CheckIfError(err)

		_, err = gitignoreFile.WriteString("\n")
		CheckIfError(err)
	}

	gitignoreFile.Close()

	return nil
}

func ProjectInSlice(name string, list []dao.Project) bool {
	for _, p := range list {
		if p.Name == name {
			return true
		}
	}
	return false
}

func (c Config) CloneRepos(configPath string, projects []Project) {
	for _, project := range projects {
		if project.Url != "" {
			err := CloneRepo(configPath, project)

			if err != nil {
				fmt.Println(err)
			}
		}
	}
}

func GetClosestConfigFile() (string, error) {
	wd, _ := os.Getwd()
	filename, err := findFileInParentDirs(wd, ACCEPTABLE_FILE_NAMES)
	return filename, err
}
