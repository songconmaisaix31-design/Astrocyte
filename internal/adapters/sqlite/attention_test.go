package sqlite

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/songconmaisaix31-design/Astrocyte/internal/adapters/objects"
	"github.com/songconmaisaix31-design/Astrocyte/internal/apierrors"
	"github.com/songconmaisaix31-design/Astrocyte/internal/attention/app"
	"github.com/songconmaisaix31-design/Astrocyte/migrations"
)

func openAttentionDB(t *testing.T, path string) *DB {
	t.Helper()
	db, err := Open(path)
	if err != nil {
		t.Fatal(err)
	}
	if err = db.RunMigrations(context.Background(), migrations.FS); err != nil {
		db.Close()
		t.Fatal(err)
	}
	t.Cleanup(func() { db.Close() })
	return db
}

func materialFixture(t *testing.T) app.MaterialDetail {
	t.Helper()
	s, err := objects.New(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	digest, err := s.Publish(context.Background(), []byte("readable original"))
	if err != nil {
		t.Fatal(err)
	}
	return app.MaterialDetail{SchemaVersion: 1, Material: app.Material{ID: "m1", Version: 1, CurrentRevision: 1, Kind: "paper", Lifecycle: "active", Title: "actual source"}, Revisions: []app.MaterialRevision{{MaterialID: "m1", Revision: 1, SourceKey: "arxiv:2501.12345", ContentDigest: digest, ObjectRef: digest}}, Uses: []app.Usage{}, Distillations: []app.Distillation{}}
}

func TestAttentionRestartAtomicReceiptAndRollback(t *testing.T) {
	ctx := context.Background()
	path := filepath.Join(t.TempDir(), "state.sqlite")
	db := openAttentionDB(t, path)
	repo := NewAttentionRepository(db)
	m := materialFixture(t)
	receipt := app.Receipt{Digest: "request-digest", Result: json.RawMessage(`{"schema_version":1,"job_id":"j1","status":"queued"}`)}
	job := app.Job{JobID: "j1", Version: 1, DedupeKey: "source+configuration", OperationID: "stable-op", Status: "queued", Caller: app.Principal{ID: "local", Kind: "human"}, Payload: json.RawMessage(`{"source_locator":"https://arxiv.org/abs/2501.12345"}`), MaxAttempts: 3, DeadlineAt: time.Now().Add(time.Hour)}
	if err := repo.WithTx(ctx, func(tx app.AttentionTx) error {
		if err := tx.SaveMaterial(m, 0); err != nil {
			return err
		}
		if err := tx.SaveJob(job, 0); err != nil {
			return err
		}
		if err := tx.SaveReceipt("human", "ImportMaterial", "key", receipt); err != nil {
			return err
		}
		return tx.AppendEvent(app.OutboxEvent{ID: "e1", Type: "attention.material_imported", AggregateID: "m1", AggregateVersion: 1, OccurredAt: time.Now(), Payload: json.RawMessage(`{"material_id":"m1"}`)})
	}); err != nil {
		t.Fatal(err)
	}
	if err := db.Close(); err != nil {
		t.Fatal(err)
	}
	db = openAttentionDB(t, path)
	defer db.Close()
	repo = NewAttentionRepository(db)
	if err := repo.WithTx(ctx, func(tx app.AttentionTx) error {
		got, err := tx.FindMaterialBySourceKey("arxiv:2501.12345")
		if err != nil {
			return err
		}
		if got.Material.ID != "m1" || len(got.Revisions) != 1 {
			t.Fatal("material lost")
		}
		j, err := tx.LoadJob("j1")
		if err != nil {
			return err
		}
		if j.OperationID != job.OperationID || j.Caller.ID != job.Caller.ID || string(j.Payload) != string(job.Payload) {
			t.Fatal("job lost hidden state")
		}
		r, err := tx.LoadReceipt("human", "ImportMaterial", "key")
		if err != nil {
			return err
		}
		if string(r.Result) != string(receipt.Result) {
			t.Fatal("lost complete receipt")
		}
		return nil
	}); err != nil {
		t.Fatal(err)
	}
	rollback := errors.New("injected late failure")
	if err := repo.WithTx(ctx, func(tx app.AttentionTx) error {
		j, _ := tx.LoadJob("j1")
		j.Version++
		j.Status = "running"
		if err := tx.SaveJob(j, 1); err != nil {
			return err
		}
		if err := tx.SaveReceipt("human", "retry", "new", receipt); err != nil {
			return err
		}
		return rollback
	}); !errors.Is(err, rollback) {
		t.Fatal(err)
	}
	if err := repo.WithTx(ctx, func(tx app.AttentionTx) error {
		j, err := tx.LoadJob("j1")
		if err != nil {
			return err
		}
		if j.Status != "queued" || j.Version != 1 {
			t.Fatal("partial state committed")
		}
		_, err = tx.LoadReceipt("human", "retry", "new")
		var service *apierrors.ServiceError
		if !errors.As(err, &service) || service.Code != apierrors.NotFound {
			t.Fatal("receipt escaped rollback", err)
		}
		return nil
	}); err != nil {
		t.Fatal(err)
	}
	var count int
	if err := db.conn.QueryRow("SELECT count(*) FROM attention_outbox").Scan(&count); err != nil || count != 1 {
		t.Fatal(count, err)
	}
}

func TestAttentionTwoConnectionsDeduplicateAndCAS(t *testing.T) {
	ctx := context.Background()
	path := filepath.Join(t.TempDir(), "state.sqlite")
	db := openAttentionDB(t, path)
	defer db.Close()
	other, err := Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer other.Close()
	repos := []*AttentionRepository{NewAttentionRepository(db), NewAttentionRepository(other)}
	m := materialFixture(t)
	var wg sync.WaitGroup
	var mu sync.Mutex
	created := 0
	for i := range 16 {
		wg.Go(func() {
			err := repos[i%2].WithTx(ctx, func(tx app.AttentionTx) error {
				_, err := tx.FindMaterialBySourceKey(m.Revisions[0].SourceKey)
				if err == nil {
					return nil
				}
				var se *apierrors.ServiceError
				if !errors.As(err, &se) || se.Code != apierrors.NotFound {
					return err
				}
				if err := tx.SaveMaterial(m, 0); err != nil {
					return err
				}
				mu.Lock()
				created++
				mu.Unlock()
				return nil
			})
			if err != nil {
				t.Error(err)
			}
		})
	}
	wg.Wait()
	if created != 1 {
		t.Fatal("created", created)
	}
	job := app.Job{JobID: "j", Version: 1, Status: "queued", DedupeKey: "one", OperationID: "stable", Payload: json.RawMessage(`{}`)}
	if err := repos[0].WithTx(ctx, func(tx app.AttentionTx) error { return tx.SaveJob(job, 0) }); err != nil {
		t.Fatal(err)
	}
	won := 0
	for i := range 2 {
		wg.Go(func() {
			err := repos[i].WithTx(ctx, func(tx app.AttentionTx) error {
				j, err := tx.LoadJob("j")
				if err != nil {
					return err
				}
				j.Status = "running"
				j.Version = 2
				return tx.SaveJob(j, 1)
			})
			if err == nil {
				mu.Lock()
				won++
				mu.Unlock()
			} else {
				var se *apierrors.ServiceError
				if !errors.As(err, &se) || se.Code != apierrors.VersionConflict {
					t.Error(err)
				}
			}
		})
	}
	wg.Wait()
	if won != 1 {
		t.Fatal("CAS winners", won)
	}
}

func TestAttentionImmutableHistoryAndReceiptConflict(t *testing.T) {
	ctx := context.Background()
	db := openAttentionDB(t, filepath.Join(t.TempDir(), "state.sqlite"))
	defer db.Close()
	r := NewAttentionRepository(db)
	m := materialFixture(t)
	if err := r.WithTx(ctx, func(tx app.AttentionTx) error { return tx.SaveMaterial(m, 0) }); err != nil {
		t.Fatal(err)
	}
	if err := r.WithTx(ctx, func(tx app.AttentionTx) error {
		v, _ := tx.LoadMaterial("m1")
		v.Material.Version++
		v.Revisions[0].SourceLocator = "changed"
		return tx.SaveMaterial(v, 1)
	}); err == nil {
		t.Fatal("overwrote original revision")
	}
	if err := r.WithTx(ctx, func(tx app.AttentionTx) error {
		v, _ := tx.LoadMaterial("m1")
		v.Material.Version++
		v.Material.CurrentRevision = 2
		rev := v.Revisions[0]
		rev.Revision = 2
		v.Revisions = append(v.Revisions, rev)
		return tx.SaveMaterial(v, 1)
	}); err == nil {
		t.Fatal("duplicate digest accepted")
	}
	if err := r.WithTx(ctx, func(tx app.AttentionTx) error {
		v, _ := tx.LoadMaterial("m1")
		if v.Material.Version != 1 {
			t.Fatal("failed save advanced CAS")
		}
		return tx.SaveReceipt("actor", "command", "key", app.Receipt{Digest: "a", Result: json.RawMessage(`{"id":"m1"}`)})
	}); err != nil {
		t.Fatal(err)
	}
	for _, digest := range []string{"a", "b"} {
		err := r.WithTx(ctx, func(tx app.AttentionTx) error {
			return tx.SaveReceipt("actor", "command", "key", app.Receipt{Digest: digest, Result: json.RawMessage(`{"id":"m1"}`)})
		})
		if (err == nil) != (digest == "a") {
			t.Fatal(digest, err)
		}
	}
	// Different command/caller namespaces can use the same idempotency key.
	if err := r.WithTx(ctx, func(tx app.AttentionTx) error {
		return tx.SaveReceipt("agent", "command", "key", app.Receipt{Digest: "b", Result: json.RawMessage(`null`)})
	}); err != nil {
		t.Fatal(err)
	}
}

func TestAttentionJobCancellationFailureAndRetryPersist(t *testing.T) {
	ctx := context.Background()
	path := filepath.Join(t.TempDir(), "state.sqlite")
	db := openAttentionDB(t, path)
	r := NewAttentionRepository(db)
	j := app.Job{JobID: "cancelled", Version: 1, Status: "queued", DedupeKey: "source", OperationID: "op", Attempts: 1, MaxAttempts: 3, Payload: json.RawMessage(`{}`)}
	if err := r.WithTx(ctx, func(tx app.AttentionTx) error { return tx.SaveJob(j, 0) }); err != nil {
		t.Fatal(err)
	}
	for i, status := range []string{"cancelled", "queued", "failed"} {
		err := r.WithTx(ctx, func(tx app.AttentionTx) error {
			v, err := tx.LoadJob(j.JobID)
			if err != nil {
				return err
			}
			v.Version++
			v.Status = status
			v.CancelRequested = status == "cancelled"
			if status == "failed" {
				v.Error = &apierrors.ServiceError{Code: apierrors.ProviderUnavailable, Message: "network failed"}
			}
			return tx.SaveJob(v, i+1)
		})
		if err != nil {
			t.Fatal(err)
		}
	}
	db.Close()
	db = openAttentionDB(t, path)
	defer db.Close()
	r = NewAttentionRepository(db)
	if err := r.WithTx(ctx, func(tx app.AttentionTx) error {
		v, err := tx.FindJobByDedupeKey("source")
		if err != nil {
			return err
		}
		if v.Status != "failed" || v.Version != 4 || v.Error == nil || v.OperationID != "op" {
			return fmt.Errorf("failed job lost: %+v", v)
		}
		return nil
	}); err != nil {
		t.Fatal(err)
	}
}

func TestAttentionOpportunityFeedbackHistoryAndDistillationReuse(t *testing.T) {
	ctx := context.Background()
	db := openAttentionDB(t, filepath.Join(t.TempDir(), "state.sqlite"))
	r := NewAttentionRepository(db)
	first := app.Opportunity{ID: "o", Revision: 1, Version: 1, State: "ready_for_review", Title: "original candidate", NextStep: "inspect evidence"}
	detail := app.OpportunityDetail{SchemaVersion: 1, Opportunity: first, Revisions: []app.Opportunity{first}, Reviews: []app.Review{}}
	d := app.Distillation{ID: "d", ReuseKey: "fixed-input+question", Stage: "content", OutputRef: "published-output", OutputText: "actual output"}
	if err := r.WithTx(ctx, func(tx app.AttentionTx) error {
		if err := tx.SaveDistillation(d); err != nil {
			return err
		}
		return tx.SaveOpportunity(detail, 0)
	}); err != nil {
		t.Fatal(err)
	}
	if err := r.WithTx(ctx, func(tx app.AttentionTx) error {
		v, err := tx.LoadOpportunity("o")
		if err != nil {
			return err
		}
		v.Opportunity.Version = 2
		v.Opportunity.State = "deferred"
		v.Reviews = append(v.Reviews, app.Review{ID: "review", OpportunityID: "o", Revision: 1, Feedback: "later", Reason: "future work"})
		return tx.SaveOpportunity(v, 1)
	}); err != nil {
		t.Fatal(err)
	}
	if err := r.WithTx(ctx, func(tx app.AttentionTx) error {
		v, err := tx.LoadOpportunity("o")
		if err != nil {
			return err
		}
		if v.Revisions[0].State != "ready_for_review" || v.Opportunity.State != "deferred" || len(v.Reviews) != 1 {
			t.Fatal("feedback replaced history")
		}
		v.Opportunity.Version = 3
		v.Opportunity.Revision = 2
		v.Opportunity.Title = "new evidence"
		v.Revisions = append(v.Revisions, v.Opportunity)
		return tx.SaveOpportunity(v, 2)
	}); err != nil {
		t.Fatal(err)
	}
	if err := r.WithTx(ctx, func(tx app.AttentionTx) error {
		v, _ := tx.LoadOpportunity("o")
		v.Opportunity.Version = 4
		v.Revisions[0].Title = "overwritten"
		return tx.SaveOpportunity(v, 3)
	}); err == nil {
		t.Fatal("edited historical candidate")
	}
	if err := r.WithTx(ctx, func(tx app.AttentionTx) error {
		got, err := tx.FindDistillationByReuseKey(d.ReuseKey)
		if err != nil {
			return err
		}
		if got.ID != d.ID {
			t.Fatal("reuse missed")
		}
		return tx.SaveDistillation(d)
	}); err != nil {
		t.Fatal(err)
	}
	for _, id := range []string{"d", "other"} {
		changed := d
		changed.ID = id
		changed.OutputText = "different output"
		if err := r.WithTx(ctx, func(tx app.AttentionTx) error { return tx.SaveDistillation(changed) }); err == nil {
			t.Fatal("overwrote/distinct duplicate", id)
		}
	}
}

func TestAttentionReturnToEarlierContentWithoutNewRevision(t *testing.T) {
	ctx := context.Background()
	path := filepath.Join(t.TempDir(), "state.sqlite")
	db := openAttentionDB(t, path)
	r := NewAttentionRepository(db)
	store, err := objects.New(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	digestA, err := store.Publish(ctx, []byte("content A"))
	if err != nil {
		t.Fatal(err)
	}
	digestB, err := store.Publish(ctx, []byte("content B"))
	if err != nil {
		t.Fatal(err)
	}
	m := materialFixture(t)
	m.Revisions[0].ContentDigest = digestA
	m.Revisions[0].ObjectRef = digestA
	if err := r.WithTx(ctx, func(tx app.AttentionTx) error { return tx.SaveMaterial(m, 0) }); err != nil {
		t.Fatal(err)
	}
	if err := r.WithTx(ctx, func(tx app.AttentionTx) error {
		v, err := tx.LoadMaterial("m1")
		if err != nil {
			return err
		}
		v.Material.Version = 2
		v.Material.CurrentRevision = 2
		rev := v.Revisions[0]
		rev.Revision = 2
		rev.ContentDigest = digestB
		rev.ObjectRef = digestB
		v.Revisions = append(v.Revisions, rev)
		return tx.SaveMaterial(v, 1)
	}); err != nil {
		t.Fatal(err)
	}
	if err := r.WithTx(ctx, func(tx app.AttentionTx) error {
		v, err := tx.LoadMaterial("m1")
		if err != nil {
			return err
		}
		v.Material.Version = 3
		v.Material.CurrentRevision = 1
		return tx.SaveMaterial(v, 2)
	}); err != nil {
		t.Fatal("could not return to immutable prior content", err)
	}
	db.Close()
	db = openAttentionDB(t, path)
	r = NewAttentionRepository(db)
	if err := r.WithTx(ctx, func(tx app.AttentionTx) error {
		v, err := tx.LoadMaterial("m1")
		if err != nil {
			return err
		}
		if v.Material.CurrentRevision != 1 || v.Material.Version != 3 || len(v.Revisions) != 2 || v.Revisions[1].ContentDigest != digestB {
			t.Fatal("return lost history", v)
		}
		return nil
	}); err != nil {
		t.Fatal(err)
	}
	var count int
	if err := db.conn.QueryRow("SELECT count(*) FROM attention_material_revisions WHERE material_id='m1'").Scan(&count); err != nil || count != 2 {
		t.Fatal("return created duplicate revision", count, err)
	}
	for _, invalid := range []int{0, 3} {
		if err := r.WithTx(ctx, func(tx app.AttentionTx) error {
			v, err := tx.LoadMaterial("m1")
			if err != nil {
				return err
			}
			v.Material.Version = 4
			v.Material.CurrentRevision = invalid
			return tx.SaveMaterial(v, 3)
		}); err == nil {
			t.Fatal("accepted absent current revision", invalid)
		}
	}
	if _, err := store.Read(ctx, digestA); err != nil {
		t.Fatal(err)
	}
	if _, err := store.Read(ctx, digestB); err != nil {
		t.Fatal(err)
	}
}
