package agents

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/songconmaisaix31-design/Astrocyte/internal/workspace/domain"
)

func TestNativeMetadataIdentifiesStandaloneProjectWithoutReadGrant(t *testing.T) {
	root, history, nonproject := t.TempDir(), t.TempDir(), t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "package.json"), []byte("PRIVATE SOURCE NOT PARSED"), 0600); err != nil {
		t.Fatal(err)
	}
	stamp := time.Now().UTC()
	for i, cwd := range []string{root, root, nonproject} {
		data, _ := json.Marshal(map[string]any{"type": "session_meta", "payload": map[string]any{"id": "session-root", "cwd": cwd, "timestamp": stamp.Format(time.RFC3339Nano)}})
		if err := os.WriteFile(filepath.Join(history, []string{"first.jsonl", "duplicate.jsonl", "not-project.jsonl"}[i]), append(data, '\n'), 0600); err != nil {
			t.Fatal(err)
		}
	}
	s := &RegisteredProjects{metadataSources: func() []projectMetadataSource { return []projectMetadataSource{{"codex", history, 3}} }}
	snapshot := domain.ProjectDiscoverySnapshot{Status: "complete", ObservedAt: &stamp, Projects: []domain.RegisteredProject{}}
	s.observeProjectMetadata(context.Background(), &snapshot)
	if len(snapshot.Projects) != 1 || snapshot.Projects[0].Root != root || snapshot.Projects[0].Source != "native_project_metadata" || len(snapshot.Projects[0].Contributors) != 1 || snapshot.Projects[0].RepoID != "" {
		t.Fatalf("metadata root discovery %+v", snapshot)
	}
	if _, err := metadataProjectRoot(nonproject); err == nil {
		t.Fatal("nonproject cwd accepted")
	}
	if home, err := os.UserHomeDir(); err == nil {
		if _, err := metadataProjectRoot(home); err == nil {
			t.Fatal("home root accepted")
		}
	}
	secret := filepath.Join(root, ".codex")
	if err := os.Mkdir(secret, 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(secret, "package.json"), []byte("{}"), 0600); err != nil {
		t.Fatal(err)
	}
	if _, err := metadataProjectRoot(secret); err == nil {
		t.Fatal("native config root accepted")
	}
}

func TestNativeMetadataSessionSampleDoesNotHideLaterProjectRoots(t *testing.T) {
	history, first, later := t.TempDir(), t.TempDir(), t.TempDir()
	for _, root := range []string{first, later} {
		if err := os.WriteFile(filepath.Join(root, "go.mod"), []byte("not decoded"), 0600); err != nil {
			t.Fatal(err)
		}
	}
	for i := 0; i < 12; i++ {
		cwd := first
		if i == 11 {
			cwd = later
		}
		data, _ := json.Marshal(map[string]any{"type": "session_meta", "payload": map[string]any{"id": fmt.Sprintf("sample-%d", i), "cwd": cwd}})
		if err := os.WriteFile(filepath.Join(history, fmt.Sprintf("%02d.jsonl", i)), append(data, '\n'), 0600); err != nil {
			t.Fatal(err)
		}
	}
	s := &RegisteredProjects{metadataSources: func() []projectMetadataSource { return []projectMetadataSource{{"codex", history, 3}} }}
	snapshot := domain.ProjectDiscoverySnapshot{Status: "complete"}
	s.observeProjectMetadata(context.Background(), &snapshot)
	if len(snapshot.Projects) != 2 || snapshot.Sources[0].MatchedHeaders != 12 || snapshot.Sources[0].HeadersExamined != 12 || snapshot.Sources[0].MatchedRoots != 2 || snapshot.Sources[0].RetainedAssociations != 5 || snapshot.Sources[0].Status != "partial" {
		t.Fatalf("session sampling hid roots %+v", snapshot)
	}
	for _, p := range snapshot.Projects {
		if len(p.Contributors) > 4 {
			t.Fatal("perroot samples exceed bound")
		}
	}
}
