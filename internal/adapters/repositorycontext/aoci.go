package repositorycontext

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"os/exec"
	"slices"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/songconmaisaix31-design/Astrocyte/internal/apierrors"
	"github.com/songconmaisaix31-design/Astrocyte/internal/workspace/domain"
)

const aociVersion = "v0.1.0-rc20/45990a1dcfc4415efdc6fbfa13bc159a415f127c"

type aociEntry struct {
	Path string `json:"path"`
	Kind string `json:"kind"`
	F    string `json:"f"`
	R    string `json:"r"`
	A    string `json:"a"`
	S    string `json:"s"`
}
type aociState struct {
	BinaryVersion string `json:"binary_version"`
	Error         string `json:"error"`
	Facts         *struct {
		Aligned bool   `json:"governance_aligned"`
		Result  string `json:"result"`
	} `json:"facts"`
}

func aociJSON(ctx context.Context, client *http.Client, base *url.URL, endpoint, root string, max int64, out any) error {
	target := *base
	target.Path = endpoint
	query := url.Values{}
	query.Set("repo", root)
	target.RawQuery = query.Encode()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, target.String(), nil)
	if err != nil {
		return err
	}
	response, err := client.Do(req)
	if err != nil {
		return failure(apierrors.ProviderUnavailable, "AOCI observation is unavailable")
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		return failure(apierrors.ProviderUnavailable, "AOCI observation was refused")
	}
	data, err := io.ReadAll(io.LimitReader(response.Body, max+1))
	if err != nil {
		return err
	}
	if int64(len(data)) > max {
		return failure(apierrors.BudgetExhausted, "AOCI response exceeds the approved byte budget")
	}
	if err := json.Unmarshal(data, out); err != nil {
		return failure(apierrors.ProviderUnavailable, "AOCI JSON response is invalid")
	}
	return nil
}

// 只在显式刷新内运行官方只读JSON服务，禁用跨仓库发现，不打开浏览器。
func observeAOCI(ctx context.Context, binary, root string, b domain.RepositoryBudget) ([]domain.RepositoryMapEntry, error) {
	if binary == "" {
		return nil, failure(apierrors.ProviderUnavailable, "AOCI is not configured")
	}
	child, cancel := context.WithCancel(ctx)
	defer cancel()
	cmd := exec.CommandContext(child, binary, "--repo", root, "ui", "--port", "0", "--host", "127.0.0.1", "--discover=false", "--json")
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, err
	}
	cmd.Stderr = io.Discard
	if err := cmd.Start(); err != nil {
		return nil, failure(apierrors.ProviderUnavailable, "AOCI observation process could not start")
	}
	defer func() { cancel(); _ = cmd.Wait() }()
	var ready struct {
		URL          string `json:"url"`
		Repositories int    `json:"repositories"`
	}
	if err := json.NewDecoder(io.LimitReader(stdout, 4096)).Decode(&ready); err != nil {
		return nil, failure(apierrors.ProviderUnavailable, "AOCI readiness was not confirmed")
	}
	base, err := url.Parse(ready.URL)
	if err != nil || base.Scheme != "http" || base.Hostname() != "127.0.0.1" || base.Port() == "" || base.User != nil || ready.Repositories != 1 {
		return nil, failure(apierrors.ProviderUnavailable, "AOCI did not bind the isolated loopback endpoint")
	}
	transport := &http.Transport{}
	defer transport.CloseIdleConnections()
	client := &http.Client{Transport: transport, Timeout: 30 * time.Second, CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }}
	var state aociState
	if err := aociJSON(ctx, client, base, "/api/state", root, b.MaxBytes, &state); err != nil {
		return nil, err
	}
	if state.BinaryVersion != "v0.1.0-rc20" && state.BinaryVersion != "0.1.0-rc20" {
		return nil, failure(apierrors.ProviderUnavailable, "AOCI version is not supported")
	}
	if state.Error != "" || state.Facts == nil {
		return nil, failure(apierrors.EvidenceMissing, "AOCI has no valid repository map")
	}
	if !state.Facts.Aligned || state.Facts.Result != "aligned" {
		return nil, failure(apierrors.IndexStale, "AOCI map is missing entries or differs from its source")
	}
	var entries struct {
		Total   int         `json:"total"`
		Matched int         `json:"matched"`
		Entries []aociEntry `json:"entries"`
	}
	if err := aociJSON(ctx, client, base, "/api/entries", root, b.MaxBytes, &entries); err != nil {
		return nil, err
	}
	if len(entries.Entries) == 0 {
		return nil, failure(apierrors.EvidenceMissing, "AOCI map contains no authored entries")
	}
	if len(entries.Entries) > b.MaxResults || entries.Total != len(entries.Entries) || entries.Matched != entries.Total {
		return nil, failure(apierrors.BudgetExhausted, "the complete AOCI map exceeds the result budget")
	}
	result := []domain.RepositoryMapEntry{}
	for _, entry := range entries.Entries {
		if entry.Kind != "code" || !safeRelative(entry.Path) || privatePath(entry.Path) {
			return nil, failure(apierrors.ScopeDenied, "AOCI map includes an unsupported or private object")
		}
		if strings.Contains(entry.F, root) || strings.Contains(entry.R, root) || strings.Contains(entry.A, root) || strings.Contains(entry.S, root) {
			return nil, failure(apierrors.ScopeDenied, "AOCI entry contains a private absolute root")
		}
		result = append(result, domain.RepositoryMapEntry{Path: entry.Path, Responsibility: entry.F, Relationships: entry.R, API: entry.A, Constraints: entry.S, EvidenceKind: "aoci_model_authored_fras"})
	}
	return result, nil
}

