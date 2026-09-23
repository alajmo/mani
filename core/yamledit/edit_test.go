package yamledit

import "testing"

func mustParse(t *testing.T, src string) *Document {
	t.Helper()
	doc, err := Parse([]byte(src))
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	return doc
}

func projectsOf(t *testing.T, doc *Document) *Mapping {
	t.Helper()
	e := doc.Root().Entry("projects")
	if e == nil {
		t.Fatal("projects entry not found")
	}
	return e.EnsureMapping(2)
}

func TestInsertEntry(t *testing.T) {
	src := `projects:
  repo-a:
    url: git@github.com:org/repo-a.git

tasks:
  hello:
    cmd: echo hi
`
	doc := mustParse(t, src)
	projectsOf(t, doc).InsertEntry("repo-b", []string{
		"  url: git@github.com:org/repo-b.git",
		"  tags: [private]",
	})

	want := `projects:
  repo-a:
    url: git@github.com:org/repo-a.git
  repo-b:
    url: git@github.com:org/repo-b.git
    tags: [private]

tasks:
  hello:
    cmd: echo hi
`
	if got := string(doc.Bytes()); got != want {
		t.Fatalf("insert mismatch\n--- want ---\n%s\n--- got ---\n%s", want, got)
	}
}

func TestInsertEntry_FourSpaceIndent(t *testing.T) {
	// Indentation is read off the mapping, so a 4-space file stays 4-space.
	src := `projects:
    repo-a:
        url: a
`
	doc := mustParse(t, src)
	projectsOf(t, doc).InsertEntry("repo-b", []string{"    url: b"})

	want := `projects:
    repo-a:
        url: a
    repo-b:
        url: b
`
	if got := string(doc.Bytes()); got != want {
		t.Fatalf("indent mismatch\n--- want ---\n%s\n--- got ---\n%s", want, got)
	}
}

func TestInsertEntrySpaced(t *testing.T) {
	src := `projects:
  repo-a:
    url: a
`
	doc := mustParse(t, src)
	p := projectsOf(t, doc)
	// First spaced insert into a non-empty mapping is separated from repo-a;
	// the second is separated from the first.
	p.InsertEntrySpaced("repo-b", []string{"  url: b"})
	p.InsertEntrySpaced("repo-c", []string{"  url: c"})

	want := `projects:
  repo-a:
    url: a

  repo-b:
    url: b

  repo-c:
    url: c
`
	if got := string(doc.Bytes()); got != want {
		t.Fatalf("spaced insert mismatch\n--- want ---\n%s\n--- got ---\n%s", want, got)
	}
}

func TestInsertEntrySpaced_FirstEntryNoBlankLine(t *testing.T) {
	// Into an empty `projects:` mapping the first entry gets no leading blank.
	doc := mustParse(t, "projects:\ntasks:\n  t:\n    cmd: x\n")
	doc.Root().Entry("projects").EnsureMapping(2).InsertEntrySpaced("repo-a", []string{"  url: a"})

	want := `projects:
  repo-a:
    url: a
tasks:
  t:
    cmd: x
`
	if got := string(doc.Bytes()); got != want {
		t.Fatalf("first-entry mismatch\n--- want ---\n%s\n--- got ---\n%s", want, got)
	}
}

func TestInsertEntry_NoFinalNewline(t *testing.T) {
	// A source without a trailing newline must not get the new entry jammed
	// onto its last line (regression: previously produced "url: a  repo-b:").
	src := "projects:\n  repo-a:\n    url: a"
	doc := mustParse(t, src)
	projectsOf(t, doc).InsertEntry("repo-b", []string{"  url: b"})

	want := `projects:
  repo-a:
    url: a
  repo-b:
    url: b
`
	if got := string(doc.Bytes()); got != want {
		t.Fatalf("insert mismatch\n--- want ---\n%s\n--- got ---\n%s", want, got)
	}
}

func TestInsertEntry_RootNoFinalNewline(t *testing.T) {
	// Inserting a new top-level key into a newline-less file must not jam it
	// onto the last existing line (regression: previously "bar baz:").
	src := "foo: bar"
	doc := mustParse(t, src)
	doc.Root().InsertEntry("projects", []string{"  a:", "    url: x"})

	want := `foo: bar
projects:
  a:
    url: x
`
	if got := string(doc.Bytes()); got != want {
		t.Fatalf("root insert mismatch\n--- want ---\n%s\n--- got ---\n%s", want, got)
	}
}

func TestInsertEntry_CRLFNoFinalNewline(t *testing.T) {
	// The inserted separator must match the file's newline style.
	src := "projects:\r\n  repo-a:\r\n    url: a"
	doc := mustParse(t, src)
	projectsOf(t, doc).InsertEntry("repo-b", []string{"  url: b"})

	want := "projects:\r\n  repo-a:\r\n    url: a\r\n  repo-b:\r\n    url: b\r\n"
	if got := string(doc.Bytes()); got != want {
		t.Fatalf("crlf insert mismatch\n--- want ---\n%q\n--- got ---\n%q", want, got)
	}
}

func TestRemoveEntry_TakesHeadTrivia(t *testing.T) {
	src := `projects:
  # keep this comment with repo-a
  repo-a:
    url: a

  # this comment belongs to repo-gone
  repo-gone:
    url: gone
  repo-b:
    url: b
tasks:
  t:
    cmd: x
`
	doc := mustParse(t, src)
	if !projectsOf(t, doc).RemoveEntry("repo-gone") {
		t.Fatal("repo-gone not removed")
	}

	// The blank line and comment directly above repo-gone go with it.
	want := `projects:
  # keep this comment with repo-a
  repo-a:
    url: a
  repo-b:
    url: b
tasks:
  t:
    cmd: x
`
	if got := string(doc.Bytes()); got != want {
		t.Fatalf("remove mismatch\n--- want ---\n%s\n--- got ---\n%s", want, got)
	}
}

func TestRemoveEntry_PreservesUntouchedFormatting(t *testing.T) {
	src := `# header
projects:
  a:
    url: a
  b:
    url: b
tasks:
  t:
    cmd: |
      echo hi
`
	doc := mustParse(t, src)
	projectsOf(t, doc).RemoveEntry("a")

	want := `# header
projects:
  b:
    url: b
tasks:
  t:
    cmd: |
      echo hi
`
	if got := string(doc.Bytes()); got != want {
		t.Fatalf("mismatch\n--- want ---\n%s\n--- got ---\n%s", want, got)
	}
}

func TestRemoveEntry_NotFound(t *testing.T) {
	doc := mustParse(t, "projects:\n  a:\n    url: a\n")
	if projectsOf(t, doc).RemoveEntry("missing") {
		t.Fatal("RemoveEntry reported success for a missing key")
	}
}

func TestMapping_Entry(t *testing.T) {
	doc := mustParse(t, "projects:\n  a:\n    url: a\n  b:\n    url: b\n")
	p := projectsOf(t, doc)
	if p.Entry("a") == nil || p.Entry("b") == nil {
		t.Fatal("expected entries a and b")
	}
	if p.Entry("c") != nil {
		t.Fatal("unexpected entry c")
	}
	// A mapping value parses into a child; a scalar value does not.
	if p.Entry("a").child == nil {
		t.Fatal("entry a (mapping value) should have a child mapping")
	}
	if p.Entry("a").child.Entry("url").child != nil {
		t.Fatal("entry url (scalar value) should have no child mapping")
	}
}
