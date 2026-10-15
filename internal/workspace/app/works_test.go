package app_test

import (
	"context"
	"encoding/json"
	"errors"
	"math"
	"os"
	"path/filepath"
	"testing"

	"github.com/songconmaisaix31-design/Astrocyte/internal/adapters/events"
	"github.com/songconmaisaix31-design/Astrocyte/internal/adapters/sqlite"
	"github.com/songconmaisaix31-design/Astrocyte/internal/adapters/worksscan"
	"github.com/songconmaisaix31-design/Astrocyte/internal/apierrors"
	attention "github.com/songconmaisaix31-design/Astrocyte/internal/attention/app"
	"github.com/songconmaisaix31-design/Astrocyte/internal/workspace/app"
	"github.com/songconmaisaix31-design/Astrocyte/internal/workspace/domain"
	"github.com/songconmaisaix31-design/Astrocyte/migrations"
)

func humanCaller() domain.Caller { return domain.Caller{Kind: "human", ID: "u1"} }

func worksApplication(t *testing.T) (app.Works, *sqlite.DB, *sqlite.EventRepository) {
	t.Helper()
	db, err := sqlite.Open(filepath.Join(t.TempDir(), "state.sqlite"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := db.Close(); err != nil {
			t.Error(err)
		}
	})
	if err = db.RunMigrations(context.Background(), migrations.FS); err != nil {
		t.Fatal(err)
	}
	receipts := sqlite.NewEventRepository(db)
	sink := events.NewWorksEventSink(attention.NewPendingEventService(receipts), nil)
	return app.NewWorksService(db, worksscan.NewScanner(), worksscan.Delivery{}, sink, app.WorksServiceOptions{}), db, receipts
}

// 平台标记是调用者实际写入的格式输入，文件数量与 mtime 均从磁盘观察。
func workDirectory(t *testing.T, platform, name string) string {
	t.Helper()
	root := t.TempDir()
	if platform != "" {
		data, err := json.Marshal(map[string]string{"name": name, "platform": platform})
		if err != nil {
			t.Fatal(err)
		}
		if err = os.WriteFile(filepath.Join(root, ".astrocyte-works.json"), data, 0600); err != nil {
			t.Fatal(err)
		}
	}
	if err := os.WriteFile(filepath.Join(root, "manual.txt"), []byte("本次实际创建的测试文件"), 0600); err != nil {
		t.Fatal(err)
	}
	return root
}

func refreshWorks(t *testing.T, service app.Works, roots ...string) domain.WorksSnapshot {
	t.Helper()
	snapshot, err := service.RefreshWorks(context.Background(), humanCaller(), app.WorksRefreshCommand{Roots: roots})
	if err != nil {
		t.Fatal(err)
	}
	return snapshot
}

func workReceipts(t *testing.T, repository *sqlite.EventRepository) []domain.WorkEvent {
	t.Helper()
	rows, err := repository.ListEvents(context.Background(), 0, math.MaxInt64)
	if err != nil {
		t.Fatal(err)
	}
	result := make([]domain.WorkEvent, len(rows))
	for i, row := range rows {
		if row.Event != nil || row.ID == "" {
			t.Fatal("作品活动产生了额外认知事实或缺少实际回执")
		}
		if err := json.Unmarshal(row.RawBytes, &result[i]); err != nil {
			t.Fatal(err)
		}
	}
	return result
}

func deliveryInput(t *testing.T, entries ...domain.WorkDeliveryEntry) app.WorksDeliveryCommand {
	t.Helper()
	// JSON 为明确的投递输入，生产 parser 核实文件前必须保持未核实。
	data, err := json.Marshal(domain.WorkDelivery{Works: entries})
	if err != nil {
		t.Fatal(err)
	}
	return app.WorksDeliveryCommand{Content: data}
}

func TestWorksListIsEmptyWhenNothingScanned(t *testing.T) {
	service, _, _ := worksApplication(t)
	snapshot, err := service.ListWorks(context.Background(), humanCaller())
	if err != nil {
		t.Fatal(err)
	}
	if len(snapshot.Works) != 0 || len(snapshot.Platforms) != 0 || snapshot.Status != "unknown" || snapshot.ObservedAt != nil {
		t.Fatalf("fresh album invented observations: %+v", snapshot)
	}
}

func TestWorksRequiresHumanCaller(t *testing.T) {
	service, _, _ := worksApplication(t)
	if _, err := service.ListWorks(context.Background(), domain.Caller{Kind: "agent", ID: "a1"}); err == nil {
		t.Fatal("agent caller must be rejected")
	}
}

