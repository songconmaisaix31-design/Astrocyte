package sqlite

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/songconmaisaix31-design/Astrocyte/internal/apierrors"
	swarmapp "github.com/songconmaisaix31-design/Astrocyte/internal/swarm/app"
	"github.com/songconmaisaix31-design/Astrocyte/internal/swarm/domain"
	workspaceapp "github.com/songconmaisaix31-design/Astrocyte/internal/workspace/app"
	"github.com/songconmaisaix31-design/Astrocyte/migrations"
)

func TestSwarmActualWorkspaceTransactionBorrowingAndRollback(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	dir := swarmTestDirectory(t)
	db := swarmTestDB(t, filepath.Join(dir, "shared.sqlite"))
	db.conn.SetMaxOpenConns(1)
	other := swarmTestDB(t, filepath.Join(dir, "other.sqlite"))
	repo, err := NewSwarmRepository(db)
	if err != nil {
		t.Fatal(err)
	}
	wrongDB, err := NewSwarmRepository(other)
	if err != nil {
		t.Fatal(err)
	}
	m, _ := swarmSQLParameters(t)
	rollback := errors.New("candidate transaction rejected")
	err = NewWorkspaceRepository(db).WithWorkspaceTx(ctx, func(wtx workspaceapp.WorkspaceTx) error {
		if err := wrongDB.Within(wtx.Context(), func(swarmapp.Transaction) error {
			t.Fatal("cross-database callback was invoked")
			return nil
		}); err == nil {
			t.Fatal("cross-database transaction was accepted")
		}
		return repo.Within(wtx.Context(), func(tx swarmapp.Transaction) error {
			if err := tx.SaveMission(m, 0); err != nil {
				return err
			}
			swarmAssertTransactionReads(t, wtx.Context(), repo, m.ID)
			swarmAssertTransactionReads(t, tx.(*swarmTransaction).ctx, repo, m.ID)
			return rollback
		})
	})
	if !errors.Is(err, rollback) {
		t.Fatal(err)
	}
	if _, err = repo.Load(ctx, m.ID); err == nil {
		t.Fatal("borrowed transaction left a partial mission after rollback")
	} else {
		var missing *apierrors.ServiceError
		if !errors.As(err, &missing) || missing.Code != apierrors.NotFound {
			t.Fatal(err)
		}
	}
	err = NewWorkspaceRepository(db).WithWorkspaceTx(ctx, func(wtx workspaceapp.WorkspaceTx) error {
		return repo.Within(wtx.Context(), func(tx swarmapp.Transaction) error { return tx.SaveMission(m, 0) })
	})
	if err != nil {
		t.Fatal(err)
	}
	if _, err = repo.Load(ctx, m.ID); err != nil {
		t.Fatal(err)
	}
}

func swarmAssertTransactionReads(t *testing.T, ctx context.Context, repo *SwarmRepository, missionID string) {
	t.Helper()
	if _, err := repo.Load(ctx, missionID); err != nil {
		t.Fatal(err)
	}
	rows, err := repo.List(ctx)
	if err != nil || len(rows) != 1 || rows[0].ID != missionID {
		t.Fatal("transaction did not expose its actual uncommitted mission", err)
	}
	if _, err = repo.Timeline(ctx, missionID); err != nil {
		t.Fatal(err)
	}
	if _, err = repo.ExecutionHistory(ctx, missionID); err != nil {
		t.Fatal(err)
	}
	if _, err = repo.ReadAcceptedArtifactEvent(ctx, uuid.NewString()); err == nil {
		t.Fatal("missing adoption event was accepted")
	} else {
		var missing *apierrors.ServiceError
		if !errors.As(err, &missing) || missing.Code != apierrors.NotFound {
			t.Fatal(err)
		}
	}
}

func swarmTestDirectory(t *testing.T) string {
	t.Helper()
	if err := os.MkdirAll(".verification", 0700); err != nil {
		t.Fatal(err)
	}
	dir, err := os.MkdirTemp(".verification", "swarm-sql-")
	if err != nil {
		t.Fatal(err)
	}
	abs, err := filepath.Abs(dir)
	if err != nil {
		t.Fatal(err)
	}
	return abs
}

func swarmTestDB(t *testing.T, path string) *DB {
	t.Helper()
	db, err := Open(path)
	if err != nil {
		t.Fatal(err)
	}
	if err = db.RunMigrations(context.Background(), migrations.FS); err != nil {
		db.Close()
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := db.Close(); err != nil {
			t.Error(err)
		}
	})
	return db
}

