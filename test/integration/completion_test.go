package integration

import (
	"testing"
)

var completionTests = []TemplateTest{
	{
		TestName:   "Print bash completion",
		InputFiles: []string{},
		TestCmd:    "mani completion bash",
		Golden:     "completion/bash",
		WantErr:    false,
	},

	{
		TestName:   "Print zsh completion",
		InputFiles: []string{},
		TestCmd:    "mani completion zsh",
		Golden:     "completion/zsh",
		WantErr:    false,
	},

	{
		TestName:   "Print fish completion",
		InputFiles: []string{},
		TestCmd:    "mani completion fish",
		Golden:     "completion/fish",
		WantErr:    false,
	},

	{
		TestName:   "Print powershell completion",
		InputFiles: []string{},
		TestCmd:    "mani completion powershell",
		Golden:     "completion/powershell",
		WantErr:    false,
	},
}

func TestCompletionCmd(t *testing.T) {
	for _, tt := range completionTests {
		t.Run(tt.TestName, func(t *testing.T) {
			Run(t, tt)
		})
	}
}
