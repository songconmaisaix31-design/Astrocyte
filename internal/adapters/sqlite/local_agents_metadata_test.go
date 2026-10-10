package sqlite

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	"github.com/songconmaisaix31-design/Astrocyte/internal/adapters/agents"
	"github.com/songconmaisaix31-design/Astrocyte/internal/apierrors"
	workspace "github.com/songconmaisaix31-design/Astrocyte/internal/workspace/app"
	"github.com/songconmaisaix31-design/Astrocyte/internal/workspace/domain"
)

func TestProjectMetadataPersistsThroughRefreshArchiveAndProcessRestart(t *testing.T) {
	ctx := context.Background()
	human := domain.Caller{Kind: "human", ID: "h"}
	if path := os.Getenv("ASTROCYTE_METADATA_RESTART_DB"); path != "" {
		db := openAttentionDB(t, path)
		service := workspace.NewLocalProjectService(db, nil, nil, nil)
		service.ConfigureRegisteredDiscovery(db, nil)
		snapshot, err := service.ListRegisteredProjects(ctx, human)
		if err != nil || len(snapshot.Board) != 1 || snapshot.Board[0].Human.Notes != "human notes" || snapshot.Board[0].Human.Review != "human review" || !snapshot.Board[0].Human.Archived || snapshot.Board[0].Human.Revision != 1 {
			t.Fatalf("new process lost metadata %+v %v", snapshot, err)
		}
		return
	}
	dbPath := filepath.Join(t.TempDir(), "state.sqlite")
	db := openAttentionDB(t, dbPath)
	stamp := time.Now().UTC()
	source := &registeredSourceFixture{snapshot: domain.ProjectDiscoverySnapshot{Status: "complete", ObservedAt: &stamp, Projects: []domain.RegisteredProject{{ProjectKey: "git:observed-family", Root: t.TempDir(), Name: "main", RepoID: "real-registered-family", Source: "orca_registered"}, {ProjectKey: "git:observed-family", Root: t.TempDir(), Name: "tree", RepoID: "real-registered-family", Source: "git_worktree"}}}}
	service := workspace.NewLocalProjectService(db, nil, nil, nil)
	service.ConfigureRegisteredDiscovery(db, source)
	snapshot, err := service.RefreshRegisteredProjects(ctx, human)
	if err != nil || len(snapshot.Board) != 1 || len(snapshot.Board[0].Roots) != 2 {
		t.Fatalf("aggregation %+v %v", snapshot, err)
	}
	id := snapshot.Board[0].ID
	metadata := domain.ProjectHumanMetadata{Notes: "human notes", Review: "human review", Group: "research", Intent: "await user progress decision", Archived: true}
	for _, caller := range []domain.Caller{{Kind: "agent", ID: "a", ProjectID: id}, {Kind: "human"}, {}} {
		if _, err := service.SetRegisteredProjectMetadata(ctx, caller, id, metadata); err == nil {
			t.Fatal("Agent/anonymous annotation write accepted")
		}
	}
	p, err := service.SetRegisteredProjectMetadata(ctx, human, id, metadata)
	if err != nil || p.Human.Revision != 1 || p.Human.UpdatedAt == nil {
		t.Fatalf("metadata %+v %v", p, err)
	}
	if _, err := service.SetRegisteredProjectMetadata(ctx, human, id, metadata); err == nil {
		t.Fatal("stale overwrite accepted")
	}
	metadata.UpdatedAt = &stamp
	if _, err := service.SetRegisteredProjectMetadata(ctx, human, id, metadata); err == nil {
		t.Fatal("forged timestamp accepted")
	}
	for i := 0; i < 2; i++ {
		source.snapshot.Projects[0].Name = "fresh observed name"
		source.snapshot.Projects[0].ProjectKey = "" // Transient Git identity failure.
		snapshot, err = service.RefreshRegisteredProjects(ctx, human)
		if err != nil || snapshot.Board[0].Human.Notes != "human notes" || !snapshot.Board[0].Human.Archived || snapshot.Board[0].Human.Revision != 1 {
			t.Fatalf("refresh overwrote fields %+v %v", snapshot, err)
		}
	}
	calls := source.calls
	if _, err := service.ListRegisteredProjects(ctx, human); err != nil || source.calls != calls {
		t.Fatal("cached GET scanned")
	}
	if err := db.Close(); err != nil {
		t.Fatal(err)
	}
	command := exec.Command(os.Args[0], "-test.run=^TestProjectMetadataPersistsThroughRefreshArchiveAndProcessRestart$", "-test.v")
	command.Env = append(os.Environ(), "ASTROCYTE_METADATA_RESTART_DB="+dbPath)
	if output, err := command.CombinedOutput(); err != nil {
		t.Fatalf("fresh OS process %+v %s", err, output)
	}
	db = openAttentionDB(t, dbPath)
	service = workspace.NewLocalProjectService(db, nil, nil, nil)
	service.ConfigureRegisteredDiscovery(db, source)
	snapshot, err = service.ListRegisteredProjects(ctx, human)
	if err != nil {
		t.Fatal(err)
	}
	metadata = snapshot.Board[0].Human
	metadata.UpdatedAt = nil
	metadata.Archived = false
	p, err = service.SetRegisteredProjectMetadata(ctx, human, id, metadata)
	if err != nil || p.Human.Archived || p.Human.Revision != 2 || p.Human.Notes != "human notes" {
		t.Fatalf("unarchive %+v %v", p, err)
	}
	projects, err := db.ListProjects(ctx)
	if err != nil || len(projects) != 0 {
		t.Fatal("annotation created grants/execution project")
	}
	if _, err := db.LoadGrant(ctx, id, "a"); err == nil {
		t.Fatal("annotation granted Agent")
	}
}