// 参数只验证 SQL 与领域规则，不表示存在批准或原生执行者。
func swarmSQLParameters(t *testing.T) (domain.Mission, domain.WorkItem) {
	t.Helper()
	now := time.Now().UTC()
	calls := int64(3)
	g := domain.Grant{ID: uuid.NewString(), Epoch: 1, ProposalID: uuid.NewString(), ProposalRevision: 1, ProjectID: uuid.NewString(), Root: "component-parameter", Goal: "verify atomic reservation", WritePaths: []string{"."}, Budget: domain.Budget{Calls: &calls}, Limits: domain.Limits{MaxWorkItems: 3, MaxAttempts: 3, MaxNoProgress: 3, MaxActive: 2, Deadline: now.Add(time.Hour)}, ExpiresAt: now.Add(time.Hour)}
	m, err := domain.NewMission(uuid.NewString(), g, now)
	if err != nil {
		t.Fatal(err)
	}
	w := domain.WorkItem{ID: uuid.NewString(), MissionID: m.ID, Version: 1, State: "open", Title: "verify transaction", Inputs: []domain.SourceVersion{{ID: "component-input", Revision: "1"}}, WritePaths: []string{"src"}}
	if err = w.Validate(m); err != nil {
		t.Fatal(err)
	}
	return m, w
}

func TestSwarmActualSQLiteTwoConnectionsAtomicClaimAndRestart(t *testing.T) {
	ctx := context.Background()
	dbPath := filepath.Join(swarmTestDirectory(t), "state.sqlite")
	a := swarmTestDB(t, dbPath)
	b := swarmTestDB(t, dbPath)
	ra, err := NewSwarmRepository(a)
	if err != nil {
		t.Fatal(err)
	}
	rb, err := NewSwarmRepository(b)
	if err != nil {
		t.Fatal(err)
	}
	m, w := swarmSQLParameters(t)
	if err = ra.Within(ctx, func(tx swarmapp.Transaction) error {
		if err := tx.SaveMission(m, 0); err != nil {
			return err
		}
		return tx.SaveWorkItem(w, 0)
	}); err != nil {
		t.Fatal(err)
	}
	start := make(chan struct{})
	results := make(chan error, 2)
	var wg sync.WaitGroup
	for _, repo := range []*SwarmRepository{ra, rb} {
		wg.Add(1)
		go func(r *SwarmRepository) {
			defer wg.Done()
			<-start
			results <- r.Within(ctx, func(tx swarmapp.Transaction) error {
				current, err := tx.Mission(m.ID)
				if err != nil {
					return err
				}
				old := current.Version
				current.Version++
				current.Reserved, err = current.Reserved.Add(domain.Resources{Calls: 1})
				if err != nil {
					return err
				}
				if err = tx.SaveMission(current, old); err != nil {
					return err
				}
				claimed := w
				claimed.Version = 2
				claimed.Fence = 1
				claimed.Attempts = 1
				claimed.State = "leased"
				if err = tx.SaveWorkItem(claimed, 1); err != nil {
					return err
				}
				l := domain.Lease{WorkItemID: w.ID, MissionID: m.ID, ProjectID: m.Grant.ProjectID, HolderID: uuid.NewString(), Fence: 1, OperationID: uuid.NewString(), Reserved: domain.Resources{Calls: 1}, WritePaths: w.WritePaths, ExpiresAt: time.Now().Add(-time.Second)}
				if err = tx.SaveLease(l); err != nil {
					return err
				}
				data, err := json.Marshal(l)
				if err != nil {
					return err
				}
				if err = tx.SaveReceipt("component", "claim", "same-key", swarmapp.Receipt{Digest: "component-input", Result: data}); err != nil {
					return err
				}
				return tx.Event(swarmapp.Event{ID: uuid.NewString(), Type: "swarm.component_claim", AggregateID: w.ID, Version: 2, At: time.Now().UTC(), Payload: l})
			})
		}(repo)
	}
	close(start)
	wg.Wait()
	close(results)
	success, conflicts := 0, 0
	for err := range results {
		if err == nil {
			success++
			continue
		}
		var rule *domain.RuleError
		if errors.As(err, &rule) && rule.Code == string(apierrors.VersionConflict) {
			conflicts++
			continue
		}
		t.Fatal(err)
	}
	if success != 1 || conflicts != 1 {
		t.Fatalf("success=%d conflicts=%d", success, conflicts)
	}
	if err = rb.Within(ctx, func(tx swarmapp.Transaction) error {
		current, err := tx.Mission(m.ID)
		if err != nil {
			return err
		}
		if current.Reserved.Calls != 1 {
			return errors.New("failed claimant left partial reservation")
		}
		leases, err := tx.ActiveLeases(m.Grant.ProjectID)
		if err != nil {
			return err
		}
		if len(leases) != 1 {
			return errors.New("unique lease violated")
		}
		stored, err := tx.WorkItem(w.ID)
		if err != nil {
			return err
		}
		if stored.Fence != 1 || stored.Attempts != 1 {
			return errors.New("failed claimant changed fence or attempt count")
		}
		_, found, err := tx.Receipt("component", "claim", "same-key")
		if err != nil {
			return err
		}
		if !found {
			return errors.New("claim receipt missing")
		}
		return nil
	}); err != nil {
		t.Fatal(err)
	}
	var events int
	if err = b.conn.QueryRowContext(ctx, "SELECT COUNT(*) FROM foundation_outbox WHERE aggregate_id=?", w.ID).Scan(&events); err != nil {
		t.Fatal(err)
	}
	if events != 1 {
		t.Fatalf("events=%d", events)
	}
	if err = a.Close(); err != nil {
		t.Fatal(err)
	}
	if err = b.Close(); err != nil {
		t.Fatal(err)
	}
	restarted := swarmTestDB(t, dbPath)
	rr, err := NewSwarmRepository(restarted)
	if err != nil {
		t.Fatal(err)
	}
	if err = rr.Within(ctx, func(tx swarmapp.Transaction) error {
		leases, err := tx.ActiveLeases(m.Grant.ProjectID)
		if err != nil {
			return err
		}
		if len(leases) != 1 || leases[0].ExpiresAt.After(time.Now()) {
			return errors.New("expired unknown lease was released on restart")
		}
		return nil
	}); err != nil {
		t.Fatal(err)
	}
}

