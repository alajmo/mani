package print

import (
	"github.com/jedib0t/go-pretty/v6/table"
)

type ListFlags struct {
	NoHeaders bool
	NoBorders bool
	Output    string
}

type TreeFlags struct {
	Output string
	Tags   []string
}

type TableOutput struct {
	Headers table.Row
	Rows    []table.Row
}
