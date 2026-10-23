package httpapi

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/cookiejar"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/songconmaisaix31-design/Astrocyte/internal/adapters/agents"
	"github.com/songconmaisaix31-design/Astrocyte/internal/adapters/sqlite"
	attention "github.com/songconmaisaix31-design/Astrocyte/internal/attention/app"
	swarm "github.com/songconmaisaix31-design/Astrocyte/internal/swarm/app"
	swarmdomain "github.com/songconmaisaix31-design/Astrocyte/internal/swarm/domain"
	workspace "github.com/songconmaisaix31-design/Astrocyte/internal/workspace/app"
	"github.com/songconmaisaix31-design/Astrocyte/internal/workspace/domain"
	"github.com/songconmaisaix31-design/Astrocyte/migrations"
)

type dynamicDAGActual struct {
	db       *sqlite.DB
	service  *swarm.Service
	projects workspace.LocalProjects
	project  domain.LocalProject
	mission  swarmdomain.Mission
	server   *httptest.Server
	client   *http.Client
	csrf     string
}

func dynamicDAGActualService(t *testing.T, idleSeconds int, lifetime time.Duration) *dynamicDAGActual {
	t.Helper()
	ctx := context.Background()
	root, err := filepath.Abs(filepath.Join("..", "..", ".."))
	if err != nil {
		t.Fatal(err)
	}
	root, err = agents.CanonicalRoot(root)
	if err != nil {
		t.Fatal(err)
	}
	base := filepath.Join(root, ".verification", "dynamic-dag-http")
	if err = os.MkdirAll(base, 0700); err != nil {
		t.Fatal(err)
	}
	dir, err := os.MkdirTemp(base, "actual-")
	if err != nil {
		t.Fatal(err)
	}
	db, err := sqlite.Open(filepath.Join(dir, "state.sqlite"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := db.Close(); err != nil {
			t.Error(err)
		}
	})
	if err = db.RunMigrations(ctx, migrations.FS); err != nil {
		t.Fatal(err)
	}
	human := domain.Caller{Kind: "human", ID: "local-human"}
	attentionService := attention.NewAttentionService(sqlite.NewAttentionRepository(db), nil, nil, attention.ServiceOptions{ProjectSpaces: httpAttentionSpaces(db)})
	space, err := attentionService.CreateProjectSpace(ctx, attention.Principal{Kind: human.Kind, ID: human.ID}, attention.ProjectSpaceCommand{CommandMeta: attention.CommandMeta{SchemaVersion: 1, RequestID: uuid.NewString(), IdempotencyKey: uuid.NewString(), ExpectedVersion: 1}, Title: "公开Astrocyte任务依赖验证"})
	if err != nil {
		t.Fatal(err)
	}
	projects := workspace.NewLocalProjectService(db, nil, agents.ProjectFiles{}, nil)
	project, err := projects.RegisterProject(ctx, human, domain.RegisterProjectCommand{Name: "Astrocyte动态DAG验证", Root: root, SpaceID: space.Space.ID})
	if err != nil {
		t.Fatal(err)
	}
	settings := project.Settings
	settings.AllowDirectory = true
	settings.AllowedSubdirs = []string{"."}
	settings.AllowedActions = []string{"create_mission", "manage_work_items", "read_swarm", "recover_mission"}
	project, err = projects.SetProjectSettings(ctx, human, project.ID, domain.SettingsCommand{ExpectedRevision: project.Settings.Revision, Settings: settings})
	if err != nil {
		t.Fatal(err)
	}
	s2 := workspace.NewS2Service(sqlite.NewWorkspaceRepository(db), agents.ProjectFiles{}, nil)
	goal, err := s2.PutGoal(ctx, human, "", workspace.GoalCommand{WorkspaceCommandMeta: swarmWorkspaceMeta(0), Goal: domain.Goal{ProjectID: project.ID, Title: "核查真实任务依赖、权限和停止历史", Status: "active", SuccessCriteria: []string{"每个创建事件可在实际SQLite回查"}}})
	if err != nil {
		t.Fatal(err)
	}
	calls := int64(3)
	deadline := time.Now().Add(lifetime)
	content := domain.ProposalContent{Title: "人工创建公开项目核查任务", Goal: domain.GoalReference{ID: goal.ID, Revision: goal.Revision}, Scope: domain.ExecutionScope{ProjectID: project.ID, SettingsRevision: project.Settings.Revision, Roots: []string{root}, WritePaths: []string{"internal/swarm"}, Actions: settings.AllowedActions, MaxConcurrency: 1}, Budget: domain.ResourceBudget{Calls: &calls}, Stops: domain.StopConditions{Deadline: deadline, Conditions: []string{"按批准的无进展时间停止"}, MaxWorkItems: 4, MaxAttempts: 2, MaxNoProgressRounds: 2, NoProgressSeconds: idleSeconds}, Deliverables: []string{"真实HTTP与SQLite核查记录"}, SuccessCriteria: []string{"保留人工任务来源并拒绝越权"}}
	proposal, err := s2.CreateProposal(ctx, human, workspace.CreateProposalCommand{WorkspaceCommandMeta: swarmWorkspaceMeta(0), Content: content})
	if err != nil {
		t.Fatal(err)
	}
	proposal, err = s2.TransitionProposal(ctx, human, proposal.ID, swarmWorkspaceMeta(proposal.Version), domain.ProposalInReview)
	if err != nil {
		t.Fatal(err)
	}
	approval, err := s2.ApproveProposal(ctx, human, proposal.ID, workspace.ApproveProposalCommand{WorkspaceCommandMeta: swarmWorkspaceMeta(proposal.Version), ExpectedRevision: proposal.Revision, ExpiresAt: deadline})
	if err != nil {
		t.Fatal(err)
	}
	repo, err := sqlite.NewSwarmRepository(db)
	if err != nil {
		t.Fatal(err)
	}
	bridge, err := sqlite.NewSwarmApprovalBridge(db, s2)
	if err != nil {
		t.Fatal(err)
	}
	service, err := swarm.NewSwarmLedger(repo, bridge)
	if err != nil {
		t.Fatal(err)
	}
	mission, err := service.CreateMission(ctx, swarmdomain.Principal{Kind: human.Kind, ID: human.ID}, swarm.CreateMissionCommand{Meta: swarmHTTPMeta(0), GrantID: approval.ID, Epoch: approval.Epoch, ProposalRevision: approval.ProposalRevision, ProjectID: project.ID, Root: root})
	if err != nil {
		t.Fatal(err)
	}
	mux := http.NewServeMux()
	h := &handler{}
	h.registerSwarmS3(mux, service)
	guard := newSessionGuard(Config{Services: Services{LocalProjects: projects}})
	mux.HandleFunc("GET /api/v1/auth/session", guard.bootstrap)
	server := httptest.NewServer(requestIDMiddleware(guard.middleware(mux)))
	t.Cleanup(server.Close)
	jar, err := cookiejar.New(nil)
	if err != nil {
		t.Fatal(err)
	}
	client := server.Client()
	client.Jar = jar
	response, err := client.Get(server.URL + "/api/v1/auth/session")
	if err != nil {
		t.Fatal(err)
	}
	defer response.Body.Close()
	var session struct {
		CSRF string `json:"csrf_token"`
	}
	if err = json.NewDecoder(response.Body).Decode(&session); err != nil {
		t.Fatal(err)
	}
	if response.StatusCode != 200 || session.CSRF == "" {
		t.Fatal("actual human session bootstrap failed")
	}
	t.Logf("actual database=%s", filepath.Join(dir, "state.sqlite"))
	return &dynamicDAGActual{db: db, service: service, projects: projects, project: project, mission: mission, server: server, client: client, csrf: session.CSRF}
}

