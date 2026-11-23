package dao

import (
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
	"time"

	"gopkg.in/yaml.v3"
)

func TestProjectDefinitionsRetainPresenceAndSource(t *testing.T) {
	config := Config{Path: "/configs/mani.yaml", ProjectDir: "../repos"}
	err := yaml.Unmarshal([]byte("projects:\n  app:\n    path: relative\n    desc: ''\n    sync: false\n    env:\n      VERSION: 1.0\n      TOKEN: $(exit 1)\n    worktrees:\n      - path: staging\n"), &config)
	if err != nil {
		t.Fatal(err)
	}
	defs, problems := config.parseProjectDefinitions()
	if len(problems) != 0 {
		t.Fatal(problems)
	}
	def := defs[0]
	if !def.desc.set || def.desc.value != "" || !def.sync.set || def.sync.value || def.branch.set {
		t.Fatalf("explicit empty/false values and absent fields lost: %+v", def)
	}
	if def.path.value != "relative" || def.worktrees[0].value.Branch != "" || def.env[1].value != "$(exit 1)" {
		t.Fatal("parsing must not resolve paths, apply defaults or evaluate env")
	}
	if def.env[0].value != "1.0" || def.env[0].source.line != 7 || def.env[0].source.file != config.Path || def.path.source.projectDir != "../repos" {
		t.Fatalf("lost scalar spelling or source: %+v", def)
	}
}

func TestProjectExplicitZeroOverrides(t *testing.T) {
	config, dir, err := readTestConfig(t, map[string]string{
		"mani.yaml":      "import: [base/mani.yaml]\nproject_dir: repos\nprojects:\n  app:\n    path: ''\n    desc: ''\n    url: ''\n    clone: ''\n    branch: ''\n    sync: false\n    single_branch: false\n    env: {VERSION: ''}\n",
		"base/mani.yaml": "projects:\n  app:\n    path: old\n    desc: inherited\n    url: repo\n    clone: clone-command\n    branch: main\n    sync: true\n    single_branch: true\n    env: {VERSION: 1.0}\n",
	})
	if err != nil {
		t.Fatal(err)
	}
	p := config.ProjectList[0]
	if p.Path != filepath.Join(dir, "repos/app") || p.Desc != "" || p.URL != "" || p.Clone != "" || p.Branch != "" || p.IsSync() || p.IsSingleBranch() || !slices.Equal(p.EnvList, []string{"VERSION="}) {
		t.Fatalf("zero values did not override: %+v", p)
	}
}

func TestProjectMergeConflicts(t *testing.T) {
	for _, test := range []struct{ name, a, b, root, want, source string }{
		{"scalar", "desc: a", "desc: b", "", "conflicting `desc`", "b.yaml:3"},
		{"bool", "sync: true", "sync: false", "", "conflicting `sync`", "b.yaml:3"},
		{"path", "path: a", "path: b", "", "conflicting `path`", "b.yaml:3"},
		{"env", "env: {VERSION: 1.0}", "env: {VERSION: 1}", "", "conflicting env `VERSION`", "b.yaml:3"},
		{"remote", "remotes: {origin: a}", "remotes: {origin: b}", "", "conflicting remote `origin`", "b.yaml:3"},
		{"worktree", "worktrees: [{path: ./wt, branch: a}]", "worktrees: [{path: wt, branch: b}]", "", "conflicting worktree `wt`", "b.yaml:3"},
		{"same env spelling", "env: {VERSION: 1.0}", "env: {VERSION: '1.0'}", "", "", ""},
		{"typed bool spelling", "sync: TRUE", "sync: true", "", "", ""},
		{"root resolves conflict", "desc: a", "desc: b", "projects: {app: {desc: root}}", "", ""},
	} {
		t.Run(test.name, func(t *testing.T) {
			_, dir, err := readTestConfig(t, map[string]string{
				"mani.yaml": "import: [a.yaml, b.yaml]\n" + test.root,
				"a.yaml":    "projects:\n  app:\n    " + test.a + "\n",
				"b.yaml":    "projects:\n  app:\n    " + test.b + "\n",
			})
			if test.want == "" {
				if err != nil {
					t.Fatal(err)
				}
				return
			}
			if err == nil || !strings.Contains(err.Error(), test.want) || !strings.Contains(err.Error(), filepath.Join(dir, test.source)) || !strings.Contains(err.Error(), filepath.Join(dir, "a.yaml:3")) {
				t.Fatalf("expected conflict with both field locations, got %v", err)
			}
		})
	}
}

