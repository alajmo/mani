package yamledit

import (
	"fmt"
	"strings"
)

// kind classifies a physical line of source.
type kind int

const (
	// kindTrivia is a blank line or a whole-line comment.
	kindTrivia kind = iota
	// kindContent is a structural line: it begins a mapping entry or a
	// sequence item at its own indentation.
	kindContent
	// kindOpaque is a continuation line of a multi-line value (block
	// scalar, multi-line flow collection, or multi-line quoted scalar).
	// The parser never inspects it for structure.
	kindOpaque
)

// srcLine is one physical line, with byte offsets into the original source.
type srcLine struct {
	start  int    // byte offset of the first byte
	end    int    // byte offset just past the line terminator (== len(src) for a final unterminated line)
	indent int    // count of leading ASCII spaces
	text   string // line content WITHOUT the terminator
	kind   kind
}

// scanned is the result of scanLines.
type scanned struct {
	lines []srcLine
	eol   string // "\n" or "\r\n" — the newline used for rendered (inserted) lines
}

// scanLines splits src into physical lines and classifies each one,
// tracking multi-line opaque regions (| / > block scalars, flow
// collections, quoted scalars) so the parser never mistakes their
// interior for structure.
func scanLines(src []byte) (*scanned, error) {
	lines, eol := splitLines(src)

	var (
		inBlock        bool
		blockKeyIndent int
		inFlow         bool
		flowDepth      int
		inQuote        bool
		quoteChar      byte
	)

	for i := range lines {
		ln := &lines[i]
		blank := strings.TrimLeft(ln.text, " ") == ""

		switch {
		case inQuote:
			ln.kind = kindOpaque
			if quoteClosed(ln.text, quoteChar) {
				inQuote = false
			}
			continue
		case inFlow:
			ln.kind = kindOpaque
			flowDepth, inQuote, quoteChar = scanInline(ln.text, flowDepth, false, 0)
			if inQuote {
				inFlow = false
			} else if flowDepth == 0 {
				inFlow = false
			}
			continue
		case inBlock:
			if blank || ln.indent > blockKeyIndent {
				ln.kind = kindOpaque
				continue
			}
			inBlock = false // dedent ends the block scalar; reclassify this line
		}

		if blank {
			ln.kind = kindTrivia
			continue
		}
		if ln.text[ln.indent] == '\t' {
			return nil, fmt.Errorf("yamledit: tab in indentation at line %d", i+1)
		}
		if ln.text[ln.indent] == '#' {
			ln.kind = kindTrivia
			continue
		}
		ln.kind = kindContent

		switch o := classifyValue(valueOf(ln.text)); o.kind {
		case openBlock:
			inBlock = true
			blockKeyIndent = ln.indent
		case openFlow:
			inFlow = true
			flowDepth = o.depth
		case openQuote:
			inQuote = true
			quoteChar = o.quote
		}
	}

	return &scanned{lines: lines, eol: eol}, nil
}

// splitLines breaks src into physical lines and detects the dominant
// newline style.
func splitLines(src []byte) ([]srcLine, string) {
	var lines []srcLine
	crlf := 0
	n := len(src)
	i := 0
	for i < n {
		start := i
		for i < n && src[i] != '\n' {
			i++
		}
		var text string
		var end int
		if i < n { // src[i] == '\n'
			end = i + 1
			if i > start && src[i-1] == '\r' {
				text = string(src[start : i-1])
				crlf++
			} else {
				text = string(src[start:i])
			}
		} else {
			end = i
			text = string(src[start:i])
		}
		indent := 0
		for indent < len(text) && text[indent] == ' ' {
			indent++
		}
		lines = append(lines, srcLine{start: start, end: end, indent: indent, text: text})
		i = end
	}
	eol := "\n"
	if crlf > 0 && crlf == countNewlines(src) {
		eol = "\r\n"
	}
	return lines, eol
}

func countNewlines(src []byte) int {
	c := 0
	for _, b := range src {
		if b == '\n' {
			c++
		}
	}
	return c
}

