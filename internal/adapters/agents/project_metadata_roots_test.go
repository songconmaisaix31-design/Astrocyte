package agents

import (
	"context"
	"encoding/json"
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
