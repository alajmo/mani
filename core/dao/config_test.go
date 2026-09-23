package dao

import (
	"testing"

	"gopkg.in/yaml.v3"
)

func TestConfig_DuplicateProjectName(t *testing.T) {
	originalProjects := []Project{
		{Name: "project-a", Path: "sub-1/project-a"},
		{Name: "project-a", Path: "sub-2/project-a"},
		{Name: "project-b", Path: "sub-3/project-b"},
	}

	var projects []Project
	projects = append(projects, originalProjects...)
	RenameDuplicates(projects)

	if projects[0].Name != originalProjects[0].Path {
		t.Fatalf(`Wanted: %q, Found: %q`, projects[0].Path, originalProjects[0].Name)
	}

	if projects[1].Name != originalProjects[1].Path {
		t.Fatalf(`Wanted: %q, Found: %q`, projects[1].Path, originalProjects[1].Name)
	}

	if originalProjects[2].Name != projects[2].Name {
		t.Fatalf(`Wanted: %q, Found: %q`, projects[2].Name, originalProjects[2].Name)
	}
}

func names(ps []Project) []string {
	out := make([]string, len(ps))
	for i, p := range ps {
		out[i] = p.Name
	}
	return out
}

func TestReconcileDiscovered(t *testing.T) {
	// Collected = what a remote --source fetch produced. `alpha` uses a source
	// path, so its path differs from where it'd be cloned on disk.
	collected := []Project{
		{Name: "alpha", Path: "acme/alpha", URL: "git@github.com:acme/alpha.git"},
		{Name: "beta", Path: "acme/beta", URL: "git@github.com:acme/beta.git"},
	}
	discovered := []Project{
		// Same repo as a source repo by URL (cloned flat at ./alpha) — dropped.
		{Name: "alpha", Path: "alpha", URL: "git@github.com:acme/alpha.git"},
		// Genuinely new on-disk repo — kept verbatim.
		{Name: "gamma", Path: "gamma", URL: "git@github.com:acme/gamma.git"},
		// Distinct repo whose name collides with source `beta` — keyed uniquely.
		{Name: "beta", Path: "vendor/beta", URL: "git@example.com:other/beta.git"},
	}

	got := reconcileDiscovered(collected, discovered)

	want := []string{"gamma", "beta-2"}
	if g := names(got); !equalStrings(g, want) {
		t.Fatalf("reconcileDiscovered names = %v, want %v", g, want)
	}
	// The renamed entry keeps its real path/URL so it still clones correctly.
	if got[1].Path != "vendor/beta" || got[1].URL != "git@example.com:other/beta.git" {
		t.Fatalf("renamed entry lost its fields: %+v", got[1])
	}
}

func TestReconcileDiscovered_DropsSamePath(t *testing.T) {
	// A URL-less collected project (e.g. the root repo, or a path-only entry)
	// still shadows an on-disk repo at the same path.
	collected := []Project{{Name: "root", Path: ".", URL: ""}, {Name: "lib", Path: "lib"}}
	discovered := []Project{
		{Name: "lib", Path: "lib", URL: "git@github.com:x/lib.git"}, // same path -> dropped
		{Name: "app", Path: "app", URL: "git@github.com:x/app.git"},
	}
	if g := names(reconcileDiscovered(collected, discovered)); !equalStrings(g, []string{"app"}) {
		t.Fatalf("names = %v, want [app]", g)
	}
}

func TestReconcileDiscovered_KeepsLocalCloneSiblings(t *testing.T) {
	// Two on-disk clones of one remote are distinct projects — intra-discovered
	// URLs are NOT deduped (matches plain `mani init`).
	discovered := []Project{
		{Name: "a", Path: "a", URL: "git@github.com:x/repo.git"},
		{Name: "b", Path: "b", URL: "git@github.com:x/repo.git"},
	}
	if g := names(reconcileDiscovered(nil, discovered)); !equalStrings(g, []string{"a", "b"}) {
		t.Fatalf("names = %v, want [a b]", g)
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

func TestYamlQuote(t *testing.T) {
	// Safe scalars round-trip bare (preserving mani's plain `mani init` output);
	// only values that would otherwise be mis-parsed get quoted.
	cases := []struct {
		in   string
		want string
	}{
		{"git@github.com:acme/alpha.git", "git@github.com:acme/alpha.git"}, // colon, no space → bare
		{"acme/alpha", "acme/alpha"},
		{"go", "go"},
		{"my-org", "my-org"},
	}
	for _, c := range cases {
		if got := yamlQuote(c.in); got != c.want {
			t.Errorf("yamlQuote(%q) = %q, want bare %q", c.in, got, c.want)
		}
	}

	// These MUST be quoted, else they break or change meaning in YAML. We don't
	// pin the exact quoting style (single vs double), only that the value
	// survives a round-trip as the original string.
	for _, in := range []string{
		"weird: value", // ": " would start a nested mapping
		"- dashy",      // leading "-" looks like a sequence item
		"# hashy",      // leading "#" is a comment
		"path with: colon and #hash",
		"123",  // would decode as int without quoting
		"true", // would decode as bool
		"",     // empty
	} {
		out := yamlQuote(in)
		var got string
		if err := yaml.Unmarshal([]byte("v: "+out), &struct{ V *string }{&got}); err != nil {
			t.Errorf("yamlQuote(%q) = %q produced unparseable YAML: %v", in, out, err)
			continue
		}
		if got != in {
			t.Errorf("yamlQuote(%q) = %q round-tripped to %q", in, out, got)
		}
	}
}
