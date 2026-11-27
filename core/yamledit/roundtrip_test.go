package yamledit

import (
	"bytes"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// roundTripCorpus exercises the core invariant: Parse(src) → Bytes() is
// byte-identical to src. Each entry stresses a different formatting trait.
var roundTripCorpus = map[string]string{
	"two-space": `projects:
  repo-a:
    url: git@github.com:org/repo-a.git
tasks:
  hello:
    cmd: echo hi
`,
	"four-space": `projects:
    repo-a:
        url: git@github.com:org/repo-a.git
`,
	"blanks-and-comments": `# top comment

sources:
  - source: github/my-org

# projects below
projects:

  # a hand-edited project
  repo-a:
    url: git@github.com:org/repo-a.git   # inline comment

  repo-b:
    url: git@github.com:org/repo-b.git


tasks:
  hello:
    cmd: echo hi
`,
	"block-scalars": `tasks:
  build:
    cmd: |
      set -e
      go build ./...

      echo done
  doc:
    cmd: >
      folded
      text
projects:
  a:
    desc: x
`,
	"flow-collections": `projects:
  a:
    tags: [private, archived]
  b:
    tags:
      - one
      - two
`,
	"multiline-flow": `projects:
  a:
    tags: [
      private,
      archived,
    ]
  b:
    url: x
`,
	"anchors-opaque": `projects:
  base: &base
    tags: [core]
  a:
    <<: *base
    url: x
`,
	"quoted-values": `projects:
  a:
    desc: "a: tricky value"
    clone: 'single quoted'
`,
	"no-final-newline": "projects:\n  a:\n    url: x",
	"trailing-blanks":  "projects:\n  a:\n    url: x\n\n\n",
}

func TestRoundTrip(t *testing.T) {
	for name, src := range roundTripCorpus {
		t.Run(name, func(t *testing.T) {
			doc, err := Parse([]byte(src))
			if err != nil {
				t.Fatalf("parse: %v", err)
			}
			if got := string(doc.Bytes()); got != src {
				t.Fatalf("round-trip mismatch\n--- want ---\n%q\n--- got ---\n%q", src, got)
			}
		})
	}
}

func TestRoundTrip_CRLF(t *testing.T) {
	src := "projects:\r\n  a:\r\n    url: x\r\n"
	doc, err := Parse([]byte(src))
	if err != nil {
		t.Fatal(err)
	}
	if got := string(doc.Bytes()); got != src {
		t.Fatalf("CRLF round-trip mismatch: %q", got)
	}
}

func TestParse_RejectsMultiDoc(t *testing.T) {
	_, err := Parse([]byte("projects:\n  a: x\n---\nfoo: bar\n"))
	if err == nil || !strings.Contains(err.Error(), "multi-document") {
		t.Fatalf("expected multi-document error, got %v", err)
	}
}

func TestParse_RejectsTabIndent(t *testing.T) {
	_, err := Parse([]byte("projects:\n\ta: x\n"))
	if err == nil || !strings.Contains(err.Error(), "tab") {
		t.Fatalf("expected tab error, got %v", err)
	}
}

func TestParse_TabInsideBlockScalarIsFine(t *testing.T) {
	// A tab in block-scalar content is data, not indentation — it must not
	// trip the tab-in-indentation guard.
	src := "tasks:\n  t:\n    cmd: |\n      echo\tindented\n"
	doc, err := Parse([]byte(src))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got := string(doc.Bytes()); got != src {
		t.Fatalf("round-trip mismatch: %q", got)
	}
}

// A block sequence may sit at the same indentation as its key. It's still the
// value of that key, and so are the keys that follow it.
var indentlessCorpus = map[string]string{
	"in-project": `projects:
  foo:
    url: x
    tags:
    - a
    - b
    desc: hello
  bar:
    url: y
`,
	"at-root": `import:
- a.yaml
- b.yaml
projects:
  foo:
    url: x
tasks:
  t: echo
`,
	"multi-line-items": `projects:
  foo:
    url: x
tasks:
  t:
    env:
      A: 1
    commands:
    - name: one
      cmd: echo 1
    - cmd: |
        echo 2
        echo 3
    desc: after
`,
	"comments-between-items": `projects:
  foo:
    tags:
    # first
    - a

    - b
    url: x
`,
	"last-without-newline": "projects:\n  foo:\n    tags:\n    - a\n    - b",
	"crlf":                 "projects:\r\n  foo:\r\n    tags:\r\n    - a\r\n    url: x\r\n",
	"nested-twice": `projects:
  foo:
    worktrees:
    - path: a
      branch: b
    remotes:
      origin: x
`,
}

func TestRoundTrip_IndentlessSequence(t *testing.T) {
	for name, src := range indentlessCorpus {
		t.Run(name, func(t *testing.T) {
			doc, err := Parse([]byte(src))
			if err != nil {
				t.Fatalf("parse: %v", err)
			}
			if got := string(doc.Bytes()); got != src {
				t.Fatalf("round-trip mismatch\n--- want ---\n%q\n--- got ---\n%q", src, got)
			}
		})
	}
}

// Keys after an indentless sequence are still entries of the mapping.
func TestParse_IndentlessSequenceKeepsSiblings(t *testing.T) {
	doc := mustParse(t, indentlessCorpus["at-root"])
	if got, want := doc.Root().Keys(), []string{"import", "projects", "tasks"}; !equalStrings(got, want) {
		t.Fatalf("root keys = %v; want %v", got, want)
	}

	doc = mustParse(t, indentlessCorpus["in-project"])
	projects := projectsOf(t, doc)
	if got, want := projects.Keys(), []string{"foo", "bar"}; !equalStrings(got, want) {
		t.Fatalf("project keys = %v; want %v", got, want)
	}
}

func TestParse_DocumentMarkers(t *testing.T) {
	// A single document may start with `---` and end with `...`
	for _, src := range []string{
		"---\nprojects:\n  a:\n    url: x\n",
		"# comment\n---\nprojects:\n  a:\n    url: x\n",
		"--- # config\nprojects:\n  a:\n    url: x\n",
		"---\nprojects:\n  a:\n    url: x\n...\n",
	} {
		doc, err := Parse([]byte(src))
		if err != nil {
			t.Fatalf("%q: unexpected error: %v", src, err)
		}
		if got := string(doc.Bytes()); got != src {
			t.Fatalf("round-trip mismatch\n--- want ---\n%q\n--- got ---\n%q", src, got)
		}
		if doc.Root().Entry("projects") == nil {
			t.Fatalf("%q: projects not found", src)
		}
	}

	// More than one document, or content on the marker line, is not supported
	for _, src := range []string{
		"projects:\n  a: x\n---\nfoo: bar\n",
		"---\nprojects:\n  a: x\n---\nfoo: bar\n",
		"---\nprojects:\n  a: x\n...\nfoo: bar\n",
		"--- foo: bar\n",
	} {
		if _, err := Parse([]byte(src)); err == nil || !strings.Contains(err.Error(), "multi-document") {
			t.Fatalf("%q: expected multi-document error, got %v", src, err)
		}
	}
}

func equalStrings(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

// Every config in the repository has to be reproduced byte for byte.
func TestRoundTrip_RepositoryConfigs(t *testing.T) {
	var files []string
	for _, root := range []string{"../../test/fixtures", "../../test/integration/golden", "../../examples"} {
		err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if ext := filepath.Ext(path); !d.IsDir() && (ext == ".yaml" || ext == ".yml") {
				files = append(files, path)
			}
			return nil
		})
		if err != nil {
			t.Fatal(err)
		}
	}
	if len(files) < 50 {
		t.Fatalf("found %d configs, expected the fixtures and golden files", len(files))
	}

	for _, file := range files {
		src, err := os.ReadFile(file)
		if err != nil {
			t.Fatal(err)
		}
		doc, err := Parse(src)
		if err != nil {
			t.Errorf("%s: %v", file, err)
			continue
		}
		if !bytes.Equal(doc.Bytes(), src) {
			t.Errorf("%s: round-trip mismatch", file)
		}
	}
}
