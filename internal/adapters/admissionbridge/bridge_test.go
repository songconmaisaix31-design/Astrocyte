package admissionbridge_test

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/songconmaisaix31-design/Astrocyte/internal/adapters/admissionbridge"
	"github.com/songconmaisaix31-design/Astrocyte/internal/adapters/importers"
	"github.com/songconmaisaix31-design/Astrocyte/internal/adapters/objects"
	"github.com/songconmaisaix31-design/Astrocyte/internal/adapters/sqlite"
	"github.com/songconmaisaix31-design/Astrocyte/internal/apierrors"
	attention "github.com/songconmaisaix31-design/Astrocyte/internal/attention/app"
	workspace "github.com/songconmaisaix31-design/Astrocyte/internal/workspace/app"
	"github.com/songconmaisaix31-design/Astrocyte/migrations"
)

func metadata() attention.CommandMeta {
	return attention.CommandMeta{SchemaVersion: 1, RequestID: uuid.NewString(), IdempotencyKey: uuid.NewString(), ExpectedVersion: 1}
}

func ownedDatabase(t *testing.T) (*sqlite.DB, string) {
	t.Helper()
	root, err := filepath.Abs("../../../.verification")
	if err != nil {
		t.Fatal(err)
	}
	dir, err := os.MkdirTemp(root, "admission-actual-")
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("owned admission evidence directory=%s", dir)
	db, err := sqlite.Open(filepath.Join(dir, "state.sqlite"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := db.Close(); err != nil {
			t.Error(err)
		}
	})
	if err := db.RunMigrations(context.Background(), migrations.FS); err != nil {
		t.Fatal(err)
	}
	return db, dir
}

func rows(t *testing.T, db *sqlite.DB, table string) int {
	t.Helper()
	var count int
	if err := db.Conn().QueryRow("SELECT COUNT(*) FROM " + table).Scan(&count); err != nil {
		t.Fatal(err)
	}
	return count
}

func requireCode(t *testing.T, err error, code apierrors.Code) {
	t.Helper()
	var service *apierrors.ServiceError
	if !errors.As(err, &service) || service.Code != code {
		t.Fatalf("expected %s: %v", code, err)
	}
}

