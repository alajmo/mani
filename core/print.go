package core

import (
	"fmt"
	tabby "github.com/cheynewallace/tabby"
	color "github.com/logrusorgru/aurora"
	"strings"
)

func PrintProjects(projects []Project, listRaw bool) {
	if listRaw {
		for _, project := range projects {
			fmt.Println(project.Name)
		}
	} else {
		t := tabby.New()
		t.AddHeader("Project", "Description", "Tags")
		for _, project := range projects {
			t.AddLine(project.Name, project.Description, strings.Join(project.Tags, ", "))
		}
		t.Print()
	}
}

func PrintCommands(commands []Command, listRaw bool) {
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
