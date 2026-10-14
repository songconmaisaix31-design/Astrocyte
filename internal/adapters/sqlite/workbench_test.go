package sqlite

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"runtime"
	"strings"
	"testing"
	"time"

	workbench "github.com/songconmaisaix31-design/Astrocyte/internal/workbench/app"
	"github.com/songconmaisaix31-design/Astrocyte/migrations"
)

func workbenchFixture(t *testing.T) *WorkbenchReader {
	t.Helper()
	root := t.TempDir()
	dir := filepath.Join(root, "snapshots", "projection-input")
	if err := os.MkdirAll(dir, 0700); err != nil {
		t.Fatal(err)
	}
	db, err := Open(filepath.Join(dir, "snapshot.sqlite"))
	if err != nil {
		t.Fatal(err)
	}
	if err = db.RunMigrations(context.Background(), migrations.FS); err != nil {
		db.Close()
		t.Fatal(err)
	}
	// 使用实际迁移后的 SQLite。以下行是明确的投影输入，不声称是原生会话或正式 Sleep 来源。
	for _, statement := range []string{
		`INSERT INTO attention_engine_nodes(id,topic,attention,momentum,confidence,last_active,updated_at,data) VALUES ('projection-node','topic',.7,.2,.8,1791820799999,1791820799999,'{}')`,
		`INSERT INTO attention_engine_sleep_runs(id,day,started_at,completed_at,data) VALUES ('projection-sleep','2026-10-12',1,1791820799999,'{}')`,
	} {
		if _, err = db.Conn().Exec(statement); err != nil {
			db.Close()
			t.Fatal(err)
		}
	}
	for i, id := range []string{"b", "a", "c"} {
		ms := int64(1791820800000)
		if i == 2 {
			ms--
		}
		payload := `{"tokens":null,"scope":"safe-scope","raw":{"input":"PRIVATE_CANARY"}}`
		if i == 0 {
			payload = `{"tokens":7,"scope":"safe-scope","raw":{"input":"PRIVATE_CANARY"}}`
		}
		if _, err = db.Conn().Exec(`INSERT INTO attention_engine_events(id,source,ts,kind,topic,data,raw_bytes,observation) VALUES (?,?,?,?,?,?,?,?)`, id, "manual", ms, "projection_input", "topic", payload, []byte("PRIVATE_CANARY"), "{}"); err != nil {
			db.Close()
			t.Fatal(err)
		}
	}
	if err = db.Close(); err != nil {
		t.Fatal(err)
	}
	if err = os.WriteFile(filepath.Join(root, "current.json"), []byte(`{"generation":"projection-input","snapshot_at":"2026-10-13T08:00:00Z"}`), 0600); err != nil {
		t.Fatal(err)
	}
	if err = os.WriteFile(filepath.Join(dir, "history.json"), []byte(`{"topic":[{"ts":1791820799999,"attention":0.7},{"ts":1791820799999,"attention":0.7}]}`), 0600); err != nil {
		t.Fatal(err)
	}
	reader := NewWorkbenchReader(root)
	reader.clock = func() time.Time { return time.Date(2026, 10, 13, 8, 0, 0, 0, time.UTC) }
	return reader
}

func TestWorkbenchProjectionPaginationAndShanghaiDays(t *testing.T) {
	s := workbenchFixture(t)
	q := workbench.Query{Start: "2026-10-13", End: "2026-10-13", Page: 1, PageSize: 1}
	v, err := s.Read(context.Background(), "events", q)
	if err != nil {
		t.Fatal(err)
	}
	if v.Events.Total != 2 || len(v.Events.Events) != 1 || v.Events.Events[0].ID != "a" || v.Events.Events[0].Tokens != nil {
		t.Fatal("pagination/date mismatch")
	}
	q.Page = 2
	v, err = s.Read(context.Background(), "events", q)
	if err != nil || v.Events.Events[0].ID != "b" || *v.Events.Events[0].Tokens != 7 {
		t.Fatal("tie order or tokens mismatch")
	}
	data, err := json.Marshal(v.Events.Events[0])
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(data), "PRIVATE_CANARY") || strings.Contains(string(data), "raw") {
		t.Fatal("private payload exposed")
	}
	var fields map[string]any
	if err = json.Unmarshal(data, &fields); err != nil {
		t.Fatal(err)
	}
	if len(fields) != 6 {
		t.Fatal("event projection is not six fields")
	}
	v, err = s.Read(context.Background(), "events", workbench.Query{Topic: "missing"})
	if err != nil || v.Events.Total != 0 || len(v.Events.Events) != 0 {
		t.Fatal("unknown filter not empty")
	}
	v, err = s.Read(context.Background(), "events", workbench.Query{Topic: "' OR 1=1 --"})
	if err != nil || v.Events.Total != 0 {
		t.Fatal("filter not bound")
	}
	v, err = s.Read(context.Background(), "topics/evidence", workbench.Query{Topic: "topic"})
	if err != nil || len(v.Evidence.Aggregates) != 2 || v.Evidence.Total != 3 {
		t.Fatal("aggregates lost full evidence")
	}
}

