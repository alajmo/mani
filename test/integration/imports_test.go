package integration

import (
	"fmt"
	"testing"
)

func TestImports(t *testing.T) {
	var cases = []TemplateTest{
		// Projects
		{
			TestName:   "List projects combined from imported configs",
			InputFiles: []string{"mani-imports/mani.yaml", "mani-imports/shared.yaml"},
			TestCmd:    "mani list projects --output table --headers project,relpath,tags,worktree",
			WantErr:    false,
		},
		{
			TestName:   "Describe projects combined from imported configs",
			InputFiles: []string{"mani-imports/mani.yaml", "mani-imports/shared.yaml"},
			TestCmd:    "mani describe projects",
			WantErr:    false,
		},
		{
			TestName:   "List projects tree with project_dir",
			InputFiles: []string{"mani-imports/mani.yaml", "mani-imports/shared.yaml"},
			TestCmd:    "mani list projects --tree",
			WantErr:    false,
		},
		{
			TestName:   "List projects matching a path below project_dir",
			InputFiles: []string{"mani-imports/mani.yaml", "mani-imports/shared.yaml"},
			TestCmd:    "mani list projects --paths repos",
			WantErr:    false,
		},
		{
			TestName:   "Describe project combined with user config",
			InputFiles: []string{"mani-imports/mani.yaml", "mani-imports/shared.yaml", "mani-imports/user.yaml"},
			TestCmd:    "mani --user-config user.yaml describe projects workspace",
			WantErr:    false,
		},
		{
			TestName:   "Describe project with fields overridden by importing config",
			InputFiles: []string{"mani-imports/override.yaml", "mani-imports/shared.yaml"},
			TestCmd:    "mani --config override.yaml describe projects api",
			WantErr:    false,
		},

		// Tasks
		{
			TestName:   "Describe tasks combined from imported configs",
			InputFiles: []string{"mani-imports/mani.yaml", "mani-imports/shared.yaml"},
			TestCmd:    "mani describe tasks",
			WantErr:    false,
		},
		{
			TestName:   "Run task with env combined from imported configs",
			InputFiles: []string{"mani-imports/mani.yaml", "mani-imports/shared.yaml"},
			TestCmd:    "mani run info --projects workspace",
			WantErr:    false,
		},
		{
			TestName:   "Run task with cmd overridden by importing config",
			InputFiles: []string{"mani-imports/mani.yaml", "mani-imports/shared.yaml"},
			TestCmd:    "mani run build --projects workspace",
			WantErr:    false,
		},

		// Errors
		{
			TestName:   "Throw error on conflicts between configs that don't import each other",
			InputFiles: []string{"mani-imports/siblings.yaml", "mani-imports/sibling-a.yaml", "mani-imports/sibling-b.yaml"},
			TestCmd:    "mani --config siblings.yaml list projects",
			WantErr:    true,
		},
		{
			TestName:   "Throw error when a project is defined twice in the same config",
			InputFiles: []string{"mani-imports/duplicate.yaml"},
			TestCmd:    "mani --config duplicate.yaml list projects",
			WantErr:    true,
		},
	}

	for i, tt := range cases {
		cases[i].Golden = fmt.Sprintf("imports/golden-%d", i)
		cases[i].Index = i

		t.Run(tt.TestName, func(t *testing.T) {
			Run(t, cases[i])
		})
	}
}
