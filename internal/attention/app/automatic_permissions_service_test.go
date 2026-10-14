package app_test

import (
	"context"
	"testing"

	"github.com/songconmaisaix31-design/Astrocyte/internal/adapters/importers"
	"github.com/songconmaisaix31-design/Astrocyte/internal/adapters/sqlite"
	"github.com/songconmaisaix31-design/Astrocyte/internal/apierrors"
	. "github.com/songconmaisaix31-design/Astrocyte/internal/attention/app"
)

func automaticPublicSource(t *testing.T) (*Service, *sqlite.DB, *observedObjects, *observedSource, MaterialDetail) {
	t.Helper()
	ctx := context.Background()
	_, db, objects := fixture(t)
	reader := &observedSource{Reader: importers.NewReader(nil)}
	repository := sqlite.NewAttentionRepository(db)
	s := NewAttentionService(repository, reader, objects, ServiceOptions{})
	queued, err := s.ImportMaterial(ctx, human, ImportMaterialCommand{CommandMeta: meta("permissions-official-source", 1), Kind: "paper", Adapter: "arxiv", SourceLocator: "https://arxiv.org/abs/2501.12948v1"})
	if err != nil {
		t.Fatal("queue authorized public fixed paper")
	}
	if _, err = s.ProcessNextJob(ctx); err != nil {
		t.Fatal("read actual public fixed paper")
	}
	imported, err := s.GetJob(ctx, human, queued.JobID)
	if err != nil || imported.Status != "succeeded" || imported.MaterialID == nil {
		t.Fatal("actual official source prerequisite failed")
	}
	material, err := s.GetMaterial(ctx, human, *imported.MaterialID)
	if err != nil || len(material.Revisions) != 1 || material.Revisions[0].Provenance.Processor != "arxiv" || material.Revisions[0].SourceKey == "" {
		t.Fatal("actual source identity or provenance missing")
	}
	return s, db, objects, reader, material
}

func TestAutomaticPermissionsAndUnconfiguredProcessorUseActualSource(t *testing.T) {
	ctx := context.Background()
	s, db, objects, reader, material := automaticPublicSource(t)
	repository := sqlite.NewAttentionRepository(db)
	command := RequestDistillationCommand{CommandMeta: meta("deny-agent", 1), InputRefs: sourceRefs(material), Stage: "content", ProcessingConfig: "selected-public-fixed-source", Question: "仅依据选定公开论文整理已知信息和未知问题。"}
	beforeJobs := countRows(t, db, "attention_jobs")
	beforeReceipts := countRows(t, db, "attention_receipts")
	beforePublications := objects.publishes.Load()
	_, err := s.RequestDistillation(ctx, agent, command)
	errorCode(t, err, apierrors.ScopeDenied)
	command.CommandMeta = meta("deny-scope", 1)
	_, err = s.RequestDistillation(ctx, human, command)
	errorCode(t, err, apierrors.ScopeDenied)
	if countRows(t, db, "attention_jobs") != beforeJobs || countRows(t, db, "attention_receipts") != beforeReceipts || objects.publishes.Load() != beforePublications || reader.calls.Load() != 1 {
		t.Fatal("denied automatic request produced work or repeated source I/O")
	}
	// 授权采用生产 Reader 实际生成的 key，未配置处理器保持明确失败。
	s = NewAttentionService(repository, reader, objects, ServiceOptions{AllowedProcessingSourceKeys: []string{material.Revisions[0].SourceKey}})
	command.CommandMeta = meta("notconfigured", 1)
	failed, err := s.RequestDistillation(ctx, human, command)
	if err != nil {
		t.Fatal("persist actual unavailable processor state")
	}
	job, err := s.GetJob(ctx, human, failed.JobID)
	if err != nil || job.Status != "failed" || job.Error == nil || job.Error.Code != apierrors.ProviderUnavailable || job.ExternalStarted || job.DeliveryUnknown || job.DistillationID != nil || countRows(t, db, "attention_distillations") != 0 {
		t.Fatal("unconfigured processor claimed a summary or external operation")
	}
	if _, err = s.ProcessNextJob(ctx); err != nil {
		t.Fatal("inspect queue after unavailable processor")
	}
	if reader.calls.Load() != 1 || objects.publishes.Load() != beforePublications || countRows(t, db, "attention_jobs") != beforeJobs+1 {
		t.Fatal("unconfigured processor started source work or another operation")
	}
}

func TestAutomaticPriorScopeCannotExpandSelectedSources(t *testing.T) {
	ctx := context.Background()
	s, db, objects, reader, selected := automaticPublicSource(t)
	unselected := importFixture(t, s, "explicit-manual-context", "这是人工提交的附加上下文参数，未选入自动处理请求。")
	prior, err := s.RecordDistillation(ctx, human, RecordDistillationCommand{CommandMeta: meta("explicit-human-prior", 1), Stage: "content", ProcessingConfig: "manual", OutputText: "人工记录的比较包含附加上下文。", InputRefs: sourceRefs(selected), RelatedRefs: sourceRefs(unselected)})
	if err != nil || prior.Distillation.Status != "succeeded" || prior.Distillation.Provenance.Mode != "manual" {
		t.Fatal("persist actual explicit human prior record")
	}
	beforeJobs := countRows(t, db, "attention_jobs")
	beforeReceipts := countRows(t, db, "attention_receipts")
	beforePublications := objects.publishes.Load()
	beforeReads := reader.calls.Load()
	s = NewAttentionService(sqlite.NewAttentionRepository(db), reader, objects, ServiceOptions{AllowedProcessingSourceKeys: []string{selected.Revisions[0].SourceKey, unselected.Revisions[0].SourceKey}})
	command := RequestDistillationCommand{CommandMeta: meta("prior-unselected-context", 1), Stage: "content", ProcessingConfig: "selected-public-fixed-source", InputRefs: sourceRefs(selected), PriorDistillationIDs: []string{prior.Distillation.ID}}
	_, err = s.RequestDistillation(ctx, human, command)
	errorCode(t, err, apierrors.ScopeDenied)
	// 实际 manual 来源即使其真实 key 在许可列表中，也不能取得公开来源处理权限。
	command.CommandMeta = meta("allowkey-manual-source", 1)
	command.InputRefs, command.PriorDistillationIDs = sourceRefs(unselected), nil
	_, err = s.RequestDistillation(ctx, human, command)
	errorCode(t, err, apierrors.ScopeDenied)
	if countRows(t, db, "attention_jobs") != beforeJobs || countRows(t, db, "attention_receipts") != beforeReceipts || countRows(t, db, "attention_distillations") != 1 || objects.publishes.Load() != beforePublications || reader.calls.Load() != beforeReads {
		t.Fatal("unselected context or manual provenance started work or changed records")
	}
}
