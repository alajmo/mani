package print

import (
	"github.com/alajmo/mani/core"
	"fmt"
	"github.com/jedib0t/go-pretty/v6/table"
	color "github.com/logrusorgru/aurora"
	"os"
)

func PrintRun(format string, outputs []core.ProjectOutput) {
	if (format == "list") {
		printList(outputs)
	} else {
		printOther(format, outputs)
	}
}

func printList(outputs []core.ProjectOutput) {
	for _, output := range outputs {
		fmt.Println()
		fmt.Println(color.Bold(color.Blue(output.ProjectName)))
		fmt.Println(output.Output)
	}
}

func printOther(format string, outputs []core.ProjectOutput) {
	t := table.NewWriter()
	t.SetOutputMirror(os.Stdout)
	t.SetStyle(core.ManiList)

	t.AppendHeader(table.Row {"Name", "Output"})

	for _, output := range outputs {
		t.AppendRow(table.Row { output.ProjectName, output.Output })
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