func TestActualAdmissionDraftAtomicLifecycle(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	defer cancel()
	db, dir := ownedDatabase(t)
	store, err := objects.New(filepath.Join(dir, "objects"))
	if err != nil {
		t.Fatal(err)
	}
	core := attention.NewAttentionService(sqlite.NewAttentionRepository(db), importers.NewReader(nil), store, attention.ServiceOptions{})
	human := attention.Principal{Kind: "human", ID: "admission-test-human"}
	queued, err := core.ImportMaterial(ctx, human, attention.ImportMaterialCommand{CommandMeta: metadata(), Kind: "paper", Adapter: "arxiv", SourceLocator: "https://arxiv.org/abs/2501.12948v1"})
	if err != nil {
		t.Fatal(err)
	}
	if _, err = core.ProcessNextJob(ctx); err != nil {
		t.Fatal("read actual public fixed paper", err)
	}
	job, err := core.GetJob(ctx, human, queued.JobID)
	if err != nil || job.Status != "succeeded" || job.MaterialID == nil {
		if job.Error != nil {
			t.Logf("actual import status=%s code=%s required_action=%s", job.Status, job.Error.Code, job.Error.RequiredAction)
		}
		t.Fatal("actual public paper import did not succeed", err)
	}
	material, err := core.GetMaterial(ctx, human, *job.MaterialID)
	if err != nil || len(material.Revisions) != 1 || material.Revisions[0].SourceKey == "" {
		t.Fatal("actual fixed paper identity missing", err)
	}
	refs := []attention.SourceRef{{MaterialID: material.Material.ID, Revision: material.Material.CurrentRevision, Locator: material.Material.SourceLocator}}
	// 这是对实际公开来源的人工作业记录，不声明调用模型或完成判断。
	topic, err := core.RecordDistillation(ctx, human, attention.RecordDistillationCommand{CommandMeta: metadata(), InputRefs: refs, Stage: "topic", ProcessingConfig: "manual", OutputText: "人工审阅：所选论文的结论仍需检查原文与适用条件。", PendingQuestions: []string{"哪些证据支持论文结论，哪些条件仍待核实？"}})
	if err != nil {
		t.Fatal(err)
	}
	ready, err := core.CreateOpportunity(ctx, human, attention.OpportunityCommand{CommandMeta: metadata(), Title: material.Material.Title, EvidenceRefs: refs, DistillationIDs: []string{topic.Distillation.ID}, Purpose: "审阅实际固定论文的证据。", NextStep: "人工检查原文与适用条件。"})
	if err != nil || ready.Opportunity.State != "ready_for_review" {
		t.Fatal("actual human record did not produce a review-ready candidate", err)
	}
	other, _ := ownedDatabase(t)
	wrongWorkspace := workspace.NewS2Service(sqlite.NewWorkspaceRepository(other), nil, nil)
	admissions, err := admissionbridge.Connect(core, sqlite.NewAdmissionRepository(db), wrongWorkspace, attention.AdmissionOptions{})
	if err != nil {
		t.Fatal(err)
	}
	command := attention.AdmissionCommand{CommandMeta: metadata(), ExpectedRevision: 1}
	beforeEvents, beforeReceipts := rows(t, db, "attention_outbox"), rows(t, db, "attention_receipts")
	_, err = admissions.AdmitOpportunity(ctx, attention.Principal{Kind: "agent", ID: "ordinary-agent"}, ready.Opportunity.ID, command)
	requireCode(t, err, apierrors.ScopeDenied)
	_, err = admissions.AdmitOpportunityWithJudgment(ctx, "unconfigured-workflow", "unrecorded-decision", ready.Opportunity.ID, command)
	requireCode(t, err, apierrors.ApprovalRequired)
	// 实际消费者指向另一实际数据库，事务身份检查在生产连接层拒绝它。
	_, err = admissions.AdmitOpportunity(ctx, human, ready.Opportunity.ID, command)
	requireCode(t, err, apierrors.InternalError)
	for table, want := range map[string]int{"attention_admissions": 0, "attention_outbox": beforeEvents, "attention_receipts": beforeReceipts, "workspace_proposals": 0} {
		if rows(t, db, table) != want {
			t.Fatalf("failed actual consumer left committed rows in %s", table)
		}
	}
	if rows(t, other, "workspace_proposals") != 0 {
		t.Fatal("another database received a draft")
	}
	unchanged, err := core.GetOpportunity(ctx, human, ready.Opportunity.ID)
	if err != nil || unchanged.Opportunity.State != "ready_for_review" || unchanged.Opportunity.Version != 1 {
		t.Fatal("failed consumer changed actual candidate", err)
	}
	proposals := workspace.NewS2Service(sqlite.NewWorkspaceRepository(db), nil, nil)
	admissions, err = admissionbridge.Connect(core, sqlite.NewAdmissionRepository(db), proposals, attention.AdmissionOptions{})
	if err != nil {
		t.Fatal(err)
	}
	waits := db.Conn().Stats().WaitCount
	accepted, err := admissions.AdmitOpportunity(ctx, human, ready.Opportunity.ID, command)
	if err != nil || accepted.Draft == nil || accepted.Draft.ID == "" || accepted.Admission.Version != 1 {
		t.Fatal("actual admission and draft did not commit", err)
	}
	if db.Conn().Stats().WaitCount != waits || rows(t, db, "attention_admissions") != 1 || rows(t, db, "workspace_proposals") != 1 || rows(t, db, "attention_outbox") != beforeEvents+1 || rows(t, db, "attention_receipts") != beforeReceipts+1 {
		t.Fatal("admission borrowed another connection or duplicated its atomic state")
	}
	var proposalVersion int
	var projectID, goalID, assignment string
	if err := db.Conn().QueryRow("SELECT version,json_extract(data,'$.content.scope.project_id'),json_extract(data,'$.content.goal.id'),json_extract(data,'$.assignment_status') FROM workspace_proposals WHERE id=?", accepted.Draft.ID).Scan(&proposalVersion, &projectID, &goalID, &assignment); err != nil || proposalVersion != 1 || projectID != "" || goalID != "" || assignment != "unassigned" {
		t.Fatal("admission invented a project, goal or execution assignment", err)
	}
	command.RequestID = uuid.NewString()
	replayed, err := admissions.AdmitOpportunity(ctx, human, ready.Opportunity.ID, command)
	if err != nil || !reflect.DeepEqual(replayed, accepted) {
		t.Fatal("same semantic request did not replay its actual receipt", err)
	}
	command.IdempotencyKey = uuid.NewString()
	_, err = admissions.AdmitOpportunity(ctx, human, ready.Opportunity.ID, command)
	requireCode(t, err, apierrors.VersionConflict)
	snapshot, err := admissions.ReadAdmittedOpportunitySnapshot(ctx, accepted.Admission.AdmittedEventID)
	if err != nil || !snapshot.Active || !reflect.DeepEqual(snapshot.Inputs, refs) || snapshot.OpportunityRevision != 1 || snapshot.OwnerNamespace != "human:"+human.ID {
		t.Fatal("actual admission source snapshot changed", err)
	}
	_, err = admissions.RevokeAdmission(ctx, attention.Principal{Kind: "human", ID: "other-human"}, accepted.Admission.ID, attention.RevokeAdmissionCommand{CommandMeta: metadata(), Reason: "人工撤销"})
	requireCode(t, err, apierrors.ScopeDenied)
	revoked, err := admissions.RevokeAdmission(ctx, human, accepted.Admission.ID, attention.RevokeAdmissionCommand{CommandMeta: metadata(), Reason: "人工撤销本次资料准入"})
	if err != nil || revoked.Admission.Status != "revoked" || revoked.Admission.Version != 2 || revoked.Admission.AdmittedEventID != accepted.Admission.AdmittedEventID {
		t.Fatal("actual independent admission revocation failed", err)
	}
	snapshot, err = admissions.ReadAdmittedOpportunitySnapshot(ctx, accepted.Admission.AdmittedEventID)
	if err != nil || snapshot.Active || !reflect.DeepEqual(snapshot.Inputs, refs) {
		t.Fatal("revocation changed fixed sources or left admission active", err)
	}
	if _, err := proposals.ConsumeOpportunityAdmission(ctx, accepted.Admission.AdmittedEventID); err == nil {
		t.Fatal("revoked admission authorized another draft consumption")
	}
	if rows(t, db, "workspace_approvals") != 0 || rows(t, db, "local_agent_sessions") != 0 || rows(t, db, "workspace_proposals") != 1 {
		t.Fatal("admission or its revocation created execution approval or native work")
	}
	if err := db.Close(); err != nil {
		t.Fatal("close actual admission database", err)
	}
	db, err = sqlite.Open(filepath.Join(dir, "state.sqlite"))
	if err != nil {
		t.Fatal("reopen actual admission database", err)
	}
	t.Cleanup(func() {
		if err := db.Close(); err != nil {
			t.Error(err)
		}
	})
	core = attention.NewAttentionService(sqlite.NewAttentionRepository(db), importers.NewReader(nil), store, attention.ServiceOptions{})
	proposals = workspace.NewS2Service(sqlite.NewWorkspaceRepository(db), nil, nil)
	admissions, err = admissionbridge.Connect(core, sqlite.NewAdmissionRepository(db), proposals, attention.AdmissionOptions{})
	if err != nil {
		t.Fatal(err)
	}
	loaded, err := admissions.GetAdmission(ctx, human, accepted.Admission.ID)
	if err != nil || !reflect.DeepEqual(loaded.Admission, revoked.Admission) || rows(t, db, "workspace_proposals") != 1 {
		t.Fatal("actual restart lost fixed admission, revocation or draft", err)
	}
	snapshot, err = admissions.ReadAdmittedOpportunitySnapshot(ctx, accepted.Admission.AdmittedEventID)
	if err != nil || snapshot.Active || !reflect.DeepEqual(snapshot.Inputs, refs) {
		t.Fatal("actual restart changed immutable admission evidence", err)
	}
	if _, err := core.UpdateMaterial(ctx, human, material.Material.ID, attention.UpdateMaterialCommand{CommandMeta: attention.CommandMeta{SchemaVersion: 1, RequestID: uuid.NewString(), IdempotencyKey: uuid.NewString(), ExpectedVersion: material.Material.Version}, Lifecycle: "withdrawn"}); err != nil {
		t.Fatal("withdraw actual source through production lifecycle", err)
	}
	_, err = admissions.ReadAdmittedOpportunitySnapshot(ctx, accepted.Admission.AdmittedEventID)
	requireCode(t, err, apierrors.ScopeDenied)
}
