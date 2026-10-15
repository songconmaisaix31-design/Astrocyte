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

func swarmWorkspaceMeta(version int) workspace.WorkspaceCommandMeta {
	return workspace.WorkspaceCommandMeta{SchemaVersion: 1, RequestID: uuid.NewString(), IdempotencyKey: uuid.NewString(), ExpectedVersion: version}
}
func swarmHTTPMeta(version int) swarmdomain.Meta {
	return swarmdomain.Meta{SchemaVersion: 1, RequestID: uuid.NewString(), IdempotencyKey: uuid.NewString(), ExpectedVersion: version}
}

func TestSwarmActualHTTPHumanAuthorizationTimelineAndRevocation(t *testing.T) {
	ctx := context.Background()
	root, err := filepath.Abs(filepath.Join("..", "..", ".."))
	if err != nil {
		t.Fatal(err)
	}
	root, err = agents.CanonicalRoot(root)
	if err != nil {
		t.Fatal(err)
	}
	base := filepath.Join(root, ".verification", "s3-http")
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
	defer db.Close()
	if err = db.RunMigrations(ctx, migrations.FS); err != nil {
		t.Fatal(err)
	}
	human := domain.Caller{Kind: "human", ID: "local-human"}
	attentionService := attention.NewAttentionService(sqlite.NewAttentionRepository(db), nil, nil, attention.ServiceOptions{})
	space, err := attentionService.CreateProjectSpace(ctx, attention.Principal{Kind: human.Kind, ID: human.ID}, attention.ProjectSpaceCommand{CommandMeta: attention.CommandMeta{SchemaVersion: 1, RequestID: uuid.NewString(), IdempotencyKey: uuid.NewString(), ExpectedVersion: 1}, Title: "Public Swarm HTTP verification"})
	if err != nil {
		t.Fatal(err)
	}
	projects := workspace.NewLocalProjectService(db, nil, agents.ProjectFiles{}, nil)
	project, err := projects.RegisterProject(ctx, human, domain.RegisterProjectCommand{Name: "Astrocyte Swarm HTTP", Root: root, SpaceID: space.Space.ID})
	if err != nil {
		t.Fatal(err)
	}
	settings := project.Settings
	settings.AllowDirectory = true
	settings.AllowedSubdirs = []string{"."}
	settings.AllowedActions = []string{"create_mission", "manage_work_items", "read_swarm", "pause_mission", "cancel_mission", "recover_mission", "reject_artifact"}
	project, err = projects.SetProjectSettings(ctx, human, project.ID, domain.SettingsCommand{ExpectedRevision: project.Settings.Revision, Settings: settings})
	if err != nil {
		t.Fatal(err)
	}
	s2 := workspace.NewS2Service(sqlite.NewWorkspaceRepository(db), agents.ProjectFiles{}, nil)
	goal, err := s2.PutGoal(ctx, human, "", workspace.GoalCommand{WorkspaceCommandMeta: swarmWorkspaceMeta(0), Goal: domain.Goal{ProjectID: project.ID, Title: "Verify actual Swarm HTTP permissions", Status: "active", SuccessCriteria: []string{"Unauthenticated commands cannot create missions"}}})
	if err != nil {
		t.Fatal(err)
	}
	calls := int64(3)
	deadline := time.Now().Add(time.Hour)
	content := domain.ProposalContent{Title: "Verify HTTP ledger commands without native execution", Goal: domain.GoalReference{ID: goal.ID, Revision: goal.Revision}, Scope: domain.ExecutionScope{ProjectID: project.ID, SettingsRevision: project.Settings.Revision, Roots: []string{root}, WritePaths: []string{"internal/swarm"}, Actions: settings.AllowedActions, MaxConcurrency: 2}, Budget: domain.ResourceBudget{Calls: &calls}, Stops: domain.StopConditions{Deadline: deadline, Conditions: []string{"Stop after HTTP authorization verification"}, MaxWorkItems: 2, MaxAttempts: 2, MaxNoProgressRounds: 2}, Deliverables: []string{"Actual HTTP and SQL receipts"}, SuccessCriteria: []string{"Revocation rejects subsequent protected commands"}}
	proposal, err := s2.CreateProposal(ctx, human, workspace.CreateProposalCommand{WorkspaceCommandMeta: swarmWorkspaceMeta(0), Content: content})
	if err != nil {
		t.Fatal(err)
	}
	proposal, err = s2.TransitionProposal(ctx, human, proposal.ID, swarmWorkspaceMeta(proposal.Version), domain.ProposalInReview)
	if err != nil {
		t.Fatal(err)
	}
	grant, err := s2.ApproveProposal(ctx, human, proposal.ID, workspace.ApproveProposalCommand{WorkspaceCommandMeta: swarmWorkspaceMeta(proposal.Version), ExpectedRevision: proposal.Revision, ExpiresAt: deadline})
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
	// 使用生产路由与生产 session guard，通过真实 HTTP listener 请求。
	mux := http.NewServeMux()
	h := &handler{}
	h.registerSwarmS3(mux, service)
	guard := newSessionGuard(Config{Services: Services{LocalProjects: projects}})
	mux.HandleFunc("GET /api/v1/auth/session", guard.bootstrap)
	server := httptest.NewServer(requestIDMiddleware(guard.middleware(mux)))
	defer server.Close()
	jar, err := cookiejar.New(nil)
	if err != nil {
		t.Fatal(err)
	}
	client := server.Client()
	client.Jar = jar
	response, err := client.Get(server.URL + "/api/v1/missions")
	if err != nil {
		t.Fatal(err)
	}
	response.Body.Close()
	if response.StatusCode != http.StatusForbidden {
		t.Fatalf("unauthenticated status=%d", response.StatusCode)
	}
	response, err = client.Get(server.URL + "/api/v1/auth/session")
	if err != nil {
		t.Fatal(err)
	}
	var session struct {
		CSRFToken string `json:"csrf_token"`
	}
	if err = json.NewDecoder(response.Body).Decode(&session); err != nil {
		t.Fatal(err)
	}
	response.Body.Close()
	if response.StatusCode != http.StatusOK || session.CSRFToken == "" {
		t.Fatal("actual human bootstrap failed")
	}
	post := func(path string, body any, key, csrf string) (int, []byte) {
		t.Helper()
		data, err := json.Marshal(body)
		if err != nil {
			t.Fatal(err)
		}
		request, err := http.NewRequest(http.MethodPost, server.URL+path, bytes.NewReader(data))
		if err != nil {
			t.Fatal(err)
		}
		request.Header.Set("Content-Type", "application/json")
		request.Header.Set("Idempotency-Key", key)
		request.Header.Set("X-CSRF-Token", csrf)
		response, err := client.Do(request)
		if err != nil {
			t.Fatal(err)
		}
		defer response.Body.Close()
		raw, err := io.ReadAll(response.Body)
		if err != nil {
			t.Fatal(err)
		}
		return response.StatusCode, raw
	}
	command := swarm.CreateMissionCommand{Meta: swarmHTTPMeta(0), GrantID: grant.ID, Epoch: grant.Epoch, ProposalRevision: grant.ProposalRevision, ProjectID: project.ID, Root: root}
	status, _ := post("/api/v1/missions", command, command.IdempotencyKey, "")
	if status != http.StatusForbidden {
		t.Fatalf("missing CSRF status=%d", status)
	}
	status, raw := post("/api/v1/missions", command, command.IdempotencyKey, session.CSRFToken)
	if status != http.StatusOK {
		t.Fatalf("approved mission status=%d", status)
	}
	var created struct {
		Mission swarmdomain.Mission `json:"mission"`
	}
	if err = json.Unmarshal(raw, &created); err != nil {
		t.Fatal(err)
	}
	if created.Mission.ID == "" {
		t.Fatal("created mission has no actual identity")
	}
	status, replayed := post("/api/v1/missions", command, command.IdempotencyKey, session.CSRFToken)
	if status != http.StatusOK {
		t.Fatalf("replay status=%d", status)
	}
	var replay struct {
		Mission swarmdomain.Mission `json:"mission"`
	}
	if err = json.Unmarshal(replayed, &replay); err != nil {
		t.Fatal(err)
	}
	if replay.Mission.ID != created.Mission.ID {
		t.Fatal("HTTP replay changed mission identity")
	}
	var forged map[string]any
	if err = json.Unmarshal(raw, &forged); err != nil {
		t.Fatal(err)
	}
	encoded, err := json.Marshal(command)
	if err != nil {
		t.Fatal(err)
	}
	if err = json.Unmarshal(encoded, &forged); err != nil {
		t.Fatal(err)
	}
	delete(forged, "mission")
	forged["actor_type"] = "human"
	status, _ = post("/api/v1/missions", forged, uuid.NewString(), session.CSRFToken)
	if status != http.StatusBadRequest {
		t.Fatalf("forged principal status=%d", status)
	}
	response, err = client.Get(server.URL + "/api/v1/missions/" + created.Mission.ID)
	if err != nil {
		t.Fatal(err)
	}
	var detail swarmDetailEnvelope
	if err = json.NewDecoder(response.Body).Decode(&detail); err != nil {
		t.Fatal(err)
	}
	response.Body.Close()
	if response.StatusCode != http.StatusOK || len(detail.Timeline) != 1 {
		t.Fatal("actual mission timeline is missing")
	}
	for _, event := range detail.Timeline {
		var count int
		if err = db.Conn().QueryRowContext(ctx, "SELECT COUNT(*) FROM foundation_outbox WHERE event_id=?", event.ID).Scan(&count); err != nil {
			t.Fatal(err)
		}
		if count != 1 {
			t.Fatal("HTTP event_id did not resolve in actual outbox")
		}
	}
	_, err = s2.ChangeApproval(ctx, human, grant.ID, workspace.ChangeApprovalCommand{WorkspaceCommandMeta: swarmWorkspaceMeta(grant.Epoch), Reason: "End public HTTP approval validation"}, "revoked")
	if err != nil {
		t.Fatal(err)
	}
	command.Meta = swarmHTTPMeta(0)
	status, _ = post("/api/v1/missions", command, command.IdempotencyKey, session.CSRFToken)
	if status != http.StatusForbidden {
		t.Fatalf("revoked approval status=%d", status)
	}
	pause := swarm.TransitionCommand{Meta: swarmHTTPMeta(created.Mission.Version), MissionID: created.Mission.ID, Reason: "Human safely pauses the revoked mission"}
	status, _ = post("/api/v1/missions/"+created.Mission.ID+"/pause", pause, pause.IdempotencyKey, session.CSRFToken)
	if status != http.StatusOK {
		t.Fatalf("safe human pause status=%d", status)
	}
	var leases int
	if err = db.Conn().QueryRowContext(ctx, "SELECT COUNT(*) FROM swarm_leases").Scan(&leases); err != nil {
		t.Fatal(err)
	}
	if leases != 0 {
		t.Fatal("HTTP ledger validation started execution")
	}
}
