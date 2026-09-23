package integration

import (
	"fmt"
	"testing"
)

// TestInitSource covers `mani init --source <provider>/<name>`, which
// fetches repositories from a remote provider (via the stubbed gh/tea CLIs)
// and scaffolds mani.yaml with both a `sources:` block and the discovered
// `projects:`. --sync-gitignore=false keeps the output to a single
// mani.yaml so the golden is stable.
func TestInitSource(t *testing.T) {
	var cases = []TemplateTest{
		{
			TestName:   "Initialize from a GitHub organization",
			InputFiles: []string{},
			TestCmd:    "mani init --source github-org/acme --sync-gitignore=false --color=false",
			WantErr:    false,
		},

		{
			TestName:   "Initialize from a Gitea organization",
			InputFiles: []string{},
			TestCmd:    "mani init --source gitea-org/acme --sync-gitignore=false --color=false",
			WantErr:    false,
		},

		{
			TestName:   "Error on unsupported source provider",
			InputFiles: []string{},
			TestCmd:    "mani init --source gitlab/acme --color=false",
			WantErr:    true,
		},
	}

	for i, tt := range cases {
		cases[i].Golden = fmt.Sprintf("init-source/golden-%d", i)
		cases[i].Index = i
		t.Run(tt.TestName, func(t *testing.T) {
			Run(t, cases[i])
		})
	}
}
