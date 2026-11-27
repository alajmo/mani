package dao

import (
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strings"
	"testing"

	"gopkg.in/yaml.v3"

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

// A config that uses lists at the indentation of their key, which is how many
// formatters write them. Adding and removing a project must leave everything
// else as it is.
const indentlessConfig = `import:
- shared.yaml

projects:
  foo:
    url: x
    tags:
    - a
    - b
    desc: hello
  bar:
    url: y

tasks:
  t:
    env:
      A: 1
    commands:
    - name: one
      cmd: echo 1
    - cmd: echo 2
    desc: after
`

func TestAddProject_IndentlessSequences(t *testing.T) {
	path := writeConfig(t, indentlessConfig)

	if err := AddProject(path, Project{Name: "new", Path: "new", URL: "u"}); err != nil {
		t.Fatal(err)
	}

	want := strings.Replace(indentlessConfig, "  bar:\n    url: y\n", "  bar:\n    url: y\n\n  new:\n    url: u\n", 1)
	if got := readConfig(t, path); got != want {
		t.Fatalf("unexpected config\n--- want ---\n%s\n--- got ---\n%s", want, got)
	}
}

func TestRemoveProject_IndentlessSequences(t *testing.T) {
	path := writeConfig(t, indentlessConfig)

	if err := RemoveProject(path, "foo"); err != nil {
		t.Fatal(err)
	}

	want := strings.Replace(indentlessConfig, "  foo:\n    url: x\n    tags:\n    - a\n    - b\n    desc: hello\n", "", 1)
	if got := readConfig(t, path); got != want {
		t.Fatalf("unexpected config\n--- want ---\n%s\n--- got ---\n%s", want, got)
	}
}

func TestAddProject_DocumentStartMarker(t *testing.T) {
	input := "---\nprojects:\n  foo:\n    url: x\n"
	path := writeConfig(t, input)

	if err := AddProject(path, Project{Name: "new", Path: "new", URL: "u"}); err != nil {
		t.Fatal(err)
	}
	if got, want := readConfig(t, path), input+"\n  new:\n    url: u\n"; got != want {
		t.Fatalf("unexpected config\n--- want ---\n%s\n--- got ---\n%s", want, got)
	}
}

func TestAddProject_MultiLineDescription(t *testing.T) {
	path := writeConfig(t, "projects:\n  foo:\n    url: x\n")

	desc := "line one\nline two: with \"quotes\" and a # hash\n"
	if err := AddProject(path, Project{Name: "new", Path: "new", URL: "u", Desc: desc}); err != nil {
		t.Fatal(err)
	}

	var config struct {
		Projects map[string]struct {
			Desc string `yaml:"desc"`
		} `yaml:"projects"`
	}
	if err := yaml.Unmarshal([]byte(readConfig(t, path)), &config); err != nil {
		t.Fatalf("config is not valid YAML: %v\n%s", err, readConfig(t, path))
	}
	if got := config.Projects["new"].Desc; got != desc {
		t.Fatalf("desc = %q; want %q", got, desc)
	}
}

func TestYamlQuote_SingleLine(t *testing.T) {
	for _, s := range []string{
		"plain",
		"a: b",
		"- dash",
		"line one\nline two",
		"trailing newline\n",
		strings.Repeat("a long description ", 20) + "\nwith a second line",
		"tab\there and \"quotes\" and 'single' and \\ backslash",
		"unicode ü → ✓\nsecond",
	} {
		quoted := yamlQuote(s)
		if strings.Contains(quoted, "\n") {
			t.Errorf("%q is quoted over several lines: %q", s, quoted)
		}

		var got struct {
			Value string `yaml:"value"`
		}
		if err := yaml.Unmarshal([]byte("value: "+quoted+"\n"), &got); err != nil {
			t.Errorf("%q: quoted as %s, which is not valid YAML: %v", s, quoted, err)
		} else if got.Value != s {
			t.Errorf("%q: quoted as %s, which reads back as %q", s, quoted, got.Value)
		}
	}
}

func TestAddProject_SymlinkedConfig(t *testing.T) {
	target := writeConfig(t, "projects:\n  foo:\n    url: x\n")
	link := filepath.Join(t.TempDir(), "mani.yaml")
	if err := os.Symlink(target, link); err != nil {
		t.Skipf("symlinks are not available: %v", err)
	}

	if err := AddProject(link, Project{Name: "new", Path: "new", URL: "u"}); err != nil {
		t.Fatal(err)
	}

	if info, err := os.Lstat(link); err != nil || info.Mode()&os.ModeSymlink == 0 {
		t.Fatalf("the config is no longer a symlink (%v, %v)", info.Mode(), err)
	}
	if got := readConfig(t, target); !strings.Contains(got, "  new:\n    url: u\n") {
		t.Fatalf("the file the symlink points to was not updated:\n%s", got)
	}
}

// A layout the editor can't follow is refused, and the config left as it is.
func TestAddProject_UnsupportedProjectsLayout(t *testing.T) {
	for _, input := range []string{
		"projects: {foo: {url: x}}\n",
		"projects:\n  foo:\n    url: x\n  ? [a, b]\n  : url: y\n",
	} {
		path := writeConfig(t, input)

		err := AddProject(path, Project{Name: "new", Path: "new", URL: "u"})
		if err == nil || !strings.Contains(err.Error(), "not supported") {
			t.Errorf("%q: expected an unsupported layout error, got %v", input, err)
		}
		if got := readConfig(t, path); got != input {
			t.Errorf("%q: the config was modified:\n%s", input, got)
		}
	}
}

func TestVerifyProjectEdit(t *testing.T) {
	before := "shell: bash -c\nprojects:\n  foo:\n    url: x\n    tags:\n    - a\n  bar:\n    url: y\ntasks:\n  t:\n    commands:\n    - cmd: echo 1\n"

	tests := []struct {
		name    string
		after   string
		added   []string
		removed []string
		wantErr string
	}{
		{name: "nothing changed", after: before},
		{name: "only formatting changed", after: strings.ReplaceAll(before, "    - a", "      - a")},
		{name: "project added", after: strings.Replace(before, "tasks:", "  new:\n    url: u\ntasks:", 1), added: []string{"new"}},
		{name: "project removed", after: strings.Replace(before, "  bar:\n    url: y\n", "", 1), removed: []string{"bar"}},
		{
			name:    "list items of another project lost",
			after:   strings.Replace(strings.Replace(before, "    - a\n", "", 1), "tasks:", "  new:\n    url: u\ntasks:", 1),
			added:   []string{"new"},
			wantErr: "project `foo`",
		},
		{
			name:    "task commands lost",
			after:   strings.Replace(strings.Replace(before, "    - cmd: echo 1\n", "", 1), "tasks:", "  new:\n    url: u\ntasks:", 1),
			added:   []string{"new"},
			wantErr: "`tasks`",
		},
		{name: "root key lost", after: strings.Replace(before, "shell: bash -c\n", "", 1), wantErr: "`shell`"},
		{name: "root key added", after: before + "extra: 1\n", wantErr: "`extra`"},
		{name: "unexpected project added", after: strings.Replace(before, "tasks:", "  new:\n    url: u\ntasks:", 1), wantErr: "project `new`"},
		{name: "unexpected project removed", after: strings.Replace(before, "  bar:\n    url: y\n", "", 1), wantErr: "project `bar`"},
		{name: "added project missing", after: before, added: []string{"new"}, wantErr: "project `new`"},
		{name: "removed project still there", after: before, removed: []string{"bar"}, wantErr: "project `bar`"},
		{name: "duplicate key", after: strings.Replace(before, "tasks:", "  foo:\n    url: u\ntasks:", 1), added: []string{"foo"}, wantErr: "invalid YAML"},
		{name: "invalid yaml", after: before + "  : :\n - x", wantErr: "invalid YAML"},
	}

	for _, tt := range tests {
		err := verifyProjectEdit([]byte(before), []byte(tt.after), tt.added, tt.removed)
		switch {
		case tt.wantErr == "" && err != nil:
			t.Errorf("%s: unexpected error: %v", tt.name, err)
		case tt.wantErr != "" && (err == nil || !strings.Contains(err.Error(), tt.wantErr)):
			t.Errorf("%s: got error %v, want one containing %q", tt.name, err, tt.wantErr)
		}
	}
}

func TestAddRemoveProject_NameThatNeedsQuoting(t *testing.T) {
	input := "projects:\n  a:\n    url: x\n"
	path := writeConfig(t, input)

	for _, name := range []string{"null", "we: ird"} {
		if err := AddProject(path, Project{Name: name, Path: name, URL: "u", Tags: []string{"a]b", "c"}}); err != nil {
			t.Fatalf("add %q: %v", name, err)
		}
		if err := AddProject(path, Project{Name: name, Path: name, URL: "u"}); !errors.As(err, new(*core.ProjectAlreadyExists)) {
			t.Fatalf("add %q again: got %v, want ProjectAlreadyExists", name, err)
		}
	}
	for _, name := range []string{"null", "we: ird"} {
		if err := RemoveProject(path, name); err != nil {
			t.Fatalf("remove %q: %v", name, err)
		}
	}

	if got := readConfig(t, path); got != input {
		t.Fatalf("adding and removing should leave the config as it was:\n%s", got)
	}
}

func TestFindVCSystems_SkipsUnreadableDirectories(t *testing.T) {
	if os.Geteuid() == 0 {
		t.Skip("root can read any directory")
	}

	root := t.TempDir()
	for _, dir := range []string{"repo/.git", "data/pg/inner", "data/other/.git"} {
		if err := os.MkdirAll(filepath.Join(root, dir), 0755); err != nil {
			t.Fatal(err)
		}
	}
	locked := filepath.Join(root, "data", "pg")
	if err := os.Chmod(locked, 0); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chmod(locked, 0755) })

	projects, err := FindVCSystems(root)
	if err != nil {
		t.Fatalf("an unreadable directory should be skipped, got: %v", err)
	}

	var paths []string
	for _, p := range projects {
		paths = append(paths, filepath.ToSlash(p.Path))
	}
	sort.Strings(paths)
	if want := []string{"data/other", "repo"}; !reflect.DeepEqual(paths, want) {
		t.Fatalf("found %v; want %v", paths, want)
	}
}
