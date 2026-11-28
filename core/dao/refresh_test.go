package dao

import (
	"context"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strings"
	"testing"

	"github.com/alajmo/mani/core/sources"
	"gopkg.in/yaml.v3"
)

// fixedFetch returns the same projects regardless of source. For tests with
// multiple sources you can switch on the source's discriminator.
func fixedFetch(projects []Project) fetchFunc {
	return func(_ context.Context, _ sources.Source) (sourceRepos, error) {
		return sourceRepos{Projects: projects}, nil
	}
}

// projectsFetch turns a fake that returns projects into a fetchFunc, for a
// source that lists every repo that exists.
func projectsFetch(fetch func(context.Context, sources.Source) ([]Project, error)) fetchFunc {
	return func(ctx context.Context, src sources.Source) (sourceRepos, error) {
		projects, err := fetch(ctx, src)
		return sourceRepos{Projects: projects}, err
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
	capturing := projectsFetch(func(_ context.Context, src sources.Source) ([]Project, error) {
		gotPath = src.Path
		return nil, nil
	})

	if _, err := refreshMani(path, false, false, capturing, nil, nil); err != nil {
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

	res, err := refreshMani(path, false, false, fetch, nil, nil)
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

	fetch := projectsFetch(func(_ context.Context, src sources.Source) ([]Project, error) {
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
	})

	noDiscovery := false
	res, err := refreshMani(path, false, false, fetch, &noDiscovery, nil)
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

	res, err := refreshMani(path, true, false, fetch, nil, nil)
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

	res, err := refreshMani(path, false, false, fetch, nil, nil)
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

	res, err := refreshMani(path, false, true, fetch, nil, nil)
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

	if _, err := refreshMani(path, false, false, fetch, nil, nil); err != nil {
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

	fetch := projectsFetch(func(_ context.Context, src sources.Source) ([]Project, error) {
		switch src.Name() {
		case "org-a":
			return []Project{{Name: "alpha", Path: "alpha", URL: "git@github.com:org-a/alpha.git"}}, nil
		case "org-b":
			return []Project{{Name: "beta", Path: "beta", URL: "git@github.com:org-b/beta.git"}}, nil
		}
		t.Fatalf("unexpected source %q", src.Source)
		return nil, nil
	})

	res, err := refreshMani(path, true, false, fetch, nil, nil)
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
	// A hand-added project `api` (no url, in another directory) collides
	// with an upstream repo also named `api`. The upstream repo must not
	// clobber the existing key — it is added under a disambiguated name so
	// the config stays valid YAML with no duplicate keys.
	input := `sources:
  - source: github/my-org
projects:
  api:
    path: vendor/api
`
	path := writeConfig(t, input)

	fetch := fixedFetch([]Project{
		{Name: "api", Path: "api", URL: "git@github.com:my-org/api.git"},
	})

	res, err := refreshMani(path, false, false, fetch, nil, nil)
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

	res, err := refreshMani(path, true, false, fixedFetch(nil), nil, nil)
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
	if _, err := refreshMani(path, false, false, fetch, nil, nil); err != nil {
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
	_, err := refreshMani(path, false, false, fixedFetch(nil), &off, nil)
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

	res, err := refreshMani(path, false, false, fixedFetch(nil), nil, nil)
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

	res, err := refreshMani(path, false, false, fetch, nil, nil)
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
	}), &off, nil)
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

	res, err := refreshMani(path, false, false, fixedFetch(nil), nil, nil)
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

	res, err := refreshMani(path, false, false, fetch, nil, nil)
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

func TestRefresh_DiscoveryWithProjectDir(t *testing.T) {
	// Project paths are relative to project_dir, while discovery reports paths
	// relative to the config. A configured project must be recognized, and a
	// new one written with a path relative to project_dir.
	input := "project_dir: repos\nprojects:\n  app:\n    path: local\n  lib:\n"
	path := writeConfig(t, input)

	withDiscoverLocal(t, func(string) ([]Project, error) {
		return []Project{
			{Name: "local", Path: filepath.Join("repos", "local")}, // project app
			{Name: "lib", Path: filepath.Join("repos", "lib")},     // project lib, default path
			{Name: "fresh", Path: filepath.Join("repos", "fresh")}, // new, in project_dir
			{Name: "tool", Path: filepath.Join("other", "tool")},   // new, outside project_dir
		}, nil
	})

	res, err := refreshMani(path, false, false, fixedFetch(nil), nil, nil)
	if err != nil {
		t.Fatal(err)
	}

	got := map[string]string{}
	for _, p := range res.Added {
		got[p.Name] = p.Path
	}
	want := map[string]string{"fresh": "fresh", "tool": "../other/tool"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("added = %v; expected %v", got, want)
	}

	out := readConfig(t, path)
	if strings.Contains(out, "local:") || strings.Count(out, "lib:") != 1 {
		t.Fatalf("configured projects were added again:\n%s", out)
	}
	if !strings.Contains(out, "  fresh:\n") || strings.Contains(out, "path: fresh") || strings.Contains(out, "repos/fresh") {
		t.Fatalf("fresh should be written without a path, it defaults to its name in project_dir:\n%s", out)
	}
	if !strings.Contains(out, "path: ../other/tool") {
		t.Fatalf("tool should be written with a path relative to project_dir:\n%s", out)
	}
}

func TestRefresh_DiscoverySkipsProjectsOfImportedConfigs(t *testing.T) {
	// The projects of the loaded config include those of imported configs.
	// They are projects already, and must not be added to the refreshed config.
	input := "import:\n  - team/mani.yaml\nprojects:\n  repo-a:\n    url: git@github.com:me/repo-a.git\n"
	path := writeConfig(t, input)
	dir := filepath.Dir(path)

	configured := []Project{
		{Name: "repo-a", Path: filepath.Join(dir, "repo-a"), URL: "git@github.com:me/repo-a.git"},
		{Name: "imported", Path: filepath.Join(dir, "team", "imported")},
		{Name: "tool", Path: filepath.Join(dir, "team", "tool"), URL: "git@github.com:team/tool.git"},
	}

	withDiscoverLocal(t, func(string) ([]Project, error) {
		return []Project{
			{Name: "repo-a", Path: "repo-a", URL: "git@github.com:me/repo-a.git"},
			{Name: "imported", Path: filepath.Join("team", "imported")},                          // known by directory
			{Name: "moved", Path: "moved", URL: "git@github.com:team/tool.git"},                  // known by url
			{Name: "tool", Path: filepath.Join("elsewhere", "tool"), URL: "git@host:other/tool"}, // new, name is taken
		}, nil
	})

	res, err := refreshMani(path, false, false, fixedFetch(nil), nil, configured)
	if err != nil {
		t.Fatal(err)
	}
	if len(res.Added) != 1 || res.Added[0].Name != "tool-2" || res.Added[0].Path != "elsewhere/tool" {
		t.Fatalf("added = %+v; expected only elsewhere/tool, named tool-2", res.Added)
	}

	out := readConfig(t, path)
	if strings.Contains(out, "imported:") || strings.Contains(out, "moved:") {
		t.Fatalf("projects of an imported config were added:\n%s", out)
	}
	// A project with the same name would be combined with the imported one
	if !strings.Contains(out, "  tool-2:\n") || strings.Contains(out, "  tool:\n") {
		t.Fatalf("the new repo should get a name that isn't taken by an imported project:\n%s", out)
	}
}

func TestRefresh_SourceSkipsProjectsOfImportedConfigs(t *testing.T) {
	// A repo of a source that is a project in an imported config isn't added,
	// and since it isn't in the refreshed config it's never stale or pruned.
	input := "sources:\n  - source: github/me\nimport:\n  - team/mani.yaml\nprojects:\n"
	path := writeConfig(t, input)
	dir := filepath.Dir(path)

	configured := []Project{
		{Name: "repo-a", Path: filepath.Join(dir, "team", "repo-a"), URL: "git@github.com:me/repo-a.git"},
		{Name: "gone", Path: filepath.Join(dir, "team", "gone"), URL: "git@github.com:me/gone.git"},
	}
	fetch := fixedFetch([]Project{
		{Name: "repo-a", Path: "repo-a", URL: "git@github.com:me/repo-a.git"},
		{Name: "repo-b", Path: "repo-b", URL: "git@github.com:me/repo-b.git"},
	})
	withDiscoverLocal(t, func(string) ([]Project, error) { return nil, nil })

	res, err := refreshMani(path, true, false, fetch, nil, configured)
	if err != nil {
		t.Fatal(err)
	}
	if len(res.Added) != 1 || res.Added[0].Name != "repo-b" {
		t.Fatalf("added = %+v; expected only repo-b", res.Added)
	}
	if len(res.Removed) != 0 || len(res.Stale) != 0 {
		t.Fatalf("removed = %+v, stale = %+v; projects of imported configs must be left alone", res.Removed, res.Stale)
	}

	out := readConfig(t, path)
	if strings.Contains(out, "repo-a:") || !strings.Contains(out, "repo-b:") {
		t.Fatalf("expected repo-b only:\n%s", out)
	}
}

func TestRefresh_CreatesProjectsBlock(t *testing.T) {
	// A config without a `projects:` key (only tasks, or projects that are all
	// defined in imported configs) gets one when there is something to add.
	input := "tasks:\n  hello:\n    cmd: echo hi\n"
	path := writeConfig(t, input)

	withDiscoverLocal(t, func(string) ([]Project, error) {
		return []Project{
			{Name: "repo-a", Path: "repo-a", URL: "git@github.com:me/repo-a.git"},
			{Name: "repo-b", Path: filepath.Join("libs", "repo-b")},
		}, nil
	})

	res, err := refreshMani(path, false, false, fixedFetch(nil), nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(res.Added) != 2 {
		t.Fatalf("added = %+v; expected both repos", res.Added)
	}

	want := input + "projects:\n  repo-a:\n    url: git@github.com:me/repo-a.git\n\n  repo-b:\n    path: libs/repo-b\n"
	if got := readConfig(t, path); got != want {
		t.Fatalf("unexpected config\n--- want ---\n%s\n--- got ---\n%s", want, got)
	}
}

func TestUrlKey(t *testing.T) {
	same := [][]string{
		{
			"git@github.com:Acme/API.git",
			"https://github.com/acme/api.git",
			"https://github.com/acme/api",
			"https://github.com/acme/api/",
			"ssh://git@github.com/acme/api.git",
			"ssh://git@github.com:22/acme/api.git",
			"git://github.com/acme/api.git",
			"https://user:token@github.com/acme/api.git",
			"github.com:acme/api",
		},
		{"git@gitlab.com:group/sub/repo.git", "https://gitlab.com/group/sub/repo"},
	}
	for _, urls := range same {
		for _, u := range urls[1:] {
			if urlKey(u) != urlKey(urls[0]) {
				t.Errorf("%q and %q are the same repository, got keys %q and %q", urls[0], u, urlKey(urls[0]), urlKey(u))
			}
		}
	}

	different := []string{
		"git@github.com:acme/api.git",
		"git@gitlab.com:acme/api.git",
		"git@github.com:acme/api-2.git",
		"git@github.com:other/api.git",
		"../local/api",
		"/srv/git/api.git",
		"file:///srv/git/api.git",
		`C:\repos\api`,
	}
	seen := map[string]string{}
	for _, u := range different {
		if other, ok := seen[urlKey(u)]; ok {
			t.Errorf("%q and %q are different repositories, both got key %q", other, u, urlKey(u))
		}
		seen[urlKey(u)] = u
	}

	// Anything that isn't a remote URL is left as it is
	for _, u := range []string{"../local/api", "/srv/git/Api.git", `C:\repos\api`} {
		if urlKey(u) != u {
			t.Errorf("urlKey(%q) = %q; expected it unchanged", u, urlKey(u))
		}
	}
}

func TestRefresh_MatchesUrlsInAnyForm(t *testing.T) {
	// Sources report ssh URLs. A project configured with the https URL of the
	// same repository is the same project: not added again, and not stale.
	input := `sources:
  - source: github/acme
projects:
  api:
    url: https://github.com/acme/api.git
  web:
    url: ssh://git@github.com/Acme/web
`
	path := writeConfig(t, input)

	fetch := fixedFetch([]Project{
		{Name: "api", Path: "api", URL: "git@github.com:acme/api.git"},
		{Name: "web", Path: "web", URL: "git@github.com:acme/web.git"},
	})
	withDiscoverLocal(t, func(string) ([]Project, error) {
		return []Project{{Name: "api", Path: "elsewhere/api", URL: "git@github.com:acme/api.git"}}, nil
	})

	res, err := refreshMani(path, true, false, fetch, nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(res.Added) != 0 || len(res.Removed) != 0 || len(res.Stale) != 0 {
		t.Fatalf("added = %+v, removed = %+v, stale = %+v; expected no changes", res.Added, res.Removed, res.Stale)
	}
	if got := readConfig(t, path); got != input {
		t.Fatalf("config was modified:\n%s", got)
	}
}

func TestRefresh_PruneIsScopedToHostAndOwner(t *testing.T) {
	// Only projects of the source's own host and owner can be stale. The owner
	// name appearing on another host, or elsewhere in the path, isn't enough.
	input := `sources:
  - source: github/joe
projects:
  kept:
    url: git@github.com:joe/kept.git
  gone:
    url: https://github.com/joe/gone.git
  other-host:
    url: git@gitlab.com:joe/tool.git
  other-host-https:
    url: https://codeberg.org/joe/tool.git
  owner-in-path:
    url: git@github.com:acme/joe/tool.git
  other-owner:
    url: git@github.com:joey/tool.git
  local:
    path: local
`
	path := writeConfig(t, input)

	fetch := fixedFetch([]Project{
		{Name: "kept", Path: "kept", URL: "git@github.com:joe/kept.git"},
	})
	withDiscoverLocal(t, func(string) ([]Project, error) { return nil, nil })

	res, err := refreshMani(path, true, false, fetch, nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(res.Removed) != 1 || res.Removed[0].Name != "gone" {
		t.Fatalf("removed = %+v; expected only gone", res.Removed)
	}

	want := strings.Replace(input, "  gone:\n    url: https://github.com/joe/gone.git\n", "", 1)
	if got := readConfig(t, path); got != want {
		t.Fatalf("unexpected config\n--- want ---\n%s\n--- got ---\n%s", want, got)
	}
}

// filteredFetch returns all for a source that includes archived repos and
// forks and has no limit, and filtered otherwise, as the providers do.
func filteredFetch(all, filtered []Project, calls *int) fetchFunc {
	return projectsFetch(func(_ context.Context, src sources.Source) ([]Project, error) {
		*calls++
		if src.Archived && src.Fork && src.Limit == 0 {
			return all, nil
		}
		return filtered, nil
	})
}

func TestRefresh_FilteredReposAreNotStale(t *testing.T) {
	// A repo that is archived, a fork, or beyond the limit is left out of the
	// listing of the source, but it still exists upstream. It is not stale,
	// and must not be pruned.
	input := `sources:
  - source: github/acme
    limit: 2
projects:
  api:
    url: git@github.com:acme/api.git
  archived:
    url: git@github.com:acme/archived.git
  beyond-limit:
    url: git@github.com:acme/beyond-limit.git
  deleted:
    url: git@github.com:acme/deleted.git
`
	path := writeConfig(t, input)

	filtered := []Project{
		{Name: "api", Path: "api", URL: "git@github.com:acme/api.git"},
		{Name: "web", Path: "web", URL: "git@github.com:acme/web.git"},
	}
	all := append([]Project{
		{Name: "archived", Path: "archived", URL: "git@github.com:acme/archived.git"},
		{Name: "beyond-limit", Path: "beyond-limit", URL: "git@github.com:acme/beyond-limit.git"},
	}, filtered...)

	calls := 0
	withDiscoverLocal(t, func(string) ([]Project, error) { return nil, nil })
	res, err := refreshMani(path, true, false, filteredFetch(all, filtered, &calls), nil, nil)
	if err != nil {
		t.Fatal(err)
	}

	if len(res.Removed) != 1 || res.Removed[0].Name != "deleted" {
		t.Fatalf("removed = %+v; expected only deleted", res.Removed)
	}
	if len(res.Added) != 1 || res.Added[0].Name != "web" {
		t.Fatalf("added = %+v; expected only web, the filters still apply to what is added", res.Added)
	}
	if calls != 2 {
		t.Fatalf("fetched %d times; expected the listing and one complete listing to check the stale projects", calls)
	}

	out := readConfig(t, path)
	if !strings.Contains(out, "archived:") || !strings.Contains(out, "beyond-limit:") || strings.Contains(out, "deleted:") {
		t.Fatalf("expected only deleted to be pruned:\n%s", out)
	}
}

func TestRefresh_CompleteListingIsFetchedOnce(t *testing.T) {
	// Nothing is left out of the listing of a source that includes archived
	// repos and forks and has no limit, and nothing needs to be checked when
	// no project is stale.
	for name, input := range map[string]string{
		"no filters": "sources:\n  - source: github/acme\n    archived: true\n    fork: true\nprojects:\n  gone:\n    url: git@github.com:acme/gone.git\n",
		"none stale": "sources:\n  - source: github/acme\nprojects:\n  api:\n    url: git@github.com:acme/api.git\n",
	} {
		path := writeConfig(t, input)
		listing := []Project{{Name: "api", Path: "api", URL: "git@github.com:acme/api.git"}}

		calls := 0
		withDiscoverLocal(t, func(string) ([]Project, error) { return nil, nil })
		if _, err := refreshMani(path, false, true, filteredFetch(listing, listing, &calls), nil, nil); err != nil {
			t.Fatal(err)
		}
		if calls != 1 {
			t.Errorf("%s: fetched %d times; expected 1", name, calls)
		}
	}
}

func TestRefresh_EmptyListingMarksNothingStale(t *testing.T) {
	// An empty listing is more likely a problem with the provider (not logged
	// in, a renamed organization) than every repository being deleted.
	input := "sources:\n  - source: github/acme\nprojects:\n  api:\n    url: git@github.com:acme/api.git\n"
	path := writeConfig(t, input)

	withDiscoverLocal(t, func(string) ([]Project, error) { return nil, nil })
	res, err := refreshMani(path, true, false, fixedFetch(nil), nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(res.Removed) != 0 || len(res.Stale) != 0 {
		t.Fatalf("removed = %+v, stale = %+v; expected none", res.Removed, res.Stale)
	}
	if got := readConfig(t, path); got != input {
		t.Fatalf("config was modified:\n%s", got)
	}
}

func TestRefresh_SourceRepoInTheDirectoryOfAProject(t *testing.T) {
	// A repo of a source that would be cloned into the directory of a project
	// is that project, also when the URLs can't be matched: an ssh host alias,
	// or a project without a URL. Adding it under another name would give two
	// projects in one directory.
	input := `sources:
  - source: github/acme
projects:
  alpha:
    url: git@gh-work:acme/alpha.git
  beta:
  renamed:
    path: gamma
`
	path := writeConfig(t, input)

	fetch := fixedFetch([]Project{
		{Name: "alpha", Path: "alpha", URL: "git@github.com:acme/alpha.git"},
		{Name: "beta", Path: "beta", URL: "git@github.com:acme/beta.git"},
		{Name: "gamma", Path: "gamma", URL: "git@github.com:acme/gamma.git"},
		{Name: "delta", Path: "delta", URL: "git@github.com:acme/delta.git"},
	})
	withDiscoverLocal(t, func(string) ([]Project, error) { return nil, nil })

	res, err := refreshMani(path, true, false, fetch, nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(res.Added) != 1 || res.Added[0].Name != "delta" {
		t.Fatalf("added = %+v; expected only delta", res.Added)
	}
	if len(res.Removed) != 0 || len(res.Stale) != 0 {
		t.Fatalf("removed = %+v, stale = %+v; expected none", res.Removed, res.Stale)
	}
}

func TestRefresh_PruneFreesNameAndDirectory(t *testing.T) {
	// A project that is gone upstream is pruned, and a new repo of the source
	// with the same name takes its name and directory in the same run.
	input := `sources:
  - source: github/acme
    archived: true
    fork: true
projects:
  alpha:
    url: git@github.com:acme/alpha-old.git
  beta:
    url: git@github.com:acme/beta.git
`
	path := writeConfig(t, input)

	fetch := fixedFetch([]Project{
		{Name: "alpha", Path: "alpha", URL: "git@github.com:acme/alpha.git"},
		{Name: "beta", Path: "beta", URL: "git@github.com:acme/beta.git"},
	})
	withDiscoverLocal(t, func(string) ([]Project, error) { return nil, nil })

	res, err := refreshMani(path, true, false, fetch, nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(res.Removed) != 1 || res.Removed[0].URL != "git@github.com:acme/alpha-old.git" {
		t.Fatalf("removed = %+v; expected the old alpha", res.Removed)
	}
	if len(res.Added) != 1 || res.Added[0].Name != "alpha" {
		t.Fatalf("added = %+v; expected the new alpha, under its own name", res.Added)
	}

	out := readConfig(t, path)
	if strings.Contains(out, "alpha-old") || strings.Count(out, "  alpha:\n") != 1 || !strings.Contains(out, "git@github.com:acme/alpha.git") {
		t.Fatalf("expected alpha to be replaced:\n%s", out)
	}

	// Without --prune the old project stays, and keeps its directory
	path = writeConfig(t, input)
	res, err = refreshMani(path, false, false, fetch, nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(res.Added) != 0 || len(res.Stale) != 1 {
		t.Fatalf("added = %+v, stale = %+v; expected alpha to be stale and nothing added", res.Added, res.Stale)
	}
}

func TestRefresh_PrunedCloneIsNotDiscoveredAgain(t *testing.T) {
	// The clone of a pruned project stays on disk. Discovery must not add it
	// back: the next refresh would prune it again, and so on.
	input := `sources:
  - source: github/acme
projects:
  api:
    url: git@github.com:acme/api.git
  old:
    url: git@github.com:acme/old.git
`
	path := writeConfig(t, input)

	listing := []Project{{Name: "api", Path: "api", URL: "git@github.com:acme/api.git"}}
	complete := append([]Project{{Name: "archived", Path: "archived", URL: "git@github.com:acme/archived.git"}}, listing...)
	calls := 0
	fetch := filteredFetch(complete, listing, &calls)
	withDiscoverLocal(t, func(string) ([]Project, error) {
		return []Project{
			{Name: "api", Path: "api", URL: "git@github.com:acme/api.git"},
			{Name: "old", Path: "old", URL: "https://github.com/acme/old"},                // gone upstream
			{Name: "archived", Path: "archived", URL: "git@github.com:acme/archived.git"}, // left out by the source, but exists
			{Name: "mine", Path: "mine", URL: "git@github.com:me/mine.git"},               // not of the source
			{Name: "scratch", Path: "scratch"},                                            // no remote
		}, nil
	})

	added := func(res *RefreshResult) []string {
		var names []string
		for _, p := range res.Added {
			names = append(names, p.Name)
		}
		sort.Strings(names)
		return names
	}

	res, err := refreshMani(path, true, false, fetch, nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(res.Removed) != 1 || res.Removed[0].Name != "old" {
		t.Fatalf("removed = %+v; expected old", res.Removed)
	}
	if got, want := added(res), []string{"archived", "mine", "scratch"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("added = %v; expected %v", got, want)
	}
	if calls != 2 {
		t.Fatalf("fetched %d times; expected the complete listing to be fetched once", calls)
	}

	// The second run has nothing to do
	res, err = refreshMani(path, true, false, fetch, nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(res.Added) != 0 || len(res.Removed) != 0 || len(res.Stale) != 0 {
		t.Fatalf("second run: added = %+v, removed = %+v, stale = %+v; expected no changes", res.Added, res.Removed, res.Stale)
	}
}

func TestRefresh_ProjectNamesAndTagsThatNeedQuoting(t *testing.T) {
	// The name of a project comes from a directory or a repo, and may be
	// anything. It has to be written so that it's read back as that name.
	path := writeConfig(t, "projects:\n  a:\n    url: x\n")

	names := []string{"null", "we: ird", "@scope", "[archive]", "!old", "123", "2fa-tool", "plain"}
	withDiscoverLocal(t, func(string) ([]Project, error) {
		var found []Project
		for _, name := range names {
			found = append(found, Project{Name: name, Path: "dir-" + name, Tags: []string{"a]b", "c, d", "{e}", "plain"}})
		}
		return found, nil
	})

	if _, err := refreshMani(path, false, false, fixedFetch(nil), nil, nil); err != nil {
		t.Fatal(err)
	}

	var config struct {
		Projects map[string]struct {
			Path string   `yaml:"path"`
			Tags []string `yaml:"tags"`
		} `yaml:"projects"`
	}
	out := readConfig(t, path)
	if err := yaml.Unmarshal([]byte(out), &config); err != nil {
		t.Fatalf("config is not valid YAML: %v\n%s", err, out)
	}
	for _, name := range names {
		project, ok := config.Projects[name]
		if !ok || project.Path != "dir-"+name {
			t.Errorf("project %q was not read back:\n%s", name, out)
			continue
		}
		if want := []string{"a]b", "c, d", "{e}", "plain"}; !reflect.DeepEqual(project.Tags, want) {
			t.Errorf("project %q: tags = %q; want %q", name, project.Tags, want)
		}
	}
	if !strings.Contains(out, "  plain:\n") || !strings.Contains(out, "  2fa-tool:\n") {
		t.Errorf("names that don't need quotes should be written without them:\n%s", out)
	}
}

func TestRefresh_PartialListingMarksNothingStale(t *testing.T) {
	// The source of a user that isn't the one logged in to the provider's CLI
	// only lists the public repos of that user. A private repo is missing from
	// it, and exists: it is not stale, and its clone is still discovered.
	input := `sources:
  - source: github/me-personal
    archived: true
    fork: true
projects:
  public:
    url: git@github.com:me-personal/public.git
  secret:
    url: git@github.com:me-personal/secret.git
    tags: [hand-written]
`
	fetch := func(partial bool) fetchFunc {
		return func(context.Context, sources.Source) (sourceRepos, error) {
			return sourceRepos{
				Projects: []Project{{Name: "public", Path: "public", URL: "git@github.com:me-personal/public.git"}},
				Partial:  partial,
			}, nil
		}
	}
	withDiscoverLocal(t, func(string) ([]Project, error) {
		return []Project{{Name: "other-secret", Path: "other-secret", URL: "git@github.com:me-personal/other-secret.git"}}, nil
	})

	path := writeConfig(t, input)
	res, err := refreshMani(path, true, false, fetch(true), nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(res.Removed) != 0 || len(res.Stale) != 0 {
		t.Fatalf("removed = %+v, stale = %+v; a partial listing can't tell that a repo is gone", res.Removed, res.Stale)
	}
	if len(res.Added) != 1 || res.Added[0].Name != "other-secret" {
		t.Fatalf("added = %+v; expected the clone of the private repo to be discovered", res.Added)
	}
	if out := readConfig(t, path); !strings.Contains(out, "secret:\n    url: git@github.com:me-personal/secret.git\n    tags: [hand-written]") {
		t.Fatalf("the private project was changed:\n%s", out)
	}

	// The same listing from a source that lists every repo: secret is gone
	path = writeConfig(t, input)
	res, err = refreshMani(path, true, false, fetch(false), nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(res.Removed) != 1 || res.Removed[0].Name != "secret" || len(res.Added) != 0 {
		t.Fatalf("removed = %+v, added = %+v; expected secret to be pruned and its sibling's clone left alone", res.Removed, res.Added)
	}
}

func TestRefresh_SameRepoNameInTwoSources(t *testing.T) {
	// Two repos can't be cloned into one directory. The second repo with a
	// name gets another name, and a directory of that name.
	input := "sources:\n  - source: github/org-a\n  - source: github/org-b\n  - source: github/org-c\n    path: c\nprojects:\n"
	path := writeConfig(t, input)

	fetch := projectsFetch(func(_ context.Context, src sources.Source) ([]Project, error) {
		p := Project{Name: "docs", Path: "docs", URL: "git@github.com:" + src.Name() + "/docs.git"}
		if src.Path != "" {
			p.Path = src.Path + "/docs" // as applySourcePath does
		}
		return []Project{p}, nil
	})
	// A clone on disk that takes the name the second repo would get
	withDiscoverLocal(t, func(string) ([]Project, error) {
		return []Project{{Name: "docs", Path: "vendor/docs", URL: "git@gitlab.com:x/docs.git"}}, nil
	})

	res, err := refreshMani(path, false, false, fetch, nil, nil)
	if err != nil {
		t.Fatal(err)
	}

	got := map[string]string{}
	for _, p := range res.Added {
		got[p.URL] = p.Name + " in " + p.Path
	}
	want := map[string]string{
		"git@github.com:org-a/docs.git": "docs in docs",
		"git@github.com:org-b/docs.git": "docs-2 in docs-2",
		"git@github.com:org-c/docs.git": "docs-3 in c/docs", // its own directory already
		"git@gitlab.com:x/docs.git":     "docs-4 in vendor/docs",
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("added = %v\nwant    %v", got, want)
	}
}

func TestRefresh_DryRunFailsLikeTheRealRun(t *testing.T) {
	// A dry run that reports changes which the real run then refuses to write
	// is of no use.
	input := "projects: {a: {url: x}}\n"
	withDiscoverLocal(t, func(string) ([]Project, error) {
		return []Project{{Name: "repo", Path: "repo"}}, nil
	})

	for _, dryRun := range []bool{true, false} {
		path := writeConfig(t, input)
		_, err := refreshMani(path, false, dryRun, fixedFetch(nil), nil, nil)
		if err == nil || !strings.Contains(err.Error(), "not supported") {
			t.Errorf("dry run %t: expected an unsupported layout error, got %v", dryRun, err)
		}
		if got := readConfig(t, path); got != input {
			t.Errorf("dry run %t: the config was modified:\n%s", dryRun, got)
		}
	}
}
