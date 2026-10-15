package sqlite

import (
	"context"
	"errors"
	"path/filepath"
	"reflect"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/songconmaisaix31-design/Astrocyte/internal/adapters/agents"
	"github.com/songconmaisaix31-design/Astrocyte/internal/apierrors"
	attention "github.com/songconmaisaix31-design/Astrocyte/internal/attention/app"
	workspace "github.com/songconmaisaix31-design/Astrocyte/internal/workspace/app"
	"github.com/songconmaisaix31-design/Astrocyte/internal/workspace/domain"
)

func TestLocalProjectUsesActualWorkspaceTransaction(t *testing.T) {
	root, err := filepath.Abs("../../..")
	if err != nil {
		t.Fatal(err)
	}
	for _, rollback := range []bool{false, true} {
		name := "commit"
		if rollback {
			name = "actual_constraint_rollback"
		}
		t.Run(name, func(t *testing.T) {
			directory := acceptedExperienceDirectory(t)
			db := openAttentionDB(t, filepath.Join(directory, "state.sqlite"))
			service := workspace.NewLocalProjectService(db, nil, agents.ProjectFiles{}, agents.NewRegistry(filepath.Join(directory, "native")))
			principal := domain.Caller{Kind: "human", ID: uuid.NewString()}
			ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
			defer cancel()
			before := db.Conn().Stats().WaitCount
			var project domain.LocalProject
			err := NewWorkspaceRepository(db).WithWorkspaceTx(ctx, func(tx workspace.WorkspaceTx) error {
				space, err := attention.NewAttentionService(NewAttentionRepository(db), nil, nil, attention.ServiceOptions{}).CreateProjectSpace(tx.Context(), attention.Principal{Kind: principal.Kind, ID: principal.ID}, attention.ProjectSpaceCommand{
					CommandMeta: attention.CommandMeta{SchemaVersion: 1, RequestID: uuid.NewString(), IdempotencyKey: uuid.NewString(), ExpectedVersion: 1},
					Title:       "Actual public project transaction validation",
				})
				if err != nil {
					return err
				}
				project, err = service.RegisterProject(tx.Context(), principal, domain.RegisterProjectCommand{Name: "Current public checkout", Root: root, SpaceID: space.Space.ID})
				if err != nil {
					return err
				}
				grant, err := service.GrantProjectAgent(tx.Context(), principal, project.ID, domain.ProjectGrant{AgentID: "transaction-observer", Actions: []string{"observe"}})
				if err != nil {
					return err
				}
				projects, err := service.ListProjects(tx.Context(), principal)
				if err != nil || len(projects) != 1 || projects[0].ID != project.ID {
					t.Fatal("actual project was not visible through the original transaction", err)
				}
				stored, err := db.LoadGrant(tx.Context(), project.ID, grant.AgentID)
				if err != nil || stored.ProjectID != project.ID || len(stored.Actions) != 1 || stored.Actions[0] != "observe" {
					t.Fatal("actual project grant was not visible through the original transaction", err)
				}
				sessions, err := service.ListNativeSessions(tx.Context(), principal, project.ID)
				if err != nil || sessions == nil || len(sessions) != 0 {
					t.Fatal("actual empty native session list changed", err)
				}
				_, err = db.LoadSession(tx.Context(), "explicit-missing-session-input")
				var missing *apierrors.ServiceError
				if !errors.As(err, &missing) || missing.Code != apierrors.NotFound {
					t.Fatal("missing session did not fail within the actual transaction", err)
				}
				if rollback {
					// 重新创建同一真实项目，触发现有主键约束并回滚全部动作。
					return db.SaveProject(tx.Context(), project, 0)
				}
				return nil
			})
			if rollback {
				var conflict *apierrors.ServiceError
				if !errors.As(err, &conflict) || conflict.Code != apierrors.VersionConflict {
					t.Fatal("actual project uniqueness conflict did not roll back", err)
				}
			} else if err != nil {
				t.Fatal(err)
			}
			if db.Conn().Stats().WaitCount != before {
				t.Fatal("local project authorization or sessions borrowed another connection")
			}
			want := 1
			if rollback {
				want = 0
			}
			for _, table := range []string{"attention_project_spaces", "attention_outbox", "attention_receipts", "local_agent_projects", "local_agent_grants"} {
				var count int
				if err := db.Conn().QueryRow("SELECT COUNT(*) FROM " + table).Scan(&count); err != nil || count != want {
					t.Fatal("actual local project transaction left unexpected rows", table, count, err)
				}
			}
			var nativeCount int
			if err := db.Conn().QueryRow("SELECT COUNT(*) FROM local_agent_sessions").Scan(&nativeCount); err != nil || nativeCount != 0 {
				t.Fatal("project transaction validation created a native source", nativeCount, err)
			}
		})
	}
}

