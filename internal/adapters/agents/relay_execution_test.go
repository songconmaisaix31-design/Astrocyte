package agents

import (
	"context"
	"errors"
	"testing"

	"github.com/songconmaisaix31-design/Astrocyte/internal/apierrors"
	"github.com/songconmaisaix31-design/Astrocyte/internal/workspace/domain"
)

// 同时可在普通 Native 基线运行：只检查拒绝，不制造持有进程或成功状态。
func TestRelayExecutionPortableRejectsIncompleteIdentity(t *testing.T) {
	n := &Native{id: "codex"}
	for _, session := range []domain.NativeSession{
		{},
		{NativeID: "public-native", ProjectID: "public-project", CLI: "codex", Ownership: "owned"},
		{ID: "public-session", ProjectID: "public-project", CLI: "codex", Ownership: "owned"},
		{ID: "public-session", NativeID: "public-native", CLI: "codex", Ownership: "owned"},
		{ID: "public-session", NativeID: "public-native", ProjectID: "public-project", CLI: "codex", Ownership: "owned"},
	} {
		evidence, err := n.OwnedRelayExecution(context.Background(), session)
		var failure *apierrors.ServiceError
		if !errors.As(err, &failure) || failure.Code != apierrors.DeliveryUnknown || evidence.ProcessID != 0 || evidence.LaunchPolicy != "" || !evidence.StartedAt.IsZero() {
			t.Fatal("missing original identity or controller produced execution evidence")
		}
	}
}

func TestRelayExecutionPortableRejectsExternalOrWrongCLI(t *testing.T) {
	n := &Native{id: "codex"}
	for _, session := range []domain.NativeSession{
		{ID: "public-session", NativeID: "public-native", ProjectID: "public-project", CLI: "codex", Ownership: "external_observed"},
		{ID: "public-session", NativeID: "public-native", ProjectID: "public-project", CLI: "claude", Ownership: "owned"},
	} {
		_, err := n.OwnedRelayExecution(context.Background(), session)
		var failure *apierrors.ServiceError
		if !errors.As(err, &failure) || failure.Code != apierrors.ScopeDenied {
			t.Fatal("external session or wrong CLI passed the owned execution gate")
		}
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := n.OwnedRelayExecution(ctx, domain.NativeSession{}); !errors.Is(err, context.Canceled) {
		t.Fatal("canceled evidence read did not stop before ownership lookup")
	}
}
