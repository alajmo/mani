package github

import (
	"context"
	"errors"
	"fmt"
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
	got, _, err := ListAll(context.Background(), ListOptions{Name: "acme"}, f.run)
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
	got, _, err := ListAll(context.Background(), ListOptions{Name: "joe"}, f.run)
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
	got, _, err := ListAll(context.Background(), ListOptions{Name: "joe"}, f.run)
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
	_, _, err := ListAll(context.Background(), ListOptions{Name: "acme"}, f.run)
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
	got, _, err := ListAll(context.Background(), ListOptions{
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
	got, _, err := ListAll(context.Background(), ListOptions{Name: "acme", Limit: 2}, f.run)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 2 {
		t.Fatalf("expected 2 repos, got %d", len(got))
	}
}

func TestList_RequiresName(t *testing.T) {
	_, _, err := ListAll(context.Background(), ListOptions{}, (&fakeRunner{}).run)
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
	_, _, err := ListAll(context.Background(), ListOptions{Name: "joe", Hint: HintOrg}, f.run)
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
			"user":                          "someone-else\n",
			"/users/joe/repos?per_page=100": `[{"name":"x"}]`,
		},
	}
	_, _, err := ListAll(context.Background(), ListOptions{Name: "joe", Hint: HintUser}, f.run)
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
	_, _, err := ListAll(context.Background(), ListOptions{Name: "acme"}, f.run)
	if err == nil || !strings.Contains(err.Error(), "decode") {
		t.Fatalf("expected decode error, got %v", err)
	}
}

func TestIsNotFound_OnlyLooksAtWhatTheCLIPrinted(t *testing.T) {
	notFound := []error{
		&cliError{args: []string{"api", "/orgs/acme/repos"}, output: "gh: Not Found (HTTP 404)"},
		&cliError{args: []string{"api", "/orgs/acme/repos"}, output: "404 Not Found"},
		fmt.Errorf("wrapped: %w", &cliError{args: []string{"api", "/orgs/acme/repos"}, output: "HTTP 404: Not Found"}),
		errors.New("tea api: HTTP 404 Not Found"), // an error of another runner
	}
	for _, err := range notFound {
		if !isNotFound(err) {
			t.Errorf("%v: expected not found", err)
		}
	}

	// The name of the source is part of the call, not of what went wrong
	other := []error{
		nil,
		&cliError{args: []string{"api", "/orgs/team404/repos?page=1"}, output: "401 Unauthorized"},
		&cliError{args: []string{"api", "/orgs/notfound/repos"}, output: "dial tcp: connection refused"},
	}
	for _, err := range other {
		if isNotFound(err) {
			t.Errorf("%v: not a 404", err)
		}
	}
}

func TestListAll_ReportsAPartialListing(t *testing.T) {
	list := func(responses map[string]string, name string) bool {
		t.Helper()
		// Not an organization: the user endpoints are tried
		f := &fakeRunner{responses: responses, errors: map[string]error{orgReposKey(name): errors.New("HTTP 404: Not Found")}}
		_, partial, err := ListAll(context.Background(), ListOptions{Name: name}, f.run)
		if err != nil {
			t.Fatal(err)
		}
		return partial
	}

	// The authenticated user: every repo is listed
	if list(ownReposResponses, "joe") {
		t.Error("the listing of the authenticated user is complete")
	}
	// Another user: only the public repos
	if !list(otherUserResponses, "jane") {
		t.Error("the listing of another user is partial")
	}
}

func orgReposKey(name string) string { return "/orgs/" + name + "/repos?per_page=100" }

var ownReposResponses = map[string]string{
	"user": "joe\n",
	"/user/repos?per_page=100&affiliation=owner": `[{"name":"x","ssh_url":"git@github.com:joe/x.git"}]`,
}

var otherUserResponses = map[string]string{
	"user":                           "joe\n",
	"/users/jane/repos?per_page=100": `[{"name":"x","ssh_url":"git@github.com:jane/x.git"}]`,
}

func TestListAll_OrganizationListingIsPartial(t *testing.T) {
	// An organization lists what the account can see: its private repos only
	// for a member whose token is authorized. What is missing may still exist.
	responses := map[string]string{orgReposKey("acme"): `[{"name":"x","ssh_url":"git@github.com:acme/x.git"}]`}
	f := &fakeRunner{responses: responses}
	repos, partial, err := ListAll(context.Background(), ListOptions{Name: "acme"}, f.run)
	if err != nil {
		t.Fatal(err)
	}
	if len(repos) != 1 || !partial {
		t.Fatalf("repos = %d, partial = %t; want 1 repo of a partial listing", len(repos), partial)
	}
}
