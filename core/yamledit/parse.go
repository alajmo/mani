package yamledit

import (
	"bytes"
	"errors"
	"fmt"
	"strings"
)

// Parse reads src into a Document. The top level is expected to be a block
// mapping (every mani.yaml is). Multi-document streams are rejected with a
// clear error rather than silently mangled.
//
// A Document that doesn't reproduce src byte for byte is rejected as well, so
// that a layout the parser doesn't understand results in an error and never in
// an edit that loses content.
func Parse(src []byte) (*Document, error) {
	sc, err := scanLines(src)
	if err != nil {
		return nil, err
	}
	if err := markDocumentMarkers(sc); err != nil {
		return nil, err
	}

	doc := parseDocument(src, sc)
	if !bytes.Equal(doc.Bytes(), src) {
		return nil, errors.New("yamledit: the layout of the document is not supported")
	}

	return doc, nil
}

// markDocumentMarkers turns the markers of a single document into trivia: a
// `---` before the first content line, and a `...` after the last. They are
// kept verbatim, as any other line the parser doesn't look into. Any other
// marker means more than one document, which is not supported.
func markDocumentMarkers(sc *scanned) error {
	multiDoc := func(i int) error {
		return fmt.Errorf("yamledit: multi-document streams are not supported (line %d)", i+1)
	}

	seenContent, seenStart, seenEnd := false, false, false
	for i := range sc.lines {
		l := &sc.lines[i]
		if l.kind == kindOpaque || l.kind == kindTrivia {
			continue
		}

		marker, rest, _ := strings.Cut(strings.TrimRight(l.text, " "), " ")
		bare := rest == "" || strings.HasPrefix(strings.TrimLeft(rest, " "), "#")
		switch {
		case seenEnd:
			// Nothing but trivia may follow the end of the document
			return multiDoc(i)
		case marker == "---" && l.indent == 0:
			if seenContent || seenStart || !bare {
				return multiDoc(i)
			}
			seenStart = true
			l.kind = kindTrivia
		case marker == "..." && l.indent == 0:
			if !bare {
				return multiDoc(i)
			}
			seenEnd = true
			l.kind = kindTrivia
		default:
			seenContent = true
		}
	}

	return nil
}

func parseDocument(src []byte, sc *scanned) *Document {
	doc := &Document{src: src, eol: sc.eol}

	rootIndent := -1
	for _, l := range sc.lines {
		if l.kind == kindContent {
			rootIndent = l.indent
			break
		}
	}
	if rootIndent < 0 {
		doc.root = &Mapping{doc: doc}
		doc.tailStart = 0
		return doc
	}

	root, next := parseMapping(doc, sc, 0, len(sc.lines), rootIndent)
	doc.root = root
	if next < len(sc.lines) {
		doc.tailStart = sc.lines[next].start
	} else {
		doc.tailStart = len(src)
	}
	return doc
}

// parseMapping parses the block-mapping entries at exactly `indent` within
// the line range [lo, hi). It returns the mapping and the index of the
// first line it did not consume — trailing trivia is left for the caller
// so it can be attached as the next sibling's head (or the document tail).
func parseMapping(doc *Document, sc *scanned, lo, hi, indent int) (*Mapping, int) {
	m := &Mapping{doc: doc, indent: indent}
	i := lo
	for i < hi {
		headLo := i
		for i < hi && sc.lines[i].kind == kindTrivia {
			i++
		}
		if i >= hi {
			return m, headLo
		}
		ln := sc.lines[i]
		if ln.kind != kindContent || ln.indent != indent || isSeqItem(ln.text) {
			// Dedent, unexpected indentation, or a sequence at this level:
			// not a mapping entry. Hand the gathered trivia back.
			return m, headLo
		}

		// The value runs until the next line that isn't indented deeper than
		// the key. The exception is a block sequence, which may sit at the
		// indentation of its key and is still the value of that key:
		//
		//   tags:
		//   - a
		//   - b
		keyIdx := i
		i++
		for i < hi {
			l := sc.lines[i]
			if l.kind == kindContent && (l.indent < indent || (l.indent == indent && !isSeqItem(l.text))) {
				break
			}
			i++
		}
		valEnd := i
		for valEnd > keyIdx+1 && sc.lines[valEnd-1].kind == kindTrivia {
			valEnd--
		}

		m.entries = append(m.entries, buildEntry(doc, sc, headLo, keyIdx, valEnd, indent))
		i = valEnd
	}
	return m, i
}

func buildEntry(doc *Document, sc *scanned, headLo, keyIdx, valEnd, indent int) *Entry {
	e := &Entry{
		key:       parseKey(sc.lines[keyIdx].text),
		doc:       doc,
		headStart: sc.lines[headLo].start,
		keyStart:  sc.lines[keyIdx].start,
		bodyEnd:   sc.lines[valEnd-1].end,
	}

	firstContent := -1
	for j := keyIdx + 1; j < valEnd; j++ {
		if sc.lines[j].kind == kindContent {
			firstContent = j
			break
		}
	}
	if firstContent >= 0 &&
		sc.lines[firstContent].indent > indent &&
		!isSeqItem(sc.lines[firstContent].text) {
		childIndent := sc.lines[firstContent].indent
		child, next := parseMapping(doc, sc, keyIdx+1, valEnd, childIndent)
		// A mapping that doesn't cover the whole value isn't one the parser
		// understands. The value is then kept as is, since the lines after
		// the mapping would otherwise belong to no entry and be lost.
		if len(child.entries) > 0 && next == valEnd {
			e.child = child
			e.childStart = sc.lines[keyIdx+1].start
			e.childEnd = e.bodyEnd
		}
	}
	return e
}

func isSeqItem(text string) bool {
	s := strings.TrimLeft(text, " ")
	return s == "-" || strings.HasPrefix(s, "- ")
}

// parseKey extracts the key of a mapping-entry line for lookup. Quoted
// keys are unwrapped; plain keys are returned trimmed of surrounding
// whitespace.
func parseKey(text string) string {
	s := strings.TrimLeft(text, " ")
	if s == "" {
		return ""
	}
	if s[0] == '"' || s[0] == '\'' {
		if end := closeQuoteIndex(s, 0); end > 0 {
			return unquoteKey(s[:end+1])
		}
		return s
	}
	depth := 0
	inq := false
	var qc byte
	for i := 0; i < len(s); i++ {
		c := s[i]
		if inq {
			if c == qc {
				inq = false
			}
			continue
		}
		switch c {
		case '\'', '"':
			inq = true
			qc = c
		case '[', '{':
			depth++
		case ']', '}':
			if depth > 0 {
				depth--
			}
		case ':':
			if depth == 0 && (i+1 >= len(s) || s[i+1] == ' ') {
				return strings.TrimRight(s[:i], " ")
			}
		}
	}
	return strings.TrimRight(s, " ")
}

func unquoteKey(tok string) string {
	if len(tok) < 2 {
		return tok
	}
	inner := tok[1 : len(tok)-1]
	if tok[0] == '\'' {
		return strings.ReplaceAll(inner, "''", "'")
	}
	return strings.NewReplacer(`\"`, `"`, `\\`, `\`, `\n`, "\n", `\t`, "\t").Replace(inner)
}
