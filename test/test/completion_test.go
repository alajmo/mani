package test

import (
	"testing"
)

var completionTests = []TemplateTest {
	{ "", "Print bash completion", "completion/bash.golden", false, "completion sh" },
	{ "", "Print zsh completion", "completion/zsh.golden", false, "completion zsh" },
	{ "", "Print powershell completion", "completion/powershell.golden", false, "completion powershell" },
}

func TestCompletion(t *testing.T) {
	for _, tt := range completionTests {
		t.Run(tt.Tmpl, func (t *testing.T) {
			Run(t, tt)
		})
	}
}
