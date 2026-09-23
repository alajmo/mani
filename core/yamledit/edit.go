package yamledit

import "strings"

// Root returns the top-level mapping.
func (d *Document) Root() *Mapping { return d.root }

// Entry returns the entry with the given key, or nil if absent.
func (m *Mapping) Entry(key string) *Entry {
	for _, e := range m.entries {
		if e.key == key {
			return e
		}
	}
	return nil
}

// Indent returns the indentation column shared by the mapping's keys.
func (m *Mapping) Indent() int { return m.indent }

// EnsureMapping returns the nested block mapping that is this entry's
// value, creating an empty one (with keys at childIndent) when the value
// is currently empty. childIndent is ignored if a mapping already exists.
func (e *Entry) EnsureMapping(childIndent int) *Mapping {
	if e.child != nil {
		return e.child
	}
	e.child = &Mapping{doc: e.doc, indent: childIndent}
	e.childStart = e.bodyEnd
	e.childEnd = e.bodyEnd
	return e.child
}

// InsertEntry appends a new entry to the mapping. valueLines are the lines
// of the entry's value, each written with its indentation relative to the
// key (e.g. "  url: git@github.com:org/x.git"); InsertEntry prepends the
// mapping's own indentation so the result nests correctly. Keys are not
// deduplicated — the caller must ensure uniqueness.
func (m *Mapping) InsertEntry(key string, valueLines []string) {
	m.insertEntry(key, valueLines, false)
}

// InsertEntrySpaced is like InsertEntry but, when the mapping already has
// entries, separates the new entry from the preceding content with one blank
// line. This matches the per-project spacing `mani init` writes, so a project
// appended by `mani refresh` / `mani project add` doesn't look crammed against
// the entry above it. The first entry of an empty mapping gets no blank line.
func (m *Mapping) InsertEntrySpaced(key string, valueLines []string) {
	m.insertEntry(key, valueLines, true)
}

func (m *Mapping) insertEntry(key string, valueLines []string, spaced bool) {
	ind := strings.Repeat(" ", m.indent)
	eol := m.doc.eol

	var b strings.Builder
	if spaced && len(m.entries) > 0 {
		b.WriteString(eol) // blank line separating from the previous entry
	}
	b.WriteString(ind)
	b.WriteString(key)
	b.WriteString(":")
	b.WriteString(eol)
	for _, vl := range valueLines {
		b.WriteString(ind)
		b.WriteString(vl)
		b.WriteString(eol)
	}
	m.entries = append(m.entries, &Entry{key: key, inserted: true, text: b.String()})
}

// RemoveEntry removes the entry with the given key, together with the
// blank and comment lines directly above it. It reports whether an entry
// was actually removed.
func (m *Mapping) RemoveEntry(key string) bool {
	for i, e := range m.entries {
		if e.key == key {
			m.entries = append(m.entries[:i], m.entries[i+1:]...)
			return true
		}
	}
	return false
}
