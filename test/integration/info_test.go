package integration

import (
	"testing"
)

var infoTests = []TemplateTest {
	{
		TestName: "Print info",
		InputFiles: []string { "mani-advanced/mani.yaml" },
		TestCmd: "$MANI info",
		Golden: "info/simple",
		WantErr: false,
	},

	{
		TestName: "Print info when specifying config file",
		InputFiles: []string { "mani-advanced/mani.yaml" },
		TestCmd: "$MANI info -c ./mani.yaml",
		Golden: "info/config",
		WantErr: false,
	},

	{
		TestName: "Print error when no config file found",
		InputFiles: []string {},
		TestCmd: "cd /tmp && $MANI info",
		Golden: "info/no-config",
		WantErr: true,
	},
}

func TestInfoCmd(t *testing.T) {
	for _, tt := range infoTests {
		t.Run(tt.TestName, func (t *testing.T) {
			Run(t, tt)
		})
	}
}
