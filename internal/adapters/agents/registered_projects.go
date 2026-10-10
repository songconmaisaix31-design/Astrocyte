package agents

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"sort"
	"strings"
	"time"

	"github.com/songconmaisaix31-design/Astrocyte/internal/workspace/domain"
)

type RegisteredProjects struct {
	read func(context.Context, string, string) ([]byte, error)
}

func NewRegisteredProjects() *RegisteredProjects {
	return &RegisteredProjects{read: registeredReadCommand}
}

type orcaRegisteredRepo struct {
	ID   string `json:"id"`
	Path string `json:"path"`
	Name string `json:"displayName"`
}

type orcaRegisteredWorktree struct {
	RepoID           string `json:"repoId"`
	HostID           string `json:"hostId"`
	Path             string `json:"path"`
	Name             string `json:"displayName"`
	CreatedWithAgent string `json:"createdWithAgent"`
	LastActivityAt   int64  `json:"lastActivityAt"`
}

func emptyRegisteredSnapshot() domain.ProjectDiscoverySnapshot {
	return domain.ProjectDiscoverySnapshot{Status: "unknown", Projects: []domain.RegisteredProject{}, Failures: []domain.ProjectDiscoveryFailure{}}
}

func rootKey(root string) string {
	root = filepath.Clean(root)
	if runtime.GOOS == "windows" {
		return strings.ToLower(root)
	}
	return root
}

// DiscoverRegistered reads only Orca's public registration/worktree metadata,
// then bounded candidate directories within those roots. It neither reads the
// Orca private database nor launches or controls any Agent. No cwd fallback.
func (s *RegisteredProjects) DiscoverRegistered(ctx context.Context) (domain.ProjectDiscoverySnapshot, error) {
	ctx, cancel := context.WithTimeout(ctx, 120*time.Second)
	defer cancel()
	result := emptyRegisteredSnapshot()
	data, err := s.read(ctx, "orca_repos", "")
	if err != nil {
		return result, errors.New("registered project source unavailable")
	}
	var repos struct {
		OK     bool `json:"ok"`
		Result struct {
			Repos []orcaRegisteredRepo `json:"repos"`
		} `json:"result"`
	}
	if json.Unmarshal(data, &repos) != nil || !repos.OK || repos.Result.Repos == nil || len(repos.Result.Repos) > 256 {
		return result, errors.New("registered project source invalid or over bound")
	}
	now := time.Now().UTC()
	result.ObservedAt = &now
	result.Status = "complete"
	byRoot := map[string]int{}
	byRepo := map[string]string{}
	fail := func(root, reason string) {
		result.Status = "partial"
		result.Failures = append(result.Failures, domain.ProjectDiscoveryFailure{Root: root, Reason: reason})
	}
	add := func(root, name, repoID, source, parent string) bool {
		canonical, e := CanonicalRoot(root)
		if e != nil {
			fail(root, "registered_root_unavailable")
			return false
		}
		key := rootKey(canonical)
		if _, found := byRoot[key]; found {
			return true
		}
		if len(result.Projects) >= 256 {
			fail(root, "project_limit_reached")
			return false
		}
		if name == "" {
			name = filepath.Base(canonical)
		}
		if len(name) > 200 {
			name = filepath.Base(canonical)
		}
		byRoot[key] = len(result.Projects)
		result.Projects = append(result.Projects, domain.RegisteredProject{Root: canonical, Name: name, RepoID: repoID, Source: source, ParentRoot: parent,
			Git:         domain.ProjectGitObservation{Status: "unknown", Reason: "not_observed"},
			Activity:    domain.ProjectActivityObservation{Status: "unknown", Source: "orca_worktree_metadata"},
			Limitations: []string{"observation_only_no_context_or_execution_permission", "tracked_git_changes_only", "orca_activity_is_not_native_session_history"}})
		return true
	}
	for _, repo := range repos.Result.Repos {
		if ctx.Err() != nil {
			fail(repo.Path, "discovery_deadline")
			break
		}
		if !regexp.MustCompile(`^[A-Za-z0-9-]{1,80}$`).MatchString(repo.ID) {
			fail(repo.Path, "registration_id_invalid")
			continue
		}
		if !add(repo.Path, repo.Name, repo.ID, "orca_registered", "") {
			continue
		}
		canonical, _ := CanonicalRoot(repo.Path)
		byRepo[repo.ID] = canonical
		candidates, e := DiscoverProjects(ctx, canonical)
		if e != nil {
			fail(canonical, "bounded_subproject_discovery_incomplete")
		}
		for _, candidate := range candidates {
			if rootKey(candidate.Root) != rootKey(canonical) {
				add(candidate.Root, candidate.Name, repo.ID, "subproject", canonical)
			}
		}
	}
	// One bounded public listing rather than a command per repo. Remote rows do
	// not authorize filesystem reads on this machine.
	data, err = s.read(ctx, "orca_worktrees", "")
	if err != nil {
		fail("", "worktree_source_unavailable")
	} else {
		var trees struct {
			OK     bool `json:"ok"`
			Result struct {
				Worktrees []orcaRegisteredWorktree `json:"worktrees"`
				Truncated bool                     `json:"truncated"`
			} `json:"result"`
		}
		if json.Unmarshal(data, &trees) != nil || !trees.OK || trees.Result.Worktrees == nil || len(trees.Result.Worktrees) > 256 {
			fail("", "worktree_source_invalid")
		} else {
			if trees.Result.Truncated {
				fail("", "worktree_listing_truncated")
			}
			for _, tree := range trees.Result.Worktrees {
				parent, found := byRepo[tree.RepoID]
				if !found || tree.HostID != "local" {
					continue
				}
				if !add(tree.Path, tree.Name, tree.RepoID, "git_worktree", parent) {
					continue
				}
				canonical, _ := CanonicalRoot(tree.Path)
				index, found := byRoot[rootKey(canonical)]
				if !found {
					continue
				}
				a := &result.Projects[index].Activity
				if tree.LastActivityAt > 0 && tree.LastActivityAt <= now.Add(time.Minute).UnixMilli() {
					t := time.UnixMilli(tree.LastActivityAt).UTC()
					a.LastActivityAt = &t
					a.Status = "known"
				}
				for _, def := range definitions {
					if def.id == tree.CreatedWithAgent {
						cli := def.id
						a.CreatedWithCLI = &cli
						a.Status = "known"
						break
					}
				}
			}
		}
	}
	for i := range result.Projects {
		if ctx.Err() != nil {
			fail(result.Projects[i].Root, "discovery_deadline")
			continue
		}
		result.Projects[i].Git = s.observeGit(ctx, result.Projects[i].Root)
		if result.Projects[i].Git.Status == "unknown" {
			fail(result.Projects[i].Root, "git_observation_incomplete")
		}
	}
	sort.Slice(result.Projects, func(i, j int) bool { return rootKey(result.Projects[i].Root) < rootKey(result.Projects[j].Root) })
	return result, nil
}

