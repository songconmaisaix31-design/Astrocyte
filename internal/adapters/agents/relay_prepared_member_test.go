package agents

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/songconmaisaix31-design/Astrocyte/internal/apierrors"
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
