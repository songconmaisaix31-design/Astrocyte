package attentionengine_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/songconmaisaix31-design/Astrocyte/internal/adapters/sqlite"
	"github.com/songconmaisaix31-design/Astrocyte/internal/adapters/worksscan"
	workspaceapp "github.com/songconmaisaix31-design/Astrocyte/internal/workspace/app"
	"github.com/songconmaisaix31-design/Astrocyte/internal/workspace/domain"
)

// W7 acceptance: real local works aggregation, durable metadata across a cold
// restart, and works activity emitted as a chain-1 event source. No network, no
// model, no Agent, no whole-disk scan.

type eventCollector struct{ events []domain.WorkEvent }

func (c *eventCollector) EmitWorkEvent(_ context.Context, e domain.WorkEvent) error {
	c.events = append(c.events, e)
	return nil
}

func openWorksDB(t *testing.T) (*sqlite.DB, string) {
	t.Helper()
	path := filepath.Join(t.TempDir(), "state.sqlite")
	db, err := sqlite.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { db.Close() })
	for _, ddl := range []string{
		"CREATE TABLE works_album (key TEXT PRIMARY KEY, data TEXT NOT NULL)",
		"CREATE TABLE works_metadata (work_id TEXT PRIMARY KEY, data TEXT NOT NULL, revision INTEGER NOT NULL CHECK (revision >= 1))",
	} {
		if _, err := db.Conn().Exec(ddl); err != nil {
			t.Fatal(err)
		}
	}
	return db, path
}

func writeMarker(t *testing.T, root, platform, name string) {
	t.Helper()
	if err := os.MkdirAll(root, 0o755); err != nil {
		t.Fatal(err)
	}
	content := `{"name":"` + name + `","platform":"` + platform + `"}`
	if err := os.WriteFile(filepath.Join(root, ".astrocyte-works.json"), []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "main.go"), []byte("package main"), 0o644); err != nil {
		t.Fatal(err)
	}
}

