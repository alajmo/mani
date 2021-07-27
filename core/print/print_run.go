package print

import (
	"fmt"
	"os"

	"github.com/jedib0t/go-pretty/v6/table"
	color "github.com/logrusorgru/aurora"
)

func PrintRun(output string, data TableOutput) {
	if (output == "list") {
		printList(data)
	} else {
		printTable(output, data)
	}
}

func printList(data TableOutput) {
	for _, row := range data.Rows {
		fmt.Println()
		fmt.Println(color.Bold(row[0])) // Project Name

		for _, out := range row[1:] {
			fmt.Println(out)
		}
	}
}

func printTable(output string, data TableOutput) {
	t := table.NewWriter()
	t.SetOutputMirror(os.Stdout)
	t.SetStyle(ManiList)

	t.AppendHeader(data.Headers)

	for _, row := range data.Rows {
		t.AppendRow(row)
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
