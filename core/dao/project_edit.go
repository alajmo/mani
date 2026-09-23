package dao

import (
	"fmt"
	"os"
	"strings"

	"github.com/alajmo/mani/core"
	"github.com/alajmo/mani/core/yamledit"
	"gopkg.in/yaml.v3"
)

// AddProject inserts project p into the `projects:` mapping of the mani config
// at configPath, preserving every byte outside the inserted entry: comments,
// blank lines, indentation, and any hand-customized fields all survive. If
// configPath is empty it is resolved via the standard config-file search in
// the current directory. It errors if a project with the same key already
// exists, and never writes a file that yaml.v3 cannot parse.
func AddProject(configPath string, p Project) error {
	configPath, raw, err := readConfigForEdit(configPath)
	if err != nil {
		return err
	}

	ydoc, err := yamledit.Parse(raw)
	if err != nil {
		return fmt.Errorf("parsing %s for editing: %w", configPath, err)
	}

	if pe := ydoc.Root().Entry("projects"); pe != nil {
		projects := pe.EnsureMapping(ydoc.Root().Indent() + 2)
		if projects.Entry(p.Name) != nil {
			return &core.ProjectAlreadyExists{Name: p.Name}
		}
		projects.InsertEntrySpaced(p.Name, renderProjectLines(p))
	} else {
		// No `projects:` block at all — create one at the root.
		ydoc.Root().InsertEntry("projects", renderProjectBlock(p, ydoc.Root().Indent()+2))
	}

	return writeValidatedConfig(configPath, ydoc.Bytes())
}

// RemoveProject deletes the project named name — together with the comment and
// blank lines directly above it — from configPath's `projects:` mapping. All
// other formatting is preserved. It errors with ProjectNotFound when no such
// entry exists.
func RemoveProject(configPath string, name string) error {
	configPath, raw, err := readConfigForEdit(configPath)
	if err != nil {
		return err
	}

	ydoc, err := yamledit.Parse(raw)
	if err != nil {
		return fmt.Errorf("parsing %s for editing: %w", configPath, err)
	}

	pe := ydoc.Root().Entry("projects")
	if pe == nil {
		return &core.ProjectNotFound{Name: []string{name}}
	}
	projects := pe.EnsureMapping(ydoc.Root().Indent() + 2)
	if !projects.RemoveEntry(name) {
		return &core.ProjectNotFound{Name: []string{name}}
	}

	return writeValidatedConfig(configPath, ydoc.Bytes())
}

// IsRemoteURL reports whether s looks like a git remote URL rather than a bare
// project name. It recognizes scheme URLs (https://, http://, ssh://, git://)
// and scp-style SSH (git@host:org/repo.git).
func IsRemoteURL(s string) bool {
	if strings.Contains(s, "://") {
		return true
	}
	// scp-style: user@host:path — an "@" with a ":" after the host.
	if at := strings.IndexByte(s, '@'); at > 0 {
		return strings.Contains(s[at+1:], ":")
	}
	return false
}

// ProjectNameFromURL derives a project name from a git URL: the final path
// segment, with any trailing "/" and ".git" suffix removed
// (e.g. "https://github.com/alajmo/mani.git" -> "mani"). Returns "" if no
// usable segment can be extracted.
func ProjectNameFromURL(url string) string {
	s := strings.TrimRight(url, "/")
	if i := strings.LastIndexAny(s, "/:"); i >= 0 {
		s = s[i+1:]
	}
	return strings.TrimSuffix(s, ".git")
}

// readConfigForEdit resolves an empty configPath via the standard search and
// reads the file.
func readConfigForEdit(configPath string) (string, []byte, error) {
	if configPath == "" {
		resolved, err := findManiConfig()
		if err != nil {
			return "", nil, err
		}
		configPath = resolved
	}
	raw, err := os.ReadFile(configPath)
	if err != nil {
		return "", nil, fmt.Errorf("reading %s: %w", configPath, err)
	}
	return configPath, raw, nil
}

// writeValidatedConfig runs the yaml.v3 validation gate over edited and, only
// if it parses, writes it atomically. A latent yamledit bug therefore degrades
// to "refused to write", never a corrupted config.
func writeValidatedConfig(configPath string, edited []byte) error {
	var probe yaml.Node
	if err := yaml.Unmarshal(edited, &probe); err != nil {
		return fmt.Errorf("edit would produce invalid YAML, aborting: %w", err)
	}
	return atomicWrite(configPath, edited)
}

// renderProjectBlock renders a full `name:`-rooted project block: the project
// key at `indent` spaces and its fields nested two further. Used only when the
// config has no `projects:` block yet and one must be created from scratch.
func renderProjectBlock(p Project, indent int) []string {
	pad := strings.Repeat(" ", indent)
	lines := []string{pad + p.Name + ":"}
	for _, l := range renderProjectLines(p) {
		lines = append(lines, pad+l)
	}
	return lines
}
