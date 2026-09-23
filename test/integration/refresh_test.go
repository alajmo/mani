package integration

import (
	"fmt"
	"testing"
)

// TestRefresh exercises `mani refresh` end-to-end. The `gh`/`tea` provider
// CLIs are stubbed (see test/scripts/gh and test/scripts/tea), which serve
// a fixed upstream repo set for the org `acme`; each case varies the input
// mani.yaml to drive a different add / prune / no-op path. The golden
// mani.yaml proves the lossless edit: comments, blank lines, 2-space
// indentation, and hand-customized fields outside the touched entries
// survive byte-for-byte.
func TestRefresh(t *testing.T) {
	var cases = []TemplateTest{
		{
			TestName:   "Add new upstream repos, preserving comments and formatting",
			InputFiles: []string{"mani-refresh-add/mani.yaml"},
			TestCmd:    "mani refresh --color=false",
			WantErr:    false,
		},

		{
			TestName:   "No changes when all upstream repos are already present",
			InputFiles: []string{"mani-refresh-uptodate/mani.yaml"},
			TestCmd:    "mani refresh --color=false",
			WantErr:    false,
		},

		{
			TestName:   "Dry-run reports additions without modifying mani.yaml",
			InputFiles: []string{"mani-refresh-add/mani.yaml"},
			TestCmd:    "mani refresh --dry-run --color=false",
			WantErr:    false,
		},

		{
			TestName:   "Prune removes upstream-gone projects scoped to the source remote",
			InputFiles: []string{"mani-refresh-prune/mani.yaml"},
			TestCmd:    "mani refresh --prune --color=false",
			WantErr:    false,
		},

		{
			TestName:   "Stale upstream-gone projects are reported but kept without --prune",
			InputFiles: []string{"mani-refresh-prune/mani.yaml"},
			TestCmd:    "mani refresh --color=false",
			WantErr:    false,
		},

		{
			TestName:   "Error when no sources block and auto-discovery disabled",
			InputFiles: []string{"mani-refresh-nosources/mani.yaml"},
			TestCmd:    "mani refresh --auto-discovery=false --color=false",
			WantErr:    true,
		},

		{
			TestName:   "Add repos from a Gitea source into an empty projects block",
			InputFiles: []string{"mani-refresh-gitea/mani.yaml"},
			TestCmd:    "mani refresh --color=false",
			WantErr:    false,
		},

		{
			// A sources-less config (plain `mani init`) still picks up repos
			// cloned into the workspace by hand. The non-git directory and the
			// already-listed project must be left alone.
			TestName:   "Discover locally-cloned repos into a sources-less config",
			InputFiles: []string{"mani-refresh-discover/mani.yaml"},
			TestCmd: `
			(mkdir -p newrepo && cd newrepo && git init -b main && git remote add origin https://github.com/acme/newrepo);
			(mkdir -p nested/lib && cd nested/lib && git init -b main && git remote add origin https://github.com/acme/lib);
			(mkdir -p notarepo && touch notarepo/empty);
			mani refresh --color=false
			`,
			WantErr: false,
		},
	}

	for i, tt := range cases {
		cases[i].Golden = fmt.Sprintf("refresh/golden-%d", i)
		cases[i].Index = i
		t.Run(tt.TestName, func(t *testing.T) {
			Run(t, cases[i])
		})
	}
}
