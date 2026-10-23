package main

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/songconmaisaix31-design/Astrocyte/internal/adapters/agents"
	"github.com/songconmaisaix31-design/Astrocyte/internal/adapters/sqlite"
	"github.com/songconmaisaix31-design/Astrocyte/internal/apierrors"
	attention "github.com/songconmaisaix31-design/Astrocyte/internal/attention/app"
	workspace "github.com/songconmaisaix31-design/Astrocyte/internal/workspace/app"
	"github.com/songconmaisaix31-design/Astrocyte/internal/workspace/domain"
	"github.com/songconmaisaix31-design/Astrocyte/migrations"
)

// 使用本仓库公开文件建立独立 Git 项目，所有授权和读取均经过真实 SQLite 服务。
func TestRelayPublicContextActualGitSQLite(t *testing.T) {
	ctx := context.Background()
	local := relayActualTestDirectory(t)
	t.Setenv("LOCALAPPDATA", local)
	access, err := relayRoot()
	if err != nil {
		relayActualTestFailure(t, local, "独立接入根不可用", err, map[string]string{"stage": "relayRoot"})
	}
	root := filepath.Join(access, "verification", "public-context")
	if err := os.MkdirAll(root, 0700); err != nil {
		t.Fatal("公开项目目录不可用")
	}
	files := []string{"README.md", "LICENSE", "go.mod", ".gitignore"}
	for _, file := range files {
		data, err := os.ReadFile(filepath.Join("..", "..", file))
		if err != nil || os.WriteFile(filepath.Join(root, file), data, 0600) != nil {
			t.Fatal("公开仓库文件复制失败")
		}
	}
	git := func(args ...string) string {
		t.Helper()
		command := exec.CommandContext(ctx, "git", append([]string{"-C", root}, args...)...)
		command.Stderr = io.Discard
		output, err := command.Output()
		if err != nil {
			t.Fatal("独立公开 Git 操作失败")
		}
		return strings.TrimSpace(string(output))
	}
	git("init", "--initial-branch=main")
	git("config", "core.autocrlf", "false")
	git("add", "--", "README.md", "LICENSE", "go.mod", ".gitignore")
	git("-c", "user.name=Relay Verification", "-c", "user.email=relay@example.invalid", "-c", "commit.gpgsign=false", "commit", "-m", "Select public repository context")
	commit := git("rev-parse", "HEAD")
	encoded, err := json.Marshal(files)
	if err != nil {
		t.Fatal("明确文件选择编码失败")
	}
	t.Setenv("ASTROCYTE_RELAY_PUBLIC_GIT_ROOT", root)
	t.Setenv("ASTROCYTE_RELAY_PUBLIC_GIT_COMMIT", commit)
	t.Setenv("ASTROCYTE_RELAY_PUBLIC_CONTEXT_FILES", string(encoded))
	db, err := sqlite.Open(filepath.Join(local, "application.sqlite"))
	if err != nil {
		t.Fatal("实际数据库不可用")
	}
	defer db.Close()
	if db.RunMigrations(ctx, migrations.FS) != nil {
		t.Fatal("实际数据库迁移失败")
	}
	human := domain.Caller{Kind: "human", ID: uuid.NewString()}
	spaces := attention.NewAttentionService(sqlite.NewAttentionRepository(db), nil, nil, attention.ServiceOptions{ProjectSpaces: testAttentionSpaces(db)})
	space, err := spaces.CreateProjectSpace(ctx, attention.Principal{Kind: human.Kind, ID: human.ID}, attention.ProjectSpaceCommand{CommandMeta: attention.CommandMeta{SchemaVersion: 1, RequestID: uuid.NewString(), IdempotencyKey: uuid.NewString(), ExpectedVersion: 1}, Title: "Public Relay context verification"})
	if err != nil {
		t.Fatal("实际项目空间创建失败")
	}
	projects := workspace.NewLocalProjectService(db, nil, agents.ProjectFiles{}, nil)
	project, err := projects.RegisterProject(ctx, human, domain.RegisterProjectCommand{Name: "Public Relay context", Root: root, SpaceID: space.Space.ID})
	if err != nil {
		t.Fatal("实际项目登记失败")
	}
	readers, err := relayPublicLaunchContexts(projects)
	if err != nil || len(readers) != 1 {
		t.Fatal("固定上下文没有配置准确读取器")
	}
	reader := readers[0]
	launch := workspace.RelayLaunch{Project: project}
	if _, err := reader.ReadRelayLaunchContext(ctx, human, launch); err == nil {
		t.Fatal("未批准目录读取没有拒绝")
	}
	settings := project.Settings
	settings.AllowDirectory, settings.AllowedSubdirs = true, []string{"."}
	settings.AllowedActions = []string{"read_context"}
	project, err = projects.SetProjectSettings(ctx, human, project.ID, domain.SettingsCommand{ExpectedRevision: project.Settings.Revision, Settings: settings})
	if err != nil {
		t.Fatal("实际项目目录授权失败")
	}
	launch.Project = project
	packet, err := reader.ReadRelayLaunchContext(ctx, human, launch)
	if err != nil || len(packet.Files) != 4 || packet.ProjectID != project.ID || packet.SettingsRevision != project.Settings.Revision {
		t.Fatal("固定四文件真实读取失败")
	}
	for i, file := range packet.Files {
		original, err := os.ReadFile(filepath.Join(root, files[i]))
		if err != nil || file.Path != files[i] || file.Text != string(original) || file.Version == "" {
			t.Fatal("固定上下文没有保留实际文件正文和版本")
		}
	}
	assertCode := func(err error, expected apierrors.Code) {
		t.Helper()
		var failure *apierrors.ServiceError
		if !errors.As(err, &failure) || failure.Code != expected || strings.Contains(failure.Message, root) {
			t.Fatal("上下文拒绝没有返回准确安全错误")
		}
	}
	settings = project.Settings
	settings.AllowDirectory = false
	_, err = projects.SetProjectSettings(ctx, human, project.ID, domain.SettingsCommand{ExpectedRevision: project.Settings.Revision, Settings: settings})
	if err != nil {
		t.Fatal("实际目录授权撤销失败")
	}
	_, err = reader.ReadRelayLaunchContext(ctx, human, launch)
	assertCode(err, apierrors.ScopeDenied)
	project, err = projects.SetProjectSettings(ctx, human, project.ID, domain.SettingsCommand{ExpectedRevision: project.Settings.Revision + 1, Settings: project.Settings})
	if err != nil {
		t.Fatal("实际目录授权更新失败")
	}
	_, err = reader.ReadRelayLaunchContext(ctx, human, launch)
	assertCode(err, apierrors.ScopeDenied)
	launch.Project = project
	if err := os.WriteFile(filepath.Join(root, "README.md"), []byte(packet.Files[0].Text+"\n"), 0600); err != nil {
		t.Fatal("实际公开文件变更失败")
	}
	_, err = reader.ReadRelayLaunchContext(ctx, human, launch)
	assertCode(err, apierrors.ContextStale)
	git("add", "--", "README.md")
	git("-c", "user.name=Relay Verification", "-c", "user.email=relay@example.invalid", "-c", "commit.gpgsign=false", "commit", "-m", "Update public context revision")
	_, err = reader.ReadRelayLaunchContext(ctx, human, launch)
	assertCode(err, apierrors.ContextStale)
	for _, selection := range []string{`["../README.md"]`, `["README.md","readme.md"]`, `["README.md"] {}`, `[]`} {
		t.Setenv("ASTROCYTE_RELAY_PUBLIC_CONTEXT_FILES", selection)
		if _, err := relayPublicLaunchContexts(projects); err == nil {
			t.Fatal("不安全或不完整文件选择没有拒绝")
		}
	}
	var sessions int
	if db.Conn().QueryRowContext(ctx, "SELECT COUNT(*) FROM local_agent_sessions").Scan(&sessions) != nil || sessions != 0 {
		t.Fatal("固定上下文检查产生原生会话")
	}
}
