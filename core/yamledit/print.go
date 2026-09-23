package yamledit

import "bytes"

// Bytes renders the document. For an unmodified Document the result is
// byte-identical to the source passed to Parse.
func (d *Document) Bytes() []byte {
	if d.root == nil || len(d.root.entries) == 0 {
		return d.src
	}
	var b bytes.Buffer
	b.Grow(len(d.src) + 256)
	d.root.write(&b, d.src, d.eol)
	b.Write(d.src[d.tailStart:])
	return b.Bytes()
}

func (m *Mapping) write(b *bytes.Buffer, src []byte, eol string) {
	for _, e := range m.entries {
		e.write(b, src, eol)
	}
}

func (e *Entry) write(b *bytes.Buffer, src []byte, eol string) {
	if e.inserted {
		// An inserted entry's text begins at its own indentation, so the
		// preceding content must end on its own line. When the source has no
		// final newline (or a nested mapping is appended to a `key:`-only line
		// that lacks one), the verbatim bytes written so far don't end in a
		// terminator and the new entry would be jammed onto them — so add one.
		// This is the only place rendering can concatenate across a line that
		// the source did not terminate.
		ensureNewline(b, eol)
		b.WriteString(e.text)
		return
	}
	b.Write(src[e.headStart:e.keyStart])
	if e.child != nil {
		b.Write(src[e.keyStart:e.childStart])
		e.child.write(b, src, eol)
		b.Write(src[e.childEnd:e.bodyEnd])
		return
	}
	b.Write(src[e.keyStart:e.bodyEnd])
}

// ensureNewline appends eol unless the buffer is empty or already ends in a
// line terminator (a trailing "\n" covers both "\n" and "\r\n").
func ensureNewline(b *bytes.Buffer, eol string) {
	if bs := b.Bytes(); len(bs) > 0 && bs[len(bs)-1] != '\n' {
		b.WriteString(eol)
	}
}