func TestProjectInvalidConfigDoesNotEvaluateEnv(t *testing.T) {
	for _, test := range []struct {
		name, root string
		files      map[string]string
		want       string
	}{
		{"parse", "projects:\n  broken: {tags: {bad: value}}\n", nil, "cannot unmarshal"},
		{"merge", "import: [a.yaml, b.yaml]\n", map[string]string{"a.yaml": "projects: {broken: {desc: a}}", "b.yaml": "projects: {broken: {desc: b}}"}, "conflicting `desc`"},
		{"resolve", "import: [a/mani.yaml, b/mani.yaml]\n", map[string]string{"a/mani.yaml": "projects: {broken: {}}", "b/mani.yaml": "projects: {broken: {}}"}, "set `path`"},
		{"task reference", "tasks: {build: {cmd: echo, theme: missing}}\n", nil, "cannot find theme"},
		{"cycle", "import: [a.yaml]\n", map[string]string{"a.yaml": "import: [mani.yaml]"}, "circular dependency"},
	} {
		t.Run(test.name, func(t *testing.T) {
			files := map[string]string{"env.yaml": "projects:\n  marker:\n    env:\n      TOKEN: $(touch {dir}/ran)\n"}
			for name, content := range test.files {
				files[name] = content
			}
			// Import the side-effecting definition even for a root parse error.
			files["mani.yaml"] = "import: [entry.yaml, env.yaml]\n"
			files["entry.yaml"] = test.root
			_, dir, err := readTestConfig(t, files)
			if err == nil || !strings.Contains(err.Error(), test.want) {
				t.Fatalf("expected %q, got %v", test.want, err)
			}
			if _, err := os.Stat(filepath.Join(dir, "ran")); !os.IsNotExist(err) {
				t.Fatalf("env ran for invalid config: %v", err)
			}
		})
	}
}

func TestProjectPathsResolveAfterMerge(t *testing.T) {
	config, dir, err := readTestConfig(t, map[string]string{
		"mani.yaml":   "import: [a/mani.yaml, b/mani.yaml]\n",
		"a/mani.yaml": "project_dir: ..\nprojects: {app: {path: ./repos/app}}\n",
		"b/mani.yaml": "projects: {app: {path: ../repos/app}}\n",
	})
	if err != nil {
		t.Fatal(err)
	}
	if config.ProjectList[0].Path != filepath.Join(dir, "repos/app") {
		t.Fatal(config.ProjectList[0].Path)
	}
}

func TestClearedProjectURLUsesWorkspaceDefaultPath(t *testing.T) {
	config, dir, err := readTestConfig(t, map[string]string{
		"mani.yaml":      "import: [base/mani.yaml]\nprojects: {app: {url: ''}}\n",
		"base/mani.yaml": "projects: {app: {url: inherited-repo}}\n",
	})
	if err != nil {
		t.Fatal(err)
	}
	if p := config.ProjectList[0]; p.URL != "" || p.Path != filepath.Join(dir, "app") {
		t.Fatalf("cleared URL should not determine path ownership: %+v", p)
	}
}

func TestProjectNamesOnSameLineRemainDistinct(t *testing.T) {
	config, _, err := readTestConfig(t, map[string]string{
		"mani.yaml": "import: [base.yaml]\nprojects: {app: {desc: local}, tools: {desc: tools}}\n",
		"base.yaml": "projects: {app: {url: repo}, other: {}}\n",
	})
	if err != nil {
		t.Fatal(err)
	}
	var names []string
	for _, project := range config.ProjectList {
		names = append(names, project.Name)
	}
	if !slices.Equal(names, []string{"tools", "app", "other"}) {
		t.Fatalf("expected distinct projects in owner order, got %v", names)
	}
}

