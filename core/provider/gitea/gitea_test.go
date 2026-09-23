package gitea

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"strings"
	"testing"
)

func boolPtr(b bool) *bool { return &b }

type fakeRunner struct {
	responses map[string]string
	errors    map[string]error
	calls     [][]string
}

func (f *fakeRunner) run(_ context.Context, args ...string) ([]byte, error) {
	f.calls = append(f.calls, args)
	if len(args) < 2 || args[0] != "api" {
		return nil, errors.New("unexpected args: " + strings.Join(args, " "))
	}
	key := args[1]
	if e, ok := f.errors[key]; ok {
		return nil, e
	}
	if body, ok := f.responses[key]; ok {
		return []byte(body), nil
	}
	return nil, errors.New("no fixture for " + key)
}

func TestList_OrgSinglePage(t *testing.T) {
	f := &fakeRunner{
		responses: map[string]string{
			"/orgs/acme/repos?page=1&limit=50": `[
				{"name":"a","ssh_url":"git@gitea.com:acme/a.git","default_branch":"main"},
				{"name":"b","ssh_url":"git@gitea.com:acme/b.git","default_branch":"main"}
			]`,
			// A short page no longer ends pagination by itself; the server must
			// hand back an empty page to terminate (so an instance whose page
			// cap is below our limit isn't mistaken for "no more repos").
			"/orgs/acme/repos?page=2&limit=50": `[]`,
		},
	}
	got, err := List(context.Background(), ListOptions{Name: "acme"}, f.run)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 2 {
		t.Fatalf("got %d, want 2: %+v", len(got), got)
	}
	// 2 calls: page 1 had repos, page 2 was empty → stop.
	if len(f.calls) != 2 {
		t.Fatalf("expected 2 calls, got %d", len(f.calls))
	}
}

func TestList_OrgPaginates(t *testing.T) {
	// First page is full (50 distinct repos) → keep paging until an empty page.
	page1 := "["
	for i := 0; i < 50; i++ {
		if i > 0 {
			page1 += ","
		}
		page1 += fmt.Sprintf(`{"name":"r%d","full_name":"acme/r%d","ssh_url":"git@gitea.com:acme/r%d.git"}`, i, i, i)
	}
	page1 += "]"

	f := &fakeRunner{
		responses: map[string]string{
			"/orgs/acme/repos?page=1&limit=50": page1,
			"/orgs/acme/repos?page=2&limit=50": `[{"name":"last","full_name":"acme/last","ssh_url":"git@gitea.com:acme/last.git"}]`,
			"/orgs/acme/repos?page=3&limit=50": `[]`,
		},
	}
	got, err := List(context.Background(), ListOptions{Name: "acme"}, f.run)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 51 {
		t.Fatalf("expected 51 repos across pages, got %d", len(got))
	}
	if len(f.calls) != 3 {
		t.Fatalf("expected 3 paginated calls (incl. empty terminator), got %d", len(f.calls))
	}
}

func TestList_OrgStopsWhenPageCapBelowLimit(t *testing.T) {
	// Regression for the under-fetch bug: an instance that caps every page at
	// 2 items (below our requested 50) must NOT be mistaken for "done" after
	// the first short page — every repo across the capped pages is returned.
	f := &fakeRunner{
		responses: map[string]string{
			"/orgs/acme/repos?page=1&limit=50": `[{"name":"a","full_name":"acme/a"},{"name":"b","full_name":"acme/b"}]`,
			"/orgs/acme/repos?page=2&limit=50": `[{"name":"c","full_name":"acme/c"},{"name":"d","full_name":"acme/d"}]`,
			"/orgs/acme/repos?page=3&limit=50": `[]`,
		},
	}
	got, err := List(context.Background(), ListOptions{Name: "acme"}, f.run)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 4 {
		t.Fatalf("expected 4 repos across capped pages, got %d: %+v", len(got), got)
	}
}

