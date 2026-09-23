package dao

import (
	"errors"
	"testing"

	"github.com/alajmo/mani/core"
)

func TestAddProject_AppendsPreservingFormatting(t *testing.T) {
	input := `# my workspace
projects:
  # hand-written
  repo-a:
    url: git@github.com:org/repo-a.git
    tags: [core]

tasks:
  hello:
    cmd: |
      echo hi
`
	path := writeConfig(t, input)

	err := AddProject(path, Project{
		Name: "repo-b",
		Path: "repo-b",
		URL:  "git@github.com:org/repo-b.git",
		Tags: []string{"backend"},
	})
	if err != nil {
		t.Fatal(err)
	}

	want := `# my workspace
projects:
  # hand-written
  repo-a:
    url: git@github.com:org/repo-a.git
    tags: [core]

  repo-b:
    url: git@github.com:org/repo-b.git
    tags: [backend]

tasks:
  hello:
    cmd: |
      echo hi
`
	if got := readConfig(t, path); got != want {
		t.Fatalf("add mismatch\n--- want ---\n%s\n--- got ---\n%s", want, got)
	}
}

func TestAddProject_PathOmittedWhenEqualToName(t *testing.T) {
	path := writeConfig(t, "projects:\n  a:\n    url: x\n")
	if err := AddProject(path, Project{Name: "b", Path: "b"}); err != nil {
		t.Fatal(err)
	}
	// path == name, no url/tags/desc -> bare key, nothing else emitted.
	want := "projects:\n  a:\n    url: x\n\n  b:\n"
	if got := readConfig(t, path); got != want {
		t.Fatalf("mismatch\n--- want ---\n%q\n--- got ---\n%q", want, got)
	}
}

func TestAddProject_CustomPathIsWritten(t *testing.T) {
	path := writeConfig(t, "projects:\n  a:\n    url: x\n")
	if err := AddProject(path, Project{Name: "b", Path: "libs/b", Desc: "lib"}); err != nil {
		t.Fatal(err)
	}
	want := "projects:\n  a:\n    url: x\n\n  b:\n    path: libs/b\n    desc: lib\n"
	if got := readConfig(t, path); got != want {
		t.Fatalf("mismatch\n--- want ---\n%q\n--- got ---\n%q", want, got)
	}
}

func TestAddProject_PopulatesEmptyProjectsBlock(t *testing.T) {
	path := writeConfig(t, "sources:\n  - source: github/org\nprojects:\ntasks:\n  t:\n    cmd: x\n")
	if err := AddProject(path, Project{Name: "a", Path: "a", URL: "u"}); err != nil {
		t.Fatal(err)
	}
	want := "sources:\n  - source: github/org\nprojects:\n  a:\n    url: u\ntasks:\n  t:\n    cmd: x\n"
	if got := readConfig(t, path); got != want {
		t.Fatalf("mismatch\n--- want ---\n%q\n--- got ---\n%q", want, got)
	}
}

func TestAddProject_DuplicateErrors(t *testing.T) {
	path := writeConfig(t, "projects:\n  a:\n    url: x\n")
	err := AddProject(path, Project{Name: "a", Path: "a", URL: "y"})
	var dup *core.ProjectAlreadyExists
	if !errors.As(err, &dup) {
		t.Fatalf("expected ProjectAlreadyExists, got %v", err)
	}
	// File must be untouched on error.
	if got := readConfig(t, path); got != "projects:\n  a:\n    url: x\n" {
		t.Fatalf("file was modified on duplicate add: %q", got)
	}
}

func TestRemoveProject_TakesHeadTrivia(t *testing.T) {
	input := `projects:
  repo-a:
    url: a

  # belongs to repo-b
  repo-b:
    url: b
  repo-c:
    url: c
tasks:
  t:
    cmd: x
`
	path := writeConfig(t, input)
	if err := RemoveProject(path, "repo-b"); err != nil {
		t.Fatal(err)
	}

	want := `projects:
  repo-a:
    url: a
  repo-c:
    url: c
tasks:
  t:
    cmd: x
`
	if got := readConfig(t, path); got != want {
		t.Fatalf("remove mismatch\n--- want ---\n%s\n--- got ---\n%s", want, got)
	}
}

func TestRemoveProject_NotFoundErrors(t *testing.T) {
	path := writeConfig(t, "projects:\n  a:\n    url: x\n")
	err := RemoveProject(path, "missing")
	var nf *core.ProjectNotFound
	if !errors.As(err, &nf) {
		t.Fatalf("expected ProjectNotFound, got %v", err)
	}
	if got := readConfig(t, path); got != "projects:\n  a:\n    url: x\n" {
		t.Fatalf("file was modified on not-found remove: %q", got)
	}
}

func TestIsRemoteURL(t *testing.T) {
	urls := []string{
		"https://github.com/alajmo/mani.git",
		"https://github.com/alajmo/mani",
		"http://example.com/x.git",
		"ssh://git@host:2222/org/repo.git",
		"git://example.com/repo.git",
		"git@github.com:alajmo/mani.git",
	}
	for _, u := range urls {
		if !IsRemoteURL(u) {
			t.Errorf("IsRemoteURL(%q) = false, want true", u)
		}
	}
	names := []string{"kaka", "my-repo", "my_lib", "mani", "libs-foo"}
	for _, n := range names {
		if IsRemoteURL(n) {
			t.Errorf("IsRemoteURL(%q) = true, want false (plain name)", n)
		}
	}
}

func TestProjectNameFromURL(t *testing.T) {
	cases := map[string]string{
		"https://github.com/alajmo/mani.git": "mani",
		"https://github.com/alajmo/mani":     "mani",
		"https://github.com/alajmo/mani/":    "mani",
		"git@github.com:alajmo/mani.git":     "mani",
		"ssh://git@host:2222/org/repo.git":   "repo",
		"git://example.com/foo-bar.git":      "foo-bar",
	}
	for in, want := range cases {
		if got := ProjectNameFromURL(in); got != want {
			t.Errorf("ProjectNameFromURL(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestRemoveProject_NoProjectsBlockErrors(t *testing.T) {
	path := writeConfig(t, "tasks:\n  t:\n    cmd: x\n")
	err := RemoveProject(path, "a")
	var nf *core.ProjectNotFound
	if !errors.As(err, &nf) {
		t.Fatalf("expected ProjectNotFound, got %v", err)
	}
}
