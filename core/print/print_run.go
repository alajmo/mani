package print

import (
	"fmt"
	"github.com/jedib0t/go-pretty/v6/table"
	color "github.com/logrusorgru/aurora"
	"os"

	"github.com/alajmo/mani/core/dao"
)

func PrintRun(output string, outputs []dao.ProjectOutput) {
	if (output == "list") {
		printList(outputs)
	} else {
		printOther(output, outputs)
	}
}

func printList(outputs []dao.ProjectOutput) {
	for _, output := range outputs {
		fmt.Println()
		fmt.Println(color.Bold(color.Blue(output.ProjectName)))
		fmt.Println(output.Output)
	}
}

func printOther(output string, outputs []dao.ProjectOutput) {
	t := table.NewWriter()
	t.SetOutputMirror(os.Stdout)
	t.SetStyle(ManiList)

	t.AppendHeader(table.Row {"Name", "Output"})

	for _, output := range outputs {
		t.AppendRow(table.Row { output.ProjectName, output.Output })
		t.AppendSeparator()
	}

	switch output {
	case "markdown":
		t.RenderMarkdown()
	case "html":
		t.RenderHTML()
	default:
		t.Render()
	}
}
