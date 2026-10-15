package agents

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/songconmaisaix31-design/Astrocyte/internal/apierrors"
	"github.com/songconmaisaix31-design/Astrocyte/internal/workspace/app"
	"github.com/songconmaisaix31-design/Astrocyte/internal/workspace/domain"
)

func TestRelayOwnedRootsReadOnlyDirectoryValidation(t *testing.T) {
	root := t.TempDir()
	private := filepath.Join(root, "not-a-directory")
	if err := os.WriteFile(private, []byte("public fixture"), 0600); err != nil {
		t.Fatal(err)
	}
	validator := RelayOwnedRoots{}
	if _, err := validator.CanonicalRoot(root); err != nil {
		t.Fatal(err)
	}
	if _, err := validator.CanonicalRoot(private); err == nil {
		t.Fatal("file used as neutral root")
	}
	if _, err := validator.CanonicalRoot("relative-root"); err == nil {
		t.Fatal("relative root accepted")
	}
	link := filepath.Join(root, "redirected")
	if err := os.Symlink(t.TempDir(), link); err != nil {
		t.Log("NOT_RUN: symlink fixture unsupported on host")
		return
	}
	if _, err := validator.CanonicalRoot(link); err == nil {
		t.Fatal("redirected relay root accepted")
	}
}

// This is a public controlled RPC fixture on the actual nativeProcess port,
// not Codex startup, a model call, or a synthetic acceptance of a real session.
func relayMetadataRPCFixture(t *testing.T) (*Native, domain.NativeSession, chan map[string]json.RawMessage, map[string]any) {
	t.Helper()
	root := t.TempDir()
	session := domain.NativeSession{ID: "public-controller", NativeID: "public-owned-thread", CLI: "codex", ProjectID: "public-project", Ownership: "owned"}
	requestReader, requestWriter := io.Pipe()
	responseReader, responseWriter := io.Pipe()
	p := &nativeProcess{cli: "codex", cmd: &exec.Cmd{Dir: root, Cancel: func() error { return nil }}, nativeID: session.NativeID, stdin: requestWriter, done: make(chan struct{}), pending: map[string]chan map[string]json.RawMessage{}, events: []domain.NativeEvent{}}
	n := &Native{id: "codex", processes: map[string]*nativeProcess{session.ID: p}, capabilities: domain.UnknownNativeCapabilities()}
	metadata := map[string]any{"id": session.NativeID, "cwd": root, "path": filepath.Join(t.TempDir(), "public-unopened.jsonl"), "preview": "PRIVATE_METADATA_PREVIEW_SENTINEL", "turns": []any{}}
	requests := make(chan map[string]json.RawMessage, 8)
	readDone := make(chan struct{})
	go func() { p.read(responseReader); close(readDone) }()
	go func() {
		defer responseWriter.Close()
		scanner := bufio.NewScanner(requestReader)
		for scanner.Scan() {
			var request map[string]json.RawMessage
			if json.Unmarshal(scanner.Bytes(), &request) != nil {
				return
			}
			requests <- request
			response, _ := json.Marshal(map[string]any{"id": json.RawMessage(request["id"]), "result": map[string]any{"thread": metadata}})
			if _, err := responseWriter.Write(append(response, '\n')); err != nil {
				return
			}
		}
	}()
	t.Cleanup(func() {
		_ = requestWriter.Close()
		_ = requestReader.Close()
		_ = responseReader.Close()
		_ = responseWriter.Close()
		select {
		case <-p.done:
		default:
			close(p.done)
		}
		select {
		case <-readDone:
		case <-time.After(time.Second):
			t.Error("controlled metadata reader did not join")
		}
	})
	return n, session, requests, metadata
}

