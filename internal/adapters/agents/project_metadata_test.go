package agents

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/songconmaisaix31-design/Astrocyte/internal/workspace/domain"
)

func TestProjectMetadataHeaderOnlyScopeDuplicateAndUnknown(t *testing.T) {
	root, history, outside := t.TempDir(), t.TempDir(), t.TempDir()
	stamp := time.Now().Add(-time.Hour).UTC()
	write := func(name, cli, cwd, id string) {
		t.Helper()
		var record any = map[string]any{"type": "session", "id": id, "cwd": cwd, "timestamp": stamp.Format(time.RFC3339Nano)}
		if cli == "codex" {
			record = map[string]any{"type": "session_meta", "payload": map[string]any{"id": id, "cwd": cwd, "timestamp": stamp.Format(time.RFC3339Nano)}}
		}
		data, _ := json.Marshal(record)
		// Invalid huge private conversation below the header must never be read.
		data = append(data, []byte("\n"+strings.Repeat("PRIVATE_CONTENT_MUST_NOT_BE_READ", 10000))...)
		if err := os.WriteFile(filepath.Join(history, name), data, 0600); err != nil {
			t.Fatal(err)
		}
	}
	write("codex.jsonl", "codex", root, "codex-s")
	write("duplicate.jsonl", "codex", root, "codex-s")
	write("pi.jsonl", "pi", root, "pi-s")
	write("outside.jsonl", "codex", outside, "outside")
	source := &RegisteredProjects{metadataSources: func() []projectMetadataSource {
		return []projectMetadataSource{{"codex", history, 3}, {"pi", history, 1}}
	}}
	snapshot := domain.ProjectDiscoverySnapshot{Status: "complete", ObservedAt: &stamp, Projects: []domain.RegisteredProject{{Root: root, Name: "actual"}}}
	source.observeProjectMetadata(context.Background(), &snapshot)
	if len(snapshot.Projects[0].Contributors) != 2 {
		t.Fatalf("duplicate/outside header accepted: %+v", snapshot.Projects[0].Contributors)
	}
	encoded, _ := json.Marshal(snapshot)
	if strings.Contains(string(encoded), "PRIVATE_CONTENT") || strings.Contains(string(encoded), outside) {
		t.Fatal("private conversation/outside root retained")
	}
	if len(snapshot.Sources) != len(definitions) || snapshot.Status != "partial" {
		t.Fatalf("all-clients unknowns missing: %+v", snapshot.Sources)
	}
	for _, c := range snapshot.Projects[0].Contributors {
		if c.CreatedAt == nil || !c.CreatedAt.Equal(stamp) || c.ActivityAt != nil {
			t.Fatal("header creation timestamp lost or inferred as recent activity")
		}
	}
}

func TestProjectMetadataRejectOversizeHeadersAndCancelledScan(t *testing.T) {
	history := t.TempDir()
	if err := os.WriteFile(filepath.Join(history, "large.jsonl"), []byte(strings.Repeat("x", 64*1024+1)), 0600); err != nil {
		t.Fatal(err)
	}
	if _, _, _, err := readProjectHeader(filepath.Join(history, "large.jsonl"), "codex"); err == nil {
		t.Fatal("over-bound header accepted")
	}
	source := &RegisteredProjects{metadataSources: func() []projectMetadataSource { return []projectMetadataSource{{"codex", history, 3}} }}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	snapshot := domain.ProjectDiscoverySnapshot{Status: "complete"}
	source.observeProjectMetadata(ctx, &snapshot)
	if snapshot.Status != "partial" || snapshot.Sources[0].EntriesExamined != 0 {
		t.Fatal("cancelled metadata source scanned")
	}
}
