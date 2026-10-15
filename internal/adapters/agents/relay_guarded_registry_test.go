package agents

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/songconmaisaix31-design/Astrocyte/internal/apierrors"
	"github.com/songconmaisaix31-design/Astrocyte/internal/workspace/domain"
)

// 只核真实缺失前置的拒绝，不启动控制器或模拟协议成功。
func TestRelayGuardedRegistryRequiresOriginalOperationAndController(t *testing.T) {
	if registry, err := NewRelayGuardedRegistry("unopened-guarded-root", nil); err == nil || registry != nil {
		t.Fatal("missing source checkpoint was accepted")
	}
	called := false
	registry, err := NewRelayGuardedRegistry("unopened-guarded-root", func(context.Context, domain.NativeSession, *domain.RelaySessionArchive, string) error {
		called = true
		return errors.New("source denied")
	})
	if err != nil {
		t.Fatal(err)
	}
	adapter, err := registry.Adapter("codex")
	if err != nil {
		t.Fatal(err)
	}
	native := adapter.(*Native)
	_, err = native.Start(context.Background(), domain.NativeRequest{})
	var prelaunch *domain.NativePrelaunchFailure
	if !errors.As(err, &prelaunch) {
		t.Fatal("guarded startup without original operation reached launch")
	}
	session := domain.NativeSession{ID: uuid.NewString(), NativeID: uuid.NewString(), ProjectID: uuid.NewString(), CLI: "codex", Ownership: "owned", PendingOperation: uuid.NewString()}
	session.Operations = map[string]domain.NativeOperation{session.PendingOperation: {Action: "send", Status: "pending"}}
	_, err = native.Send(context.Background(), session, "public fixed input")
	var service *apierrors.ServiceError
	if !errors.As(err, &service) || service.Code != apierrors.DeliveryUnknown || called || len(native.processes) != 0 {
		t.Fatal("missing original controller produced a guarded input")
	}
}