func (s *Service) RefreshRepositoryMap(ctx context.Context, c domain.Caller, q domain.RepositoryQuery) (domain.RepositoryIndex, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	var result domain.RepositoryIndex
	if c.Kind != "human" || c.ID == "" {
		return result, failure(apierrors.ScopeDenied, "map refresh requires the human control plane")
	}
	before, q, err := s.capture(ctx, c, q)
	if err != nil {
		return result, err
	}
	if !slices.Equal(q.Roots, []string{"."}) {
		return result, failure(apierrors.ScopeDenied, "AOCI whole-index observation requires explicit full project directory authorization")
	}
	entries, err := observeAOCI(ctx, s.options.AOCI, before.Root, q.Budget)
	if err != nil {
		return result, err
	}
	after, _, err := s.capture(ctx, c, q)
	if err != nil {
		return result, err
	}
	if !SameVersion(before, after) {
		return result, failure(apierrors.IndexStale, "repository changed during map observation")
	}
	expected := ""
	old, err := s.indexes.LoadRepositoryIndex(ctx, q.ProjectID, before.WorktreeID, "aoci")
	if err == nil {
		expected = old.Generation
	} else if !isNotFound(err) {
		return result, err
	}
	result = domain.RepositoryIndex{Backend: "aoci", Version: aociVersion, Generation: uuid.NewString(), Snapshot: before, State: "ready", BuiltAt: time.Now().UTC(), MapEntries: entries}
	if err := s.indexes.SaveRepositoryIndex(ctx, result, expected); err != nil {
		return result, err
	}
	return result, nil
}

func (s *Service) ReadRepositoryMap(ctx context.Context, c domain.Caller, q domain.RepositoryQuery) (domain.RepositoryResult, error) {
	var result domain.RepositoryResult
	before, q, err := s.capture(ctx, c, q)
	if err != nil {
		return result, err
	}
	index, err := s.indexes.LoadRepositoryIndex(ctx, q.ProjectID, before.WorktreeID, "aoci")
	if err != nil {
		return result, err
	}
	if index.State != "ready" || index.Version != aociVersion || !SameVersion(index.Snapshot, before) {
		return result, failure(apierrors.IndexStale, "saved AOCI map does not match the authorized source snapshot")
	}
	if len(index.MapEntries) > q.Budget.MaxResults {
		return result, failure(apierrors.BudgetExhausted, "saved AOCI map exceeds the result budget")
	}
	after, _, err := s.capture(ctx, c, q)
	if err != nil {
		return result, err
	}
	if !SameVersion(before, after) {
		return result, failure(apierrors.IndexStale, "repository changed during map query")
	}
	result = domain.RepositoryResult{Index: index, MapEntries: index.MapEntries, Symbols: []domain.RepositorySymbol{}, Relations: []domain.RepositoryRelation{}, Coverage: slices.Clone(q.Roots), Gaps: []string{"model_authored_meaning_requires_source_review", "repository_map_is_not_embedding_or_compiler_evidence"}}
	return result, nil
}
