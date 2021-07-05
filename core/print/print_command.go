package print

import (
	"fmt"
	"github.com/alajmo/mani/core"
	tabby "github.com/cheynewallace/tabby"
	"strings"
)

func PrintCommands(commands []core.Command, format string, listRaw bool) {
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
