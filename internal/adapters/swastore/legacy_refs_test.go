package swastore_test

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/google/uuid"
	"github.com/songconmaisaix31-design/Astrocyte/internal/adapters/objects"
	"github.com/songconmaisaix31-design/Astrocyte/internal/adapters/sqlite"
	"github.com/songconmaisaix31-design/Astrocyte/internal/adapters/swastore"
	attention "github.com/songconmaisaix31-design/Astrocyte/internal/attention/app"
	workspace "github.com/songconmaisaix31-design/Astrocyte/internal/workspace/app"
	"github.com/songconmaisaix31-design/Astrocyte/internal/workspace/domain"
	legacy "github.com/songconmaisaix31-design/Astrocyte/migrations"
	migrations "github.com/songconmaisaix31-design/Astrocyte/migrations/swa"
)

func TestLegacyMaterialReferenceAndReceiptSurviveStandaloneMigration(t *testing.T) {
	ctx := context.Background()
	root := t.TempDir()
	path := filepath.Join(root, "state.sqlite")
	db, err := sqlite.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	if err := db.RunMigrations(ctx, legacy.FS); err != nil {
		t.Fatal(err)
	}
	originals, err := objects.New(filepath.Join(root, "objects"))
	if err != nil {
		t.Fatal(err)
	}
	source, err := filepath.Abs("../../../README.md")
	if err != nil {
		t.Fatal(err)
	}
	text, err := os.ReadFile(source)
	if err != nil {
		t.Fatal(err)
	}
	digest, err := originals.Publish(ctx, text)
	if err != nil {
		t.Fatal(err)
	}
	materialID := uuid.NewString()
	material := attention.MaterialDetail{SchemaVersion: 1, Material: attention.Material{ID: materialID, Version: 1, CurrentRevision: 1, Kind: "paper", Lifecycle: "active", Title: "Public repository documentation"}, Revisions: []attention.MaterialRevision{{MaterialID: materialID, Revision: 1, SourceKey: source, SourceLocator: source, ContentDigest: digest, ObjectRef: digest}}, Uses: []attention.Usage{}, Distillations: []attention.Distillation{}}
	repo := sqlite.NewAttentionRepository(db)
	if err := repo.WithTx(ctx, func(tx attention.AttentionTx) error { return tx.SaveMaterial(material, 0) }); err != nil {
		t.Fatal(err)
	}
	id := uuid.NewString()
	caller := domain.Caller{ID: "local-human", Kind: "human"}
	spaces := workspace.NewProjectSpaceService(sqlite.NewProjectSpaceRepository(db), workspace.ProjectSpaceOptions{})
	created, err := spaces.CreateProjectSpace(ctx, caller, workspace.ProjectSpaceCommand{WorkspaceCommandMeta: workspace.WorkspaceCommandMeta{SchemaVersion: 1, RequestID: id, IdempotencyKey: id, ExpectedVersion: 1}, Title: "Preserved reference"})
	if err != nil {
		t.Fatal(err)
	}
	id = uuid.NewString()
	command := attention.ReferenceMaterialCommand{CommandMeta: attention.CommandMeta{SchemaVersion: 1, RequestID: id, IdempotencyKey: id, ExpectedVersion: 1}, MaterialID: materialID, Revision: 1}
	service := attention.NewAttentionService(repo, nil, nil, attention.ServiceOptions{})
	referenced, err := service.ReferenceMaterial(ctx, attention.Principal{ID: caller.ID, Kind: caller.Kind}, created.Space.ID, command)
	if err != nil {
		t.Fatal(err)
	}
	if len(referenced.Space.MaterialRefs) != 1 {
		t.Fatal("actual reference command did not persist reference")
	}
	var before []byte
	if err := db.Conn().QueryRow("SELECT data FROM attention_project_spaces WHERE id=?", created.Space.ID).Scan(&before); err != nil {
		t.Fatal(err)
	}
	var receiptsBefore int
	if err := db.Conn().QueryRow("SELECT COUNT(*) FROM attention_receipts").Scan(&receiptsBefore); err != nil {
		t.Fatal(err)
	}
	if err := db.Close(); err != nil {
		t.Fatal(err)
	}
	standalone, err := swastore.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer standalone.Close()
	if err := standalone.RunMigrations(ctx, migrations.FS); err != nil {
		t.Fatal(err)
	}
	var after []byte
	if err := standalone.Conn().QueryRow("SELECT data FROM attention_project_spaces WHERE id=?", created.Space.ID).Scan(&after); err != nil || !bytes.Equal(before, after) {
		t.Fatal("migration rewrote original reference or locator", err)
	}
	read, err := workspace.NewProjectSpaceService(swastore.NewProjectSpaceRepository(standalone), workspace.ProjectSpaceOptions{}).GetProjectSpace(ctx, caller, created.Space.ID)
	if err != nil || read.Space.Version != referenced.Space.Version || len(read.Space.MaterialRefs) != 1 || read.Space.MaterialRefs[0].MaterialID != materialID || read.Space.MaterialRefs[0].Locator != source {
		t.Fatal("independent consumer lost actual reference", err)
	}
	if err := standalone.Close(); err != nil {
		t.Fatal(err)
	}
	reopened, err := sqlite.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer reopened.Close()
	service = attention.NewAttentionService(sqlite.NewAttentionRepository(reopened), nil, nil, attention.ServiceOptions{})
	replay, err := service.ReferenceMaterial(ctx, attention.Principal{ID: caller.ID, Kind: caller.Kind}, created.Space.ID, command)
	if err != nil || !reflect.DeepEqual(replay, referenced) {
		t.Fatal("original material-reference receipt was not replayed", err)
	}
	var receiptsAfter int
	if err := reopened.Conn().QueryRow("SELECT COUNT(*) FROM attention_receipts").Scan(&receiptsAfter); err != nil || receiptsAfter != receiptsBefore {
		t.Fatal("migration or replay duplicated receipt", err)
	}
}
