package core

import (
	"bufio"
	"fmt"
	color "github.com/logrusorgru/aurora"
	"gopkg.in/yaml.v3"
	"io/ioutil"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"strings"
)

var ACCEPTABLE_FILE_NAMES = []string{"mani.yaml", "mani.yml", ".mani", ".mani.yaml", ".mani.yml"}

func GetProjectsByTag(tags []string, projects []Project) []Project {
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

func GetProjects(flagProjects []string, projects []Project) []Project {
	var matchedProjects []Project

	for _, v := range flagProjects {
		for _, p := range projects {
			if v == p.Name {
				matchedProjects = append(matchedProjects, p)
			}
		}
	}

	return matchedProjects
}

func GetUnionProjects(a []Project, b []Project) []Project {
	m := make(map[string]Project)

	for _, project := range a {
		m[project.Name] = project
	}

	for _, project := range b {
		m[project.Name] = project
	}

	projects := []Project{}
	for _, p := range m {
		projects = append(projects, p)
	}

	return projects
}

func GetCommand(command string, commands []Command) (*Command, error) {
	for _, cmd := range commands {
		if command == cmd.Name {
			return &cmd, nil
		}
	}

	return nil, &CommandNotFound{command}
}

func GetCommands(commands []Command) []string {
	var list []string
	for _, cmd := range commands {
		list = append(list, cmd.Name)
	}

	return list
}

func findFileInParentDirs(path string, files []string) (string, error) {
	for _, file := range files {
		pathToFile := filepath.Join(path, file)

		if _, err := os.Stat(pathToFile); err == nil {
			return pathToFile, nil
		}
	}

	parentDir := filepath.Dir(path)

	if parentDir == "/" {
		return "", &ConfigNotFound{files}
	}

	return findFileInParentDirs(parentDir, files)
}

func GetClosestConfigFile() (string, error) {
	wd, _ := os.Getwd()
	filename, err := findFileInParentDirs(wd, ACCEPTABLE_FILE_NAMES)

	return filename, err
}

func ReadConfig(cfgName string) (string, Config, error) {
	var configFilename string

	if cfgName != "" {
		filename, err := filepath.Abs(cfgName)
		if err != nil {
			return "", Config{}, err
		}
		configFilename = filename
	} else {
		wd, err := os.Getwd()
		if err != nil {
			return "", Config{}, err
		}

		filename, err := findFileInParentDirs(wd, ACCEPTABLE_FILE_NAMES)
		if err != nil {
			return "", Config{}, err
		}

		filename, err = filepath.Abs(filename)
		if err != nil {
			return "", Config{}, err
		}

		configFilename = filename
	}

	dat, err := ioutil.ReadFile(configFilename)

	if err != nil {
		return "", Config{}, err
	}

	var config Config
	err = yaml.Unmarshal(dat, &config)
	if err != nil {
		parseError := &FailedToParseFile{configFilename, err}
		return "", config, parseError
	}

	return configFilename, config, nil
}

// Get the absolute path to a project
// Need to support following path types:
//		lala/land
//		./lala/land
//		../lala/land
//		/lala/land
//		$HOME/lala/land
//		~/lala/land
//		~root/lala/land
func GetAbsolutePath(configPath string, projectPath string, projectName string) (string, error) {
    projectPath = os.ExpandEnv(projectPath)

    usr, err := user.Current()
    if err != nil {
      return "", err
    }

    homeDir := usr.HomeDir
	configDir := filepath.Dir(configPath)

	// TODO: Remove any .., make path absolute and then cut of configDir
	var path string
    if projectPath == "~" {
        path = homeDir
    } else if strings.HasPrefix(projectPath, "~/") {
        path = filepath.Join(homeDir, projectPath[2:])
    } else if len(projectPath) > 0 && filepath.IsAbs(projectPath) {
		path = projectPath
	} else if len(projectPath) > 0 {
		path = filepath.Join(configDir, projectPath)
	} else {
		path = filepath.Join(configDir, projectName)
	}

    return path, nil
}

func ExecCmd(configPath string, project Project, cmdString string, dryRun bool) error {
	fmt.Println()
	fmt.Println(color.Bold(color.Blue(project.Name)))

    projectPath, err := GetAbsolutePath(configPath, project.Path, project.Name)
    if err != nil {
      return &FailedToParsePath{projectPath}
    }
	if _, err := os.Stat(projectPath); os.IsNotExist(err) {
		return &PathDoesNotExist{projectPath}
	}

	cmd := exec.Command("sh", "-c", cmdString)
	cmd.Dir = projectPath
	cmd.Env = os.Environ()
	if dryRun {
		fmt.Println(os.ExpandEnv(cmdString))
	} else {
		out, _ := cmd.CombinedOutput()
		fmt.Println(string(out))
	}

	return nil
}

func RunCommand(configPath string, project Project, command *Command, userArguments []string, dryRun bool) error {
	fmt.Println()
	fmt.Println(color.Bold(color.Blue(project.Name)))

    projectPath, err := GetAbsolutePath(configPath, project.Path, project.Name)
    if err != nil {
      return &FailedToParsePath{projectPath}
    }
	if _, err := os.Stat(projectPath); os.IsNotExist(err) {
		return &PathDoesNotExist{projectPath}
	}

	// Set Arguments
	// Format to key=value string
	projectNameEnv := fmt.Sprintf("project_name=%s", project.Name)
	projectUrlEnv := fmt.Sprintf("project_url=%s", project.Url)
	projectPathEnv := fmt.Sprintf("project_path=%s", project.Path)

	userArguments = append(userArguments, projectNameEnv, projectUrlEnv, projectPathEnv)

	var userArgumentKeys []string
	for _, arg := range userArguments {
		kv := strings.SplitN(arg, "=", 2)
		userArgumentKeys = append(userArgumentKeys, kv[0])
	}

	for k, v := range command.Args {
		if !StringInSlice(k, userArgumentKeys) {
			fmt.Println(k, v)
			defaultArg := fmt.Sprintf("%s=%s", k, v)
			userArguments = append(userArguments, defaultArg)
		}
	}

	// Execute Command
	cmd := exec.Command("sh", "-c", command.Command)
	cmd.Dir = projectPath
	if dryRun {
		for _, arg := range userArguments {
			env := strings.SplitN(arg, "=", 2)
			os.Setenv(env[0], env[1])
		}

		fmt.Println(os.ExpandEnv(command.Command))
	} else {
		cmd.Env = append(os.Environ(), userArguments...)
		out, _ := cmd.CombinedOutput()
		fmt.Println(string(out))
	}

	return nil
}

func CloneRepos(configPath string, projects []Project) {
	for _, project := range projects {
		if project.Url != "" {
			err := cloneRepo(configPath, project)

			if err != nil {
				fmt.Println(err)
			}
		}
	}
}

func cloneRepo(configPath string, project Project) error {
    projectPath, err := GetAbsolutePath(configPath, project.Path, project.Name)
    if err != nil {
      return &FailedToParsePath{projectPath}
    }

	if _, err := os.Stat(projectPath); os.IsNotExist(err) {
		cmd := exec.Command("git", "clone", project.Url, projectPath)
		_, err := cmd.CombinedOutput()
		if err != nil {
			return err
		}
	}

	fmt.Println(color.Green("\u2713"), "synced", color.Bold(project.Name))

	return nil
}

func AddStringToFile(name string, filename string) {
	fmt.Println(name, filename)
}

func AddProjectsToGitignore(projects []Project, gitignoreFilename string) error {
	gitignoreFile, err := os.OpenFile(gitignoreFilename, os.O_RDWR, 0644)
	if err != nil {
		return &FailedToOpenFile{gitignoreFilename}
	}

	for _, project := range projects {
      if project.Path != "." {
        gitignoreFile.WriteString(project.Path)
        gitignoreFile.WriteString("\n")
      }
	}
	gitignoreFile.Close()

	return nil
}

func IsSubDirectory(rootPath string, subPath string) bool {
	return false
}

func UpdateProjectsToGitignore(projects map[string]bool, gitignoreFilename string) error {
	gitignoreFile, err := os.OpenFile(gitignoreFilename, os.O_RDWR, 0644)

	if err != nil {
		return &FailedToOpenFile{gitignoreFilename}
	}

	scanner := bufio.NewScanner(gitignoreFile)
	for scanner.Scan() {
		line := scanner.Text()
		if _, ok := projects[line]; ok {
			projects[line] = true
		}
	}

	for project, found := range projects {
		if !found {
			gitignoreFile.WriteString(project)
			gitignoreFile.WriteString("\n")
			fmt.Println(color.Green("\u2713"), "added project", color.Bold(project), "to .gitignore")
		}
	}
	gitignoreFile.Close()

	return nil
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

        // Return nil
        gitDir := filepath.Join(path, ".git")
        if _, err := os.Stat(gitDir); !os.IsNotExist(err) {
          name := filepath.Base(path)
          relPath, _ := filepath.Rel(rootPath, path)
          url := GetRemoteUrl(path)
          project := Project{Name: name, Path: relPath, Url: url}
          projects = append(projects, project)

          return filepath.SkipDir
        }

		return nil
	})

  return projects, err
}

func GetRemoteUrl(path string) string {
	cmd := exec.Command("git", "config", "--get", "remote.origin.url")
	cmd.Dir = path
	output, err := cmd.CombinedOutput()
	var url string
	if err != nil {
		url = ""
	} else {
		url = strings.TrimSuffix(string(output), "\n")
	}

    return url
}
