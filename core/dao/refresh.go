package dao

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/alajmo/mani/core/sources"
	"github.com/alajmo/mani/core/yamledit"
	"gopkg.in/yaml.v3"
)

// RefreshResult summarizes what RefreshMani did (or would have done in
// --dry-run). Lists are per-source-aggregated; each project appears in
// exactly one bucket.
type RefreshResult struct {
	ConfigPath string
	Added      []Project
	Removed    []Project // populated only when Prune is true
	Stale      []Project // upstream-missing but not pruned (so callers can warn)
	Kept       []Project
	DryRun     bool
}

// fetchFunc fetches repos for a single source and returns them mapped to
// Projects. The default implementation dispatches by provider; tests inject
// a fake.
type fetchFunc func(ctx context.Context, src sources.Source) ([]Project, error)

func defaultFetch(ctx context.Context, src sources.Source) ([]Project, error) {
	if err := CheckProviderCLI(src); err != nil {
		return nil, err
	}
	return FetchAndMap(ctx, src)
}

// discoverLocal walks a directory for git repositories. It's a package var
// so tests can substitute a fake; production uses FindVCSystems (the same
// walk `mani init` performs).
var discoverLocal = FindVCSystems

// RefreshMani updates the projects: mapping in configPath from two sources:
// the remote repos returned by each entry in the sources: block (appended,
// and optionally pruned when their upstream is gone), and — when
// auto-discovery is enabled — git repositories found on disk under the
// config's directory but not yet listed (the same local walk `mani init`
// performs, so a freshly `git clone`d repo is picked up). Comments, ordering,
// indentation, blank lines, and custom fields are preserved: the file is
// edited losslessly with the yamledit package and only the touched entries
// are rewritten.
//
// autoDiscovery overrides the config's `auto_discovery` field when non-nil
// (used to plumb the --auto-discovery CLI flag); nil falls back to the field,
// which itself defaults to true.
//
// If configPath is empty, it's resolved via the standard ACCEPTABLE_FILE_NAMES
// chain in the current directory.
func RefreshMani(configPath string, prune, dryRun bool, autoDiscovery *bool) (*RefreshResult, error) {
	return refreshMani(configPath, prune, dryRun, defaultFetch, autoDiscovery)
}