func TestProjectYAMLValidation(t *testing.T) {
	for _, input := range []string{
		"projects: [app]", "projects: {app: [bad]}",
		"projects:\n  app:\n  app:\n",
		"projects: {app: {sync: true, sync: false}}",
		"projects: {app: {env: [bad]}}", "projects: {app: {env: {A: [bad]}}}",
		"projects: {app: {env: {A: a, A: b}}}",
		"projects: {app: {remotes: [bad]}}", "projects: {app: {remotes: {origin: [bad]}}}",
		"projects: {app: {worktrees: {path: wt}}}", "projects: {app: {worktrees: [{branch: main}]}}",
		"projects: {app: {worktrees: [{path: ./wt, branch: a}, {path: wt, branch: b}]}}",
		"defaults: &defaults {<<: *defaults}\nprojects: {app: {<<: *defaults}}",
		"projects: {app: ''}",
		"projects:\n  app:\n    ? [sync]\n    : false\n",
		"more: &more {other: {}}\nprojects: {app: {}, <<: *more}",
		"projects: {app: {env: {A: a, A: a}}}",
		"projects: {app: {env: ''}}",
		"projects: ''",
		"defaults: &defaults {desc: a, desc: b}\nprojects: {app: {<<: *defaults}}",
		"env_defaults: &env {A: a, A: b}\nprojects: {app: {env: {<<: *env}}}",
	} {
		t.Run(input, func(t *testing.T) {
			if _, _, err := readTestConfig(t, map[string]string{"mani.yaml": input}); err == nil {
				t.Fatal("expected invalid project error")
			}
		})
	}
}

func TestProjectAliasesAndMergeKeys(t *testing.T) {
	config, _, err := readTestConfig(t, map[string]string{"mani.yaml": `
env_defaults: &env {VERSION: 1.0, MODE: base}
defaults: &defaults
  sync: false
  desc: base
  tags: [shared]
projects:
  app:
    <<: *defaults
    desc: local
    env:
      <<: *env
      MODE: local
  other: *defaults
`})
	if err != nil {
		t.Fatal(err)
	}
	app := config.ProjectList[0]
	if app.Desc != "local" || app.IsSync() || !slices.Equal(app.Tags, []string{"shared"}) || !slices.Equal(app.EnvList, []string{"VERSION=1.0", "MODE=local"}) {
		t.Fatalf("aliases lost: %+v", app)
	}
}

func TestProjectUserPrecedence(t *testing.T) {
	config, _, err := readTestConfig(t, map[string]string{
		"mani.yaml":        "import: [shared.yaml]\n",
		"shared.yaml":      "projects: {app: {desc: workspace, url: repo}}\n",
		"user-config.yaml": "import: [mani.yaml, extra.yaml]\nprojects: {app: {desc: user}}\n",
		"extra.yaml":       "projects: {app: {desc: extra, tags: [personal]}}\n",
	})
	if err != nil {
		t.Fatal(err)
	}
	if config.ProjectList[0].Desc != "workspace" || !slices.Equal(config.ProjectList[0].Tags, []string{"personal"}) {
		t.Fatal(config.ProjectList)
	}
}