func TestList_OrgTerminatesWhenServerIgnoresPaging(t *testing.T) {
	// A server that ignores `page` and returns the same repos forever must
	// still terminate: a page with no NEW repos ends the walk.
	f := &fakeRunner{
		responses: map[string]string{
			"/orgs/acme/repos?page=1&limit=50": `[{"name":"a","full_name":"acme/a"},{"name":"b","full_name":"acme/b"}]`,
			"/orgs/acme/repos?page=2&limit=50": `[{"name":"a","full_name":"acme/a"},{"name":"b","full_name":"acme/b"}]`,
		},
	}
	got, err := List(context.Background(), ListOptions{Name: "acme"}, f.run)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 2 {
		t.Fatalf("expected 2 deduped repos, got %d: %+v", len(got), got)
	}
	if len(f.calls) != 2 {
		t.Fatalf("expected to stop after the repeated page, got %d calls", len(f.calls))
	}
}

func TestList_FallsBackToUserPublicEndpointOn404(t *testing.T) {
	f := &fakeRunner{
		responses: map[string]string{
			"user":                             `{"login":"someone-else"}`,
			"/users/joe/repos?page=1&limit=50": `[{"name":"x","ssh_url":"git@gitea.com:joe/x.git"}]`,
			"/users/joe/repos?page=2&limit=50": `[]`,
		},
		errors: map[string]error{
			"/orgs/joe/repos?page=1&limit=50": errors.New("tea api: HTTP 404 Not Found"),
		},
	}
	got, err := List(context.Background(), ListOptions{Name: "joe"}, f.run)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 || got[0].Name != "x" {
		t.Fatalf("got %+v", got)
	}
}

func TestList_FallsBackToAuthdUserEndpoint(t *testing.T) {
	f := &fakeRunner{
		responses: map[string]string{
			"user": `{"login":"joe"}`,
			"/user/repos?page=1&limit=50": `[
				{"name":"pub","private":false},
				{"name":"priv","private":true}
			]`,
			"/user/repos?page=2&limit=50": `[]`,
		},
		errors: map[string]error{
			"/orgs/joe/repos?page=1&limit=50": errors.New("tea api: HTTP 404 Not Found"),
		},
	}
	got, err := List(context.Background(), ListOptions{Name: "joe"}, f.run)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 2 {
		t.Fatalf("expected 2 repos, got %d", len(got))
	}
	for _, c := range f.calls {
		// We must NOT have called the public-only path.
		if last := c[len(c)-1]; last == "/users/joe/repos?page=1&limit=50" {
			t.Fatalf("should not have hit public-only endpoint; calls=%+v", f.calls)
		}
	}
}

func TestList_FiltersAndLimits(t *testing.T) {
	f := &fakeRunner{
		responses: map[string]string{
			"/orgs/acme/repos?page=1&limit=50": `[
				{"name":"a","archived":false,"fork":false},
				{"name":"b","archived":true,"fork":false},
				{"name":"c","archived":false,"fork":true},
				{"name":"d","archived":false,"fork":false}
			]`,
			"/orgs/acme/repos?page=2&limit=50": `[]`,
		},
	}
	got, err := List(context.Background(), ListOptions{
		Name:     "acme",
		Limit:    1,
		Archived: boolPtr(false),
		Fork:     boolPtr(false),
	}, f.run)
	if err != nil {
		t.Fatal(err)
	}
	names := make([]string, len(got))
	for i, r := range got {
		names[i] = r.Name
	}
	if !reflect.DeepEqual(names, []string{"a"}) {
		t.Fatalf("got %v, want [a]", names)
	}
}

func TestRepo_Visibility(t *testing.T) {
	cases := []struct {
		name string
		repo Repo
		want string
	}{
		{"public", Repo{}, "public"},
		{"private", Repo{Private: true}, "private"},
		{"internal", Repo{Internal: true}, "internal"},
		{"internal wins over private", Repo{Internal: true, Private: true}, "internal"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := tc.repo.Visibility(); got != tc.want {
				t.Fatalf("got %q, want %q", got, tc.want)
			}
		})
	}
}

func TestList_RequiresName(t *testing.T) {
	_, err := List(context.Background(), ListOptions{}, (&fakeRunner{}).run)
	if err == nil {
		t.Fatal("expected error for empty name")
	}
}
