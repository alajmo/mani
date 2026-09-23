package integration

import (
	"fmt"
	"testing"
)

// TestProject exercises `mani project add` and `mani project rm` end-to-end.
// These commands edit mani.yaml losslessly via the yamledit engine (no
// provider CLI or network needed). The golden mani.yaml proves comments,
// blank lines, 2-space indentation, and hand-customized fields outside the
// touched entry survive byte-for-byte.
func TestProject(t *testing.T) {
	var cases = []TemplateTest{
		{
			TestName:   "Add a remote project, preserving comments and formatting",
			InputFiles: []string{"mani-project-add/mani.yaml"},
			TestCmd:    "mani project add delta --url git@github.com:acme/delta.git --tags backend,go --color=false",
			WantErr:    false,
		},

		{
			TestName:   "Add a project with a custom path and description",
			InputFiles: []string{"mani-project-add/mani.yaml"},
			TestCmd:    `mani project add my-lib --url git@github.com:acme/my-lib.git --path ./libs/my-lib --desc "shared lib" --color=false`,
			WantErr:    false,
		},

		{
			TestName:   "Error when adding a project that already exists",
			InputFiles: []string{"mani-project-add/mani.yaml"},
			TestCmd:    "mani project add alpha --url git@github.com:acme/alpha.git --color=false",
			WantErr:    true,
		},

		{
			TestName:   "Remove a project with --force, taking its head comment",
			InputFiles: []string{"mani-project-remove/mani.yaml"},
			TestCmd:    "mani project rm legacy --force --color=false",
			WantErr:    false,
		},

		{
			TestName:   "Error when removing a project that does not exist",
			InputFiles: []string{"mani-project-remove/mani.yaml"},
			TestCmd:    "mani project rm ghost --force --color=false",
			WantErr:    true,
		},

		{
			TestName:   "Add from a URL, deriving the project name",
			InputFiles: []string{"mani-project-add/mani.yaml"},
			TestCmd:    "mani project add https://github.com/alajmo/mani.git --color=false",
			WantErr:    false,
		},

		{
			TestName:   "Add from a URL with an explicit --name",
			InputFiles: []string{"mani-project-add/mani.yaml"},
			TestCmd:    "mani project add https://github.com/alajmo/koko.git --name kaka --color=false",
			WantErr:    false,
		},

		{
			TestName:   "Error when adding a name without a url",
			InputFiles: []string{"mani-project-add/mani.yaml"},
			TestCmd:    "mani project add localthing --color=false",
			WantErr:    true,
		},
	}

	for i, tt := range cases {
		cases[i].Golden = fmt.Sprintf("project/golden-%d", i)
		cases[i].Index = i
		t.Run(tt.TestName, func(t *testing.T) {
			Run(t, cases[i])
		})
	}
}
