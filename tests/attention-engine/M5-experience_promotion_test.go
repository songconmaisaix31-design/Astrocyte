package attention_engine_test

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/songconmaisaix31-design/Astrocyte/internal/apierrors"
	"github.com/songconmaisaix31-design/Astrocyte/internal/attention/app"
)

func TestM5ExperienceHumanGatesVersionAndAtomicAudit(t *testing.T) {
	ctx := context.Background()
	db := experienceDB(t, filepath.Join(t.TempDir(), "experience.sqlite"))
	defer db.Close()
	experienceReceipt(t, db, "failed", "A", "agent_task", "agent_task", "failed", "mapped")
	svc := experienceService(db, "A")
	candidate, err := svc.CaptureFailure(ctx, "failed")
	if err != nil || candidate == nil {
		t.Fatal(err)
	}
	id := candidate.Experience.ID
	_, err = svc.ApproveExperience(ctx, experienceAgent, id, app.ApproveExperienceCommand{CommandMeta: experienceMeta("agent-approve", 1), OwnerProject: "A"})
	experienceError(t, err, apierrors.ApprovalRequired)
	_, err = svc.DemoteExperience(ctx, experienceAgent, id, app.DemoteMemoryCommand{CommandMeta: experienceMeta("agent-demote", 1), OwnerProject: "A"})
	experienceError(t, err, apierrors.ApprovalRequired)
	_, err = svc.ShareExperience(ctx, experienceAgent, id, app.ShareMemoryCommand{CommandMeta: experienceMeta("agent-share", 1), OwnerProject: "A", TargetProject: "B"})
	experienceError(t, err, apierrors.ApprovalRequired)
	_, err = svc.DemoteExperience(ctx, experienceHuman, id, app.DemoteMemoryCommand{CommandMeta: experienceMeta("illegal-demote", 1), OwnerProject: "A"})
	experienceError(t, err, apierrors.ValidationFailed)
	_, err = svc.ShareExperience(ctx, experienceHuman, id, app.ShareMemoryCommand{CommandMeta: experienceMeta("quarantine-share", 1), OwnerProject: "A", TargetProject: "B"})
	experienceError(t, err, apierrors.ApprovalRequired)
	result, err := svc.ApproveExperience(ctx, experienceHuman, id, app.ApproveExperienceCommand{CommandMeta: experienceMeta("approve-1", 1), OwnerProject: "A"})
	if err != nil || result.Version != 2 || result.Experience.Status != "approved" {
		t.Fatalf("human approve: %+v %v", result, err)
	}
	_, err = svc.DemoteExperience(ctx, experienceHuman, id, app.DemoteMemoryCommand{CommandMeta: experienceMeta("stale", 1), OwnerProject: "A"})
	experienceError(t, err, apierrors.VersionConflict)
	_, err = svc.ApproveExperience(ctx, experienceHuman, id, app.ApproveExperienceCommand{CommandMeta: experienceMeta("illegal-approve", 2), OwnerProject: "A"})
	experienceError(t, err, apierrors.ValidationFailed)
	// Audit request collision occurs after share insertion and head update;
	// rollback must restore all three, not leave cross-project visibility behind.
	_, err = svc.ShareExperience(ctx, experienceHuman, id, app.ShareMemoryCommand{CommandMeta: experienceMeta("approve-1", 2), OwnerProject: "A", TargetProject: "B"})
	experienceError(t, err, apierrors.VersionConflict)
	record, err := svc.GetExperience(ctx, experienceHuman, id, "A")
	if err != nil || record.Version != 2 || len(record.Shares) != 0 || len(record.Actions) != 1 || record.Actions[0].ActorID != experienceHuman.ID || record.Actions[0].Action != "approve" {
		t.Fatalf("audit rollback: %+v %v", record, err)
	}
}
