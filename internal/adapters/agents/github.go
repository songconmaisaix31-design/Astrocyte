package agents

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"

	"github.com/songconmaisaix31-design/Astrocyte/internal/apierrors"
	"github.com/songconmaisaix31-design/Astrocyte/internal/workspace/domain"
)

var githubOwner = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9-]{0,38}$`)
var githubRepoName = regexp.MustCompile(`^[A-Za-z0-9_.-]{1,100}$`)

// ParsePublicGitHubInput 只接受人类选定的公开 GitHub 账号或仓库。
func ParsePublicGitHubInput(input string) (string, string, error) {
	input = strings.TrimSpace(input)
	if strings.Contains(input, "://") {
		u, err := url.Parse(input)
		if err != nil || u.Scheme != "https" || u.Host != "github.com" || u.User != nil || u.RawQuery != "" || u.Fragment != "" || u.RawPath != "" {
			return "", "", errors.New("only public github.com repository/account URLs are accepted")
		}
		input = strings.Trim(u.Path, "/")
	}
	parts := strings.Split(input, "/")
	if len(parts) < 1 || len(parts) > 2 || !githubOwner.MatchString(parts[0]) {
		return "", "", errors.New("a public GitHub account or owner/repository is required")
	}
	repo := ""
	if len(parts) == 2 {
		repo = strings.TrimSuffix(parts[1], ".git")
		if !githubRepoName.MatchString(repo) || repo == "." || repo == ".." {
			return "", "", errors.New("invalid public GitHub repository name")
		}
	}
	return parts[0], repo, nil
}

type PublicGitHubSource struct{ client *http.Client }

func NewPublicGitHubSource() *PublicGitHubSource {
	return &PublicGitHubSource{client: &http.Client{Timeout: 30 * time.Second, CheckRedirect: rejectGitHubRedirect}}
}

func rejectGitHubRedirect(*http.Request, []*http.Request) error {
	return errors.New("GitHub redirects require a new explicit input")
}

type publicGitHubResponse struct {
	known         map[string]bool
	ID            int64     `json:"id"`
	FullName      string    `json:"full_name"`
	Private       bool      `json:"private"`
	HTMLURL       string    `json:"html_url"`
	CloneURL      string    `json:"clone_url"`
	Description   string    `json:"description"`
	DefaultBranch string    `json:"default_branch"`
	Language      string    `json:"language"`
	Stars         int       `json:"stargazers_count"`
	Archived      bool      `json:"archived"`
	Fork          bool      `json:"fork"`
	UpdatedAt     time.Time `json:"updated_at"`
	PushedAt      time.Time `json:"pushed_at"`
}

func (r *publicGitHubResponse) UnmarshalJSON(data []byte) error {
	type wire publicGitHubResponse
	var value wire
	if err := json.Unmarshal(data, &value); err != nil {
		return err
	}
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(data, &fields); err != nil {
		return err
	}
	*r = publicGitHubResponse(value)
	r.known = map[string]bool{}
	for key, value := range fields {
		r.known[key] = string(value) != "null"
	}
	return nil
}

func (r publicGitHubResponse) unknownFields() []string {
	result := []string{}
	for _, field := range [][2]string{{"description", "description"}, {"default_branch", "default_branch"}, {"language", "language"}, {"stargazers_count", "stars"}, {"archived", "archived"}, {"fork", "fork"}, {"updated_at", "updated_at"}, {"pushed_at", "pushed_at"}} {
		if !r.known[field[0]] {
			result = append(result, field[1])
		}
	}
	return result
}

func publicGitHubRequest(ctx context.Context, owner, repo string) (*http.Request, error) {
	path := "/repos/" + owner + "/" + repo
	if repo == "" {
		path = "/users/" + owner + "/repos?type=owner&sort=updated&per_page=100&page=1"
	}
	request, err := http.NewRequestWithContext(ctx, "GET", "https://api.github.com"+path, nil)
	if err != nil {
		return nil, err
	}
	request.Header.Set("Accept", "application/vnd.github+json")
	request.Header.Set("X-GitHub-Api-Version", "2022-11-28")
	request.Header.Set("User-Agent", "Astrocyte-public-metadata")
	return request, nil
}

func publicGitHubStatus(status int) error {
	if status != http.StatusOK {
		return &apierrors.ServiceError{Code: apierrors.ProviderUnavailable, Message: fmt.Sprintf("public GitHub API returned HTTP %d", status), RequiredAction: "check_public_input_or_rate_limit"}
	}
	return nil
}

func (s *PublicGitHubSource) FetchPublicRepositories(ctx context.Context, input string) ([]domain.GitHubMetadata, error) {
	owner, repo, err := ParsePublicGitHubInput(input)
	if err != nil {
		return nil, &apierrors.ServiceError{Code: apierrors.ValidationFailed, Message: err.Error()}
	}
	request, err := publicGitHubRequest(ctx, owner, repo)
	if err != nil {
		return nil, err
	}
	response, err := s.client.Do(request)
	if err != nil {
		return nil, &apierrors.ServiceError{Code: apierrors.ProviderUnavailable, Message: "public GitHub API unavailable", RequiredAction: "retry_public_repository_sync"}
	}
	defer response.Body.Close()
	if err := publicGitHubStatus(response.StatusCode); err != nil {
		return nil, err
	}
	data, err := io.ReadAll(io.LimitReader(response.Body, 2*1024*1024+1))
	if err != nil {
		return nil, err
	}
	if len(data) > 2*1024*1024 {
		return nil, errors.New("public GitHub response exceeds metadata bound")
	}
	return decodePublicGitHubMetadata(data, owner, repo)
}

func decodePublicGitHubMetadata(data []byte, owner, repo string) ([]domain.GitHubMetadata, error) {
	items := []publicGitHubResponse{}
	if repo != "" {
		var item publicGitHubResponse
		if err := json.Unmarshal(data, &item); err != nil {
			return nil, err
		}
		items = append(items, item)
	} else if err := json.Unmarshal(data, &items); err != nil {
		return nil, err
	}
	if len(items) > 100 {
		return nil, errors.New("invalid public GitHub metadata response")
	}
	result := []domain.GitHubMetadata{}
	for _, item := range items {
		o, r, e := ParsePublicGitHubInput(item.FullName)
		if !item.known["private"] || item.Private || item.ID <= 0 || e != nil || r == "" || !strings.EqualFold(o, owner) || (repo != "" && !strings.EqualFold(r, repo)) || item.CloneURL != "https://github.com/"+item.FullName+".git" || item.HTMLURL != "https://github.com/"+item.FullName {
			return nil, errors.New("GitHub response is not the selected public repository")
		}
		result = append(result, domain.GitHubMetadata{MetadataSource: "github_rest", UnknownFields: item.unknownFields(), GitHubID: item.ID, FullName: item.FullName, HTMLURL: item.HTMLURL, CloneURL: item.CloneURL, Description: item.Description, DefaultBranch: item.DefaultBranch, Language: item.Language, Stars: item.Stars, Archived: item.Archived, Fork: item.Fork, UpdatedAt: item.UpdatedAt, PushedAt: item.PushedAt})
	}
	return result, nil
}
