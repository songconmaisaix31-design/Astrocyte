package sqlite

import (
	"context"
	"encoding/json"
	"path/filepath"
	"testing"
	"time"

	"github.com/google/uuid"
	swarmapp "github.com/songconmaisaix31-design/Astrocyte/internal/swarm/app"
	"github.com/songconmaisaix31-design/Astrocyte/internal/swarm/domain"
)

func TestSwarmActualPrivateLeaseTuplePersistsWithoutPublicSerialization(t *testing.T) {
	ctx := context.Background()
	path := filepath.Join(swarmTestDirectory(t), "state.sqlite")
	db := swarmTestDB(t, path)
	repo, err := NewSwarmRepository(db)
	if err != nil {
		t.Fatal(err)
	}
	m, w := swarmSQLParameters(t)
	l := domain.Lease{WorkItemID: w.ID, MissionID: m.ID, ProjectID: m.Grant.ProjectID, HolderID: uuid.NewString(), Fence: 1, SessionID: uuid.NewString(), OperationID: uuid.NewString(), NativeSessionID: "native-reference-parameter", AdapterID: "adapter-reference-parameter", ControllerID: uuid.NewString(), StartupOperationID: uuid.NewString(), ExpiresAt: time.Now().Add(time.Minute)}
	if err = repo.Within(ctx, func(tx swarmapp.Transaction) error {
		if err := tx.SaveMission(m, 0); err != nil {
			return err
		}
		if err := tx.SaveWorkItem(w, 0); err != nil {
			return err
		}
		return tx.SaveLease(l)
	}); err != nil {
		t.Fatal(err)
	}
	check := func(got domain.Lease) {
		t.Helper()
		if got.NativeSessionID != l.NativeSessionID || got.AdapterID != l.AdapterID || got.ControllerID != l.ControllerID || got.StartupOperationID != l.StartupOperationID {
			t.Fatal("private original controller tuple was not retained")
		}
		data, err := json.Marshal(got)
		if err != nil {
			t.Fatal(err)
		}
		var public map[string]json.RawMessage
		if err = json.Unmarshal(data, &public); err != nil {
			t.Fatal(err)
		}
		for _, key := range []string{"native_session_id", "adapter_id", "controller_id", "startup_operation_id", "private_native_session_id", "private_adapter_id", "private_controller_id", "private_startup_operation_id"} {
			if _, exists := public[key]; exists {
				t.Fatal("private original tuple was exposed in public lease JSON")
			}
		}
	}
	if err = repo.Within(ctx, func(tx swarmapp.Transaction) error {
		stored, err := tx.Lease(w.ID)
		if err != nil {
			return err
		}
		check(stored)
		rows, err := tx.ActiveLeases(m.Grant.ProjectID)
		if err != nil {
			return err
		}
		if len(rows) != 1 {
			t.Fatal("actual lease was not listed")
		}
		check(rows[0])
		return nil
	}); err != nil {
		t.Fatal(err)
	}
	if err = db.Close(); err != nil {
		t.Fatal(err)
	}
	repo, err = NewSwarmRepository(swarmTestDB(t, path))
	if err != nil {
		t.Fatal(err)
	}
	stored, err := repo.LeaseForOperation(ctx, l.OperationID, l.SessionID)
	if err != nil {
		t.Fatal(err)
	}
	check(stored)
	if _, err = repo.LeaseForOperation(ctx, l.OperationID, uuid.NewString()); err == nil {
		t.Fatal("original operation was accepted under another session")
	}
}
