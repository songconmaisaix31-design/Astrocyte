package s3_test

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/songconmaisaix31-design/Astrocyte/internal/adapters/agents"
	"github.com/songconmaisaix31-design/Astrocyte/internal/adapters/sqlite"
	"github.com/songconmaisaix31-design/Astrocyte/internal/apierrors"
	attention "github.com/songconmaisaix31-design/Astrocyte/internal/attention/app"
	swarm "github.com/songconmaisaix31-design/Astrocyte/internal/swarm/app"
	swarmdomain "github.com/songconmaisaix31-design/Astrocyte/internal/swarm/domain"
	workspace "github.com/songconmaisaix31-design/Astrocyte/internal/workspace/app"
	"github.com/songconmaisaix31-design/Astrocyte/internal/workspace/domain"
	"github.com/songconmaisaix31-design/Astrocyte/migrations"
)

func meta(version int) swarmdomain.Meta {
	return swarmdomain.Meta{SchemaVersion: 1, RequestID: uuid.NewString(), IdempotencyKey: uuid.NewString(), ExpectedVersion: version}
}
func workspaceMeta(version int) workspace.WorkspaceCommandMeta {
	return workspace.WorkspaceCommandMeta{SchemaVersion: 1, RequestID: uuid.NewString(), IdempotencyKey: uuid.NewString(), ExpectedVersion: version}
}
func code(t *testing.T, err error, want apierrors.Code) {
	t.Helper()
	var rule *swarmdomain.RuleError
	var service *apierrors.ServiceError
	if errors.As(err, &rule) && rule.Code == string(want) {
		return
	}
	if errors.As(err, &service) && service.Code == want {
		return
	}
	t.Fatalf("wanted %s, got %v", want, err)
}

func ledger(t *testing.T, db *sqlite.DB) (*swarm.Service, *workspace.S2Service) {
	t.Helper()
	s2 := workspace.NewS2Service(sqlite.NewWorkspaceRepository(db), agents.ProjectFiles{}, nil)
	repo, err := sqlite.NewSwarmRepository(db)
	if err != nil {
		t.Fatal(err)
	}
	approval, err := sqlite.NewSwarmApprovalBridge(db, s2)
	if err != nil {
		t.Fatal(err)
	}
	service, err := swarm.NewSwarmLedger(repo, approval)
	if err != nil {
		t.Fatal(err)
	}
	return service, s2
}

