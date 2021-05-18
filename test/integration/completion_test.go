package integration

import (
	"testing"
)

var completionTests = []TemplateTest {
	{
		TestName: "Print bash completion",
		InputFiles: []string { "" },
		TestCmd: "$MANI completion sh",
		Golden: "completion/sh",
		WantErr: false,
	},

	// {
	// 	TestName: "Print zsh completion",
	// 	InputFiles: []string { "" },
	// 	TestCmd: "$MANI completion zsh",
	// 	Golden: "completion/zsh",
	// 	WantErr: false,
	// },

	// {
	// 	TestName: "Print powershell completion",
	// 	InputFiles: []string { "" },
	// 	TestCmd: "$MANI completion psh",
	// 	Golden: "completion/psh",
	// 	WantErr: false,
	// },
}

func TestCompletionCmd(t *testing.T) {
	for _, tt := range completionTests {
		t.Run(tt.TestName, func (t *testing.T) {
			Run(t, tt)
		})
	}
}
