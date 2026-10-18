package sqlite

import (
	"context"
	"encoding/json"
	"errors"
	"path/filepath"
	"testing"

	"github.com/google/uuid"
	"github.com/songconmaisaix31-design/Astrocyte/internal/adapters/agents"
	"github.com/songconmaisaix31-design/Astrocyte/internal/apierrors"
	workspace "github.com/songconmaisaix31-design/Astrocyte/internal/workspace/domain"
)

func TestDynamicDAGReconciliationConfirmationEncoding(t *testing.T) {
	for _, tc := range []struct {
		name  string
		raw   json.RawMessage
		valid bool
	}{
		{"legacy_v1_absent", nil, true},
		{"explicit_confirmed", json.RawMessage(`true`), true},
		{"explicit_false", json.RawMessage(`false`), false},
		{"explicit_null", json.RawMessage(`null`), false},
		{"wrong_type", json.RawMessage(`"true"`), false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			confirmed, err := reconciliationExitConfirmation(tc.raw)
			if (err == nil) != tc.valid || confirmed != tc.valid {
				t.Fatalf("confirmed=%v error=%v", confirmed, err)
			}
		})
	}
}

func TestDynamicDAGActualMissingRegistryControllerClassification(t *testing.T) {
	registry := agents.NewRegistry(filepath.Join(swarmTestDirectory(t), "native"))
	_, err := registry.ReadOwnedNativeSession(context.Background(), uuid.NewString())
	var unavailable *workspace.RelayObservationUnavailable
	if !errors.As(err, &unavailable) || unavailable.Reason != "controller_not_held" {
		t.Fatalf("actual Registry absence lacks its explicit classification: %v", err)
	}
	var original *apierrors.ServiceError
	if !errors.As(err, &original) || original.Code != apierrors.DeliveryUnknown {
		t.Fatalf("actual Registry absence changed its external error: %v", err)
	}
}
