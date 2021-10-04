package print

import (
	"fmt"
	"os"

	"github.com/jedib0t/go-pretty/v6/table"
	color "github.com/logrusorgru/aurora"
)

func PrintRun(output string, data TableOutput) {
	if output == "list" || output == "" {
		printList(data)
	} else {
		printTable(output, data)
	}
}

func printList(data TableOutput) {
	for _, row := range data.Rows {
		fmt.Println()
		fmt.Println(color.Bold(row[0])) // Project Name

		fmt.Println(row[1])
		fmt.Println()

		// Print headers for sub-commands
		for i, out := range row[2:] {
			fmt.Printf("# %v\n", data.Headers[i+2])
			fmt.Println(out)
			fmt.Println()
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