func refreshMani(configPath string, prune, dryRun bool, fetch fetchFunc, autoDiscovery *bool) (*RefreshResult, error) {
	if configPath == "" {
		resolved, err := findManiConfig()
		if err != nil {
			return nil, err
		}
		configPath = resolved
	}

	raw, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("reading %s: %w", configPath, err)
	}

	var root yaml.Node
	if err := yaml.Unmarshal(raw, &root); err != nil {
		return nil, fmt.Errorf("parsing %s: %w", configPath, err)
	}
	doc := documentNode(&root)
	if doc == nil || doc.Kind != yaml.MappingNode {
		return nil, fmt.Errorf("%s: top-level is not a YAML mapping", configPath)
	}

	// Resolve auto-discovery: explicit flag (non-nil) wins, else the config
	// field, which defaults to true.
	autoDisc := readAutoDiscovery(doc)
	if autoDiscovery != nil {
		autoDisc = *autoDiscovery
	}

	// The sources: block is optional now — a config scaffolded by plain
	// `mani init` has none, yet local discovery still gives refresh useful
	// work. Only error when there's genuinely nothing to do.
	var srcs []sources.Source
	if sourcesNode := findChild(doc, "sources"); sourcesNode != nil {
		if err := sourcesNode.Decode(&srcs); err != nil {
			return nil, fmt.Errorf("decoding sources: %w", err)
		}
	}
	if len(srcs) == 0 && !autoDisc {
		return nil, fmt.Errorf("%s: nothing to refresh — no `sources:` block and auto-discovery is disabled (--auto-discovery)", configPath)
	}

	var existing []Project
	if projectsNode := findChild(doc, "projects"); projectsNode != nil {
		existing, err = projectsFromNode(projectsNode)
		if err != nil {
			return nil, fmt.Errorf("decoding projects: %w", err)
		}
	}

	result := &RefreshResult{ConfigPath: configPath, DryRun: dryRun}

	// Two sources can return the same upstream repo (e.g. a fork mirrored
	// across orgs, or the same source listed twice). diffProjects only knows
	// about `existing`, not what an earlier source already queued, so without
	// these guards the repo is added once per source — and the later name
	// dedup would keep both as `repo` and `repo-2`, two entries for one URL.
	// Track by projectKey across all sources so each repo lands at most once.
	addedSeen := make(map[string]bool)
	staleSeen := make(map[string]bool)
	for _, src := range srcs {
		fetched, err := fetch(context.Background(), src)
		if err != nil {
			return nil, fmt.Errorf("refreshing %s: %w", src.Label(), err)
		}

		added, stale, _ := diffProjects(existing, fetched, src.Name())
		for _, p := range added {
			if addedSeen[projectKey(p)] {
				continue
			}
			addedSeen[projectKey(p)] = true
			result.Added = append(result.Added, p)
		}
		for _, p := range stale {
			if staleSeen[projectKey(p)] {
				continue
			}
			staleSeen[projectKey(p)] = true
			if prune {
				result.Removed = append(result.Removed, p)
			} else {
				result.Stale = append(result.Stale, p)
			}
		}
	}

	// Local discovery: append on-disk git repos not already configured (or
	// already queued by a source above). Matched by path, then URL, so a repo
	// pulled in by a source isn't added a second time from disk.
	if autoDisc {
		local, err := discoverLocalProjects(filepath.Dir(configPath), existing, result.Added)
		if err != nil {
			return nil, fmt.Errorf("discovering local repositories in %s: %w", filepath.Dir(configPath), err)
		}
		result.Added = append(result.Added, local...)
	}

	// Projects are matched by URL, so an upstream repo whose name equals an
	// existing project key (or another added project) can slip through as
	// "added" — writing it verbatim would produce a duplicate YAML mapping
	// key and corrupt the config. Disambiguate added names by suffixing.
	taken := make(map[string]bool, len(existing))
	for _, p := range existing {
		taken[p.Name] = true
	}
	for _, p := range result.Removed {
		delete(taken, p.Name)
	}
	for i := range result.Added {
		name := result.Added[i].Name
		if taken[name] {
			name = uniqueProjectName(name, taken)
			result.Added[i].Name = name
		}
		taken[name] = true
	}

	result.Kept = keptProjects(existing, result.Added, result.Removed, result.Stale)

	if dryRun || (len(result.Added) == 0 && len(result.Removed) == 0) {
		return result, nil
	}

	// Apply the additions/removals losslessly: yamledit edits only the
	// touched mapping entries and re-emits every other byte verbatim, so
	// indentation, blank lines, and comments survive untouched.
	edited, err := applyProjectEdits(raw, result.Added, result.Removed, prune)
	if err != nil {
		return nil, err
	}

	// Validation gate: yaml.v3 must accept the edited bytes before we
	// touch disk. A latent yamledit bug degrades to "refresh refused to
	// write", never "refresh corrupted my config".
	var probe yaml.Node
	if err := yaml.Unmarshal(edited, &probe); err != nil {
		return nil, fmt.Errorf("refresh would produce invalid YAML, aborting: %w", err)
	}
	if err := atomicWrite(configPath, edited); err != nil {
		return nil, err
	}
	return result, nil
}

// applyProjectEdits rewrites raw's `projects:` mapping — appending added
// projects and (when prune) removing removed ones — using the lossless
// yamledit engine.
func applyProjectEdits(raw []byte, added, removed []Project, prune bool) ([]byte, error) {
	ydoc, err := yamledit.Parse(raw)
	if err != nil {
		return nil, fmt.Errorf("parsing config for editing: %w", err)
	}
	pe := ydoc.Root().Entry("projects")
	if pe == nil {
		return nil, fmt.Errorf("no `projects:` block found in config")
	}
	projects := pe.EnsureMapping(ydoc.Root().Indent() + 2)

	if prune {
		for _, p := range removed {
			projects.RemoveEntry(p.Name)
		}
	}
	for _, p := range added {
		projects.InsertEntrySpaced(p.Name, renderProjectLines(p))
	}
	return ydoc.Bytes(), nil
}

