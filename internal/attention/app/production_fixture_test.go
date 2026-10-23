package app_test

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"sync/atomic"
	"testing"
	"time"

	"github.com/songconmaisaix31-design/Astrocyte/internal/adapters/importers"
	"github.com/songconmaisaix31-design/Astrocyte/internal/adapters/objects"
	"github.com/songconmaisaix31-design/Astrocyte/internal/adapters/projectspacebridge"
	"github.com/songconmaisaix31-design/Astrocyte/internal/adapters/sqlite"
	"github.com/songconmaisaix31-design/Astrocyte/internal/apierrors"
	. "github.com/songconmaisaix31-design/Astrocyte/internal/attention/app"
	workspace "github.com/songconmaisaix31-design/Astrocyte/internal/workspace/app"
	"github.com/songconmaisaix31-design/Astrocyte/migrations"
)

var human = Principal{ID: "local-user", Kind: "human"}
var agent = Principal{ID: "agent-reader", Kind: "agent"}

type observedObjects struct {
	*objects.Store
	publishes     atomic.Int32
	beforePublish func()
}

func (o *observedObjects) Publish(ctx context.Context, content []byte) (string, error) {
	if o.beforePublish != nil {
		o.beforePublish()
	}
	o.publishes.Add(1)
	return o.Store.Publish(ctx, content)
}

// 只记录生产 Reader 调用，不提供替代来源响应。
type observedSource struct {
	*importers.Reader
	calls             atomic.Int32
	beforeRead        func()
	beforeReadContext func(context.Context)
	onError           func(error)
}

func (r *observedSource) ReadSource(ctx context.Context, command ImportMaterialCommand) (ImportedSource, error) {
	if r.beforeReadContext != nil {
		r.beforeReadContext(ctx)
	}
	if r.beforeRead != nil {
		r.beforeRead()
	}
	r.calls.Add(1)
	result, err := r.Reader.ReadSource(ctx, command)
	if err != nil && r.onError != nil {
		r.onError(err)
	}
	return result, err
}

type observedRepository struct {
	Repository
	active atomic.Int32
}

func (r *observedRepository) WithTx(ctx context.Context, fn func(AttentionTx) error) error {
	r.active.Add(1)
	defer r.active.Add(-1)
	return r.Repository.WithTx(ctx, fn)
}

func meta(key string, version int) CommandMeta {
	return CommandMeta{SchemaVersion: 1, RequestID: "request-" + key, IdempotencyKey: key, ExpectedVersion: version}
}

func evidenceRoot(t *testing.T) string {
	t.Helper()
	root := os.Getenv("ASTROCYTE_TEST_OUTPUT_ROOT")
	if root == "" || !filepath.IsAbs(root) {
		t.Fatal("ASTROCYTE_TEST_OUTPUT_ROOT must name an explicit absolute owned evidence directory")
	}
	if err := os.MkdirAll(root, 0700); err != nil {
		t.Fatal("create evidence directory")
	}
	return root
}

func projectSpaces(db *sqlite.DB) ProjectSpaces {
	return projectspacebridge.Attention{Spaces: workspace.NewProjectSpaceService(sqlite.NewProjectSpaceRepository(db), workspace.ProjectSpaceOptions{})}
}

func fixture(t *testing.T) (*Service, *sqlite.DB, *observedObjects) {
	t.Helper()
	dir, err := os.MkdirTemp(evidenceRoot(t), "attention-")
	if err != nil {
		t.Fatal("create owned test directory")
	}
	t.Logf("owned SQLite and object evidence directory=%s", dir)
	db, err := sqlite.Open(filepath.Join(dir, "state.sqlite"))
	if err != nil {
		t.Fatal("open owned SQLite")
	}
	t.Cleanup(func() {
		if err := db.Close(); err != nil {
			t.Error("close owned SQLite")
		}
	})
	if err = db.RunMigrations(context.Background(), migrations.FS); err != nil {
		t.Fatal("migrate owned SQLite")
	}
	store, err := objects.New(filepath.Join(dir, "objects"))
	if err != nil {
		t.Fatal("create owned objects")
	}
	now := time.Date(2026, 10, 9, 12, 0, 0, 0, time.UTC)
	o := &observedObjects{Store: store}
	s := NewAttentionService(sqlite.NewAttentionRepository(db), importers.NewReader(nil), o, ServiceOptions{Clock: func() time.Time { return now }, ProjectSpaces: projectSpaces(db)})
	return s, db, o
}

// 这些正文是人工提交的用例参数，Reader 实际处理 manual 导入，不声明外部读取成功。
func importCommand(key, text string) ImportMaterialCommand {
	return ImportMaterialCommand{CommandMeta: meta(key, 1), SourceLocator: "https://github.com/songconmaisaix31-design/Astrocyte", Kind: "text", ExportText: text, Adapter: "manual", Title: "Original source"}
}

func importFixture(t *testing.T, s *Service, key, text string) MaterialDetail {
	t.Helper()
	queued, err := s.ImportMaterial(context.Background(), human, importCommand(key, text))
	if err != nil {
		t.Fatal("manual import request failed")
	}
	if _, err = s.ProcessNextJob(context.Background()); err != nil {
		t.Fatal("manual import processing failed")
	}
	job, err := s.GetJob(context.Background(), human, queued.JobID)
	if err != nil || job.Status != "succeeded" || job.MaterialID == nil {
		t.Fatal("manual import did not complete")
	}
	row, err := s.GetMaterial(context.Background(), human, *job.MaterialID)
	if err != nil {
		t.Fatal("load imported material")
	}
	return row
}

func sourceRefs(row MaterialDetail) []SourceRef {
	return []SourceRef{{MaterialID: row.Material.ID, Revision: row.Material.CurrentRevision, Locator: row.Material.SourceLocator}}
}

func errorCode(t *testing.T, err error, want apierrors.Code) {
	t.Helper()
	var service *apierrors.ServiceError
	if !errors.As(err, &service) || service.Code != want {
		t.Fatalf("expected service error code %s", want)
	}
}

func countRows(t *testing.T, db *sqlite.DB, table string) int {
	t.Helper()
	var count int
	if err := db.Conn().QueryRow("SELECT COUNT(*) FROM " + table).Scan(&count); err != nil {
		t.Fatal("count persisted rows")
	}
	return count
}

func refuseOutbox(t *testing.T, db *sqlite.DB) func() {
	t.Helper()
	if _, err := db.Conn().Exec("CREATE TRIGGER attention_test_outbox_failure BEFORE INSERT ON attention_outbox BEGIN SELECT * FROM attention_test_unavailable; END"); err != nil {
		t.Fatal("restrict owned outbox")
	}
	return func() {
		if _, err := db.Conn().Exec("DROP TRIGGER attention_test_outbox_failure"); err != nil {
			t.Fatal("restore owned outbox")
		}
	}
}
