package attention_engine_test

import (
	"context"
	"database/sql"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/songconmaisaix31-design/Astrocyte/internal/adapters/events"
	_ "modernc.org/sqlite"
)

// Opt-in consumes only already persisted S1 project/session metadata, never
// native transcript bodies. A missing client remains a failed acceptance.
func TestM1CachedNativeMetadataLive(t *testing.T) {
	path, scope := os.Getenv("ASTROCYTE_M1_METADATA_DB"), os.Getenv("ASTROCYTE_M1_PROJECT_ROOT")
	if path == "" || scope == "" {
		t.Skip("selected S1 metadata database and approved project root required")
	}
	db, err := sql.Open("sqlite", "file:"+filepath.ToSlash(path)+"?mode=ro")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	records, partial, err := events.ReadCachedProjectMetadata(context.Background(), db, scope, 128)
	if err != nil {
		t.Fatal(err)
	}
	clients := map[string]int{}
	for _, raw := range records {
		var v struct {
			CLI     string `json:"cli"`
			Root    string `json:"root"`
			Session string `json:"session_id"`
			Source  string `json:"source"`
		}
		if err = json.Unmarshal(raw, &v); err != nil {
			t.Fatal(err)
		}
		if v.Root != scope || v.Session == "" || v.Source != "native_session_header" {
			t.Fatal("invalid selected metadata provenance")
		}
		clients[v.CLI]++
	}
	t.Logf("actual cached metadata: claude=%d codex=%d partial=%v; outcome/depth/tokens/topic unknown; no messages read", clients["claude"], clients["codex"], partial)
	if clients["claude"] == 0 || clients["codex"] == 0 {
		t.Fatal("both real client sources required")
	}
}