// renderProjectLines renders project p's value as YAML lines, each indented
// two spaces relative to the project key. Mirrors the field set the diff
// produces: path is emitted only when it differs from the name.
func renderProjectLines(p Project) []string {
	var lines []string
	if p.Path != "" && p.Path != p.Name {
		lines = append(lines, "  path: "+yamlQuote(p.Path))
	}
	if p.URL != "" {
		lines = append(lines, "  url: "+yamlQuote(p.URL))
	}
	if p.Desc != "" {
		lines = append(lines, "  desc: "+yamlQuote(p.Desc))
	}
	if p.Clone != "" {
		lines = append(lines, "  clone: "+yamlQuote(p.Clone))
	}
	if len(p.Tags) > 0 {
		quoted := make([]string, len(p.Tags))
		for i, t := range p.Tags {
			quoted[i] = yamlQuote(t)
		}
		lines = append(lines, "  tags: ["+strings.Join(quoted, ", ")+"]")
	}
	return lines
}

// diffProjects compares existing vs fetched, keyed by URL with Name fallback.
// stale is only the subset of "in existing, not in fetched" whose URL is
// owned by owner (so we don't prune projects from a different remote).
func diffProjects(existing, fetched []Project, owner string) (added, stale, kept []Project) {
	existingKey := map[string]Project{}
	for _, p := range existing {
		existingKey[projectKey(p)] = p
	}
	fetchedKey := map[string]Project{}
	for _, p := range fetched {
		fetchedKey[projectKey(p)] = p
	}
	for _, p := range fetched {
		if _, ok := existingKey[projectKey(p)]; ok {
			kept = append(kept, p)
		} else {
			added = append(added, p)
		}
	}
	for _, p := range existing {
		if _, ok := fetchedKey[projectKey(p)]; ok {
			continue
		}
		if owner != "" && !urlOwnedBy(p.URL, owner) {
			continue
		}
		stale = append(stale, p)
	}
	return added, stale, kept
}

func projectKey(p Project) string {
	if p.URL != "" {
		return "url:" + p.URL
	}
	return "name:" + p.Name
}

// urlOwnedBy reports whether a git URL's owner segment is exactly owner.
// It anchors on the separator preceding the owner (":" for scp-style SSH
// URLs, "/" for https URLs) so source "bar" does not match a URL owned by
// "foobar".
func urlOwnedBy(url, owner string) bool {
	return strings.Contains(url, ":"+owner+"/") || strings.Contains(url, "/"+owner+"/")
}

// uniqueProjectName returns base, or base-2, base-3, … — the first that is
// not already present in taken.
func uniqueProjectName(base string, taken map[string]bool) string {
	for i := 2; ; i++ {
		candidate := fmt.Sprintf("%s-%d", base, i)
		if !taken[candidate] {
			return candidate
		}
	}
}

func keptProjects(existing, added, removed, stale []Project) []Project {
	excluded := map[string]bool{}
	for _, set := range [][]Project{added, removed, stale} {
		for _, p := range set {
			excluded[projectKey(p)] = true
		}
	}
	var kept []Project
	for _, p := range existing {
		if excluded[projectKey(p)] {
			continue
		}
		kept = append(kept, p)
	}
	return kept
}

// discoverLocalProjects walks dir for git repositories and returns those not
// already represented in existing or alreadyAdded. A repo is considered known
// if its relative path matches (projects written without an explicit path use
// the project name as path, which is what FindVCSystems also reports) or its
// remote URL matches — the latter prevents a repo just appended from a source
// from being added a second time off disk.
func discoverLocalProjects(dir string, existing, alreadyAdded []Project) ([]Project, error) {
	found, err := discoverLocal(dir)
	if err != nil {
		return nil, err
	}

	knownPath := map[string]bool{}
	knownURL := map[string]bool{}
	note := func(p Project) {
		knownPath[filepath.Clean(p.Path)] = true
		if p.URL != "" {
			knownURL[p.URL] = true
		}
	}
	for _, p := range existing {
		note(p)
	}
	for _, p := range alreadyAdded {
		note(p)
	}

	var local []Project
	for _, p := range found {
		if knownPath[filepath.Clean(p.Path)] {
			continue
		}
		if p.URL != "" && knownURL[p.URL] {
			continue
		}
		// Worktrees discovered on disk aren't emitted by renderProjectLines,
		// so drop them to keep the returned project in step with what's
		// actually written to mani.yaml.
		p.WorktreeList = nil
		local = append(local, p)
		note(p) // guard against two discovered repos sharing a path/URL
	}
	return local, nil
}

