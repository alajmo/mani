package print

import (
	"io"

	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/jedib0t/go-pretty/v6/text"
	"golang.org/x/term"

	"github.com/alajmo/mani/core/dao"
)

func CreateTable[T Items](
	options PrintTableOptions,
	defaultHeaders []string,
	taskHeaders []string,
	data []T,
	writer io.Writer,
) table.Writer {
	t := table.NewWriter()

	theme := options.Theme

	t.SetOutputMirror(writer)
	t.SetStyle(FormatTable(theme))

	if options.OmitEmptyColumns {
		t.SuppressEmptyColumns()
	}

	canWrap, maxColumnWidths := calcColumnWidths(defaultHeaders, taskHeaders, data)

	headers := []table.ColumnConfig{}
	for i := range defaultHeaders {
		headerStyle := table.ColumnConfig{
			Number: i + 1,
		}
		if options.AutoWrap && canWrap {
			headerStyle.WidthMaxEnforcer = text.WrapText
			headerStyle.WidthMax = maxColumnWidths[i]
		}
		headers = append(headers, headerStyle)
	}

	for i := range taskHeaders {
		offset := len(defaultHeaders) + i
		headerStyle := table.ColumnConfig{
			Number: len(defaultHeaders) + 1 + i,
		}

		if options.AutoWrap && canWrap {
			headerStyle.WidthMaxEnforcer = text.WrapText
			headerStyle.WidthMax = maxColumnWidths[offset]
		}
		headers = append(headers, headerStyle)
	}

	t.SetColumnConfigs(headers)

	return t
}

func FormatTable(theme dao.Theme) table.Style {
	return table.Style{
		Name: theme.Name,
		Box:  theme.Table.Box,

		Options: table.Options{
			DrawBorder:      *theme.Table.Border.Around,
			SeparateColumns: *theme.Table.Border.Columns,
			SeparateHeader:  *theme.Table.Border.Header,
			SeparateRows:    *theme.Table.Border.Rows,
		},
	}
}

func RenderTable(t table.Writer, output string) {
	switch output {
	case "markdown":
		t.RenderMarkdown()
	case "html":
		t.RenderHTML()
	default:
		t.Render()
	}
}

func calcColumnWidths[T Items](
	defaultHeaders []string,
	taskHeaders []string,
	data []T,
) (bool, []int) {
	headers := append(defaultHeaders, taskHeaders...)
	columnWidths := make([]int, len(headers))
	headerPaddingsSum := 3*len(headers) + 1

	// Initialize column widths based on headers
	for i, header := range headers {
		columnWidths[i] = GetMaxTextWidth(header)
	}

	// Update column widths based on rows
	for _, row := range data {
		for j, column := range headers {
			value := row.GetValue(column, j)
			columnWidth := GetMaxTextWidth(value)
			if columnWidths[j] < columnWidth {
				columnWidths[j] = columnWidth
			}
		}
	}

	// Calculate total width and check against terminal width
	columnSum := headerPaddingsSum
	for _, width := range columnWidths {
		columnSum += width
	}

	terminalWidth, _, _ := term.GetSize(0)
	if columnSum < terminalWidth {
		return false, columnWidths
	}

	maxColumnWidth := (terminalWidth - headerPaddingsSum) / (len(columnWidths))
	var affectedColumns []int
	for i := range columnWidths {
		if columnWidths[i] > maxColumnWidth {
			columnWidths[i] = maxColumnWidth
			affectedColumns = append(affectedColumns, i)
		}
	}

	columnSum = headerPaddingsSum
	for _, width := range columnWidths {
		columnSum += width
	}

	addToEach := (terminalWidth - columnSum) / len(affectedColumns)
	for _, col := range affectedColumns {
		columnWidths[col] += addToEach
	}

	return true, columnWidths
}
