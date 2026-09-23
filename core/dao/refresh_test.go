package dao

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/alajmo/mani/core/sources"
	"gopkg.in/yaml.v3"
)

// fixedFetch returns the same projects regardless of source. For tests with
// multiple sources you can switch on the source's discriminator.
func fixedFetch(projects []Project) fetchFunc {
	return func(_ context.Context, _ sources.Source) ([]Project, error) {
		return projects, nil
	}
}

// TestRefresh_ReadsSourcePathFromConfig proves that a hand-written `path:` on a
// source in mani.yaml is decoded and handed to the fetcher. The path prefix is
// applied inside the real FetchAndMap (covered by TestApplySourcePath); here we
// only need to confirm refresh reads the field off disk and propagates it, which
// is the half that depends on the YAML struct tag rather than init.
func TestRefresh_ReadsSourcePathFromConfig(t *testing.T) {
	input := `sources:
  - source: github/my-org
    path: acme
projects:
tasks:
  hello:
    cmd: echo hi
`
	path := writeConfig(t, input)

	var gotPath string
	capturing := func(_ context.Context, src sources.Source) ([]Project, error) {
		gotPath = src.Path
		return nil, nil
	}

	if _, err := refreshMani(path, false, false, capturing, nil); err != nil {
		t.Fatal(err)
	}
	if gotPath != "acme" {
		t.Fatalf("source path = %q, want %q (path: not decoded from mani.yaml)", gotPath, "acme")
	}
}

const inputBasic = `sources:
  - source: github/my-org
projects:
  repo-a:
    url: git@github.com:my-org/repo-a.git
tasks:
  hello:
    cmd: echo hi
`

func writeConfig(t *testing.T, body string) string {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "mani.yaml")
	if err := os.WriteFile(path, []byte(body), 0644); err != nil {
		t.Fatal(err)
	}
	return path
}

func readConfig(t *testing.T, path string) string {
	t.Helper()
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return string(b)
}

