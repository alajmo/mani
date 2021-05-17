package test

import (
	"testing"
)

var infoTests = []TemplateTest {
	{
		TestName: "Print info",
		InputFiles: []string { "mani.yaml" },
		TestCmd: "info",
		Golden: "info/simple",
	},

	{
		TestName: "Print info when specifying config file",
		InputFiles: []string { "mani.yaml" },
		TestCmd: "info -c ./mani.yaml",
		Golden: "info/config",
	},

	{
		TestName: "Print error when no config file found",
		InputFiles: []string { "" },
		BootstrapCmds: []string { "cd /tmp" },
		TestCmd: "info",
		Golden: "info/no-config",
	},
}

func TestInfoCmd(t *testing.T) {
	for _, tt := range infoTests {
		t.Run(tt.TestName, func (t *testing.T) {
			Run(t, tt)
		})
	}
}
