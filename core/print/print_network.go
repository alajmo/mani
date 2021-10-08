package print

import (
	"fmt"
	"github.com/jedib0t/go-pretty/v6/table"
	"os"

	"github.com/alajmo/mani/core"
	"github.com/alajmo/mani/core/dao"
)

type ListNetworkFlags struct {
	Name    string
	Tags    []string
	Headers []string
	Edit    bool
}

func PrintNetworks(
	networks []dao.Network,
	listFlags ListFlags,
	networkFlags ListNetworkFlags,
) {
	t := table.NewWriter()
	t.SetOutputMirror(os.Stdout)
	t.SetStyle(core.ManiList)

	var headers []interface{}
	for _, h := range networkFlags.Headers {
		headers = append(headers, h)
	}

	if !listFlags.NoHeaders {
		t.AppendHeader(headers)
	}

	for _, network := range networks {
		var row []interface{}
		for _, h := range headers {
			value := network.GetValue(fmt.Sprintf("%v", h))
			row = append(row, value)
		}

		t.AppendRow(row)
	}

	if listFlags.NoBorders {
		t.Style().Box = core.StyleNoBorders
		t.Style().Options.SeparateHeader = false
		t.Style().Options.DrawBorder = false
	}

	switch listFlags.Output {
	case "markdown":
		t.RenderMarkdown()
	case "html":
		t.RenderHTML()
	default:
		t.Render()
	}
}

func PrintNetworkBlocks(networks []dao.Network) {
	t := table.NewWriter()
	t.SetOutputMirror(os.Stdout)
	t.SetStyle(core.ManiList)

	for _, network := range networks {
		t.AppendRows([]table.Row{
			{"Name: ", network.Name},
			{"Description: ", network.Description},
			{"Hosts: ", network.GetValue("Hosts")},
			{"Tags: ", network.GetValue("Tags")},
		})

		t.AppendSeparator()
		t.AppendRow(table.Row{})
		t.AppendSeparator()
	}

	t.Style().Box = core.StyleNoBorders
	t.Style().Options.SeparateHeader = false
	t.Style().Options.DrawBorder = false

	t.Render()
}
