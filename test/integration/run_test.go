package integration

import (
	"testing"
)

var runTests = []TemplateTest{
	{
		TestName:   "Should fail to run when no configuration file found",
		InputFiles: []string{},
		TestCmd: `
			$MANI run pwd -a
		`,
		Golden:  "run/no-config",
		WantErr: true,
	},

	{
		TestName:   "Should run in zero projects",
		InputFiles: []string{"mani-advanced/mani.yaml", "mani-advanced/.gitignore"},
		TestCmd: `
			$MANI sync
			$MANI run pwd
		`,
		Golden:  "run/zero",
		WantErr: false,
	},

	{
		TestName:   "Should run in all projects",
		InputFiles: []string{"mani-advanced/mani.yaml", "mani-advanced/.gitignore"},
		TestCmd: `
			$MANI sync
			$MANI run -a pwd
		`,
		Golden:  "run/all",
		WantErr: false,
	},

	{
		TestName:   "Should run when filtered on project name",
		InputFiles: []string{"mani-advanced/mani.yaml", "mani-advanced/.gitignore"},
		TestCmd: `
			$MANI sync
			$MANI run --projects pinto pwd
		`,
		Golden:  "run/filter-on-1-project",
		WantErr: false,
	},

	{
		TestName:   "Should run when filtered on tags",
		InputFiles: []string{"mani-advanced/mani.yaml", "mani-advanced/.gitignore"},
		TestCmd: `
			$MANI sync
			$MANI run --tags frontend pwd
		`,
		Golden:  "run/filter-on-1-tag",
		WantErr: false,
	},

	{
		TestName:   "Should run when filtered on cwd",
		InputFiles: []string{"mani-advanced/mani.yaml", "mani-advanced/.gitignore"},
		TestCmd: `
			$MANI sync
			cd template-generator
			$MANI run --cwd pwd
		`,
		Golden:  "run/filter-on-cwd",
		WantErr: false,
	},

	{
		TestName:   "Should dry run run",
		InputFiles: []string{"mani-advanced/mani.yaml", "mani-advanced/.gitignore"},
		TestCmd: `
			$MANI sync
			$MANI run --dry-run --projects template-generator pwd
		`,
		Golden:  "run/dry-run",
		WantErr: false,
	},
}

func TestRunCmd(t *testing.T) {
	for _, tt := range runTests {
		t.Run(tt.TestName, func(t *testing.T) {
			Run(t, tt)
		})
	}
}
