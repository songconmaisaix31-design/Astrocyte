package sqlite

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net"
	"net/http"
	"net/http/cookiejar"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/songconmaisaix31-design/Astrocyte/internal/adapters/agents"
	"github.com/songconmaisaix31-design/Astrocyte/internal/adapters/httpapi"
	"github.com/songconmaisaix31-design/Astrocyte/internal/apierrors"
	attention "github.com/songconmaisaix31-design/Astrocyte/internal/attention/app"
	swarm "github.com/songconmaisaix31-design/Astrocyte/internal/swarm/app"
	workspace "github.com/songconmaisaix31-design/Astrocyte/internal/workspace/app"
	"github.com/songconmaisaix31-design/Astrocyte/internal/workspace/domain"
)

func acceptedArtifactActualSource(t *testing.T, db *DB) (*AcceptedArtifactCompletionSource, workspace.LocalProjects) {
	t.Helper()
	projects := workspace.NewLocalProjectService(db, nil, agents.ProjectFiles{}, nil)
	s2 := workspace.NewS2Service(NewWorkspaceRepository(db), agents.ProjectFiles{}, nil)
	grants, err := NewSwarmApprovalBridge(db, s2)
	if err != nil {
		t.Fatal(err)
	}
	repository, err := NewSwarmRepository(db)
	if err != nil {
		t.Fatal(err)
	}
	ledger, err := swarm.NewSwarmLedger(repository, grants)
	if err != nil {
		t.Fatal(err)
	}
	source, err := NewAcceptedArtifactCompletionSource(db, ledger, projects)
	if err != nil {
		t.Fatal(err)
	}
	return source, projects
}

func acceptedArtifactExpectCode(t *testing.T, err error, code apierrors.Code) {
	t.Helper()
	var failure *apierrors.ServiceError
	if !errors.As(err, &failure) || failure.Code != code {
		t.Fatalf("expected %s from the actual interface: %v", code, err)
	}
}

func acceptedArtifactAssertNoSource(t *testing.T, db *DB) {
	t.Helper()
	for _, table := range []string{"swarm_missions", "swarm_artifacts", "foundation_outbox", "attention_engine_events", "attention_engine_experiences", "experience_completion_sources", "local_agent_grants", "local_agent_sessions"} {
		var count int
		if err := db.conn.QueryRow("SELECT COUNT(*) FROM " + table).Scan(&count); err != nil || count != 0 {
			t.Fatal("source rejection created an execution, adoption or experience record", table, count, err)
		}
	}
}

