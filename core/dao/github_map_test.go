package dao

import (
	"reflect"
	"strings"
	"testing"

	"github.com/alajmo/mani/core/provider/github"
)

func TestProjectFromGitHubRepo_AllTagsOn(t *testing.T) {
	r := github.Repo{
		Name:        "alpha",
		SSHURL:      "git@github.com:acme/alpha.git",
		Description: "the alpha repo",
		Visibility:  "PUBLIC",
		Archived:    false,
		Fork:        false,
		Topics:      []string{"go", "cli"},
	}
	p := ProjectFromGitHubRepo(r, github.MapOptions{})
	if p.Name != "alpha" || p.Path != "alpha" {
		t.Fatalf("name/path mismatch: %+v", p)
	}
	if p.URL != "git@github.com:acme/alpha.git" {
		t.Fatalf("url mismatch: %s", p.URL)
	}
	if p.Desc != "the alpha repo" {
		t.Fatalf("desc mismatch: %s", p.Desc)
	}
	// Boolean state is presence-only; visibility flattened; topics bare;
	// default branch not emitted.
	want := []string{"public", "go", "cli"}
	if !reflect.DeepEqual(p.Tags, want) {
		t.Fatalf("tags mismatch:\n got: %v\nwant: %v", p.Tags, want)
	}
}

func TestProjectFromGitHubRepo_BooleanFlagsArePresenceOnly(t *testing.T) {
	r := github.Repo{
		Name:       "alpha",
		Archived:   true,
		Fork:       true,
		Visibility: "private",
	}
	p := ProjectFromGitHubRepo(r, github.MapOptions{})
	want := []string{"archived", "fork", "private"}
	if !reflect.DeepEqual(p.Tags, want) {
		t.Fatalf("tags mismatch:\n got: %v\nwant: %v", p.Tags, want)
	}
}

func TestProjectFromGitHubRepo_DedupesTopicCollidingWithSynth(t *testing.T) {
	// Repo is archived AND has a topic literally named "archived" — the
	// tag must appear only once.
	r := github.Repo{
		Name:     "alpha",
		Archived: true,
		Topics:   []string{"archived", "go", "go"}, // also dedup repeated topic
	}
	p := ProjectFromGitHubRepo(r, github.MapOptions{})
	want := []string{"archived", "go"}
	if !reflect.DeepEqual(p.Tags, want) {
		t.Fatalf("tags mismatch:\n got: %v\nwant: %v", p.Tags, want)
	}
}

func TestProjectFromGitHubRepo_NoTags(t *testing.T) {
	r := github.Repo{Name: "a", Visibility: "public", Topics: []string{"x"}}
	p := ProjectFromGitHubRepo(r, github.MapOptions{NoTags: true})
	if p.Tags != nil {
		t.Fatalf("expected no tags, got %v", p.Tags)
	}
}

func TestProjectFromGitHubRepo_NoTopicTags(t *testing.T) {
	r := github.Repo{Name: "a", Visibility: "public", Topics: []string{"x", "y"}}
	p := ProjectFromGitHubRepo(r, github.MapOptions{NoTopicTags: true})
	for _, tag := range p.Tags {
		if tag == "x" || tag == "y" {
			t.Fatalf("unexpected topic in tags: %s", tag)
		}
	}
	if len(p.Tags) == 0 {
		t.Fatalf("expected non-topic tags, got none")
	}
}

func TestProjectFromGitHubRepo_VisibilityLowercased(t *testing.T) {
	p := ProjectFromGitHubRepo(github.Repo{Name: "a", Visibility: "Internal"}, github.MapOptions{})
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

func TestProjectFromGitHubRepo_DefaultBranchNotEmittedAsTag(t *testing.T) {
	// We intentionally do not emit defaultBranch as a tag — it's not
	// useful to filter on and would be the lone key=value tag in the set.
	p := ProjectFromGitHubRepo(github.Repo{Name: "a", DefaultBranch: "main"}, github.MapOptions{})
	for _, tag := range p.Tags {
		if strings.HasPrefix(tag, "defaultBranch") {
			t.Fatalf("unexpected defaultBranch tag: %s", tag)
		}
	}
}
