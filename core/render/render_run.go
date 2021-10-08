package render

import (
	"fmt"
	"os"

	"github.com/jedib0t/go-pretty/v6/table"
	color "github.com/logrusorgru/aurora"

	"github.com/alajmo/mani/core"
)

// TASK [<name>: <description>] ************>
// <project|dir|host> | OUTPUT

func Render(output string, data core.TableOutput) {
	// if runFlags.Describe {
	// 	render.PrintTaskBlock([]Task{*t})
	// }

	// Table Style
	// switch config.Theme.Table {
	// case "ascii":
	// 	core.ManiList.Box = core.StyleBoxASCII
	// default:
	// 	core.ManiList.Box = core.StyleBoxDefault
	// }

	if output == "list" || output == "" {
		printList(data)
	} else {
		printTable(output, data)
	}
}

func printList(data core.TableOutput) {
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

func printTable(output string, data core.TableOutput) {
	t := table.NewWriter()
	t.SetOutputMirror(os.Stdout)
	t.SetStyle(core.ManiList)

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
