package dao

import (
	"errors"
	"fmt"
	"os"
	"reflect"
	"slices"
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

	projects, err := editableProjects(ydoc, raw)
	if err != nil {
		return fmt.Errorf("%s: %w", configPath, err)
	}

	if projects != nil {
		if projects.Entry(p.Name) != nil {
			return &core.ProjectAlreadyExists{Name: p.Name}
		}
		projects.InsertEntrySpaced(yamlQuote(p.Name), renderProjectLines(p))
	} else {
		// No `projects:` block at all — create one at the root.
		ydoc.Root().InsertEntry("projects", renderProjectBlock(p, ydoc.Root().Indent()+2))
	}

	return writeValidatedConfig(configPath, raw, ydoc.Bytes(), []string{p.Name}, nil)
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

	projects, err := editableProjects(ydoc, raw)
	if err != nil {
		return fmt.Errorf("%s: %w", configPath, err)
	}
	if projects == nil || !projects.RemoveEntry(name) {
		return &core.ProjectNotFound{Name: []string{name}}
	}

	return writeValidatedConfig(configPath, raw, ydoc.Bytes(), nil, []string{name})
}

// ProjectInConfig reports whether the config at configPath defines the project
// itself. A project of the loaded config may also come from a config that
// configPath imports, and those can't be edited through configPath.
func ProjectInConfig(configPath string, name string) (bool, error) {
	configPath, raw, err := readConfigForEdit(configPath)
	if err != nil {
		return false, err
	}

	var root yaml.Node
	if err := yaml.Unmarshal(raw, &root); err != nil {
		return false, fmt.Errorf("parsing %s: %w", configPath, err)
	}
	projectsNode := findChild(documentNode(&root), "projects")
	if projectsNode == nil {
		return false, nil
	}
	projects, err := projectsFromNode(projectsNode)
	if err != nil {
		return false, fmt.Errorf("%s: %w", configPath, err)
	}

	return slices.ContainsFunc(projects, func(p Project) bool { return p.Name == name }), nil
}

