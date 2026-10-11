package attention_engine_test

import (
	"bytes"
	"context"
	"encoding/json"
	"math"
	"os"
	"path/filepath"
	"testing"

	"github.com/songconmaisaix31-design/Astrocyte/internal/adapters/events"
	"github.com/songconmaisaix31-design/Astrocyte/internal/adapters/sqlite"
	"github.com/songconmaisaix31-design/Astrocyte/internal/attention/app"
	"github.com/songconmaisaix31-design/Astrocyte/migrations"
)

func TestM1NativeFormatFixtureReplayAccounting(t *testing.T) {
	ctx := context.Background()
	project := t.TempDir()
	nativeRoot := t.TempDir()
	input := []app.StoredEvent{}
	for _, source := range []string{"claude_code", "codex"} {
		var header, tool []byte
		if source == "claude_code" {
			header, _ = json.Marshal(map[string]any{"type": "system", "sessionId": "fixture-claude", "cwd": project, "timestamp": "2026-10-10T00:00:00.123Z"})
			tool = []byte(`{ "type":"assistant", "timestamp":"2026-10-10T00:00:01.123Z", "message":{"content":[{"type":"tool_use","name":"Edit","input":{"file_path":"selected_test.go","new_string":"private fixture payload"}}]}}`)
		} else {
			header, _ = json.Marshal(map[string]any{"type": "session_meta", "payload": map[string]any{"id": "fixture-codex", "cwd": project, "timestamp": "2026-10-10T00:00:00.123Z"}})
			tool = []byte(`{ "type":"response_item", "timestamp":"2026-10-10T00:00:01.123Z", "payload":{"type":"function_call","name":"apply_patch","arguments":"private fixture payload"}}`)
		}
		path := filepath.Join(nativeRoot, source+".jsonl")
		if err := os.WriteFile(path, append(append(append([]byte{}, header...), '\n'), append(tool, '\n')...), 0600); err != nil {
			t.Fatal(err)
		}
		records, partial, err := events.ReplayNativeFile(ctx, source, nativeRoot, path, project, 128, 1791590400500)
		if err != nil || partial || len(records) != 2 || records[0].Event != nil || records[1].Event == nil || records[1].Event.Kind != "edit_file" || !bytes.Equal(records[1].RawBytes, tool) {
			t.Fatalf("contract_local %s native accounting: count=%d partial=%v err=%v", source, len(records), partial, err)
		}
		for _, v := range records {
			if v.Observation.Tokens != nil || v.Observation.Outcome != "unknown" || v.Observation.Depth != "unknown" || v.Event != nil && v.Event.Topic != "" {
				t.Fatal("native parser invented missing semantics")
			}
		}
		input = append(input, records...)
	}
	db, err := sqlite.Open(filepath.Join(t.TempDir(), "replay.sqlite"))
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	if err = db.RunMigrations(ctx, migrations.FS); err != nil {
		t.Fatal(err)
	}
	repo := sqlite.NewEventRepository(db)
	svc := app.NewEventService(repo, 1)
	first, err := svc.Record(ctx, input)
	if err != nil || len(first) != 4 {
		t.Fatal("native-format fixture receipt accounting failed")
	}
	replayed, err := svc.Record(ctx, input)
	if err != nil || len(replayed) != 4 {
		t.Fatal("native-format fixture replay failed")
	}
	for i, v := range replayed {
		if v.ID != first[i].ID || !bytes.Equal(v.RawBytes, input[i].RawBytes) {
			t.Fatal("native-format fixture stable identity lost")
		}
	}
	ss, err := repo.ListSignals(ctx, 0, math.MaxInt64)
	if err != nil || len(ss) != 2 {
		t.Fatal("native-format fixture signal accounting failed")
	}
	for _, s := range ss {
		if s.Weight != .6 || s.Strength != 1 || s.Topic != "" {
			t.Fatal("native-format fixture weight or unknown topic changed")
		}
	}
}
