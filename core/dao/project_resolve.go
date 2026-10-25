package dao

import (
	"fmt"
	"path/filepath"

	"github.com/alajmo/mani/core"
)

// Stage 3 resolves surviving paths and applies defaults. Evaluation is a
// separate final pass so an invalid config never executes project env commands.
func resolveProjects(definitions []projectDefinition, rootDir string) ([]Project, []ResourceErrors[projectSource]) {
	var projects []Project
	var problems []ResourceErrors[projectSource]
	for _, def := range definitions {
		project := Project{Name: def.name, Desc: def.desc.value, URL: def.url.value,
			Clone: def.clone.value, Branch: def.branch.value, Tags: def.tags.value,
			context: def.source.file, contextLine: def.source.line,
			Sync: core.Ptr(true), SingleBranch: core.Ptr(false)}
		var resolved []projectValue[string]
		for _, candidate := range def.pathCandidates {
			base, err := core.GetAbsolutePath(filepath.Dir(candidate.source.file), candidate.source.projectDir, "")
			if err == nil {
				candidate.value, err = core.GetAbsolutePath(base, candidate.value, def.name)
			}
			if err != nil {
				problems = append(problems, candidate.source.problem(err))
				continue
			}
			candidate.value = filepath.Clean(candidate.value)
			resolved = append(resolved, candidate)
		}
		if len(resolved) == 0 {
			continue
		}
		project.Path = resolved[0].value
		for _, path := range resolved[1:] {
			if path.value == project.Path {
				continue
			}
			if def.path.set {
				problems = append(problems, projectConflict(def.name, "`path`", path, resolved[0]))
			} else {
				problems = append(problems, path.source.problem(fmt.Errorf("project `%s` resolves to `%s` here and `%s` in %s, set `path` in one of the definitions", def.name, path.value, project.Path, resolved[0].source.location())))
			}
		}
		var err error
		project.RelPath, err = core.GetRelativePath(rootDir, project.Path)
		if err != nil {
			problems = append(problems, def.source.problem(err))
		}
		if def.sync.set {
			project.Sync = core.Ptr(def.sync.value)
		}
		if def.singleBranch.set {
			project.SingleBranch = core.Ptr(def.singleBranch.value)
		}
		for _, remote := range def.remotes.value {
			project.RemoteList = append(project.RemoteList, Remote{Name: remote.key, URL: remote.value})
		}
		for _, entry := range def.worktrees.value {
			worktree := entry.value
			if worktree.Branch == "" {
				worktree.Branch = filepath.Base(worktree.Path)
			}
			project.WorktreeList = append(project.WorktreeList, worktree)
		}
		projects = append(projects, project)
	}
	return projects, problems
}

func evaluateProjectEnv(projects []Project, definitions []projectDefinition) []ResourceErrors[projectSource] {
	var problems []ResourceErrors[projectSource]
	for i, def := range definitions {
		for _, env := range def.env.value {
			values, err := EvaluateEnv([]string{env.key + "=" + env.value})
			if err != nil {
				problems = append(problems, env.source.problem(err))
				continue
			}
			projects[i].EnvList = append(projects[i].EnvList, values[0])
		}
	}
	return problems
}
