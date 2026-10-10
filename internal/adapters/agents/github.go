package agents

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"html"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/songconmaisaix31-design/Astrocyte/internal/apierrors"
	"github.com/songconmaisaix31-design/Astrocyte/internal/workspace/domain"
)

var githubOwner = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9-]{0,38}$`)
var githubRepoName = regexp.MustCompile(`^[A-Za-z0-9_.-]{1,100}$`)

// ParsePublicGitHubInput accepts only manual public GitHub names/URLs. Nothing
// supplied by the caller selects an API host, credential or filesystem path.
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

type PublicGitHubSource struct {
	client        *http.Client
	baseURL       string
	publicBaseURL string
}

func NewPublicGitHubSource() *PublicGitHubSource {
	return &PublicGitHubSource{client: &http.Client{Timeout: 30 * time.Second, CheckRedirect: func(*http.Request, []*http.Request) error {
		return errors.New("GitHub redirects require a new explicit input")
	}}, baseURL: "https://api.github.com", publicBaseURL: "https://github.com"}
}

type publicGitHubResponse struct {
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

func (s *PublicGitHubSource) FetchPublicRepositories(ctx context.Context, input string) ([]domain.GitHubMetadata, error) {
	owner, repo, err := ParsePublicGitHubInput(input)
	if err != nil {
		return nil, &apierrors.ServiceError{Code: apierrors.ValidationFailed, Message: err.Error()}
	}
	path := "/repos/" + owner + "/" + repo
	if repo == "" {
		path = "/users/" + owner + "/repos?type=owner&sort=updated&per_page=100&page=1"
	}
	request, err := http.NewRequestWithContext(ctx, "GET", s.baseURL+path, nil)
	if err != nil {
		return nil, err
	}
	request.Header.Set("Accept", "application/vnd.github+json")
	request.Header.Set("X-GitHub-Api-Version", "2022-11-28")
	request.Header.Set("User-Agent", "Astrocyte-public-metadata")
	response, err := s.client.Do(request)
	if err != nil {
		return nil, &apierrors.ServiceError{Code: apierrors.ProviderUnavailable, Message: "public GitHub API unavailable", RequiredAction: "retry_public_repository_sync"}
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		// This is a bounded read of the SAME explicitly selected official public
		// repository, only after a proven anonymous rate limit. Account listing,
		// authentication failures, redirects and private access never use it.
		if repo != "" && response.StatusCode == http.StatusForbidden && response.Header.Get("X-RateLimit-Remaining") == "0" {
			return s.fetchPublicRepositoryPage(ctx, owner, repo)
		}
		return nil, &apierrors.ServiceError{Code: apierrors.ProviderUnavailable, Message: fmt.Sprintf("public GitHub API returned HTTP %d", response.StatusCode), RequiredAction: "check_public_input_or_rate_limit"}
	}
	data, err := io.ReadAll(io.LimitReader(response.Body, 2*1024*1024+1))
	if err != nil || len(data) > 2*1024*1024 {
		return nil, errors.New("public GitHub response exceeds metadata bound")
	}
	items := []publicGitHubResponse{}
	if repo != "" {
		var item publicGitHubResponse
		err = json.Unmarshal(data, &item)
		items = append(items, item)
	} else {
		err = json.Unmarshal(data, &items)
	}
	if err != nil || len(items) > 100 {
		return nil, errors.New("invalid public GitHub metadata response")
	}
	result := []domain.GitHubMetadata{}
	for _, item := range items {
		o, r, e := ParsePublicGitHubInput(item.FullName)
		if item.Private || item.ID <= 0 || e != nil || r == "" || !strings.EqualFold(o, owner) || (repo != "" && !strings.EqualFold(r, repo)) || item.CloneURL != "https://github.com/"+item.FullName+".git" || item.HTMLURL != "https://github.com/"+item.FullName {
			return nil, errors.New("GitHub response is not the selected public repository")
		}
		result = append(result, domain.GitHubMetadata{MetadataSource: "github_rest", UnknownFields: []string{}, GitHubID: item.ID, FullName: item.FullName, HTMLURL: item.HTMLURL, CloneURL: item.CloneURL, Description: item.Description, DefaultBranch: item.DefaultBranch, Language: item.Language, Stars: item.Stars, Archived: item.Archived, Fork: item.Fork, UpdatedAt: item.UpdatedAt, PushedAt: item.PushedAt})
	}
	return result, nil
}

var githubMetaTag = regexp.MustCompile(`(?i)<meta\s+[^>]+>`)
var githubMetaAttribute = regexp.MustCompile(`(?i)\b(name|content)\s*=\s*(?:"([^"]*)"|'([^']*)')`)

