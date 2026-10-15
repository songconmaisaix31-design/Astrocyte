package app_test

import (
	"context"
	"path/filepath"
	"reflect"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/songconmaisaix31-design/Astrocyte/internal/adapters/agents"
	"github.com/songconmaisaix31-design/Astrocyte/internal/adapters/importers"
	"github.com/songconmaisaix31-design/Astrocyte/internal/adapters/sqlite"
	"github.com/songconmaisaix31-design/Astrocyte/internal/apierrors"
	. "github.com/songconmaisaix31-design/Astrocyte/internal/attention/app"
	workspace "github.com/songconmaisaix31-design/Astrocyte/internal/workspace/app"
	workdomain "github.com/songconmaisaix31-design/Astrocyte/internal/workspace/domain"
)

// 观察调用主体和次数，来源结果完全由生产 Reader 返回。
type observedCompletionSource struct {
	ExperienceCompletionReader
	reads, authorizations int
	principal             Principal
}

func (s *observedCompletionSource) ReadExperienceCompletion(ctx context.Context, p Principal, project, event string) (ExperienceCompletionEvidence, error) {
	s.reads++
	s.principal = p
	return s.ExperienceCompletionReader.ReadExperienceCompletion(ctx, p, project, event)
}

func (s *observedCompletionSource) AuthorizeExperienceCompletion(ctx context.Context, p Principal, evidence ExperienceCompletionEvidence) error {
	s.authorizations++
	return s.ExperienceCompletionReader.AuthorizeExperienceCompletion(ctx, p, evidence)
}

