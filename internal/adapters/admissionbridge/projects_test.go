package admissionbridge_test

import (
	"context"
	"errors"
	"path/filepath"
	"testing"

	"github.com/google/uuid"
	"github.com/songconmaisaix31-design/Astrocyte/internal/adapters/admissionbridge"
	"github.com/songconmaisaix31-design/Astrocyte/internal/adapters/agents"
	"github.com/songconmaisaix31-design/Astrocyte/internal/adapters/sqlite"
	"github.com/songconmaisaix31-design/Astrocyte/internal/apierrors"
	attention "github.com/songconmaisaix31-design/Astrocyte/internal/attention/app"
	workspace "github.com/songconmaisaix31-design/Astrocyte/internal/workspace/app"
	"github.com/songconmaisaix31-design/Astrocyte/internal/workspace/domain"
)

func TestHumanAdmissionActualProjectAndTransaction(t *testing.T) {
	ctx := context.Background()
	db, _ := ownedDatabase(t)
	core := attention.NewAttentionService(sqlite.NewAttentionRepository(db), nil, nil, attention.ServiceOptions{})
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
	authorizer, err := admissionbridge.NewHumanProjectAccess(projects)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := admissionbridge.NewHumanProjectAccess(nil); err == nil {
		t.Fatal("missing actual project service was accepted")
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
}
