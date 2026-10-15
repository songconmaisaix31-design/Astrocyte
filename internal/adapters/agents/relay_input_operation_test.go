package agents

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/songconmaisaix31-design/Astrocyte/internal/apierrors"
)

func TestRelayInputOperationRejectsColdRegistryAndInvalidPurpose(t *testing.T) {
	registry := NewRegistry("unopened-input-operation-root")
	for _, purpose := range []string{"claim_work_item", "submit_artifact"} {
		result, err := registry.ReadOwnedInputOperation(context.Background(), uuid.NewString(), uuid.NewString(), purpose)
		var known *apierrors.ServiceError
		if !errors.As(err, &known) || known.Code != apierrors.DeliveryUnknown || result.InputAccepted || result.ToolsKnown || result.TargetsKnown || result.CallsKnown || result.TokensKnown {
			t.Fatal("cold Registry produced execution facts")
		}
		encoded, err := json.Marshal(result)
		if err != nil || string(encoded) != "{}" {
			t.Fatal("private input evidence was exposed through JSON")
		}
	}
	_, err := registry.ReadOwnedInputOperation(context.Background(), uuid.NewString(), uuid.NewString(), "unknown-purpose")
	var known *apierrors.ServiceError
	if !errors.As(err, &known) || known.Code != apierrors.ValidationFailed {
		t.Fatal("unknown membership purpose was accepted")
	}
	_, err = registry.StopOwnedInputOperation(context.Background(), uuid.NewString(), uuid.NewString())
	if !errors.As(err, &known) || known.Code != apierrors.DeliveryUnknown {
		t.Fatal("cold Registry claimed an original process stop")
	}
}
