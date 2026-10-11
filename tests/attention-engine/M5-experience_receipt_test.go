package attention_engine_test

import (
	"context"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/songconmaisaix31-design/Astrocyte/internal/apierrors"
	"github.com/songconmaisaix31-design/Astrocyte/internal/attention/app"
)

func TestM5ExperienceGovernanceReceiptReplaysImmutableResultAfterRestart(t *testing.T) {
	ctx := context.Background()
	path := filepath.Join(t.TempDir(), "experience.sqlite")
	db := experienceDB(t, path)
	defer func() { db.Close() }()
	experienceReceipt(t, db, "failed", "A", "agent_task", "agent_task", "failed", "mapped")
	svc := experienceService(db, "A")
	candidate, err := svc.CaptureFailure(ctx, "failed")
	if err != nil || candidate == nil {
		t.Fatal(err)
	}
	id := candidate.Experience.ID
	approve := app.ApproveExperienceCommand{CommandMeta: experienceMeta("approve", 1), OwnerProject: "A"}
	share := app.ShareMemoryCommand{CommandMeta: experienceMeta("share", 2), OwnerProject: "A", TargetProject: "B"}
	demote := app.DemoteMemoryCommand{CommandMeta: experienceMeta("demote", 3), OwnerProject: "A"}
	approved, err := svc.ApproveExperience(ctx, experienceHuman, id, approve)
	if err != nil {
		t.Fatal(err)
	}
	shared, err := svc.ShareExperience(ctx, experienceHuman, id, share)
	if err != nil {
		t.Fatal(err)
	}
	demoted, err := svc.DemoteExperience(ctx, experienceHuman, id, demote)
	if err != nil {
		t.Fatal(err)
	}
	if err = db.Close(); err != nil {
		t.Fatal(err)
	}
	db = experienceDB(t, path)
	svc = experienceService(db, "A")
	// New tracing IDs do not change the semantic retry bound to the same key.
	approve.RequestID, share.RequestID, demote.RequestID = "retry-approve", "retry-share", "retry-demote"
	got, err := svc.ApproveExperience(ctx, experienceHuman, id, approve)
	if err != nil || !reflect.DeepEqual(got, approved) || got.Experience.Status != "approved" || got.Version != 2 {
		t.Fatalf("historical approve result: %+v %v", got, err)
	}
	got, err = svc.ShareExperience(ctx, experienceHuman, id, share)
	if err != nil || !reflect.DeepEqual(got, shared) {
		t.Fatalf("historical share result: %+v %v", got, err)
	}
	got, err = svc.DemoteExperience(ctx, experienceHuman, id, demote)
	if err != nil || !reflect.DeepEqual(got, demoted) {
		t.Fatalf("historical demote result: %+v %v", got, err)
	}
	changed := share
	changed.TargetProject = "C"
	_, err = svc.ShareExperience(ctx, experienceHuman, id, changed)
	experienceError(t, err, apierrors.VersionConflict)
	changedApprove := approve
	changedApprove.ExpectedVersion = 4
	_, err = svc.ApproveExperience(ctx, experienceHuman, id, changedApprove)
	experienceError(t, err, apierrors.VersionConflict)
	_, err = svc.ApproveExperience(ctx, app.Principal{ID: "other-human", Kind: "human"}, id, approve)
	experienceError(t, err, apierrors.VersionConflict)
	_, err = experienceService(db, "B").ApproveExperience(ctx, experienceHuman, id, approve)
	experienceError(t, err, apierrors.ScopeDenied)
	record, err := svc.GetExperience(ctx, experienceHuman, id, "A")
	if err != nil || record.Version != 4 || record.Experience.Status != "quarantine" || len(record.Actions) != 3 || len(record.Shares) != 1 {
		t.Fatalf("retry changed live state/audit: %+v %v", record, err)
	}
	var receipts int
	if err = db.Conn().QueryRowContext(ctx, `SELECT count(*) FROM attention_receipts WHERE command LIKE 'Experience/%'`).Scan(&receipts); err != nil || receipts != 3 {
		t.Fatalf("receipt count: %d %v", receipts, err)
	}
}

func TestM5ExperienceUnknownReceiptDoesNotRetryGovernance(t *testing.T) {
	ctx := context.Background()
	db := experienceDB(t, filepath.Join(t.TempDir(), "experience.sqlite"))
	defer db.Close()
	experienceReceipt(t, db, "failed", "A", "agent_task", "agent_task", "failed", "mapped")
	svc := experienceService(db, "A")
	candidate, err := svc.CaptureFailure(ctx, "failed")
	if err != nil || candidate == nil {
		t.Fatal(err)
	}
	cmd := app.ApproveExperienceCommand{CommandMeta: experienceMeta("approve", 1), OwnerProject: "A"}
	if _, err = svc.ApproveExperience(ctx, experienceHuman, candidate.Experience.ID, cmd); err != nil {
		t.Fatal(err)
	}
	// Inject a recovery receipt with an unknown result, keeping the original
	// input digest. It must stop at reconciliation rather than re-run the gate.
	if _, err = db.Conn().ExecContext(ctx, `UPDATE attention_receipts SET result='{"status":"unknown"}' WHERE command='Experience/approve/A'`); err != nil {
		t.Fatal(err)
	}
	_, err = svc.ApproveExperience(ctx, experienceHuman, candidate.Experience.ID, cmd)
	experienceError(t, err, apierrors.DeliveryUnknown)
	record, err := svc.GetExperience(ctx, experienceHuman, candidate.Experience.ID, "A")
	if err != nil || record.Version != 2 || len(record.Actions) != 1 {
		t.Fatalf("unknown receipt retried: %+v %v", record, err)
	}
}
