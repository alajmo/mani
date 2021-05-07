package test

import (
	"testing"
)

// TODO:
// Check existence of .gitignore file when init success
// Check existence of directories when auto-discovery=true
// Check non-existence of directories when auto-discovery=false

var initTests = []TemplateTest {
	{ "", "Simple init", "init/simple.golden", false, "init" },
	{ "mani.yaml", "Existing init", "init/simple-existing-mani.golden", false, "init" },
}

func TestInit(t *testing.T) {
	for _, tt := range initTests {
		t.Run(tt.Tmpl, func (t *testing.T) {
			Run(t, tt)
		})
	}
}

