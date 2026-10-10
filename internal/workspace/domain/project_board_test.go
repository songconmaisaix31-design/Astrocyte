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