// readAutoDiscovery reads the top-level `auto_discovery` field from the parsed
// config document, defaulting to true when absent or unparseable.
func readAutoDiscovery(doc *yaml.Node) bool {
	n := findChild(doc, "auto_discovery")
	if n == nil {
		return true
	}
	var v bool
	if err := n.Decode(&v); err != nil {
		return true
	}
	return v
}

// --- yaml.Node helpers --------------------------------------------------

func documentNode(n *yaml.Node) *yaml.Node {
	if n == nil {
		return nil
	}
	if n.Kind == yaml.DocumentNode && len(n.Content) > 0 {
		return n.Content[0]
	}
	return n
}

// findChild looks up a key in a YAML mapping node, returning its value node
// (or nil if absent).
func findChild(mapping *yaml.Node, key string) *yaml.Node {
	if mapping == nil || mapping.Kind != yaml.MappingNode {
		return nil
	}
	for i := 0; i+1 < len(mapping.Content); i += 2 {
		if mapping.Content[i].Value == key {
			return mapping.Content[i+1]
		}
	}
	return nil
}

// projectsFromNode decodes a `projects:` mapping into []Project, preserving
// only the fields refresh's diff cares about (Name, URL).
func projectsFromNode(projectsNode *yaml.Node) ([]Project, error) {
	// An empty `projects:` key parses as a null scalar; treat it as no
	// projects rather than an error, so refresh can populate it.
	if projectsNode.Kind == yaml.ScalarNode && (projectsNode.Tag == "!!null" || projectsNode.Value == "") {
		return nil, nil
	}
	if projectsNode.Kind != yaml.MappingNode {
		return nil, fmt.Errorf("`projects:` is not a mapping (got kind %d)", projectsNode.Kind)
	}
	var projects []Project
	for i := 0; i+1 < len(projectsNode.Content); i += 2 {
		name := projectsNode.Content[i].Value
		val := projectsNode.Content[i+1]
		p := Project{Name: name, Path: name}
		// Value can be null (shorthand `name:` with no body) or a mapping.
		if val.Kind == yaml.MappingNode {
			var pm map[string]any
			if err := val.Decode(&pm); err != nil {
				return nil, fmt.Errorf("decoding project %q: %w", name, err)
			}
			if u, ok := pm["url"].(string); ok {
				p.URL = u
			}
			if d, ok := pm["desc"].(string); ok {
				p.Desc = d
			}
			if path, ok := pm["path"].(string); ok {
				p.Path = path
			}
		}
		projects = append(projects, p)
	}
	return projects, nil
}

func atomicWrite(path string, data []byte) error {
	dir := filepath.Dir(path)
	// Preserve the original file's permissions — os.CreateTemp makes the
	// temp file 0600, which would otherwise silently lock down a config
	// that was previously group/world-readable.
	mode := os.FileMode(0644)
	if fi, err := os.Stat(path); err == nil {
		mode = fi.Mode().Perm()
	}
	tmp, err := os.CreateTemp(dir, filepath.Base(path)+".tmp-")
	if err != nil {
		return fmt.Errorf("creating temp file: %w", err)
	}
	cleanup := func() { _ = os.Remove(tmp.Name()) }
	if _, err := tmp.Write(data); err != nil {
		_ = tmp.Close()
		cleanup()
		return fmt.Errorf("writing temp file: %w", err)
	}
	if err := tmp.Close(); err != nil {
		cleanup()
		return err
	}
	if err := os.Chmod(tmp.Name(), mode); err != nil {
		cleanup()
		return fmt.Errorf("setting permissions on temp file: %w", err)
	}
	if err := os.Rename(tmp.Name(), path); err != nil {
		cleanup()
		return fmt.Errorf("replacing %s: %w", path, err)
	}
	return nil
}

func findManiConfig() (string, error) {
	wd, err := os.Getwd()
	if err != nil {
		return "", err
	}
	for _, name := range ACCEPTABLE_FILE_NAMES {
		candidate := filepath.Join(wd, name)
		if _, err := os.Stat(candidate); err == nil {
			return candidate, nil
		}
	}
	return "", fmt.Errorf("no mani config found in %s (looked for %s)", wd, strings.Join(ACCEPTABLE_FILE_NAMES, ", "))
}
