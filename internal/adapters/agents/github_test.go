package agents

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestPublicGitHubInputCannotSelectCredentialsHostsOrPaths(t *testing.T) {
	for _, input := range []string{"https://user:secret@github.com/o/r", "http://github.com/o/r", "https://github.com.evil/o/r", "https://github.com/o/r?token=secret", "../r", "o/..", "o/r/tree/main", "file:///tmp/r", "https://github.com/o/%72", "o/r\n"} {
		// Surrounding whitespace is intentionally normalized for manual inputs.
		if input == "o/r\n" {
			continue
		}
		if _, _, err := ParsePublicGitHubInput(input); err == nil {
			t.Errorf("accepted %q", input)
		}
	}
	for _, input := range []string{"steipete/summarize", "https://github.com/steipete/summarize.git", "steipete", "https://github.com/steipete/"} {
		if _, _, err := ParsePublicGitHubInput(input); err != nil {
			t.Errorf("rejected %q: %v", input, err)
		}
	}
}

func TestPublicGitHubRateLimitFallbackKeepsStableIdentityAndUnknowns(t *testing.T) {
	apiCalls, pageCalls := 0, 0
	rateLimited := true
	public := true
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Cookie") != "" || r.Header.Get("Authorization") != "" {
			t.Fatal("fallback sent authentication")
		}
		if strings.HasPrefix(r.URL.Path, "/repos/") || strings.HasPrefix(r.URL.Path, "/users/") {
			apiCalls++
			if rateLimited {
				w.Header().Set("X-RateLimit-Remaining", "0")
			}
			w.WriteHeader(403)
			return
		}
		pageCalls++
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		visibility := "true"
		if !public {
			visibility = "false"
		}
		_, _ = w.Write([]byte(`<meta name="octolytics-dimension-repository_id" content="1118209243"><meta content="steipete/summarize" name="octolytics-dimension-repository_nwo"><meta name="octolytics-dimension-repository_public" content="` + visibility + `"><meta name="octolytics-dimension-repository_is_fork" content="false">`))
	}))
	defer server.Close()
	s := NewPublicGitHubSource()
	s.baseURL = server.URL
	s.publicBaseURL = server.URL
	items, err := s.FetchPublicRepositories(context.Background(), "steipete/summarize")
	if err != nil || len(items) != 1 || items[0].GitHubID != 1118209243 || items[0].MetadataSource != "github_public_html" || len(items[0].UnknownFields) != 7 || apiCalls != 1 || pageCalls != 1 {
		t.Fatalf("official fallback %+v %v", items, err)
	}
	if !items[0].UpdatedAt.IsZero() || items[0].Stars != 0 {
		t.Fatal("unobserved fields invented")
	}
	if _, err = s.FetchPublicRepositories(context.Background(), "steipete"); err == nil || pageCalls != 1 {
		t.Fatal("account rate limit expanded scope")
	}
	rateLimited = false
	if _, err = s.FetchPublicRepositories(context.Background(), "steipete/summarize"); err == nil || pageCalls != 1 {
		t.Fatal("authentication403 triggered fallback")
	}
	rateLimited = true
	public = false
	if _, err = s.FetchPublicRepositories(context.Background(), "steipete/summarize"); err == nil {
		t.Fatal("private page accepted")
	}
	if _, err = parsePublicGitHubPage([]byte(`<meta name="octolytics-dimension-repository_id" content="1">`), "steipete", "summarize"); err == nil {
		t.Fatal("missing page markers reported success")
	}
}

