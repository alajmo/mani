package print

import (
	"github.com/alajmo/mani/core"
	"fmt"
	color "github.com/logrusorgru/aurora"
	"github.com/jedib0t/go-pretty/v6/table"
	"os"
)

func PrintRun(format string, outputs map[string]string) {
	if (format == "list") {
		printList(outputs)
	} else {
		printOther(format, outputs)
	}
}

func printList(outputs map[string]string) {
	for projectName, output := range outputs {
		fmt.Println()
		fmt.Println(color.Bold(color.Blue(projectName)))
		fmt.Println(output)
	}
}

func printOther(format string, outputs map[string]string) {
	t := table.NewWriter()
	t.SetOutputMirror(os.Stdout)
	t.SetStyle(core.ManiList)

	t.AppendHeader(table.Row {"Name", "Output"})

	for projectName, output := range outputs {
		t.AppendRow(table.Row { projectName, output })
		t.AppendSeparator()
	}

	switch format {
	case "markdown":
		t.RenderMarkdown()
	case "html":
		t.RenderHTML()
	default:
		t.Render()
	}
}
