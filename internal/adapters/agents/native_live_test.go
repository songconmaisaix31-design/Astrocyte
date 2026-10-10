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

func TestNativeLiveTurnAndOriginalSessionResume(t *testing.T) {
	if os.Getenv("ASTROCYTE_TEST_NATIVE_MODEL") != "1" {
		t.Skip("coordinator sole live model/native slot required")
	}
	registry := NewRegistry(t.TempDir())
	for _, cli := range registry.List() {
		if !t.Run(cli, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 180*time.Second)
			defer cancel()
			project := domain.LocalProject{ID: uuid.NewString(), Root: t.TempDir(), Settings: domain.ProjectSettings{Revision: 1}}
			marker := "PUBLIC_CONTEXT_" + strings.ReplaceAll(uuid.NewString(), "-", "")
			packet := domain.ContextPacket{ID: uuid.NewString(), SchemaVersion: 1, ProjectID: project.ID, SettingsRevision: 1, Mode: "selected_context", Materials: []domain.ContextMaterial{{Reference: domain.FixedReference{MaterialID: "public-test-marker", Revision: 1}, Text: marker}}}
			n := registry.adapters[cli]
			sid := uuid.NewString()
			s, err := n.Start(ctx, domain.NativeRequest{SessionID: sid, Project: project, Session: domain.NativeSession{ID: sid, ProjectID: project.ID}, Command: domain.NativeCommand{DeadlineSeconds: 180, Message: "Reply with the supplied public context marker only."}, Packet: packet})
			if s.NativeID != "" {
				defer func() {
					obs, err := n.Stop(context.Background(), s)
					if err != nil || !obs.StopConfirmed {
						t.Errorf("cleanup positive owned exit: %+v %v", obs, err)
					}
				}()
			}
			if err != nil {
				t.Fatal(err)
			}
			wait := func(s domain.NativeSession) domain.NativeObservation {
				timer := time.NewTicker(100 * time.Millisecond)
				defer timer.Stop()
				for {
					obs, err := n.Observe(ctx, s)
					if err != nil {
						t.Fatal(err)
					}
					if obs.Status == "failed" || obs.Status == "blocked" || obs.OutputTruncated {
						t.Fatalf("actual native result: %+v", obs)
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
			obs := wait(s)
			var text strings.Builder
			for _, e := range obs.Events {
				if e.Kind == "text" {
					text.WriteString(e.Text)
				}
			}
			if !strings.Contains(text.String(), marker) {
				t.Fatalf("native output omitted selected marker; chars=%d", text.Len())
			}
			originalNativeID := s.NativeID
			obs, err = n.Stop(ctx, s)
			if err != nil || !obs.StopConfirmed {
				t.Fatalf("original stop not confirmed: %+v %v", obs, err)
			}
			s.StopConfirmed = true
			s.Status = "stopped"
			resumed, err := n.Resume(ctx, domain.NativeRequest{SessionID: sid, Project: project, Session: s, Command: domain.NativeCommand{DeadlineSeconds: 180, Message: "What was the public context marker from the earlier turn? Reply with that marker only."}, Packet: domain.ContextPacket{ID: uuid.NewString(), SchemaVersion: 1, ProjectID: project.ID, SettingsRevision: 1, Mode: "native_resume"}})
			if err != nil {
				t.Fatal(err)
			}
			s = resumed
			if resumed.NativeID != originalNativeID {
				t.Fatal("native resume silently created another session")
			}
			obs = wait(resumed)
			text.Reset()
			for _, e := range obs.Events {
				if e.Kind == "text" {
					text.WriteString(e.Text)
				}
			}
			if !strings.Contains(text.String(), marker) {
				t.Fatalf("original native context did not survive resume; chars=%d", text.Len())
			}
			obs, err = n.Stop(ctx, resumed)
			if err != nil || !obs.StopConfirmed {
				t.Fatalf("resumed stop not confirmed: %+v %v", obs, err)
			}
			process, _ := n.get(resumed)
			t.Logf("cli=%s version=%s model=%s provider=%s actual_turns=2 same_native_id=true stop_confirmed=true", cli, resumed.Version, process.model, process.provider)
		}) {
			break
		}
	}
}
