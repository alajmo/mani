package sources

import (
	"strings"
	"testing"
)

func TestParse(t *testing.T) {
	cases := []struct {
		in      string
		wantSrc string // empty means parse should error
		wantErr string
	}{
		{"github/my-org", "github/my-org", ""},
		{"github-org/my-org", "github-org/my-org", ""},
		{"github-user/joe", "github-user/joe", ""},
		{"gitea/my-org", "gitea/my-org", ""},
		{"gitea-org/my-org", "gitea-org/my-org", ""},
		{"gitea-user/joe", "gitea-user/joe", ""},
		{"forgejo/my-org", "forgejo/my-org", ""},
		{"forgejo-org/my-org", "forgejo-org/my-org", ""},
		{"codeberg/joe", "codeberg/joe", ""},
		{"codeberg-user/joe", "codeberg-user/joe", ""},
		{"", "", "expected <provider>/<name>"},
		{"github", "", "expected <provider>/<name>"},
		{"github/", "", "expected <provider>/<name>"},
		{"/my-org", "", "expected <provider>/<name>"},
		{"gitlab/my-group", "", "unsupported source provider"},
		{"github-team/x", "", "unsupported source provider"},
	}
	for _, tc := range cases {
		t.Run(tc.in, func(t *testing.T) {
			got, err := Parse(tc.in)
			if tc.wantErr != "" {
				if err == nil || !strings.Contains(err.Error(), tc.wantErr) {
					t.Fatalf("err = %v, want containing %q", err, tc.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected err: %v", err)
			}
			if got.Source != tc.wantSrc {
				t.Fatalf("got Source=%q, want %q", got.Source, tc.wantSrc)
			}
		})
	}
}

func TestSource_ProviderNameHint(t *testing.T) {
	cases := []struct {
		src                        string
		wantProv, wantNm, wantHint string
	}{
		{"github/my-org", "github", "my-org", ""},
		{"github-org/my-org", "github", "my-org", "org"},
		{"github-user/joe", "github", "joe", "user"},
		{"gitea/my-org", "gitea", "my-org", ""},
		{"gitea-org/my-org", "gitea", "my-org", "org"},
		{"forgejo/my-org", "gitea", "my-org", ""},     // canonicalized
		{"codeberg/joe", "gitea", "joe", ""},          // canonicalized
		{"codeberg-user/joe", "gitea", "joe", "user"}, // canonicalized + hint
		{"", "", "", ""},
	}
	for _, tc := range cases {
		t.Run(tc.src, func(t *testing.T) {
			s := Source{Source: tc.src}
			if got := s.Provider(); got != tc.wantProv {
				t.Fatalf("Provider() = %q, want %q", got, tc.wantProv)
			}
			if got := s.Name(); got != tc.wantNm {
				t.Fatalf("Name() = %q, want %q", got, tc.wantNm)
			}
			if got := s.Hint(); got != tc.wantHint {
				t.Fatalf("Hint() = %q, want %q", got, tc.wantHint)
			}
		})
	}
}

func TestSource_Label(t *testing.T) {
	// Label preserves what the user typed — codeberg stays codeberg.
	if got := (Source{Source: "codeberg/joe"}).Label(); got != "codeberg/joe" {
		t.Fatalf("got %q, want codeberg/joe", got)
	}
}

func TestSource_ToGitHubListOptions(t *testing.T) {
	// Default (Archived/Fork false) means EXCLUDE: the provider filter is the
	// non-nil pointer to false (keep only repos where the attribute is false).
	s := Source{Source: "github/x", Limit: 50}
	opts := s.ToGitHubListOptions()
	if opts.Name != "x" || opts.Limit != 50 {
		t.Fatalf("opts mismatch: %+v", opts)
	}
	if opts.Archived == nil || *opts.Archived != false {
		t.Fatalf("default archived should map to filter=&false, got %v", opts.Archived)
	}
	if opts.Fork == nil || *opts.Fork != false {
		t.Fatalf("default fork should map to filter=&false, got %v", opts.Fork)
	}

	// Include (true) means NO filter: the provider pointer is nil.
	incl := Source{Source: "github/x", Archived: true, Fork: true}.ToGitHubListOptions()
	if incl.Archived != nil {
		t.Fatalf("archived:true should map to no filter (nil), got %v", *incl.Archived)
	}
	if incl.Fork != nil {
		t.Fatalf("fork:true should map to no filter (nil), got %v", *incl.Fork)
	}
}

func TestSource_ToGiteaListOptions(t *testing.T) {
	s := Source{Source: "codeberg/y", Limit: 25}
	opts := s.ToGiteaListOptions()
	if opts.Name != "y" || opts.Limit != 25 {
		t.Fatalf("opts mismatch: %+v", opts)
	}
	// Same include→filter mapping as GitHub.
	if opts.Archived == nil || *opts.Archived != false {
		t.Fatalf("default archived should map to filter=&false, got %v", opts.Archived)
	}
	incl := Source{Source: "codeberg/y", Archived: true}.ToGiteaListOptions()
	if incl.Archived != nil {
		t.Fatalf("archived:true should map to no filter (nil), got %v", *incl.Archived)
	}
}

func TestIncludeFilter(t *testing.T) {
	// include=true → no filter (nil)
	if got := includeFilter(true); got != nil {
		t.Fatalf("includeFilter(true) = %v, want nil", *got)
	}
	// include=false → &false (keep only attribute==false)
	got := includeFilter(false)
	if got == nil || *got != false {
		t.Fatalf("includeFilter(false) = %v, want &false", got)
	}
}

func TestSource_MapOptions(t *testing.T) {
	s := Source{Source: "github/x", NoTags: true, NoTopicTags: true}
	gh := s.ToGitHubMapOptions()
	if !gh.NoTags || !gh.NoTopicTags {
		t.Fatalf("github map opts mismatch: %+v", gh)
	}
	gt := s.ToGiteaMapOptions()
	if !gt.NoTags || !gt.NoTopicTags {
		t.Fatalf("gitea map opts mismatch: %+v", gt)
	}
}
