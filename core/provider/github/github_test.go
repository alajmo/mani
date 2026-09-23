package github

import (
	"context"
	"errors"
	"reflect"
	"strings"
	"testing"
)

func boolPtr(b bool) *bool { return &b }

// fakeRunner builds a Runner whose response is keyed off the API path
// embedded in args. It records every call for assertions.
type fakeRunner struct {
	responses map[string]string // path → JSON body
	errors    map[string]error  // path → error
	calls     [][]string
}

func (f *fakeRunner) run(_ context.Context, args ...string) ([]byte, error) {
	f.calls = append(f.calls, args)
	// Two call shapes we care about:
	//   ["api", "--paginate", "<path>"]      → key = <path>
	//   ["api", "user", "-q", ".login"]      → key = "user"
	if len(args) < 2 || args[0] != "api" {
		return nil, errors.New("unexpected args: " + strings.Join(args, " "))
	}
	var key string
	switch args[1] {
	case "--paginate":
		key = args[2]
	default:
		key = args[1]
	}
	if e, ok := f.errors[key]; ok {
		return nil, e
	}
	if body, ok := f.responses[key]; ok {
		return []byte(body), nil
	}
	return nil, errors.New("no fixture for " + key)
}

func TestList_OrgSuccess(t *testing.T) {
	f := &fakeRunner{
		responses: map[string]string{
			"/orgs/acme/repos?per_page=100": `[
				{"name":"a","ssh_url":"git@github.com:acme/a.git","default_branch":"main"},
				{"name":"b","ssh_url":"git@github.com:acme/b.git","default_branch":"main"}
			]`,
		},
	}
	got, err := List(context.Background(), ListOptions{Name: "acme"}, f.run)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 2 || got[0].Name != "a" || got[1].Name != "b" {
		t.Fatalf("unexpected repos: %+v", got)
	}
	if len(f.calls) != 1 {
		t.Fatalf("expected one call, got %d", len(f.calls))
	}
}

func TestList_FallsBackToOtherUserPublicEndpointOn404(t *testing.T) {
	// Name is NOT the auth'd user → falls back to /users/<name>/repos.
	f := &fakeRunner{
		responses: map[string]string{
			"/users/joe/repos?per_page=100": `[{"name":"x","ssh_url":"git@github.com:joe/x.git"}]`,
			"user":                          "someone-else\n", // gh api user -q .login
		},
		errors: map[string]error{
			"/orgs/joe/repos?per_page=100": errors.New("gh api: HTTP 404: Not Found"),
		},
	}
	got, err := List(context.Background(), ListOptions{Name: "joe"}, f.run)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 || got[0].Name != "x" {
		t.Fatalf("unexpected repos: %+v", got)
	}
	// Calls: orgs/joe (404) → gh api user → users/joe = 3 calls
	if len(f.calls) != 3 {
		t.Fatalf("expected 3 calls (org, whoami, user-public), got %d: %+v", len(f.calls), f.calls)
	}
}

func TestList_FallsBackToAuthdUserEndpointForOwnRepos(t *testing.T) {
	// Name matches auth'd user → uses /user/repos (private+public).
	f := &fakeRunner{
		responses: map[string]string{
			"/user/repos?per_page=100&affiliation=owner": `[
				{"name":"pub","ssh_url":"git@github.com:joe/pub.git","visibility":"public"},
				{"name":"priv","ssh_url":"git@github.com:joe/priv.git","visibility":"private"}
			]`,
			"user": "joe\n",
		},
		errors: map[string]error{
			"/orgs/joe/repos?per_page=100": errors.New("gh api: HTTP 404: Not Found"),
		},
	}
	got, err := List(context.Background(), ListOptions{Name: "joe"}, f.run)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 2 {
		t.Fatalf("expected 2 repos (incl. private), got %d", len(got))
	}
	// Verify we didn't hit the public-only endpoint.
	for _, c := range f.calls {
		if c[len(c)-1] == "/users/joe/repos?per_page=100" {
			t.Fatalf("should not have called the public-only user endpoint, calls = %+v", f.calls)
		}
	}
}

func TestList_NonNotFoundErrorPropagates(t *testing.T) {
	f := &fakeRunner{
		errors: map[string]error{
			"/orgs/acme/repos?per_page=100": errors.New("gh api: HTTP 500: server error"),
		},
	}
	_, err := List(context.Background(), ListOptions{Name: "acme"}, f.run)
	if err == nil || !strings.Contains(err.Error(), "500") {
		t.Fatalf("expected 500 error, got %v", err)
	}
	if len(f.calls) != 1 {
		t.Fatalf("should not have fallen back, got %d calls", len(f.calls))
	}
}

func TestList_FiltersArchivedAndFork(t *testing.T) {
	f := &fakeRunner{
		responses: map[string]string{
			"/orgs/acme/repos?per_page=100": `[
				{"name":"a","archived":false,"fork":false},
				{"name":"b","archived":true,"fork":false},
				{"name":"c","archived":false,"fork":true},
				{"name":"d","archived":true,"fork":true}
			]`,
		},
	}
	got, err := List(context.Background(), ListOptions{
		Name:     "acme",
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
		t.Fatalf("filter mismatch: %v", names)
	}
}

func TestList_LimitTruncates(t *testing.T) {
	f := &fakeRunner{
		responses: map[string]string{
			"/orgs/acme/repos?per_page=100": `[
				{"name":"a"},{"name":"b"},{"name":"c"},{"name":"d"}
			]`,
		},
	}
	got, err := List(context.Background(), ListOptions{Name: "acme", Limit: 2}, f.run)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 2 {
		t.Fatalf("expected 2 repos, got %d", len(got))
	}
}

func TestList_RequiresName(t *testing.T) {
	_, err := List(context.Background(), ListOptions{}, (&fakeRunner{}).run)
	if err == nil {
		t.Fatal("expected error for empty name")
	}
}

func TestList_HintOrgSkipsUserFallback(t *testing.T) {
	// With HintOrg, a 404 should propagate rather than fall back to /users.
	f := &fakeRunner{
		errors: map[string]error{
			"/orgs/joe/repos?per_page=100": errors.New("gh api: HTTP 404: Not Found"),
		},
	}
	_, err := List(context.Background(), ListOptions{Name: "joe", Hint: HintOrg}, f.run)
	if err == nil {
		t.Fatal("expected error to propagate when HintOrg and org returns 404")
	}
	if len(f.calls) != 1 {
		t.Fatalf("HintOrg should only make 1 call; got %d: %+v", len(f.calls), f.calls)
	}
}

func TestList_HintUserSkipsOrgProbe(t *testing.T) {
	f := &fakeRunner{
		responses: map[string]string{
			"user": "someone-else\n",
			"/users/joe/repos?per_page=100": `[{"name":"x"}]`,
		},
	}
	_, err := List(context.Background(), ListOptions{Name: "joe", Hint: HintUser}, f.run)
	if err != nil {
		t.Fatal(err)
	}
	for _, c := range f.calls {
		last := c[len(c)-1]
		if strings.HasPrefix(last, "/orgs/") {
			t.Fatalf("HintUser should not probe /orgs/; got calls %+v", f.calls)
		}
	}
}

func TestList_DecodeError(t *testing.T) {
	f := &fakeRunner{
		responses: map[string]string{
			"/orgs/acme/repos?per_page=100": `not json`,
		},
	}
	_, err := List(context.Background(), ListOptions{Name: "acme"}, f.run)
	if err == nil || !strings.Contains(err.Error(), "decode") {
		t.Fatalf("expected decode error, got %v", err)
	}
}