func TestActualApprovalMissionConcurrentIdempotencyRevokeAndRestart(t *testing.T) {
	ctx := context.Background()
	root, err := filepath.Abs("../..")
	if err != nil {
		t.Fatal(err)
	}
	root, err = agents.CanonicalRoot(root)
	if err != nil {
		t.Fatal(err)
	}
	base := filepath.Join(root, ".verification", "s3-approved")
	if err = os.MkdirAll(base, 0700); err != nil {
		t.Fatal(err)
	}
	dir, err := os.MkdirTemp(base, "actual-")
	if err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(dir, "state.sqlite")
	a, err := sqlite.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer a.Close()
	if err = a.RunMigrations(ctx, migrations.FS); err != nil {
		t.Fatal(err)
	}
	b, err := sqlite.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer b.Close()
	human := domain.Caller{Kind: "human", ID: uuid.NewString()}
	p := swarmdomain.Principal{Kind: human.Kind, ID: human.ID}
	attentionService := attention.NewAttentionService(sqlite.NewAttentionRepository(a), nil, nil, attention.ServiceOptions{})
	space, err := attentionService.CreateProjectSpace(ctx, attention.Principal{Kind: human.Kind, ID: human.ID}, attention.ProjectSpaceCommand{CommandMeta: attention.CommandMeta{SchemaVersion: 1, RequestID: uuid.NewString(), IdempotencyKey: uuid.NewString(), ExpectedVersion: 1}, Title: "Public Swarm authorization verification"})
	if err != nil {
		t.Fatal(err)
	}
	projects := workspace.NewLocalProjectService(a, nil, agents.ProjectFiles{}, nil)
	project, err := projects.RegisterProject(ctx, human, domain.RegisterProjectCommand{Name: "Astrocyte public Swarm implementation", Root: root, SpaceID: space.Space.ID})
	if err != nil {
		t.Fatal(err)
	}
	settings := project.Settings
	settings.AllowDirectory = true
	settings.AllowedSubdirs = []string{"."}
	settings.AllowedActions = []string{"create_mission", "manage_work_items", "read_swarm", "pause_mission", "cancel_mission", "recover_mission"}
	project, err = projects.SetProjectSettings(ctx, human, project.ID, domain.SettingsCommand{ExpectedRevision: project.Settings.Revision, Settings: settings})
	if err != nil {
		t.Fatal(err)
	}
	sa, s2 := ledger(t, a)
	sb, _ := ledger(t, b)
	goal, err := s2.PutGoal(ctx, human, "", workspace.GoalCommand{WorkspaceCommandMeta: workspaceMeta(0), Goal: domain.Goal{ProjectID: project.ID, Title: "Verify actual Swarm approval isolation", Status: "active", SuccessCriteria: []string{"One mission per current approval across two connections"}}})
	if err != nil {
		t.Fatal(err)
	}
	calls := int64(4)
	content := domain.ProposalContent{Title: "Verify Swarm authorization without launching native execution", Goal: domain.GoalReference{ID: goal.ID, Revision: goal.Revision}, Scope: domain.ExecutionScope{ProjectID: project.ID, SettingsRevision: project.Settings.Revision, Roots: []string{root}, WritePaths: []string{"internal/swarm"}, Actions: settings.AllowedActions, MaxConcurrency: 2}, Budget: domain.ResourceBudget{Calls: &calls}, Stops: domain.StopConditions{Deadline: time.Now().Add(time.Hour), Conditions: []string{"Stop when ledger verification completes"}, MaxWorkItems: 3, MaxAttempts: 2, MaxNoProgressRounds: 2}, Deliverables: []string{"Actual SQLite receipts and outbox events"}, SuccessCriteria: []string{"Current approval is required for each protected action"}}
	proposal, err := s2.CreateProposal(ctx, human, workspace.CreateProposalCommand{WorkspaceCommandMeta: workspaceMeta(0), Content: content})
	if err != nil {
		t.Fatal(err)
	}
	proposal, err = s2.TransitionProposal(ctx, human, proposal.ID, workspaceMeta(proposal.Version), domain.ProposalInReview)
	if err != nil {
		t.Fatal(err)
	}
	grant, err := s2.ApproveProposal(ctx, human, proposal.ID, workspace.ApproveProposalCommand{WorkspaceCommandMeta: workspaceMeta(proposal.Version), ExpectedRevision: proposal.Revision, ExpiresAt: content.Stops.Deadline})
	if err != nil {
		t.Fatal(err)
	}
	command := swarm.CreateMissionCommand{Meta: meta(0), GrantID: grant.ID, Epoch: grant.Epoch, ProposalRevision: grant.ProposalRevision, ProjectID: project.ID, Root: root}
	start := make(chan struct{})
	results := make(chan swarmdomain.Mission, 2)
	failures := make(chan error, 2)
	var wg sync.WaitGroup
	for _, service := range []*swarm.Service{sa, sb} {
		wg.Add(1)
		go func(s *swarm.Service) {
			defer wg.Done()
			<-start
			m, err := s.CreateMission(ctx, p, command)
			if err != nil {
				failures <- err
				return
			}
			results <- m
		}(service)
	}
	close(start)
	wg.Wait()
	close(failures)
	close(results)
	for err := range failures {
		t.Fatal(err)
	}
	var mission swarmdomain.Mission
	count := 0
	for m := range results {
		if count > 0 && m.ID != mission.ID {
			t.Fatal("same approved command created distinct missions")
		}
		mission = m
		count++
	}
	if count != 2 {
		t.Fatalf("successful requests=%d", count)
	}
	redelivery := command
	redelivery.Meta = meta(0)
	reused, err := sa.CreateMission(ctx, p, redelivery)
	if err != nil || reused.ID != mission.ID {
		t.Fatalf("same grant redelivery: %v", err)
	}
	var createdEvents int
	if err = a.Conn().QueryRowContext(ctx, "SELECT COUNT(*) FROM foundation_outbox WHERE aggregate_id=? AND event_type='swarm.mission_created'", mission.ID).Scan(&createdEvents); err != nil {
		t.Fatal(err)
	}
	if createdEvents != 1 {
		t.Fatalf("same grant emitted %d creation events", createdEvents)
	}
	changed := command
	changed.Root = filepath.Join(root, "internal")
	_, err = sa.CreateMission(ctx, p, changed)
	code(t, err, apierrors.VersionConflict)
	item, err := sa.AddWorkItem(ctx, p, swarm.AddWorkItemCommand{Meta: meta(mission.Version), MissionID: mission.ID, Title: "Review actual Swarm ledger rules", Inputs: []swarmdomain.SourceVersion{{ID: proposal.ID, Revision: "1"}}, WritePaths: []string{"internal/swarm/domain"}})
	if err != nil {
		t.Fatal(err)
	}
	if item.State != "open" {
		t.Fatal("new work item unexpectedly executed")
	}
	_, err = sa.AddWorkItem(ctx, p, swarm.AddWorkItemCommand{Meta: meta(mission.Version + 1), MissionID: mission.ID, Title: "Reject an unapproved source", Inputs: []swarmdomain.SourceVersion{{ID: uuid.NewString(), Revision: "1"}}, WritePaths: []string{"internal/swarm/domain"}})
	code(t, err, apierrors.ScopeDenied)
	detail, err := sa.ReadMission(ctx, p, mission.ID)
	if err != nil {
		t.Fatal(err)
	}
	if detail.Mission.WorkItemCount != 1 || len(detail.WorkItems) != 1 || detail.Mission.Version != mission.Version+1 {
		t.Fatal("rejected source left partial mission or work item state")
	}
	if err = a.Close(); err != nil {
		t.Fatal(err)
	}
	if err = b.Close(); err != nil {
		t.Fatal(err)
	}
	a, err = sqlite.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer a.Close()
	if err = a.RunMigrations(ctx, migrations.FS); err != nil {
		t.Fatal(err)
	}
	sa, s2 = ledger(t, a)
	result, err := sa.CreateMission(ctx, p, command)
	if err != nil || result.ID != mission.ID {
		t.Fatalf("restart receipt: %v", err)
	}
	_, err = s2.ChangeApproval(ctx, human, grant.ID, workspace.ChangeApprovalCommand{WorkspaceCommandMeta: workspaceMeta(grant.Epoch), Reason: "End current Swarm authorization validation"}, "revoked")
	if err != nil {
		t.Fatal(err)
	}
	_, err = sa.CreateMission(ctx, p, command)
	code(t, err, apierrors.ApprovalRevoked)
	var missions, leases int
	if err = a.Conn().QueryRowContext(ctx, "SELECT COUNT(*) FROM swarm_missions").Scan(&missions); err != nil {
		t.Fatal(err)
	}
	if err = a.Conn().QueryRowContext(ctx, "SELECT COUNT(*) FROM swarm_leases").Scan(&leases); err != nil {
		t.Fatal(err)
	}
	if missions != 1 || leases != 0 {
		t.Fatalf("missions=%d leases=%d", missions, leases)
	}
}
