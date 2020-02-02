package core

import (
	"fmt"
	color "github.com/logrusorgru/aurora"
)

func PrintProjects(projects []Project) {
	for _, project := range projects {
		fmt.Println(project.Name)
	}
}

func PrintCommands(commands []Command) {
	for _, command := range commands {
		fmt.Println(command.Name)
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
