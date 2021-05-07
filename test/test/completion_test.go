package test

import (
	"testing"
)

var completionTests = []TemplateTest {
	{ "", "mani.yaml", "completion/bash.golden", false, "completion sh" },
	{ "", "mani.yaml", "completion/zsh.golden", false, "completion zsh" },
	{ "", "mani.yaml", "completion/powershell.golden", false, "completion powershell" },
}

func TestCompletion(t *testing.T) {
	for _, tt := range completionTests {
		t.Run(tt.Tmpl, func (t *testing.T) {
			Run(t, tt)
		})
	}
}
