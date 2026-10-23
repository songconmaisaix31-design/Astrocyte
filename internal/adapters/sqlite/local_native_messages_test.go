package sqlite

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/songconmaisaix31-design/Astrocyte/internal/adapters/agents"
	"github.com/songconmaisaix31-design/Astrocyte/internal/adapters/objects"
	"github.com/songconmaisaix31-design/Astrocyte/internal/apierrors"
	attention "github.com/songconmaisaix31-design/Astrocyte/internal/attention/app"
	workspace "github.com/songconmaisaix31-design/Astrocyte/internal/workspace/app"
	"github.com/songconmaisaix31-design/Astrocyte/internal/workspace/domain"
)

func TestNativeMessageActualTransactionBoundaries(t *testing.T) {
	directory := acceptedExperienceDirectory(t)
	db := openAttentionDB(t, filepath.Join(directory, "state.sqlite"))
	other := openAttentionDB(t, filepath.Join(directory, "other.sqlite"))
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	if err := db.RequireNativeMessageOutsideTx(ctx); err != nil {
		t.Fatal(err)
	}
	before := db.Conn().Stats().WaitCount
	err := db.WithNativeMessageTx(ctx, func(txContext context.Context) error {
		if err := db.RequireNativeMessageOutsideTx(txContext); err == nil {
			t.Fatal("input publication or sending was allowed within the actual transaction")
		}
		if err := other.RequireNativeMessageOutsideTx(txContext); err == nil {
			t.Fatal("native input accepted another database's actual transaction")
		}
		if _, err := db.LoadNativeMessagePreparation(txContext, "explicit-missing-session-input", "explicit-missing-operation-input"); err == nil {
			t.Fatal("missing original native preparation became available")
		}
		called := false
		if err := other.WithNativeMessageTx(txContext, func(context.Context) error { called = true; return nil }); err == nil || called {
			t.Fatal("native preparation transaction crossed databases", err)
		}
		return nil
	})
	if err != nil || db.Conn().Stats().WaitCount != before || other.Conn().Stats().WaitCount != 0 {
		t.Fatal("actual native preparation transaction borrowed another connection", err)
	}
	canceled, stop := context.WithCancel(ctx)
	stop()
	if err := db.RequireNativeMessageOutsideTx(canceled); !errors.Is(err, context.Canceled) {
		t.Fatal("native preparation lost actual cancellation", err)
	}
	var count int
	if err := db.Conn().QueryRow("SELECT COUNT(*) FROM local_agent_sessions").Scan(&count); err != nil || count != 0 {
		t.Fatal("transaction boundary checks created a native source", count, err)
	}
}

func TestPreparedNativeMessageMissingOriginalReaderDoesNotPublish(t *testing.T) {
	directory := acceptedExperienceDirectory(t)
	db := openAttentionDB(t, filepath.Join(directory, "state.sqlite"))
	store, err := objects.New(filepath.Join(directory, "objects"))
	if err != nil {
		t.Fatal(err)
	}
	root, err := filepath.Abs("../../..")
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	caller := domain.Caller{Kind: "human", ID: uuid.NewString(), OperationID: uuid.NewString()}
	space, err := attention.NewAttentionService(NewAttentionRepository(db), nil, nil, attention.ServiceOptions{ProjectSpaces: attentionSpaces(db)}).CreateProjectSpace(ctx, attention.Principal{Kind: caller.Kind, ID: caller.ID}, attention.ProjectSpaceCommand{
		CommandMeta: attention.CommandMeta{SchemaVersion: 1, RequestID: uuid.NewString(), IdempotencyKey: uuid.NewString(), ExpectedVersion: 1},
		Title:       "Actual preparation dependency refusal",
	})
	if err != nil {
		t.Fatal(err)
	}
	service := workspace.NewLocalProjectService(db, nil, agents.ProjectFiles{}, agents.NewRegistry(filepath.Join(directory, "native"))).WithNativeMessageObjects(store)
	project, err := service.RegisterProject(ctx, caller, domain.RegisterProjectCommand{Name: "Current public checkout", Root: root, SpaceID: space.Space.ID})
	if err != nil {
		t.Fatal(err)
	}
	message, err := os.ReadFile(filepath.Join(root, "README.md"))
	if err != nil {
		t.Fatal(err)
	}
	_, err = service.PrepareNativeMessage(ctx, caller, project.ID, "explicit-missing-session-input", string(message))
	var unavailable *apierrors.ServiceError
	if !errors.As(err, &unavailable) || unavailable.Code != apierrors.UnsupportedCapability {
		t.Fatal("missing original controller reader did not fail before publishing", err)
	}
	if _, err := service.ReadPreparedNativeSession(ctx, "explicit-missing-session-input", caller.OperationID); !errors.As(err, &unavailable) || unavailable.Code != apierrors.UnsupportedCapability {
		t.Fatal("missing original controller reader yielded prepared evidence", err)
	}
	if _, err := service.SendNativeMessage(ctx, caller, project.ID, "explicit-missing-session-input", string(message)); !errors.As(err, &unavailable) || unavailable.Code != apierrors.UnsupportedCapability {
		t.Fatal("missing original controller reader allowed an unprepared send", err)
	}
	files, err := os.ReadDir(filepath.Join(directory, "objects"))
	if err != nil || len(files) != 0 {
		t.Fatal("unsupported native preparation published input", err)
	}
	var count int
	if err := db.Conn().QueryRow("SELECT COUNT(*) FROM local_agent_sessions").Scan(&count); err != nil || count != 0 {
		t.Fatal("unsupported native preparation created source metadata", count, err)
	}
}

func TestNativeMessageExplicitJSONRules(t *testing.T) {
	// 这些字符串只验证 JSON 规则，不作为会话或控制器来源。
	for _, parameter := range []string{"", "null", "[]", "{", `{"private_native_messages":null}`, `{"private_native_messages":[]}`, `{"private_native_messages":"invalid-object-parameter"}`} {
		if _, _, err := nativeMessageRecords([]byte(parameter)); err == nil {
			t.Fatalf("invalid JSON object parameter was accepted: %q", parameter)
		}
	}
	for _, parameter := range []string{`{}`, `{"private_native_messages":{}}`} {
		row, records, err := nativeMessageRecords([]byte(parameter))
		if err != nil || row == nil || records == nil || len(records) != 0 {
			t.Fatal("empty JSON object parameter changed", err)
		}
	}
}
