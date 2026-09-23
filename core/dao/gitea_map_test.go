package dao

import (
	"reflect"
	"testing"

	"github.com/alajmo/mani/core/provider/gitea"
)

func TestProjectFromGiteaRepo_PublicRepoTags(t *testing.T) {
	r := gitea.Repo{
		Name:        "alpha",
		SSHURL:      "git@gitea.com:acme/alpha.git",
		Description: "the alpha repo",
		Topics:      []string{"go", "cli"},
	}
	p := ProjectFromGiteaRepo(r, gitea.MapOptions{})
	if p.Name != "alpha" || p.Path != "alpha" || p.URL != "git@gitea.com:acme/alpha.git" {
		t.Fatalf("project mismatch: %+v", p)
	}
	want := []string{"public", "go", "cli"}
	if !reflect.DeepEqual(p.Tags, want) {
		t.Fatalf("tags: got %v want %v", p.Tags, want)
	}
}

func TestProjectFromGiteaRepo_ArchivedForkPrivate(t *testing.T) {
	r := gitea.Repo{Name: "a", Archived: true, Fork: true, Private: true}
	p := ProjectFromGiteaRepo(r, gitea.MapOptions{})
	want := []string{"archived", "fork", "private"}
	if !reflect.DeepEqual(p.Tags, want) {
		t.Fatalf("tags: got %v want %v", p.Tags, want)
	}
}

func TestProjectFromGiteaRepo_Internal(t *testing.T) {
	p := ProjectFromGiteaRepo(gitea.Repo{Name: "a", Internal: true}, gitea.MapOptions{})
	found := false
	for _, tag := range p.Tags {
		if tag == "internal" {
			found = true
		}
	}
	if !found {
		t.Fatalf("expected `internal` in tags, got %v", p.Tags)
	}
}

func TestProjectFromGiteaRepo_DedupesTopicCollidingWithSynth(t *testing.T) {
	r := gitea.Repo{Name: "a", Archived: true, Topics: []string{"archived", "go"}}
	p := ProjectFromGiteaRepo(r, gitea.MapOptions{})
	want := []string{"archived", "public", "go"}
	if !reflect.DeepEqual(p.Tags, want) {
		t.Fatalf("tags: got %v want %v", p.Tags, want)
	}
}

func TestProjectFromGiteaRepo_NoTags(t *testing.T) {
	r := gitea.Repo{Name: "a", Topics: []string{"x"}}
	p := ProjectFromGiteaRepo(r, gitea.MapOptions{NoTags: true})
	if p.Tags != nil {
		t.Fatalf("expected nil tags, got %v", p.Tags)
	}
}
