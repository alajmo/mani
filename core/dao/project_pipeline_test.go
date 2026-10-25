package dao

import (
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

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
	if def.path.value != "relative" || def.worktrees.value[0].value.Branch != "" || def.env.value[1].value != "$(exit 1)" {
		t.Fatal("parsing must not resolve paths, apply defaults or evaluate env")
	}
	if def.env.value[0].value != "1.0" || def.env.value[0].source.line != 7 || def.env.value[0].source.file != config.Path || def.path.source.projectDir != "../repos" {
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
