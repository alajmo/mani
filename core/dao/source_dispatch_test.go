package dao

import (
	"testing"

	"github.com/alajmo/mani/core/sources"
)

func TestApplySourcePath(t *testing.T) {
	tests := []struct {
		name     string
		path     string
		in       Project
		wantPath string
	}{
		{
			name:     "no path leaves project flat",
			path:     "",
			in:       Project{Name: "alpha", Path: "alpha"},
			wantPath: "alpha",
		},
		{
			name:     "path prefixes project path",
			path:     "acme",
			in:       Project{Name: "alpha", Path: "alpha"},
			wantPath: "acme/alpha",
		},
		{
			name:     "nested path prefix",
			path:     "team/acme",
			in:       Project{Name: "alpha", Path: "alpha"},
			wantPath: "team/acme/alpha",
		},
		{
			name:     "trailing slash is cleaned",
			path:     "acme/",
			in:       Project{Name: "alpha", Path: "alpha"},
			wantPath: "acme/alpha",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := sources.Source{Source: "github-org/acme", Path: tt.path}
			got := applySourcePath(s, tt.in)
			if got.Path != tt.wantPath {
				t.Errorf("applySourcePath path = %q, want %q", got.Path, tt.wantPath)
			}
			// The project key (Name) must never change — refresh keys its
			// diff on URL/Name, and the YAML mapping key stays the repo name.
			if got.Name != tt.in.Name {
				t.Errorf("applySourcePath mutated Name = %q, want %q", got.Name, tt.in.Name)
			}
		})
	}
}
