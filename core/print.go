package core

import (
	"fmt"
	color "github.com/logrusorgru/aurora"
	tabby "github.com/cheynewallace/tabby"
)

func PrintProjects(projects []Project, listRaw bool) {
	if (listRaw) {
		for _, project := range projects {
			fmt.Println(project.Name)
		}
	} else {
		t := tabby.New()
		t.AddHeader("Project", "Description")
		for _, project := range projects {
			t.AddLine(project.Name, project.Description)
		}
		t.Print()
	}
}

func PrintCommands(commands []Command, listRaw bool) {
	if (listRaw) {
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

func PrintTags(tags map[string]struct{}) {
	for tag := range tags {
		fmt.Println(tag)
	}
}