func TestProjectMergeLeavesOtherResourcesUnmerged(t *testing.T) {
	config, _, err := readTestConfig(t, map[string]string{
		"mani.yaml": "import: [base.yaml]\nprojects: {app: {tags: [local]}}\ntasks: {build: echo local}\nenv: {MODE: local}\nspecs: {test: {parallel: true}}\ntargets: {test: {tags: [local]}}\nthemes: {test: {color: true}}\n",
		"base.yaml": "projects: {app: {tags: [shared]}}\ntasks: {build: echo shared}\nenv: {MODE: shared}\nspecs: {test: {parallel: false}}\ntargets: {test: {tags: [shared]}}\nthemes: {test: {color: false}}\n",
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(config.ProjectList) != 1 || len(config.TaskList) != 2 || len(config.SpecList) != 3 || len(config.TargetList) != 3 || len(config.ThemeList) != 3 || !slices.Equal(config.EnvList, []string{"MODE=local", "MODE=shared"}) {
		t.Fatal("only projects should be merged")
	}
}

func TestProjectNullValueDoesNotOverride(t *testing.T) {
	config, dir, err := readTestConfig(t, map[string]string{
		"mani.yaml": "import: [base.yaml]\nprojects:\n  app:\n    path:\n    desc:\n    branch: dev\n",
		"base.yaml": "projects:\n  app:\n    path: services/app\n    desc: inherited\n    branch: main\n",
	})
	if err != nil {
		t.Fatal(err)
	}
	p := config.ProjectList[0]
	if p.Path != filepath.Join(dir, "services/app") || p.Desc != "inherited" || p.Branch != "dev" {
		t.Fatalf("null values must not override inherited ones: %+v", p)
	}
}

func TestProjectEnvOrderWithMultipleProjects(t *testing.T) {
	config, _, err := readTestConfig(t, map[string]string{
		"mani.yaml": "projects:\n  a:\n    env: {X: one}\n  b:\n    env: {X: two}\n  c: {}\n",
	})
	if err != nil {
		t.Fatal(err)
	}
	got := map[string][]string{}
	for _, p := range config.ProjectList {
		got[p.Name] = p.EnvList
	}
	if !slices.Equal(got["a"], []string{"X=one"}) || !slices.Equal(got["b"], []string{"X=two"}) || len(got["c"]) != 0 {
		t.Fatalf("env assigned to wrong project: %v", got)
	}
}

func TestImportCycleStillReadsLaterImports(t *testing.T) {
	// b.yaml imports mani.yaml, closing a cycle, and then c.yaml, which must still be read: its
	// error is reported.
	_, dir, err := readTestConfig(t, map[string]string{
		"mani.yaml": "import: [b.yaml]\n",
		"b.yaml":    "import: [mani.yaml, c.yaml]\n",
		"c.yaml":    "projects:\n  c:\n    sync: notbool\n",
	})
	if err == nil || !strings.Contains(err.Error(), "circular dependency") || !strings.Contains(err.Error(), filepath.Join(dir, "c.yaml")) {
		t.Fatalf("expected cycle error and an error from c.yaml, got %v", err)
	}
}

func TestProjectMergeKeyPrecedence(t *testing.T) {
	// Earlier merged mappings take precedence over later ones, keys set in the mapping take
	// precedence over merged ones, and merging the same mapping twice isn't a duplicate
	config, _, err := readTestConfig(t, map[string]string{"mani.yaml": `
a: &a {desc: a, branch: a}
b: &b {desc: b, clone: b}
projects:
  app:
    <<: [*a, *b, *a]
    branch: local
`})
	if err != nil {
		t.Fatal(err)
	}
	if p := config.ProjectList[0]; p.Desc != "a" || p.Clone != "b" || p.Branch != "local" {
		t.Fatalf("wrong merge key precedence: %+v", p)
	}
}

func TestProjectMergeKeyChainIsRejected(t *testing.T) {
	// Each anchor merges the previous one twice, which is 2^60 expansions. The YAML library
	// refuses to expand it, so reading the config must fail right away instead of not finishing.
	var content strings.Builder
	content.WriteString("a0: &a0 {desc: x}\n")
	for i := 1; i <= 60; i++ {
		fmt.Fprintf(&content, "a%d: &a%d {<<: [*a%d, *a%d]}\n", i, i, i-1, i-1)
	}
	content.WriteString("projects:\n  app: {<<: *a60}\n")
	dir := writeTestFiles(t, map[string]string{"mani.yaml": content.String(), "user.yaml": ""})

	done := make(chan error, 1)
	go func() {
		_, err := ReadConfig(filepath.Join(dir, "mani.yaml"), filepath.Join(dir, "user.yaml"), false)
		done <- err
	}()

	select {
	case err := <-done:
		if err == nil || !strings.Contains(err.Error(), "excessive aliasing") || !strings.Contains(err.Error(), filepath.Join(dir, "mani.yaml")+":63") {
			t.Fatalf("expected an aliasing error at the project, got %v", err)
		}
	case <-time.After(10 * time.Second):
		t.Fatal("reading a config with a chain of merge keys didn't finish")
	}
}

func TestProjectEntriesKeepFileOrder(t *testing.T) {
	// The entries of env and remotes are ordered by where their values are in the config, so
	// entries merged from an anchor that's defined earlier come first
	config, _, err := readTestConfig(t, map[string]string{"mani.yaml": `
env_defaults: &env {VERSION: 1.0, MODE: base}
token: &token secret
projects:
  app:
    env:
      ZEBRA: z
      MODE: local
      <<: *env
      TOKEN: *token
      EMPTY:
      ALPHA: a
    remotes: {zulu: z, alpha: a}
`})
	if err != nil {
		t.Fatal(err)
	}
	app := config.ProjectList[0]
	if !slices.Equal(app.EnvList, []string{"VERSION=1.0", "ZEBRA=z", "MODE=local", "TOKEN=secret", "EMPTY=", "ALPHA=a"}) {
		t.Fatalf("wrong env order: %v", app.EnvList)
	}
	if len(app.RemoteList) != 2 || app.RemoteList[0].Name != "zulu" || app.RemoteList[1].Name != "alpha" {
		t.Fatalf("wrong remote order: %v", app.RemoteList)
	}
}

func TestProjectErrorsPointToTheProject(t *testing.T) {
	// A project is reported at the line of its name, also when its definition starts on the next line
	_, dir, err := readTestConfig(t, map[string]string{
		"mani.yaml":   "import: [a/mani.yaml, b/mani.yaml]\n",
		"a/mani.yaml": "projects:\n  first:\n    desc: x\n  app:\n    desc: a\n",
		"b/mani.yaml": "projects:\n  app:\n    tags: [b]\n",
	})
	if err == nil || !strings.Contains(err.Error(), "--> "+filepath.Join(dir, "b/mani.yaml")+":2\n") || !strings.Contains(err.Error(), "in "+filepath.Join(dir, "a/mani.yaml")+":4,") {
		t.Fatalf("expected the ambiguous path to point to the names of the projects, got %v", err)
	}
}

func TestYAMLTypeErrorsKeepTheirLine(t *testing.T) {
	invalid := "projects:\n  app:\n    desc: x\nsync_remotes: notbool\n"
	for _, test := range []struct {
		name   string
		files  map[string]string
		source string
	}{
		{"root config", map[string]string{"mani.yaml": invalid}, "mani.yaml:4"},
		{"imported config", map[string]string{"mani.yaml": "import: [shared.yaml]\n", "shared.yaml": invalid}, "shared.yaml:4"},
		{"project field", map[string]string{"mani.yaml": "projects:\n  app:\n    desc: x\n    tags: {bad: value}\n"}, "mani.yaml:4"},
	} {
		t.Run(test.name, func(t *testing.T) {
			_, dir, err := readTestConfig(t, test.files)
			if err == nil || !strings.Contains(err.Error(), "cannot unmarshal") {
				t.Fatalf("expected a type error, got %v", err)
			}
			if !strings.Contains(err.Error(), "--> "+filepath.Join(dir, test.source)+"\n") || strings.Contains(err.Error(), "yaml: unmarshal errors") {
				t.Fatalf("expected the error to point to %s, got %v", test.source, err)
			}
		})
	}
}

func TestProjectUserConfigPathIsCleaned(t *testing.T) {
	t.Run("user config is a fallback", func(t *testing.T) {
		dir := writeTestFiles(t, map[string]string{
			"ws/mani.yaml":   "projects: {app: {desc: workspace, url: repo}}\n",
			"home/user.yaml": "projects: {app: {desc: user, tags: [personal]}}\n",
		})
		// Not filepath.Join, which cleans the path
		config, err := ReadConfig(filepath.Join(dir, "ws", "mani.yaml"), dir+"/ws/../home/user.yaml", false)
		if err != nil {
			t.Fatal(err)
		}
		if p := config.ProjectList[0]; p.Desc != "workspace" || !slices.Equal(p.Tags, []string{"personal"}) {
			t.Fatalf("user config should supply fallback values: %+v", p)
		}
	})

	t.Run("user config the workspace imports is read once", func(t *testing.T) {
		dir := writeTestFiles(t, map[string]string{
			"ws/mani.yaml":   "import: [../home/user.yaml]\n",
			"home/user.yaml": "projects:\n  solo:\n    env:\n      TOKEN: $(echo ran >> {dir}/ran; echo value)\n",
		})
		config, err := ReadConfig(filepath.Join(dir, "ws", "mani.yaml"), dir+"/ws/../home/user.yaml", false)
		if err != nil {
			t.Fatal(err)
		}
		ran, err := os.ReadFile(filepath.Join(dir, "ran"))
		if err != nil {
			t.Fatal(err)
		}
		if len(config.ProjectList) != 1 || string(ran) != "ran\n" {
			t.Fatalf("expected one project whose env ran once, got %d projects and %q", len(config.ProjectList), ran)
		}
	})
}
