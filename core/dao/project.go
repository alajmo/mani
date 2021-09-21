package dao

import (
	"strings"
	"fmt"
	"sync"
	"bytes"
	"os"
	"os/exec"
	"path/filepath"

	color "github.com/logrusorgru/aurora"

	"github.com/alajmo/mani/core"
)

type Project struct {
	Name        string   `yaml:"name"`
	Path        string   `yaml:"path"`
	Description string   `yaml:"description"`
	Url         string   `yaml:"url"`
	Clone       string   `yaml:"clone"`
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
		syncErrors[project.Name] = (&core.FailedToParsePath { Name: projectPath }).Error()
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