func TestRelayOwnedCodexMetadataExactOriginalIDNoTurnsAndNoArchiveRead(t *testing.T) {
	n, session, requests, metadata := relayMetadataRPCFixture(t)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	archive, err := n.OwnedRelayArchive(ctx, session)
	if err != nil {
		t.Fatal(err)
	}
	request := <-requests
	var method string
	_ = json.Unmarshal(request["method"], &method)
	var params struct {
		ThreadID     string `json:"threadId"`
		IncludeTurns bool   `json:"includeTurns"`
	}
	if json.Unmarshal(request["params"], &params) != nil || method != "thread/read" || params.ThreadID != session.NativeID || params.IncludeTurns {
		t.Fatal("metadata queried another session or hydrated turns")
	}
	if archive.SessionID != session.ID || archive.NativeID != session.NativeID || archive.Path != metadata["path"] || archive.Root != filepath.Dir(archive.Path) || archive.WorkspaceRoot != metadata["cwd"] || archive.Format != "codex" || archive.Finalized || archive.AppendOnly {
		t.Fatal("metadata source identity/framing/closure incorrect")
	}
	if _, err := os.Stat(archive.Path); !os.IsNotExist(err) {
		t.Fatal("metadata test unexpectedly created/read an archive fixture")
	}
	encoded, _ := json.Marshal(archive)
	if strings.Contains(string(encoded), "PRIVATE_METADATA_PREVIEW_SENTINEL") || string(encoded) != "{}" {
		t.Fatal("metadata/path fields escaped the internal source object")
	}
	if obs := n.processes[session.ID].observe(); len(obs.Events) != 0 || obs.StopConfirmed {
		t.Fatal("metadata-only read invented activity or exit")
	}
}

func TestRelayOwnedCodexMetadataMissingOrConflictingEvidenceRejected(t *testing.T) {
	for _, reason := range []string{"id", "cwd", "null_path", "relative_path", "unknown_format"} {
		t.Run(reason, func(t *testing.T) {
			n, session, requests, metadata := relayMetadataRPCFixture(t)
			switch reason {
			case "id":
				metadata["id"] = "different-thread"
			case "cwd":
				metadata["cwd"] = t.TempDir()
			case "null_path":
				metadata["path"] = nil
			case "relative_path":
				metadata["path"] = "public.jsonl"
			case "unknown_format":
				metadata["path"] = filepath.Join(t.TempDir(), "public.unknown")
			}
			ctx, cancel := context.WithTimeout(context.Background(), time.Second)
			defer cancel()
			archive, err := n.OwnedRelayArchive(ctx, session)
			if err == nil || archive.Path != "" || archive.Root != "" {
				t.Fatal("unknown/mismatched source selected")
			}
			if len(requests) != 1 {
				t.Fatal("metadata operation retried or searched other sessions")
			}
		})
	}
}

func TestRelayOwnedMetadataUnownedStoppedOrUnsupportedNeverQueries(t *testing.T) {
	for _, reason := range []string{"external", "wrong_native_id", "stopped", "unsupported"} {
		t.Run(reason, func(t *testing.T) {
			n, session, requests, _ := relayMetadataRPCFixture(t)
			switch reason {
			case "external":
				session.Ownership = "external_observed"
			case "wrong_native_id":
				session.NativeID = "different-thread"
			case "stopped":
				close(n.processes[session.ID].done)
			case "unsupported":
				n.id = "claude"
				session.CLI = "claude"
			}
			archive, err := n.OwnedRelayArchive(context.Background(), session)
			var failure *apierrors.ServiceError
			if !errors.As(err, &failure) || archive.Path != "" || len(requests) != 0 {
				t.Fatal("unowned/stopped/unsupported source read protocol or exposed path")
			}
		})
	}
}

func TestRelayNativeDriverReadsActualOwnedMetadataProvider(t *testing.T) {
	n, session, requests, metadata := relayMetadataRPCFixture(t)
	driver := NewRelayNativeDriver(&Registry{adapters: map[string]app.NativeAdapter{"codex": n}})
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	obs, err := driver.Read(ctx, app.RelayLaunch{Spec: domain.AgentSpec{Name: "codex"}, Session: session})
	if err != nil || obs.Archive == nil || obs.Archive.Path != metadata["path"] || len(obs.Events) != 0 || obs.ExitConfirmed {
		t.Fatalf("native driver did not consume metadata without activity/exit: %v", err)
	}
	if len(requests) != 1 {
		t.Fatal("driver resolved anything besides the exact owned metadata")
	}
}
