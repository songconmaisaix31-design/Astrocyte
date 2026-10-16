package main

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net"
	"net/http"
	"net/http/cookiejar"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/songconmaisaix31-design/Astrocyte/internal/adapters/agents"
	"github.com/songconmaisaix31-design/Astrocyte/internal/adapters/distillers"
	"github.com/songconmaisaix31-design/Astrocyte/internal/adapters/httpapi"
	"github.com/songconmaisaix31-design/Astrocyte/internal/adapters/importers"
	"github.com/songconmaisaix31-design/Astrocyte/internal/adapters/objects"
	"github.com/songconmaisaix31-design/Astrocyte/internal/adapters/sqlite"
	"github.com/songconmaisaix31-design/Astrocyte/internal/apierrors"
	attention "github.com/songconmaisaix31-design/Astrocyte/internal/attention/app"
	workspace "github.com/songconmaisaix31-design/Astrocyte/internal/workspace/app"
	"github.com/songconmaisaix31-design/Astrocyte/internal/workspace/domain"
	"github.com/songconmaisaix31-design/Astrocyte/migrations"
)

func TestSelectedTextActualAssemblyPrerequisitesWithoutNative(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	publication := os.Getenv("ASTROCYTE_TEST_PUBLIC_PAPER_SNAPSHOT")
	if !filepath.IsAbs(publication) || !filepath.IsAbs(os.Getenv("ASTROCYTE_TEST_OUTPUT_ROOT")) {
		t.Fatal("必须明确选择实际公开论文快照和自有输出绝对目录")
	}
	sourceFile, err := os.Open(publication)
	if err != nil {
		t.Fatal(err)
	}
	sourceInfo, statErr := sourceFile.Stat()
	raw, readErr := io.ReadAll(sourceFile)
	if err := errors.Join(statErr, readErr, sourceFile.Close()); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		currentInfo, err := os.Stat(publication)
		if err != nil {
			t.Error(err)
			return
		}
		current, err := os.ReadFile(publication)
		if err != nil || !os.SameFile(sourceInfo, currentInfo) || sourceInfo.Size() != currentInfo.Size() || !sourceInfo.ModTime().Equal(currentInfo.ModTime()) || !bytes.Equal(raw, current) {
			t.Error("公开快照的原文件身份、大小、修改时间或字节发生变化", err)
		}
	})
	var paper importers.PaperSnapshot
	if err := json.Unmarshal(raw, &paper); err != nil {
		t.Fatal(err)
	}
	if paper.SourceURL != "https://arxiv.org/html/2501.12948v1" || paper.SourceKey != "arxiv:2501.12948" || paper.ObservedVersion != "v1" || paper.ContentState != "readable_fulltext" || paper.Truncated || paper.Text == "" {
		t.Fatal("本次来源必须为获准的固定版本完整公开论文 DOM 快照")
	}
	directory := relayActualTestDirectory(t)
	db, err := sqlite.Open(filepath.Join(directory, "state.sqlite"))
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
	objectRoot, nativeRoot := filepath.Join(directory, "objects"), filepath.Join(directory, "unused-native")
	store, err := objects.New(objectRoot)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir(nativeRoot, 0700); err != nil {
		t.Fatal(err)
	}
	registry := agents.NewRegistry(nativeRoot)
	baseline, err := registry.SnapshotNative(ctx)
	if err != nil {
		t.Fatal(err)
	}
	core := attention.NewAttentionService(sqlite.NewAttentionRepository(db), importers.NewReader(nil), store, attention.ServiceOptions{})
	projects := workspace.NewLocalProjectService(db, &projectReferenceBridge{attention: core}, agents.ProjectFiles{}, registry)
	bridge := &selectedTextBridge{projects: projects, processor: registry}
	server := httpapi.NewServer(httpapi.Config{
		Logger:   slog.New(slog.NewTextHandler(io.Discard, nil)),
		Services: httpapi.Services{LocalProjects: projects, Attention: core},
	})
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	origin := "http://" + listener.Addr().String()
	done := make(chan error, 1)
	go func() { done <- server.Serve(listener) }()
	stopped := false
	stop := func() {
		if stopped {
			return
		}
		shutdown, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := server.Shutdown(shutdown); err != nil {
			t.Fatal(err)
		}
		if err := <-done; !errors.Is(err, http.ErrServerClosed) {
			t.Fatal("实际 HTTP 服务没有正常关闭", err)
		}
		stopped = true
	}
	t.Cleanup(stop)
	jar, err := cookiejar.New(nil)
	if err != nil {
		t.Fatal(err)
	}
	client := &http.Client{Jar: jar, Timeout: 5 * time.Second}
	t.Cleanup(client.CloseIdleConnections)
	var session struct {
		ID   string `json:"actor_id"`
		Kind string `json:"actor_kind"`
		CSRF string `json:"csrf_token"`
	}
	request := func(method, route string, command map[string]any, status int, output any) {
		t.Helper()
		var body []byte
		if command != nil {
			command["schema_version"], command["request_id"], command["expected_version"] = 1, uuid.NewString(), 1
			var err error
			body, err = json.Marshal(command)
			if err != nil {
				t.Fatal(err)
			}
		}
		req, err := http.NewRequestWithContext(ctx, method, origin+route, bytes.NewReader(body))
		if err != nil {
			t.Fatal(err)
		}
		if command != nil {
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("X-CSRF-Token", session.CSRF)
			req.Header.Set("Idempotency-Key", uuid.NewString())
		}
		response, err := client.Do(req)
		if err != nil {
			t.Fatal(err)
		}
		data, readErr := io.ReadAll(response.Body)
		if err := errors.Join(readErr, response.Body.Close()); err != nil || response.StatusCode != status {
			t.Fatal("实际 HTTP 状态不符", method, route, response.StatusCode, status, err)
		}
		if err := json.Unmarshal(data, output); err != nil {
			t.Fatal(err)
		}
	}
	request(http.MethodGet, "/api/v1/auth/session", nil, http.StatusOK, &session)
	if session.Kind != "human" || session.ID == "" || session.CSRF == "" {
		t.Fatal("实际 bootstrap 没有产生完整人类身份")
	}
	human := attention.Principal{ID: session.ID, Kind: session.Kind}
	var space attention.ProjectSpaceResult
	request(http.MethodPost, "/api/v1/project-spaces", map[string]any{"title": "公开论文实际选定文本前置验证"}, http.StatusCreated, &space)
	root, err := filepath.Abs("../..")
	if err != nil {
		t.Fatal(err)
	}
	root, err = agents.CanonicalRoot(root)
	if err != nil {
		t.Fatal(err)
	}
	var registered struct {
		Project domain.LocalProject `json:"project"`
	}
	request(http.MethodPost, "/api/v1/local-projects", map[string]any{"name": "公开工作树实际模型许可", "root": root, "space_id": space.Space.ID}, http.StatusOK, &registered)
	project := registered.Project
	if project.ID == "" || project.Root != root || project.SpaceID != space.Space.ID || project.Settings.ExternalModelCLI != "" {
		t.Fatal("实际注册改变了公开项目身份或默认模型许可")
	}
	expectCode := func(err error, code apierrors.Code) {
		t.Helper()
		var service *apierrors.ServiceError
		if !errors.As(err, &service) || service.Code != code || service.RequiredAction == "" {
			t.Fatal("实际前置拒绝没有保留准确错误和后续操作", code, err)
		}
	}
	factory := distillers.NewSelectedTextFactory(bridge, 30*time.Minute)
	_, err = factory.Resolve(ctx, human, project.ID, "codex")
	expectCode(err, apierrors.ApprovalRequired)
	configure := func(cli string) {
		settings := project.Settings
		settings.ExternalModelCLI = cli
		request(http.MethodPut, "/api/v1/local-projects/"+project.ID+"/settings", map[string]any{"expected_revision": project.Settings.Revision, "settings": settings}, http.StatusOK, &registered)
		if registered.Project.Settings.Revision != project.Settings.Revision+1 || registered.Project.Settings.ExternalModelCLI != cli || registered.Project.Settings.AllowDirectory || registered.Project.Settings.AllowAgentControl {
			t.Fatal("实际 HTTP 设置没有保留当前模型许可和其他权限边界")
		}
		project = registered.Project
	}
	configure("codex")
	distiller, err := factory.Resolve(ctx, human, project.ID, "codex")
	if err != nil {
		t.Fatal("明确许可后的实际项目不能解析选定文本处理器", err)
	}
	scoped, ok := distiller.(interface {
		ProjectSpaceID(context.Context) (string, error)
	})
	if !ok {
		t.Fatal("实际工厂没有返回项目范围接口")
	}
	actualSpace, err := scoped.ProjectSpaceID(ctx)
	if err != nil || actualSpace != space.Space.ID {
		t.Fatal("生产 bridge 没有返回实际项目空间", err)
	}
	_, err = distiller.ConfigurationID(ctx)
	expectCode(err, apierrors.EvidenceMissing)
	// 原始公开正文仅送同一生产业务服务端口，HTTP 请求只处理身份与许可。
	imported, err := core.ImportMaterial(ctx, human, attention.ImportMaterialCommand{
		CommandMeta: attention.CommandMeta{SchemaVersion: 1, RequestID: uuid.NewString(), IdempotencyKey: uuid.NewString(), ExpectedVersion: 1},
		Kind:        "paper", Adapter: "paper_snapshot", SourceLocator: paper.SourceURL, ExportText: string(raw),
	})
	if err != nil {
		t.Fatal("实际公开快照不能通过生产业务端口导入", err)
	}
	if processed, err := core.ProcessNextJob(ctx); err != nil || !processed {
		t.Fatal("实际公开快照导入作业没有执行", err)
	}
	job, err := core.GetJob(ctx, human, imported.JobID)
	if err != nil || job.Status != "succeeded" || job.MaterialID == nil {
		t.Fatal("实际公开快照没有发布资料版本", err)
	}
	material, err := core.GetMaterial(ctx, human, *job.MaterialID)
	if err != nil || len(material.Revisions) != 1 || material.Material.CurrentRevision != material.Revisions[0].Revision || material.Revisions[0].SourceKey != paper.SourceKey || material.Material.SourceLocator != paper.SourceURL {
		t.Fatal("实际资料没有保留固定公开来源", err)
	}
	ref := attention.SourceRef{MaterialID: material.Material.ID, Revision: material.Material.CurrentRevision, Locator: paper.SourceURL}
	request(http.MethodPost, "/api/v1/project-spaces/"+space.Space.ID+"/references", map[string]any{"material_id": ref.MaterialID, "revision": ref.Revision}, http.StatusOK, &space)
	source, err := core.ReadProjectReference(ctx, human, project.SpaceID, ref, false)
	if err != nil {
		t.Fatal(err)
	}
	stored, err := store.Read(ctx, material.Revisions[0].ObjectRef)
	if err != nil {
		t.Fatal(err)
	}
	digest := sha256.Sum256(stored)
	if source.Ref != ref || source.SourceKey != paper.SourceKey || source.Provenance.Mode != "browser_snapshot_fulltext" || source.Text != string(stored) || !strings.Contains(source.Text, paper.Text) || source.ContentDigest != hex.EncodeToString(digest[:]) {
		t.Fatal("生产固定引用没有返回完整公开论文及原版本摘要")
	}
	objectsBefore, err := os.ReadDir(objectRoot)
	if err != nil {
		t.Fatal(err)
	}
	// 作业及操作标识仅为明确调用参数，不声明已经产生模型作业或回执。
	input := attention.DistillationInput{JobID: uuid.NewString(), OperationID: uuid.NewString(), Stage: "summary", Inputs: []attention.SourceSnapshot{source}}
	_, err = distiller.Distill(ctx, input)
	expectCode(err, apierrors.EvidenceMissing)
	configure("")
	_, err = distiller.Distill(ctx, input)
	expectCode(err, apierrors.ApprovalRequired)
	_, err = scoped.ProjectSpaceID(ctx)
	expectCode(err, apierrors.ApprovalRequired)
	configure("codex")
	// 取消来自真实调用者上下文，处理器内部没有取消或替代实现。
	cancelled, cancelCaller := context.WithCancel(ctx)
	cancelCaller()
	_, err = factory.Resolve(cancelled, human, project.ID, "codex")
	if !errors.Is(err, context.Canceled) {
		t.Fatal("实际 Resolve 丢失调用者取消", err)
	}
	_, err = distiller.Distill(cancelled, input)
	if !errors.Is(err, context.Canceled) {
		t.Fatal("实际 Distill 丢失调用者取消", err)
	}
	assertNoNative := func() {
		t.Helper()
		current, err := registry.SnapshotNative(ctx)
		if err != nil || !reflect.DeepEqual(current, baseline) {
			t.Fatal("前置失败改变了实际未探测 Registry", err)
		}
		entries, err := os.ReadDir(nativeRoot)
		if err != nil || len(entries) != 0 {
			t.Fatal("前置失败创建了原生输入或控制器目录", err)
		}
		currentObjects, err := os.ReadDir(objectRoot)
		if err != nil || !reflect.DeepEqual(currentObjects, objectsBefore) {
			t.Fatal("前置失败写入了额外文本对象", err)
		}
	}
	assertNoNative()
	var sessions int
	if err := db.Conn().QueryRowContext(ctx, "SELECT COUNT(*) FROM local_agent_sessions").Scan(&sessions); err != nil || sessions != 0 {
		t.Fatal("前置失败创建了原生会话", err)
	}
	stop()
	if err := db.Close(); err != nil {
		t.Fatal(err)
	}
	closedError := db.Conn().PingContext(ctx)
	if closedError == nil {
		t.Fatal("实际关闭的 SQLite 连接仍可使用")
	}
	_, err = factory.Resolve(ctx, human, project.ID, "codex")
	if !errors.Is(err, closedError) {
		t.Fatal("实际关闭 SQLite 后 Resolve 丢失原始生命周期错误", err)
	}
	_, err = distiller.Distill(ctx, input)
	if !errors.Is(err, closedError) {
		t.Fatal("实际关闭 SQLite 后 Distill 丢失原始生命周期错误", err)
	}
	assertNoNative()
	t.Logf("实际组装前置已验证：固定公开论文正文=%d UTF-8 bytes；默认无许可、真实空间、未探测配置、撤回、调用者取消及真实 SQLite 关闭；Native/模型调用0；联合容量、超限真实来源、推荐及非法模型响应 NOT_RUN", len(paper.Text))
}
