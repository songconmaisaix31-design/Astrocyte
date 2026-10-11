package app

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/songconmaisaix31-design/Astrocyte/internal/apierrors"
	"github.com/songconmaisaix31-design/Astrocyte/internal/workspace/domain"
)

type WorksRequestReceipt struct {
	Input  []byte
	Status string
	Result json.RawMessage
	Error  *apierrors.ServiceError
}

// The transport supplies Caller.OperationID from Idempotency-Key. Reservations
// survive process loss and stop an uncertain command from being executed again.
type WorksRequestRepository interface {
	BeginWorksRequest(context.Context, string, string, string, []byte) (WorksRequestReceipt, bool, error)
	FinishWorksRequest(context.Context, string, string, string, WorksRequestReceipt) error
}

func worksRequest[T any](ctx context.Context, s *worksService, caller domain.Caller, command string, input any, run func() (T, error)) (T, error) {
	var zero T
	if err := human(caller); err != nil {
		return zero, err
	}
	// Existing internal callers remain transport-free. HTTP requires an operation
	// key, and a keyed call must have durable receipt storage before any effect.
	if caller.OperationID == "" {
		return run()
	}
	repo, ok := s.repo.(WorksRequestRepository)
	if !ok {
		return zero, workError(apierrors.UnsupportedCapability, "durable works requests are not connected")
	}
	payload, err := json.Marshal(input)
	if err != nil {
		return zero, err
	}
	actor := caller.Kind + ":" + caller.ID
	receipt, fresh, err := repo.BeginWorksRequest(ctx, actor, command, caller.OperationID, payload)
	if err != nil {
		return zero, err
	}
	if !fresh {
		switch receipt.Status {
		case "accepted":
			if err := json.Unmarshal(receipt.Result, &zero); err != nil {
				return zero, workError(apierrors.DeliveryUnknown, "stored works result cannot be read; do not replay")
			}
			return zero, nil
		case "rejected":
			if receipt.Error != nil {
				return zero, receipt.Error
			}
		}
		return zero, workError(apierrors.DeliveryUnknown, "works request is pending or has unknown effects; reload without replay")
	}
	result, effectErr := run()
	receipt.Status = "accepted"
	if effectErr != nil {
		receipt.Status = "unknown"
		var known *apierrors.ServiceError
		if errors.As(effectErr, &known) {
			receipt.Error = known
			switch known.Code {
			case apierrors.ValidationFailed, apierrors.NotFound, apierrors.VersionConflict, apierrors.ScopeDenied, apierrors.ApprovalRequired, apierrors.UnsupportedCapability:
				receipt.Status = "rejected"
			}
		}
	} else {
		receipt.Result, err = json.Marshal(result)
		if err != nil {
			return zero, workError(apierrors.DeliveryUnknown, "works effect completed but result could not be saved; reload without replay")
		}
	}
	if receipt.Status == "unknown" {
		if receipt.Error == nil || receipt.Error.Code != apierrors.DeliveryUnknown {
			unknown := workError(apierrors.DeliveryUnknown, "works command effects are unknown; reload without replay")
			errors.As(unknown, &receipt.Error)
		}
		effectErr = receipt.Error
	}
	if err := repo.FinishWorksRequest(ctx, actor, command, caller.OperationID, receipt); err != nil {
		return result, workError(apierrors.DeliveryUnknown, "works receipt could not be finalized; reload without replay")
	}
	return result, effectErr
}
