package s1_test

import (
	"database/sql"
	"testing"

	"github.com/songconmaisaix31-design/Astrocyte/migrations"
	_ "modernc.org/sqlite"
)

// Exercise a populated legacy database with foreign keys enabled. Source IDs,
// saved JSON, item revisions and selections must survive the identity upgrade.
func TestTrackingIdentityMigrationPreservesLegacyItems(t *testing.T) {
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	db.SetMaxOpenConns(1)
	exec := func(query string, args ...any) {
		t.Helper()
		if _, err := db.Exec(query, args...); err != nil {
			t.Fatal(err)
		}
	}
	exec("PRAGMA foreign_keys=ON")
	exec("CREATE TABLE foundation_schema(context TEXT PRIMARY KEY, version INTEGER); INSERT INTO foundation_schema VALUES('attention',4); CREATE TABLE _migrations(id TEXT PRIMARY KEY)")
	for _, name := range []string{"006_attention_tracking.sql"} {
		body, err := migrations.FS.ReadFile(name)
		if err != nil {
			t.Fatal(err)
		}
		exec(string(body))
	}
	legacy := `{"id":"legacy","owner_id":"owner-one","version":3}`
	item := `{"source_id":"legacy","external_id":"video","revision":2,"selected":true,"import_job_id":"old-job"}`
	exec("INSERT INTO attention_tracking_sources VALUES('legacy',3,'douyin','favorites','folder',?)", legacy)
	exec("INSERT INTO attention_source_items VALUES('legacy','video',2,?)", item)
	body, err := migrations.FS.ReadFile("010_tracking_source_identity.sql")
	if err != nil {
		t.Fatal(err)
	}
	tx, err := db.Begin()
	if err != nil {
		t.Fatal(err)
	}
	if _, err := tx.Exec(string(body)); err != nil {
		tx.Rollback()
		t.Fatal(err)
	}
	if err := tx.Commit(); err != nil {
		t.Fatal(err)
	}
	var saved, owner, mode string
	var version int
	if err := db.QueryRow("SELECT data,version,owner_id,access_mode FROM attention_tracking_sources WHERE id='legacy'").Scan(&saved, &version, &owner, &mode); err != nil {
		t.Fatal(err)
	}
	if saved != legacy || version != 3 || owner != "owner-one" || mode != "public" {
		t.Fatalf("legacy source changed: %s %d %s %s", saved, version, owner, mode)
	}
	if err := db.QueryRow("SELECT data FROM attention_source_items WHERE source_id='legacy'").Scan(&saved); err != nil || saved != item {
		t.Fatalf("selected item changed: %s %v", saved, err)
	}
	exec("INSERT INTO attention_tracking_sources VALUES('other-owner',1,'douyin','favorites','folder','owner-two','browser_selected','{}')")
	exec("INSERT INTO attention_tracking_sources VALUES('other-mode',1,'douyin','favorites','folder','owner-one','browser_selected','{}')")
	if _, err := db.Exec("INSERT INTO attention_tracking_sources VALUES('duplicate',1,'douyin','favorites','folder','owner-one','public','{}')"); err == nil {
		t.Fatal("duplicate complete source identity accepted")
	}
	if _, err := db.Exec("INSERT INTO attention_source_items VALUES('missing','video',1,'{}')"); err == nil {
		t.Fatal("source item foreign key lost")
	}
	rows, err := db.Query("PRAGMA foreign_key_check")
	if err != nil {
		t.Fatal(err)
	}
	defer rows.Close()
	if rows.Next() || rows.Err() != nil {
		t.Fatal("migration left broken foreign keys")
	}
}
