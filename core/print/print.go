package print

import (
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
