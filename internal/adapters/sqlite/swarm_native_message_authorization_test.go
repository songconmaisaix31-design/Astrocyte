package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/songconmaisaix31-design/Astrocyte/internal/adapters/agents"
	"github.com/songconmaisaix31-design/Astrocyte/internal/adapters/objects"
	"github.com/songconmaisaix31-design/Astrocyte/internal/apierrors"
	attention "github.com/songconmaisaix31-design/Astrocyte/internal/attention/app"
	swarm "github.com/songconmaisaix31-design/Astrocyte/internal/swarm/app"
	workspace "github.com/songconmaisaix31-design/Astrocyte/internal/workspace/app"
	"github.com/songconmaisaix31-design/Astrocyte/internal/workspace/domain"
	"github.com/songconmaisaix31-design/Astrocyte/migrations"
)

func actualNativeMessageAuthorizer(t *testing.T, db *DB, prepared domain.RelayPreparedInputReader, registry *agents.Registry, store workspace.ContextObjectStore) *SwarmNativeMessageAuthorizer {
	t.Helper()
	grants, err := NewSwarmApprovalBridge(db, workspace.NewS2Service(NewWorkspaceRepository(db), agents.ProjectFiles{}, nil))
	if err != nil {
		t.Fatal(err)
	}
	repo, err := NewSwarmRepository(db)
	if err != nil {
		t.Fatal(err)
	}
	source, err := agents.NewRelayPreparedMemberEvidenceSource(registry, prepared)
	if err != nil {
		t.Fatal(err)
	}
	members, err := NewSwarmNativeBridge(repo, source, source)
	if err != nil {
		t.Fatal(err)
	}
	evidence, err := NewSwarmGitEvidence(store)
	if err != nil {
		t.Fatal(err)
	}
	service, err := swarm.NewSwarmService(repo, grants, members, evidence, members)
	if err != nil {
		t.Fatal(err)
	}
	authorizer, err := NewSwarmNativeMessageAuthorizer(db, grants, service)
	if err != nil {
		t.Fatal(err)
	}
	return authorizer
}

