package app

import (
	"context"

	"github.com/songconmaisaix31-design/Astrocyte/internal/apierrors"
	"github.com/songconmaisaix31-design/Astrocyte/internal/workspace/domain"
)

type NativeMessageScope struct {
	Kind      string `json:"-"`
	BindingID string `json:"-"`
}

type NativeMessageLease struct {
	WorkItemID string `json:"-"`
	MissionID  string `json:"-"`
	Fence      int64  `json:"-"`
}

type NativeMessageAuthorization struct {
	Caller      domain.Caller            `json:"-"`
	Preparation NativeMessagePreparation `json:"-"`
}

// 回调只读写同一主库；发送回调的 context 携带当前授权的最早截止时间。
type NativeMessageAuthorizer interface {
	WithNativeMessagePreparation(context.Context, NativeMessageAuthorization, func(context.Context) error) error
	WithNativeMessageSend(context.Context, NativeMessageAuthorization, func(context.Context) error) error
}

type SwarmNativeMessagePreparationStore interface {
	ReadSwarmNativeMessagePreparation(context.Context, string, string) (NativeMessagePreparation, error)
	BindSwarmNativeMessageLease(context.Context, string, string, NativeMessageLease) error
}

func (s *localProjectService) WithNativeMessageAuthorizer(authorizer NativeMessageAuthorizer) *localProjectService {
	s.nativeMessageMu.Lock()
	defer s.nativeMessageMu.Unlock()
	s.nativeMessageOptions.authorizer = authorizer
	return s
}

func (s *localProjectService) ConfigureNativeMessageAuthorizer(authorizer NativeMessageAuthorizer) error {
	if authorizer == nil {
		return projectError(apierrors.UnsupportedCapability, "actual native send authorizer is required")
	}
	s.WithNativeMessageAuthorizer(authorizer)
	return nil
}

func ValidateNativeMessageScope(scope NativeMessageScope) error {
	if scope.Kind == "project" && scope.BindingID == "" || scope.Kind == "swarm" && scope.BindingID != "" {
		return nil
	}
	return projectError(apierrors.EvidenceMissing, "native preparation requires its explicit original purpose")
}

func NativeMessageRequiresSwarmLease(actions []string) (bool, error) {
	if len(actions) == 0 {
		return false, projectError(apierrors.EvidenceMissing, "original approved actions are required to determine input purpose")
	}
	swarm := false
	for _, action := range actions {
		if !projectActionKnown(action) {
			return false, projectError(apierrors.EvidenceMissing, "original approval contains an unknown protected action")
		}
		if action == "claim_work_item" || action == "submit_artifact" {
			swarm = true
		}
	}
	return swarm, nil
}

func (s *localProjectService) PrepareSwarmNativeMessage(ctx context.Context, caller domain.Caller, projectID, sessionID, bindingID, message string) (domain.NativeSession, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.prepareNativeMessageForScope(ctx, caller, projectID, sessionID, message, NativeMessageScope{Kind: "swarm", BindingID: bindingID})
}

func (s *localProjectService) ReadSwarmNativeMessagePreparation(ctx context.Context, sessionID, operationID string) (NativeMessagePreparation, error) {
	if _, err := s.ReadPreparedNativeSession(ctx, sessionID, operationID); err != nil {
		return NativeMessagePreparation{}, err
	}
	repository, _, err := s.nativeMessageServices()
	if err != nil {
		return NativeMessagePreparation{}, err
	}
	prepared, err := repository.LoadNativeMessagePreparation(ctx, sessionID, operationID)
	if err != nil {
		return prepared, err
	}
	if prepared.Scope.Kind != "swarm" || prepared.Scope.BindingID == "" {
		return NativeMessagePreparation{}, projectError(apierrors.ScopeDenied, "work item claim requires an explicit Swarm preparation")
	}
	return prepared, nil
}

func (s *localProjectService) BindSwarmNativeMessageLease(ctx context.Context, sessionID, operationID string, lease NativeMessageLease) error {
	repository, _, err := s.nativeMessageServices()
	if err != nil {
		return err
	}
	binder, ok := repository.(interface {
		BindSwarmNativeMessageLease(context.Context, string, string, NativeMessageLease) error
	})
	if !ok {
		return projectError(apierrors.UnsupportedCapability, "actual prepared lease binder is required")
	}
	return binder.BindSwarmNativeMessageLease(ctx, sessionID, operationID, lease)
}