func TestRefreshWorksScansAndStats(t *testing.T) {
	service, _, receipts := worksApplication(t)
	roots := []string{workDirectory(t, "codex", "app"), workDirectory(t, "codex", "deck"), workDirectory(t, "lingxi", "report")}
	snapshot := refreshWorks(t, service, roots...)
	if len(snapshot.Works) != 3 || snapshot.Status != "complete" || len(snapshot.Platforms) != 2 {
		t.Fatalf("actual scan counts: %+v", snapshot)
	}
	for _, work := range snapshot.Works {
		if !work.Files.Verified || work.Files.FileCount != 2 || work.Files.Mtime == nil {
			t.Fatalf("actual files lost: %+v", work.Files)
		}
	}
	if rows := workReceipts(t, receipts); len(rows) != 3 {
		t.Fatalf("expected 3 persisted synced events, got %d", len(rows))
	}
}

func TestRefreshPreservesHumanMetadata(t *testing.T) {
	service, _, _ := worksApplication(t)
	root := workDirectory(t, "codex", "app")
	first := refreshWorks(t, service, root)
	id := first.Works[0].ID
	if _, err := service.SetWorkAttitude(context.Background(), humanCaller(), id, app.WorkAttitudeCommand{Attitude: domain.AttitudePark}); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "new.txt"), []byte("实际新文件"), 0600); err != nil {
		t.Fatal(err)
	}
	snapshot := refreshWorks(t, service, root)
	if len(snapshot.Works) != 1 || snapshot.Works[0].Attitude != domain.AttitudePark || snapshot.Works[0].Files.FileCount != 3 {
		t.Fatalf("metadata or new observation lost: %+v", snapshot.Works)
	}
}

func TestImportDeliverySeedsGeneratedMetadata(t *testing.T) {
	service, _, receipts := worksApplication(t)
	root := workDirectory(t, "", "notes")
	snapshot, err := service.ImportWorksDelivery(context.Background(), humanCaller(), deliveryInput(t, domain.WorkDeliveryEntry{Name: "notes", Platform: "workbuddy", Path: root, Kind: domain.WorkKindFolder, Attitude: domain.AttitudeContinue, Handoff: "generated guide"}))
	if err != nil {
		t.Fatal(err)
	}
	if len(snapshot.Works) != 1 {
		t.Fatalf("expected 1 work, got %d", len(snapshot.Works))
	}
	work := snapshot.Works[0]
	if work.Attitude != domain.AttitudeContinue || work.Handoff.Source != domain.HandoffSourceGenerated || work.Handoff.Content != "generated guide" || work.Files.Verified {
		t.Fatalf("delivery metadata/uncertainty lost: %+v", work)
	}
	if rows := workReceipts(t, receipts); len(rows) != 1 || rows[0].Kind != domain.WorkEventImported {
		t.Fatalf("import event lost: %+v", rows)
	}
}

func TestSetWorkHandoffUpdatesTimestamp(t *testing.T) {
	service, _, _ := worksApplication(t)
	first := refreshWorks(t, service, workDirectory(t, "codex", "app"))
	work, err := service.SetWorkHandoff(context.Background(), humanCaller(), first.Works[0].ID, app.WorkHandoffCommand{Content: "manual notes"})
	if err != nil {
		t.Fatal(err)
	}
	if work.Handoff.Source != domain.HandoffSourceHuman || work.Handoff.UpdatedAt == nil || work.Handoff.Content != "manual notes" {
		t.Fatalf("manual handoff lost: %+v", work.Handoff)
	}
}

func TestSetWorkMetadataRevisionConflict(t *testing.T) {
	service, _, _ := worksApplication(t)
	first := refreshWorks(t, service, workDirectory(t, "codex", "app"))
	id := first.Works[0].ID
	if _, err := service.SetWorkAttitude(context.Background(), humanCaller(), id, app.WorkAttitudeCommand{Attitude: domain.AttitudeDone}); err != nil {
		t.Fatal(err)
	}
	_, err := service.SetWorkAttitude(context.Background(), humanCaller(), id, app.WorkAttitudeCommand{ExpectedRevision: 0, Attitude: domain.AttitudePark})
	var failure *apierrors.ServiceError
	if !errors.As(err, &failure) || failure.Code != apierrors.VersionConflict {
		t.Fatalf("stale metadata must conflict: %v", err)
	}
}

