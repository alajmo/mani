package integration

import (
	"testing"
)

var execTests = []TemplateTest{
	{
		TestName:   "Should fail to exec when no configuration file found",
		InputFiles: []string{},
		TestCmd: `
			mani exec --all -o table ls
		`,
		Golden:  "exec/no-config",
		WantErr: true,
	},

	{
		TestName:   "Should exec in zero projects",
		InputFiles: []string{"mani-advanced/mani.yaml", "mani-advanced/.gitignore"},
		TestCmd: `
			mani sync --parallel=false
			mani exec -o table ls
		`,
		Golden:  "exec/zero",
		WantErr: false,
	},

	{
		TestName:   "Should exec in all projects",
		InputFiles: []string{"mani-advanced/mani.yaml", "mani-advanced/.gitignore"},
		TestCmd: `
			mani sync --parallel=false
			mani exec --all -o table ls
		`,
		Golden:  "exec/all",
		WantErr: false,
	},

	{
		TestName:   "Should exec when filtered on project name",
		InputFiles: []string{"mani-advanced/mani.yaml", "mani-advanced/.gitignore"},
		TestCmd: `
			mani sync --parallel=false
			mani exec -o table --projects pinto ls
		`,
		Golden:  "exec/filter-on-1-project",
		WantErr: false,
	},

	{
		TestName:   "Should exec when filtered on tags",
		InputFiles: []string{"mani-advanced/mani.yaml", "mani-advanced/.gitignore"},
		TestCmd: `
			mani sync --parallel=false
			mani exec -o table --tags frontend ls
		`,
		Golden:  "exec/filter-on-1-tag",
		WantErr: false,
	},

	{
		TestName:   "Should exec when filtered on cwd",
		InputFiles: []string{"mani-advanced/mani.yaml", "mani-advanced/.gitignore"},
		TestCmd: `
			mani sync --parallel=false
			cd template-generator
			mani exec -o table --cwd pwd
		`,
		Golden:  "exec/filter-on-cwd",
		WantErr: false,
	},

	{
		TestName:   "Should dry run exec",
		InputFiles: []string{"mani-advanced/mani.yaml", "mani-advanced/.gitignore"},
		TestCmd: `
			mani sync --parallel=false
			mani exec -o table --dry-run --projects template-generator pwd
		`,
		Golden:  "exec/dry-run",
		WantErr: false,
	},
}

func TestExecCmd(t *testing.T) {
	for _, tt := range execTests {
		t.Run(tt.TestName, func(t *testing.T) {
			Run(t, tt)
		})
	}
}
