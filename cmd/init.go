package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"text/template"
)

func initCmd() *cobra.Command {
	cmd := cobra.Command{
		Use:   "init",
		Short: "Init mani repo",
		Long:  "Init mani repo",
		Run:   runInit,
	}

	return &cmd
}

func runInit(cmd *cobra.Command, args []string) {
	configFilename, _ := filepath.Abs("mani.yaml")
	if _, err := os.Stat(configFilename); os.IsNotExist(err) {
		gitignoreFilename, _ := filepath.Abs(".gitignore")
		if _, err := os.Stat(gitignoreFilename); os.IsNotExist(err) {
			err := ioutil.WriteFile(gitignoreFilename, []byte(""), 0644)
			check(err)
		}

		wd, _ := os.Getwd()
		cmd := exec.Command("git", "config", "--get", "remote.origin.url")
		url, err := cmd.CombinedOutput()

		values := map[string]string{
			"Name": filepath.Base(wd),
			"Path": ".",
			"Url":  strings.TrimSuffix(string(url), "\n"),
		}
		tmpl, err := template.New("default").Parse(`version: alpha

projects:
  - name: {{ .Name }}
    path: {{ .Path }}
    url: {{ .Url }}

commands:
  - name: hello-world
    description: Print Hello World
    command: echo "Hello World"
`,
		)

		f, err := os.Create("mani.yaml")
		check(err)
		err = tmpl.Execute(f, values)
		check(err)
		f.Close()
		fmt.Println("Initialized mani repository in", configFilename)

	} else {
		fmt.Println("fatal: already mani repository")
	}
}