func TestSwarmActualSQLiteRollbackPreservesBudgetAndFencing(t *testing.T) {
	ctx := context.Background()
	db := swarmTestDB(t, filepath.Join(swarmTestDirectory(t), "state.sqlite"))
	repo, err := NewSwarmRepository(db)
	if err != nil {
		t.Fatal(err)
	}
	m, w := swarmSQLParameters(t)
	if err = repo.Within(ctx, func(tx swarmapp.Transaction) error {
		if err := tx.SaveMission(m, 0); err != nil {
			return err
		}
		return tx.SaveWorkItem(w, 0)
	}); err != nil {
		t.Fatal(err)
	}
	constraint := domain.Failure(apierrors.BudgetExhausted, "component resource limit reached")
	err = repo.Within(ctx, func(tx swarmapp.Transaction) error {
		m.Version++
		m.Reserved.Calls = 2
		if err := tx.SaveMission(m, 1); err != nil {
			return err
		}
		w.Version++
		w.Fence = 1
		w.State = "leased"
		if err := tx.SaveWorkItem(w, 1); err != nil {
			return err
		}
		return constraint
	})
	if !errors.Is(err, constraint) {
		t.Fatalf("unexpected failure %v", err)
	}
	if err = repo.Within(ctx, func(tx swarmapp.Transaction) error {
		m, err := tx.Mission(m.ID)
		if err != nil {
			return err
		}
		w, err := tx.WorkItem(w.ID)
		if err != nil {
			return err
		}
		if m.Reserved.Calls != 0 || m.Version != 1 || w.Version != 1 || w.Fence != 0 || w.State != "open" {
			return errors.New("transaction rollback changed persistent aggregates")
		}
		return nil
	}); err != nil {
		t.Fatal(err)
	}
}

func TestSwarmActualSQLiteGrantUniqueConsumerAndStaleFence(t *testing.T) {
	ctx := context.Background()
	db := swarmTestDB(t, filepath.Join(swarmTestDirectory(t), "state.sqlite"))
	repo, err := NewSwarmRepository(db)
	if err != nil {
		t.Fatal(err)
	}
	m, w := swarmSQLParameters(t)
	l := domain.Lease{WorkItemID: w.ID, MissionID: m.ID, ProjectID: m.Grant.ProjectID, Fence: 7, OperationID: uuid.NewString()}
	if err = repo.Within(ctx, func(tx swarmapp.Transaction) error {
		if err := tx.SaveMission(m, 0); err != nil {
			return err
		}
		if err := tx.SaveWorkItem(w, 0); err != nil {
			return err
		}
		if err := tx.SaveLease(l); err != nil {
			return err
		}
		fresh, err := tx.Consume("component-consumer", "component-event")
		if err != nil {
			return err
		}
		if !fresh {
			return errors.New("first event was not fresh")
		}
		fresh, err = tx.Consume("component-consumer", "component-event")
		if err != nil {
			return err
		}
		if fresh {
			return errors.New("event replay was not deduplicated")
		}
		return nil
	}); err != nil {
		t.Fatal(err)
	}
	duplicate := m
	duplicate.ID = uuid.NewString()
	if err = repo.Within(ctx, func(tx swarmapp.Transaction) error { return tx.SaveMission(duplicate, 0) }); err == nil {
		t.Fatal("duplicate grant created a second mission")
	}
	if err = repo.Within(ctx, func(tx swarmapp.Transaction) error { return tx.DeleteLease(w.ID, 6) }); err == nil {
		t.Fatal("stale fencing token deleted lease")
	}
	if err = repo.Within(ctx, func(tx swarmapp.Transaction) error {
		current, err := tx.Lease(w.ID)
		if err != nil {
			return err
		}
		if current.Fence != 7 {
			return errors.New("stale fencing token changed lease")
		}
		return nil
	}); err != nil {
		t.Fatal(err)
	}
}
