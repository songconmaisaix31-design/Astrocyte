package app_test

import (
	"context"
	"testing"

	"github.com/songconmaisaix31-design/Astrocyte/internal/adapters/sqlite"
	. "github.com/songconmaisaix31-design/Astrocyte/internal/attention/app"
)

func realReadyOpportunity(t *testing.T) (*Service, *sqlite.DB, *observedObjects, MaterialDetail, OpportunityDetail) {
	t.Helper()
	ctx := context.Background()
	s, db, objects, _, material := automaticPublicSource(t)
	// 显式人工记录作为待审候选前置，不提供模型输出或已完成判断。
	topic, err := s.RecordDistillation(ctx, human, RecordDistillationCommand{CommandMeta: meta("admission-human-question", 1), InputRefs: sourceRefs(material), Stage: "topic", ProcessingConfig: "manual", OutputText: "人工记录：后续验证仍需核查原文和附加证据。", PendingQuestions: []string{"选定论文的证据覆盖哪些范围，仍有哪些未核实内容？"}})
	if err != nil || topic.Distillation.Provenance.Mode != "manual" || topic.Distillation.Status != "succeeded" {
		t.Fatal("persist explicit human question against actual fixed source")
	}
	beforePublications := objects.publishes.Load()
	beforeEvents := countRows(t, db, "attention_outbox")
	ready, err := s.CreateOpportunity(ctx, human, OpportunityCommand{CommandMeta: meta("admission-ready-candidate", 1), Title: material.Material.Title, EvidenceRefs: sourceRefs(material), DistillationIDs: []string{topic.Distillation.ID}, Purpose: "审阅选定论文的固定来源和公开证据，保留未核实问题。", NextStep: "人工检查论文版本、引用关系和待补证据。"})
	if err != nil || ready.Opportunity.State != "ready_for_review" || ready.Opportunity.Revision != 1 || ready.Opportunity.Version != 1 || len(ready.Revisions) != 1 {
		t.Fatal("actual fixed source did not create a review-ready candidate")
	}
	if len(ready.Opportunity.EvidenceRefs) != 1 || ready.Opportunity.EvidenceRefs[0] != sourceRefs(material)[0] || len(ready.Opportunity.GoalRefs) != 0 || ready.Opportunity.Dimensions.GoalProgress.Value != nil || ready.Opportunity.CompositeScore != nil {
		t.Fatal("ready candidate replaced fixed evidence or invented a goal or score")
	}
	if objects.publishes.Load() != beforePublications || countRows(t, db, "attention_outbox") != beforeEvents+1 {
		t.Fatal("creating a ready candidate produced unrelated work")
	}
	var created, admitted int
	if err = db.Conn().QueryRow("SELECT COUNT(*) FROM attention_outbox WHERE type='attention.opportunity_created' AND aggregate_id=?", ready.Opportunity.ID).Scan(&created); err != nil || created != 1 {
		t.Fatal("actual candidate creation event missing or duplicated")
	}
	if err = db.Conn().QueryRow("SELECT COUNT(*) FROM attention_outbox WHERE type='attention.opportunity_admitted'").Scan(&admitted); err != nil || admitted != 0 || countRows(t, db, "local_agent_sessions") != 0 {
		t.Fatal("ready candidate started admission or a native session")
	}
	return s, db, objects, material, ready
}

func TestActualFixedPaperReadyCandidateDoesNotAdmitOrExecute(t *testing.T) {
	s, db, _, material, ready := realReadyOpportunity(t)
	loaded, err := s.GetOpportunity(context.Background(), human, ready.Opportunity.ID)
	if err != nil || loaded.Opportunity.Revision != ready.Opportunity.Revision || loaded.Opportunity.State != "ready_for_review" || loaded.Revisions[0].EvidenceRefs[0] != sourceRefs(material)[0] || countRows(t, db, "attention_opportunities") != 1 {
		t.Fatal("actual review-ready candidate or immutable evidence was not persisted")
	}
}
