package dao

import (
	"strings"
	"fmt"
	"time"

	"os"
	"os/exec"
	"os/user"
	"path/filepath"

	color "github.com/logrusorgru/aurora"
	"github.com/theckman/yacspin"

	"github.com/alajmo/mani/core"
)

type Project struct {
	Name        string   `yaml:"name"`
	Path        string   `yaml:"path"`
	Description string   `yaml:"description"`
	Url         string   `yaml:"url"`
	Tags        []string `yaml:"tags"`

	RelPath     string
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

func CloneRepo(configPath string, project Project) error {
	cfg := yacspin.Config{
		Frequency:       100 * time.Millisecond,
		CharSet:         yacspin.CharSets[9],
		SuffixAutoColon: false,

		Message:         fmt.Sprintf(" syncing %v", color.Bold(project.Name)),

		StopMessage:	 fmt.Sprintf(" synced %v", color.Bold(project.Name)),
		StopCharacter:   "✓",
		StopColors:      []string{"fgGreen"},

		StopFailCharacter:   "✗",
		StopFailColors:      []string{"fgRed"},
	}

	spinner, err := yacspin.New(cfg)
	if err != nil {
		return err
	}

	projectPath, err := GetAbsolutePath(configPath, project.Path, project.Name)
	if err != nil {
		return &core.FailedToParsePath{projectPath}
	}

	if _, err := os.Stat(projectPath); os.IsNotExist(err) {
		cmd := exec.Command("git", "clone", project.Url, projectPath)
		cmd.Env = os.Environ()

		// s.Suffix = fmt.Sprintf(" syncing %v", color.Bold(project.Name))
		err = spinner.Start()
		if err != nil {
			return err
		}

		stdoutStderr, err := cmd.CombinedOutput()

		if err != nil {
			spinner.StopFailMessage(fmt.Sprintf(" failed to sync %v \n%s", color.Bold(project.Name), stdoutStderr))

			serr := spinner.StopFail()
			if serr != nil {
				return serr
			}

			return err
		}

		err = spinner.Stop()
		if err != nil {
			return err
		}
	}

	// fmt.Println(color.Green("\u2713"), "synced", color.Bold(project.Name))

	return nil
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

