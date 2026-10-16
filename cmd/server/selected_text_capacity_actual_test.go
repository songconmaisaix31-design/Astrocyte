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
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"
	"testing"
	"time"
	"unicode/utf8"

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

func selectedCapacityPublicBytes(t *testing.T, variable, digest string, expectedBytes int) []byte {
	t.Helper()
	path := os.Getenv(variable)
	if !filepath.IsAbs(path) || strings.TrimSpace(path) != path || filepath.Base(path) != digest {
		t.Fatal("必须明确选择获准公开对象的原绝对路径", variable)
	}
	read := func() ([]byte, os.FileInfo) {
		t.Helper()
		volumeRoot := filepath.VolumeName(path) + string(filepath.Separator)
		relative, err := filepath.Rel(volumeRoot, path)
		if err != nil {
			t.Fatal(err)
		}
		// 沿用生产路径保护，逐分量拒绝源文件及祖先重定向；不枚举来源目录。
		physical, err := agents.ValidateProjectPath(volumeRoot, relative)
		if err != nil {
			t.Fatal(err)
		}
		file, err := os.Open(physical)
		if err != nil {
			t.Fatal(err)
		}
		info, statErr := file.Stat()
		data, readErr := io.ReadAll(file)
		if err := errors.Join(statErr, readErr, file.Close()); err != nil {
			t.Fatal(err)
		}
		sum := sha256.Sum256(data)
		if !info.Mode().IsRegular() || info.Size() != int64(expectedBytes) || len(data) != expectedBytes || hex.EncodeToString(sum[:]) != digest {
			t.Fatal("获准公开对象的普通文件、完整字节或摘要不符", variable)
		}
		return data, info
	}
	original, before := read()
	t.Cleanup(func() {
		current, after := read()
		if !os.SameFile(before, after) || before.Size() != after.Size() || !before.ModTime().Equal(after.ModTime()) || !bytes.Equal(original, current) {
			t.Error("获准公开对象的原文件身份、大小、修改时间或完整字节改变", variable)
		}
	})
	return original
}

