package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/http/cookiejar"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/songconmaisaix31-design/Astrocyte/internal/adapters/agents"
	"github.com/songconmaisaix31-design/Astrocyte/internal/adapters/httpapi"
	"github.com/songconmaisaix31-design/Astrocyte/internal/adapters/objects"
	"github.com/songconmaisaix31-design/Astrocyte/internal/adapters/sqlite"
	"github.com/songconmaisaix31-design/Astrocyte/internal/apierrors"
	swarm "github.com/songconmaisaix31-design/Astrocyte/internal/swarm/app"
	swarmdomain "github.com/songconmaisaix31-design/Astrocyte/internal/swarm/domain"
	workspace "github.com/songconmaisaix31-design/Astrocyte/internal/workspace/app"
	"github.com/songconmaisaix31-design/Astrocyte/migrations"
)

func TestRelaySwarmActualPreparedAssemblySQLiteHTTP(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	root := relayActualTestDirectory(t)
	db, err := sqlite.Open(filepath.Join(root, "state.sqlite"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := db.Close(); err != nil {
			t.Error(err)
		}
	})
	if err := db.RunMigrations(ctx, migrations.FS); err != nil {
		t.Fatal(err)
	}
	db.Conn().SetMaxOpenConns(1)
	nativeRoot := filepath.Join(root, "native")
	registry := agents.NewRegistry(nativeRoot)
	projects := workspace.NewLocalProjectService(db, nil, agents.ProjectFiles{}, registry)
	store, err := objects.New(filepath.Join(root, "objects"))
	if err != nil {
		t.Fatal(err)
	}
	services := httpapi.Services{WorkspaceS2: workspace.NewS2Service(sqlite.NewWorkspaceRepository(db), agents.ProjectFiles{}, nil)}
	if err := assembleSwarmLedger(&services, db, registry, store); err == nil || services.SwarmS3 != nil {
		t.Fatal("缺少实际 Workspace 准备读取端口时完成了 Swarm 装配")
	}
	services.LocalProjects = projects
	if err := assembleSwarmLedger(&services, db, registry, store); err != nil || services.SwarmS3 == nil {
		t.Fatal("实际 Swarm 装配失败", err)
	}
	source, err := agents.NewRelayPreparedMemberEvidenceSource(registry, projects)
	if err != nil {
		t.Fatal(err)
	}
	repo, err := sqlite.NewSwarmRepository(db)
	if err != nil {
		t.Fatal(err)
	}
	if bridge, err := sqlite.NewSwarmNativeBridge(repo, source, nil); bridge != nil || err == nil {
		t.Fatal("缺少实际准备读取端口时建立了 Native bridge")
	}
	bridge, err := sqlite.NewSwarmNativeBridge(repo, source, source)
	if err != nil {
		t.Fatal(err)
	}
	assertCode := func(err error, code apierrors.Code) {
		t.Helper()
		var failure *apierrors.ServiceError
		if errors.As(err, &failure) && failure.Code == code {
			return
		}
		var rule *swarmdomain.RuleError
		if errors.As(err, &rule) && rule.Code == string(code) {
			return
		}
		t.Fatalf("拒绝码应为 %s，实际错误为 %v", code, err)
	}
	// UUID 只标识本次缺失查询，不登记原生会话、输入或租约。
	sessionID, operationID := uuid.NewString(), uuid.NewString()
	if err := sqlite.NewWorkspaceRepository(db).WithWorkspaceTx(ctx, func(tx workspace.WorkspaceTx) error {
		if lease, err := repo.LeaseForOperation(tx.Context(), operationID, sessionID); err == nil || lease.OperationID != "" {
			t.Fatal("空 SQLite 产生了租约")
		} else {
			assertCode(err, apierrors.NotFound)
		}
		if prepared, err := db.LoadNativeMessagePreparation(tx.Context(), sessionID, operationID); err == nil || prepared.Input.Session.ID != "" {
			t.Fatal("空 SQLite 产生了输入准备记录")
		} else {
			assertCode(err, apierrors.NotFound)
		}
		if input, err := source.ReadPreparedMemberInput(tx.Context(), sessionID, operationID); err == nil || input != (swarm.PreparedMemberInput{}) {
			t.Fatal("空原 Registry 产生了 Swarm 输入身份")
		}
		return nil
	}); err != nil {
		t.Fatal("真实单连接授权事务内读取失败", err)
	}
	if _, err := projects.ReadPreparedNativeSession(ctx, sessionID, operationID); err == nil {
		t.Fatal("生产未配置的准备服务意外启用")
	} else {
		assertCode(err, apierrors.UnsupportedCapability)
	}
	assertCode(bridge.CheckMember(ctx, swarmdomain.Principal{}, swarm.Member{}), apierrors.ContextStale)
	if descriptor, err := bridge.DescribeExecution(ctx, swarmdomain.Principal{}, swarm.Member{}); err == nil || descriptor.Known {
		t.Fatal("缺少当前授权绑定时产生了执行能力")
	}
	if _, err := bridge.Reconcile(ctx, operationID, sessionID); err == nil {
		t.Fatal("缺少租约时完成了原生对账")
	} else {
		assertCode(err, apierrors.NotFound)
	}
	if _, err := bridge.Stop(ctx, operationID, sessionID); err == nil {
		t.Fatal("缺少租约时完成了原生停止")
	} else {
		assertCode(err, apierrors.NotFound)
	}
	api := httpapi.NewServer(httpapi.Config{Services: services, Logger: slog.New(slog.NewTextHandler(io.Discard, nil))})
	server := httptest.NewServer(api.Handler())
	defer server.Close()
	jar, err := cookiejar.New(nil)
	if err != nil {
		t.Fatal(err)
	}
	client := &http.Client{Jar: jar, Timeout: 5 * time.Second}
	request := func(method, path string, body []byte, csrf string, status int, result any) {
		t.Helper()
		req, err := http.NewRequestWithContext(ctx, method, server.URL+path, bytes.NewReader(body))
		if err != nil {
			t.Fatal(err)
		}
		if body != nil {
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Idempotency-Key", uuid.NewString())
		}
		if csrf != "" {
			req.Header.Set("X-CSRF-Token", csrf)
		}
		response, err := client.Do(req)
		if err != nil {
			t.Fatal(err)
		}
		decodeErr := json.NewDecoder(response.Body).Decode(result)
		closeErr := response.Body.Close()
		if response.StatusCode != status || decodeErr != nil || closeErr != nil {
			t.Fatalf("HTTP %s %s：状态 %d，解码 %v，关闭 %v", method, path, response.StatusCode, decodeErr, closeErr)
		}
	}
	var denied apierrors.ErrorEnvelope
	request(http.MethodGet, "/api/v1/missions", nil, "", http.StatusForbidden, &denied)
	if denied.Error == nil || denied.Error.Code != apierrors.ScopeDenied {
		t.Fatal("未认证的实际 HTTP 没有拒绝 Swarm 读取")
	}
	var session struct {
		CSRF string `json:"csrf_token"`
	}
	request(http.MethodGet, "/api/v1/auth/session", nil, "", http.StatusOK, &session)
	if session.CSRF == "" {
		t.Fatal("实际人工会话缺少 CSRF")
	}
	var listing apierrors.ListResult
	request(http.MethodGet, "/api/v1/missions", nil, "", http.StatusOK, &listing)
	if len(listing.Items) != 0 {
		t.Fatal("实际空 SQLite 返回了 Mission")
	}
	command, err := json.Marshal(swarm.ClaimCommand{Meta: swarmdomain.Meta{SchemaVersion: 1, RequestID: uuid.NewString(), ExpectedVersion: 1}})
	if err != nil {
		t.Fatal(err)
	}
	denied = apierrors.ErrorEnvelope{}
	request(http.MethodPost, "/api/v1/work-items/"+uuid.NewString()+"/claims", command, session.CSRF, http.StatusForbidden, &denied)
	if denied.Error == nil || denied.Error.Code != apierrors.ScopeDenied || denied.Error.Message != "bound agent identity required" {
		t.Fatal("实际 Claim 路由没有执行 Swarm 成员身份拒绝")
	}
	for _, table := range []string{"swarm_missions", "swarm_work_items", "swarm_leases", "foundation_idempotency", "foundation_outbox", "local_agent_sessions", "workspace_session_bindings"} {
		var count int
		if err := db.Conn().QueryRowContext(ctx, "SELECT COUNT(*) FROM "+table).Scan(&count); err != nil || count != 0 {
			t.Fatalf("真实拒绝后 %s 出现记录或查询失败：%d，%v", table, count, err)
		}
	}
	if _, err := os.Stat(nativeRoot); !errors.Is(err, os.ErrNotExist) {
		t.Fatal("无原生检查创建了 Registry 会话目录或目录检查失败", err)
	}
}
