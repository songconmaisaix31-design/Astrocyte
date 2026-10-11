package app

import (
	"context"
	"testing"
	"time"

	"github.com/songconmaisaix31-design/Astrocyte/internal/apierrors"
	"github.com/songconmaisaix31-design/Astrocyte/internal/workspace/domain"
)

type fakeWorksRepo struct {
	snapshot domain.WorksSnapshot
	metadata map[string]domain.WorkHumanMetadata
}

func newFakeWorksRepo() *fakeWorksRepo {
	return &fakeWorksRepo{snapshot: domain.WorksSnapshot{Works: []domain.Work{}}, metadata: map[string]domain.WorkHumanMetadata{}}
}

func (f *fakeWorksRepo) LoadWorksSnapshot(context.Context) (domain.WorksSnapshot, error) {
	return f.snapshot, nil
}

func (f *fakeWorksRepo) SaveWorksSnapshot(_ context.Context, s domain.WorksSnapshot) error {
	f.snapshot = s
	return nil
}

func (f *fakeWorksRepo) LoadWorkMetadata(_ context.Context, id string) (domain.WorkHumanMetadata, error) {
	return f.metadata[id], nil
}

func (f *fakeWorksRepo) SaveWorkMetadata(_ context.Context, id string, m domain.WorkHumanMetadata, expected int) error {
	if m.Revision != expected+1 {
		return &apierrors.ServiceError{Code: apierrors.VersionConflict, Message: "revision mismatch"}
	}
	f.metadata[id] = m
	return nil
}

type fakeScanSource struct {
	works    []domain.Work
	failures []domain.WorkDiscoveryFailure
	err      error
}

func (f *fakeScanSource) ScanWorks(context.Context, []string) ([]domain.Work, []domain.WorkDiscoveryFailure, error) {
	return f.works, f.failures, f.err
}

type fakeDeliverySource struct {
	works []domain.Work
	err   error
}

func (f *fakeDeliverySource) ParseDelivery(context.Context, []byte) ([]domain.Work, error) {
	return f.works, f.err
}

type fakeEventSink struct {
	events []domain.WorkEvent
}

func (f *fakeEventSink) EmitWorkEvent(_ context.Context, e domain.WorkEvent) error {
	f.events = append(f.events, e)
	return nil
}

func fixedClock() func() time.Time {
	now := time.Date(2026, 10, 11, 12, 0, 0, 0, time.UTC)
	return func() time.Time { return now }
}

func humanCaller() domain.Caller { return domain.Caller{Kind: "human", ID: "u1"} }

func scanWork(platform, name, root string) domain.Work {
	mtime := time.Now().Add(-time.Hour)
	return domain.Work{
		ID: domain.WorkIdentity(platform, root), Platform: platform, Name: name, Root: root,
		Kind: domain.WorkKindFolder, Source: domain.WorkSourceScan,
		Files: domain.WorkFileFact{SourcePath: root, Mtime: &mtime, FileCount: 3, Verified: true},
	}
}

func TestWorksListIsEmptyWhenNothingScanned(t *testing.T) {
	s := NewWorksService(newFakeWorksRepo(), &fakeScanSource{}, &fakeDeliverySource{}, nil, WorksServiceOptions{Clock: fixedClock()})
	snapshot, err := s.ListWorks(context.Background(), humanCaller())
	if err != nil {
		t.Fatal(err)
	}
	if len(snapshot.Works) != 0 || len(snapshot.Platforms) != 0 {
		t.Fatalf("fresh album must be empty, got %d works", len(snapshot.Works))
	}
}

func TestWorksRequiresHumanCaller(t *testing.T) {
	s := NewWorksService(newFakeWorksRepo(), &fakeScanSource{}, &fakeDeliverySource{}, nil, WorksServiceOptions{Clock: fixedClock()})
	if _, err := s.ListWorks(context.Background(), domain.Caller{Kind: "agent", ID: "a1"}); err == nil {
		t.Fatal("agent caller must be rejected")
	}
}

func TestRefreshWorksScansAndStats(t *testing.T) {
	scan := &fakeScanSource{works: []domain.Work{
		scanWork("codex", "app", "C:\\w\\app"),
		scanWork("codex", "deck", "C:\\w\\deck"),
		scanWork("lingxi", "report", "C:\\w\\report"),
	}}
	sink := &fakeEventSink{}
	s := NewWorksService(newFakeWorksRepo(), scan, &fakeDeliverySource{}, sink, WorksServiceOptions{Clock: fixedClock()})
	snapshot, err := s.RefreshWorks(context.Background(), humanCaller(), WorksRefreshCommand{})
	if err != nil {
		t.Fatal(err)
	}
	if len(snapshot.Works) != 3 {
		t.Fatalf("expected 3 works, got %d", len(snapshot.Works))
	}
	if snapshot.Status != "complete" {
		t.Fatalf("expected complete, got %q", snapshot.Status)
	}
	if len(snapshot.Platforms) != 2 {
		t.Fatalf("expected 2 platforms, got %d", len(snapshot.Platforms))
	}
	if len(sink.events) != 3 {
		t.Fatalf("expected 3 synced events, got %d", len(sink.events))
	}
}