func (env *dynamicDAGActual) request(t *testing.T, method, path string, body any, token string) (int, []byte) {
	t.Helper()
	var raw []byte
	var err error
	if body != nil {
		raw, err = json.Marshal(body)
		if err != nil {
			t.Fatal(err)
		}
	}
	request, err := http.NewRequest(method, env.server.URL+"/api/v1"+path, bytes.NewReader(raw))
	if err != nil {
		t.Fatal(err)
	}
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Idempotency-Key", uuid.NewString())
	request.Header.Set("X-CSRF-Token", env.csrf)
	if token != "" {
		request.Header.Set("Authorization", "Bearer "+token)
	}
	response, err := env.client.Do(request)
	if err != nil {
		t.Fatal(err)
	}
	defer response.Body.Close()
	raw, err = io.ReadAll(response.Body)
	if err != nil {
		t.Fatal(err)
	}
	return response.StatusCode, raw
}

func (env *dynamicDAGActual) read(t *testing.T) swarm.MissionDAG {
	t.Helper()
	status, raw := env.request(t, "GET", "/missions/"+env.mission.ID+"/dag", nil, "")
	if status != 200 {
		t.Fatalf("DAG HTTP status=%d: %s", status, raw)
	}
	var result swarm.MissionDAG
	if err := json.Unmarshal(raw, &result); err != nil {
		t.Fatal(err)
	}
	return result
}

