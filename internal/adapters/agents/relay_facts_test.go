package agents

import (
	"context"
	"errors"
	"github.com/google/uuid"
	"github.com/songconmaisaix31-design/Astrocyte/internal/apierrors"
	"github.com/songconmaisaix31-design/Astrocyte/internal/workspace/app"
	"github.com/songconmaisaix31-design/Astrocyte/internal/workspace/domain"
	"testing"
)

// 只核实际空 Registry 的拒绝，不构造 owned 进程或成功原生回复。
func TestRelayFactsRejectMissingOriginalControllerAndFork(t *testing.T) {
	ctx := context.Background()
	registry := NewRegistry("unopened-controller-root")
	a, err := registry.Adapter("codex")
	if err != nil {
		t.Fatal(err)
	}
	n := a.(*Native)
	request := domain.NativeRequest{Session: domain.NativeSession{ID: uuid.NewString(), NativeID: uuid.NewString(), ProjectID: uuid.NewString(), CLI: "codex", Ownership: "owned", StopConfirmed: true}}
	assert := func(err error, code apierrors.Code) {
		t.Helper()
		var known *apierrors.ServiceError
		if !errors.As(err, &known) || known.Code != code {
			t.Fatal("original controller refusal differs")
		}
	}
	_, err = n.StartOperation(ctx, request, "")
	assert(err, apierrors.ValidationFailed)
	_, err = n.ResumeOperation(ctx, request, uuid.NewString())
	assert(err, apierrors.DeliveryUnknown)
	_, err = n.SendOperation(ctx, request.Session, "public fixed task", uuid.NewString())
	assert(err, apierrors.DeliveryUnknown)
	evidence, err := n.OwnedRelayControllerEvidence(ctx, request.Session)
	assert(err, apierrors.DeliveryUnknown)
	if evidence.ControllerID != "" || len(evidence.Facts) != 0 {
		t.Fatal("missing controller produced ownership or execution evidence")
	}
	_, err = NewRelayNativeDriver(registry).Fork(ctx, app.RelayLaunch{})
	assert(err, apierrors.UnsupportedCapability)
	_, err = NewRelayNativeDriver(registry).CurrentControllerEvidence(ctx, app.RelayLaunch{Spec: domain.AgentSpec{Name: "codex"}, Session: request.Session})
	assert(err, apierrors.DeliveryUnknown)
	_, err = registry.ReadOwnedNativeSession(ctx, request.Session.ID)
	assert(err, apierrors.DeliveryUnknown)
	_, err = nativeRequestOperation(domain.NativeRequest{Session: domain.NativeSession{PendingOperation: uuid.NewString()}}, "start")
	assert(err, apierrors.ScopeDenied)
	if len(n.processes) != 0 {
		t.Fatal("rejection created a controller process")
	}
}
