package agents

import (
	"context"
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
		// A registry with no adapters would panic if launch/configuration were
		// reached; validation must reject first and preserve unknown effects.
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
