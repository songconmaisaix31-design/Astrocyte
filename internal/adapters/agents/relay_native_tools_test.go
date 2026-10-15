package agents

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/songconmaisaix31-design/Astrocyte/internal/apierrors"
	"github.com/songconmaisaix31-design/Astrocyte/internal/workspace/app"
	"github.com/songconmaisaix31-design/Astrocyte/internal/workspace/domain"
)

func TestRelayCodexFileToolRestrictionRejectsBeforeNativeEffects(t *testing.T) {
	const privateInput = "PRIVATE_UNSUBMITTED_INPUT_CANARY"
	registry := NewRegistry("unopened-restricted-codex-root")
	native := registry.adapters["codex"].(*Native)
	request := domain.NativeRequest{SessionID: uuid.NewString(), Command: domain.NativeCommand{Message: privateInput}}
	assert := func(session domain.NativeSession, err error) {
		t.Helper()
		var known *apierrors.ServiceError
		var unstarted *domain.NativePrelaunchFailure
		if !errors.As(err, &known) || known.Code != apierrors.UnsupportedCapability || !errors.As(err, &unstarted) || session.NativeID != "" || session.Ownership == "owned" || strings.Contains(err.Error(), privateInput) {
			t.Fatal("unsupported file tool restriction was not rejected before native effects")
		}
	}
	assert(native.Start(context.Background(), request))
	request.Command.Message = ""
	assert(native.StartRelay(context.Background(), request))
	assert(native.ResumeRelayOperation(context.Background(), request, uuid.NewString()))
	guard := func(context.Context, domain.NativeSession, *domain.RelaySessionArchive, string) error {
		panic("unsupported restricted native launch must not reach a source checkpoint")
	}
	assert(native.StartRelayOperationWithInputGuard(context.Background(), request, uuid.NewString(), guard))
	driver := NewRelayNativeDriver(registry, RelayNativeDriverOptions{CaptureNativeNotifications: true, BeforeFirstInput: guard})
	launch := app.RelayLaunch{Spec: domain.AgentSpec{Name: "codex"}, OperationID: uuid.NewString()}
	for _, effect := range []func(context.Context, app.RelayLaunch) (app.RelayObservation, error){driver.Initialize, driver.Launch, driver.Resume} {
		observation, err := effect(context.Background(), launch)
		assert(observation.Session, err)
	}
	if len(native.processes) != 0 || len(native.relayCaptures) != 0 {
		t.Fatal("unsupported file tool restriction created a controller or capture")
	}
}
