package agents

import (
	"context"
	"errors"
	"github.com/google/uuid"
	"github.com/songconmaisaix31-design/Astrocyte/internal/apierrors"
	"github.com/songconmaisaix31-design/Astrocyte/internal/workspace/app"
	"github.com/songconmaisaix31-design/Astrocyte/internal/workspace/domain"
	"strings"
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
	_, err = n.Send(ctx, request.Session, "public fixed task")
	assert(err, apierrors.DeliveryUnknown)
	request.Session.PendingOperation = uuid.NewString()
	request.Session.Operations = map[string]domain.NativeOperation{request.Session.PendingOperation: {Action: "start", Status: "pending"}}
	_, err = n.Send(ctx, request.Session, "public fixed task")
	assert(err, apierrors.ScopeDenied)
	request.Session.Operations[request.Session.PendingOperation] = domain.NativeOperation{Action: "send", Status: "pending"}
	operation, operationErr := nativeRequestOperation(request, "send")
	if operationErr != nil || operation != request.Session.PendingOperation {
		t.Fatal("original pending send receipt lost its operation identity")
	}
	_, err = n.Send(ctx, request.Session, "public fixed task")
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

func TestRelaySendRequiresOriginalControllerBeforeRestrictedAdmission(t *testing.T) {
	const input = "PRIVATE_UNSUBMITTED_INPUT_CANARY"
	for _, provider := range []string{"codex", "claude", "pi"} {
		t.Run(provider, func(t *testing.T) {
			registry := NewRegistry("unopened-controller-root")
			native := registry.adapters[provider].(*Native)
			operation := uuid.NewString()
			session := domain.NativeSession{ID: uuid.NewString(), NativeID: uuid.NewString(), ProjectID: uuid.NewString(), CLI: provider, Ownership: "owned", PendingOperation: operation, Operations: map[string]domain.NativeOperation{operation: {Action: "send", Status: "pending"}}}
			launch := app.RelayLaunch{Spec: domain.AgentSpec{Name: provider}, Session: session, OperationID: operation}
			driver := NewRelayNativeDriver(registry)
			for _, send := range []struct {
				name string
				run  func() error
			}{
				{"native_send", func() error { _, err := native.send(context.Background(), session, input, operation); return err }},
				{"pending_send", func() error { _, err := native.Send(context.Background(), session, input); return err }},
				{"operation_send", func() error {
					_, err := native.SendOperation(context.Background(), session, input, operation)
					return err
				}},
				{"relay_send", func() error { _, err := driver.Send(context.Background(), launch, input); return err }},
			} {
				t.Run(send.name, func(t *testing.T) {
					err := send.run()
					var known *apierrors.ServiceError
					var notSubmitted *domain.RelayInputNotSubmitted
					if !errors.As(err, &known) || known.Code != apierrors.DeliveryUnknown || errors.As(err, &notSubmitted) || strings.Contains(err.Error(), input) {
						t.Fatal("missing original controller must retain unknown delivery without exposing input")
					}
				})
			}
			if len(native.processes) != 0 || len(native.relayCaptures) != 0 || session.Operations[operation].Status != "pending" {
				t.Fatal("missing original controller changed pending input or created native effects")
			}
		})
	}
}
