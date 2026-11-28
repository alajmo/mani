package yamledit

import "strings"

// Root returns the top-level mapping.
func (d *Document) Root() *Mapping { return d.root }

// Entry returns the entry with the given key, or nil if absent.
func (m *Mapping) Entry(key string) *Entry {
	for _, e := range m.entries {
		if e.key == key && !e.removed {
			return e
		}
	}
	return nil
}

// Keys returns the keys of the mapping's entries, in order.
func (m *Mapping) Keys() []string {
	keys := make([]string, 0, len(m.entries))
	for _, e := range m.entries {
		if !e.removed {
			keys = append(keys, e.key)
		}
	}
	return keys
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
//
// The key is written as given, so a key that isn't a plain scalar (`null`,
// `a: b`) has to be quoted by the caller. It's looked up by its value, without
// the quotes, as any other key.
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
	m.entries = append(m.entries, &Entry{key: parseKey(key + ":"), inserted: true, text: b.String()})
}

// RemoveEntry removes the entry with the given key, together with the comment
// lines directly above it and the blank lines that set it apart from the
// entry before. It reports whether an entry was actually removed.
//
// Other comments above the entry stay: a block that is commented out, or a
// heading for the entries that follow, isn't about the entry just because it
// is the next one.
//
//	alpha:
//	  url: a
//
//	# ---- services ----        stays, with the blank line below it
//
//	# the billing service       removed with billing
//	billing:
//	  url: b
func (m *Mapping) RemoveEntry(key string) bool {
	for i, e := range m.entries {
		if e.key != key || e.removed {
			continue
		}

		keptEnd := e.headStart
		if !e.inserted {
			keptEnd = e.headStart + keptHead(m.doc.src[e.headStart:e.keyStart])
		}
		if keptEnd > e.headStart {
			e.removed = true
			e.keptEnd = keptEnd
		} else {
			m.entries = append(m.entries[:i], m.entries[i+1:]...)
		}
		return true
	}
	return false
}

// keptHead returns how much of the head trivia of an entry stays when the
// entry is removed: everything up to the comment lines directly above the
// key. Nothing stays when only blank lines are above those.
func keptHead(head []byte) int {
	var lines []string
	for rest := string(head); rest != ""; {
		line := rest
		if i := strings.IndexByte(rest, '\n'); i >= 0 {
			line = rest[:i+1]
		}
		lines = append(lines, line)
		rest = rest[len(line):]
	}

	i := len(lines)
	for i > 0 && strings.HasPrefix(strings.TrimLeft(lines[i-1], " "), "#") {
		i--
	}
	attached := i
	for i > 0 && strings.TrimSpace(lines[i-1]) == "" {
		i--
	}
	if i == 0 {
		return 0
	}

	kept := 0
	for _, line := range lines[:attached] {
		kept += len(line)
	}
	return kept
}