func TestManagedCheckoutPublicationNeverReplacesExistingTarget(t *testing.T) {
	base := t.TempDir()
	source := filepath.Join(base, "source")
	target := filepath.Join(base, "target")
	if err := os.Mkdir(source, 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir(target, 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(source, "own"), []byte("stage"), 0600); err != nil {
		t.Fatal(err)
	}
	if err := publishManagedCheckout(source, target); err == nil {
		t.Fatal("existing empty directory replaced")
	}
	if _, err := os.Stat(filepath.Join(source, "own")); err != nil {
		t.Fatal("owned stage lost on failed publication")
	}
}

func TestPublicGitHubSyncIsAnonymousMetadataOnlyAndRejectsPrivateOrRedirect(t *testing.T) {
	mode, calls := "repo", 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		if r.Header.Get("Authorization") != "" || r.Header.Get("Cookie") != "" {
			t.Error("sent private authentication")
		}
		if r.Header.Get("X-GitHub-Api-Version") != "2022-11-28" {
			t.Error("missing pinned API version")
		}
		if mode == "redirect" {
			http.Redirect(w, r, "/other", 302)
			return
		}
		if mode == "rate" {
			w.WriteHeader(403)
			return
		}
		if mode == "account" {
			if r.URL.Path != "/users/steipete/repos" || r.URL.Query().Get("per_page") != "100" || r.URL.Query().Get("page") != "1" {
				t.Error("unbounded account lookup")
			}
			_, _ = w.Write([]byte(`[]`))
			return
		}
		if r.URL.Path != "/repos/steipete/summarize" {
			t.Error("unexpected endpoint")
		}
		private := "false"
		if mode == "private" {
			private = "true"
		}
		_, _ = w.Write([]byte(`{"id":123,"full_name":"steipete/summarize","private":` + private + `,"html_url":"https://github.com/steipete/summarize","clone_url":"https://github.com/steipete/summarize.git","description":"metadata only"}`))
	}))
	defer server.Close()
	source := NewPublicGitHubSource()
	source.baseURL = server.URL
	base := filepath.Join(t.TempDir(), "managed")
	_ = NewManagedRepositoryCloner(base)
	items, err := source.FetchPublicRepositories(context.Background(), "steipete/summarize")
	if err != nil || len(items) != 1 || items[0].GitHubID != 123 || calls != 1 {
		t.Fatalf("metadata: %v %v calls=%d", items, err, calls)
	}
	if items[0].MetadataSource != "github_rest" || len(items[0].UnknownFields) != 7 {
		t.Fatal("omitted REST statistics/timestamps invented", items[0])
	}
	if _, err = os.Stat(base); !os.IsNotExist(err) {
		t.Fatal("metadata created code directory")
	}
	for _, m := range []string{"private", "redirect", "rate"} {
		mode = m
		before := calls
		if _, err = source.FetchPublicRepositories(context.Background(), "steipete/summarize"); err == nil {
			t.Fatalf("%s accepted", m)
		}
		if calls != before+1 {
			t.Fatal("automatic retry/redirect")
		}
	}
	mode = "account"
	items, err = source.FetchPublicRepositories(context.Background(), "steipete")
	if err != nil || len(items) != 0 {
		t.Fatal("empty public list not preserved", err)
	}
}

func TestManagedCloneRejectsUnsafeInputsAndPreservesExistingDirectory(t *testing.T) {
	base := filepath.Join(t.TempDir(), "managed")
	cloner := NewManagedRepositoryCloner(base)
	for _, test := range [][2]string{{"../escape", "https://github.com/steipete/summarize.git"}, {"github-1", "file:///tmp/repo"}, {"github-1", "https://u:p@github.com/steipete/summarize.git"}} {
		if _, err := cloner.CloneRepository(context.Background(), test[0], test[1]); err == nil {
			t.Fatal("unsafe clone accepted", test)
		}
	}
	if _, err := os.Stat(base); !os.IsNotExist(err) {
		t.Fatal("invalid request created checkout root")
	}
	target := filepath.Join(base, "github-1")
	if err := os.MkdirAll(target, 0700); err != nil {
		t.Fatal(err)
	}
	sentinel := filepath.Join(target, "user.txt")
	if err := os.WriteFile(sentinel, []byte("preserve"), 0600); err != nil {
		t.Fatal(err)
	}
	if _, err := cloner.CloneRepository(context.Background(), "github-1", "https://github.com/steipete/summarize.git"); err == nil {
		t.Fatal("unrelated directory accepted")
	}
	data, err := os.ReadFile(sentinel)
	if err != nil || string(data) != "preserve" {
		t.Fatal("existing directory overwritten")
	}
}
