package agents

import (
	"context"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/songconmaisaix31-design/Astrocyte/internal/workspace/domain"
)

// TestOpencodeLiveSessionChain runs a real `opencode serve` session chain: start
// (idle), explicit send with a real paid turn, read context, resume the same
// native ID, stop, then a cross-session context handoff that must produce a new
// native ID while carrying the original selected context. It is gated so normal
// CI never launches the CLI or charges a model turn.
func TestOpencodeLiveSessionChain(t *testing.T) {
	if os.Getenv("ASTROCYTE_TEST_OPENCODE_LIVE") != "1" {
		t.Skip("real OpenCode serve session chain requires the sole paid slot")
	}
	registry := NewRegistry(t.TempDir())
	adapter, err := registry.Adapter("opencode")
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 300*time.Second)
	defer cancel()
	project := domain.LocalProject{ID: uuid.NewString(), Root: t.TempDir(), Settings: domain.ProjectSettings{Revision: 1}}
	marker := "PUBLIC_MARKER_" + strings.ReplaceAll(uuid.NewString(), "-", "")
	native := adapter.(*opencodeNative)

	newSession := func() domain.NativeSession {
		t.Helper()
		sid := uuid.NewString()
		s, err := adapter.Start(ctx, domain.NativeRequest{
			SessionID: sid,
			Project:   project,
			Session:   domain.NativeSession{ID: sid, ProjectID: project.ID},
			Command:   domain.NativeCommand{DeadlineSeconds: 180},
			Packet:    domain.ContextPacket{SchemaVersion: 1, ProjectID: project.ID, Mode: "selected_context"},
		})
		if err != nil {
			t.Fatalf("open code start: %v", err)
		}
		if s.NativeID == "" {
			t.Fatal("open code start did not report a native identity")
		}
		return s
	}
	text := func(obs domain.NativeObservation) string {
		var b strings.Builder
		for _, e := range obs.Events {
			if e.Kind == "text" {
				b.WriteString(e.Text)
			}
		}
		return b.String()
	}
	wait := func(s domain.NativeSession) domain.NativeObservation {
		t.Helper()
		timer := time.NewTicker(100 * time.Millisecond)
		defer timer.Stop()
		for {
			obs, err := adapter.Observe(ctx, s)
			if err != nil {
				t.Fatalf("observe: %v", err)
			}
			if obs.Status == "failed" || obs.Status == "blocked" || obs.OutputTruncated {
				t.Fatalf("native turn: %+v", obs)
			}
			if obs.Status == "completed" {
				return obs
			}
			select {
			case <-ctx.Done():
				t.Fatal("native result unknown after deadline; no automatic retry")
			case <-timer.C:
			}
		}
	}

	// 1. Start (idle, no message) then an explicit Send positive turn.
	s1 := newSession()
	sendObs, err := adapter.Send(ctx, s1, "Reply with exactly the word "+marker+" and nothing else.")
	if err != nil || sendObs.Status != "completed" {
		t.Fatalf("explicit send: %+v %v", sendObs, err)
	}
	if !strings.Contains(text(sendObs), marker) {
		t.Fatalf("explicit send output omitted the selected marker: %q", text(sendObs))
	}
	native.mu.Lock()
	proc := native.processes[s1.ID]
	native.mu.Unlock()
	proc.mu.Lock()
	model, provider := proc.model, proc.provider
	proc.mu.Unlock()
	if model == "" || provider == "" {
		t.Fatalf("model/provider not observed from the native reply schema: model=%q provider=%q", model, provider)
	}

	// 2. Read context: both the user marker and the assistant reply are present.
	history, err := adapter.ReadContext(ctx, s1)
	if err != nil {
		t.Fatal(err)
	}
	hasUser, hasAssistant := false, false
	for _, e := range history {
		if e.Kind == "user_text" && strings.Contains(e.Text, marker) {
			hasUser = true
		}
		if e.Kind == "text" && strings.Contains(e.Text, marker) {
			hasAssistant = true
		}
	}
	if !hasUser || !hasAssistant {
		t.Fatalf("read context missing user/assistant marker: %+v", history)
	}

	// 3. Stop and resume the same native ID; a second real turn continues it.
	stopObs, err := adapter.Stop(ctx, s1)
	if err != nil || !stopObs.StopConfirmed {
		t.Fatalf("open code stop: %+v %v", stopObs, err)
	}
	s1.StopConfirmed = true
	s1.Status = "stopped"
	resumed, err := adapter.Resume(ctx, domain.NativeRequest{
		SessionID: s1.ID,
		Project:   project,
		Session:   s1,
		Command:   domain.NativeCommand{DeadlineSeconds: 180, Message: "What was the earlier public marker? Reply with that exact word only."},
		Packet:    domain.ContextPacket{SchemaVersion: 1, ProjectID: project.ID, Mode: "selected_context"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if resumed.NativeID != s1.NativeID {
		t.Fatalf("resume changed the native identity: %q -> %q", s1.NativeID, resumed.NativeID)
	}
	resumeObs := wait(resumed)
	if !strings.Contains(text(resumeObs), marker) {
		t.Fatalf("resumed turn lost the original context marker: %q", text(resumeObs))
	}
	if _, err := adapter.Stop(ctx, resumed); err != nil {
		t.Fatal(err)
	}

	// 4. Cross-session context handoff: a new session gets a new native ID but
	// carries the original selected context and must reference it, not just echo
	// a fresh marker.
	handoffPacket := domain.ContextPacket{
		SchemaVersion: 1, ProjectID: project.ID, Mode: "context_handoff",
		Materials: []domain.ContextMaterial{{Reference: domain.FixedReference{MaterialID: "handoff-marker", Revision: 1}, Title: "prior-selected-context", Text: marker}},
	}
	s2id := uuid.NewString()
	s2, err := adapter.Start(ctx, domain.NativeRequest{
		SessionID: s2id,
		Project:   project,
		Session:   domain.NativeSession{ID: s2id, ProjectID: project.ID},
		Command:   domain.NativeCommand{DeadlineSeconds: 180, Message: "The prior selected context contained a public marker. Reply with exactly that marker word only."},
		Packet:    handoffPacket,
	})
	if err != nil {
		t.Fatal(err)
	}
	if s2.NativeID == "" || s2.NativeID == s1.NativeID {
		t.Fatalf("cross-session handoff reused the source native identity: %q", s2.NativeID)
	}
	handoffObs := wait(s2)
	if !strings.Contains(text(handoffObs), marker) {
		t.Fatalf("handoff session did not reference the carried context marker: %q", text(handoffObs))
	}
	if _, err := adapter.Stop(ctx, s2); err != nil {
		t.Fatal(err)
	}
	t.Logf("opencode live: version=%s model=%s provider=%s start_native_id=%s send=true read=true resume_same_id=true handoff_new_id=%s handoff_context=true", s1.Version, model, provider, s1.NativeID, s2.NativeID)
}
