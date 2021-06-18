package integration

import (
	"testing"
)

var runTests = []TemplateTest{
	{
		TestName:   "Should fail to run when no configuration file found",
		InputFiles: []string{},
		TestCmd: `
			mani run pwd -a
		`,
		Golden:  "run/no-config",
		WantErr: true,
	},

	{
		TestName:   "Should run in zero projects",
		InputFiles: []string{"mani-advanced/mani.yaml", "mani-advanced/.gitignore"},
		TestCmd: `
			mani sync
			mani run pwd
		`,
		Golden:  "run/zero-projects",
		WantErr: false,
	},

	{
		TestName:   "Should run in all projects",
		InputFiles: []string{"mani-advanced/mani.yaml", "mani-advanced/.gitignore"},
		TestCmd: `
			mani sync
			mani run -a pwd
		`,
		Golden:  "run/all-projects",
		WantErr: false,
	},

	{
		TestName:   "Should run when filtered on project",
		InputFiles: []string{"mani-advanced/mani.yaml", "mani-advanced/.gitignore"},
		TestCmd: `
			mani sync
			mani run --projects pinto pwd
		`,
		Golden:  "run/1-project-flag-projects",
		WantErr: false,
	},

	{
		TestName:   "Should run when filtered on tags",
		InputFiles: []string{"mani-advanced/mani.yaml", "mani-advanced/.gitignore"},
		TestCmd: `
			mani sync
			mani run --tags frontend pwd
		`,
		Golden:  "run/1-project-flag-tags",
		WantErr: false,
	},

	{
		TestName:   "Should run when filtered on cwd",
		InputFiles: []string{"mani-advanced/mani.yaml", "mani-advanced/.gitignore"},
		TestCmd: `
			mani sync
			cd template-generator
			mani run --cwd pwd
		`,
		Golden:  "run/1-project-flag-cwd",
		WantErr: false,
	},

	{
		TestName:   "Should run on default tags",
		InputFiles: []string{"mani-advanced/mani.yaml", "mani-advanced/.gitignore"},
		TestCmd: `
			mani sync
			mani run default-tags
		`,
		Golden:  "run/filter-default-tags",
		WantErr: false,
	},

	{
		TestName:   "Should run on default projects",
		InputFiles: []string{"mani-advanced/mani.yaml", "mani-advanced/.gitignore"},
		TestCmd: `
			mani sync
			mani run default-projects
		`,
		Golden:  "run/filter-default-projects",
		WantErr: false,
	},

	{
		TestName:   "Should print table when output set to table in task",
		InputFiles: []string{"mani-advanced/mani.yaml", "mani-advanced/.gitignore"},
		TestCmd: `
			mani sync
			mani run default-output -p dashgrid
		`,
		Golden:  "run/default-output",
		WantErr: false,
	},

	{
		TestName:   "Should dry run",
		InputFiles: []string{"mani-advanced/mani.yaml", "mani-advanced/.gitignore"},
		TestCmd: `
			mani sync
			mani run --dry-run --projects template-generator pwd
		`,
		Golden:  "run/dry-run",
		WantErr: false,
	},

	{
		TestName:   "Should run multiple commands",
		InputFiles: []string{"mani-advanced/mani.yaml", "mani-advanced/.gitignore"},
		TestCmd: `
			mani sync
			mani run pwd multi -a
		`,
		Golden:  "run/multiple-commands",
		WantErr: false,
	},

	{
		TestName:   "Should run sub-commands",
		InputFiles: []string{"mani-advanced/mani.yaml", "mani-advanced/.gitignore"},
		TestCmd: `
			mani sync
			mani run submarine -a
		`,
		Golden:  "run/sub-commands",
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
