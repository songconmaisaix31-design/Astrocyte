package sqlite

import (
	"context"
	"path/filepath"
	"testing"
	"time"

	"github.com/songconmaisaix31-design/Astrocyte/internal/apierrors"
	"github.com/songconmaisaix31-design/Astrocyte/internal/attention/app"
)

type sqliteListingFunc func(context.Context, app.TrackingSource, string, int) (app.ListingPage, error)

func (f sqliteListingFunc) ReadPage(ctx context.Context, source app.TrackingSource, cursor string, limit int) (app.ListingPage, error) {
	return f(ctx, source, cursor, limit)
}
func trackingMeta(key string, version int) app.CommandMeta {
	return app.CommandMeta{SchemaVersion: 1, ExpectedVersion: version, IdempotencyKey: key, RequestID: key}
}

func TestTrackingRealSQLiteRestartKeepsPartialMetadataAndResumesCursor(t *testing.T) {
	ctx := context.Background()
	path := filepath.Join(t.TempDir(), "tracking.sqlite")
	db := openAttentionDB(t, path)
	repo := NewAttentionRepository(db)
	human := app.Principal{ID: "test-human", Kind: "human"}
	fail := true
	cursorCalls := []string{}
	reader := sqliteListingFunc(func(_ context.Context, _ app.TrackingSource, cursor string, _ int) (app.ListingPage, error) {
		cursorCalls = append(cursorCalls, cursor)
		if cursor == "" {
			next := "2"
			return app.ListingPage{Items: []app.ListingMetadata{{ExternalID: "2:123", Locator: "https://www.bilibili.com/video/BV1PReT6EEqR/", Title: "contract-local metadata", Description: "preserved exact description"}}, HasMore: true, NextCursor: &next}, nil
		}
		if fail {
			return app.ListingPage{}, &apierrors.ServiceError{Code: apierrors.ProviderUnavailable, Message: "contract-local safe failure", Retryable: true, RequiredAction: "retry_public_source"}
		}
		return app.ListingPage{Items: []app.ListingMetadata{{ExternalID: "2:456", Title: "unavailable", UnavailableReason: "provider video unavailable"}}}, nil
	})
	s := app.NewAttentionService(repo, nil, nil, app.ServiceOptions{ListingReader: reader, JobTimeout: time.Minute})
	source, err := s.BindTrackingSource(ctx, human, app.BindTrackingSourceCommand{CommandMeta: trackingMeta("bind", 1), Platform: "bilibili", SourceKind: "uploads", Locator: "3494358764489275"})
	if err != nil {
		t.Fatal(err)
	}
	queued, err := s.SyncTrackingSource(ctx, human, source.Source.ID, app.SyncTrackingSourceCommand{CommandMeta: trackingMeta("sync", source.Source.Version)})
	if err != nil {
		t.Fatal(err)
	}
	if _, err = s.ProcessNextJob(ctx); err != nil {
		t.Fatal(err)
	}
	before, err := s.GetTrackingSource(ctx, human, source.Source.ID)
	if err != nil || before.Source.Status != "failed" || len(before.Items) != 1 || !before.Items[0].Stale {
		t.Fatal("partial failure did not persist", err, before)
	}
	job, err := s.GetJob(ctx, human, queued.Jobs[0].JobID)
	if err != nil || job.Attempts != 1 || job.Status != "failed" {
		t.Fatal("first failure lost", err, job)
	}
	if err = db.Close(); err != nil {
		t.Fatal(err)
	}
	db = openAttentionDB(t, path)
	repo = NewAttentionRepository(db)
	s = app.NewAttentionService(repo, nil, nil, app.ServiceOptions{ListingReader: reader, JobTimeout: time.Minute})
	after, err := s.GetTrackingSource(ctx, human, source.Source.ID)
	if err != nil || after.Items[0].Metadata.Description != before.Items[0].Metadata.Description || after.Items[0].Revision != 1 || after.Source.LastSuccessAt != nil {
		t.Fatal("restart lost partial source or invented success", err)
	}
	fail = false
	retried, err := s.RetryJob(ctx, human, job.JobID, trackingMeta("retry-after-restart", job.Version))
	if err != nil || retried.OperationID != job.OperationID || !retried.DeadlineAt.Equal(job.DeadlineAt) {
		t.Fatal("restart reset operation/budget", err)
	}
	if _, err = s.ProcessNextJob(ctx); err != nil {
		t.Fatal(err)
	}
	after, err = s.GetTrackingSource(ctx, human, source.Source.ID)
	if err != nil || after.Source.Status != "succeeded" || len(after.Items) != 2 || after.Items[0].Stale || after.Items[1].Stale || len(cursorCalls) != 3 || cursorCalls[2] != "2" {
		t.Fatal("restart did not resume saved cursor/cache", err, after, cursorCalls)
	}
	var materials, failures int
	if err = db.conn.QueryRow("SELECT count(*) FROM attention_materials").Scan(&materials); err != nil {
		t.Fatal(err)
	}
	if err = db.conn.QueryRow("SELECT count(*) FROM attention_outbox WHERE type='attention.job_failed'").Scan(&failures); err != nil {
		t.Fatal(err)
	}
	if materials != 0 || failures != 1 {
		t.Fatalf("metadata became KB or first failure vanished: materials%d failures%d", materials, failures)
	}
}
