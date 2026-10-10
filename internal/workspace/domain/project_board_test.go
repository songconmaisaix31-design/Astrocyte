package domain

import (
	"testing"
	"time"
)

func TestProjectBoardIdentityAndHistoricalContributors(t *testing.T) {
	stamp := time.Now().UTC()
	created := "claude"
	projects := []RegisteredProject{
		{Root: "/project", Name: "main", RepoID: "r", Source: "orca_registered", Activity: ProjectActivityObservation{CreatedWithCLI: &created}},
		{Root: "/tree", Name: "worker", RepoID: "r", Source: "git_worktree", Contributors: []ProjectContributor{{CLI: "codex", Source: "native_session_header", SessionID: "s", CreatedAt: &stamp}}},
		{Root: "/project/nested", Name: "nested", RepoID: "r", Source: "subproject"},
		{Root: "/clone", Name: "main", RepoID: "different", Source: "orca_registered"},
	}
	board := AggregateProjects(append(projects, projects[0]))
	if len(board) != 3 {
		t.Fatalf("duplicate/nested/clone identity wrong: %+v", board)
	}
	for _, p := range board {
		if len(p.Roots) == 2 {
			if p.Name != "main" || len(p.Contributors) != 1 || p.Contributors[0].CLI != "codex" || p.LastActivityAt != nil {
				t.Fatalf("created-with inferred contributor: %+v", p)
			}
		}
	}
	projects[0].ProjectKey = "git:common"
	projects[1].ProjectKey = "git:common"
	projects[1].RepoID = "independent-registration"
	if len(AggregateProjects(projects)) != 3 {
		t.Fatal("observed Git common-dir did not join registrations")
	}
}

func TestProjectBoardLatestRecordedActivity(t *testing.T) {
	older := time.Date(2026, 10, 8, 10, 0, 0, 0, time.UTC)
	newer := older.Add(24 * time.Hour)
	created := newer.Add(24 * time.Hour)
	header := ProjectContributor{CLI: "codex", Source: "native_session_header", CreatedAt: &created, ActivityAt: &created}
	for _, tt := range []struct {
		name string
		git  ProjectGitObservation
		orca *time.Time
		want *time.Time
	}{
		{"native_only_known_commit", ProjectGitObservation{Status: "known", LastCommitAt: &older}, nil, &older},
		{"git_newer_than_orca", ProjectGitObservation{Status: "known", LastCommitAt: &newer}, &older, &newer},
		{"orca_newer_than_git", ProjectGitObservation{Status: "known", LastCommitAt: &older}, &newer, &newer},
		{"equal_observations", ProjectGitObservation{Status: "known", LastCommitAt: &older}, &older, &older},
		{"commit_known_status_query_failed", ProjectGitObservation{Status: "unknown", LastCommitAt: &older}, nil, &older},
		{"unknown_commit_creation_only", ProjectGitObservation{Status: "unknown"}, nil, nil},
		{"known_git_nil_commit_creation_only", ProjectGitObservation{Status: "known"}, nil, nil},
		{"unavailable_git_creation_only", ProjectGitObservation{Status: "unavailable"}, nil, nil},
	} {
		t.Run(tt.name, func(t *testing.T) {
			projects := []RegisteredProject{{Root: "/native", Source: "native_session_root", Git: tt.git, Activity: ProjectActivityObservation{LastActivityAt: tt.orca}, Contributors: []ProjectContributor{header}}}
			board := AggregateProjects(projects)
			if len(board) != 1 {
				t.Fatalf("expected one project: %+v", board)
			}
			got := board[0].LastActivityAt
			if (tt.want == nil) != (got == nil) || (got != nil && !got.Equal(*tt.want)) {
				t.Fatalf("latest activity = %v, want %v", got, tt.want)
			}
			if !board[0].Contributors[0].CreatedAt.Equal(created) {
				t.Fatal("session creation time changed")
			}
		})
	}
}

func TestProjectBoardLatestActivityAcrossMemberDirectories(t *testing.T) {
	old := time.Date(2026, 10, 8, 10, 0, 0, 0, time.UTC)
	middle := old.Add(time.Hour)
	latest := middle.Add(time.Hour)
	projects := []RegisteredProject{
		{Root: "/main", ProjectKey: "git:common", Source: "orca_registered", Activity: ProjectActivityObservation{LastActivityAt: &middle}},
		{Root: "/older-tree", ProjectKey: "git:common", Git: ProjectGitObservation{LastCommitAt: &old}},
		{Root: "/newer-tree", ProjectKey: "git:common", Git: ProjectGitObservation{LastCommitAt: &latest}},
	}
	for i := 0; i < len(projects); i++ {
		board := AggregateProjects(projects)
		if len(board) != 1 || len(board[0].Roots) != 3 || board[0].LastActivityAt == nil || !board[0].LastActivityAt.Equal(latest) {
			t.Fatalf("member order %d lost latest commit: %+v", i, board)
		}
		projects = append(projects[1:], projects[0])
	}
}