func TestSelectedTextCapacityActualManualSourcesPreflightWithoutNative(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	if !filepath.IsAbs(os.Getenv("ASTROCYTE_TEST_OUTPUT_ROOT")) {
		t.Fatal("实际容量前置必须明确选择自有输出绝对目录")
	}
	definitions := []struct {
		variable, digest, locator, kind string
		length                          int
	}{
		{"ASTROCYTE_TEST_CAPACITY_PAPER_OBJECT", "05acdc763e444fa7b9762df6dc2c879867d6a2e38d42884dcf3b80e07cf89d64", "https://arxiv.org/html/2504.16054v1", "paper", 95314},
		{"ASTROCYTE_TEST_CAPACITY_VIDEO_OBJECT", "030165629c4b38e519ec79e97e52b41e3c470ddfb82df268a86aef361d35f5b0", "https://www.bilibili.com/video/BV1PReT6EEqR/", "video", 35807},
	}
	bodies := make([][]byte, len(definitions))
	total := 0
	for i, source := range definitions {
		bodies[i] = selectedCapacityPublicBytes(t, source.variable, source.digest, source.length)
		total += len(bodies[i])
	}
	if total != 131121 || total <= 128*1024 {
		t.Fatal("两份完整原文的自然总量未超过128KiB")
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
	store, err := objects.New(filepath.Join(directory, "objects"))
	if err != nil {
		t.Fatal(err)
	}
	nativeRoot := filepath.Join(directory, "unused-native")
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
			t.Fatal("实际容量前置 HTTP 服务没有正常退出", err)
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
	request := func(method, route string, command map[string]any, expected int, output any) {
		t.Helper()
		var body []byte
		if command != nil {
			command["schema_version"], command["request_id"] = 1, uuid.NewString()
			if _, ok := command["expected_version"]; !ok {
				command["expected_version"] = 1
			}
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
		if err := errors.Join(readErr, response.Body.Close()); err != nil || response.StatusCode != expected {
			t.Fatal("实际容量前置 HTTP 状态不符", method, route, response.StatusCode, expected, err)
		}
		if err := json.Unmarshal(data, output); err != nil {
			t.Fatal(err)
		}
	}
	request(http.MethodGet, "/api/v1/auth/session", nil, http.StatusOK, &session)
	if session.ID == "" || session.Kind != "human" || session.CSRF == "" {
		t.Fatal("实际 bootstrap 没有完整人类身份")
	}
	human := attention.Principal{ID: session.ID, Kind: session.Kind}
	var space attention.ProjectSpaceResult
	request(http.MethodPost, "/api/v1/project-spaces", map[string]any{"title": "两份公开正文的新人工来源容量前置"}, http.StatusCreated, &space)
	root, err := filepath.Abs("../..")
	if err != nil {
		t.Fatal(err)
	}
	root, err = agents.CanonicalRoot(root)
	if err != nil {
		t.Fatal(err)
	}
	var configured struct {
		Project domain.LocalProject `json:"project"`
	}
	request(http.MethodPost, "/api/v1/local-projects", map[string]any{"name": "实际公开双源前置许可", "root": root, "space_id": space.Space.ID}, http.StatusOK, &configured)
	project := configured.Project
	if project.ID == "" || project.SpaceID != space.Space.ID || project.Root != root || project.Settings.ExternalModelCLI != "" {
		t.Fatal("实际注册没有保留项目身份和默认模型许可")
	}
	settings := project.Settings
	settings.ExternalModelCLI = "codex"
	request(http.MethodPut, "/api/v1/local-projects/"+project.ID+"/settings", map[string]any{"expected_revision": project.Settings.Revision, "settings": settings}, http.StatusOK, &configured)
	if configured.Project.Settings.ExternalModelCLI != "codex" || configured.Project.Settings.AllowDirectory || configured.Project.Settings.AllowAgentControl || len(configured.Project.Settings.AllowedTools) != 0 {
		t.Fatal("实际容量前置配置扩大了其他权限")
	}
	inputs := make([]attention.SourceSnapshot, 0, len(definitions))
	for i, source := range definitions {
		started := time.Now().UTC()
		// 以原URL作本次人工标题，完整原正文新入库；不复造历史ID或出版provenance。
		imported, err := core.ImportMaterial(ctx, human, attention.ImportMaterialCommand{
			CommandMeta: attention.CommandMeta{SchemaVersion: 1, RequestID: uuid.NewString(), IdempotencyKey: uuid.NewString(), ExpectedVersion: 1},
			Kind:        source.kind, Adapter: "manual", SourceLocator: source.locator, Title: source.locator, ExportText: string(bodies[i]),
		})
		if err != nil {
			t.Fatal("获准公开正文不能经生产人工导入端口入库", err)
		}
		if processed, err := core.ProcessNextJob(ctx); err != nil || !processed {
			t.Fatal("实际人工导入作业未执行", err)
		}
		job, err := core.GetJob(ctx, human, imported.JobID)
		if err != nil || job.Status != "succeeded" || job.MaterialID == nil {
			t.Fatal("实际人工导入没有发布资料", err)
		}
		material, err := core.GetMaterial(ctx, human, *job.MaterialID)
		if err != nil || len(material.Revisions) != 1 || material.Material.CurrentRevision != 1 || material.Material.SourceLocator != source.locator {
			t.Fatal("新人工来源的实际版本或URL不符", err)
		}
		revision := material.Revisions[0]
		if revision.ContentDigest != source.digest || revision.Provenance.Processor != "manual" || revision.Provenance.Version != "1" || revision.Provenance.Mode != "manual" || revision.Provenance.Source != source.locator || revision.CreatedAt.Before(started) || revision.CreatedAt.After(time.Now().UTC()) {
			t.Fatal("新人工来源没有保留原摘要、人工provenance或本次时间")
		}
		ref := attention.SourceRef{MaterialID: material.Material.ID, Revision: revision.Revision, Locator: source.locator}
		request(http.MethodPost, "/api/v1/project-spaces/"+project.SpaceID+"/references", map[string]any{"material_id": ref.MaterialID, "revision": ref.Revision, "expected_version": space.Space.Version}, http.StatusOK, &space)
		selected, err := core.ReadProjectReference(ctx, human, project.SpaceID, ref, false)
		if err != nil || selected.Ref != ref || selected.Text != string(bodies[i]) || selected.ContentDigest != source.digest || selected.Provenance != revision.Provenance {
			t.Fatal("实际新固定引用没有保留完整公开原文和人工来源", err)
		}
		inputs = append(inputs, selected)
	}
	distiller, err := distillers.NewSelectedTextFactory(bridge, 30*time.Minute).Resolve(ctx, human, project.ID, "codex")
	if err != nil {
		t.Fatal(err)
	}
	// UUID只标识本次直接调用参数，不声明创建了模型作业或原生回执。
	input := attention.DistillationInput{JobID: uuid.NewString(), OperationID: uuid.NewString(), Stage: "topic", Inputs: inputs}
	var observed bytes.Buffer
	previousLogger := slog.Default()
	t.Cleanup(func() { slog.SetDefault(previousLogger) })
	slog.SetDefault(slog.New(slog.NewJSONHandler(&observed, nil)))
	_, err = distiller.Distill(ctx, input)
	slog.SetDefault(previousLogger)
	var failure *apierrors.ServiceError
	if !errors.As(err, &failure) || failure.Code != apierrors.EvidenceMissing {
		t.Fatal("完整双源未通过生产预检到达实际未探测配置边界", err)
	}
	var observation struct {
		Time      time.Time `json:"time"`
		Level     string    `json:"level"`
		Message   string    `json:"msg"`
		Job       string    `json:"job_id"`
		Operation string    `json:"operation_id"`
		Prompt    int       `json:"prompt_utf8_bytes"`
		Wire      int       `json:"wire_utf8_bytes"`
		Output    int       `json:"output_limit_bytes"`
	}
	decoder := json.NewDecoder(bytes.NewReader(observed.Bytes()))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&observation); err != nil {
		t.Fatal("生产容量安全日志不是实际标准JSON", err)
	}
	if err := decoder.Decode(new(any)); err != io.EOF {
		t.Fatal("本次生产容量前置没有恰好一个观察记录", err)
	}
	if observation.Time.IsZero() || observation.Level != "INFO" || observation.Message != "selected-text request preflight" || observation.Job != input.JobID || observation.Operation != input.OperationID || observation.Prompt <= 128*1024 || observation.Wire <= observation.Prompt || observation.Wire >= 512*1024 || observation.Output != 128*1024 {
		t.Fatal("生产实际安全字段的完整输入或固定输出大小不符")
	}
	current, err := registry.SnapshotNative(ctx)
	if err != nil || !reflect.DeepEqual(current, baseline) {
		t.Fatal("容量前置改变了实际未探测Registry", err)
	}
	entries, err := os.ReadDir(nativeRoot)
	if err != nil || len(entries) != 0 {
		t.Fatal("容量前置创建了原生输入或控制器目录", err)
	}
	for table, expected := range map[string]int{"local_agent_sessions": 0, "attention_jobs": 2, "attention_distillations": 0} {
		var count int
		if err := db.Conn().QueryRowContext(ctx, "SELECT COUNT(*) FROM "+table).Scan(&count); err != nil || count != expected {
			t.Fatal("容量前置增加了模型或原生记录", table, err)
		}
	}
	stop()
	if err := db.Close(); err != nil {
		t.Fatal(err)
	}
	t.Logf("实际新人工双源正文=%d UTF-8 bytes，生产prompt=%d、wire=%d、output=%d；仅容量/配置前置，完整prompt内容保真、模型接受及真实超限来源 NOT_RUN", total, observation.Prompt, observation.Wire, observation.Output)
}

func TestSelectedTextCapacityActualGitDocumentsOversize(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	if !filepath.IsAbs(os.Getenv("ASTROCYTE_TEST_OUTPUT_ROOT")) {
		t.Fatal("实际Git文档超限检查需要明确自有输出绝对目录")
	}
	commit := os.Getenv("ASTROCYTE_TEST_CAPACITY_PUBLIC_COMMIT")
	decoded, err := hex.DecodeString(commit)
	if err != nil || len(decoded) != 20 || len(commit) != 40 {
		t.Fatal("必须明确选择准确已push的完整公开Git commit")
	}
	root, err := filepath.Abs("../..")
	if err != nil {
		t.Fatal(err)
	}
	root, err = agents.CanonicalRoot(root)
	if err != nil {
		t.Fatal(err)
	}
	gitPath, err := exec.LookPath("git")
	if err != nil {
		t.Fatal(err)
	}
	gitRead := func(input []byte, args ...string) []byte {
		t.Helper()
		command := exec.CommandContext(ctx, gitPath, args...)
		command.Dir, command.Stdin = root, bytes.NewReader(input)
		output, err := command.Output()
		if err != nil {
			var failure *exec.ExitError
			if errors.As(err, &failure) {
				t.Fatalf("实际Git只读命令失败：%v；stderr=%s", err, failure.Stderr)
			}
			t.Fatal(err)
		}
		return output
	}
	if strings.TrimSpace(string(gitRead(nil, "rev-parse", "--verify", commit+"^{commit}"))) != commit || strings.TrimSpace(string(gitRead(nil, "cat-file", "-t", commit))) != "commit" {
		t.Fatal("固定公开Git commit身份不符")
	}
	documents := []struct {
		path string
		size int
	}{
		{"STATUS.md", 209211},
		{"docs/standards/backend.md", 230066},
		{"HANDOFF.md", 78620},
		{"docs/SPEC.md", 60277},
	}
	bodies := make([][]byte, len(documents))
	blobs := make([]string, len(documents))
	total := 0
	for i, document := range documents {
		blobs[i] = strings.TrimSpace(string(gitRead(nil, "rev-parse", "--verify", commit+":"+document.path)))
		if strings.TrimSpace(string(gitRead(nil, "cat-file", "-t", blobs[i]))) != "blob" {
			t.Fatal("固定文档没有实际Git blob身份", document.path)
		}
		size, err := strconv.Atoi(strings.TrimSpace(string(gitRead(nil, "cat-file", "-s", blobs[i]))))
		if err != nil || size != document.size {
			t.Fatal("固定公开文档的真实Git大小不符", document.path, err)
		}
		bodies[i] = gitRead(nil, "cat-file", "blob", blobs[i])
		if len(bodies[i]) != size || !utf8.Valid(bodies[i]) || strings.TrimSpace(string(gitRead(bodies[i], "hash-object", "--no-filters", "--stdin"))) != blobs[i] {
			t.Fatal("固定公开文档的完整UTF-8原bytes或Git blob不符", document.path)
		}
		total += size
	}
	if total != 578174 || total <= 512*1024 {
		t.Fatal("四份不同真实Git文档的自然总量未超过512KiB")
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
	store, err := objects.New(filepath.Join(directory, "objects"))
	if err != nil {
		t.Fatal(err)
	}
	nativeRoot := filepath.Join(directory, "unused-native")
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
	server := httpapi.NewServer(httpapi.Config{
		Logger:   slog.New(slog.NewTextHandler(io.Discard, nil)),
		Services: httpapi.Services{LocalProjects: projects, Attention: core},
	})
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
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
			t.Fatal("实际文档超限HTTP服务没有正常退出", err)
		}
		stopped = true
	}
	t.Cleanup(stop)
	client := &http.Client{Timeout: 5 * time.Second}
	t.Cleanup(client.CloseIdleConnections)
	response, err := client.Get("http://" + listener.Addr().String() + "/api/v1/auth/session")
	if err != nil {
		t.Fatal(err)
	}
	var session struct {
		ID   string `json:"actor_id"`
		Kind string `json:"actor_kind"`
		CSRF string `json:"csrf_token"`
	}
	decodeErr := json.NewDecoder(response.Body).Decode(&session)
	if err := errors.Join(decodeErr, response.Body.Close()); err != nil || response.StatusCode != http.StatusOK || session.ID == "" || session.Kind != "human" || session.CSRF == "" || len(response.Cookies()) == 0 {
		t.Fatal("实际bootstrap没有产生完整人类身份", err)
	}
	human := attention.Principal{ID: session.ID, Kind: session.Kind}
	caller := domain.Caller{ID: session.ID, Kind: session.Kind}
	meta := func() attention.CommandMeta {
		return attention.CommandMeta{SchemaVersion: 1, RequestID: uuid.NewString(), IdempotencyKey: uuid.NewString(), ExpectedVersion: 1}
	}
	space, err := core.CreateProjectSpace(ctx, human, attention.ProjectSpaceCommand{CommandMeta: meta(), Title: "固定公开Git文档实际超限前置"})
	if err != nil {
		t.Fatal(err)
	}
	project, err := projects.RegisterProject(ctx, caller, domain.RegisterProjectCommand{Name: "实际公开文档输入许可", Root: root, SpaceID: space.Space.ID})
	if err != nil {
		t.Fatal(err)
	}
	settings := project.Settings
	settings.ExternalModelCLI = "codex"
	project, err = projects.SetProjectSettings(ctx, caller, project.ID, domain.SettingsCommand{ExpectedRevision: project.Settings.Revision, Settings: settings})
	if err != nil || project.Settings.ExternalModelCLI != "codex" || project.Settings.AllowDirectory || project.Settings.AllowAgentControl || len(project.Settings.AllowedTools) != 0 {
		t.Fatal("真实公开项目没有保留当前模型和工具范围", err)
	}
	inputs := make([]attention.SourceSnapshot, 0, len(documents))
	for i, document := range documents {
		locator, err := url.JoinPath("https://github.com/songconmaisaix31-design/Astrocyte/blob", commit, document.path)
		if err != nil {
			t.Fatal(err)
		}
		started := time.Now().UTC()
		// 四份各一次的真实公开Git文档，file分类及manual来源不声明论文或视频。
		imported, err := core.ImportMaterial(ctx, human, attention.ImportMaterialCommand{CommandMeta: meta(), Kind: "file", Adapter: "manual", Title: document.path, SourceLocator: locator, ExportText: string(bodies[i])})
		if err != nil {
			t.Fatal("真实Git文档不能经生产manual端口导入", err)
		}
		if processed, err := core.ProcessNextJob(ctx); err != nil || !processed {
			t.Fatal("实际Git文档导入作业未执行", err)
		}
		job, err := core.GetJob(ctx, human, imported.JobID)
		if err != nil || job.Status != "succeeded" || job.MaterialID == nil {
			t.Fatal("实际Git文档没有生成新资料", err)
		}
		material, err := core.GetMaterial(ctx, human, *job.MaterialID)
		if err != nil || material.Material.Kind != "file" || material.Material.SourceLocator != locator || material.Material.CurrentRevision != 1 || len(material.Revisions) != 1 {
			t.Fatal("新Git文档的实际分类/URL/版本不符", err)
		}
		revision := material.Revisions[0]
		sum := sha256.Sum256(bodies[i])
		if revision.ContentDigest != hex.EncodeToString(sum[:]) || revision.CreatedAt.Before(started) || revision.CreatedAt.After(time.Now().UTC()) || revision.Provenance.Processor != "manual" || revision.Provenance.Mode != "manual" || revision.Provenance.Source != locator {
			t.Fatal("新Git文档没有保留原bytes摘要和本次人工provenance")
		}
		ref := attention.SourceRef{MaterialID: material.Material.ID, Revision: 1, Locator: locator}
		command := meta()
		command.ExpectedVersion = space.Space.Version
		space, err = core.ReferenceMaterial(ctx, human, project.SpaceID, attention.ReferenceMaterialCommand{CommandMeta: command, MaterialID: ref.MaterialID, Revision: ref.Revision})
		if err != nil {
			t.Fatal(err)
		}
		selected, err := core.ReadProjectReference(ctx, human, project.SpaceID, ref, false)
		if err != nil || selected.Ref != ref || selected.Text != string(bodies[i]) || selected.ContentDigest != revision.ContentDigest || selected.Provenance != revision.Provenance {
			t.Fatal("实际新固定引用没有保留完整Git文档", err)
		}
		inputs = append(inputs, selected)
	}
	bridge := &selectedTextBridge{projects: projects, processor: registry}
	distiller, err := distillers.NewSelectedTextFactory(bridge, 30*time.Minute).Resolve(ctx, human, project.ID, "codex")
	if err != nil {
		t.Fatal(err)
	}
	_, err = distiller.Distill(ctx, attention.DistillationInput{JobID: uuid.NewString(), OperationID: uuid.NewString(), Stage: "topic", Inputs: inputs})
	var failure *apierrors.ServiceError
	if !errors.As(err, &failure) || failure.Code != apierrors.ValidationFailed || failure.RequiredAction != "review_selected_text_size" {
		t.Fatal("四份完整真实文档未被生产输入超限预检准确拒绝", err)
	}
	current, err := registry.SnapshotNative(ctx)
	if err != nil || !reflect.DeepEqual(current, baseline) {
		t.Fatal("实际超限预检改变了未探测Registry", err)
	}
	entries, err := os.ReadDir(nativeRoot)
	if err != nil || len(entries) != 0 {
		t.Fatal("实际超限拒绝创建了原生输入或控制器目录", err)
	}
	for table, expected := range map[string]int{"attention_jobs": 4, "attention_distillations": 0, "local_agent_sessions": 0} {
		var count int
		if err := db.Conn().QueryRowContext(ctx, "SELECT COUNT(*) FROM "+table).Scan(&count); err != nil || count != expected {
			t.Fatal("真实文档超限检查增加了模型或原生记录", table, err)
		}
	}
	stop()
	if err := db.Close(); err != nil {
		t.Fatal(err)
	}
	for i, document := range documents {
		if strings.TrimSpace(string(gitRead(nil, "rev-parse", "--verify", commit+":"+document.path))) != blobs[i] || !bytes.Equal(gitRead(nil, "cat-file", "blob", blobs[i]), bodies[i]) {
			t.Fatal("固定commit的原Git blob或全bytes改变", document.path)
		}
	}
	t.Logf("准确公开Git commit=%s，四份不同真实文档各一次，自然正文=%d bytes；实际新manual来源/Ref完整，生产ValidationFailed+review_selected_text_size，Native/模型0；未验证recommendation或模型响应", commit, total)
}
