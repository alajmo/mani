package print

type ListFlags struct {
	NoHeaders bool
	NoBorders bool
	Output string
}

type TableOutput struct {
	Headers []string
	Rows []string
}
