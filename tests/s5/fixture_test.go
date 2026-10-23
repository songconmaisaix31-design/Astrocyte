package s5

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
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/songconmaisaix31-design/Astrocyte/internal/adapters/agents"
	"github.com/songconmaisaix31-design/Astrocyte/internal/adapters/httpapi"
	"github.com/songconmaisaix31-design/Astrocyte/internal/adapters/importers"
	adapter "github.com/songconmaisaix31-design/Astrocyte/internal/adapters/judgment"
	"github.com/songconmaisaix31-design/Astrocyte/internal/adapters/objects"
	"github.com/songconmaisaix31-design/Astrocyte/internal/adapters/projectspacebridge"
	"github.com/songconmaisaix31-design/Astrocyte/internal/adapters/repositorycontext"
	"github.com/songconmaisaix31-design/Astrocyte/internal/adapters/sqlite"
	attention "github.com/songconmaisaix31-design/Astrocyte/internal/attention/app"
	"github.com/songconmaisaix31-design/Astrocyte/internal/judgment"
	workspace "github.com/songconmaisaix31-design/Astrocyte/internal/workspace/app"
	"github.com/songconmaisaix31-design/Astrocyte/internal/workspace/domain"
	"github.com/songconmaisaix31-design/Astrocyte/migrations"
)

type fixture struct {
	db            *sqlite.DB
	attention     *attention.Service
	projects      workspace.LocalProjects
	access        *adapter.ProjectAuthorization
	actor         judgment.Actor
	caller        domain.Caller
	client        *http.Client
	url, csrf     string
	server        *http.Server
	path, objects string
}

func commandMeta() attention.CommandMeta {
	return attention.CommandMeta{SchemaVersion: 1, RequestID: uuid.NewString(), IdempotencyKey: uuid.NewString(), ExpectedVersion: 1}
}

func newFixture(t *testing.T, path, objectRoot string) *fixture {
	t.Helper()
	ctx := context.Background()
	db, err := sqlite.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	if err := db.RunMigrations(ctx, migrations.FS); err != nil {
		db.Close()
		t.Fatal(err)
	}
	store, err := objects.New(objectRoot)
	if err != nil {
		db.Close()
		t.Fatal(err)
	}
	spaces := workspace.NewProjectSpaceService(sqlite.NewProjectSpaceRepository(db), workspace.ProjectSpaceOptions{})
	a := attention.NewAttentionService(sqlite.NewAttentionRepository(db), importers.NewReader([]string{filepath.Join(s5Root(t), ".verification", "s5")}), store, attention.ServiceOptions{ProjectSpaces: projectspacebridge.Attention{Spaces: spaces}})
	refs := adapter.AttentionReferences{Attention: a}
	projects := workspace.NewLocalProjectService(db, refs, agents.ProjectFiles{}, nil)
	environment, err := os.Hostname()
	if err != nil {
		t.Fatal(err)
	}
	snapshots, err := repositorycontext.New(db, db, repositorycontext.Options{Git: repositorycontext.Git{EnvironmentID: environment}, CacheRoot: filepath.Join(filepath.Dir(path), "projections"), SnapshotBudget: domain.RepositoryBudget{MaxFiles: 4, MaxBytes: 1024 * 1024, MaxResults: 1, MaxDepth: 1}})
	if err != nil {
		t.Fatal(err)
	}
	access := &adapter.ProjectAuthorization{Projects: db, References: refs, Policies: db, Snapshots: snapshots}
	transport := httpapi.NewServer(httpapi.Config{Services: httpapi.Services{Attention: a, ProjectSpaces: spaces, LocalProjects: projects}, Logger: slog.New(slog.NewTextHandler(io.Discard, nil))})
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	server := &http.Server{Handler: transport.Handler()}
	served := make(chan error, 1)
	go func() { served <- server.Serve(listener) }()
	jar, err := cookiejar.New(nil)
	if err != nil {
		t.Fatal(err)
	}
	f := &fixture{db: db, attention: a, projects: projects, access: access, client: &http.Client{Jar: jar}, url: "http://" + listener.Addr().String(), server: server, path: path, objects: objectRoot}
	t.Cleanup(func() {
		if err := server.Close(); err != nil {
			t.Error(err)
		}
		if err := <-served; err != http.ErrServerClosed {
			t.Error(err)
		}
		if err := db.Close(); err != nil {
			t.Error(err)
		}
	})
	var session struct {
		ID   string `json:"actor_id"`
		Kind string `json:"actor_kind"`
		CSRF string `json:"csrf_token"`
	}
	f.request(t, http.MethodGet, "/api/v1/auth/session", nil, http.StatusOK, &session)
	if session.ID == "" || session.Kind != "human" || session.CSRF == "" {
		t.Fatal("production HTTP bootstrap did not authenticate a real local human session")
	}
	f.actor = judgment.Actor{ID: session.ID, Kind: session.Kind}
	f.caller = domain.Caller{ID: session.ID, Kind: session.Kind}
	f.csrf = session.CSRF
	return f
}

