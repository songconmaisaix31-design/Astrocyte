package sqlite

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"

	workbench "github.com/songconmaisaix31-design/Astrocyte/internal/workbench/app"
)

func workbenchFixture(t *testing.T) *WorkbenchReader {
	t.Helper()
	root := t.TempDir()
	dir := filepath.Join(root, "snapshots", "fixture")
	if e := os.MkdirAll(dir, 0700); e != nil {
		t.Fatal(e)
	}
	db, e := sql.Open("sqlite", filepath.Join(dir, "snapshot.sqlite"))
	if e != nil {
		t.Fatal(e)
	}
	defer db.Close()
	for _, stmt := range []string{
		`CREATE TABLE attention_engine_events(id TEXT PRIMARY KEY,source TEXT,ts INTEGER,kind TEXT,topic TEXT,data TEXT,raw_bytes BLOB,observation TEXT)`,
		`CREATE TABLE attention_engine_nodes(topic TEXT PRIMARY KEY,attention REAL,momentum REAL,confidence REAL,updated_at INTEGER)`,
		`CREATE TABLE attention_engine_sleep_runs(id TEXT,day TEXT,started_at INTEGER,completed_at INTEGER)`,
		`INSERT INTO attention_engine_nodes VALUES ('topic',.7,.2,.8,1791820799999)`,
		`INSERT INTO attention_engine_sleep_runs VALUES ('sleep','2026-10-12',1,1791820799999)`,
	} {
		if _, e = db.Exec(stmt); e != nil {
			t.Fatal(e)
		}
	}
	for i, id := range []string{"b", "a", "c"} {
		ms := int64(1791820800000)
		if i == 2 {
			ms--
		}
		payload := `{"tokens":null,"scope":"safe-scope","raw":{"raw_unique":"PRIVATE_CANARY"}}`
		if i == 0 {
			payload = `{"tokens":7,"scope":"safe-scope","raw":{"raw_unique":"PRIVATE_CANARY"}}`
		}
		if _, e = db.Exec(`INSERT INTO attention_engine_events VALUES (?,?,?,?,?,?,?,?)`, id, "codex", ms, "agent_task", "topic", payload, []byte("PRIVATE_CANARY"), "PRIVATE_CANARY"); e != nil {
			t.Fatal(e)
		}
	}
	if e = os.WriteFile(filepath.Join(root, "current.json"), []byte(`{"generation":"fixture","snapshot_at":"2026-10-13T08:00:00Z"}`), 0600); e != nil {
		t.Fatal(e)
	}
	if e = os.WriteFile(filepath.Join(dir, "history.json"), []byte(`{"topic":[{"ts":1791820799999,"attention":0.7},{"ts":1791820799999,"attention":0.7}]}`), 0600); e != nil {
		t.Fatal(e)
	}
	reader := NewWorkbenchReader(root)
	reader.clock = func() time.Time { return time.Date(2026, 10, 13, 8, 0, 0, 0, time.UTC) }
	return reader
}
func TestWorkbenchProjectionPaginationAndShanghaiDays(t *testing.T) {
	s := workbenchFixture(t)
	q := workbench.Query{Start: "2026-10-13", End: "2026-10-13", Page: 1, PageSize: 1}
	v, e := s.Read(context.Background(), "events", q)
	if e != nil {
		t.Fatal(e)
	}
	if v.Events.Total != 2 || len(v.Events.Events) != 1 || v.Events.Events[0].ID != "a" {
		t.Fatal("pagination/date mismatch")
	}
	q.Page = 2
	v, e = s.Read(context.Background(), "events", q)
	if e != nil || v.Events.Events[0].ID != "b" || *v.Events.Events[0].Tokens != 7 {
		t.Fatal("tie order or tokens mismatch")
	}
	b, e := json.Marshal(v.Events.Events[0])
	if e != nil {
		t.Fatal(e)
	}
	if strings.Contains(string(b), "PRIVATE_CANARY") || strings.Contains(string(b), "raw") {
		t.Fatal("private payload exposed")
	}
	var fields map[string]any
	json.Unmarshal(b, &fields)
	if len(fields) != 6 {
		t.Fatal("event projection is not six fields")
	}
	q = workbench.Query{Topic: "missing"}
	v, e = s.Read(context.Background(), "events", q)
	if e != nil || v.Events.Total != 0 || len(v.Events.Events) != 0 {
		t.Fatal("unknown filter not empty")
	}
	q = workbench.Query{Topic: "' OR 1=1 --"}
	v, e = s.Read(context.Background(), "events", q)
	if e != nil || v.Events.Total != 0 {
		t.Fatal("filter not bound")
	}
	v, e = s.Read(context.Background(), "topics/evidence", workbench.Query{Topic: "topic"})
	if e != nil || len(v.Evidence.Aggregates) != 2 || v.Evidence.Total != 3 {
		t.Fatal("aggregates lost full evidence")
	}
}
func TestWorkbenchTodayEmptyDayAndConfirmedHistory(t *testing.T) {
	s := workbenchFixture(t)
	s.clock = func() time.Time { return time.Date(2026, 10, 14, 8, 0, 0, 0, time.UTC) }
	v, e := s.Read(context.Background(), "today", workbench.Query{})
	if e != nil {
		t.Fatal(e)
	}
	if v.Today.Day != "2026-10-14" || len(v.Today.Activity) != 0 || v.Today.Sleep.Status != "not_produced" || v.Today.Focus[0].Delta != nil || len(v.Today.DailyCounts) != 7 {
		t.Fatal("empty day fabricated")
	}
	v, e = s.Read(context.Background(), "focus", workbench.Query{})
	if e != nil || len(v.Focus.Series[0].Points) != 1 || v.Focus.Series[0].Points[0].Attention != .7 {
		t.Fatal("history was interpolated or not deduplicated")
	}
	s.clock = func() time.Time { return time.Date(2026, 10, 13, 8, 0, 0, 0, time.UTC) }
	v, e = s.Read(context.Background(), "today", workbench.Query{})
	if e != nil || v.Today.Sleep.Status != "completed" || len(v.Today.Activity) != 2 {
		t.Fatal("today/sleep mismatch")
	}
	v, e = s.Read(context.Background(), "agents", workbench.Query{})
	if e != nil || v.Agents.Groups[0].Today != 2 || v.Agents.Groups[0].Days7 != 3 || len(v.Agents.Groups[0].Recent) != 3 {
		t.Fatal("source groups mismatch")
	}
	v, e = s.Read(context.Background(), "filters", workbench.Query{})
	if e != nil || !reflect.DeepEqual(v.Filters.Sources, []string{"codex"}) {
		t.Fatal("filters mismatch")
	}
}
func TestWorkbenchReadOnlyAndPublicationReopen(t *testing.T) {
	s := workbenchFixture(t)
	db, _, _, e := s.open(context.Background())
	if e != nil {
		t.Fatal(e)
	}
	if _, e = db.Exec(`DELETE FROM attention_engine_events`); e == nil {
		t.Fatal("copy writable")
	}
	var qo int
	if e = db.row(`PRAGMA query_only`).Scan(&qo); e != nil || qo != 1 {
		t.Fatal("query_only not enabled")
	}
	db.Close()
	if e = os.WriteFile(filepath.Join(s.root, "current.json"), []byte(`{"generation":"fixture","snapshot_at":"2026-10-14T08:00:00Z"}`), 0600); e != nil {
		t.Fatal(e)
	}
	v, e := s.Read(context.Background(), "health", workbench.Query{})
	if e != nil || v.Health.SnapshotAt != "2026-10-14T08:00:00Z" {
		t.Fatal("publication not reopened")
	}
}
func TestWorkbenchSanitizedFailureAndQueryValidation(t *testing.T) {
	s := workbenchFixture(t)
	for _, q := range []workbench.Query{{Page: -1}, {PageSize: 101}, {Start: "2026-13-01"}, {Start: "2026-10-14", End: "2026-10-13"}} {
		if _, e := s.Read(context.Background(), "events", q); !errors.Is(e, workbench.ErrInvalidQuery) {
			t.Fatal("invalid query accepted")
		}
	}
	if e := os.WriteFile(filepath.Join(s.root, "current.json"), []byte(`{"generation":"../PRIVATE_CANARY","snapshot_at":"x"}`), 0600); e != nil {
		t.Fatal(e)
	}
	if _, e := s.Read(context.Background(), "health", workbench.Query{}); e != workbench.ErrUnavailable || strings.Contains(e.Error(), "PRIVATE_CANARY") {
		t.Fatal("private error escaped")
	}
}