func TestActualRegisteredBoardSQLite(t *testing.T) {
	if os.Getenv("ASTROCYTE_TEST_REGISTERED_BOARD") != "1" {
		t.Skip("explicit approved-root metadata refresh opt-in")
	}
	ctx := context.Background()
	db := openAttentionDB(t, filepath.Join(t.TempDir(), "state.sqlite"))
	service := workspace.NewLocalProjectService(db, nil, nil, nil)
	service.ConfigureRegisteredDiscovery(db, agents.NewRegisteredProjects())
	human := domain.Caller{Kind: "human", ID: "actual-local-human"}
	snapshot, err := service.RefreshRegisteredProjects(ctx, human)
	if err != nil || len(snapshot.Board) == 0 {
		t.Fatalf("actual refresh %+v %v", snapshot, err)
	}
	contributors, duplicateGroups, multipleClients, standalone := 0, 0, 0, 0
	for _, p := range snapshot.Board {
		clients := map[string]bool{}
		for _, c := range p.Contributors {
			clients[c.CLI] = true
		}
		if len(clients) > 1 {
			multipleClients++
		}
		contributors += len(p.Contributors)
		if len(p.Roots) > 1 {
			duplicateGroups++
		}
	}
	for _, p := range snapshot.Projects {
		if p.Source == "native_project_metadata" {
			standalone++
		}
	}
	t.Logf("actual status=%s roots=%d board=%d groups_with_multiple_roots=%d historical_contributors=%d multiple_client_groups=%d standalone_metadata_roots=%d", snapshot.Status, len(snapshot.Projects), len(snapshot.Board), duplicateGroups, contributors, multipleClients, standalone)
	for _, s := range snapshot.Sources {
		t.Logf("scope source=%s cli=%s status=%s reason=%s examined=%d matched=%d", s.Source, s.CLI, s.Status, s.Reason, s.EntriesExamined, s.MatchedHeaders)
	}
	for _, f := range snapshot.Failures {
		t.Logf("partial root=%s reason=%s", f.Root, f.Reason)
	}
	p, err := service.SetRegisteredProjectMetadata(ctx, human, snapshot.Board[0].ID, domain.ProjectHumanMetadata{Notes: "actual SQLite refresh note", Review: "actual storage review", Archived: true})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := service.SetRegisteredProjectMetadata(ctx, domain.Caller{Kind: "agent", ID: "a", ProjectID: p.ID}, p.ID, domain.ProjectHumanMetadata{}); err == nil {
		t.Fatal("actual store Agent write accepted")
	} else if e, ok := err.(*apierrors.ServiceError); !ok || e.Code != apierrors.ScopeDenied {
		t.Fatalf("wrong denial %v", err)
	}
	snapshot, err = service.RefreshRegisteredProjects(ctx, human)
	if err != nil {
		t.Fatal(err)
	}
	retained := false
	for _, item := range snapshot.Board {
		if item.ID == p.ID {
			retained = item.Human.Notes == p.Human.Notes && item.Human.Archived
		}
	}
	if !retained {
		t.Fatal("actual second refresh lost metadata")
	}
	t.Log("actual second source refresh retained note/review/archive without grants or native launch")
}
