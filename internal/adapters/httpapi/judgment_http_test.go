package httpapi

import (
	"bytes"
	"context"
	"encoding/json"
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
	"github.com/songconmaisaix31-design/Astrocyte/internal/adapters/importers"
	adapter "github.com/songconmaisaix31-design/Astrocyte/internal/adapters/judgment"
	"github.com/songconmaisaix31-design/Astrocyte/internal/adapters/objects"
	"github.com/songconmaisaix31-design/Astrocyte/internal/adapters/repositorycontext"
	"github.com/songconmaisaix31-design/Astrocyte/internal/adapters/sqlite"
	"github.com/songconmaisaix31-design/Astrocyte/internal/apierrors"
	attention "github.com/songconmaisaix31-design/Astrocyte/internal/attention/app"
	"github.com/songconmaisaix31-design/Astrocyte/internal/judgment"
	workspace "github.com/songconmaisaix31-design/Astrocyte/internal/workspace/app"
	"github.com/songconmaisaix31-design/Astrocyte/internal/workspace/domain"
	"github.com/songconmaisaix31-design/Astrocyte/migrations"
)

type judgmentHTTPRun struct {
	db          *sqlite.DB
	projects    workspace.LocalProjects
	access      *adapter.ProjectAuthorization
	client      *http.Client
	url         string
	csrf        string
	privateRoot string
	close       func()
}

func judgmentHTTPStart(t *testing.T, repoRoot, runRoot string) *judgmentHTTPRun {
	t.Helper()
	db, err := sqlite.Open(filepath.Join(runRoot, "state.sqlite"))
	if err != nil {
		t.Fatal(err)
	}
	if err := db.RunMigrations(context.Background(), migrations.FS); err != nil {
		t.Fatal(err)
	}
	store, err := objects.New(filepath.Join(runRoot, "objects"))
	if err != nil {
		t.Fatal(err)
	}
	a := attention.NewAttentionService(sqlite.NewAttentionRepository(db), importers.NewReader([]string{filepath.Join(repoRoot, ".verification", "s5")}), store, attention.ServiceOptions{})
	refs := adapter.AttentionReferences{Attention: a}
	projects := workspace.NewLocalProjectService(db, refs, agents.ProjectFiles{}, nil)
	environment, err := os.Hostname()
	if err != nil {
		t.Fatal(err)
	}
	snapshots, err := repositorycontext.New(db, db, repositorycontext.Options{Git: repositorycontext.Git{EnvironmentID: environment}, CacheRoot: filepath.Join(runRoot, "projections"), SnapshotBudget: domain.RepositoryBudget{MaxFiles: 4, MaxBytes: 1048576, MaxResults: 1, MaxDepth: 1}})
	if err != nil {
		t.Fatal(err)
	}
	s2 := workspace.NewS2Service(sqlite.NewWorkspaceRepository(db), agents.ProjectFiles{}, refs)
	access := &adapter.ProjectAuthorization{Projects: db, References: refs, Policies: db, Snapshots: snapshots, Goals: s2, Taste: a, Ranking: a, Rules: db}
	core, err := judgment.NewService(db, access)
	if err != nil {
		t.Fatal(err)
	}
	mux := http.NewServeMux()
	services := Services{Attention: a, LocalProjects: projects}
	h := &handler{services: services, logger: slog.New(slog.NewTextHandler(io.Discard, nil))}
	h.registerAttention(mux)
	h.registerLocalProjects(mux)
	h.registerJudgment(mux, JudgmentServices{Judgments: core, Policies: access, Taste: a})
	guard := newSessionGuard(Config{Services: services})
	mux.HandleFunc("GET /api/v1/auth/session", guard.bootstrap)
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	server := &http.Server{Handler: requestIDMiddleware(loopbackMiddleware(guard.middleware(mux)))}
	served := make(chan error, 1)
	go func() { served <- server.Serve(listener) }()
	jar, err := cookiejar.New(nil)
	if err != nil {
		t.Fatal(err)
	}
	f := &judgmentHTTPRun{db: db, projects: projects, access: access, client: &http.Client{Jar: jar, Timeout: 30 * time.Second}, url: "http://" + listener.Addr().String()}
	closed := false
	f.close = func() {
		if closed {
			return
		}
		closed = true
		if err := server.Close(); err != nil {
			t.Error(err)
		}
		if err := <-served; err != http.ErrServerClosed {
			t.Error(err)
		}
		if err := db.Close(); err != nil {
			t.Error(err)
		}
	}
	t.Cleanup(f.close)
	var session struct {
		Actor string `json:"actor_id"`
		Kind  string `json:"actor_kind"`
		CSRF  string `json:"csrf_token"`
	}
	f.call(t, http.MethodGet, "/api/v1/auth/session", nil, http.StatusOK, &session)
	if session.Actor != "local-human" || session.Kind != "human" || session.CSRF == "" {
		t.Fatal("the actual production session guard did not authenticate the local human")
	}
	f.csrf = session.CSRF
	return f
}