func (s *RegisteredProjects) observeGit(ctx context.Context, root string) domain.ProjectGitObservation {
	g := domain.ProjectGitObservation{Status: "unknown", Reason: "git_observation_failed"}
	marker, e := os.Lstat(filepath.Join(root, ".git"))
	if e != nil {
		g.Status, g.Reason = "unavailable", "no_project_git_marker"
		return g
	}
	if marker.Mode()&os.ModeSymlink != 0 {
		g.Reason = "git_marker_symlink_excluded"
		return g
	}
	head, e := s.read(ctx, "git_head", root)
	if e != nil {
		g.Reason = "git_head_unavailable_or_unborn"
		return g
	}
	value := strings.TrimSpace(string(head))
	if !regexp.MustCompile(`^[0-9a-f]{40,64}$`).MatchString(value) {
		return g
	}
	g.Head = &value
	branch, e := s.read(ctx, "git_branch", root)
	if e == nil && len(branch) > 0 && len(branch) < 1024 {
		value := strings.TrimSpace(string(branch))
		g.Branch = &value
	}
	stamp, e := s.read(ctx, "git_time", root)
	if e != nil {
		return g
	}
	t, e := time.Parse(time.RFC3339, strings.TrimSpace(string(stamp)))
	if e != nil {
		return g
	}
	t = t.UTC()
	g.LastCommitAt = &t
	status, e := s.read(ctx, "git_status", root)
	if e != nil {
		return g
	}
	dirty := len(status) > 0
	g.Dirty = &dirty
	g.Status, g.Reason = "known", "tracked_changes_observed"
	return g
}