func TestRefreshPreservesHumanMetadata(t *testing.T) {
	repo := newFakeWorksRepo()
	scan := &fakeScanSource{works: []domain.Work{scanWork("codex", "app", "C:\\w\\app")}}
	s := NewWorksService(repo, scan, &fakeDeliverySource{}, nil, WorksServiceOptions{Clock: fixedClock()})
	if _, err := s.RefreshWorks(context.Background(), humanCaller(), WorksRefreshCommand{}); err != nil {
		t.Fatal(err)
	}
	id := domain.WorkIdentity("codex", "C:\\w\\app")
	if _, err := s.SetWorkAttitude(context.Background(), humanCaller(), id, WorkAttitudeCommand{Attitude: domain.AttitudePark}); err != nil {
		t.Fatal(err)
	}
	// Re-scan with a fresh observation (human fields zeroed) must not lose the park.
	scan.works = []domain.Work{scanWork("codex", "app", "C:\\w\\app")}
	snapshot, err := s.RefreshWorks(context.Background(), humanCaller(), WorksRefreshCommand{})
	if err != nil {
		t.Fatal(err)
	}
	if len(snapshot.Works) != 1 || snapshot.Works[0].Attitude != domain.AttitudePark {
		t.Fatalf("attitude lost on refresh: %+v", snapshot.Works[0])
	}
}

func TestImportDeliverySeedsGeneratedMetadata(t *testing.T) {
	delivery := &fakeDeliverySource{works: []domain.Work{{
		ID: domain.WorkIdentity("workbuddy", "C:\\w\\notes"), Platform: "workbuddy", Name: "notes",
		Root: "C:\\w\\notes", Kind: domain.WorkKindFolder, Source: domain.WorkSourceJSON,
		Attitude: domain.AttitudeContinue, Handoff: domain.WorkHandoff{Content: "generated guide"},
	}}}
	sink := &fakeEventSink{}
	s := NewWorksService(newFakeWorksRepo(), &fakeScanSource{}, delivery, sink, WorksServiceOptions{Clock: fixedClock()})
	snapshot, err := s.ImportWorksDelivery(context.Background(), humanCaller(), WorksDeliveryCommand{Content: []byte(`{"works":[]}`)})
	if err != nil {
		t.Fatal(err)
	}
	if len(snapshot.Works) != 1 {
		t.Fatalf("expected 1 work, got %d", len(snapshot.Works))
	}
	work := snapshot.Works[0]
	if work.Attitude != domain.AttitudeContinue {
		t.Fatalf("delivered attitude not seeded: %q", work.Attitude)
	}
	if work.Handoff.Source != domain.HandoffSourceGenerated || work.Handoff.Content != "generated guide" {
		t.Fatalf("delivered handoff not marked generated: %+v", work.Handoff)
	}
	if work.Files.Verified {
		t.Fatal("JSON-delivered work must not be marked file-verified")
	}
	if len(sink.events) != 1 || sink.events[0].Kind != domain.WorkEventImported {
		t.Fatalf("expected one imported event, got %+v", sink.events)
	}
}

func TestSetWorkHandoffUpdatesTimestamp(t *testing.T) {
	repo := newFakeWorksRepo()
	scan := &fakeScanSource{works: []domain.Work{scanWork("codex", "app", "C:\\w\\app")}}
	s := NewWorksService(repo, scan, &fakeDeliverySource{}, nil, WorksServiceOptions{Clock: fixedClock()})
	if _, err := s.RefreshWorks(context.Background(), humanCaller(), WorksRefreshCommand{}); err != nil {
		t.Fatal(err)
	}
	id := domain.WorkIdentity("codex", "C:\\w\\app")
	work, err := s.SetWorkHandoff(context.Background(), humanCaller(), id, WorkHandoffCommand{Content: "manual notes"})
	if err != nil {
		t.Fatal(err)
	}
	if work.Handoff.Source != domain.HandoffSourceHuman || work.Handoff.UpdatedAt == nil {
		t.Fatalf("handoff not marked human with timestamp: %+v", work.Handoff)
	}
}

func TestSetWorkMetadataRevisionConflict(t *testing.T) {
	repo := newFakeWorksRepo()
	scan := &fakeScanSource{works: []domain.Work{scanWork("codex", "app", "C:\\w\\app")}}
	s := NewWorksService(repo, scan, &fakeDeliverySource{}, nil, WorksServiceOptions{Clock: fixedClock()})
	if _, err := s.RefreshWorks(context.Background(), humanCaller(), WorksRefreshCommand{}); err != nil {
		t.Fatal(err)
	}
	id := domain.WorkIdentity("codex", "C:\\w\\app")
	if _, err := s.SetWorkAttitude(context.Background(), humanCaller(), id, WorkAttitudeCommand{Attitude: domain.AttitudeDone}); err != nil {
		t.Fatal(err)
	}
	// Stale revision must be rejected, not overwritten.
	if _, err := s.SetWorkAttitude(context.Background(), humanCaller(), id, WorkAttitudeCommand{ExpectedRevision: 0, Attitude: domain.AttitudePark}); err == nil {
		t.Fatal("stale revision must conflict")
	}
}