func TestWorksEventsDoNotInventHumanAttention(t *testing.T) {
	service, _, receipts := worksApplication(t)
	first := refreshWorks(t, service, workDirectory(t, "codex", "app"))
	rows := workReceipts(t, receipts)
	if len(rows) != 1 || rows[0].Source != domain.WorkEventSourceMachine || rows[0].Topic != "" {
		t.Fatalf("machine refresh invented human topic: %+v", rows)
	}
	if _, err := service.ListWorks(context.Background(), humanCaller()); err != nil {
		t.Fatal(err)
	}
	if len(workReceipts(t, receipts)) != 1 {
		t.Fatal("GET emitted new event")
	}
	root := workDirectory(t, "", "other")
	imported, err := service.ImportWorksDelivery(context.Background(), humanCaller(), deliveryInput(t, domain.WorkDeliveryEntry{Name: "other", Platform: "claude", Path: root}))
	if err != nil {
		t.Fatal(err)
	}
	rows = workReceipts(t, receipts)
	if len(rows) != 2 || rows[1].WorkID != domain.WorkIdentity("claude", root) || rows[1].Kind != domain.WorkEventImported || len(imported.Works) != 2 {
		t.Fatalf("import emitted unrelated cached work: %+v", rows)
	}
	work, err := service.SetWorkGroup(context.Background(), humanCaller(), first.Works[0].ID, app.WorkGroupCommand{Group: "research"})
	if err != nil {
		t.Fatal(err)
	}
	rows = workReceipts(t, receipts)
	if len(rows) != 3 || rows[2].Kind != domain.WorkEventGroupSet {
		t.Fatalf("group became approval: %+v", rows)
	}
	if _, err = service.SetWorkHandoff(context.Background(), humanCaller(), work.ID, app.WorkHandoffCommand{ExpectedRevision: work.Revision, Content: "human guide"}); err != nil {
		t.Fatal(err)
	}
	for _, row := range workReceipts(t, receipts) {
		if row.Topic != "" {
			t.Fatalf("platform became topic: %+v", row)
		}
	}
	signals, err := receipts.ListSignals(context.Background(), 0, math.MaxInt64)
	if err != nil || len(signals) != 0 {
		t.Fatalf("work metadata created cognitive signals: %+v %v", signals, err)
	}
}

func TestWorksScanVerifiesExistingDeliveryWithoutInventingPlatform(t *testing.T) {
	service, _, _ := worksApplication(t)
	root := workDirectory(t, "", "delivered")
	snapshot, err := service.ImportWorksDelivery(context.Background(), humanCaller(), deliveryInput(t, domain.WorkDeliveryEntry{Name: "human title", Platform: "codex", Path: root}))
	if err != nil {
		t.Fatal(err)
	}
	delivered := snapshot.Works[0]
	work, err := service.SetWorkAttitude(context.Background(), humanCaller(), delivered.ID, app.WorkAttitudeCommand{Attitude: domain.AttitudePark})
	if err != nil {
		t.Fatal(err)
	}
	work, err = service.SetWorkHandoff(context.Background(), humanCaller(), delivered.ID, app.WorkHandoffCommand{ExpectedRevision: work.Revision, Content: "manual guide"})
	if err != nil {
		t.Fatal(err)
	}
	snapshot = refreshWorks(t, service, root)
	if len(snapshot.Works) != 1 {
		t.Fatalf("scan invented another platform work: %+v", snapshot.Works)
	}
	got := snapshot.Works[0]
	if got.ID != delivered.ID || got.Platform != "codex" || got.Source != domain.WorkSourceJSON || got.Name != delivered.Name || !got.Files.Verified || got.Attitude != domain.AttitudePark || got.Handoff.Content != "manual guide" || got.Revision != work.Revision {
		t.Fatalf("provenance/metadata lost: %+v", got)
	}
}

func TestWorksFailedRootNoLongerClaimsCurrentVerification(t *testing.T) {
	service, _, _ := worksApplication(t)
	root := workDirectory(t, "codex", "app")
	refreshWorks(t, service, root)
	// 实际移动本次自有根，使原授权路径失效；保留全部文件作为证据。
	if err := os.Rename(root, root+"-unavailable"); err != nil {
		t.Fatal(err)
	}
	snapshot := refreshWorks(t, service, root)
	if snapshot.Status != "partial" || len(snapshot.Works) != 1 || snapshot.Works[0].Files.Verified || snapshot.Works[0].Files.Reason != "root_unavailable" {
		t.Fatalf("missing root claims current verification: %+v", snapshot)
	}
}

func TestWorksMalformedScanReturnsFailureAndPreservesStaleCache(t *testing.T) {
	service, _, _ := worksApplication(t)
	root := workDirectory(t, "codex", "app")
	refreshWorks(t, service, root)
	if err := os.WriteFile(filepath.Join(root, ".astrocyte-works.json"), []byte("{"), 0600); err != nil {
		t.Fatal(err)
	}
	if _, err := service.RefreshWorks(context.Background(), humanCaller(), app.WorksRefreshCommand{Roots: []string{root}}); err == nil {
		t.Fatal("provided malformed marker became successful refresh")
	}
	cached, err := service.ListWorks(context.Background(), humanCaller())
	if err != nil || cached.Status != "stale" || len(cached.Works) != 1 || len(cached.Failures) != 1 {
		t.Fatalf("stale original cache lost: %+v %v", cached, err)
	}
}
