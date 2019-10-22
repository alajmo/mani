package core

import (
	"errors"
	"fmt"
	color "github.com/logrusorgru/aurora"
	"gopkg.in/yaml.v3"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func GetProjectsByTag(tags []string, projects []Project) ([]Project, error) {
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

	if len(matchedProjects) < 1 {
		return []Project{}, errors.New("tag: One or more of the provided tags does not have project associated with it")
	}

	return matchedProjects, nil
}

func GetProjects(flagProjects []string, projects []Project) ([]Project, error) {
	var matchedProjects []Project

	for _, v := range flagProjects {
		for _, p := range projects {
			if v == p.Name {
				matchedProjects = append(matchedProjects, p)
			}
		}
	}

	if len(matchedProjects) < 1 {
		return []Project{}, errors.New("project: One or more of the provided projects do not exist")
	}

	return matchedProjects, nil
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

	return nil, fmt.Errorf("command: Could not find command %q", command)
}

func ReadConfig() Config {
	filename, _ := filepath.Abs("./mani.yaml")
	dat, err := ioutil.ReadFile(filename)

	check(err)

	var config Config
	err = yaml.Unmarshal(dat, &config)

	check(err)

	return config
}

func RunCommand(project Project, command *Command, userArguments []string, dryRun bool) error {
	fmt.Println("")
	fmt.Println(color.Bold(color.Blue(project.Name)))

	cmd := exec.Command("sh", "-c", command.Command)
	var projectPath string
	if len(project.Path) > 0 {
		projectPath = project.Path
	} else {
		wd, _ := os.Getwd()
		projectPath = filepath.Join(wd, project.Name)
	}

	if _, err := os.Stat(projectPath); os.IsNotExist(err) {
		return fmt.Errorf("Path %q does not exist", projectPath)
	} else {
		cmd.Dir = projectPath
	}

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

func CloneRepo(project string, url string, path string) {
	if path != "" {
		if _, err := os.Stat(path); os.IsNotExist(err) {
			cmd := exec.Command("git", "clone", url, path)
			out, err := cmd.CombinedOutput()
			if err != nil {
				fmt.Println(string(out))
			} else {
				fmt.Println(color.Green("\u2713"), "synced", project)
			}
		} else {
			// fmt.Println(color.Green("\u2713"), "synced", project)
		}
	} else {
		if _, err := os.Stat(project); os.IsNotExist(err) {
			cmd := exec.Command("git", "clone", url)
			out, err := cmd.CombinedOutput()
			if err != nil {
				fmt.Println(string(out))
			} else {
				fmt.Println(color.Green("\u2713"), "synced", project)
			}

		} else {
			// fmt.Println(project, "", color.Green("\u2713"))
		}
	}
}

func AddStringToFile(name string, filename string) {
	fmt.Println(name, filename)
}
