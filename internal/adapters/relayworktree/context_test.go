package relayworktree

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/songconmaisaix31-design/Astrocyte/internal/adapters/agents"
	"github.com/songconmaisaix31-design/Astrocyte/internal/adapters/sqlite"
	"github.com/songconmaisaix31-design/Astrocyte/internal/apierrors"
	"github.com/songconmaisaix31-design/Astrocyte/internal/workspace/app"
	"github.com/songconmaisaix31-design/Astrocyte/internal/workspace/domain"
	"github.com/songconmaisaix31-design/Astrocyte/migrations"
)

// 真实 Git、SQLite、项目授权和文件读取；不启动原生进程或构造成功响应。
func TestFixedRelayContextUsesAuthorizedProjectAndExactBaseline(t *testing.T) {
	ctx := context.Background()
	root := cleanRepo(t)
	head, err := git(ctx, root, "rev-parse", "HEAD")
	if err != nil {
		t.Fatal(err)
	}
	db, err := sqlite.Open(filepath.Join(t.TempDir(), "context.sqlite"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = db.Close() })
	if err = db.RunMigrations(ctx, migrations.FS); err != nil {
		t.Fatal(err)
	}
	projects := app.NewLocalProjectService(db, nil, agents.ProjectFiles{}, agents.NewRegistry(filepath.Join(t.TempDir(), "controller")))
	caller := domain.Caller{Kind: "human", ID: "public-context", OperationID: uuid.NewString()}
	p, err := projects.RegisterProject(ctx, caller, domain.RegisterProjectCommand{Name: "Public fixed context", Root: root, SpaceID: "public"})
	if err != nil {
		t.Fatal(err)
	}
	p, err = projects.SetProjectSettings(ctx, caller, p.ID, domain.SettingsCommand{ExpectedRevision: p.Settings.Revision, Settings: domain.ProjectSettings{AllowDirectory: true, AllowedSubdirs: []string{"."}, AllowedActions: []string{"read_context", "start"}, ExternalModelCLI: "codex"}})
	if err != nil {
		t.Fatal(err)
	}
	selection := FixedContextSelection{ProjectID: p.ID, Root: root, BaselineCommit: strings.TrimSpace(head), SettingsRevision: p.Settings.Revision, Files: []string{"public.txt"}}
	reader, err := NewFixedContext(projects, []FixedContextSelection{selection})
	if err != nil {
		t.Fatal(err)
	}
	// 组装后外部切片修改不能扩大原来的固定文件选择。
	selection.Files[0] = "other.txt"
	launch := app.RelayLaunch{Project: p, Workspace: domain.RelayWorkspace{SourceRoot: root, Root: t.TempDir()}}
	packet, err := reader.ReadRelayLaunchContext(ctx, caller, launch)
	if err != nil || len(packet.Files) != 1 || packet.Files[0].Path != "public.txt" || packet.Files[0].Text != "committed public baseline\n" || packet.SettingsRevision != p.Settings.Revision {
		t.Fatal("real authorized fixed source was not passed as its exact packet")
	}
	assertDenied := func(err error, code apierrors.Code) {
		t.Helper()
		var known *apierrors.ServiceError
		if !errors.As(err, &known) || known.Code != code {
			t.Fatal("fixed context refusal differs")
		}
	}
	altered := launch
	altered.Workspace.SourceRoot = t.TempDir()
	_, err = reader.ReadRelayLaunchContext(ctx, caller, altered)
	assertDenied(err, apierrors.ScopeDenied)
	if err = os.WriteFile(filepath.Join(root, "public.txt"), []byte("uncommitted public change\n"), 0600); err != nil {
		t.Fatal(err)
	}
	_, err = reader.ReadRelayLaunchContext(ctx, caller, launch)
	assertDenied(err, apierrors.VersionConflict)
	testGit(t, root, "add", "public.txt")
	testGit(t, root, "-c", "user.name=Public Test", "-c", "user.email=public-test@example.invalid", "commit", "-qm", "new public baseline")
	_, err = reader.ReadRelayLaunchContext(ctx, caller, launch)
	assertDenied(err, apierrors.VersionConflict)
}
