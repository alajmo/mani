package yamledit

import (
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
