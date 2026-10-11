package events

import (
	"context"
	"database/sql"
	"encoding/json"
	"path/filepath"
	"testing"

	_ "modernc.org/sqlite"
)

func TestS1MetadataSelectedProjectAndReplayBounds(t *testing.T) {
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	if _, err = db.Exec(`CREATE TABLE local_project_discovery(id TEXT, data TEXT)`); err != nil {
		t.Fatal(err)
	}
	root, other := filepath.Join(t.TempDir(), "approved"), filepath.Join(t.TempDir(), "other")
	contributor := func(cli, session, project string) map[string]any {
		return map[string]any{"cli": cli, "source": "native_session_header", "root": project, "session_id": session, "created_at": "2026-10-10T00:00:00Z", "activity_at": nil}
	}
	raw, _ := json.Marshal(map[string]any{"projects": []any{
		map[string]any{"root": root, "contributors": []any{contributor("claude", "claude-a", root), contributor("codex", "codex-a", root)}},
		map[string]any{"root": other, "contributors": []any{contributor("codex", "not-authorized", other)}},
	}})
	if _, err = db.Exec(`INSERT INTO local_project_discovery VALUES('registered', ?)`, string(raw)); err != nil {
		t.Fatal(err)
	}
	records, partial, err := ReadCachedProjectMetadata(context.Background(), db, root, 2)
	if err != nil || partial || len(records) != 2 {
		t.Fatalf("selected metadata count=%d partial=%v err=%v", len(records), partial, err)
	}
	for _, record := range records {
		source, o, err := cachedMetadata(record)
		if err != nil || source == "" || o.cwd != root || o.kind != "unknown" || o.outcome != "unknown" || o.depth != "unknown" || o.tokens != nil || o.ts != 1791590400000 {
			t.Fatalf("metadata projection: %+v, %v", o, err)
		}
	}
	records, partial, err = ReadCachedProjectMetadata(context.Background(), db, root, 1)
	if err != nil || !partial || len(records) != 1 {
		t.Fatalf("bounded metadata count=%d partial=%v err=%v", len(records), partial, err)
	}
}