func TestLocalProjectRejectsAnotherActualDatabaseTransaction(t *testing.T) {
	directory := acceptedExperienceDirectory(t)
	db := openAttentionDB(t, filepath.Join(directory, "first.sqlite"))
	other := openAttentionDB(t, filepath.Join(directory, "other.sqlite"))
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	err := NewWorkspaceRepository(db).WithWorkspaceTx(ctx, func(tx workspace.WorkspaceTx) error {
		if _, err := other.ListProjects(tx.Context()); err == nil {
			t.Fatal("project list accepted another database transaction")
		}
		if _, err := other.LoadProject(tx.Context(), "explicit-missing-project-input"); err == nil {
			t.Fatal("project read accepted another database transaction")
		}
		if _, err := other.LoadGrant(tx.Context(), "explicit-project-input", "explicit-agent-input"); err == nil {
			t.Fatal("grant read accepted another database transaction")
		}
		if _, err := other.ListSessions(tx.Context(), "explicit-project-input"); err == nil {
			t.Fatal("session list accepted another database transaction")
		}
		if _, err := other.LoadSession(tx.Context(), "explicit-session-input"); err == nil {
			t.Fatal("session read accepted another database transaction")
		}
		return nil
	})
	if err != nil || other.Conn().Stats().WaitCount != 0 {
		t.Fatal("actual cross-database rejection borrowed a connection", err)
	}
	canceled, stop := context.WithCancel(context.Background())
	stop()
	if _, err := db.ListProjects(canceled); !errors.Is(err, context.Canceled) {
		t.Fatal("actual cancellation was not retained", err)
	}
	if err := other.Close(); err != nil {
		t.Fatal(err)
	}
	if _, err := other.ListProjects(ctx); err == nil {
		t.Fatal("closed actual database returned project data")
	}
}

func TestNativeSessionUsesActualWorkspaceTransactionWithoutModel(t *testing.T) {
	directory := acceptedExperienceDirectory(t)
	db := openAttentionDB(t, filepath.Join(directory, "state.sqlite"))
	root, err := filepath.Abs("../../..")
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()
	principal := domain.Caller{Kind: "human", ID: uuid.NewString()}
	space, err := attention.NewAttentionService(NewAttentionRepository(db), nil, nil, attention.ServiceOptions{}).CreateProjectSpace(ctx, attention.Principal{Kind: principal.Kind, ID: principal.ID}, attention.ProjectSpaceCommand{
		CommandMeta: attention.CommandMeta{SchemaVersion: 1, RequestID: uuid.NewString(), IdempotencyKey: uuid.NewString(), ExpectedVersion: 1},
		Title:       "Actual native receipt transaction validation",
	})
	if err != nil {
		t.Fatal(err)
	}
	service := workspace.NewLocalProjectService(db, nil, agents.ProjectFiles{}, agents.NewRegistry(filepath.Join(directory, "native")))
	project, err := service.RegisterProject(ctx, principal, domain.RegisterProjectCommand{Name: "Actual public checkout native receipt", Root: root, SpaceID: space.Space.ID})
	if err != nil {
		t.Fatal(err)
	}
	settings := project.Settings
	settings.ExternalModelCLI, settings.AllowedActions = "codex", []string{"start", "stop"}
	project, err = service.SetProjectSettings(ctx, principal, project.ID, domain.SettingsCommand{ExpectedRevision: settings.Revision, Settings: settings})
	if err != nil {
		t.Fatal(err)
	}
	principal.OperationID = uuid.NewString()
	session, err := service.ProbeNativeCLI(ctx, principal, project.ID, "codex")
	t.Cleanup(func() {
		if session.ID == "" || session.StopConfirmed {
			return
		}
		cleanup, stop := context.WithTimeout(context.Background(), 10*time.Second)
		defer stop()
		observation, err := service.StopNativeSession(cleanup, principal, project.ID, session.ID)
		if err != nil || !observation.StopConfirmed {
			t.Error("actual empty-input native cleanup is unconfirmed", err)
		}
	})
	if err != nil || session.NativeID == "" || !session.StopConfirmed {
		t.Fatal("actual model-free controller did not generate a positively stopped session", err)
	}
	before := db.Conn().Stats().WaitCount
	err = NewWorkspaceRepository(db).WithWorkspaceTx(ctx, func(tx workspace.WorkspaceTx) error {
		if err := db.SaveSession(tx.Context(), session); err != nil {
			return err
		}
		stored, err := db.LoadSession(tx.Context(), session.ID)
		if err != nil || !reflect.DeepEqual(stored, session) {
			t.Fatal("actual native receipt changed inside the original transaction", err)
		}
		listed, err := db.ListSessions(tx.Context(), project.ID)
		if err != nil || len(listed) != 1 || !reflect.DeepEqual(listed[0], session) {
			t.Fatal("actual native receipt list differs within the original transaction", err)
		}
		return nil
	})
	if err != nil || db.Conn().Stats().WaitCount != before {
		t.Fatal("actual native receipt write borrowed another connection", err)
	}
	if err := db.Close(); err != nil {
		t.Fatal(err)
	}
	reopened := openAttentionDB(t, filepath.Join(directory, "state.sqlite"))
	stored, err := reopened.LoadSession(ctx, session.ID)
	if err != nil || !reflect.DeepEqual(stored, session) {
		t.Fatal("actual native receipt changed after SQLite reopen", err)
	}
	t.Logf("actual_native_id=%s operation=%s stop_confirmed=true model_turns=0", session.NativeID, principal.OperationID)
}
