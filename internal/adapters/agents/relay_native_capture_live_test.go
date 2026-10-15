package agents

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/songconmaisaix31-design/Astrocyte/internal/workspace/domain"
)

func TestRelayActualCodexProtocolSourceStart(t *testing.T) {
	if os.Getenv("ASTROCYTE_RELAY_L1_NATIVE") != "1" {
		t.Skip("requires authorized actual Codex protocol startup")
	}
	base := os.Getenv("ASTROCYTE_RELAY_L1_EVIDENCE")
	if !filepath.IsAbs(base) {
		t.Fatal("absolute ignored evidence directory required")
	}
	root := filepath.Join(base, "protocol-"+uuid.NewString())
	if err := os.MkdirAll(root, 0700); err != nil {
		t.Fatal(err)
	}
	for _, args := range [][]string{{"init", root}, {"-C", root, "-c", "user.name=Relay L1", "-c", "user.email=relay-l1@example.invalid", "-c", "core.hooksPath=", "commit", "--allow-empty", "-m", "Public native protocol context"}} {
		if err := exec.Command("git", args...).Run(); err != nil {
			t.Fatal(err)
		}
	}
	registry := NewRegistry(filepath.Join(root, "controller"))
	n := registry.adapters["codex"].(*Native)
	ctx, cancel := context.WithTimeout(context.Background(), 40*time.Second)
	defer cancel()
	id := uuid.NewString()
	request := domain.NativeRequest{SessionID: id, Project: domain.LocalProject{ID: uuid.NewString(), Root: root, Settings: domain.ProjectSettings{Revision: 1}}, Session: domain.NativeSession{ID: id}, Command: domain.NativeCommand{DeadlineSeconds: 60}}
	session, startErr := n.StartRelay(ctx, request)
	t.Cleanup(func() {
		obs, err := n.Stop(context.Background(), session)
		if err != nil || !obs.StopConfirmed {
			t.Error("actual controller cleanup exit unconfirmed")
		}
	})
	n.mu.Lock()
	capture := n.relayCaptures[id]
	n.mu.Unlock()
	capture.mu.Lock()
	diagnostic := capture.diagnosticMethod
	capture.mu.Unlock()
	if diagnostic != "" {
		if err := os.WriteFile(filepath.Join(root, "protocol-method-diagnostic.txt"), []byte(diagnostic), 0600); err != nil {
			t.Fatal(err)
		}
	}
	state := capture.snapshot(false)
	if startErr != nil || state.Errors != 0 || state.Unknown != 0 || state.PrefixComplete != "confirmed" {
		t.Fatalf("actual source startup failed; first_gap=%s errors=%d unknown=%d", state.FirstGap, state.Errors, state.Unknown)
	}
	archive, err := n.OwnedRelayArchive(ctx, session)
	if err != nil || archive.Path == "" || archive.Finalized {
		t.Fatal("actual new thread metadata source unavailable or already finalized")
	}
	// 直接停止真实进程没有 thread/closed 确认，不能形成归档收尾确证。
	stopped, err := n.Stop(ctx, session)
	if err != nil || !stopped.StopConfirmed {
		t.Fatal("actual owned forced stop did not confirm process exit")
	}
	select {
	case <-capture.readerDone:
	case <-ctx.Done():
		t.Fatal("actual notification reader did not terminate")
	}
	ended, captureErr := n.OwnedRelayNativeCapture(ctx, session)
	if captureErr == nil || ended.Finalized != "unknown" || ended.FirstGap != "native_stream_ended_without_shutdown_ack" || ended.Unknown == 0 {
		t.Fatal("actual process exit fabricated stream or archive finalization")
	}
	archive, err = n.OwnedRelayArchive(ctx, session)
	if err != nil || archive.Finalized {
		t.Fatal("forced exit promoted the selected archive to finalized")
	}
	t.Logf("actual_native_start=true version=%s source_prefix=confirmed notification_records=%d model_calls=0 actual_forced_exit=true stream_finalized=unknown archive_finalized=unknown", n.Version(), len(state.Events))
}
