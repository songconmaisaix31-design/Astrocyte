package agents

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/songconmaisaix31-design/Astrocyte/internal/apierrors"
	"github.com/songconmaisaix31-design/Astrocyte/internal/workspace/domain"
)

func TestRelayPortableMemberSourceRejectsColdRegistry(t *testing.T) {
	var source domain.RelayMemberEvidenceSource = NewRegistry("unopened-portable-member-root")
	for _, purpose := range []string{"claim_work_item", "submit_artifact"} {
		fact, err := source.ReadRelayMemberEvidence(context.Background(), uuid.NewString(), uuid.NewString(), purpose)
		var known *apierrors.ServiceError
		if !errors.As(err, &known) || known.Code != apierrors.DeliveryUnknown || fact.ControllerHeld || fact.InputAccepted || fact.ToolsKnown || fact.TargetsKnown || fact.CallsKnown || fact.TokensKnown {
			t.Fatal("portable source reconstructed ownership or a known execution descriptor")
		}
		body, err := json.Marshal(fact)
		if err != nil || string(body) != "{}" || fact.ReportedTotalTokens != nil || fact.AcceptedTurnID != nil || fact.ExitObservedAt != nil || fact.Outcome != nil || fact.FailureCode != nil || fact.ResultObservedAt != nil || fact.ResultSource != "" {
			t.Fatal("portable source exposed private or invented observation data")
		}
	}
	_, err := source.StopRelayMemberEvidence(context.Background(), uuid.NewString(), uuid.NewString())
	var known *apierrors.ServiceError
	if !errors.As(err, &known) || known.Code != apierrors.DeliveryUnknown {
		t.Fatal("portable source confirmed a stop without the original controller")
	}
}