func TestCompletionCaptureActualProjectMissingSourceAndReopen(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	core, db, _ := fixture(t)
	var file string
	if err := db.Conn().QueryRow("SELECT file FROM pragma_database_list WHERE name='main'").Scan(&file); err != nil {
		t.Fatal("resolve actual completion database")
	}
	space, err := core.CreateProjectSpace(ctx, human, ProjectSpaceCommand{CommandMeta: meta("completion-space", 1), Title: "Public completion source permission verification"})
	if err != nil {
		t.Fatal("create actual completion project space")
	}
	root, err := filepath.Abs("../../..")
	if err != nil {
		t.Fatal("resolve approved public Git worktree")
	}
	root, err = agents.CanonicalRoot(root)
	if err != nil {
		t.Fatal("canonicalize approved public Git worktree")
	}
	projects := workspace.NewLocalProjectService(db, nil, agents.ProjectFiles{}, nil)
	caller := workdomain.Caller{ID: human.ID, Kind: human.Kind}
	project, err := projects.RegisterProject(ctx, caller, workdomain.RegisterProjectCommand{Name: "Actual completion source verification", Root: root, SpaceID: space.Space.ID})
	if err != nil || project.ID == project.Root || project.ID == project.SpaceID {
		t.Fatal("register actual local project identity")
	}
	grant, err := projects.GrantProjectAgent(ctx, caller, project.ID, workdomain.ProjectGrant{AgentID: "completion-current-agent", Actions: []string{"discover"}})
	if err != nil {
		t.Fatal("grant actual project discovery permission")
	}
	token, err := projects.IssueProjectAgentToken(ctx, caller, project.ID, grant.AgentID)
	if err != nil {
		t.Fatal("issue actual scoped Agent token")
	}
	scoped, err := projects.AuthenticateAgentToken(ctx, token.Token)
	if err != nil || scoped.ProjectID != project.ID || scoped.Kind != "agent" || scoped.ID != grant.AgentID {
		t.Fatal("authenticate current actual project principal")
	}
	trustedAgent := Principal{ID: scoped.ID, Kind: scoped.Kind}
	t.Setenv("LOCALAPPDATA", filepath.Join(filepath.Dir(file), "relay-local"))
	relayRoot, err := sqlite.RelayStateRoot()
	if err != nil {
		t.Fatal("resolve owned Relay store")
	}
	relay, err := sqlite.OpenRelayStore(relayRoot)
	if err != nil {
		t.Fatal("open actual independent Relay store")
	}
	t.Cleanup(func() {
		if err := relay.Close(); err != nil {
			t.Error("close actual independent Relay store")
		}
	})
	accesses := 0
	var authenticated workdomain.Caller
	var owner string
	access := func(readctx context.Context, p Principal, id string) (string, error) {
		accesses++
		authenticated = workdomain.Caller{ID: p.ID, Kind: p.Kind}
		current, err := projects.ListProjects(readctx, authenticated)
		if err != nil {
			return "", err
		}
		for _, registered := range current {
			if registered.ID == id {
				owner = registered.ID
				return owner, nil
			}
		}
		return "", apierrors.NewNotFound("local project", id)
	}
	assemble := func(database *sqlite.DB) (*AcceptedExperienceService, *observedCompletionSource) {
		t.Helper()
		runtime := workspace.NewRelayService(relay, nil, database, nil, nil, nil, nil, workspace.RelayServiceOptions{})
		reader, ok := runtime.(workspace.RelayCompletionEvidenceReader)
		if !ok {
			t.Fatal("production Relay completion reader unavailable")
		}
		source, err := agents.NewRelayCompletionExperienceSource(reader, access)
		if err != nil {
			t.Fatal("assemble actual source and current project reader")
		}
		observed := &observedCompletionSource{ExperienceCompletionReader: source}
		repository := sqlite.NewAcceptedExperienceRepository(database)
		governance := NewExperienceService(repository, sqlite.NewEventRepository(database), ExperienceOptions{ProjectID: project.ID})
		return NewAcceptedExperienceService(governance, repository, observed, nil), observed
	}
	service, observed := assemble(db)
	before, err := relay.Snapshot(ctx)
	if err != nil {
		t.Fatal("read original actual Relay state")
	}
	beforeReceipts, beforeEvents := countRows(t, db, "attention_receipts"), countRows(t, db, "attention_outbox")
	// 未观察到的 UUID 仅作为缺失来源的拒绝输入，未创建完成事件。
	command := CaptureCompletionExperienceCommand{CommandMeta: meta("completion-missing-source", 1), LocalProjectID: project.ID, EventID: uuid.NewString()}
	for _, p := range []Principal{trustedAgent, {ID: "", Kind: "human"}, {ID: human.ID, Kind: "untrusted"}} {
		_, err := service.CaptureCompletionExperience(ctx, p, command)
		want := apierrors.ScopeDenied
		if p.Kind == "agent" {
			want = apierrors.ApprovalRequired
		}
		errorCode(t, err, want)
	}
	if observed.reads != 0 || accesses != 0 {
		t.Fatal("unauthorized principal entered completion source reading")
	}
	for _, invalid := range []CaptureCompletionExperienceCommand{
		{CommandMeta: meta("completion-wrong-version", 2), LocalProjectID: project.ID, EventID: command.EventID},
		{CommandMeta: meta("completion-empty-project", 1), EventID: command.EventID},
		{CommandMeta: meta("completion-empty-event", 1), LocalProjectID: project.ID},
	} {
		_, err := service.CaptureCompletionExperience(ctx, human, invalid)
		errorCode(t, err, apierrors.ValidationFailed)
	}
	if observed.reads != 0 || accesses != 0 {
		t.Fatal("invalid command entered completion source reading")
	}
	verifyMissing := func(database *sqlite.DB) {
		t.Helper()
		for range 2 {
			_, err := service.CaptureCompletionExperience(ctx, human, command)
			errorCode(t, err, apierrors.NotFound)
			if observed.principal != human || authenticated != caller || owner != project.ID || observed.authorizations != 0 {
				t.Fatal("actual source lost current principal or owner, or missing source entered persistence")
			}
		}
		for _, alias := range []string{project.Root, project.SpaceID, uuid.NewString()} {
			wrong := command
			wrong.CommandMeta = meta("completion-alias-"+alias, 1)
			wrong.LocalProjectID = alias
			_, err := service.CaptureCompletionExperience(ctx, human, wrong)
			errorCode(t, err, apierrors.NotFound)
		}
		beforeAccesses := accesses
		_, err := observed.ExperienceCompletionReader.ReadExperienceCompletion(ctx, trustedAgent, project.ID, command.EventID)
		errorCode(t, err, apierrors.ScopeDenied)
		if accesses != beforeAccesses {
			t.Fatal("unauthorized Agent completion action entered project source reading")
		}
		for _, table := range []string{"attention_engine_experiences", "experience_completion_sources", "local_agent_sessions"} {
			if countRows(t, database, table) != 0 {
				t.Fatal("missing or denied completion produced candidate, evidence or execution")
			}
		}
		if countRows(t, database, "attention_receipts") != beforeReceipts || countRows(t, database, "attention_outbox") != beforeEvents {
			t.Fatal("rejected completion changed an existing receipt or event")
		}
		state, err := relay.Snapshot(ctx)
		if err != nil || !reflect.DeepEqual(before, state) {
			t.Fatal("read-only completion rejection changed original Relay state")
		}
		if err = sqlite.NewAcceptedExperienceRepository(database).WithAcceptedExperienceTx(ctx, func(tx AcceptedExperienceTx) error {
			_, err := tx.LoadExperienceReceipt(human.Kind+":"+human.ID, "Experience/CaptureCompletion/"+project.ID, command.IdempotencyKey)
			errorCode(t, err, apierrors.NotFound)
			return nil
		}); err != nil {
			t.Fatal("inspect original rejected completion command receipt")
		}
	}
	verifyMissing(db)
	if err = relay.Close(); err != nil {
		t.Fatal("close original Relay store before reopen")
	}
	if err = db.Close(); err != nil {
		t.Fatal("close original main database before reopen")
	}
	reopened, err := sqlite.Open(file)
	if err != nil {
		t.Fatal("reopen actual completion main database")
	}
	t.Cleanup(func() {
		if err := reopened.Close(); err != nil {
			t.Error("close reopened completion main database")
		}
	})
	relay, err = sqlite.OpenRelayStore(relayRoot)
	if err != nil {
		t.Fatal("reopen original independent Relay store")
	}
	projects = workspace.NewLocalProjectService(reopened, nil, agents.ProjectFiles{}, nil)
	service, observed = assemble(reopened)
	verifyMissing(reopened)
	after, err := projects.AuthenticateAgentToken(ctx, token.Token)
	if err != nil || !reflect.DeepEqual(scoped, after) {
		t.Fatal("actual current Agent identity changed after SQLite reopen")
	}
	if _, err = projects.RevokeProjectAgent(ctx, caller, project.ID, grant.AgentID); err != nil {
		t.Fatal("revoke original actual Agent grant")
	}
	_, err = projects.AuthenticateAgentToken(ctx, token.Token)
	errorCode(t, err, apierrors.ApprovalRevoked)
	_, err = service.CaptureCompletionExperience(ctx, trustedAgent, command)
	errorCode(t, err, apierrors.ApprovalRequired)
	if countRows(t, reopened, "attention_receipts") != beforeReceipts || countRows(t, reopened, "attention_outbox") != beforeEvents || countRows(t, reopened, "attention_engine_experiences") != 0 || countRows(t, reopened, "experience_completion_sources") != 0 || countRows(t, reopened, "local_agent_sessions") != 0 {
		t.Fatal("revoked or unauthorized Agent capture changed completion state")
	}
}

