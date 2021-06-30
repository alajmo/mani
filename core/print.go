package core

import (
	"fmt"
	tabby "github.com/cheynewallace/tabby"
	color "github.com/logrusorgru/aurora"
	"path/filepath"
	"strings"
)

func PrintProjects(configPath string, projects []Project, format string, listRaw bool) {
	switch format {
	case "table":
	case "list":
		if listRaw {
			for _, project := range projects {
				fmt.Println(project.Name)
			}
		} else {
			t := tabby.New()
			t.AddHeader("Project", "Tags", "Description")
			for _, project := range projects {
				t.AddLine(project.Name, strings.Join(project.Tags, ", "), project.Description)
			}
			t.Print()
		}
	case "block":
		baseDir := filepath.Dir(configPath)
		t := tabby.New()
		for _, project := range projects {
			relPath, err := filepath.Rel(baseDir, project.Path)
			CheckIfError(err)

			t.AddLine("Name:", project.Name)
			t.AddLine("Path:", relPath)
			t.AddLine("Description:", project.Description)
			t.AddLine("Url:", project.Url)
			t.AddLine("Tags:", strings.Join(project.Tags, ", "))
			t.AddLine("")
			t.AddLine("")
		}

		t.Print()
	}
}

func PrintCommands(commands []Command, format string, listRaw bool) {
	switch format {
	case "table":
	case "list":
		if listRaw {
			for _, command := range commands {
				fmt.Println(command.Name)
			}
		} else {
			t := tabby.New()
			t.AddHeader("Command", "Description")
			for _, command := range commands {
				t.AddLine(command.Name, command.Description)
			}
			t.Print()
		}
	case "block":
		t := tabby.New()
		for _, command := range commands {
			t.AddLine("Name:", command.Name)
			t.AddLine("Description:", command.Description)
			t.AddLine("Shell:", command.Shell)

			if len(command.Args) > 0 {
				t.AddLine("Args:")
				for key, value := range command.Args {
					t.AddLine(fmt.Sprintf("  - %s=%s", key, value))
				}
			} else {
				t.AddLine("Args:")
			}

			if strings.Count(command.Command, "\n") < 2 {
				t.AddLine("Command:", strings.TrimSpace(command.Command))
				t.AddLine("")
			} else {
				t.AddLine("Command:")
				lines := strings.Split(command.Command, "\n")
				for _, l := range lines {
					t.AddLine(" ", l)
				}
			}
			t.AddLine("")
		}

		t.Print()
	}

}

func PrintCommand(command *Command) {
	fmt.Println(color.Bold("Command:"))
	fmt.Println(command.Command)
}

func PrintTags(tags []string) {
	for _, tag := range tags {
		fmt.Println(tag)
	}
}
