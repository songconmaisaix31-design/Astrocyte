package agents

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/songconmaisaix31-design/Astrocyte/internal/workspace/domain"
)

func TestProjectScopeDefaultDenyAndDirectoryGranularity(t *testing.T) {
	root := t.TempDir()
	_ = os.Mkdir(filepath.Join(root, "allowed"), 0700)
	_ = os.Mkdir(filepath.Join(root, "other"), 0700)
	for _, path := range []string{"allowed/a.txt", "other/b.txt", "allowed/.env"} {
		if err := os.WriteFile(filepath.Join(root, filepath.FromSlash(path)), []byte("selected"), 0600); err != nil {
			t.Fatal(err)
		}
	}
	p := domain.LocalProject{Root: root}
	ctx := context.Background()
	if _, err := ReadProjectFiles(ctx, p, []string{"allowed/a.txt"}); err == nil {
		t.Fatal("default A-only unexpectedly read directory")
	}
	p.Settings.AllowDirectory = true
	p.Settings.AllowedSubdirs = []string{"allowed"}
	files, err := ReadProjectFiles(ctx, p, []string{"allowed/a.txt"})
	if err != nil || len(files) != 1 || files[0].Version == "" {
		t.Fatalf("approved file read: %+v %v", files, err)
	}
	for _, path := range []string{"other/b.txt", "../escape.txt", "allowed/.env", root} {
		if _, err := ReadProjectFiles(ctx, p, []string{path}); err == nil {
			t.Fatalf("outside path read: %s", path)
		}
	}
	p.Settings.AllowDirectory = false
	if _, err := ReadProjectFiles(ctx, p, []string{"allowed/a.txt"}); err == nil {
		t.Fatal("revoked B still readable")
	}
	if _, err := CanonicalRoot(""); err == nil {
		t.Fatal("empty root silently registered cwd")
	}
}

func TestProjectSymlinkEscapeAndOutputBound(t *testing.T) {
	root := t.TempDir()
	outside := t.TempDir()
	if err := os.WriteFile(filepath.Join(outside, "private.txt"), []byte("private"), 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(outside, filepath.Join(root, "escape")); err == nil {
		if _, err := ValidateProjectPath(root, "escape/private.txt"); err == nil {
			t.Fatal("symlink escape allowed")
		}
	} else {
		t.Logf("symlink creation unavailable; separate bounded read check still runs: %v", err)
	}
	if err := os.WriteFile(filepath.Join(root, "large.txt"), []byte(strings.Repeat("x", 128*1024+1)), 0600); err != nil {
		t.Fatal(err)
	}
	p := domain.LocalProject{Root: root, Settings: domain.ProjectSettings{AllowDirectory: true, AllowedSubdirs: []string{"."}}}
	if _, err := ReadProjectFiles(context.Background(), p, []string{"large.txt"}); err == nil {
		t.Fatal("large file escaped output bound")
	}
}

func TestHistoryDiscoveryRequiresChosenRootAndFiltersBeforeBody(t *testing.T) {
	root := t.TempDir()
	history := t.TempDir()
	ctx := context.Background()
	n := NewRegistry(t.TempDir()).adapters["pi"]
	p := domain.LocalProject{ID: uuid.NewString(), Root: root, Settings: domain.ProjectSettings{Revision: 1}}
	if _, err := n.Discover(ctx, p); err == nil {
		t.Fatal("private history guessed without chosen root")
	}
	header := func(cwd, id string) string {
		b, _ := json.Marshal(map[string]any{"type": "session", "version": 3, "cwd": cwd, "id": id})
		return string(b) + "\n"
	}
	if err := os.WriteFile(filepath.Join(history, "selected.jsonl"), []byte(header(root, "real-native")+`{"type":"message","message":{"role":"assistant","content":[{"type":"text","text":"retained result"}]}}`+"\n"), 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(history, "unapproved.jsonl"), []byte(header(t.TempDir(), "foreign")+"invalid private body must not be parsed"), 0600); err != nil {
		t.Fatal(err)
	}
	p.Settings.HistoryRoots = map[string]string{"pi": history}
	sessions, err := n.Discover(ctx, p)
	if err != nil || len(sessions) != 1 || sessions[0].NativeID != "real-native" || sessions[0].Ownership != "external_observed" || sessions[0].StopConfirmed {
		t.Fatalf("history scope mismatch %+v %v", sessions, err)
	}
	events, err := n.ReadContext(ctx, sessions[0])
	if err != nil || len(events) != 1 || events[0].Text != "retained result" {
		t.Fatalf("native historical read %+v %v", events, err)
	}
	if _, err := n.Resume(ctx, domain.NativeRequest{Session: sessions[0]}); err == nil {
		t.Fatal("read-only history falsely resumed")
	}
	if _, err := n.Stop(ctx, sessions[0]); err == nil {
		t.Fatal("external process falsely stopped")
	}
}

func TestNativeHandshakeLiveNoInference(t *testing.T) {
	if os.Getenv("ASTROCYTE_TEST_NATIVE_HANDSHAKE") != "1" {
		t.Skip("explicit no-inference native protocol smoke")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	registry := NewRegistry(t.TempDir())
	for _, cli := range registry.List() {
		t.Run(cli, func(t *testing.T) {
			p := domain.LocalProject{ID: uuid.NewString(), Root: t.TempDir(), Settings: domain.ProjectSettings{Revision: 1}}
			n := registry.adapters[cli]
			id := uuid.NewString()
			s, err := n.Start(ctx, domain.NativeRequest{SessionID: id, Project: p, Session: domain.NativeSession{ID: id, ProjectID: p.ID}, Command: domain.NativeCommand{DeadlineSeconds: 30}, Packet: domain.ContextPacket{SchemaVersion: 1, ProjectID: p.ID}})
			if err != nil {
				t.Fatal(err)
			}
			defer n.Stop(context.Background(), s)
			if s.NativeID == "" || s.Version == "native_version_not_observed" {
				t.Fatalf("native ID/version unobserved: %+v", s)
			}
			obs, err := n.Observe(ctx, s)
			if err != nil || obs.StopConfirmed || obs.Status != "idle" {
				t.Fatalf("native idle: %+v %v", obs, err)
			}
			obs, err = n.Stop(ctx, s)
			if err != nil || !obs.StopConfirmed {
				t.Fatalf("owned exit: %+v %v", obs, err)
			}
			t.Logf("cli=%s version=%s native_identity_present=true stop_confirmed=%v no_model_turn=true", cli, s.Version, obs.StopConfirmed)
		})
	}
}