func TestDynamicDAGActualHTTPGrowthProvenanceAndScopedRead(t *testing.T) {
	env := dynamicDAGActualService(t, 300, time.Hour)
	ctx := context.Background()
	human := domain.Caller{Kind: "human", ID: "local-human"}
	if result := env.read(t); len(result.Nodes) != 0 || len(result.Edges) != 0 || len(result.Workers) != 0 {
		t.Fatal("empty actual mission was not empty")
	}
	var previous string
	for _, title := range []string{"读取公开SPEC与真实任务边界", "核查依赖来源并记录实际验证结果"} {
		dependencies := []string{}
		if previous != "" {
			dependencies = append(dependencies, previous)
		}
		command := swarm.AddWorkItemCommand{Meta: swarmHTTPMeta(env.mission.Version), Title: title, Inputs: env.mission.Grant.Inputs, Dependencies: dependencies, WritePaths: []string{"internal/swarm"}}
		status, raw := env.request(t, "POST", "/missions/"+env.mission.ID+"/work-items", command, "")
		if status != 200 {
			t.Fatalf("create status=%d: %s", status, raw)
		}
		var result struct {
			WorkItem swarmdomain.WorkItem `json:"work_item"`
		}
		if err := json.Unmarshal(raw, &result); err != nil {
			t.Fatal(err)
		}
		previous = result.WorkItem.ID
		env.mission.Version++
	}
	result := env.read(t)
	if len(result.Nodes) != 2 || len(result.Edges) != 1 || len(result.Workers) != 0 {
		t.Fatal("actual creation did not produce exactly two nodes, one dependency and no executions")
	}
	for _, node := range result.Nodes {
		if node.WorkItem.CreatorKind != "human" || node.WorkItem.CreatedBy != "local-human" || node.CreationEvent == nil {
			t.Fatal("actual human creation provenance missing")
		}
		var aggregate string
		if err := env.db.Conn().QueryRow("SELECT aggregate_id FROM foundation_outbox WHERE event_id=? AND event_type='swarm.work_item_created'", node.CreationEvent.ID).Scan(&aggregate); err != nil || aggregate != node.WorkItem.ID {
			t.Fatalf("creation event SQL lookup failed: %v", err)
		}
	}
	if result.Edges[0].EventID == "" {
		t.Fatal("dependency source missing")
	}
	missing := swarm.AddWorkItemCommand{Meta: swarmHTTPMeta(env.mission.Version), Title: "检查缺少端点必须拒绝", Inputs: env.mission.Grant.Inputs, Dependencies: []string{uuid.NewString()}, WritePaths: []string{"internal/swarm"}}
	if status, _ := env.request(t, "POST", "/missions/"+env.mission.ID+"/work-items", missing, ""); status != 404 {
		t.Fatalf("missing dependency status=%d", status)
	}
	if unchanged := env.read(t); len(unchanged.Nodes) != 2 || unchanged.MissionVersion != result.MissionVersion {
		t.Fatal("rejected dependency changed the actual graph")
	}
	for _, actions := range [][]string{{"read_context"}, {"read_context", "read_swarm"}} {
		if _, err := env.projects.GrantProjectAgent(ctx, human, env.project.ID, domain.ProjectGrant{AgentID: "codex", Actions: actions}); err != nil {
			t.Fatal(err)
		}
		token, err := env.projects.IssueProjectAgentToken(ctx, human, env.project.ID, "codex")
		if err != nil {
			t.Fatal(err)
		}
		status, _ := env.request(t, "GET", "/missions/"+env.mission.ID+"/dag", nil, token.Token)
		if len(actions) == 1 && status != 403 || len(actions) == 2 && status != 200 {
			t.Fatalf("read_swarm actions=%v status=%d", actions, status)
		}
	}
	other, err := env.projects.RegisterProject(ctx, human, domain.RegisterProjectCommand{Name: "公开项目的独立前端读取边界", Root: filepath.Join(env.project.Root, "web"), SpaceID: env.project.SpaceID})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := env.projects.GrantProjectAgent(ctx, human, other.ID, domain.ProjectGrant{AgentID: "codex", Actions: []string{"read_context", "read_swarm"}}); err != nil {
		t.Fatal(err)
	}
	token, err := env.projects.IssueProjectAgentToken(ctx, human, other.ID, "codex")
	if err != nil {
		t.Fatal(err)
	}
	if status, _ := env.request(t, "GET", "/missions/"+env.mission.ID+"/dag", nil, token.Token); status != 403 {
		t.Fatalf("cross-project DAG status=%d", status)
	}
	settings := env.project.Settings
	settings.AllowedActions = []string{"read_context"}
	if _, err := env.projects.SetProjectSettings(ctx, human, env.project.ID, domain.SettingsCommand{ExpectedRevision: settings.Revision, Settings: settings}); err != nil {
		t.Fatal(err)
	}
	if status, _ := env.request(t, "GET", "/missions/"+env.mission.ID+"/dag", nil, ""); status != 403 {
		t.Fatalf("removed current read_swarm status=%d", status)
	}
}

