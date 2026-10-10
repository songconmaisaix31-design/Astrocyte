package agents

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestClaudeBoundedInitialMetadataIgnoresContentAndLimitsRecords(t *testing.T) {
	initial := `{"type":"file-history-snapshot","snapshot":{"private":"discard"}}` + "\n"
	root := t.TempDir()
	data, _ := json.Marshal(map[string]any{"type": "user", "cwd": root, "sessionId": "actual-session", "timestamp": "2026-01-01T00:00:00Z", "message": map[string]any{"content": "PRIVATE words with fake cwd /outside"}})
	header := string(data) + "\n"
	id, cwd, stamp, err := readClaudeProjectMetadata(strings.NewReader(initial + header + strings.Repeat("PRIVATE_TRANSCRIPT", 10000)))
	if err != nil || id != "actual-session" || cwd != root || stamp == nil {
		t.Fatalf("metadata-only extraction %s %s %v", id, cwd, err)
	}
	if _, _, _, err := readClaudeProjectMetadata(strings.NewReader(strings.Repeat(initial, 16) + header)); err == nil {
		t.Fatal("read beyond sixteen metadata records")
	}
	if _, _, _, err := readClaudeProjectMetadata(strings.NewReader(strings.Repeat("x", 64*1024) + header)); err == nil {
		t.Fatal("read over64KiB metadata budget")
	}
	if _, _, _, err := readClaudeProjectMetadata(strings.NewReader(`{"type":"user","message":{"cwd":"/outside","sessionId":"fake"}}`)); err == nil {
		t.Fatal("used conversation fields as project metadata")
	}
}
