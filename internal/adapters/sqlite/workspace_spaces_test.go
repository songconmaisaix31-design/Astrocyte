package sqlite

import (
	"context"
	"encoding/json"
	"errors"
	"path/filepath"
	"testing"

	"github.com/google/uuid"
	"github.com/songconmaisaix31-design/Astrocyte/internal/adapters/projectspacebridge"
	"github.com/songconmaisaix31-design/Astrocyte/internal/apierrors"
	attention "github.com/songconmaisaix31-design/Astrocyte/internal/attention/app"
	workspace "github.com/songconmaisaix31-design/Astrocyte/internal/workspace/app"
	"github.com/songconmaisaix31-design/Astrocyte/internal/workspace/domain"
)

func attentionSpaces(db *DB) attention.ProjectSpaces {
	return projectspacebridge.Attention{Spaces: workspace.NewProjectSpaceService(NewProjectSpaceRepository(db), workspace.ProjectSpaceOptions{})}
}

func TestWorkspaceProjectSpacesActualPersistenceAndReplay(t *testing.T) {
	db := openAttentionDB(t, filepath.Join(acceptedExperienceDirectory(t), "state.sqlite"))
	ctx := context.Background()
	service := workspace.NewProjectSpaceService(NewProjectSpaceRepository(db), workspace.ProjectSpaceOptions{})
	caller := domain.Caller{ID: "space-owner", Kind: "human"}
	key := uuid.NewString()
	command := workspace.ProjectSpaceCommand{WorkspaceCommandMeta: workspace.WorkspaceCommandMeta{SchemaVersion: 1, RequestID: key, IdempotencyKey: key, ExpectedVersion: 1}, Title: "Actual independent Workspace space"}
	created, err := service.CreateProjectSpace(ctx, caller, command)
	if err != nil || created.Space.ID == "" || created.Space.MaterialRefs == nil {
		t.Fatal("create actual Workspace space", err)
	}
	command.RequestID = uuid.NewString()
	replayed, err := service.CreateProjectSpace(ctx, caller, command)
	if err != nil || replayed.Space.ID != created.Space.ID {
		t.Fatal("changed trace repeated the space creation", err)
	}
	legacy := attention.NewAttentionService(NewAttentionRepository(db), nil, nil, attention.ServiceOptions{ProjectSpaces: attentionSpaces(db)})
	oldReplay, err := legacy.CreateProjectSpace(ctx, attention.Principal{ID: caller.ID, Kind: caller.Kind}, attention.ProjectSpaceCommand{CommandMeta: attention.CommandMeta{SchemaVersion: 1, RequestID: uuid.NewString(), IdempotencyKey: key, ExpectedVersion: 1}, Title: command.Title})
	if err != nil || oldReplay.Space.ID != created.Space.ID {
		t.Fatal("Attention consumer did not use the original Workspace receipt", err)
	}
	command.Title = "Different semantic input"
	_, err = service.CreateProjectSpace(ctx, caller, command)
	var conflictError *apierrors.ServiceError
	if !errors.As(err, &conflictError) || conflictError.Code != apierrors.VersionConflict {
		t.Fatal("same key accepted different space input", err)
	}
	for _, table := range []string{"attention_project_spaces", "attention_receipts", "attention_outbox"} {
		var count int
		if err := db.Conn().QueryRowContext(ctx, "SELECT COUNT(*) FROM "+table).Scan(&count); err != nil || count != 1 {
			t.Fatal("space replay created additional durable state", table, count, err)
		}
	}
	read, err := service.GetProjectSpace(ctx, caller, created.Space.ID)
	if err != nil || read.Space.Version != 1 || read.Space.MaterialRefs == nil || len(read.Space.MaterialRefs) != 0 {
		t.Fatal("Workspace changed the original empty space representation", err)
	}
	var raw []byte
	if err = db.Conn().QueryRowContext(ctx, "SELECT data FROM attention_project_spaces WHERE id=?", created.Space.ID).Scan(&raw); err != nil {
		t.Fatal(err)
	}
	var original attention.ProjectSpace
	if err = json.Unmarshal(raw, &original); err != nil || original.ID != created.Space.ID || original.Version != 1 || original.MaterialRefs == nil {
		t.Fatal("legacy durable space encoding changed", err)
	}
	for _, denied := range []domain.Caller{{ID: "agent", Kind: "agent"}, {Kind: "human"}} {
		_, err = service.ListProjectSpaces(ctx, denied)
		if !errors.As(err, &conflictError) || conflictError.Code != apierrors.ScopeDenied {
			t.Fatal("unauthorized caller read spaces", err)
		}
	}
	unconfigured := attention.NewAttentionService(NewAttentionRepository(db), nil, nil, attention.ServiceOptions{})
	_, err = unconfigured.CreateProjectSpace(ctx, attention.Principal{ID: caller.ID, Kind: caller.Kind}, attention.ProjectSpaceCommand{})
	if !errors.As(err, &conflictError) || conflictError.Code != apierrors.UnsupportedCapability {
		t.Fatal("unconfigured Attention retained its own space creator", err)
	}
}
