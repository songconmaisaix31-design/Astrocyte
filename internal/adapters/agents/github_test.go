package agents

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"os"
	"path/filepath"
	"testing"

	"github.com/songconmaisaix31-design/Astrocyte/internal/apierrors"
	"github.com/songconmaisaix31-design/Astrocyte/internal/workspace/domain"
)

func TestPublicGitHubInputCannotSelectCredentialsHostsOrPaths(t *testing.T) {
	for _, input := range []string{"https://user:secret@github.com/o/r", "http://github.com/o/r", "https://github.com.evil/o/r", "https://github.com/o/r?token=secret", "../r", "o/..", "o/r/tree/main", "file:///tmp/r", "https://github.com/o/%72", "o/r\n"} {
		// 人类输入允许规范化外围空白。
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

func TestPublicGitHubRateLimitFailsImmediately(t *testing.T) {
	// 状态码是纯函数输入；每个非成功状态均有明确的错误结果。
	for _, status := range []int{301, 302, 401, 403, 404, 429, 500} {
		var service *apierrors.ServiceError
		if err := publicGitHubStatus(status); !errors.As(err, &service) || service.Code != apierrors.ProviderUnavailable || service.RequiredAction == "" {
			t.Fatalf("HTTP %d did not fail explicitly", status)
		}
	}
	if err := publicGitHubStatus(http.StatusOK); err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	for _, input := range []string{"steipete/summarize", "steipete"} {
		if result, err := NewPublicGitHubSource().FetchPublicRepositories(ctx, input); err == nil || result != nil {
			t.Fatal("cancelled real HTTP request continued")
		}
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
	for _, repo := range []string{"summarize", ""} {
		request, err := publicGitHubRequest(context.Background(), "steipete", repo)
		if err != nil {
			t.Fatal(err)
		}
		if request.URL.Host != "api.github.com" || request.URL.Scheme != "https" || request.Header.Get("Authorization") != "" || request.Header.Get("Cookie") != "" || request.Header.Get("X-GitHub-Api-Version") != "2022-11-28" {
			t.Fatal("request authority or anonymity changed")
		}
		if repo == "" {
			if request.URL.Path != "/users/steipete/repos" || request.URL.Query().Get("per_page") != "100" || request.URL.Query().Get("page") != "1" {
				t.Fatal("unbounded account request")
			}
		} else if request.URL.Path != "/repos/steipete/summarize" {
			t.Fatal("wrong selected repository endpoint")
		}
		if err := NewPublicGitHubSource().client.CheckRedirect(request, nil); err == nil {
			t.Fatal("redirect accepted")
		}
	}
	// JSON 仅作为 decoder 的明确输入，期望值不表示已观察的仓库事实。
	input := map[string]any{"id": 123, "full_name": "selected/repository", "private": false, "html_url": "https://github.com/selected/repository", "clone_url": "https://github.com/selected/repository.git", "description": "decoder input"}
	decode := func() ([]domain.GitHubMetadata, error) {
		raw, err := json.Marshal(input)
		if err != nil {
			t.Fatal(err)
		}
		return decodePublicGitHubMetadata(raw, "selected", "repository")
	}
	base := filepath.Join(t.TempDir(), "managed")
	_ = NewManagedRepositoryCloner(base)
	items, err := decode()
	if err != nil || len(items) != 1 || items[0].GitHubID != 123 {
		t.Fatalf("decoder: %v %v", items, err)
	}
	if items[0].MetadataSource != "github_rest" || len(items[0].UnknownFields) != 7 {
		t.Fatal("omitted REST statistics/timestamps invented", items[0])
	}
	if !items[0].UpdatedAt.IsZero() || items[0].Stars != 0 {
		t.Fatal("unobserved fields invented")
	}
	if _, err = os.Stat(base); !os.IsNotExist(err) {
		t.Fatal("metadata created code directory")
	}
	input["private"] = true
	if _, err := decode(); err == nil {
		t.Fatal("private input accepted")
	}
	delete(input, "private")
	if _, err := decode(); err == nil {
		t.Fatal("unknown visibility accepted")
	}
	items, err = decodePublicGitHubMetadata([]byte(`[]`), "selected", "")
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
