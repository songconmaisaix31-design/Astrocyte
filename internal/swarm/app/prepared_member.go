package app

import (
	"context"
	"time"

	"github.com/songconmaisaix31-design/Astrocyte/internal/apierrors"
	"github.com/songconmaisaix31-design/Astrocyte/internal/swarm/domain"
)

// 只消费已核验且尚未发送的原输入身份，不携带正文或授权能力。
type PreparedMemberInput struct {
	OperationID        string    `json:"-"`
	SessionID          string    `json:"-"`
	NativeSessionID    string    `json:"-"`
	AdapterID          string    `json:"-"`
	AdapterVersion     string    `json:"-"`
	ControllerID       string    `json:"-"`
	StartupOperationID string    `json:"-"`
	ProjectID          string    `json:"-"`
	Root               string    `json:"-"`
	ActorID            string    `json:"-"`
	ActorKind          string    `json:"-"`
	ActorProjectID     string    `json:"-"`
	PreparedAt         time.Time `json:"-"`
}

type PreparedMemberInputReader interface {
	ReadPreparedMemberInput(context.Context, string, string) (PreparedMemberInput, error)
}

type PreparedMemberLeaseBinder interface {
	BindPreparedMemberLease(context.Context, domain.Lease) error
}

func ValidatePreparedMemberInput(ctx context.Context, p domain.Principal, m Member, input PreparedMemberInput) error {
	b, err := AuthorizedBindingForMember(ctx, p, m)
	if err != nil {
		return err
	}
	if m.Purpose != "claim_work_item" {
		return domain.Failure(apierrors.ValidationFailed, "prepared input is only evidence for a new claim")
	}
	if input.ActorKind != p.Kind || input.ActorID != p.ID || input.ActorProjectID != p.ProjectID {
		return domain.Failure(apierrors.ScopeDenied, "prepared input belongs to a different authenticated member")
	}
	if input.OperationID != m.OperationID || input.OperationID == b.StartupOperationID || input.SessionID != b.SessionID || input.NativeSessionID != b.NativeSessionID || input.AdapterID != b.AdapterID || input.ControllerID != b.ControllerID || input.StartupOperationID != b.StartupOperationID || input.ProjectID != b.ProjectID || input.Root != b.Root {
		return domain.Failure(apierrors.ContextStale, "prepared input differs from the current authorized member binding")
	}
	if input.AdapterVersion == "" || input.PreparedAt.IsZero() {
		return domain.Failure(apierrors.EvidenceMissing, "prepared input lacks the original adapter version or preparation time")
	}
	return nil
}
