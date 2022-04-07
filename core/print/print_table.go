package print

import (
	"fmt"

	"github.com/alajmo/mani/core"
	"github.com/alajmo/mani/core/dao"
)

type Items interface {
	GetValue(string, int) string
}

type PrintTableOptions struct {
	Output    string
	Theme     string
	Tree	  bool
	OmitEmpty bool
}

func PrintTable [T Items] (
	config *dao.Config,
	data []T,
	options PrintTableOptions,
	defaultHeaders []string,
	taskHeaders []string,
) {
	// core.DebugPrint(data[0])
	// core.DebugPrint(defaultHeaders)

	theme, err := config.GetTheme(options.Theme)
	core.CheckIfError(err)

	t := CreateTable(theme, options, defaultHeaders, taskHeaders)

	// Headers
	var headers []any
	for _, h := range defaultHeaders {
		headers = append(headers, h)
	}
	for _, h := range taskHeaders {
		headers = append(headers, h)
	}

	t.AppendHeader(headers)

	// Rows
	for _, item := range data {
		var row []any
		if options.OmitEmpty && item.GetValue("", 1) == "" {
			continue
		}

		for i, h := range headers {
			value := item.GetValue(fmt.Sprintf("%v", h), i)
			row = append(row, value)
		}
		t.AppendRow(row)
	}

	RenderTable(t, options.Output)
}
