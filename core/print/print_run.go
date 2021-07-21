package print

import (
	"fmt"
	"os"

	"github.com/jedib0t/go-pretty/v6/table"
	color "github.com/logrusorgru/aurora"

	"github.com/alajmo/mani/core/dao"
)

func PrintRun(output string, outputs []TableOutput) {
	if (output == "list") {
		printList(outputs)
	} else {
		printOther(output, outputs)
	}
}

func printList(outputs []TableOutput) {
	for _, out := range outputs {
		fmt.Println()
		fmt.Println(color.Bold(color.Blue(out.Headers)))
		fmt.Println(out.Rows)
	}
}

func printOther(output string, data TableOutput) {
	t := table.NewWriter()
	t.SetOutputMirror(os.Stdout)
	t.SetStyle(ManiList)

	t.AppendHeader(table.Row { data.Headers })

	for _, row := range data.Rows {
		t.AppendRow(table.Row { row })
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
