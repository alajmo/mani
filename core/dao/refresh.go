package dao

import (
	"context"
	"fmt"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/alajmo/mani/core"
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
type fetchFunc func(ctx context.Context, src sources.Source) (sourceRepos, error)

func defaultFetch(ctx context.Context, src sources.Source) (sourceRepos, error) {
	if err := CheckProviderCLI(src); err != nil {
		return sourceRepos{}, err
	}
	return fetchListing(ctx, src)
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
// configured holds the projects of the loaded config: those of configPath and
// of every config it imports, with resolved (absolute) paths. They tell refresh
// which repositories are projects already, so that a project defined in an
// imported config isn't added to configPath as well. Only the projects in
// configPath itself are ever edited or pruned. When nil, only the projects in
// configPath are known.
//
// If configPath is empty, it's resolved via the standard ACCEPTABLE_FILE_NAMES
// chain in the current directory.
func RefreshMani(configPath string, prune, dryRun bool, autoDiscovery *bool, configured []Project) (*RefreshResult, error) {
	return refreshMani(configPath, prune, dryRun, defaultFetch, autoDiscovery, configured)
}

func refreshMani(configPath string, prune, dryRun bool, fetch fetchFunc, autoDiscovery *bool, configured []Project) (*RefreshResult, error) {
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

	// Paths in configPath are relative to its project_dir
	projectDir, err := projectDirOf(doc, filepath.Dir(configPath))
	if err != nil {
		return nil, err
	}

	configuredKey := make(map[string]bool, len(configured))
	for _, p := range configured {
		configuredKey[projectKey(p)] = true
	}

	// The directories of the projects there already are. A repo of a source is
	// the same project as the one in its directory, also when their URLs can't
	// be matched (an ssh host alias, or a project without a URL).
	projectDirs := map[string]string{}
	for _, p := range existing {
		if dir, err := core.GetAbsolutePath(projectDir, p.Path, p.Name); err == nil {
			projectDirs[filepath.Clean(dir)] = projectKey(p)
		}
	}
	for _, p := range configured {
		projectDirs[filepath.Clean(p.Path)] = projectKey(p)
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
	var listings []*sourceListing
	for _, src := range srcs {
		listed, err := fetch(context.Background(), src)
		if err != nil {
			return nil, fmt.Errorf("refreshing %s: %w", src.Label(), err)
		}
		fetched := listed.Projects
		listing := newSourceListing(src, listed, fetch)
		listings = append(listings, listing)

		added, stale, _ := diffProjects(existing, fetched, src.Name())

		// The listing leaves out archived repos and forks unless the source
		// includes them, and everything beyond its limit. Those repos still
		// exist, so a project is only stale if it's missing from the complete
		// listing as well.
		var gone []Project
		for _, p := range stale {
			exists, err := listing.upstream(projectKey(p))
			if err != nil {
				return nil, err
			}
			if !exists {
				gone = append(gone, p)
			}
		}

		removedKey := map[string]bool{}
		for _, p := range gone {
			if staleSeen[projectKey(p)] {
				continue
			}
			staleSeen[projectKey(p)] = true
			if prune {
				result.Removed = append(result.Removed, p)
				removedKey[projectKey(p)] = true
			} else {
				result.Stale = append(result.Stale, p)
			}
		}

		for _, p := range added {
			// Already a project, defined in an imported config
			if configuredKey[projectKey(p)] {
				continue
			}
			// Already a project, with a URL that isn't the one of the source.
			// Unless that project is removed, which leaves the directory to
			// the repo of the source.
			if dir, err := core.GetAbsolutePath(projectDir, p.Path, p.Name); err == nil {
				if key, ok := projectDirs[filepath.Clean(dir)]; ok && !removedKey[key] {
					continue
				}
			}
			if addedSeen[projectKey(p)] {
				continue
			}
			addedSeen[projectKey(p)] = true
			result.Added = append(result.Added, p)
		}
	}

	// The repos queued so far are those of the sources. What is discovered on
	// disk is added after them.
	fromSources := len(result.Added)

	// A clone of a repo that is gone upstream stays on disk when its project
	// is pruned. Discovery must not add it back, or the next refresh would
	// prune it again.
	gone := func(url string) (bool, error) {
		for _, listing := range listings {
			if !urlOwnedBy(url, listing.src.Name(), listing.hosts) {
				continue
			}
			exists, err := listing.upstream("url:" + urlKey(url))
			if err != nil {
				return false, err
			}
			if !exists {
				return true, nil
			}
		}
		return false, nil
	}

	// Local discovery: append on-disk git repos not already configured (or
	// already queued by a source above). Matched by path, then URL, so a repo
	// pulled in by a source isn't added a second time from disk.
	if autoDisc {
		local, err := discoverLocalProjects(filepath.Dir(configPath), projectDir, existing, result.Added, configured, gone)
		if err != nil {
			return nil, fmt.Errorf("discovering local repositories in %s: %w", filepath.Dir(configPath), err)
		}
		result.Added = append(result.Added, local...)
	}

	// Projects are matched by URL, so an upstream repo whose name equals an
	// existing project key (or another added project) can slip through as
	// "added" — writing it verbatim would produce a duplicate YAML mapping
	// key and corrupt the config. Disambiguate added names by suffixing.
	//
	// Names of projects in imported configs are taken as well: projects with
	// the same name in different configs are combined into one project.
	taken := make(map[string]bool, len(existing))
	for _, p := range existing {
		taken[p.Name] = true
	}
	for _, p := range configured {
		taken[p.Name] = true
	}
	for _, p := range result.Removed {
		delete(taken, p.Name)
	}
	//
	// Two sources may each have a repo with the same name, which without a
	// `path:` on the source also is the same directory. The second one gets
	// another name, and is moved to a directory of that name: two repos can't
	// be cloned into one directory. What is discovered on disk stays where it
	// is.
	addedDirs := map[string]bool{}
	for i := range result.Added {
		name := result.Added[i].Name
		if taken[name] {
			name = uniqueProjectName(name, taken)
			result.Added[i].Name = name
		}
		taken[name] = true

		dir := func() string {
			abs, err := core.GetAbsolutePath(projectDir, result.Added[i].Path, name)
			if err != nil {
				return ""
			}
			return filepath.Clean(abs)
		}
		if i < fromSources && addedDirs[dir()] {
			result.Added[i].Path = path.Join(path.Dir(filepath.ToSlash(result.Added[i].Path)), name)
		}
		addedDirs[dir()] = true
	}

	result.Kept = keptProjects(existing, result.Added, result.Removed, result.Stale)

	if len(result.Added) == 0 && len(result.Removed) == 0 {
		return result, nil
	}

	// Apply the additions/removals losslessly: yamledit edits only the
	// touched mapping entries and re-emits every other byte verbatim, so
	// indentation, blank lines, and comments survive untouched.
	edited, err := applyProjectEdits(raw, result.Added, result.Removed, prune)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", configPath, err)
	}

	// Validation gate: the edited config must be valid, and differ in nothing
	// but the added and removed projects, before we touch disk. A latent
	// yamledit bug degrades to "refresh refused to write", never "refresh
	// corrupted my config". A dry run goes through the same checks, so that it
	// doesn't report changes that the real run would refuse.
	names := func(projects []Project) []string {
		out := make([]string, len(projects))
		for i, p := range projects {
			out[i] = p.Name
		}
		return out
	}
	if err := verifyProjectEdit(raw, edited, names(result.Added), names(result.Removed)); err != nil {
		return nil, fmt.Errorf("%s: %w", configPath, err)
	}
	if dryRun {
		return result, nil
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
	projects, err := editableProjects(ydoc, raw)
	if err != nil {
		return nil, err
	}

	// No `projects:` block at all — create one at the root, as `mani project
	// add` does. There is nothing to remove in that case.
	if projects == nil {
		if len(added) == 0 {
			return raw, nil
		}
		var lines []string
		for i, p := range added {
			if i > 0 {
				lines = append(lines, "")
			}
			lines = append(lines, renderProjectBlock(p, ydoc.Root().Indent()+2)...)
		}
		ydoc.Root().InsertEntry("projects", lines)
		return ydoc.Bytes(), nil
	}

	if prune {
		for _, p := range removed {
			projects.RemoveEntry(p.Name)
		}
	}
	for _, p := range added {
		projects.InsertEntrySpaced(yamlQuote(p.Name), renderProjectLines(p))
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
			quoted[i] = yamlQuoteFlow(t)
		}
		lines = append(lines, "  tags: ["+strings.Join(quoted, ", ")+"]")
	}
	return lines
}

// sourceListing is what a source returned, and on demand what it left out.
type sourceListing struct {
	src     sources.Source
	hosts   map[string]bool // hosts of the repos of the source
	partial bool            // the source can't list every repo that exists
	fetch   fetchFunc

	listed   map[string]bool // keys of the repos in the listing
	complete map[string]bool // keys of all repos upstream, nil until needed
}

func newSourceListing(src sources.Source, listed sourceRepos, fetch fetchFunc) *sourceListing {
	l := &sourceListing{
		src:     src,
		hosts:   hostsOf(listed.Projects),
		partial: listed.Partial,
		fetch:   fetch,
		listed:  make(map[string]bool, len(listed.Projects)),
	}
	for _, p := range listed.Projects {
		l.listed[projectKey(p)] = true
	}
	return l
}

// upstream reports whether the repo with the given project key exists
// upstream: in the listing of the source, or among the repos that the listing
// leaves out (archived repos and forks unless the source includes them, and
// everything beyond its limit). The complete listing is only fetched when a
// repo isn't in the listing, and only once.
func (l *sourceListing) upstream(key string) (bool, error) {
	if l.listed[key] {
		return true, nil
	}
	// A source that can't list every repo can't tell that one is gone. That
	// is the case for the private repos of a user that isn't the one logged
	// in: they are taken to exist, since removing a project that is there is
	// worse than keeping one that isn't.
	if l.partial {
		return true, nil
	}
	if l.src.Archived && l.src.Fork && l.src.Limit == 0 {
		return false, nil // The listing is complete
	}

	if l.complete == nil {
		complete := l.src
		complete.Archived, complete.Fork, complete.Limit = true, true, 0
		listed, err := l.fetch(context.Background(), complete)
		if err != nil {
			return false, fmt.Errorf("refreshing %s: %w", l.src.Label(), err)
		}
		l.complete = make(map[string]bool, len(listed.Projects))
		for _, p := range listed.Projects {
			l.complete[projectKey(p)] = true
		}
	}

	return l.complete[key], nil
}

// hostsOf returns the hosts of the URLs of projects.
func hostsOf(projects []Project) map[string]bool {
	hosts := map[string]bool{}
	for _, p := range projects {
		if host, _, ok := parseGitURL(p.URL); ok {
			hosts[host] = true
		}
	}
	return hosts
}

// diffProjects compares existing vs fetched, keyed by URL with Name fallback.
// stale is only the subset of "in existing, not in fetched" whose URL is
// owned by owner on a host of the fetched repos (so we don't prune projects
// from a different remote). Nothing is stale when nothing was fetched: the
// host is not known then, and an empty listing is more likely a problem with
// the provider than every repository being gone.
func diffProjects(existing, fetched []Project, owner string) (added, stale, kept []Project) {
	hosts := hostsOf(fetched)

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
		if owner != "" && !urlOwnedBy(p.URL, owner, hosts) {
			continue
		}
		stale = append(stale, p)
	}
	return added, stale, kept
}

func projectKey(p Project) string {
	if p.URL != "" {
		return "url:" + urlKey(p.URL)
	}
	return "name:" + p.Name
}

// urlKey returns a key that is the same for every URL of a repository: the
// ssh URL that sources report, and the https or ssh:// URL that a project may
// be configured with. Anything that isn't a remote URL (a local path) is
// returned as it is.
func urlKey(rawURL string) string {
	if host, repo, ok := parseGitURL(rawURL); ok {
		return host + "/" + repo
	}
	return rawURL
}

// parseGitURL splits a remote git URL into its host and the path of the
// repository on that host, without the scheme, user, port and `.git` suffix.
// Both are lower-cased, as hosts and the repository paths of the supported
// providers are case-insensitive. ok is false for anything else, such as a
// local path.
//
//	git@github.com:acme/api.git        -> github.com, acme/api
//	https://github.com/Acme/api        -> github.com, acme/api
//	ssh://git@host.xz:2222/acme/api    -> host.xz, acme/api
func parseGitURL(rawURL string) (host, repo string, ok bool) {
	rawURL = strings.TrimSpace(rawURL)

	if strings.Contains(rawURL, "://") {
		u, err := url.Parse(rawURL)
		if err != nil || u.Scheme == "file" {
			return "", "", false
		}
		host, repo = u.Hostname(), u.Path
	} else {
		// scp-like: [user@]host:path. A host has no slash, and more than one
		// character so that a Windows drive (C:\repos) isn't taken for one.
		var found bool
		host, repo, found = strings.Cut(rawURL, ":")
		if !found || strings.ContainsAny(host, `/\`) {
			return "", "", false
		}
		if i := strings.LastIndex(host, "@"); i >= 0 {
			host = host[i+1:]
		}
		if len(host) < 2 {
			return "", "", false
		}
	}

	repo = strings.TrimSuffix(strings.Trim(repo, "/"), ".git")
	if host == "" || repo == "" {
		return "", "", false
	}

	return strings.ToLower(host), strings.ToLower(repo), true
}

// urlOwnedBy reports whether a git URL is a repository of owner on one of
// hosts: the host has to match, and owner has to be the first segment of the
// repository path. Source "bar" matches neither a URL owned by "foobar" nor a
// repository of another "bar" on a different host.
func urlOwnedBy(rawURL, owner string, hosts map[string]bool) bool {
	host, repo, ok := parseGitURL(rawURL)
	if !ok || !hosts[host] {
		return false
	}
	first, _, _ := strings.Cut(repo, "/")
	return first == strings.ToLower(owner)
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

// discoverLocalProjects walks dir for git repositories and returns those that
// aren't projects already. A repo is known if its directory is the directory of
// a project, or if its remote URL is the URL of a project — the latter prevents
// a repo just appended from a source from being added a second time off disk.
//
// The known projects are existing (the projects in the config being refreshed),
// alreadyAdded (queued by a source) and configured (the projects of the loaded
// config, which includes those of imported configs). The paths of the first two
// are relative to projectDir, the paths of configured are absolute.
//
// The path of a returned project is relative to projectDir, as it's written to
// the config.
//
// gone reports whether a URL is of a repo that a source owns and that no longer
// exists upstream. Such a repo is left alone.
func discoverLocalProjects(dir, projectDir string, existing, alreadyAdded, configured []Project, gone func(url string) (bool, error)) ([]Project, error) {
	found, err := discoverLocal(dir)
	if err != nil {
		return nil, err
	}

	knownDir := map[string]bool{}
	knownURL := map[string]bool{}
	note := func(absPath string, url string) {
		knownDir[filepath.Clean(absPath)] = true
		if url != "" {
			knownURL[urlKey(url)] = true
		}
	}
	for _, set := range [][]Project{existing, alreadyAdded} {
		for _, p := range set {
			absPath, err := core.GetAbsolutePath(projectDir, p.Path, p.Name)
			if err != nil {
				return nil, err
			}
			note(absPath, p.URL)
		}
	}
	for _, p := range configured {
		note(p.Path, p.URL)
	}

	var local []Project
	for _, p := range found {
		absPath := filepath.Join(dir, p.Path)
		if knownDir[filepath.Clean(absPath)] {
			continue
		}
		if p.URL != "" && knownURL[urlKey(p.URL)] {
			continue
		}
		if p.URL != "" && gone != nil {
			isGone, err := gone(p.URL)
			if err != nil {
				return nil, err
			}
			if isGone {
				continue
			}
		}
		note(absPath, p.URL) // guard against two discovered repos sharing a path/URL

		if relPath, err := filepath.Rel(projectDir, absPath); err == nil {
			p.Path = filepath.ToSlash(relPath)
		} else {
			p.Path = filepath.ToSlash(absPath)
		}
		// Worktrees discovered on disk aren't emitted by renderProjectLines,
		// so drop them to keep the returned project in step with what's
		// actually written to mani.yaml.
		p.WorktreeList = nil
		local = append(local, p)
	}
	return local, nil
}

// projectDirOf returns the directory that the project paths in a config are
// relative to: its `project_dir`, or the directory of the config when not set.
func projectDirOf(doc *yaml.Node, configDir string) (string, error) {
	var projectDir string
	if n := findChild(doc, "project_dir"); n != nil {
		if err := n.Decode(&projectDir); err != nil {
			return "", fmt.Errorf("decoding project_dir: %w", err)
		}
	}

	return core.GetAbsolutePath(configDir, projectDir, "")
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
	// A config that is a symlink is written through the link. Renaming the
	// temp file over the link itself would replace it with a regular file,
	// and leave the file it points to unchanged. The temp file is created
	// next to the target, so that the rename stays within one file system.
	if target, err := filepath.EvalSymlinks(path); err == nil {
		path = target
	}
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
