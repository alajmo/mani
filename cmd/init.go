package cmd

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"text/template"

	"github.com/spf13/cobra"
	color "github.com/logrusorgru/aurora"

	"github.com/alajmo/mani/core"
	"github.com/alajmo/mani/core/dao"
)

func initCmd() *cobra.Command {
	var autoDiscovery bool

	cmd := cobra.Command{
		Use:   "init",
		Short: "Initialize a mani repository",
		Long: `Initialize a mani repository.

Creates a mani repository - a directory with configuration file mani.yaml and a .gitignore file.`,
		Example: `  # Basic example
  mani init

  # Skip auto-discovery of projects
  mani init --auto-discovery=false`,

		Args: cobra.MaximumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			runInit(args, autoDiscovery)
		},
	}

	cmd.Flags().BoolVar(&autoDiscovery, "auto-discovery", true, "walk current directory and find git repositories to add to mani.yaml")

	return &cmd
}

func runInit(args []string, autoDiscovery bool) {
	// Choose to initialize mani in a different directory
	// 1. absolute or
	// 2. relative or
	// 3. working directory
	var configDir string
	if len(args) > 0 && filepath.IsAbs(args[0]) {
		configDir = args[0]
	} else if len(args) > 0 {
		wd, err := os.Getwd()
		core.CheckIfError(err)
		configDir = filepath.Join(wd, args[0])
	} else {
		wd, err := os.Getwd()
		core.CheckIfError(err)
		configDir = wd
	}

	err := os.MkdirAll(configDir, os.ModePerm)
	core.CheckIfError(err)

	configPath := filepath.Join(configDir, "mani.yaml")
	if _, err := os.Stat(configPath); err == nil {
		fmt.Printf("fatal: %q is already a mani directory\n", configDir)
		os.Exit(1)
	}

	url := core.GetWdRemoteUrl(configDir)
	rootName := filepath.Base(configDir)
	rootPath := "."
	rootUrl := url
	rootProject := dao.Project {Name: rootName, Path: rootPath, Url: rootUrl}
	projects := []dao.Project{rootProject}
	if autoDiscovery {
		prs, err := dao.FindVCSystems(configDir)

		if err != nil {
			fmt.Println(err)
		}

		projects = append(projects, prs...)
	}

	funcMap := template.FuncMap{
		"projectItem": func(name string, path string, url string) string {
			var txt = "- name: " + name

			if name != path {
				txt = txt + "\n    path: " + path
			}

			if url != "" {
				txt = txt + "\n    url: " + url
			}

			return txt
		},
	}

	// - name: {{ .Name }}
	// {{ if ne .Name .Path }}path: {{ .Path }}{{ end }}
	// {{ if .Url }}url: {{ .Url }} {{ end }}

	// Path, Name, Url
	tmpl, err := template.New("init").Funcs(funcMap).Parse(`projects:
{{- range .}}
  {{ (projectItem .Name .Path .Url) }}
{{ end }}
tasks:
  - name: hello-world
    description: Print Hello World
    command: echo "Hello World"
`,
	)

	core.CheckIfError(err)

	// Create mani.yaml
	f, err := os.Create(configPath)
	core.CheckIfError(err)

	err = tmpl.Execute(f, projects)
	core.CheckIfError(err)

	f.Close()
	fmt.Println(color.Green("\u2713"), "Initialized mani repository in", configDir)

	// Add gitignore file
	gitignoreFilepath := filepath.Join(configDir, ".gitignore")
	if _, err := os.Stat(gitignoreFilepath); os.IsNotExist(err) {
		err := ioutil.WriteFile(gitignoreFilepath, []byte(""), 0644)

		core.CheckIfError(err)
	}

	var projectNames []string
	for _, project := range projects {
		if project.Url == "" {
			continue
		}

		if project.Path == "." {
			continue
		}

		projectNames = append(projectNames, project.Path)
	}

	// Add projects to gitignore file
	err = dao.UpdateProjectsToGitignore(projectNames, gitignoreFilepath)
	core.CheckIfError(err)
}