func TestCompletionRepositorySharesActualCommandReceiptRollbackAndReopen(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	core, db, objects := fixture(t)
	var file string
	if err := db.Conn().QueryRow("SELECT file FROM pragma_database_list WHERE name='main'").Scan(&file); err != nil {
		t.Fatal("resolve original receipt database")
	}
	repository := sqlite.NewAcceptedExperienceRepository(db)
	command := UpdateRankingProfileCommand{CommandMeta: meta("completion-repository-human-profile", 1), Enabled: true, Weights: map[string]float64{"goal_progress": 4, "current_interest": 2, "project_improvement": 2, "originality": 1}}
	caller, operation := human.Kind+":"+human.ID, "UpdateRankingProfile"
	var saved RankingProfileDetail
	var original Receipt
	waits := db.Conn().Stats().WaitCount
	write := func(tx AcceptedExperienceTx) error {
		conn, err := sqlite.WorkspaceTransaction(tx.Context(), db)
		if err != nil {
			return err
		}
		if conn == nil {
			t.Fatal("accepted experience transaction lost its original database connection")
		}
		saved, err = core.UpdateRankingProfile(tx.Context(), human, command)
		if err != nil {
			return err
		}
		original, err = tx.LoadExperienceReceipt(caller, operation, command.IdempotencyKey)
		return err
	}
	err := repository.WithAcceptedExperienceTx(ctx, func(tx AcceptedExperienceTx) error {
		if err := write(tx); err != nil {
			return err
		}
		// 重复保存实际人工命令的原收据，触发真实 SQLite 唯一约束。
		return tx.SaveExperienceReceipt(caller, operation, command.IdempotencyKey, original)
	})
	if err == nil || saved.Profile == nil || original.Digest == "" || len(original.Result) == 0 || db.Conn().Stats().WaitCount != waits {
		t.Fatal("actual duplicate receipt failure or shared transaction was not observed")
	}
	errorCode(t, err, apierrors.VersionConflict)
	for _, table := range []string{"attention_ranking_profiles", "attention_ranking_profile_revisions", "attention_receipts", "attention_outbox"} {
		if countRows(t, db, table) != 0 {
			t.Fatal("duplicate receipt left partial human command state")
		}
	}
	if err = repository.WithAcceptedExperienceTx(ctx, write); err != nil || saved.Profile == nil || saved.Profile.Version != 1 {
		t.Fatal("rolled-back human command did not commit once after repair")
	}
	if countRows(t, db, "attention_receipts") != 1 || countRows(t, db, "attention_outbox") != 1 || countRows(t, db, "attention_ranking_profile_revisions") != 1 {
		t.Fatal("actual command retained duplicate receipts, events or versions")
	}
	if err = db.Close(); err != nil {
		t.Fatal("close original receipt database")
	}
	reopened, err := sqlite.Open(file)
	if err != nil {
		t.Fatal("reopen original receipt database")
	}
	t.Cleanup(func() {
		if err := reopened.Close(); err != nil {
			t.Error("close reopened receipt database")
		}
	})
	core = NewAttentionService(sqlite.NewAttentionRepository(reopened), importers.NewReader(nil), objects, ServiceOptions{})
	replay, err := core.UpdateRankingProfile(ctx, human, command)
	if err != nil || !sameAdmissionJSON(t, saved, replay) {
		t.Fatal("original command receipt or complete profile changed after reopen")
	}
	if err = sqlite.NewAcceptedExperienceRepository(reopened).WithAcceptedExperienceTx(ctx, func(tx AcceptedExperienceTx) error {
		stored, err := tx.LoadExperienceReceipt(caller, operation, command.IdempotencyKey)
		if err != nil {
			return err
		}
		if !reflect.DeepEqual(original, stored) {
			t.Fatal("original persisted human command receipt changed")
		}
		return nil
	}); err != nil {
		t.Fatal("read original persisted receipt through completion repository")
	}
	for _, table := range []string{"attention_engine_experiences", "experience_completion_sources", "local_agent_sessions"} {
		if countRows(t, reopened, table) != 0 {
			t.Fatal("receipt transaction created completion evidence or native execution")
		}
	}
	if countRows(t, reopened, "attention_receipts") != 1 || countRows(t, reopened, "attention_outbox") != 1 || countRows(t, reopened, "attention_ranking_profile_revisions") != 1 {
		t.Fatal("reopen or command replay changed durable row counts")
	}
}
