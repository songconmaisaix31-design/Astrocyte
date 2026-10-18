package agents

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/songconmaisaix31-design/Astrocyte/internal/apierrors"
	swarm "github.com/songconmaisaix31-design/Astrocyte/internal/swarm/app"
	"github.com/songconmaisaix31-design/Astrocyte/internal/workspace/domain"
)

func TestRelayPreparedMemberRejectsMissingOriginalSources(t *testing.T) {
	registry := NewRegistry("unopened-prepared-member-root")
	if source, err := NewRelayPreparedMemberEvidenceSource(registry, nil); source != nil || err == nil {
		t.Fatal("missing actual prepared input reader produced a configured source")
	}
	source := &RelayPreparedMemberEvidenceSource{registry: registry}
	if input, err := source.ReadPreparedNativeSession(context.Background(), uuid.NewString(), uuid.NewString()); err == nil || input.Session.ID != "" || input.ControllerID != "" || !input.PreparedAt.IsZero() {
		t.Fatal("missing private preparation reader produced input evidence")
	}
	if input, err := source.ReadPreparedMemberInput(context.Background(), uuid.NewString(), uuid.NewString()); err == nil || input != (swarm.PreparedMemberInput{}) {
		t.Fatal("缺少原准备来源时返回了 Swarm 输入身份")
	}
	if _, err := source.ReadRelayMemberEvidence(context.Background(), uuid.NewString(), uuid.NewString(), "claim_work_item"); err == nil {
		t.Fatal("cold original Registry produced member ownership")
	}
	if _, err := source.StopRelayMemberEvidence(context.Background(), uuid.NewString(), uuid.NewString()); err == nil {
		t.Fatal("cold original Registry produced a confirmed process stop")
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := source.ReadPreparedNativeSession(ctx, uuid.NewString(), uuid.NewString()); !errors.Is(err, context.Canceled) {
		t.Fatal("canceled preparation read continued")
	}
	if input, err := source.ReadPreparedMemberInput(ctx, uuid.NewString(), uuid.NewString()); !errors.Is(err, context.Canceled) || input != (swarm.PreparedMemberInput{}) {
		t.Fatal("取消后的 Swarm 准备读取继续执行或返回了输入身份")
	}
}

func TestRelayMemberExecutionRejectsUnownedMetadata(t *testing.T) {
	native := &Native{id: "codex"}
	for _, session := range []domain.NativeSession{{}, {ID: uuid.NewString(), CLI: "codex", NativeID: uuid.NewString(), ProjectID: uuid.NewString(), Version: "unrecorded-version", Ownership: "owned"}} {
		result, err := native.ownedRelayMemberExecution(context.Background(), session)
		var known *apierrors.ServiceError
		if !errors.As(err, &known) || result.ProcessID != 0 || !result.StartedAt.IsZero() || result.ExitConfirmed || result.EffectiveToolSet != "" {
			t.Fatal("unrecorded controller produced member execution facts")
		}
	}
}