func TestSwarmNativeMessageActualAuthorizationLifecycle(t *testing.T) {
	directory := acceptedExperienceDirectory(t)
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	db, err := Open(filepath.Join(directory, "state.sqlite"))
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
	other := openAttentionDB(t, filepath.Join(directory, "other.sqlite"))
	registry := agents.NewRegistry(filepath.Join(directory, "native"))
	store, err := objects.New(filepath.Join(directory, "objects"))
	if err != nil {
		t.Fatal(err)
	}
	projects := workspace.NewLocalProjectService(db, nil, agents.ProjectFiles{}, registry).WithNativeMessageObjects(store).WithOwnedNativeSessions(registry)
	authorizer := actualNativeMessageAuthorizer(t, db, projects, registry, store)
	projects.WithNativeMessageAuthorizer(authorizer)
	if _, err := NewSwarmNativeMessageAuthorizer(other, authorizer.grants, authorizer.service); err == nil {
		t.Fatal("发送授权接受了其他数据库的批准桥接")
	}
	called := false
	callback := func(context.Context) error { called = true; return nil }
	request := workspace.NativeMessageAuthorization{}
	for _, authorize := range []func(context.Context, workspace.NativeMessageAuthorization, func(context.Context) error) error{authorizer.WithNativeMessagePreparation, authorizer.WithNativeMessageSend} {
		canceled, stop := context.WithCancel(ctx)
		stop()
		if err := authorize(canceled, request, callback); !errors.Is(err, context.Canceled) || called {
			t.Fatal("发送授权没有在取消处停止", err)
		}
		if err := authorize(ctx, request, callback); err == nil || called {
			t.Fatal("未明确用途的输入通过授权", err)
		}
		for _, database := range []*DB{db, other} {
			if err := database.WithNativeMessageTx(ctx, func(txContext context.Context) error {
				if err := authorize(txContext, request, callback); err == nil || called {
					t.Fatal("发送授权在调用者事务或其他数据库事务内继续执行", err)
				}
				return nil
			}); err != nil {
				t.Fatal(err)
			}
		}
	}
	// UUID 只作为不存在的查询键，不保存会话、准备或租约。
	sessionID, operationID := uuid.NewString(), uuid.NewString()
	if err := db.requireProjectNativeMessageBindings(ctx, uuid.NewString(), sessionID); err == nil {
		t.Fatal("事务外读取了原成员绑定的用途")
	}
	if err := db.BindSwarmNativeMessageLease(ctx, sessionID, operationID, workspace.NativeMessageLease{}); err == nil {
		t.Fatal("事务外绑定了准备租约")
	}
	if err := db.WithNativeMessageTx(ctx, func(txContext context.Context) error {
		before := db.Conn().Stats().WaitCount
		if err := db.requireProjectNativeMessageBindings(txContext, uuid.NewString(), sessionID); err != nil {
			t.Fatal("真实空绑定查询失败", err)
		}
		if err := other.requireProjectNativeMessageBindings(txContext, uuid.NewString(), sessionID); err == nil {
			t.Fatal("原绑定用途读取接受了其他数据库事务")
		}
		if err := db.BindSwarmNativeMessageLease(txContext, sessionID, operationID, workspace.NativeMessageLease{}); err == nil {
			t.Fatal("缺失原准备时绑定了租约")
		}
		if err := db.requireProjectNativeMessageScope(txContext, sessionID); !errors.Is(err, sql.ErrNoRows) {
			t.Fatal("缺失 session 产生了 project 用途", err)
		}
		if db.Conn().Stats().WaitCount != before {
			t.Fatal("准备绑定没有复用单连接事务")
		}
		return nil
	}); err != nil {
		t.Fatal(err)
	}
	owner := domain.Caller{Kind: "human", ID: uuid.NewString(), OperationID: operationID}
	space, err := attention.NewAttentionService(NewAttentionRepository(db), nil, nil, attention.ServiceOptions{}).CreateProjectSpace(ctx, attention.Principal{Kind: owner.Kind, ID: owner.ID}, attention.ProjectSpaceCommand{
		CommandMeta: attention.CommandMeta{SchemaVersion: 1, RequestID: uuid.NewString(), IdempotencyKey: uuid.NewString(), ExpectedVersion: 1}, Title: "Send authorization actual lifecycle",
	})
	if err != nil {
		t.Fatal(err)
	}
	root, err := filepath.Abs("../../..")
	if err != nil {
		t.Fatal(err)
	}
	project, err := projects.RegisterProject(ctx, owner, domain.RegisterProjectCommand{Name: "Public source checkout", Root: root, SpaceID: space.Space.ID})
	if err != nil {
		t.Fatal(err)
	}
	message, err := os.ReadFile(filepath.Join(root, "README.md"))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := projects.PrepareSwarmNativeMessage(ctx, owner, project.ID, sessionID, uuid.NewString(), string(message)); err == nil {
		t.Fatal("缺失原生来源时保存了 Swarm 准备")
	}
	if _, err := projects.SendNativeMessage(ctx, owner, project.ID, sessionID, string(message)); err == nil {
		t.Fatal("缺失原生来源时发送了输入")
	}
	// 只向端口提供缺失查询键，没有 NativeID、controller 或准备来源。
	request = workspace.NativeMessageAuthorization{Caller: owner, Preparation: workspace.NativeMessagePreparation{
		Input: domain.RelayPreparedInput{OperationID: operationID, ActorID: owner.ID, ActorKind: owner.Kind, Session: domain.NativeSession{ID: sessionID, ProjectID: project.ID}},
		Scope: workspace.NativeMessageScope{Kind: "project"},
	}}
	if err := authorizer.WithNativeMessageSend(ctx, request, callback); !errors.Is(err, sql.ErrNoRows) || called {
		t.Fatal("缺失原 session 被推断为 project 发送", err)
	}
	request.Preparation.Scope = workspace.NativeMessageScope{Kind: "swarm", BindingID: uuid.NewString()}
	var missingLease *apierrors.ServiceError
	if err := authorizer.WithNativeMessageSend(ctx, request, callback); !errors.As(err, &missingLease) || missingLease.Code != apierrors.EvidenceMissing || called {
		t.Fatal("明确的 Swarm 用途缺少原租约时没有拒绝", err)
	}
	memberID := uuid.NewString()
	err = db.WithNativeMessageTx(ctx, func(txContext context.Context) error {
		if _, err := projects.GrantProjectAgent(txContext, owner, project.ID, domain.ProjectGrant{AgentID: memberID, Actions: []string{"send"}}); err != nil {
			return err
		}
		canceled, stop := context.WithCancel(txContext)
		stop()
		return canceled.Err()
	})
	if !errors.Is(err, context.Canceled) {
		t.Fatal("取消没有终止实际成员授权事务", err)
	}
	if _, err := db.LoadGrant(ctx, project.ID, memberID); err == nil {
		t.Fatal("取消的事务留下了成员授权")
	}
	if _, err := projects.GrantProjectAgent(ctx, owner, project.ID, domain.ProjectGrant{AgentID: memberID, Actions: []string{"send"}}); err != nil {
		t.Fatal(err)
	}
	token, err := projects.IssueProjectAgentToken(ctx, owner, project.ID, memberID)
	if err != nil {
		t.Fatal(err)
	}
	member, err := projects.AuthenticateAgentToken(ctx, token.Token)
	if err != nil {
		t.Fatal(err)
	}
	// 只核对真实成员授权的时间限制，不调用原生发送。
	if err := authorizer.withMemberDeadline(context.Background(), member, func(bounded context.Context) error {
		deadline, ok := bounded.Deadline()
		if !ok || !deadline.Equal(token.ExpiresAt) {
			t.Fatal("没有保留成员授权的绝对截止时间")
		}
		return nil
	}); err != nil {
		t.Fatal(err)
	}
	if _, err := projects.RevokeProjectAgent(ctx, owner, project.ID, memberID); err != nil {
		t.Fatal(err)
	}
	assertRevoked := func() {
		t.Helper()
		called = false
		err := authorizer.withMemberDeadline(ctx, member, callback)
		var denied *apierrors.ServiceError
		if !errors.As(err, &denied) || denied.Code != apierrors.ApprovalRevoked || called {
			t.Fatal("已撤销的真实成员授权进入发送回调", err)
		}
	}
	assertRevoked()
	if err := db.Close(); err != nil {
		t.Fatal(err)
	}
	db, err = Open(filepath.Join(directory, "state.sqlite"))
	if err != nil {
		t.Fatal(err)
	}
	db.Conn().SetMaxOpenConns(1)
	projects = workspace.NewLocalProjectService(db, nil, agents.ProjectFiles{}, registry).WithNativeMessageObjects(store).WithOwnedNativeSessions(registry)
	authorizer = actualNativeMessageAuthorizer(t, db, projects, registry, store)
	projects.WithNativeMessageAuthorizer(authorizer)
	assertRevoked()
	for _, table := range []string{"local_agent_sessions", "swarm_leases", "swarm_work_items", "swarm_missions", "workspace_session_bindings"} {
		var count int
		if err := db.Conn().QueryRowContext(ctx, "SELECT COUNT(*) FROM "+table).Scan(&count); err != nil || count != 0 {
			t.Fatal("实际拒绝及重开检查创建了执行来源", table, count, err)
		}
	}
	files, err := os.ReadDir(filepath.Join(directory, "objects"))
	if err != nil || len(files) != 0 {
		t.Fatal("缺失来源时发布了输入对象", err)
	}
	if _, err := os.Stat(filepath.Join(directory, "native")); !errors.Is(err, os.ErrNotExist) {
		t.Fatal("拒绝检查初始化了原生控制器目录", err)
	}
}