func TestW7WorksAggregationAcrossRestart(t *testing.T) {
	ctx := context.Background()
	db, path := openWorksDB(t)

	appRoot := t.TempDir()
	writeMarker(t, appRoot, "codex", "作品-app")
	reportRoot := t.TempDir()
	writeMarker(t, reportRoot, "lingxi", "作品-report")

	sink := &eventCollector{}
	service := workspaceapp.NewWorksService(db, worksscan.NewScanner(), worksscan.Delivery{}, sink, workspaceapp.WorksServiceOptions{})
	caller := domain.Caller{Kind: "human", ID: "acceptance"}

	// Capability 5: scan imports existing projects, aggregating by platform.
	snapshot, err := service.RefreshWorks(ctx, caller, workspaceapp.WorksRefreshCommand{Roots: []string{appRoot, reportRoot}})
	if err != nil {
		t.Fatal(err)
	}
	if len(snapshot.Works) != 2 {
		t.Fatalf("expected 2 scanned works, got %d", len(snapshot.Works))
	}
	if len(snapshot.Platforms) != 2 {
		t.Fatalf("expected 2 platforms, got %+v", snapshot.Platforms)
	}
	for _, work := range snapshot.Works {
		if !work.Files.Verified {
			t.Fatalf("scanned work must carry the file-verified badge: %+v", work)
		}
	}

	// Capability 5: AI 投递 JSON import, unverified until re-scanned.
	delivery := []byte(`{"works":[{"name":"笔记","platform":"workbuddy","path":"C:\\missing\\notes","handoff":"generated guide"}]}`)
	if _, err := service.ImportWorksDelivery(ctx, caller, workspaceapp.WorksDeliveryCommand{Content: delivery}); err != nil {
		t.Fatal(err)
	}
	afterImport, _ := service.ListWorks(ctx, caller)
	if len(afterImport.Works) != 3 {
		t.Fatalf("expected 3 works after import, got %d", len(afterImport.Works))
	}
	var imported domain.Work
	for _, work := range afterImport.Works {
		if work.Platform == "workbuddy" {
			imported = work
		}
	}
	if imported.Files.Verified {
		t.Fatal("JSON-delivered work must not be file-verified")
	}
	if imported.Handoff.Source != domain.HandoffSourceGenerated {
		t.Fatalf("delivered handoff must be generated: %+v", imported.Handoff)
	}

	// Capability 4: attitude, group and handoff are human metadata.
	id := domain.WorkIdentity("codex", appRoot)
	attitudeWork, err := service.SetWorkAttitude(ctx, caller, id, workspaceapp.WorkAttitudeCommand{Attitude: domain.AttitudeContinue})
	if err != nil {
		t.Fatal(err)
	}
	groupWork, err := service.SetWorkGroup(ctx, caller, id, workspaceapp.WorkGroupCommand{ExpectedRevision: attitudeWork.Revision, Group: "研究"})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := service.SetWorkHandoff(ctx, caller, id, workspaceapp.WorkHandoffCommand{ExpectedRevision: groupWork.Revision, Content: "下一步：接入预览"}); err != nil {
		t.Fatal(err)
	}
	if _, err := service.SetWorkHidden(ctx, caller, domain.WorkIdentity("lingxi", reportRoot), workspaceapp.WorkHiddenCommand{Hidden: true}); err != nil {
		t.Fatal(err)
	}

	if len(sink.events) == 0 {
		t.Fatal("works activity must be emitted as chain-1 events")
	}
	kinds := map[string]bool{}
	for _, event := range sink.events {
		kinds[event.Kind] = true
		if event.Scope == "" || event.Topic != "" || (event.Source != domain.WorkEventSourceManual && event.Source != domain.WorkEventSourceMachine) {
			t.Fatalf("event missing normalization fields: %+v", event)
		}
	}
	for _, kind := range []string{domain.WorkEventSynced, domain.WorkEventImported, domain.WorkEventAttitudeSet, domain.WorkEventHandoffSet, domain.WorkEventHidden, domain.WorkEventGroupSet} {
		if !kinds[kind] {
			t.Fatalf("missing %s event", kind)
		}
	}

	// Cold restart: the same database must yield the same album and metadata.
	if err := db.Close(); err != nil {
		t.Fatal(err)
	}
	db2, err := sqlite.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer db2.Close()
	restarted := workspaceapp.NewWorksService(db2, worksscan.NewScanner(), worksscan.Delivery{}, nil, workspaceapp.WorksServiceOptions{})
	snapshot2, err := restarted.ListWorks(ctx, caller)
	if err != nil {
		t.Fatal(err)
	}
	if len(snapshot2.Works) != 3 {
		t.Fatalf("cold restart lost works, got %d", len(snapshot2.Works))
	}
	var codexWork domain.Work
	for _, work := range snapshot2.Works {
		if work.Platform == "codex" {
			codexWork = work
		}
	}
	if codexWork.Attitude != domain.AttitudeContinue || codexWork.Group != "研究" || codexWork.Handoff.Content != "下一步：接入预览" {
		t.Fatalf("cold restart lost human metadata: %+v", codexWork)
	}
	// The hidden lingxi work must survive and stay hidden.
	var lingxiHidden bool
	for _, work := range snapshot2.Works {
		if work.Platform == "lingxi" {
			lingxiHidden = work.Hidden
		}
	}
	if !lingxiHidden {
		t.Fatal("cold restart lost hidden flag")
	}
}

func TestW7ApprovedLocalMetadataObservation(t *testing.T) {
	root := os.Getenv("ASTROCYTE_TEST_W7_APPROVED_ROOT")
	if root == "" {
		t.Skip("requires explicitly approved metadata-only root")
	}
	works, failures, err := worksscan.NewScanner().ScanWorks(context.Background(), []string{root})
	if err != nil {
		t.Fatal(err)
	}
	if len(works) != 1 {
		t.Fatalf("approved root not observed: %+v", failures)
	}
	if works[0].Root != filepath.Clean(root) || works[0].Files.SourcePath != filepath.Clean(root) {
		t.Fatalf("observation escaped root: %+v", works[0])
	}
	if works[0].Files.Mtime == nil {
		t.Fatal("actual stat time absent")
	}
	t.Logf("actual bounded root metadata: platform=%s verified=%v observed_files=%d reason=%s failures=%d", works[0].Platform, works[0].Files.Verified, works[0].Files.FileCount, works[0].Files.Reason, len(failures))
}
