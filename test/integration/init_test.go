package integration

import (
	"testing"
)

// NOTE:
// Check existence of .gitignore file
// Check existence of directories when auto-discovery=true
// Check non-existence of directories when auto-discovery=false
var initTests = []TemplateTest {
	{
		TestName: "Initialize mani in empty directory",
		InputFiles: []string { },
		TestCmd: "$MANI init",
		Golden: "init/empty",
		WantErr: false,
	},

	{
		TestName: "Initialize mani with existing git directories and auto-discovery",
		InputFiles: []string { },
		TestCmd: `
			(mkdir pinto && cd pinto && git init && git remote add origin https://github.com/alajmo/pinto);
			(mkdir -p nested/template-generator && cd nested/template-generator && git init && git remote add origin https://github.com/alajmo/template-generator);
			(mkdir nameless);
			$MANI init
		`,
		Golden: "init/auto-discovery",
		WantErr: false,
	},

	{
		TestName: "Throw error when initialize in existing mani directory",
		InputFiles: []string { "mani-advanced/mani.yaml" },
		TestCmd: "$MANI init",
		Golden: "init/existing",
		WantErr: true,
	},
}

func TestInitCmd(t *testing.T) {
	for _, tt := range initTests {
		t.Run(tt.TestName, func (t *testing.T) {
			Run(t, tt)
		})
	}
}