func TestDynamicDAGActualIdleStopRetainsHistoryWithoutWorkers(t *testing.T) {
	env := dynamicDAGActualService(t, 1, time.Hour)
	time.Sleep(time.Until(env.mission.LastProgressAt.Add(1100 * time.Millisecond)))
	if err := env.service.EnforceStopConditions(context.Background()); err != nil {
		t.Fatal(err)
	}
	result := env.read(t)
	if result.MissionState != "blocked" || result.Reason != "no_progress_time_limit" || len(result.Workers) != 0 {
		t.Fatal("actual idle stop changed worker facts or missed its stop condition")
	}
	before := len(result.Events)
	if err := env.service.EnforceStopConditions(context.Background()); err != nil {
		t.Fatal(err)
	}
	if after := env.read(t); len(after.Events) != before {
		t.Fatal("blocked mission was automatically stopped again")
	}
	var count int
	if err := env.db.Conn().QueryRow("SELECT count(*) FROM foundation_outbox WHERE aggregate_id=? AND event_type='swarm.mission_stop_requested' AND json_extract(payload,'$.source')='system:swarm-lifecycle'", env.mission.ID).Scan(&count); err != nil || count != 1 {
		t.Fatalf("actual system stop event count=%d error=%v", count, err)
	}
}

func TestDynamicDAGActualExpiredApprovalRemainsHumanReadable(t *testing.T) {
	env := dynamicDAGActualService(t, 300, 2*time.Second)
	time.Sleep(time.Until(env.mission.Grant.ExpiresAt.Add(100 * time.Millisecond)))
	if err := env.service.EnforceStopConditions(context.Background()); err != nil {
		t.Fatal(err)
	}
	result := env.read(t)
	if result.MissionState != "blocked" || result.Reason != "deadline_reached" || len(result.Workers) != 0 {
		t.Fatal("expired actual approval history was lost or fabricated worker exit")
	}
}
