package attention_engine_test

import (
	"bytes"
	"context"
	"io/fs"
	"path/filepath"
	"testing"
	"testing/fstest"

	"github.com/songconmaisaix31-design/Astrocyte/internal/adapters/sqlite"
	"github.com/songconmaisaix31-design/Astrocyte/migrations"
)

func TestM0PersistenceUpgradeRestartAndConstraints(t *testing.T) {
	ctx := context.Background()
	path := filepath.Join(t.TempDir(), "state.sqlite")
	db, err := sqlite.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { db.Close() }()
	before := fstest.MapFS{}
	names, err := fs.Glob(migrations.FS, "*.sql")
	if err != nil {
		t.Fatal(err)
	}
	for _, name := range names {
		if name >= "015" {
			continue
		}
		b, err := migrations.FS.ReadFile(name)
		if err != nil {
			t.Fatal(err)
		}
		before[name] = &fstest.MapFile{Data: b}
	}
	if err = db.RunMigrations(ctx, before); err != nil {
		t.Fatal(err)
	}
	if _, err = db.Conn().ExecContext(ctx, `INSERT INTO attention_engine_events(id,source,ts,kind,data) VALUES('legacy','manual',42,'collect','{}');
INSERT INTO attention_engine_nodes(id,topic,attention,momentum,confidence,last_active,data) VALUES('node','t',.1,0,.2,42,'{}');`); err != nil {
		t.Fatal(err)
	}
	if err = db.RunMigrations(ctx, migrations.FS); err != nil {
		t.Fatal(err)
	}
	var updated, active int64
	if err = db.Conn().QueryRowContext(ctx, `SELECT updated_at,last_active FROM attention_engine_nodes WHERE id='node'`).Scan(&updated, &active); err != nil || updated != 42 || active != 42 {
		t.Fatalf("upgrade timestamp: %d %d %v", updated, active, err)
	}
	var missingRaw bool
	if err = db.Conn().QueryRowContext(ctx, `SELECT raw_bytes IS NULL FROM attention_engine_events WHERE id='legacy'`).Scan(&missingRaw); err != nil || !missingRaw {
		t.Fatalf("legacy original unknown: %v %v", missingRaw, err)
	}
	raw := []byte(" \n{\"unknown\": 1, \"x\": \"\\u0061\"}\r\n")
	_, err = db.Conn().ExecContext(ctx, `INSERT INTO attention_engine_events(id,source,ts,kind,data,origin_key,raw_bytes,observation,mapping_status) VALUES('receipt','codex',100,'','{}','codex:session:record',?,'{"tokens":null,"outcome":"unknown"}','pending')`, raw)
	if err != nil {
		t.Fatal(err)
	}
	reject := func(query string, args ...any) {
		t.Helper()
		if _, err := db.Conn().ExecContext(ctx, query, args...); err == nil {
			t.Fatalf("constraint accepted: %s", query)
		}
	}
	reject(`INSERT INTO attention_engine_events(id,source,ts,kind,data,origin_key) VALUES('duplicate','codex',100,'','{}','codex:session:record')`)
	reject(`INSERT INTO attention_engine_events(id,source,ts,kind,data,raw_bytes) VALUES('text','codex',100,'','{}','text')`)
	reject(`INSERT INTO attention_engine_signals(event_id,topic,class,weight,ts,data) VALUES('receipt','t','implicit_regular',.6,100,'{}')`)
	_, err = db.Conn().ExecContext(ctx, `INSERT INTO attention_engine_signals(event_id,topic,class,weight,ts,data) VALUES('legacy','t','implicit_light',.3,42,'{}');
INSERT INTO attention_engine_knowledge(id,topic,status,data,owner_project) VALUES('k','t','inbox','{}','project-a');
INSERT INTO attention_engine_knowledge_versions VALUES('k',1,'{"fact":1}',100);
INSERT INTO attention_engine_knowledge_versions VALUES('k',2,'{"fact":2}',200);
INSERT INTO attention_engine_knowledge_shares VALUES('k','project-a','project-b','local-user','human',200,'share-1');
INSERT INTO attention_engine_knowledge_references VALUES('k','project-b',2,2,'local-user','human',200,'reference-1');
INSERT INTO attention_engine_experiences(id,pattern,verdict,status,data,owner_project,origin_event_id) VALUES('e','p','avoid','quarantine','{}','project-a','legacy');
INSERT INTO attention_engine_sleep_runs VALUES('sleep','2026-10-11',100,200,'{}');
INSERT INTO attention_engine_sleep_metrics VALUES('sleep',2,1,100,50,.5,.5);
INSERT INTO attention_engine_sleep_candidates VALUES('sleep','candidate','resurface','{}');
INSERT INTO attention_engine_conflicts(id,topic,kind,status,data) VALUES('conflict','t','knowledge_entry','pending','{}');
INSERT INTO attention_engine_sleep_conflicts VALUES('sleep','conflict');
UPDATE attention_engine_nodes SET updated_at=200 WHERE id='node';`)
	if err != nil {
		t.Fatal(err)
	}
	reject(`INSERT INTO attention_engine_signals(event_id,topic,class,weight,ts,data) VALUES('legacy','t','implicit_light',.3,42,'{}')`)
	reject(`UPDATE attention_engine_knowledge_versions SET content='{}' WHERE entry_id='k'`)
	reject(`DELETE FROM attention_engine_knowledge_versions WHERE entry_id='k'`)
	reject(`UPDATE attention_engine_knowledge_references SET project_id='project-c' WHERE entry_id='k'`)
	reject(`DELETE FROM attention_engine_knowledge_references WHERE entry_id='k'`)
	reject(`INSERT INTO attention_engine_knowledge_references VALUES('k','project-b',2,99,'local-user','human',200,'missing-version')`)
	reject(`INSERT INTO attention_engine_knowledge_shares VALUES('k','project-a','project-c','agent','agent',200,'share-2')`)
	reject(`INSERT INTO attention_engine_experiences(id,pattern,verdict,status,data,owner_project,origin_event_id) VALUES('e2','p','avoid','quarantine','{}','project-a','legacy')`)
	reject(`INSERT INTO attention_engine_sleep_runs VALUES('sleep2','2026-10-11',100,200,'{}')`)
	if err = db.Close(); err != nil {
		t.Fatal(err)
	}
	db, err = sqlite.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	if err = db.RunMigrations(ctx, migrations.FS); err != nil {
		t.Fatal(err)
	}
	var got []byte
	var observation string
	if err = db.Conn().QueryRowContext(ctx, `SELECT raw_bytes,observation FROM attention_engine_events WHERE id='receipt'`).Scan(&got, &observation); err != nil || !bytes.Equal(got, raw) || observation != "{\"tokens\":null,\"outcome\":\"unknown\"}" {
		t.Fatalf("receipt restart: %q %s %v", got, observation, err)
	}
	if err = db.Conn().QueryRowContext(ctx, `SELECT updated_at,last_active FROM attention_engine_nodes WHERE id='node'`).Scan(&updated, &active); err != nil || updated != 200 || active != 42 {
		t.Fatalf("independent timestamp: %d %d %v", updated, active, err)
	}
	var versions, shares, metrics, candidates, conflicts int
	for _, q := range []struct {
		query string
		out   *int
	}{
		{`SELECT count(*) FROM attention_engine_knowledge_versions`, &versions},
		{`SELECT count(*) FROM attention_engine_knowledge_shares WHERE owner_project='project-a' AND target_project='project-b' AND actor_kind='human'`, &shares},
		{`SELECT count(*) FROM attention_engine_sleep_metrics`, &metrics},
		{`SELECT count(*) FROM attention_engine_sleep_candidates`, &candidates},
		{`SELECT count(*) FROM attention_engine_sleep_conflicts`, &conflicts},
	} {
		if err = db.Conn().QueryRowContext(ctx, q.query).Scan(q.out); err != nil {
			t.Fatal(err)
		}
	}
	if versions != 2 || shares != 1 || metrics != 1 || candidates != 1 || conflicts != 1 {
		t.Fatalf("restart rows: %d %d %d %d %d", versions, shares, metrics, candidates, conflicts)
	}
	var project string
	var knowledgeVersion, crystallizedVersion int
	if err = db.Conn().QueryRowContext(ctx, `SELECT project_id,knowledge_version,crystallized_version FROM attention_engine_knowledge_references WHERE entry_id='k' AND request_id='reference-1'`).Scan(&project, &knowledgeVersion, &crystallizedVersion); err != nil || project != "project-b" || knowledgeVersion != 2 || crystallizedVersion != 2 {
		t.Fatalf("reference restart: %s %d %d %v", project, knowledgeVersion, crystallizedVersion, err)
	}
}
