package yamledit

// Document is a parsed YAML file that supports lossless structural edits:
// Parse(src) → edits → Bytes() is byte-identical to src outside the edited
// regions. The invariant holds by construction — every unmodified node
// re-emits its original byte span verbatim; only inserted entries are
// rendered.
type Document struct {
	src       []byte
	eol       string // newline used when rendering inserted lines
	root      *Mapping
	tailStart int // src[tailStart:] is trailing trivia after the root mapping
}

// Mapping is an ordered set of block-mapping entries sharing one
// indentation column.
type Mapping struct {
	doc     *Document
	indent  int
	entries []*Entry
}

// Entry is one `key: value` pair together with the trivia (blank and
// comment lines) directly above it. Values are opaque: yamledit models
// structure only and never interprets a value's bytes.
type Entry struct {
	key string
	doc *Document

	// Existing entries — byte spans into Document.src:
	headStart  int // start of the leading trivia block
	keyStart   int // start of the key line (== end of head trivia)
	bodyEnd    int // end of the value block, including its final terminator
	child      *Mapping
	childStart int // byte offset where the nested mapping begins
	childEnd   int // byte offset where the nested mapping ends

	// Inserted entries:
	inserted bool
	text     string // fully rendered text, including the trailing newline
}
