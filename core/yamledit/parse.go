package yamledit

import (
	"fmt"
	"strings"
)

// Parse reads src into a Document. The top level is expected to be a block
// mapping (every mani.yaml is). Multi-document streams are rejected with a
// clear error rather than silently mangled.
func Parse(src []byte) (*Document, error) {
	sc, err := scanLines(src)
	if err != nil {
		return nil, err
	}
	for i, l := range sc.lines {
		if l.kind == kindOpaque {
			continue
		}
		t := strings.TrimRight(l.text, " ")
		if t == "---" || t == "..." || strings.HasPrefix(t, "--- ") || strings.HasPrefix(t, "... ") {
			return nil, fmt.Errorf("yamledit: multi-document streams are not supported (line %d)", i+1)
		}
	}

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
		return doc, nil
	}

	root, next := parseMapping(doc, sc, 0, len(sc.lines), rootIndent)
	doc.root = root
	if next < len(sc.lines) {
		doc.tailStart = sc.lines[next].start
	} else {
		doc.tailStart = len(src)
	}
	return doc, nil
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

		keyIdx := i
		i++
		for i < hi {
			l := sc.lines[i]
			if l.kind == kindContent && l.indent <= indent {
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
		child, _ := parseMapping(doc, sc, keyIdx+1, valEnd, childIndent)
		if len(child.entries) > 0 {
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