func TestRefresh_AppendsNewProjects(t *testing.T) {
	path := writeConfig(t, inputBasic)

	fetch := fixedFetch([]Project{
		{Name: "repo-a", Path: "repo-a", URL: "git@github.com:my-org/repo-a.git"},
		{Name: "repo-b", Path: "repo-b", URL: "git@github.com:my-org/repo-b.git"},
	})

	res, err := refreshMani(path, false, false, fetch, nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(res.Added) != 1 || res.Added[0].Name != "repo-b" {
		t.Fatalf("added = %+v", res.Added)
	}
	if len(res.Removed) != 0 {
		t.Fatalf("removed should be empty without --prune, got %+v", res.Removed)
	}

	out := readConfig(t, path)
	if !strings.Contains(out, "repo-a:") {
		t.Fatalf("existing repo-a missing from output:\n%s", out)
	}
	if !strings.Contains(out, "repo-b:") {
		t.Fatalf("new repo-b missing from output:\n%s", out)
	}
	if !strings.Contains(out, "git@github.com:my-org/repo-b.git") {
		t.Fatalf("repo-b URL missing:\n%s", out)
	}
	// `tasks:` block must be preserved
	if !strings.Contains(out, "hello:") {
		t.Fatalf("tasks block lost:\n%s", out)
	}
}

func TestRefresh_DedupsRepoAcrossSources(t *testing.T) {
	// Two sources that both surface the same upstream repo (repo-x) must add it
	// once, not once-per-source (which the later name dedup would turn into
	// repo-x and repo-x-2 — two entries for one URL).
	input := `sources:
  - source: github/org-a
  - source: github/org-b
projects:
tasks:
  hello:
    cmd: echo hi
`
	path := writeConfig(t, input)

	fetch := func(_ context.Context, src sources.Source) ([]Project, error) {
		switch src.Name() {
		case "org-a":
			return []Project{
				{Name: "repo-x", Path: "repo-x", URL: "git@github.com:org-a/repo-x.git"},
				{Name: "repo-a", Path: "repo-a", URL: "git@github.com:org-a/repo-a.git"},
			}, nil
		case "org-b":
			return []Project{
				{Name: "repo-x", Path: "repo-x", URL: "git@github.com:org-a/repo-x.git"},
				{Name: "repo-b", Path: "repo-b", URL: "git@github.com:org-b/repo-b.git"},
			}, nil
		}
		return nil, nil
	}

	noDiscovery := false
	res, err := refreshMani(path, false, false, fetch, &noDiscovery)
	if err != nil {
		t.Fatal(err)
	}
	if len(res.Added) != 3 {
		t.Fatalf("added = %v, want 3 unique (repo-x once)", names(res.Added))
	}
	out := readConfig(t, path)
	if n := strings.Count(out, "repo-x:"); n != 1 {
		t.Fatalf("repo-x written %d times, want 1:\n%s", n, out)
	}
	// And no spurious suffixed duplicate.
	if strings.Contains(out, "repo-x-2:") {
		t.Fatalf("duplicate repo-x-2 key written:\n%s", out)
	}
}

func TestRefresh_PruneRemovesUpstreamGone(t *testing.T) {
	input := `sources:
  - source: github/my-org
projects:
  repo-a:
    url: git@github.com:my-org/repo-a.git
  repo-gone:
    url: git@github.com:my-org/repo-gone.git
  hand-added:
    url: git@gitlab.com:other/hand-added.git
`
	path := writeConfig(t, input)

	fetch := fixedFetch([]Project{
		{Name: "repo-a", Path: "repo-a", URL: "git@github.com:my-org/repo-a.git"},
	})

	res, err := refreshMani(path, true, false, fetch, nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(res.Removed) != 1 || res.Removed[0].Name != "repo-gone" {
		t.Fatalf("removed = %+v; expected only repo-gone", res.Removed)
	}

	out := readConfig(t, path)
	if strings.Contains(out, "repo-gone") {
		t.Fatalf("repo-gone should be pruned:\n%s", out)
	}
	if !strings.Contains(out, "hand-added") {
		t.Fatalf("hand-added (different remote) must survive --prune:\n%s", out)
	}
}

func TestRefresh_NoChangesIsNoOp(t *testing.T) {
	path := writeConfig(t, inputBasic)
	before := readConfig(t, path)

	fetch := fixedFetch([]Project{
		{Name: "repo-a", Path: "repo-a", URL: "git@github.com:my-org/repo-a.git"},
	})

	res, err := refreshMani(path, false, false, fetch, nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(res.Added) != 0 || len(res.Removed) != 0 {
		t.Fatalf("expected no changes; got added=%v removed=%v", res.Added, res.Removed)
	}
	if got := readConfig(t, path); got != before {
		t.Fatalf("file should be byte-identical on no-op refresh\nbefore:\n%s\nafter:\n%s", before, got)
	}
}

func TestRefresh_DryRunWritesNothing(t *testing.T) {
	path := writeConfig(t, inputBasic)
	before := readConfig(t, path)

	fetch := fixedFetch([]Project{
		{Name: "repo-a", Path: "repo-a", URL: "git@github.com:my-org/repo-a.git"},
		{Name: "repo-new", Path: "repo-new", URL: "git@github.com:my-org/repo-new.git"},
	})

	res, err := refreshMani(path, false, true, fetch, nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(res.Added) != 1 || res.Added[0].Name != "repo-new" {
		t.Fatalf("expected repo-new in added, got %+v", res.Added)
	}
	if got := readConfig(t, path); got != before {
		t.Fatalf("dry-run must not modify file\nbefore:\n%s\nafter:\n%s", before, got)
	}
}

func TestRefresh_PreservesUserComments(t *testing.T) {
	input := `# top-level comment
sources:
  - source: github/my-org
projects:
  # this is a hand-edited project
  repo-a:
    url: git@github.com:my-org/repo-a.git
    desc: my custom description
    tags: [team=core]    # inline comment
`
	path := writeConfig(t, input)

	fetch := fixedFetch([]Project{
		{Name: "repo-a", Path: "repo-a", URL: "git@github.com:my-org/repo-a.git"},
		{Name: "repo-new", Path: "repo-new", URL: "git@github.com:my-org/repo-new.git"},
	})

	if _, err := refreshMani(path, false, false, fetch, nil); err != nil {
		t.Fatal(err)
	}
	out := readConfig(t, path)
	for _, want := range []string{
		"# top-level comment",
		"# this is a hand-edited project",
		"my custom description",
		"team=core",
		"repo-new:",
	} {
		if !strings.Contains(out, want) {
			t.Fatalf("expected output to contain %q\nactual:\n%s", want, out)
		}
	}
}

func TestRefresh_PruneScopedToSource(t *testing.T) {
	// Two sources, two prefixes. Each --prune should only touch its own.
	input := `sources:
  - source: github/org-a
  - source: github/org-b
projects:
  alpha:
    url: git@github.com:org-a/alpha.git
  beta:
    url: git@github.com:org-b/beta.git
  gone-a:
    url: git@github.com:org-a/gone-a.git
  gone-b:
    url: git@github.com:org-b/gone-b.git
`
	path := writeConfig(t, input)

	fetch := func(_ context.Context, src sources.Source) ([]Project, error) {
		switch src.Name() {
		case "org-a":
			return []Project{{Name: "alpha", Path: "alpha", URL: "git@github.com:org-a/alpha.git"}}, nil
		case "org-b":
			return []Project{{Name: "beta", Path: "beta", URL: "git@github.com:org-b/beta.git"}}, nil
		}
		t.Fatalf("unexpected source %q", src.Source)
		return nil, nil
	}

	res, err := refreshMani(path, true, false, fetch, nil)
	if err != nil {
		t.Fatal(err)
	}
	removed := make(map[string]bool, len(res.Removed))
	for _, p := range res.Removed {
		removed[p.Name] = true
	}
	if !removed["gone-a"] || !removed["gone-b"] {
		t.Fatalf("expected gone-a and gone-b in removed; got %+v", res.Removed)
	}
	if removed["alpha"] || removed["beta"] {
		t.Fatalf("alive projects should not be in removed; got %+v", res.Removed)
	}

	out := readConfig(t, path)
	if strings.Contains(out, "gone-a") || strings.Contains(out, "gone-b") {
		t.Fatalf("both stale entries should be pruned:\n%s", out)
	}
	if !strings.Contains(out, "alpha") || !strings.Contains(out, "beta") {
		t.Fatalf("alive projects must survive:\n%s", out)
	}
}

func TestRefresh_RenamesCollidingAddedProject(t *testing.T) {
	// A hand-added project `api` (no url) collides with an upstream repo
	// also named `api`. The upstream repo must not clobber the existing
	// key — it is added under a disambiguated name so the config stays
	// valid YAML with no duplicate keys.
	input := `sources:
  - source: github/my-org
projects:
  api:
`
	path := writeConfig(t, input)

	fetch := fixedFetch([]Project{
		{Name: "api", Path: "api", URL: "git@github.com:my-org/api.git"},
	})

	res, err := refreshMani(path, false, false, fetch, nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(res.Added) != 1 || res.Added[0].Name != "api-2" {
		t.Fatalf("expected added project renamed to api-2, got %+v", res.Added)
	}

	out := readConfig(t, path)
	if !strings.Contains(out, "api-2:") {
		t.Fatalf("expected disambiguated api-2: key:\n%s", out)
	}
	var m map[string]any
	if err := yaml.Unmarshal([]byte(out), &m); err != nil {
		t.Fatalf("refreshed config is not valid YAML (duplicate keys?): %v\n%s", err, out)
	}
}

func TestRefresh_PruneDoesNotMatchOwnerSubstring(t *testing.T) {
	// Source `bar` must not prune a project owned by `foobar` just because
	// "foobar/" contains "bar/".
	input := `sources:
  - source: github/bar
projects:
  mine:
    url: git@github.com:foobar/mine.git
`
	path := writeConfig(t, input)

	res, err := refreshMani(path, true, false, fixedFetch(nil), nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(res.Removed) != 0 {
		t.Fatalf("project owned by foobar must not be pruned by source bar, got %+v", res.Removed)
	}
	if out := readConfig(t, path); !strings.Contains(out, "foobar/mine.git") {
		t.Fatalf("foobar-owned project should survive --prune:\n%s", out)
	}
}

func TestRefresh_PreservesExactFormatting(t *testing.T) {
	// The yamledit write path must leave every untouched byte alone:
	// blank lines, comments, the block scalar, and indentation all survive,
	// and the appended project matches the surrounding indentation.
	input := `sources:
  - source: github/my-org

projects:
  # hand-written
  repo-a:
    url: git@github.com:my-org/repo-a.git
    tags: [core]


tasks:
  hello:
    cmd: |
      echo hi
`
	path := writeConfig(t, input)

	fetch := fixedFetch([]Project{
		{Name: "repo-a", Path: "repo-a", URL: "git@github.com:my-org/repo-a.git"},
		{Name: "repo-b", Path: "repo-b", URL: "git@github.com:my-org/repo-b.git"},
	})
	if _, err := refreshMani(path, false, false, fetch, nil); err != nil {
		t.Fatal(err)
	}

	want := `sources:
  - source: github/my-org

projects:
  # hand-written
  repo-a:
    url: git@github.com:my-org/repo-a.git
    tags: [core]

  repo-b:
    url: git@github.com:my-org/repo-b.git


tasks:
  hello:
    cmd: |
      echo hi
`
	if got := readConfig(t, path); got != want {
		t.Fatalf("formatting not preserved\n--- want ---\n%s\n--- got ---\n%s", want, got)
	}
}

func TestRefresh_ErrorsWhenNoSourcesAndDiscoveryOff(t *testing.T) {
	// No sources block AND --auto-discovery=false leaves refresh with nothing
	// to do, which is an error. (With discovery on — the default — a missing
	// sources block is fine; see TestRefresh_NoSourcesDiscoversLocalRepos.)
	path := writeConfig(t, "projects:\n  a:\n    url: foo\n")
	off := false
	_, err := refreshMani(path, false, false, fixedFetch(nil), &off)
	if err == nil || !strings.Contains(err.Error(), "no `sources:` block") {
		t.Fatalf("expected nothing-to-refresh error, got %v", err)
	}
}

// withDiscoverLocal swaps the package-level discovery hook for the duration of
// a test and restores it afterwards.
func withDiscoverLocal(t *testing.T, fn func(string) ([]Project, error)) {
	t.Helper()
	prev := discoverLocal
	discoverLocal = fn
	t.Cleanup(func() { discoverLocal = prev })
}

func TestRefresh_NoSourcesDiscoversLocalRepos(t *testing.T) {
	// A config from a plain `mani init` (no sources block) must still pick up
	// repos cloned into the workspace by hand.
	input := "projects:\n  repo-a:\n    url: git@github.com:me/repo-a.git\ntasks:\n  hello:\n    cmd: echo hi\n"
	path := writeConfig(t, input)

	withDiscoverLocal(t, func(string) ([]Project, error) {
		return []Project{
			{Name: "repo-a", Path: "repo-a", URL: "git@github.com:me/repo-a.git"}, // already configured
			{Name: "repo-b", Path: "repo-b", URL: "git@github.com:me/repo-b.git"}, // freshly cloned
		}, nil
	})

	res, err := refreshMani(path, false, false, fixedFetch(nil), nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(res.Added) != 1 || res.Added[0].Name != "repo-b" {
		t.Fatalf("added = %+v; expected only repo-b", res.Added)
	}

	out := readConfig(t, path)
	if !strings.Contains(out, "repo-b:") || !strings.Contains(out, "git@github.com:me/repo-b.git") {
		t.Fatalf("discovered repo-b not written:\n%s", out)
	}
	if !strings.Contains(out, "hello:") {
		t.Fatalf("tasks block lost:\n%s", out)
	}
}

func TestRefresh_DiscoveryDoesNotDuplicateSourceRepos(t *testing.T) {
	// A repo returned by a source AND present on disk must be added once, not
	// twice — the on-disk copy is matched by URL against the source addition.
	input := "sources:\n  - source: github/me\nprojects:\ntasks:\n  hello:\n    cmd: echo hi\n"
	path := writeConfig(t, input)

	fetch := fixedFetch([]Project{
		{Name: "repo-a", Path: "repo-a", URL: "git@github.com:me/repo-a.git"},
	})
	withDiscoverLocal(t, func(string) ([]Project, error) {
		return []Project{
			{Name: "repo-a", Path: "repo-a", URL: "git@github.com:me/repo-a.git"}, // same as the source repo
			{Name: "local-only", Path: "local-only", URL: "git@gitlab.com:me/local-only.git"},
		}, nil
	})

	res, err := refreshMani(path, false, false, fetch, nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(res.Added) != 2 {
		t.Fatalf("added = %+v; expected repo-a (source) + local-only (disk)", res.Added)
	}
	out := readConfig(t, path)
	if strings.Count(out, "repo-a:") != 1 {
		t.Fatalf("repo-a should appear exactly once:\n%s", out)
	}
	if !strings.Contains(out, "local-only:") {
		t.Fatalf("local-only repo not added:\n%s", out)
	}
}

func TestRefresh_DiscoveryDisabledSkipsLocalRepos(t *testing.T) {
	input := "sources:\n  - source: github/me\nprojects:\n  repo-a:\n    url: git@github.com:me/repo-a.git\n"
	path := writeConfig(t, input)

	called := false
	withDiscoverLocal(t, func(string) ([]Project, error) {
		called = true
		return []Project{{Name: "repo-b", Path: "repo-b", URL: "git@github.com:me/repo-b.git"}}, nil
	})

	off := false
	res, err := refreshMani(path, false, false, fixedFetch([]Project{
		{Name: "repo-a", Path: "repo-a", URL: "git@github.com:me/repo-a.git"},
	}), &off)
	if err != nil {
		t.Fatal(err)
	}
	if called {
		t.Fatal("discoverLocal must not be called when auto-discovery is off")
	}
	if len(res.Added) != 0 {
		t.Fatalf("expected no additions with discovery off, got %+v", res.Added)
	}
}

func TestRefresh_DiscoveryRenamesNameCollision(t *testing.T) {
	// A discovered repo whose basename collides with an existing project but
	// has a different path/URL must be suffixed, not silently dropped or
	// emitted as a duplicate YAML key.
	input := "projects:\n  app:\n    path: services/app\n    url: git@github.com:me/app.git\n"
	path := writeConfig(t, input)

	withDiscoverLocal(t, func(string) ([]Project, error) {
		return []Project{
			{Name: "app", Path: "vendor/app", URL: "git@github.com:other/app.git"},
		}, nil
	})

	res, err := refreshMani(path, false, false, fixedFetch(nil), nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(res.Added) != 1 || res.Added[0].Name != "app-2" {
		t.Fatalf("added = %+v; expected app-2", res.Added)
	}
	out := readConfig(t, path)
	if !strings.Contains(out, "app-2:") {
		t.Fatalf("collision not disambiguated:\n%s", out)
	}
	var probe yaml.Node
	if err := yaml.Unmarshal([]byte(out), &probe); err != nil {
		t.Fatalf("output is not valid YAML: %v\n%s", err, out)
	}
}

// An empty `projects:` key (a null scalar, as `mani init --source` writes
// when a source returns no repos) must be treated as no projects and then
// populated, not rejected as "not a mapping".
func TestRefresh_PopulatesEmptyProjectsBlock(t *testing.T) {
	input := "sources:\n  - source: github/my-org\nprojects:\ntasks:\n  hello:\n    cmd: echo hi\n"
	path := writeConfig(t, input)

	fetch := fixedFetch([]Project{
		{Name: "repo-a", Path: "repo-a", URL: "git@github.com:my-org/repo-a.git"},
	})

	res, err := refreshMani(path, false, false, fetch, nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(res.Added) != 1 || res.Added[0].Name != "repo-a" {
		t.Fatalf("added = %+v", res.Added)
	}

	out := readConfig(t, path)
	if !strings.Contains(out, "repo-a:") {
		t.Fatalf("repo-a not inserted into empty projects block:\n%s", out)
	}
	if !strings.Contains(out, "hello:") {
		t.Fatalf("tasks block lost:\n%s", out)
	}
	var probe yaml.Node
	if err := yaml.Unmarshal([]byte(out), &probe); err != nil {
		t.Fatalf("output is not valid YAML: %v\n%s", err, out)
	}
}
