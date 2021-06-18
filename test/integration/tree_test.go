package integration

import (
	"testing"
)

var treeTests = []TemplateTest{
	{
		TestName:   "List empty tree",
		InputFiles: []string{"mani-empty/mani.yaml"},
		TestCmd:    "mani tree projects",
		Golden:     "tree/empty-tree",
		WantErr:    false,
	},
	{
		TestName:   "List full tree",
		InputFiles: []string{"mani-advanced/mani.yaml"},
		TestCmd:    "mani tree",
		Golden:     "tree/full-tree",
		WantErr:    false,
	},
	{
		TestName:   "List tree filtered on tag",
		InputFiles: []string{"mani-advanced/mani.yaml"},
		TestCmd:    "mani tree --tags frontend",
		Golden:     "tree/tags-tree",
		WantErr:    false,
	},
}

func TestTreeCmd(t *testing.T) {
	for _, tt := range treeTests {
		t.Run(tt.TestName, func(t *testing.T) {
			Run(t, tt)
		})
	}
}
