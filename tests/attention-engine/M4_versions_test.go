package attention_engine_test

import (
	"context"
	"encoding/json"
	"path/filepath"
	"testing"

	"github.com/songconmaisaix31-design/Astrocyte/internal/attention/app"
	"github.com/songconmaisaix31-design/Astrocyte/internal/attention/domain"
)

func TestM4CrystallizedVersionsAndPinnedReferencesAfterRestart(t *testing.T) {
	ctx := context.Background()
	path := filepath.Join(t.TempDir(), "knowledge.sqlite")
	db, _, service := openM4(t, path)
	defer func() { db.Close() }()
	source := seedM4Source(t, db, "source", "p1")
	if _, err := service.RewriteFluid(ctx, app.KnowledgeFluidUpdate{ID: "k", OwnerProject: "p1", Topic: "agents", Fluid: "first", Sources: []domain.EventRef{source}, TS: 2000}); err != nil {
		t.Fatal(err)
	}
	first := json.RawMessage(`{"fact":"one"}`)
	if _, err := service.AppendCrystallized(ctx, app.KnowledgeCrystallizedUpdate{ID: "k", OwnerProject: "p1", Content: first, TS: 3000, ExpectedVersion: 1, Policy: app.KnowledgeKeepConfirmation}); err != nil {
		t.Fatal(err)
	}
	first[9] = 'x'
	human := app.Principal{ID: "human", Kind: "human"}
	if _, err := service.ConfirmKnowledge(ctx, human, "k", app.ConfirmKnowledgeCommand{CommandMeta: m4Meta(2, "confirm"), OwnerProject: "p1"}); err != nil {
		t.Fatal(err)
	}
	if _, err := service.ReferenceKnowledge(ctx, human, "k", app.ReferenceKnowledgeCommand{CommandMeta: m4Meta(3, "pin-first"), ProjectID: "p1"}); err != nil {
		t.Fatal(err)
	}
	if _, err := service.RewriteFluid(ctx, app.KnowledgeFluidUpdate{ID: "k", OwnerProject: "p1", Topic: "agents", Fluid: "second", Sources: []domain.EventRef{source}, TS: 6000, ExpectedVersion: 4}); err != nil {
		t.Fatal(err)
	}
	if _, err := service.AppendCrystallized(ctx, app.KnowledgeCrystallizedUpdate{ID: "k", OwnerProject: "p1", Content: json.RawMessage(`{"fact":"two"}`), TS: 7000, ExpectedVersion: 5}); err == nil {
		t.Fatal("unanswered policy was defaulted")
	}
	if _, err := service.AppendCrystallized(ctx, app.KnowledgeCrystallizedUpdate{ID: "k", OwnerProject: "p1", Content: json.RawMessage(`{"fact":"two"}`), TS: 7000, ExpectedVersion: 5, Policy: app.KnowledgeKeepConfirmation}); err != nil {
		t.Fatal(err)
	}
	if err := db.Close(); err != nil {
		t.Fatal(err)
	}
	db, _, service = openM4(t, path)
	entry, err := service.GetKnowledge(ctx, human, "k", "p1")
	if err != nil || entry.Entry.Fluid != "second" || entry.Entry.Status != domain.KnowledgeConfirmed || len(entry.CrystallizedVersions) != 2 {
		t.Fatalf("reopen: %+v %v", entry, err)
	}
	if _, err = db.Conn().Exec(`UPDATE attention_engine_knowledge_versions SET content='{}' WHERE entry_id='k' AND version=1`); err == nil {
		t.Fatal("SQLite accepted version overwrite")
	}
	if _, err = db.Conn().Exec(`DELETE FROM attention_engine_knowledge_versions WHERE entry_id='k' AND version=1`); err == nil {
		t.Fatal("SQLite accepted version deletion")
	}
	if _, err = db.Conn().Exec(`UPDATE attention_engine_knowledge_references SET crystallized_version=2 WHERE entry_id='k'`); err == nil {
		t.Fatal("SQLite accepted pinned reference overwrite")
	}
	if _, err = db.Conn().Exec(`PRAGMA query_only=ON`); err != nil {
		t.Fatal(err)
	}
	rows, err := db.Conn().Query(`SELECT version,content,recorded_at FROM attention_engine_knowledge_versions WHERE entry_id='k' ORDER BY version`)
	if err != nil {
		t.Fatal(err)
	}
	for i, want := range []string{`{"fact":"one"}`, `{"fact":"two"}`} {
		if !rows.Next() {
			t.Fatal("missing SQLite version")
		}
		var version int
		var content string
		var at int64
		if err = rows.Scan(&version, &content, &at); err != nil {
			t.Fatal(err)
		}
		wantAt := []int64{3000, 7000}[i]
		if version != i+1 || content != want || at != wantAt {
			t.Fatalf("immutable SQLite version: %d %s %d", version, content, at)
		}
	}
	if rows.Next() {
		t.Fatal("unexpected SQLite version")
	}
	if err = rows.Err(); err != nil {
		t.Fatal(err)
	}
	rows.Close()
	var project string
	var knowledgeVersion, crystallizedVersion int
	if err = db.Conn().QueryRow(`SELECT project_id,knowledge_version,crystallized_version FROM attention_engine_knowledge_references WHERE entry_id='k' AND request_id='pin-first'`).Scan(&project, &knowledgeVersion, &crystallizedVersion); err != nil {
		t.Fatal(err)
	}
	if project != "p1" || knowledgeVersion != 3 || crystallizedVersion != 1 {
		t.Fatalf("pin changed: %s %d %d", project, knowledgeVersion, crystallizedVersion)
	}
}

