package main

import (
	"context"
	"encoding/json"
	"path/filepath"
	"testing"

	"github.com/songconmaisaix31-design/Astrocyte/internal/adapters/sqlite"
	attentionapp "github.com/songconmaisaix31-design/Astrocyte/internal/attention/app"
	attentiondomain "github.com/songconmaisaix31-design/Astrocyte/internal/attention/domain"
	workspaceapp "github.com/songconmaisaix31-design/Astrocyte/internal/workspace/app"
	workspacedomain "github.com/songconmaisaix31-design/Astrocyte/internal/workspace/domain"
	"github.com/songconmaisaix31-design/Astrocyte/migrations"
)

type experienceProjects struct{ workspaceapp.LocalProjects }

func (*experienceProjects) ListProjects(context.Context, workspacedomain.Caller) ([]workspacedomain.LocalProject, error) {
	return []workspacedomain.LocalProject{{ID: "project-a"}, {ID: "project-b"}}, nil
}

func TestExperienceHTTPFactoryChecksActualContextAndOwner(t *testing.T) {
	ctx := context.Background()
	db, err := sqlite.Open(filepath.Join(t.TempDir(), "state.sqlite"))
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	if err = db.RunMigrations(ctx, migrations.FS); err != nil {
		t.Fatal(err)
	}
	events := sqlite.NewEventRepository(db)
	rows, err := attentionapp.NewEventService(events, 1).Record(ctx, []attentionapp.StoredEvent{{OriginKey: "fixture:failed-task", RawBytes: []byte(" {\"outcome\":\"failed\"} \n"), Observation: attentionapp.Observation{Source: "codex", SourceKind: "agent_task", Scope: "project-a", TS: 100, Outcome: "failed", Depth: "unknown", Provenance: "contract_local"}, Event: &attentiondomain.Event{Source: "codex", Scope: "project-a", Kind: "agent_task", TS: 100, Raw: json.RawMessage(`{"outcome":"failed"}`)}}})
	if err != nil {
		t.Fatal(err)
	}
	experience := attentionapp.NewExperienceService(sqlite.NewExperienceRepository(db), events, attentionapp.ExperienceOptions{ProjectID: "project-a"})
	captured, err := experience.CaptureFailure(ctx, rows[0].ID)
	if err != nil || captured == nil {
		t.Fatalf("capture: %+v %v", captured, err)
	}
	bridge := &experienceHTTPBridge{db: db, events: events, projects: &experienceProjects{}}
	p := attentionapp.Principal{ID: "local-human", Kind: "human"}
	cmd := attentionapp.ApproveExperienceCommand{CommandMeta: attentionapp.CommandMeta{SchemaVersion: 1, RequestID: "approve", IdempotencyKey: "approve", ExpectedVersion: 1}, OwnerProject: "unknown"}
	if _, err = bridge.ApproveExperience(ctx, p, captured.Experience.ID, cmd); err == nil {
		t.Fatal("unregistered project authorized by body")
	}
	cmd.OwnerProject = "project-b"
	if _, err = bridge.ApproveExperience(ctx, p, captured.Experience.ID, cmd); err == nil {
		t.Fatal("different registered project authorized owner mutation")
	}
	cmd.OwnerProject = "project-a"
	if _, err = bridge.ApproveExperience(ctx, attentionapp.Principal{ID: "agent", Kind: "agent"}, captured.Experience.ID, cmd); err == nil {
		t.Fatal("Agent human HTTP factory access")
	}
	approved, err := bridge.ApproveExperience(ctx, p, captured.Experience.ID, cmd)
	if err != nil {
		t.Fatal(err)
	}
	share := attentionapp.ShareMemoryCommand{CommandMeta: attentionapp.CommandMeta{SchemaVersion: 1, RequestID: "share", IdempotencyKey: "share", ExpectedVersion: approved.Version}, OwnerProject: "project-a", TargetProject: "unknown"}
	if _, err = bridge.ShareExperience(ctx, p, captured.Experience.ID, share); err == nil {
		t.Fatal("unregistered share target authorized by body")
	}
	share.TargetProject = "project-b"
	if _, err = bridge.ShareExperience(ctx, p, captured.Experience.ID, share); err != nil {
		t.Fatal(err)
	}
	visible, err := bridge.GetExperience(ctx, p, captured.Experience.ID, "project-b")
	if err != nil || visible.OwnerProject != "project-a" || len(visible.Shares) != 1 {
		t.Fatalf("explicit share: %+v %v", visible, err)
	}
}
