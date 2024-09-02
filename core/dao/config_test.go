package dao

import (
	"testing"
)

func TestConfig_DuplicateProjectName(t *testing.T) {
	originalProjects := []Project{
		{Name: "project-a", Path: "sub-1/project-a"},
		{Name: "project-a", Path: "sub-2/project-a"},
		{Name: "project-b", Path: "sub-3/project-b"},
	}

	var projects []Project
	projects = append(projects, originalProjects...)
	RenameDuplicates(projects)

	if projects[0].Name != originalProjects[0].Path {
		t.Fatalf(`Wanted: %q, Found: %q`, projects[0].Path, originalProjects[0].Name)
	}

	if projects[1].Name != originalProjects[1].Path {
		t.Fatalf(`Wanted: %q, Found: %q`, projects[1].Path, originalProjects[1].Name)
	}

	if originalProjects[2].Name != projects[2].Name {
		t.Fatalf(`Wanted: %q, Found: %q`, projects[2].Name, originalProjects[2].Name)
	}
}
