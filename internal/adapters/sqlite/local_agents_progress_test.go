package sqlite

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"net/url"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/google/uuid"
	"github.com/songconmaisaix31-design/Astrocyte/internal/adapters/agents"
	"github.com/songconmaisaix31-design/Astrocyte/internal/apierrors"
	workspace "github.com/songconmaisaix31-design/Astrocyte/internal/workspace/app"
	"github.com/songconmaisaix31-design/Astrocyte/internal/workspace/domain"
)

// 只读此前公开 README 验证的准确记录；原生身份和推断收据保持原始来源。
func readPublicProgress(t *testing.T) (domain.ProjectProgress, string, int) {
	t.Helper()
	publication := os.Getenv("ASTROCYTE_TEST_PUBLIC_PROGRESS_DB")
	if !filepath.IsAbs(publication) {
		t.Fatal("select the absolute SQLite path of the actual public README progress result")
	}
	locator := &url.URL{Scheme: "file", Path: filepath.ToSlash(publication)}
	if filepath.VolumeName(publication) != "" {
		locator.Path = "/" + locator.Path
	}
	query := locator.Query()
	query.Set("mode", "ro")
	locator.RawQuery = query.Encode()
	conn, err := sql.Open("sqlite", locator.String())
	if err != nil {
		t.Fatal(err)
	}
	conn.SetMaxOpenConns(1)
	t.Cleanup(func() {
		if err := conn.Close(); err != nil {
			t.Error(err)
		}
	})
	source := &DB{conn: conn, dbPath: publication}
	const projectID = "25b3f37a-a96c-4082-9a59-6ed6707fbf70"
	record, err := source.LoadProjectProgress(context.Background(), projectID)
	if err != nil {
		t.Fatal(err)
	}
	if record.ProjectID != projectID || record.Source != "agent_inferred" || record.NativeID != "01a13a73-1de2-7761-aaa3-5cbd8fb433cc" || record.Status == "" || len(record.Evidence) == 0 || len(record.Operations) == 0 {
		t.Fatal("actual public progress provenance is unavailable")
	}
	for _, operation := range record.Operations {
		if operation.Action != "infer" || operation.Status != "accepted" || operation.Result == nil || operation.Result.NativeID != record.NativeID || len(operation.Result.Evidence) == 0 {
			t.Fatal("actual accepted public inference receipt is unavailable")
		}
	}
	var projectJSON string
	var version int
	if err := conn.QueryRow("SELECT version,data FROM local_agent_projects WHERE id=?", projectID).Scan(&version, &projectJSON); err != nil {
		t.Fatal(err)
	}
	var project domain.LocalProject
	if err := json.Unmarshal([]byte(projectJSON), &project); err != nil || project.ID != projectID || project.Settings.Revision != version {
		t.Fatal("actual registered source project identity differs", err)
	}
	return record, projectJSON, version
}

func TestProgressStorageRoundTripAndCAS(t *testing.T) {
	ctx := context.Background()
	record, projectJSON, projectVersion := readPublicProgress(t)
	db := openAttentionDB(t, filepath.Join(t.TempDir(), "state.sqlite"))
	if got, err := db.LoadProjectProgress(ctx, uuid.NewString()); err != nil || got.Status != "" || got.Revision != 0 {
		t.Fatalf("absent project should be zero value, got %+v %v", got, err)
	}
	// 自有副本保留原项目元数据；新的进度投影从本库 CAS revision 1 开始。
	if _, err := db.conn.ExecContext(ctx, "INSERT INTO local_agent_projects(id,version,data) VALUES(?,?,?)", record.ProjectID, projectVersion, projectJSON); err != nil {
		t.Fatal(err)
	}
	record.Revision = 1
	if err := db.SaveProjectProgress(ctx, record, 0); err != nil {
		t.Fatal(err)
	}
	loaded, err := db.LoadProjectProgress(ctx, record.ProjectID)
	if err != nil || !reflect.DeepEqual(loaded, record) {
		t.Fatalf("actual public observation round trip changed: %v", err)
	}

	if err := db.SaveProjectProgress(ctx, record, 0); err == nil {
		t.Fatal("stale insert accepted")
	}
	percent := 50
	service := workspace.NewProjectProgressService(db, db, agents.ProjectFiles{}, nil, nil)
	updated, err := service.SetProjectProgress(ctx, domain.Caller{Kind: "human", ID: "local-human"}, record.ProjectID, domain.ProgressInput{Status: "review", Percent: &percent})
	if err != nil || updated.Revision != 2 || updated.Source != "human" {
		t.Fatal("actual human progress update failed", err)
	}
	loaded, err = db.LoadProjectProgress(ctx, record.ProjectID)
	if err != nil || loaded.Status != "review" || loaded.Revision != 2 || loaded.Source != "human" || loaded.Percent == nil || *loaded.Percent != percent || loaded.NativeID != "" || !reflect.DeepEqual(loaded.Operations, record.Operations) {
		t.Fatal("actual human update lost its source or immutable original inference receipt", err)
	}
	projects, err := db.ListProjects(ctx)
	if err != nil || len(projects) != 1 || projects[0].ID != record.ProjectID {
		t.Fatal("progress storage created or changed the actual source project")
	}
	_, err = db.LoadGrant(ctx, record.ProjectID, "ungranted-input")
	var missing *apierrors.ServiceError
	if !errors.As(err, &missing) || missing.Code != apierrors.NotFound {
		t.Fatal("progress storage created a grant or lost the actual missing result", err)
	}
	for _, table := range []string{"local_agent_sessions", "local_agent_grants"} {
		var count int
		if err := db.conn.QueryRow("SELECT COUNT(*) FROM " + table).Scan(&count); err != nil || count != 0 {
			t.Fatal("progress persistence created native work or permission", table, count, err)
		}
	}
}
