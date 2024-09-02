package print

import (
	"io"

	"github.com/alajmo/mani/core/dao"
	"github.com/jedib0t/go-pretty/v6/table"
)

type Items interface {
	GetValue(string, int) string
}

type PrintTableOptions struct {
	Output           string
	Theme            dao.Theme
	Tree             bool
	Color            bool
	AutoWrap         bool
	OmitEmptyRows    bool
	OmitEmptyColumns bool
}

func PrintTable[T Items](
	data []T,
	options PrintTableOptions,
	defaultHeaders []string,
	taskHeaders []string,
	writer io.Writer,
) {
	// Colors not supported for markdown and html
	switch options.Output {
	case "markdown":
		options.Color = false
	case "html":
		options.Color = false
	}

	t := CreateTable(options, defaultHeaders, taskHeaders, data, writer)
	theme := options.Theme

	// Headers
	var headers table.Row
	for _, h := range defaultHeaders {
		headers = append(headers, dao.StyleString(h, *theme.Table.Header, options.Color))
	}
	for _, h := range taskHeaders {
		headers = append(headers, dao.StyleString(h, *theme.Table.Header, options.Color))
	}
	t.AppendHeader(headers)

	// Rows
	headerNames := append(defaultHeaders, taskHeaders...)
	for _, item := range data {
		row := table.Row{}
		for i, h := range headerNames {
			value := item.GetValue(h, i)
			if i == 0 {
				value = dao.StyleString(value, *theme.Table.TitleColumn, options.Color)
			}
			row = append(row, value)
		}

		if options.OmitEmptyRows {
			empty := true
			for _, v := range row[1:] {
				if v != "" {
					empty = false
					break
				}
			}
			if empty {
				continue
			}
		}
		t.AppendRow(row)
	}

	RenderTable(t, options.Output)
}
