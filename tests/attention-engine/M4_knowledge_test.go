package attention_engine_test

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/songconmaisaix31-design/Astrocyte/internal/adapters/sqlite"
	"github.com/songconmaisaix31-design/Astrocyte/internal/attention/app"
	"github.com/songconmaisaix31-design/Astrocyte/internal/attention/domain"
	"github.com/songconmaisaix31-design/Astrocyte/migrations"
)

func m4Path(t *testing.T) string {
	t.Helper()
	root := os.Getenv("ASTROCYTE_M4_ARTIFACT_DIR")
	if root == "" {
		return filepath.Join(t.TempDir(), "knowledge.sqlite")
	}
	if !filepath.IsAbs(root) {
		t.Fatal("M4 artifact directory must be absolute")
	}
	dir := filepath.Join(root, t.Name())
	if err := os.MkdirAll(dir, 0700); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(dir, "knowledge.sqlite")
	if _, err := os.Stat(path); err == nil {
		t.Fatal("M4 artifact database already exists; select a fresh artifact directory")
	}
	return path
}

func m4Evidence(t *testing.T, path string, value any) {
	t.Helper()
	if os.Getenv("ASTROCYTE_M4_ARTIFACT_DIR") == "" {
		return
	}
	data, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	if err = os.WriteFile(filepath.Join(filepath.Dir(path), "retrieval.json"), data, 0600); err != nil {
		t.Fatal(err)
	}
}

type m4Repository struct {
	app.KnowledgeRepository
	t *testing.T
}

func (r m4Repository) WithKnowledgeTx(ctx context.Context, fn func(app.KnowledgeTx) error) error {
	err := r.KnowledgeRepository.WithKnowledgeTx(ctx, fn)
	if err != nil {
		r.t.Logf("knowledge transaction: %v", err)
	}
	return err
}

func openM4(t *testing.T, path string) (*sqlite.DB, *sqlite.EngineRepository, *app.KnowledgeService) {
	t.Helper()
	db, err := sqlite.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	if err = db.RunMigrations(context.Background(), migrations.FS); err != nil {
		db.Close()
		t.Fatal(err)
	}
	base := sqlite.NewEngineRepository(db)
	service := app.NewKnowledgeService(m4Repository{KnowledgeRepository: sqlite.NewKnowledgeRepository(base), t: t}, app.KnowledgeServiceOptions{Clock: func() time.Time { return time.UnixMilli(5000) }})
	return db, base, service
}

func seedM4Source(t *testing.T, db *sqlite.DB, id, project string) domain.EventRef {
	t.Helper()
	event := domain.Event{Source: "manual", Kind: "collect", Topic: "agents", Scope: project, TS: 1000, Raw: json.RawMessage(`{"fact":"original"}`)}
	data, _ := json.Marshal(event)
	observation, _ := json.Marshal(app.Observation{Source: "manual", Scope: project, TS: 1000, Provenance: "M4 controlled acceptance"})
	_, err := db.Conn().Exec(`INSERT INTO attention_engine_events(id,source,ts,kind,topic,data,origin_key,received_at,raw_bytes,observation,mapping_status) VALUES(?,?,?,?,?,?,?,?,?,?,?)`, id, "manual", 1000, "collect", "agents", string(data), "m4:"+id, 1000, []byte(`{"fact":"original"}`), string(observation), "mapped")
	if err != nil {
		t.Fatal(err)
	}
	return domain.EventRef{EventID: id, TS: 1000}
}

func m4Meta(version int, request string) app.CommandMeta {
	return app.CommandMeta{SchemaVersion: 1, ExpectedVersion: version, RequestID: request, IdempotencyKey: request}
}

func TestM4FluidRestartAndSharedTransaction(t *testing.T) {
	ctx := context.Background()
	path := m4Path(t)
	db, base, service := openM4(t, path)
	defer func() { db.Close() }()
	source := seedM4Source(t, db, "source", "p1")
	update := app.KnowledgeFluidUpdate{ID: "k", OwnerProject: "p1", Topic: "agents", Fluid: "first", Sources: []domain.EventRef{source}, TS: 2000}
	if _, err := service.RewriteFluid(ctx, update); err != nil {
		t.Fatal(err)
	}
	update.Fluid, update.ExpectedVersion, update.TS = "second", 1, 3000
	if err := base.WithTx(ctx, func(tx sqlite.EngineTx) error {
		_, err := app.RewriteKnowledgeFluidTx(sqlite.BindKnowledgeTx(tx), update)
		return err
	}); err != nil {
		t.Fatal(err)
	}
	update.Fluid, update.ExpectedVersion, update.TS = "rolled back", 2, 4000
	if err := base.WithTx(ctx, func(tx sqlite.EngineTx) error {
		if _, err := app.RewriteKnowledgeFluidTx(sqlite.BindKnowledgeTx(tx), update); err != nil {
			return err
		}
		return domain.ErrEvidence
	}); err != domain.ErrEvidence {
		t.Fatalf("rollback result: %v", err)
	}
	if err := db.Close(); err != nil {
		t.Fatal(err)
	}
	db, _, service = openM4(t, path)
	got, err := service.GetKnowledge(ctx, app.Principal{ID: "human", Kind: "human"}, "k", "p1")
	if err != nil || got.Entry.Fluid != "second" || got.Version != 2 {
		t.Fatalf("restart: %+v %v", got, err)
	}
	if _, err = db.Conn().Exec(`PRAGMA query_only=ON`); err != nil {
		t.Fatal(err)
	}
	var fluid string
	var count int
	if err = db.Conn().QueryRow(`SELECT json_extract(data,'$.fluid') FROM attention_engine_knowledge WHERE id='k'`).Scan(&fluid); err != nil || fluid != "second" {
		t.Fatalf("SQLite fluid=%q err=%v", fluid, err)
	}
	if err = db.Conn().QueryRow(`SELECT count(*) FROM attention_engine_knowledge WHERE id='k'`).Scan(&count); err != nil || count != 1 {
		t.Fatalf("single mutable row=%d err=%v", count, err)
	}
}