func TestWorkbenchTodayEmptyDayAndConfirmedHistory(t *testing.T) {
	s := workbenchFixture(t)
	s.clock = func() time.Time { return time.Date(2026, 10, 14, 8, 0, 0, 0, time.UTC) }
	v, err := s.Read(context.Background(), "today", workbench.Query{})
	if err != nil {
		t.Fatal(err)
	}
	if v.Today.Day != "2026-10-14" || len(v.Today.Activity) != 0 || v.Today.Sleep.Status != "not_produced" || v.Today.Focus[0].Delta != nil || len(v.Today.DailyCounts) != 7 {
		t.Fatal("empty day fabricated")
	}
	v, err = s.Read(context.Background(), "focus", workbench.Query{})
	if err != nil || len(v.Focus.Series[0].Points) != 1 || v.Focus.Series[0].Points[0].Attention != .7 {
		t.Fatal("history interpolated or not deduplicated")
	}
	s.clock = func() time.Time { return time.Date(2026, 10, 13, 8, 0, 0, 0, time.UTC) }
	v, err = s.Read(context.Background(), "today", workbench.Query{})
	if err != nil || v.Today.Sleep.Status != "completed" || len(v.Today.Activity) != 2 {
		t.Fatal("today/sleep mismatch")
	}
	v, err = s.Read(context.Background(), "agents", workbench.Query{})
	if err != nil || v.Agents.Groups[0].Today != 2 || v.Agents.Groups[0].Days7 != 3 || len(v.Agents.Groups[0].Recent) != 3 {
		t.Fatal("source groups mismatch")
	}
	v, err = s.Read(context.Background(), "filters", workbench.Query{})
	if err != nil || !reflect.DeepEqual(v.Filters.Sources, []string{"manual"}) {
		t.Fatal("filters mismatch")
	}
}

func TestWorkbenchReadOnlyAndPublicationReopen(t *testing.T) {
	s := workbenchFixture(t)
	db, _, _, err := s.open(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if _, err = db.Exec(`DELETE FROM attention_engine_events`); err == nil {
		t.Fatal("copy writable")
	}
	var queryOnly int
	if err = db.row(`PRAGMA query_only`).Scan(&queryOnly); err != nil || queryOnly != 1 {
		t.Fatal("query_only not enabled")
	}
	if err = db.Close(); err != nil {
		t.Fatal(err)
	}
	pointer := filepath.Join(s.root, "current.json")
	staged := pointer + ".next"
	if err = os.WriteFile(staged, []byte(`{"generation":"projection-input","snapshot_at":"2026-10-14T08:00:00Z"}`), 0600); err != nil {
		t.Fatal(err)
	}
	if err = os.Rename(staged, pointer); err != nil {
		t.Fatal(err)
	}
	v, err := s.Read(context.Background(), "health", workbench.Query{})
	if err != nil || v.Health.SnapshotAt != "2026-10-14T08:00:00Z" {
		t.Fatal("publication not reopened")
	}
}

func TestWorkbenchSanitizedFailureAndQueryValidation(t *testing.T) {
	s := workbenchFixture(t)
	for _, q := range []workbench.Query{{Page: -1}, {PageSize: 101}, {Start: "2026-13-01"}, {Start: "2026-10-14", End: "2026-10-13"}} {
		if _, err := s.Read(context.Background(), "events", q); !errors.Is(err, workbench.ErrInvalidQuery) {
			t.Fatal("invalid query accepted")
		}
	}
	if err := os.WriteFile(filepath.Join(s.root, "current.json"), []byte(`{"generation":"../PRIVATE_CANARY","snapshot_at":"x"}`), 0600); err != nil {
		t.Fatal(err)
	}
	if _, err := s.Read(context.Background(), "health", workbench.Query{}); err != workbench.ErrUnavailable || strings.Contains(err.Error(), "PRIVATE_CANARY") {
		t.Fatal("private error escaped")
	}
}

func TestWorkbenchConfiguredSourceAndHardlinkIdentityRejected(t *testing.T) {
	s := workbenchFixture(t)
	snapshot := filepath.Join(s.root, "snapshots", "projection-input", "snapshot.sqlite")
	for _, original := range []string{snapshot, filepath.Join(s.root, "source-alias.sqlite")} {
		if original != snapshot {
			if err := os.Link(snapshot, original); err != nil {
				t.Fatal(err)
			}
		}
		reader := NewWorkbenchReader(s.root, original)
		if _, err := reader.Read(context.Background(), "health", workbench.Query{}); !errors.Is(err, workbench.ErrUnavailable) {
			t.Fatal("same-file source/snapshot accepted")
		}
	}
	if len(NewWorkbenchReader(s.root).protectedSources) != 0 {
		t.Fatal("default reader inferred a personal source")
	}
	different := filepath.Join(t.TempDir(), "actual-source.sqlite")
	db, err := Open(different)
	if err != nil {
		t.Fatal(err)
	}
	if err = db.Close(); err != nil {
		t.Fatal(err)
	}
	if _, err = NewWorkbenchReader(s.root, different).Read(context.Background(), "health", workbench.Query{}); err != nil {
		t.Fatal("independent actual source rejected", err)
	}
}

func TestWorkbenchRejectsRedirectedGeneration(t *testing.T) {
	s := workbenchFixture(t)
	alias := filepath.Join(s.root, "snapshots", "redirected")
	target := filepath.Join(s.root, "snapshots", "projection-input")
	if runtime.GOOS == "windows" {
		if output, err := exec.Command("cmd", "/c", "mklink", "/J", alias, target).CombinedOutput(); err != nil {
			t.Fatalf("owned junction creation failed: %v bytes=%d", err, len(output))
		}
	} else if err := os.Symlink(target, alias); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(s.root, "current.json"), []byte(`{"generation":"redirected","snapshot_at":"2026-10-14T08:00:00Z"}`), 0600); err != nil {
		t.Fatal(err)
	}
	if _, err := s.Read(context.Background(), "health", workbench.Query{}); !errors.Is(err, workbench.ErrUnavailable) {
		t.Fatal("redirected generation accepted")
	}
}
