package app

import (
	"testing"

	"github.com/songconmaisaix31-design/Astrocyte/internal/apierrors"
	"github.com/songconmaisaix31-design/Astrocyte/internal/swarm/domain"
)

func TestDynamicDAGStopFailureClassification(t *testing.T) {
	for _, tc := range []struct {
		name     string
		err      error
		retained bool
	}{
		{"unclassified_native_identity_or_delivery", &apierrors.ServiceError{Code: apierrors.DeliveryUnknown}, false},
		{"unmarked_unknown_execution", domain.Failure(apierrors.DeliveryUnknown, "original execution remains unknown"), false},
		{"persisted_retained_execution", &retainedExecutionStop{Cause: domain.Failure(apierrors.DeliveryUnknown, "blocked and reservation retained")}, true},
		{"structural_evidence_missing", domain.Failure(apierrors.EvidenceMissing, "original claim is missing"), false},
		{"identity_conflict", domain.Failure(apierrors.ContextStale, "original tuple changed"), false},
		{"database_error", &apierrors.ServiceError{Code: apierrors.InternalError}, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if retainedStopFailure(tc.err) != tc.retained {
				t.Fatal("stop error crossed its explicit retention boundary")
			}
		})
	}
}

func TestDynamicDAGUnknownUsageBoundaries(t *testing.T) {
	three, five, six := int64(3), int64(5), int64(6)
	budget := domain.Budget{Calls: &three, Tokens: &five}
	reserved := domain.Resources{Calls: 3, Tokens: 5}
	for _, tc := range []struct {
		name     string
		usage    domain.Usage
		retained bool
	}{
		{"both_unknown", domain.Usage{}, true},
		{"one_unknown_known_part_bounded", domain.Usage{Tokens: &five}, true},
		{"unknown_with_known_overrun", domain.Usage{Tokens: &six}, false},
		{"both_known", domain.Usage{Calls: &three, Tokens: &five}, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if unknownUsageWithinKnownBounds(tc.usage, reserved, budget) != tc.retained {
				t.Fatal("unknown usage hid a known overrun or changed a known result")
			}
		})
	}
}