func TestAcceptedArtifactCompletionActualInterfacesAndLifecycle(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	directory := acceptedExperienceDirectory(t)
	path := filepath.Join(directory, "state.sqlite")
	db := openAttentionDB(t, path)
	source, projects := acceptedArtifactActualSource(t, db)
	api := httpapi.NewServer(httpapi.Config{
		Logger: slog.New(slog.NewTextHandler(io.Discard, nil)),
		Services: httpapi.Services{
			LocalProjects: projects,
			Attention:     attention.NewAttentionService(NewAttentionRepository(db), nil, nil, attention.ServiceOptions{}),
		},
	})
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	origin := "http://" + listener.Addr().String()
	done := make(chan error, 1)
	go func() { done <- api.Serve(listener) }()
	stopped := false
	stop := func() {
		if stopped {
			return
		}
		shutdown, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := api.Shutdown(shutdown); err != nil {
			t.Fatal(err)
		}
		if err := <-done; !errors.Is(err, http.ErrServerClosed) {
			t.Fatal("production HTTP server did not shut down normally", err)
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
	response, err := client.Get(origin + "/api/v1/auth/session")
	if err != nil {
		t.Fatal(err)
	}
	var session struct {
		ID   string `json:"actor_id"`
		Kind string `json:"actor_kind"`
		CSRF string `json:"csrf_token"`
	}
	decodeErr := json.NewDecoder(response.Body).Decode(&session)
	if err := errors.Join(decodeErr, response.Body.Close()); err != nil || response.StatusCode != http.StatusOK || session.ID == "" || session.Kind != "human" || session.CSRF == "" {
		t.Fatal("actual authenticated human bootstrap failed", err)
	}
	principal := attention.Principal{ID: session.ID, Kind: session.Kind}
	post := func(route string, command any, expected int, output any) {
		data, err := json.Marshal(command)
		if err != nil {
			t.Fatal(err)
		}
		request, err := http.NewRequestWithContext(ctx, http.MethodPost, origin+route, bytes.NewReader(data))
		if err != nil {
			t.Fatal(err)
		}
		request.Header.Set("Content-Type", "application/json")
		request.Header.Set("Idempotency-Key", uuid.NewString())
		request.Header.Set("X-CSRF-Token", session.CSRF)
		response, err := client.Do(request)
		if err != nil {
			t.Fatal(err)
		}
		decodeErr := json.NewDecoder(response.Body).Decode(output)
		if err := errors.Join(decodeErr, response.Body.Close()); err != nil || response.StatusCode != expected {
			t.Fatal("actual project setup HTTP failed", route, response.StatusCode, err)
		}
	}
	var space attention.ProjectSpaceResult
	post("/api/v1/project-spaces", attention.ProjectSpaceCommand{
		CommandMeta: attention.CommandMeta{SchemaVersion: 1, RequestID: uuid.NewString(), ExpectedVersion: 1},
		Title:       "Owned acceptance source permission verification",
	}, http.StatusCreated, &space)
	root, err := filepath.Abs(filepath.Join("..", "..", ".."))
	if err != nil {
		t.Fatal(err)
	}
	root, err = filepath.EvalSymlinks(root)
	if err != nil {
		t.Fatal(err)
	}
	var created struct {
		Project domain.LocalProject `json:"project"`
	}
	post("/api/v1/local-projects", map[string]any{
		"schema_version": 1, "request_id": uuid.NewString(), "expected_version": 1,
		"name": "Current owned public source checkout", "root": root, "space_id": space.Space.ID,
	}, http.StatusOK, &created)
	project := created.Project
	if project.ID == "" || project.Root != root || project.SpaceID != space.Space.ID || project.Settings.Revision != 1 {
		t.Fatal("actual registered project differs from the owned HTTP command")
	}
	// 随机 ID 明确表示不存在的事件参数，不构造采用来源或模型输出。
	missingEventID := uuid.NewString()
	missing := attention.ExperienceCompletionEvidence{SourceKind: attention.ExperienceSourceAcceptedArtifact, LocalProjectID: project.ID, EventID: missingEventID}
	beforeWaits := db.conn.Stats().WaitCount
	t.Run("actual_registered_project_missing_adoption", func(t *testing.T) {
		if found, err := source.HasAcceptedArtifactCompletion(ctx, principal, project.ID, missingEventID); err != nil || found {
			t.Fatal("empty actual outbox was classified as an adoption", err)
		}
		result, err := source.ReadExperienceCompletion(ctx, principal, project.ID, missingEventID)
		acceptedArtifactExpectCode(t, err, apierrors.NotFound)
		if !reflect.DeepEqual(result, attention.ExperienceCompletionEvidence{}) {
			t.Fatal("missing adoption returned partial completion evidence")
		}
	})
	t.Run("current_human_identity_and_project_required", func(t *testing.T) {
		for _, caller := range []attention.Principal{{}, {ID: principal.ID, Kind: "agent"}, {ID: " ", Kind: "human"}} {
			_, err := source.HasAcceptedArtifactCompletion(ctx, caller, project.ID, missingEventID)
			acceptedArtifactExpectCode(t, err, apierrors.ScopeDenied)
			_, err = source.ReadExperienceCompletion(ctx, caller, project.ID, missingEventID)
			acceptedArtifactExpectCode(t, err, apierrors.ScopeDenied)
		}
		_, err := source.ReadExperienceCompletion(ctx, principal, uuid.NewString(), missingEventID)
		acceptedArtifactExpectCode(t, err, apierrors.ScopeDenied)
		_, err = source.HasAcceptedArtifactCompletion(ctx, principal, uuid.NewString(), missingEventID)
		acceptedArtifactExpectCode(t, err, apierrors.ScopeDenied)
		_, err = source.ReadExperienceCompletion(ctx, principal, project.ID, " ")
		acceptedArtifactExpectCode(t, err, apierrors.ValidationFailed)
	})
	t.Run("publication_requires_current_owned_transaction", func(t *testing.T) {
		acceptedArtifactExpectCode(t, source.AuthorizeExperienceCompletion(ctx, principal, missing), apierrors.ScopeDenied)
		err := NewAcceptedExperienceRepository(db).WithAcceptedExperienceTx(ctx, func(tx attention.AcceptedExperienceTx) error {
			return source.AuthorizeExperienceCompletion(tx.Context(), principal, missing)
		})
		acceptedArtifactExpectCode(t, err, apierrors.NotFound)
	})
	t.Run("transaction_reads_uncommitted_current_project_and_rolls_back", func(t *testing.T) {
		var pending domain.LocalProject
		err := NewAcceptedExperienceRepository(db).WithAcceptedExperienceTx(ctx, func(tx attention.AcceptedExperienceTx) error {
			var err error
			pending, err = projects.RegisterProject(tx.Context(), domain.Caller{ID: principal.ID, Kind: principal.Kind}, domain.RegisterProjectCommand{
				Name: "Owned uncommitted permission scope", Root: root, SpaceID: space.Space.ID,
			})
			if err != nil {
				return err
			}
			input := missing
			input.LocalProjectID = pending.ID
			if found, err := source.HasAcceptedArtifactCompletion(tx.Context(), principal, pending.ID, missingEventID); err != nil || found {
				t.Fatal("source location did not use the uncommitted current project", err)
			}
			return source.AuthorizeExperienceCompletion(tx.Context(), principal, input)
		})
		acceptedArtifactExpectCode(t, err, apierrors.NotFound)
		if pending.ID == "" {
			t.Fatal("actual project registration did not run inside the current transaction")
		}
		_, err = db.LoadProject(ctx, pending.ID)
		acceptedArtifactExpectCode(t, err, apierrors.NotFound)
	})
	t.Run("foreign_database_transaction_rejected", func(t *testing.T) {
		other := openAttentionDB(t, filepath.Join(directory, "other.sqlite"))
		err := NewAcceptedExperienceRepository(other).WithAcceptedExperienceTx(ctx, func(tx attention.AcceptedExperienceTx) error {
			if found, err := source.HasAcceptedArtifactCompletion(tx.Context(), principal, project.ID, missingEventID); err == nil || found || err.Error() != "workspace transaction cannot span databases" {
				t.Fatal("source location accepted another database's transaction", err)
			}
			return source.AuthorizeExperienceCompletion(tx.Context(), principal, missing)
		})
		if err == nil || !strings.Contains(err.Error(), "another database") {
			t.Fatal("foreign database transaction was not rejected", err)
		}
		acceptedArtifactAssertNoSource(t, other)
	})
	t.Run("standalone_swarm_transaction_rejected_without_waiting", func(t *testing.T) {
		repository, err := NewSwarmRepository(db)
		if err != nil {
			t.Fatal(err)
		}
		err = repository.Within(ctx, func(tx swarm.Transaction) error {
			current := tx.(*swarmTransaction).ctx
			_, err := source.HasAcceptedArtifactCompletion(current, principal, project.ID, missingEventID)
			acceptedArtifactExpectCode(t, err, apierrors.ScopeDenied)
			_, err = source.ReadExperienceCompletion(current, principal, project.ID, missingEventID)
			acceptedArtifactExpectCode(t, err, apierrors.ScopeDenied)
			return source.AuthorizeExperienceCompletion(current, principal, missing)
		})
		acceptedArtifactExpectCode(t, err, apierrors.ScopeDenied)
	})
	t.Run("completed_transaction_connection_rejected", func(t *testing.T) {
		var ended context.Context
		if err := NewAcceptedExperienceRepository(db).WithAcceptedExperienceTx(ctx, func(tx attention.AcceptedExperienceTx) error {
			ended = tx.Context()
			return nil
		}); err != nil {
			t.Fatal(err)
		}
		if err := source.AuthorizeExperienceCompletion(ended, principal, missing); !errors.Is(err, sql.ErrConnDone) {
			t.Fatal("ended SQLite transaction connection was reused", err)
		}
		if found, err := source.HasAcceptedArtifactCompletion(ended, principal, project.ID, missingEventID); !errors.Is(err, sql.ErrConnDone) || found {
			t.Fatal("source location reused an ended SQLite transaction", err)
		}
	})
	t.Run("actual_context_cancellation_propagates", func(t *testing.T) {
		canceled, cancel := context.WithCancel(ctx)
		cancel()
		if found, err := source.HasAcceptedArtifactCompletion(canceled, principal, project.ID, missingEventID); !errors.Is(err, context.Canceled) || found {
			t.Fatal("source location ignored cancellation", err)
		}
		if _, err := source.ReadExperienceCompletion(canceled, principal, project.ID, missingEventID); !errors.Is(err, context.Canceled) {
			t.Fatal("source read ignored cancellation", err)
		}
		if err := source.AuthorizeExperienceCompletion(canceled, principal, missing); !errors.Is(err, context.Canceled) {
			t.Fatal("publication authorization ignored cancellation", err)
		}
	})
	t.Run("missing_actual_dependencies_rejected", func(t *testing.T) {
		for _, input := range []struct {
			db       *DB
			reader   swarm.AcceptedArtifactEvidenceReader
			projects workspace.LocalProjects
		}{{nil, source.reader, projects}, {db, nil, projects}, {db, source.reader, nil}} {
			_, err := NewAcceptedArtifactCompletionSource(input.db, input.reader, input.projects)
			acceptedArtifactExpectCode(t, err, apierrors.UnsupportedCapability)
		}
	})
	if ctx.Err() != nil || db.conn.Stats().WaitCount != beforeWaits {
		t.Fatal("source authorization waited for its caller's single SQLite connection", ctx.Err())
	}
	acceptedArtifactAssertNoSource(t, db)
	stop()
	if err := db.Close(); err != nil {
		t.Fatal(err)
	}
	if _, err := source.ReadExperienceCompletion(ctx, principal, project.ID, missingEventID); err == nil || !strings.Contains(err.Error(), "database is closed") {
		t.Fatal("closed actual SQLite source was accepted", err)
	}
	if found, err := source.HasAcceptedArtifactCompletion(ctx, principal, project.ID, missingEventID); err == nil || found || !strings.Contains(err.Error(), "database is closed") {
		t.Fatal("source location ignored actual SQLite closure", err)
	}
	reopened := openAttentionDB(t, path)
	stored, err := reopened.LoadProject(ctx, project.ID)
	if err != nil || !reflect.DeepEqual(stored, project) {
		t.Fatal("HTTP project registration did not persist through actual SQLite restart", err)
	}
	restarted, _ := acceptedArtifactActualSource(t, reopened)
	if found, err := restarted.HasAcceptedArtifactCompletion(ctx, principal, project.ID, missingEventID); err != nil || found {
		t.Fatal("SQLite restart created an adoption source location", err)
	}
	_, err = restarted.ReadExperienceCompletion(ctx, principal, project.ID, missingEventID)
	acceptedArtifactExpectCode(t, err, apierrors.NotFound)
	acceptedArtifactAssertNoSource(t, reopened)
	t.Log("actual HTTP bootstrap and project registration, same-database transaction reads, rollback, cancellation, shutdown and SQLite reopen verified; real adoption positive path NOT_RUN")
}
