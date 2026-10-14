package agents

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/songconmaisaix31-design/Astrocyte/internal/apierrors"
	"github.com/songconmaisaix31-design/Astrocyte/internal/workspace/domain"
)

func TestSelectedTextFinalPromptUTF8BoundaryBeforeLaunch(t *testing.T) {
	packet := domain.ContextPacket{SchemaVersion: 1, ProjectID: "00000000-0000-0000-0000-000000000000", Mode: "selected_text"}
	overhead := len(packetPrompt(packet, ""))
	if SelectedTextPromptBytes("文") != overhead+3 {
		t.Fatal("wire UTF8 measurement differs from actual envelope")
	}
	for _, size := range []int{128*1024 + 1, 512*1024 - overhead} {
		prompt := strings.Repeat("x", size)
		if err := ValidateSelectedTextPrompt(prompt); err != nil {
			t.Fatalf("%d total bytes: %v", size+overhead, err)
		}
	}
	for _, prompt := range []string{strings.Repeat("x", 512*1024-overhead+1), strings.Repeat("文", (512*1024-overhead)/3+1), "\xff"} {
		// 空 Registry 没有原生适配器，格式错误必须在选择配置之前失败。
		_, err := (&Registry{}).ProcessSelectedText(context.Background(), domain.TextRequest{CLI: "codex", Prompt: prompt})
		serviceErr, ok := err.(*apierrors.ServiceError)
		if !ok || serviceErr.Code != apierrors.ValidationFailed {
			t.Fatalf("oversize reached native selection: %v", err)
		}
	}
	_, err := (&Registry{}).ProcessSelectedText(context.Background(), domain.TextRequest{CLI: "uninstalled", Prompt: strings.Repeat("x", 128*1024+1)})
	if e, ok := err.(*apierrors.ServiceError); !ok || e.Code != apierrors.UnsupportedCapability {
		t.Fatalf("valid expanded prompt rejected before actual registry selection: %v", err)
	}
}

func TestSelectedTextDirectoryStaysUnderControllerOwnedRoot(t *testing.T) {
	stateRoot := filepath.Join(t.TempDir(), "native-sessions")
	root, err := selectedTextRoot(stateRoot)
	if err != nil {
		t.Fatal(err)
	}
	if !inside(filepath.Join(stateRoot, "selected-text"), root) {
		t.Fatal("selected text directory escaped the controller-owned root")
	}
	info, err := os.Stat(root)
	if err != nil || !info.IsDir() {
		t.Fatalf("owned directory unavailable: %v", err)
	}
	if _, err = selectedTextRoot("relative-native-state"); err == nil {
		t.Fatal("relative runtime root was accepted")
	}
	blockedRoot := filepath.Join(t.TempDir(), "not-a-directory")
	if err = os.WriteFile(blockedRoot, []byte("人工创建的目录类型错误输入"), 0600); err != nil {
		t.Fatal(err)
	}
	if _, err = selectedTextRoot(blockedRoot); err == nil {
		t.Fatal("actual filesystem failure was suppressed")
	}
}
