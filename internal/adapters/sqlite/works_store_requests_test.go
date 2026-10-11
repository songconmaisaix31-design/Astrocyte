package sqlite

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/songconmaisaix31-design/Astrocyte/internal/adapters/worksscan"
	"github.com/songconmaisaix31-design/Astrocyte/internal/apierrors"
	workspaceapp "github.com/songconmaisaix31-design/Astrocyte/internal/workspace/app"
	"github.com/songconmaisaix31-design/Astrocyte/internal/workspace/domain"
)

type worksReceiptScan struct{ calls int }

func (s *worksReceiptScan) ScanWorks(ctx context.Context, roots []string) ([]domain.Work, []domain.WorkDiscoveryFailure, error) {
	s.calls++
	return worksscan.NewScanner().ScanWorks(ctx, roots)
}

type worksReceiptSink struct {
	calls int
	fail  bool
}

func (s *worksReceiptSink) EmitWorkEvent(context.Context, domain.WorkEvent) error {
	s.calls++
	if s.fail {
		return errors.New("test sink unavailable after observing the event")
	}
	return nil
}

func worksCode(t *testing.T, err error, code apierrors.Code) {
	t.Helper()
	var typed *apierrors.ServiceError
	if !errors.As(err, &typed) || typed.Code != code {
		t.Fatalf("want %s, got %v", code, err)
	}
}

func TestWorksRequestReceiptsPersistAndPreventUnknownReplay(t *testing.T) {
	ctx := context.Background()
	path := filepath.Join(t.TempDir(), "state.sqlite")
	db := openAttentionDB(t, path)
	ensureWorksTables(t, db)
	// openAttentionDB applies the canonical W0 migration016. Do not redefine it.
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "main.go"), []byte("metadata-only"), 0o644); err != nil {
		t.Fatal(err)
	}
	scan := &worksReceiptScan{}
	sink := &worksReceiptSink{}
	service := workspaceapp.NewWorksService(db, scan, worksscan.Delivery{}, sink, workspaceapp.WorksServiceOptions{})
	caller := domain.Caller{Kind: "human", ID: "human", OperationID: "scan-once"}
	command := workspaceapp.WorksRefreshCommand{Roots: []string{root}}
	first, err := service.RefreshWorks(ctx, caller, command)
	if err != nil {
		t.Fatal(err)
	}
	replayed, err := service.RefreshWorks(ctx, caller, command)
	if err != nil || len(replayed.Works) != 1 || scan.calls != 1 || sink.calls != 1 {
		t.Fatalf("same request reexecuted: scan=%d sink=%d result=%+v err=%v", scan.calls, sink.calls, replayed, err)
	}
	if !first.ObservedAt.Equal(*replayed.ObservedAt) {
		t.Fatal("receipt replaced the original observation")
	}
	_, err = service.RefreshWorks(ctx, caller, workspaceapp.WorksRefreshCommand{Roots: []string{t.TempDir()}})
	worksCode(t, err, apierrors.VersionConflict)
	caller.OperationID = "attitude-once"
	edit := workspaceapp.WorkAttitudeCommand{Attitude: domain.AttitudeContinue}
	work, err := service.SetWorkAttitude(ctx, caller, first.Works[0].ID, edit)
	if err != nil {
		t.Fatal(err)
	}
	repeatWork, err := service.SetWorkAttitude(ctx, caller, work.ID, edit)
	if err != nil || repeatWork.Revision != work.Revision || sink.calls != 2 {
		t.Fatalf("metadata receipt did not precede stale CAS: %+v %v", repeatWork, err)
	}
	caller.OperationID = "unknown-handoff"
	sink.fail = true
	guide := workspaceapp.WorkHandoffCommand{ExpectedRevision: work.Revision, Content: "human handoff"}
	_, err = service.SetWorkHandoff(ctx, caller, work.ID, guide)
	worksCode(t, err, apierrors.DeliveryUnknown)
	_, err = service.SetWorkHandoff(ctx, caller, work.ID, guide)
	worksCode(t, err, apierrors.DeliveryUnknown)
	if sink.calls != 3 {
		t.Fatal("unknown event delivery was retried")
	}
	metadata, err := db.LoadWorkMetadata(ctx, work.ID)
	if err != nil || metadata.Revision != 2 || metadata.Handoff.Content != guide.Content {
		t.Fatalf("unknown result lost the actual persisted effect: %+v %v", metadata, err)
	}
	input, _ := json.Marshal(command)
	if _, fresh, err := db.BeginWorksRequest(ctx, "human:human", "RefreshWorks", "crashed-pending", input); err != nil || !fresh {
		t.Fatalf("pending reservation: %v", err)
	}
	if err := db.Close(); err != nil {
		t.Fatal(err)
	}
	db2, err := Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer db2.Close()
	scan2 := &worksReceiptScan{}
	sink2 := &worksReceiptSink{}
	restarted := workspaceapp.NewWorksService(db2, scan2, worksscan.Delivery{}, sink2, workspaceapp.WorksServiceOptions{})
	caller.OperationID = "scan-once"
	if _, err := restarted.RefreshWorks(ctx, caller, command); err != nil {
		t.Fatal(err)
	}
	caller.OperationID = "unknown-handoff"
	_, err = restarted.SetWorkHandoff(ctx, caller, work.ID, guide)
	worksCode(t, err, apierrors.DeliveryUnknown)
	caller.OperationID = "crashed-pending"
	_, err = restarted.RefreshWorks(ctx, caller, command)
	worksCode(t, err, apierrors.DeliveryUnknown)
	if scan2.calls != 0 || sink2.calls != 0 {
		t.Fatal("restart replayed an accepted, pending or unknown command")
	}
}