func TestM4HumanConfirmationScopeAndShareAudit(t *testing.T) {
	ctx := context.Background()
	db, _, service := openM4(t, m4Path(t))
	defer db.Close()
	source := seedM4Source(t, db, "source", "p1")
	if _, err := service.RewriteFluid(ctx, app.KnowledgeFluidUpdate{ID: "k", OwnerProject: "p1", Topic: "agents", Fluid: "summary", Sources: []domain.EventRef{source}, TS: 2000}); err != nil {
		t.Fatal(err)
	}
	human, agent := app.Principal{ID: "human", Kind: "human"}, app.Principal{ID: "agent", Kind: "agent"}
	if _, err := service.ReferenceKnowledge(ctx, human, "k", app.ReferenceKnowledgeCommand{CommandMeta: m4Meta(1, "before-confirm"), ProjectID: "p1"}); err == nil {
		t.Fatal("inbox was referenceable")
	}
	if _, err := service.GetKnowledge(ctx, human, "k", "p2"); err == nil {
		t.Fatal("cross-project entry visible without share")
	}
	confirm := app.ConfirmKnowledgeCommand{CommandMeta: m4Meta(1, "confirm"), OwnerProject: "p1"}
	if _, err := service.ConfirmKnowledge(ctx, agent, "k", confirm); err == nil {
		t.Fatal("Agent confirmation accepted")
	}
	confirmed, err := service.ConfirmKnowledge(ctx, human, "k", confirm)
	if err != nil || confirmed.Entry.Status != domain.KnowledgeConfirmed {
		t.Fatalf("confirm: %+v %v", confirmed, err)
	}
	if again, err := service.ConfirmKnowledge(ctx, human, "k", confirm); err != nil || again.Version != confirmed.Version {
		t.Fatalf("idempotent confirm: %+v %v", again, err)
	}
	confirm.RequestID = "confirm-retry-trace"
	if again, err := service.ConfirmKnowledge(ctx, human, "k", confirm); err != nil || again.Version != confirmed.Version {
		t.Fatalf("new tracing ID changed semantic retry: %+v %v", again, err)
	}
	share := app.ShareMemoryCommand{CommandMeta: m4Meta(2, "share"), OwnerProject: "p1", TargetProject: "p2"}
	if _, err := service.ShareKnowledge(ctx, agent, "k", share); err == nil {
		t.Fatal("Agent sharing accepted")
	}
	if _, err := service.ShareKnowledge(ctx, human, "k", share); err != nil {
		t.Fatal(err)
	}
	got, err := service.GetKnowledge(ctx, human, "k", "p2")
	if err != nil || len(got.Shares) != 1 {
		t.Fatalf("explicit share: %+v %v", got, err)
	}
	audit := got.Shares[0]
	if audit.ActorID != "human" || audit.ActorKind != "human" || audit.OwnerProject != "p1" || audit.TargetProject != "p2" || audit.TS != 5000 || audit.RequestID != "share" {
		t.Fatalf("share audit: %+v", audit)
	}
	if _, err = service.GetKnowledge(ctx, human, "k", "p3"); err == nil {
		t.Fatal("sharing widened unrelated project")
	}
	if _, err = service.ReferenceKnowledge(ctx, human, "k", app.ReferenceKnowledgeCommand{CommandMeta: m4Meta(3, "reference"), ProjectID: "p2"}); err != nil {
		t.Fatal(err)
	}
	if _, err = service.DemoteKnowledge(ctx, human, "k", app.DemoteMemoryCommand{CommandMeta: m4Meta(4, "demote"), OwnerProject: "p1"}); err != nil {
		t.Fatal(err)
	}
	if _, err = service.ReferenceKnowledge(ctx, human, "k", app.ReferenceKnowledgeCommand{CommandMeta: m4Meta(5, "after-demote"), ProjectID: "p2"}); err == nil {
		t.Fatal("demoted entry remained referenceable")
	}
}