// editableProjects returns the `projects:` mapping of a config for editing, or
// nil when the config has no `projects:` key.
//
// The editor works on the layout of the file and yaml.v3 on its content. An
// edit is only safe when both see the same projects, so anything else (a flow
// mapping, for instance) is an error.
func editableProjects(ydoc *yamledit.Document, raw []byte) (*yamledit.Mapping, error) {
	unsupported := errors.New("the layout of the `projects:` block is not supported for editing")

	var root yaml.Node
	if err := yaml.Unmarshal(raw, &root); err != nil {
		return nil, err
	}
	projectsNode := findChild(documentNode(&root), "projects")
	entry := ydoc.Root().Entry("projects")
	if projectsNode == nil {
		if entry != nil {
			return nil, unsupported
		}
		return nil, nil
	}
	if entry == nil {
		return nil, unsupported
	}

	var names []string
	switch {
	case projectsNode.Kind == yaml.MappingNode && projectsNode.Style&yaml.FlowStyle == 0:
		for i := 0; i+1 < len(projectsNode.Content); i += 2 {
			names = append(names, projectsNode.Content[i].Value)
		}
	case projectsNode.Kind == yaml.ScalarNode && projectsNode.Tag == "!!null" && projectsNode.Value == "":
		// `projects:` without a value. An explicit null (`projects: ~`) is
		// not supported: the projects would be written after it.
	default:
		return nil, unsupported
	}

	projects := entry.EnsureMapping(ydoc.Root().Indent() + 2)
	if !slices.Equal(projects.Keys(), names) {
		return nil, unsupported
	}

	return projects, nil
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

// writeValidatedConfig writes edited atomically, but only if it is valid YAML
// and differs from before in nothing but the added and removed projects. A
// latent yamledit bug therefore degrades to "refused to write", never to a
// config that is corrupt or has silently lost content.
func writeValidatedConfig(configPath string, before, edited []byte, added, removed []string) error {
	if err := verifyProjectEdit(before, edited, added, removed); err != nil {
		return fmt.Errorf("%s: %w", configPath, err)
	}
	return atomicWrite(configPath, edited)
}

// verifyProjectEdit checks, on the content of the two documents as yaml.v3
// reads them, that after is before with the added projects added and the
// removed projects removed, and that nothing else has changed.
func verifyProjectEdit(before, after []byte, added, removed []string) error {
	old, err := decodeConfig(before)
	if err != nil {
		return fmt.Errorf("config is not valid YAML, aborting: %w", err)
	}
	edited, err := decodeConfig(after)
	if err != nil {
		return fmt.Errorf("edit would produce invalid YAML, aborting: %w", err)
	}

	changed := func(what string) error {
		return fmt.Errorf("edit would change %s, aborting. This is a bug in mani, the config has not been modified", what)
	}

	for key, value := range old {
		if key == "projects" {
			continue
		}
		if editedValue, ok := edited[key]; !ok || !reflect.DeepEqual(value, editedValue) {
			return changed("`" + key + "`")
		}
	}
	for key := range edited {
		if _, ok := old[key]; !ok && key != "projects" {
			return changed("`" + key + "`")
		}
	}

	oldProjects, ok := projectsOf(old)
	if !ok {
		return changed("`projects`")
	}
	editedProjects, ok := projectsOf(edited)
	if !ok {
		return changed("`projects`")
	}
	for name, project := range oldProjects {
		editedProject, ok := editedProjects[name]
		switch {
		case slices.Contains(removed, name):
			// The name of a removed project may be taken by an added one
			if ok && !slices.Contains(added, name) {
				return changed("project `" + name + "`, which should have been removed")
			}
		case !ok || !reflect.DeepEqual(project, editedProject):
			return changed("project `" + name + "`")
		}
	}
	for name := range editedProjects {
		if _, ok := oldProjects[name]; !ok && !slices.Contains(added, name) {
			return changed("project `" + name + "`")
		}
	}
	for _, name := range added {
		if _, ok := editedProjects[name]; !ok {
			return changed("project `" + name + "`, which should have been added")
		}
	}

	return nil
}

// decodeConfig decodes a config for comparison. Mappings with a key that isn't
// a string (a project named 2048, or true) are decoded by yaml.v3 with keys of
// their own type. They are turned into strings, as they are for mani, so that
// every mapping is compared the same way.
func decodeConfig(raw []byte) (map[string]any, error) {
	var decoded any
	if err := yaml.Unmarshal(raw, &decoded); err != nil {
		return nil, err
	}

	switch config := stringKeys(decoded).(type) {
	case nil:
		return map[string]any{}, nil
	case map[string]any:
		return config, nil
	}
	return nil, errors.New("the config is not a mapping")
}

func stringKeys(v any) any {
	switch v := v.(type) {
	case map[string]any:
		for key, value := range v {
			v[key] = stringKeys(value)
		}
	case map[any]any:
		out := make(map[string]any, len(v))
		for key, value := range v {
			out[fmt.Sprint(key)] = stringKeys(value)
		}
		return out
	case []any:
		for i, value := range v {
			v[i] = stringKeys(value)
		}
	}
	return v
}

// projectsOf returns the projects of a decoded config. A config without
// projects has none; ok is false when `projects` isn't a mapping.
func projectsOf(config map[string]any) (projects map[string]any, ok bool) {
	switch v := config["projects"].(type) {
	case nil:
		return nil, true
	case map[string]any:
		return v, true
	}
	return nil, false
}

// renderProjectBlock renders a full `name:`-rooted project block: the project
// key at `indent` spaces and its fields nested two further. Used only when the
// config has no `projects:` block yet and one must be created from scratch.
func renderProjectBlock(p Project, indent int) []string {
	pad := strings.Repeat(" ", indent)
	lines := []string{pad + yamlQuote(p.Name) + ":"}
	for _, l := range renderProjectLines(p) {
		lines = append(lines, pad+l)
	}
	return lines
}