func TestM4ExplicitReconfirmationPolicyAndFluidPreservation(t *testing.T) {
	ctx := context.Background()
	db, _, service := openM4(t, filepath.Join(t.TempDir(), "knowledge.sqlite"))
	defer db.Close()
	source := seedM4Source(t, db, "source", "p1")
	if _, err := service.RewriteFluid(ctx, app.KnowledgeFluidUpdate{ID: "k", OwnerProject: "p1", Topic: "agents", Fluid: "first", Sources: []domain.EventRef{source}, TS: 2000}); err != nil {
		t.Fatal(err)
	}
	human := app.Principal{ID: "human", Kind: "human"}
	if _, err := service.ConfirmKnowledge(ctx, human, "k", app.ConfirmKnowledgeCommand{CommandMeta: m4Meta(1, "confirm"), OwnerProject: "p1"}); err != nil {
		t.Fatal(err)
	}
	if _, err := service.RewriteFluid(ctx, app.KnowledgeFluidUpdate{ID: "k", OwnerProject: "p1", Topic: "agents", Fluid: "second", Sources: []domain.EventRef{source}, TS: 6000, ExpectedVersion: 2}); err != nil {
		t.Fatal(err)
	}
	got, err := service.GetKnowledge(ctx, human, "k", "p1")
	if err != nil || got.Entry.Status != domain.KnowledgeConfirmed {
		t.Fatalf("fluid lost confirmation: %+v %v", got, err)
	}
	if _, err = service.AppendCrystallized(ctx, app.KnowledgeCrystallizedUpdate{ID: "k", OwnerProject: "p1", Content: json.RawMessage(`{"fact":"new"}`), TS: 7000, ExpectedVersion: 3, Policy: app.KnowledgeReconfirmVersion}); err != nil {
		t.Fatal(err)
	}
	if _, err = service.ReferenceKnowledge(ctx, human, "k", app.ReferenceKnowledgeCommand{CommandMeta: m4Meta(4, "unconfirmed-new-version"), ProjectID: "p1"}); err == nil {
		t.Fatal("new version did not require explicit confirmation")
	}
}