func (f *fixture) request(t *testing.T, method, path string, body any, status int, out any) {
	t.Helper()
	var encoded []byte
	var err error
	if body != nil {
		encoded, err = json.Marshal(body)
		if err != nil {
			t.Fatal(err)
		}
	}
	r, err := http.NewRequest(method, f.url+path, bytes.NewReader(encoded))
	if err != nil {
		t.Fatal(err)
	}
	if body != nil {
		r.Header.Set("Content-Type", "application/json")
		r.Header.Set("X-CSRF-Token", f.csrf)
		r.Header.Set("Idempotency-Key", uuid.NewString())
	}
	response, err := f.client.Do(r)
	if err != nil {
		t.Fatal(err)
	}
	defer response.Body.Close()
	if response.StatusCode != status {
		t.Fatalf("actual production %s %s returned %d, expected %d", method, path, response.StatusCode, status)
	}
	if out != nil {
		if err := json.NewDecoder(response.Body).Decode(out); err != nil {
			t.Fatal(err)
		}
	}
}

func (f *fixture) importOfficialReadme(t *testing.T) judgment.EvidenceRef {
	t.Helper()
	root := s5Root(t)
	text, err := os.ReadFile(filepath.Join(root, ".verification", "s5", "upstream", "laya", "laya-ts", "README.md"))
	if err != nil {
		t.Fatal(err)
	}
	short := strings.Split(string(text), "## Node")[0]
	return f.importPublicText(t, short, "https://github.com/NandhaKishorM/laya/blob/432312b36714d8889eba405507b37eda25257d9c/laya-ts/README.md", "Official Laya export documentation", []string{"Export weights section"})
}

func (f *fixture) importPublicText(t *testing.T, text, locator, title string, spans []string) judgment.EvidenceRef {
	t.Helper()
	command := attention.ImportMaterialCommand{CommandMeta: commandMeta(), Kind: "text", Adapter: "manual", SourceLocator: locator, Title: title, ExportText: text, ContentDigest: judgment.DigestText(text), SourceSpans: spans}
	var job attention.ImportJobResult
	f.request(t, http.MethodPost, "/api/v1/materials/imports", command, http.StatusAccepted, &job)
	if _, err := f.attention.ProcessNextJob(context.Background()); err != nil {
		t.Fatal(err)
	}
	actual, err := f.attention.GetJob(context.Background(), attention.Principal{ID: f.actor.ID, Kind: f.actor.Kind}, job.JobID)
	if err != nil || actual.MaterialID == nil || actual.MaterialRevision == nil || actual.Status != "succeeded" {
		t.Fatal("actual original import did not complete", err)
	}
	return judgment.EvidenceRef{MaterialID: *actual.MaterialID, Revision: *actual.MaterialRevision, Digest: judgment.DigestText(text)}
}

func (f *fixture) project(t *testing.T, ref judgment.EvidenceRef, checkout string) domain.LocalProject {
	t.Helper()
	ctx := context.Background()
	p := attention.Principal{ID: f.actor.ID, Kind: f.actor.Kind}
	space, err := f.attention.CreateProjectSpace(ctx, p, attention.ProjectSpaceCommand{CommandMeta: commandMeta(), Title: "S5 fixed public context"})
	if err != nil {
		t.Fatal(err)
	}
	_, err = f.attention.ReferenceMaterial(ctx, p, space.Space.ID, attention.ReferenceMaterialCommand{CommandMeta: commandMeta(), MaterialID: ref.MaterialID, Revision: ref.Revision})
	if err != nil {
		t.Fatal(err)
	}
	project, err := f.projects.RegisterProject(ctx, f.caller, domain.RegisterProjectCommand{Name: "S5 public fixed Astrocyte", Root: filepath.Join(s5Root(t), ".verification", "s5", "projects", checkout), SpaceID: space.Space.ID})
	if err != nil {
		t.Fatal(err)
	}
	settings := project.Settings
	settings.AllowDirectory = true
	settings.AllowedSubdirs = []string{"."}
	project, err = f.projects.SetProjectSettings(ctx, f.caller, project.ID, domain.SettingsCommand{ExpectedRevision: project.Settings.Revision, Settings: settings})
	if err != nil {
		t.Fatal(err)
	}
	return project
}

func (f *fixture) policy(t *testing.T, project domain.LocalProject, questionType judgment.QuestionType, backend judgment.Backend, modelRevision string) judgment.Policy {
	t.Helper()
	snapshot, err := f.access.Snapshot(context.Background(), f.actor, project.ID, []string{"README.md"})
	if err != nil {
		t.Fatal(err)
	}
	return judgment.Policy{Revision: 1, ProjectID: project.ID, ProjectSettingsRevision: project.Settings.Revision, WorktreeID: snapshot.WorktreeID, SnapshotID: snapshot.Generation, Snapshot: snapshot, SnapshotFiles: []string{"README.md"}, QuestionType: questionType, Backend: backend, Languages: []string{"en"}, Versions: judgment.Versions{Goal: "public-goal-v1", Taste: "manual-review-v1", Policy: "policy-v1", Model: modelRevision, AccessEpoch: "public-context-epoch-v1"}, Budget: judgment.Budget{TimeoutMS: 60000, MaxInputBytes: 131072, MaxInputTokens: 2048, MaxOutputTokens: 0}}
}