func (f *judgmentHTTPRun) call(t *testing.T, method, path string, value any, status int, target any) {
	t.Helper()
	var encoded []byte
	var err error
	if value != nil {
		encoded, err = json.Marshal(value)
		if err != nil {
			t.Fatal(err)
		}
	}
	r, err := http.NewRequest(method, f.url+path, bytes.NewReader(encoded))
	if err != nil {
		t.Fatal(err)
	}
	if value != nil {
		r.Header.Set("Content-Type", "application/json")
		r.Header.Set("X-CSRF-Token", f.csrf)
		r.Header.Set("Idempotency-Key", uuid.NewString())
	}
	response, err := f.client.Do(r)
	if err != nil {
		t.Fatal(err)
	}
	defer response.Body.Close()
	data, err := io.ReadAll(response.Body)
	if err != nil {
		t.Fatal(err)
	}
	if f.privateRoot != "" && (bytes.Contains(data, []byte(f.privateRoot)) || bytes.Contains(data, []byte(strings.ReplaceAll(f.privateRoot, `\`, `\\`)))) {
		t.Fatal("actual public judgment HTTP response disclosed its local absolute project root")
	}
	if response.StatusCode != status {
		t.Fatalf("actual judgment HTTP %s %s returned %d, expected %d", method, path, response.StatusCode, status)
	}
	if target != nil {
		if err := json.Unmarshal(data, target); err != nil {
			t.Fatal(err)
		}
	}
}

func TestJudgmentActualHTTPRulesFeedbackAndRestart(t *testing.T) {
	if os.Getenv("ASTROCYTE_S5_LIVE") != "1" {
		t.Skip("requires the explicitly prepared public S5 worktree and official source assets")
	}
	repoRoot, err := filepath.Abs("../../..")
	if err != nil {
		t.Fatal(err)
	}
	evidenceRoot := os.Getenv("ASTROCYTE_S5_RUN_ROOT")
	if !filepath.IsAbs(evidenceRoot) {
		t.Fatal("ASTROCYTE_S5_RUN_ROOT must be an explicit absolute owned evidence directory")
	}
	runRoot, err := os.MkdirTemp(evidenceRoot, "http-")
	if err != nil {
		t.Fatal(err)
	}
	f := judgmentHTTPStart(t, repoRoot, runRoot)
	ctx := context.Background()
	human := domain.Caller{ID: "local-human", Kind: "human"}
	meta := func() attention.CommandMeta {
		return attention.CommandMeta{SchemaVersion: 1, RequestID: uuid.NewString(), IdempotencyKey: uuid.NewString(), ExpectedVersion: 1}
	}
	readme, err := os.ReadFile(filepath.Join(repoRoot, ".verification", "s5", "upstream", "laya", "laya-ts", "README.md"))
	if err != nil {
		t.Fatal(err)
	}
	text := strings.Split(string(readme), "## Node")[0]
	var job attention.ImportJobResult
	f.call(t, http.MethodPost, "/api/v1/materials/imports", attention.ImportMaterialCommand{CommandMeta: meta(), Kind: "text", Adapter: "manual", SourceLocator: "https://github.com/NandhaKishorM/laya/blob/432312b36714d8889eba405507b37eda25257d9c/laya-ts/README.md", Title: "Official Laya ONNX export documentation", ExportText: text, ContentDigest: judgment.DigestText(text), SourceSpans: []string{"Export weights section"}}, http.StatusAccepted, &job)
	a := attention.NewAttentionService(sqlite.NewAttentionRepository(f.db), importers.NewReader([]string{runRoot}), mustJudgmentObjects(t, filepath.Join(runRoot, "objects")), attention.ServiceOptions{})
	if _, err := a.ProcessNextJob(ctx); err != nil {
		t.Fatal(err)
	}
	principal := attention.Principal{ID: human.ID, Kind: human.Kind}
	actual, err := a.GetJob(ctx, principal, job.JobID)
	if err != nil || actual.MaterialID == nil || actual.MaterialRevision == nil || actual.Status != "succeeded" {
		t.Fatal("the actual original material import did not finish", err)
	}
	ref := judgment.EvidenceRef{MaterialID: *actual.MaterialID, Revision: *actual.MaterialRevision, Digest: judgment.DigestText(text)}
	var taste attention.TasteProfileDetail
	f.call(t, http.MethodGet, "/api/v1/attention-taste-profile", nil, http.StatusOK, &taste)
	if taste.Configured || taste.Profile != nil {
		t.Fatal("the actual new database invented a complete taste profile")
	}
	tasteCommand := attention.UpdateTasteProfileCommand{CommandMeta: meta(), Enabled: true, Rubrics: []attention.TasteRubric{{ID: "official-source-availability", QuestionType: "truth", Criteria: "The selected fixed official source section is present"}}, Examples: []attention.TasteExample{{ID: "official-export-section", RubricID: "official-source-availability", MaterialID: ref.MaterialID, Revision: ref.Revision, ContentDigest: ref.Digest, Expected: "1"}}, Delegations: []attention.TasteDelegation{}}
	f.call(t, http.MethodPut, "/api/v1/attention-taste-profile", tasteCommand, http.StatusOK, &taste)
	if !taste.Configured || taste.Profile == nil || taste.Profile.Version != 1 || taste.Profile.ConfiguredBy != human.ID {
		t.Fatal("the actual HTTP human did not create a fixed complete taste profile")
	}
	tasteCommand.CommandMeta = meta()
	tasteCommand.Rubrics[0].Criteria = "Verify that each fixed official source section is present before recommending review"
	f.call(t, http.MethodPut, "/api/v1/attention-taste-profile", tasteCommand, http.StatusOK, &taste)
	if taste.Profile == nil || taste.Profile.Version != 2 {
		t.Fatal("the actual HTTP rubric change did not create the next immutable taste version")
	}
	tasteCommand.CommandMeta = meta()
	f.call(t, http.MethodPut, "/api/v1/attention-taste-profile", tasteCommand, http.StatusConflict, nil)
	space, err := a.CreateProjectSpace(ctx, principal, attention.ProjectSpaceCommand{CommandMeta: meta(), Title: "Public judgment HTTP verification"})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := a.ReferenceMaterial(ctx, principal, space.Space.ID, attention.ReferenceMaterialCommand{CommandMeta: meta(), MaterialID: ref.MaterialID, Revision: ref.Revision}); err != nil {
		t.Fatal(err)
	}
	project, err := f.projects.RegisterProject(ctx, human, domain.RegisterProjectCommand{Name: "S5 public HTTP project", Root: filepath.Join(repoRoot, ".verification", "s5", "projects", "astro-a"), SpaceID: space.Space.ID})
	if err != nil {
		t.Fatal(err)
	}
	settings := project.Settings
	settings.AllowDirectory, settings.AllowedSubdirs = true, []string{"."}
	project, err = f.projects.SetProjectSettings(ctx, human, project.ID, domain.SettingsCommand{ExpectedRevision: settings.Revision, Settings: settings})
	if err != nil {
		t.Fatal(err)
	}
	f.privateRoot = project.Root
	goal, err := f.access.Goals.(*workspace.S2Service).PutGoal(ctx, human, "", workspace.GoalCommand{WorkspaceCommandMeta: workspace.WorkspaceCommandMeta{SchemaVersion: 1, RequestID: uuid.NewString(), IdempotencyKey: uuid.NewString()}, Goal: domain.Goal{ProjectID: project.ID, Title: "Verify the fixed official Laya CPU documentation", SuccessCriteria: []string{"Read the official export verification and local runtime contract"}, Priority: 1, Status: "active"}})
	if err != nil {
		t.Fatal(err)
	}
	prefix := "/api/v1/local-projects/" + project.ID
	var captured struct {
		Snapshot domain.ProjectContextSnapshot `json:"snapshot"`
	}
	f.call(t, http.MethodPost, prefix+"/judgment-snapshots", JudgmentSnapshotCommand{Files: []string{"README.md"}}, http.StatusOK, &captured)
	snapshot := captured.Snapshot
	if snapshot.Root != "" || snapshot.Generation == "" {
		t.Fatal("public snapshot must expose its actual fixed identity without a local absolute root")
	}
	policy := judgment.Policy{Revision: 1, ProjectID: project.ID, ProjectSettingsRevision: project.Settings.Revision, WorktreeID: snapshot.WorktreeID, SnapshotID: snapshot.Generation, Snapshot: snapshot, SnapshotFiles: []string{"README.md"}, QuestionType: judgment.Truth, Backend: judgment.RuleBackend, Languages: []string{"en"}, Versions: judgment.Versions{Goal: "public-goal-v1", Taste: "manual-review-v1", Policy: "http-policy-v1", Model: "rule:all_evidence_present/v1", AccessEpoch: "public-http-v1"}, Budget: judgment.Budget{TimeoutMS: 60000, MaxInputBytes: 131072, MaxInputTokens: 2048, MaxOutputTokens: 0}}
	policy.Sources = &judgment.SourceReferences{Goal: domain.GoalReference{ID: goal.ID, Revision: goal.Revision}, Taste: judgment.ProfileReference{ID: taste.Profile.ID, Version: taste.Profile.Version}}
	private := policy
	private.Snapshot.Root = project.Root
	f.call(t, http.MethodPut, prefix+"/judgment-policies/truth", JudgmentPolicyCommand{Policy: private}, http.StatusBadRequest, nil)
	f.call(t, http.MethodPut, prefix+"/judgment-policies/truth", JudgmentPolicyCommand{Policy: policy}, http.StatusOK, nil)
	var saved struct {
		Policy judgment.Policy `json:"policy"`
	}
	f.call(t, http.MethodGet, prefix+"/judgment-policies/truth", nil, http.StatusOK, &saved)
	if saved.Policy.Snapshot.Root != "" || saved.Policy.Snapshot.Generation != snapshot.Generation || !reflect.DeepEqual(saved.Policy.Sources, policy.Sources) {
		t.Fatal("the persisted public policy disclosed or changed its actual snapshot identity")
	}
	request := judgment.JudgmentRequest{SchemaVersion: "1.0", ProjectID: project.ID, WorktreeID: policy.WorktreeID, SnapshotID: policy.SnapshotID, Question: judgment.Question{ID: "official-export-availability", Type: judgment.Truth, Instructions: "Are all selected official export source sections available?", Rule: "all_evidence_present"}, EvidenceRefs: []judgment.EvidenceRef{ref}, Language: "en", Backend: policy.Backend, Versions: policy.Versions, Budget: policy.Budget}
	request.Sources = policy.Sources
	var answer struct {
		Judgment judgment.JudgmentResult `json:"judgment"`
	}
	f.call(t, http.MethodPost, prefix+"/judgments", request, http.StatusOK, &answer)
	original := answer.Judgment
	if original.ID == "" || original.Decision != judgment.Recommend || original.ProviderConfidence != nil || original.CacheHit || !reflect.DeepEqual(original.Sources, policy.Sources) {
		t.Fatal("actual HTTP rule judgment is incomplete")
	}
	f.call(t, http.MethodPost, prefix+"/judgments", request, http.StatusOK, &answer)
	if !answer.Judgment.CacheHit || answer.Judgment.ID != original.ID {
		t.Fatal("actual HTTP cache did not reuse the fixed result")
	}
	path := prefix + "/judgments/" + original.ID
	for _, kind := range []string{"adopt", "later", "reject", "revise", "already_solved"} {
		f.call(t, http.MethodPost, path+"/feedback", JudgmentFeedbackCommand{ResultVersion: original.Version, Kind: kind, Reason: "Record this explicit review of the fixed public documentation result"}, http.StatusOK, nil)
	}
	f.call(t, http.MethodPost, path+"/feedback", JudgmentFeedbackCommand{ResultVersion: original.Version + 1, Kind: "adopt", Reason: "Verify exact immutable result version"}, http.StatusConflict, nil)
	f.csrf = ""
	f.call(t, http.MethodPost, path+"/feedback", JudgmentFeedbackCommand{ResultVersion: original.Version, Kind: "adopt", Reason: "Verify the actual CSRF rejection"}, http.StatusForbidden, nil)
	f.close()
	f = judgmentHTTPStart(t, repoRoot, runRoot)
	f.privateRoot = project.Root
	f.call(t, http.MethodGet, path, nil, http.StatusOK, &answer)
	if answer.Judgment.ID != original.ID || answer.Judgment.InputDigest != original.InputDigest || !reflect.DeepEqual(answer.Judgment.Sources, policy.Sources) {
		t.Fatal("actual SQLite and HTTP restart changed the immutable result")
	}
	var history struct {
		Items []judgment.Feedback `json:"items"`
	}
	f.call(t, http.MethodGet, path+"/feedback", nil, http.StatusOK, &history)
	if len(history.Items) != 5 {
		t.Fatal("actual HTTP feedback did not survive SQLite restart")
	}
	for _, item := range history.Items {
		if item.ActorID != human.ID || item.ActorKind != "human" || item.ResultVersion != original.Version {
			t.Fatal("persisted feedback lost its actual authenticated human or fixed version")
		}
	}
	f.call(t, http.MethodGet, "/api/v1/attention-taste-profile", nil, http.StatusOK, &taste)
	if taste.Profile == nil || taste.Profile.Version != 2 || taste.Profile.Rubrics[0].Criteria != tasteCommand.Rubrics[0].Criteria {
		t.Fatal("the actual complete taste profile did not survive SQLite and HTTP restart")
	}
	f.call(t, http.MethodPost, prefix+"/judgments", request, http.StatusOK, &answer)
	if !answer.Judgment.CacheHit || answer.Judgment.ID != original.ID {
		t.Fatal("the complete source cache did not survive SQLite and HTTP restart")
	}
	tasteCommand.CommandMeta = meta()
	tasteCommand.ExpectedVersion = taste.Profile.Version
	tasteCommand.Rubrics[0].Criteria = "Require the current fixed source versions before recommending the documented local CPU review"
	f.call(t, http.MethodPut, "/api/v1/attention-taste-profile", tasteCommand, http.StatusOK, &taste)
	var sourceFailure apierrors.ErrorEnvelope
	f.call(t, http.MethodPost, prefix+"/judgments", request, http.StatusConflict, &sourceFailure)
	if sourceFailure.Error == nil || sourceFailure.Error.Code != apierrors.VersionConflict {
		t.Fatal("the actual HTTP taste change did not reject its fixed old version")
	}
	f.call(t, http.MethodGet, path, nil, http.StatusOK, &answer)
	goal.Title = "Verify current local CPU parity and loading requirements"
	_, err = f.access.Goals.(*workspace.S2Service).PutGoal(ctx, human, goal.ID, workspace.GoalCommand{WorkspaceCommandMeta: workspace.WorkspaceCommandMeta{SchemaVersion: 1, RequestID: uuid.NewString(), IdempotencyKey: uuid.NewString(), ExpectedVersion: goal.Revision}, Goal: goal})
	if err != nil {
		t.Fatal(err)
	}
	f.call(t, http.MethodPost, prefix+"/judgments", request, http.StatusConflict, &sourceFailure)
	if sourceFailure.Error == nil || sourceFailure.Error.Code != apierrors.ContextStale {
		t.Fatal("the actual HTTP Goal change did not reject its fixed old revision independently")
	}
	t.Log("actual judgment handlers, fixed Goal/taste source references, independent source changes, original human/CSRF guard, SQLite cache, five feedback kinds and restart passed; production registration remains a separate H integration")
}

func mustJudgmentObjects(t *testing.T, root string) *objects.Store {
	t.Helper()
	store, err := objects.New(root)
	if err != nil {
		t.Fatal(err)
	}
	return store
}