func (s *PublicGitHubSource) fetchPublicRepositoryPage(ctx context.Context, owner, repo string) ([]domain.GitHubMetadata, error) {
	request, err := http.NewRequestWithContext(ctx, "GET", s.publicBaseURL+"/"+owner+"/"+repo, nil)
	if err != nil {
		return nil, err
	}
	request.Header.Set("Accept", "text/html")
	request.Header.Set("User-Agent", "Astrocyte-public-metadata")
	response, err := s.client.Do(request)
	if err != nil {
		return nil, &apierrors.ServiceError{Code: apierrors.ProviderUnavailable, Message: "public GitHub page unavailable after anonymous rate limit"}
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK || !strings.HasPrefix(strings.ToLower(response.Header.Get("Content-Type")), "text/html") {
		return nil, &apierrors.ServiceError{Code: apierrors.ProviderUnavailable, Message: "public GitHub page did not return repository metadata"}
	}
	data, err := io.ReadAll(io.LimitReader(response.Body, 4*1024*1024+1))
	if err != nil || len(data) > 4*1024*1024 {
		return nil, errors.New("public GitHub page exceeds bound")
	}
	return parsePublicGitHubPage(data, owner, repo)
}

func parsePublicGitHubPage(data []byte, owner, repo string) ([]domain.GitHubMetadata, error) {
	meta := map[string]string{}
	for _, tag := range githubMetaTag.FindAll(data, -1) {
		if len(tag) > 1000 {
			continue
		}
		attrs := map[string]string{}
		for _, a := range githubMetaAttribute.FindAllSubmatch(tag, -1) {
			key := strings.ToLower(string(a[1]))
			value := html.UnescapeString(string(a[2]) + string(a[3]))
			if _, ok := attrs[key]; ok {
				return nil, errors.New("ambiguous public repository metadata")
			}
			attrs[key] = value
		}
		name := attrs["name"]
		if !strings.HasPrefix(name, "octolytics-dimension-repository_") {
			continue
		}
		if old, ok := meta[name]; ok && old != attrs["content"] {
			return nil, errors.New("conflicting public repository metadata")
		}
		meta[name] = attrs["content"]
	}
	name := meta["octolytics-dimension-repository_nwo"]
	o, r, err := ParsePublicGitHubInput(name)
	id, idErr := strconv.ParseInt(meta["octolytics-dimension-repository_id"], 10, 64)
	if err != nil || idErr != nil || id <= 0 || !strings.EqualFold(owner, o) || !strings.EqualFold(repo, r) || meta["octolytics-dimension-repository_public"] != "true" {
		return nil, errors.New("official public page lacks stable ID or selected public repository proof")
	}
	unknown := []string{"description", "default_branch", "language", "stars", "archived", "updated_at", "pushed_at"}
	forkValue := meta["octolytics-dimension-repository_is_fork"]
	if forkValue != "true" && forkValue != "false" {
		unknown = append(unknown, "fork")
	}
	return []domain.GitHubMetadata{{MetadataSource: "github_public_html", UnknownFields: unknown, GitHubID: id, FullName: name, HTMLURL: "https://github.com/" + name, CloneURL: "https://github.com/" + name + ".git", Fork: forkValue == "true"}}, nil
}
