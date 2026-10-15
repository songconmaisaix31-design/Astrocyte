package admissionbridge_test

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/google/uuid"
	"github.com/songconmaisaix31-design/Astrocyte/internal/adapters/admissionbridge"
	"github.com/songconmaisaix31-design/Astrocyte/internal/adapters/agents"
	"github.com/songconmaisaix31-design/Astrocyte/internal/adapters/importers"
	"github.com/songconmaisaix31-design/Astrocyte/internal/adapters/objects"
	"github.com/songconmaisaix31-design/Astrocyte/internal/adapters/sqlite"
	"github.com/songconmaisaix31-design/Astrocyte/internal/apierrors"
	attention "github.com/songconmaisaix31-design/Astrocyte/internal/attention/app"
	workspace "github.com/songconmaisaix31-design/Astrocyte/internal/workspace/app"
	"github.com/songconmaisaix31-design/Astrocyte/internal/workspace/domain"
)

func TestHumanAdmissionActualProjectAndTransaction(t *testing.T) {
	ctx := context.Background()
	db, directory := ownedDatabase(t)
	store, err := objects.New(filepath.Join(directory, "objects"))
	if err != nil {
		t.Fatal(err)
	}
	core := attention.NewAttentionService(sqlite.NewAttentionRepository(db), importers.NewReader(nil), store, attention.ServiceOptions{})
	human := attention.Principal{Kind: "human", ID: "local-human"}
	space, err := core.CreateProjectSpace(ctx, human, attention.ProjectSpaceCommand{CommandMeta: metadata(), Title: "Human-created admission project scope"})
	if err != nil {
		t.Fatal(err)
	}
	projects := workspace.NewLocalProjectService(db, nil, agents.ProjectFiles{}, nil)
	root, err := filepath.Abs("../../..")
	if err != nil {
		t.Fatal(err)
	}
	project, err := projects.RegisterProject(ctx, domain.Caller{Kind: human.Kind, ID: human.ID}, domain.RegisterProjectCommand{Name: "Public admission project", Root: root, SpaceID: space.Space.ID})
	if err != nil {
		t.Fatal(err)
	}
	authorizer, err := admissionbridge.NewHumanProjectAccess(projects, core)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := admissionbridge.NewHumanProjectAccess(nil, core); err == nil {
		t.Fatal("missing actual project service was accepted")
	}
	if _, err := admissionbridge.NewHumanProjectAccess(projects, nil); err == nil {
		t.Fatal("missing actual fixed reference service was accepted")
	}
	waits := db.Conn().Stats().WaitCount
	if err := sqlite.NewWorkspaceRepository(db).WithWorkspaceTx(ctx, func(tx workspace.WorkspaceTx) error {
		return authorizer.AuthorizeAdmissionProject(tx.Context(), human, project.ID, nil)
	}); err != nil {
		t.Fatal("current human project failed in its actual shared transaction", err)
	}
	if db.Conn().Stats().WaitCount != waits {
		t.Fatal("project authorization borrowed another SQLite connection")
	}
	for _, id := range []string{"", project.Root, project.SpaceID, " " + project.ID, uuid.NewString()} {
		err := authorizer.AuthorizeAdmissionProject(ctx, human, id, nil)
		requireCode(t, err, apierrors.ScopeDenied)
	}
	// 角色值是明确权限规则参数，不声明存在原生会话或委托判断。
	for _, principal := range []attention.Principal{{}, {Kind: "human"}, {Kind: "agent", ID: "rule-agent"}, {Kind: "delegated_judgment", ID: "unrecorded-workflow"}} {
		err := authorizer.AuthorizeAdmissionProject(ctx, principal, project.ID, nil)
		requireCode(t, err, apierrors.ScopeDenied)
	}
	other, _ := ownedDatabase(t)
	if err := sqlite.NewWorkspaceRepository(other).WithWorkspaceTx(ctx, func(tx workspace.WorkspaceTx) error {
		return authorizer.AuthorizeAdmissionProject(tx.Context(), human, project.ID, nil)
	}); err == nil {
		t.Fatal("project authorization accepted another actual database transaction")
	}
	cancelled, cancel := context.WithCancel(ctx)
	cancel()
	if err := authorizer.AuthorizeAdmissionProject(cancelled, human, project.ID, nil); !errors.Is(err, context.Canceled) {
		t.Fatal("actual cancellation was suppressed", err)
	}
	for _, table := range []string{"attention_admissions", "workspace_proposals", "workspace_approvals", "local_agent_sessions"} {
		if count := rows(t, db, table); count != 0 {
			t.Fatal("project authorization created business execution or source records", table, count)
		}
	}
	publication := os.Getenv("ASTROCYTE_TEST_PUBLIC_PAPER_SNAPSHOT")
	if !filepath.IsAbs(publication) {
		t.Fatal("select the absolute path of an actual public DOM snapshot")
	}
	raw, err := os.ReadFile(publication)
	if err != nil {
		t.Fatal(err)
	}
	var snapshot importers.PaperSnapshot
	if err := json.Unmarshal(raw, &snapshot); err != nil {
		t.Fatal(err)
	}
	imported, err := core.ImportMaterial(ctx, human, attention.ImportMaterialCommand{CommandMeta: metadata(), Kind: "paper", Adapter: "paper_snapshot", SourceLocator: snapshot.SourceURL, ExportText: string(raw)})
	if err != nil {
		t.Fatal(err)
	}
	if processed, err := core.ProcessNextJob(ctx); err != nil || !processed {
		t.Fatal("actual public DOM snapshot import failed", processed, err)
	}
	job, err := core.GetJob(ctx, human, imported.JobID)
	if err != nil || job.MaterialID == nil || job.Status != "succeeded" {
		t.Fatal("actual public DOM import did not publish its material", job.Status, err)
	}
	material, err := core.GetMaterial(ctx, human, *job.MaterialID)
	if err != nil || material.Revisions[0].SourceKey != snapshot.SourceKey {
		t.Fatal("actual imported DOM source identity changed", err)
	}
	ref := attention.SourceRef{MaterialID: material.Material.ID, Revision: material.Material.CurrentRevision, Locator: material.Material.SourceLocator}
	requireCode(t, authorizer.AuthorizeAdmissionProject(ctx, human, project.ID, []attention.SourceRef{ref}), apierrors.ScopeDenied)
	if _, err := core.ReferenceMaterial(ctx, human, project.SpaceID, attention.ReferenceMaterialCommand{CommandMeta: metadata(), MaterialID: ref.MaterialID, Revision: ref.Revision}); err != nil {
		t.Fatal("actual fixed source was not included through the production service", err)
	}
	waits = db.Conn().Stats().WaitCount
	if err := sqlite.NewWorkspaceRepository(db).WithWorkspaceTx(ctx, func(tx workspace.WorkspaceTx) error {
		return authorizer.AuthorizeAdmissionProject(tx.Context(), human, project.ID, []attention.SourceRef{ref})
	}); err != nil {
		t.Fatal("actual included fixed source failed current authorization", err)
	}
	if db.Conn().Stats().WaitCount != waits {
		t.Fatal("actual fixed source authorization borrowed another SQLite connection")
	}
	current, err := core.GetMaterial(ctx, human, material.Material.ID)
	if err != nil {
		t.Fatal(err)
	}
	withdrawal := metadata()
	withdrawal.ExpectedVersion = current.Material.Version
	if _, err := core.UpdateMaterial(ctx, human, material.Material.ID, attention.UpdateMaterialCommand{CommandMeta: withdrawal, Lifecycle: "withdrawn"}); err != nil {
		t.Fatal("actual material withdrawal failed", err)
	}
	requireCode(t, authorizer.AuthorizeAdmissionProject(ctx, human, project.ID, []attention.SourceRef{ref}), apierrors.ScopeDenied)
}