// valueOf returns the scalar value portion of a content line: the bytes
// after any leading sequence markers ("- ") and a "key:" token. For a
// line with an empty value ("key:") it returns "".
func valueOf(text string) string {
	s := strings.TrimLeft(text, " ")
	for {
		if s == "-" {
			return ""
		}
		if strings.HasPrefix(s, "- ") {
			s = strings.TrimLeft(s[2:], " ")
			continue
		}
		break
	}
	if s == "" {
		return ""
	}
	if s[0] == '"' || s[0] == '\'' {
		end := closeQuoteIndex(s, 0)
		if end < 0 {
			return s // unterminated quoted key — treat whole remainder as value
		}
		rest := strings.TrimLeft(s[end+1:], " ")
		if strings.HasPrefix(rest, ":") {
			return strings.TrimLeft(rest[1:], " ")
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
				return strings.TrimLeft(s[i+1:], " ")
			}
		}
	}
	return s
}

// closeQuoteIndex returns the index of the closing quote for the quoted
// token starting at s[open], or -1 if it is unterminated on this line.
func closeQuoteIndex(s string, open int) int {
	q := s[open]
	for i := open + 1; i < len(s); i++ {
		if s[i] != q {
			continue
		}
		if q == '\'' {
			if i+1 < len(s) && s[i+1] == '\'' {
				i++
				continue
			}
			return i
		}
		// double-quote: an even number of preceding backslashes means real close
		bs := 0
		for k := i - 1; k >= open+1 && s[k] == '\\'; k-- {
			bs++
		}
		if bs%2 == 0 {
			return i
		}
	}
	return -1
}

// quoteClosed reports whether a quoted scalar opened on a previous line
// closes somewhere on this continuation line.
func quoteClosed(text string, q byte) bool {
	for i := 0; i < len(text); i++ {
		if text[i] != q {
			continue
		}
		if q == '\'' {
			if i+1 < len(text) && text[i+1] == '\'' {
				i++
				continue
			}
			return true
		}
		bs := 0
		for k := i - 1; k >= 0 && text[k] == '\\'; k-- {
			bs++
		}
		if bs%2 == 0 {
			return true
		}
	}
	return false
}

type openerKind int

const (
	openNone openerKind = iota
	openBlock
	openFlow
	openQuote
)

type opener struct {
	kind  openerKind
	depth int
	quote byte
}

// classifyValue inspects an inline value and reports whether it opens a
// multi-line construct that the following lines continue.
func classifyValue(val string) opener {
	v := strings.TrimLeft(val, " ")
	if v == "" {
		return opener{}
	}
	switch v[0] {
	case '|', '>':
		if isBlockIndicator(v) {
			return opener{kind: openBlock}
		}
		return opener{}
	case '"', '\'', '[', '{':
		depth, inQuote, qc := scanInline(v, 0, false, 0)
		if inQuote {
			return opener{kind: openQuote, quote: qc}
		}
		if depth > 0 {
			return opener{kind: openFlow, depth: depth}
		}
		return opener{}
	default:
		return opener{}
	}
}

// isBlockIndicator reports whether val is a block-scalar header such as
// "|", ">", "|-", "|+", ">2", "|2-" (optionally followed by a comment).
func isBlockIndicator(val string) bool {
	fields := strings.Fields(val)
	if len(fields) == 0 {
		return false
	}
	tok := fields[0]
	if tok[0] != '|' && tok[0] != '>' {
		return false
	}
	for _, c := range tok[1:] {
		if !strings.ContainsRune("+-0123456789", c) {
			return false
		}
	}
	return true
}

// scanInline walks an inline region (or a flow continuation line),
// updating bracket depth and quote state. A "#" comment outside any
// bracket or quote terminates the scan.
func scanInline(s string, depth int, inQuote bool, q byte) (int, bool, byte) {
	for i := 0; i < len(s); i++ {
		c := s[i]
		if inQuote {
			if c != q {
				continue
			}
			if q == '\'' {
				if i+1 < len(s) && s[i+1] == '\'' {
					i++
					continue
				}
				inQuote = false
				continue
			}
			bs := 0
			for k := i - 1; k >= 0 && s[k] == '\\'; k-- {
				bs++
			}
			if bs%2 == 0 {
				inQuote = false
			}
			continue
		}
		switch c {
		case '\'', '"':
			inQuote = true
			q = c
		case '[', '{':
			depth++
		case ']', '}':
			if depth > 0 {
				depth--
			}
		case '#':
			if depth == 0 && i > 0 && s[i-1] == ' ' {
				return depth, inQuote, q
			}
		}
	}
	return depth, inQuote, q
}