func TestWorksEventsDoNotInventHumanAttention(t *testing.T) {
	repo := newFakeWorksRepo()
	scan := &fakeScanSource{works: []domain.Work{scanWork("codex", "app", "C:\\w\\app")}}
	sink := &fakeEventSink{}
	delivery := &fakeDeliverySource{works: []domain.Work{scanWork("claude", "other", "C:\\w\\other")}}
	s := NewWorksService(repo, scan, delivery, sink, WorksServiceOptions{Clock: fixedClock()})
	ctx := context.Background()
	if _, err := s.RefreshWorks(ctx, humanCaller(), WorksRefreshCommand{}); err != nil {
		t.Fatal(err)
	}
	if len(sink.events) != 1 || sink.events[0].Source != domain.WorkEventSourceMachine || sink.events[0].Topic != "" {
		t.Fatalf("refresh must be machine observation with unknown topic: %+v", sink.events)
	}
	sink.events = nil
	if _, err := s.ListWorks(ctx, humanCaller()); err != nil {
		t.Fatal(err)
	}
	if len(sink.events) != 0 {
		t.Fatal("GET must not emit events")
	}
	if _, err := s.ImportWorksDelivery(ctx, humanCaller(), WorksDeliveryCommand{Content: []byte(`{"works":[]}`)}); err != nil {
		t.Fatal(err)
	}
	if len(sink.events) != 1 || sink.events[0].WorkID != delivery.works[0].ID {
		t.Fatalf("import must not emit unrelated cached works: %+v", sink.events)
	}
	sink.events = nil
	work, err := s.SetWorkGroup(ctx, humanCaller(), scan.works[0].ID, WorkGroupCommand{Group: "research"})
	if err != nil {
		t.Fatal(err)
	}
	if len(sink.events) != 1 || sink.events[0].Kind != domain.WorkEventGroupSet {
		t.Fatalf("group must not become attitude approval: %+v", sink.events)
	}
	if _, err := s.SetWorkHandoff(ctx, humanCaller(), work.ID, WorkHandoffCommand{ExpectedRevision: work.Revision, Content: "human guide"}); err != nil {
		t.Fatal(err)
	}
	for _, event := range sink.events {
		if event.Topic != "" {
			t.Fatalf("platform is not a cognitive topic: %+v", event)
		}
	}
}

func TestWorksScanVerifiesExistingDeliveryWithoutInventingPlatform(t *testing.T) {
	repo := newFakeWorksRepo()
	root := "C:\\w\\delivered"
	delivered := domain.Work{ID: domain.WorkIdentity("codex", root), Platform: "codex", Root: root, Name: "human title", Kind: domain.WorkKindFolder, Source: domain.WorkSourceJSON}
	repo.snapshot = domain.WorksSnapshot{Works: []domain.Work{delivered}}
	repo.metadata[delivered.ID] = domain.WorkHumanMetadata{Attitude: domain.AttitudePark, Handoff: domain.WorkHandoff{Content: "manual guide", Source: domain.HandoffSourceHuman}, Revision: 1}
	scan := &fakeScanSource{works: []domain.Work{scanWork("unknown", "folder name", root)}}
	service := NewWorksService(repo, scan, nil, nil, WorksServiceOptions{})
	snapshot, err := service.RefreshWorks(context.Background(), humanCaller(), WorksRefreshCommand{Roots: []string{root}})
	if err != nil {
		t.Fatal(err)
	}
	if len(snapshot.Works) != 1 {
		t.Fatalf("verification invented another platform work: %+v", snapshot.Works)
	}
	got := snapshot.Works[0]
	if got.ID != delivered.ID || got.Platform != "codex" || got.Source != domain.WorkSourceJSON || got.Name != delivered.Name || !got.Files.Verified || got.Attitude != domain.AttitudePark || got.Handoff.Content != "manual guide" || got.Revision != 1 {
		t.Fatalf("scan lost provenance/metadata: %+v", got)
	}
}

func TestWorksFailedRootNoLongerClaimsCurrentVerification(t *testing.T) {
	repo := newFakeWorksRepo()
	root := "C:\\w\\unavailable"
	repo.snapshot = domain.WorksSnapshot{Works: []domain.Work{scanWork("codex", "app", root)}}
	scan := &fakeScanSource{failures: []domain.WorkDiscoveryFailure{{Root: root, Reason: "root_unavailable"}}}
	service := NewWorksService(repo, scan, nil, nil, WorksServiceOptions{})
	snapshot, err := service.RefreshWorks(context.Background(), humanCaller(), WorksRefreshCommand{Roots: []string{root}})
	if err != nil {
		t.Fatal(err)
	}
	if snapshot.Status != "partial" || len(snapshot.Works) != 1 || snapshot.Works[0].Files.Verified || snapshot.Works[0].Files.Reason != "root_unavailable" {
		t.Fatalf("failed observation masquerades as verified: %+v", snapshot)
	}
}
