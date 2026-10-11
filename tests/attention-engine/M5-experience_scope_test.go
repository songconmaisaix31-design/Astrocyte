package attention_engine_test

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/songconmaisaix31-design/Astrocyte/internal/apierrors"
	"github.com/songconmaisaix31-design/Astrocyte/internal/attention/app"
)

func TestM5ExperienceTrustedProjectAuditedSharingAndDemotionRestart(t *testing.T) {
	ctx := context.Background()
	path := filepath.Join(t.TempDir(), "experience.sqlite")
	db := experienceDB(t, path)
	defer func() { db.Close() }()
	experienceReceipt(t, db, "failed", "A", "agent_task", "agent_task", "failed", "mapped")
	a, b, c := experienceService(db, "A"), experienceService(db, "B"), experienceService(db, "C")
	candidate, err := a.CaptureFailure(ctx, "failed")
	if err != nil || candidate == nil {
		t.Fatal(err)
	}
	id := candidate.Experience.ID
	for _, service := range []*app.ExperienceService{a, b} {
		project := "A"
		if service == b {
			project = "B"
		}
		rows, err := service.GetExperiences(ctx, experienceAgent, project, "")
		if err != nil || len(rows) != 0 {
			t.Fatalf("quarantine retrieved: %+v %v", rows, err)
		}
	}
	_, err = b.GetExperiences(ctx, experienceAgent, "A", "")
	experienceError(t, err, apierrors.ScopeDenied)
	_, err = a.GetExperience(ctx, experienceAgent, id, "A")
	experienceError(t, err, apierrors.ScopeDenied)
	_, err = b.GetExperience(ctx, experienceHuman, id, "B")
	experienceError(t, err, apierrors.ScopeDenied)
	if _, err = a.ApproveExperience(ctx, experienceHuman, id, app.ApproveExperienceCommand{CommandMeta: experienceMeta("approve", 1), OwnerProject: "A"}); err != nil {
		t.Fatal(err)
	}
	rows, err := b.GetExperiences(ctx, experienceAgent, "B", "agent")
	if err != nil || len(rows) != 0 {
		t.Fatalf("default isolation: %+v %v", rows, err)
	}
	if _, err = a.ShareExperience(ctx, experienceHuman, id, app.ShareMemoryCommand{CommandMeta: experienceMeta("share", 2), OwnerProject: "A", TargetProject: "B"}); err != nil {
		t.Fatal(err)
	}
	if err = db.Close(); err != nil {
		t.Fatal(err)
	}
	db = experienceDB(t, path)
	a, b, c = experienceService(db, "A"), experienceService(db, "B"), experienceService(db, "C")
	rows, err = b.GetExperiences(ctx, experienceAgent, "B", "agent_task")
	if err != nil || len(rows) != 1 || rows[0].Experience.ID != id || len(rows[0].Shares) != 1 {
		t.Fatalf("persisted shared retrieval: %+v %v", rows, err)
	}
	share := rows[0].Shares[0]
	if share.OwnerProject != "A" || share.TargetProject != "B" || share.ActorID != experienceHuman.ID || share.ActorKind != "human" || share.TS != 1800000001000 || share.RequestID != "share" {
		t.Fatalf("who/when/where: %+v", share)
	}
	rows, err = c.GetExperiences(ctx, experienceAgent, "C", "")
	if err != nil || len(rows) != 0 {
		t.Fatalf("share widened to C: %+v %v", rows, err)
	}
	rows, err = b.GetExperiences(ctx, experienceAgent, "B", "secret")
	if err != nil || len(rows) != 0 {
		t.Fatalf("private raw text became a pattern: %+v %v", rows, err)
	}
	_, err = b.ShareExperience(ctx, experienceHuman, id, app.ShareMemoryCommand{CommandMeta: experienceMeta("re-share", 3), OwnerProject: "B", TargetProject: "C"})
	experienceError(t, err, apierrors.ScopeDenied)
	_, err = b.DemoteExperience(ctx, experienceHuman, id, app.DemoteMemoryCommand{CommandMeta: experienceMeta("foreign-demote", 3), OwnerProject: "A"})
	experienceError(t, err, apierrors.ScopeDenied)
	if _, err = a.DemoteExperience(ctx, experienceHuman, id, app.DemoteMemoryCommand{CommandMeta: experienceMeta("demote", 3), OwnerProject: "A"}); err != nil {
		t.Fatal(err)
	}
	if err = db.Close(); err != nil {
		t.Fatal(err)
	}
	db = experienceDB(t, path)
	a, b = experienceService(db, "A"), experienceService(db, "B")
	rows, err = b.GetExperiences(ctx, experienceAgent, "B", "")
	if err != nil || len(rows) != 0 {
		t.Fatalf("demotion failed to hide shared memory: %+v %v", rows, err)
	}
	_, err = b.GetExperience(ctx, experienceHuman, id, "B")
	experienceError(t, err, apierrors.ScopeDenied)
	record, err := a.GetExperience(ctx, experienceHuman, id, "A")
	if err != nil || record.Experience.Status != "quarantine" || record.Version != 4 || len(record.Actions) != 3 || len(record.Shares) != 1 {
		t.Fatalf("persisted demotion audit: %+v %v", record, err)
	}
}
