package app

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/songconmaisaix31-design/Astrocyte/internal/apierrors"
	"github.com/songconmaisaix31-design/Astrocyte/internal/attention/domain"
)

func (tx *memoryTx) ListTrackingSources() ([]TrackingSource, error) {
	r := []TrackingSource{}
	for _, source := range tx.state.TrackingSources {
		r = append(r, source)
	}
	return r, nil
}
func (tx *memoryTx) LoadTrackingSource(id string) (TrackingSource, error) {
	r, ok := tx.state.TrackingSources[id]
	if !ok {
		return r, apierrors.NewNotFound("tracking_source", id)
	}
	return r, nil
}
func (tx *memoryTx) SaveTrackingSource(r TrackingSource, expected int) error {
	old, ok := tx.state.TrackingSources[r.ID]
	if r.Version != expected+1 || (ok && old.Version != expected) || (!ok && expected != 0) {
		return domain.ErrVersion
	}
	if tx.state.TrackingSources == nil {
		tx.state.TrackingSources = map[string]TrackingSource{}
	}
	tx.state.TrackingSources[r.ID] = r
	return nil
}
func (tx *memoryTx) ListSourceItems(id string) ([]SourceItem, error) {
	r := []SourceItem{}
	for _, item := range tx.state.SourceItems {
		if item.SourceID == id {
			r = append(r, item)
		}
	}
	return r, nil
}
func (tx *memoryTx) SaveSourceItem(item SourceItem) error {
	if tx.state.SourceItems == nil {
		tx.state.SourceItems = map[string]SourceItem{}
	}
	tx.state.SourceItems[item.SourceID+":"+item.ExternalID] = item
	return nil
}

type listingFunc func(context.Context, TrackingSource, string, int) (ListingPage, error)

func (f listingFunc) ReadPage(ctx context.Context, s TrackingSource, cursor string, limit int) (ListingPage, error) {
	return f(ctx, s, cursor, limit)
}

type recommendationFunc func(context.Context, ListingRecommendationInput) (map[string]SourceRecommendation, error)

func (f recommendationFunc) ConfigurationID(context.Context, Principal, string, string) (string, error) {
	return "contract-local-selected-cli", nil
}
func (f recommendationFunc) Recommend(ctx context.Context, input ListingRecommendationInput) (map[string]SourceRecommendation, error) {
	return f(ctx, input)
}

func bindFixture(t *testing.T, s *Service) TrackingSourceResult {
	t.Helper()
	r, err := s.BindTrackingSource(context.Background(), human, BindTrackingSourceCommand{CommandMeta: meta("bind", 1), Platform: "bilibili", SourceKind: "uploads", Locator: "3494358764489275"})
	if err != nil {
		t.Fatal(err)
	}
	return r
}
func metadataItem(id string) ListingMetadata {
	return ListingMetadata{ExternalID: id, Locator: "https://www.bilibili.com/video/BV1PReT6EEqR/", Title: "Contract-local provider title " + id, Description: "metadata only"}
}

func TestSourceIdentityAndPublicOnlyBinding(t *testing.T) {
	s, repo, _ := fixture(t)
	first := bindFixture(t, s)
	r, err := s.BindTrackingSource(context.Background(), human, BindTrackingSourceCommand{CommandMeta: meta("bind-alias", 1), Platform: "bilibili", SourceKind: "uploads", ExternalID: "3494358764489275", Locator: "https://space.bilibili.com/3494358764489275/upload/video?tracking=1"})
	if err != nil || r.Source.ID != first.Source.ID || len(repo.state.TrackingSources) != 1 {
		t.Fatal("account aliases duplicated binding", err)
	}
	_, err = s.BindTrackingSource(context.Background(), human, BindTrackingSourceCommand{CommandMeta: meta("second-account", 1), Platform: "bilibili", SourceKind: "uploads", Locator: "12345"})
	if err != nil || len(repo.state.TrackingSources) != 2 {
		t.Fatal("multiple accounts not retained", err)
	}
	_, err = s.BindTrackingSource(context.Background(), human, BindTrackingSourceCommand{CommandMeta: meta("douyin-self", 1), Platform: "douyin", SourceKind: "favorites", Locator: "https://www.douyin.com/user/self?showTab=favorite_collection"})
	errorCode(t, err, apierrors.ScopeDenied)
	_, err = s.BindTrackingSource(context.Background(), human, BindTrackingSourceCommand{CommandMeta: meta("mismatch", 1), Platform: "bilibili", SourceKind: "uploads", ExternalID: "1", Locator: "3494358764489275"})
	errorCode(t, err, apierrors.ValidationFailed)
	_, err = s.BindTrackingSource(context.Background(), human, BindTrackingSourceCommand{CommandMeta: meta("folder", 1), Platform: "bilibili", SourceKind: "favorites", Locator: "https://space.bilibili.com/3494358764489275/favlist?fid=456"})
	if err != nil || len(repo.state.TrackingSources) != 3 {
		t.Fatal("explicit folder not separate source", err)
	}
}

func TestMetadataCapRecommendationGateAndHumanSelection(t *testing.T) {
	ctx := context.Background()
	s, repo, _ := fixture(t)
	source := bindFixture(t, s)
	bodyCalls := 0
	s.sources = sourceFunc(func(_ context.Context, c ImportMaterialCommand) (ImportedSource, error) {
		bodyCalls++
		return ImportedSource{SourceKey: c.SourceLocator, SourceLocator: c.SourceLocator, Kind: "video", Text: "contract-local selected body"}, nil
	})
	readCalls := 0
	s.options.ListingReader = listingFunc(func(_ context.Context, _ TrackingSource, cursor string, limit int) (ListingPage, error) {
		readCalls++
		if limit != 100 {
			t.Fatalf("default cap%d", limit)
		}
		items := []ListingMetadata{}
		for i := 0; i < 100; i++ {
			items = append(items, metadataItem(fmt.Sprint(i)))
		}
		next := "next"
		return ListingPage{Items: items, HasMore: true, NextCursor: &next}, nil
	})
	_, err := s.SyncTrackingSource(ctx, human, source.Source.ID, SyncTrackingSourceCommand{CommandMeta: meta("sync", source.Source.Version)})
	if err != nil {
		t.Fatal(err)
	}
	if _, err = s.ProcessNextJob(ctx); err != nil {
		t.Fatal(err)
	}
	r, err := s.GetTrackingSource(ctx, human, source.Source.ID)
	if err != nil || len(r.Items) != 100 || !r.HasMore || bodyCalls != 0 || len(repo.state.Materials) != 0 || readCalls != 1 {
		t.Fatalf("metadata auto-imported or wrong cap %+v calls%d %v", r.Source, bodyCalls, err)
	}
	selection := []SourceItemSelection{{ExternalID: "0", Revision: 1}}
	_, err = s.SelectSourceItems(ctx, human, r.Source.ID, SelectSourceItemsCommand{CommandMeta: meta("select-before-rec", r.Source.Version), Items: selection})
	errorCode(t, err, apierrors.EvidenceMissing)
	modelCalls := 0
	s.options.ListingRecommender = recommendationFunc(func(_ context.Context, in ListingRecommendationInput) (map[string]SourceRecommendation, error) {
		modelCalls++
		if len(in.Items) != 1 || in.Items[0].Metadata.Description != "metadata only" || in.CLI != "selected-cli" {
			t.Fatal("recommendation did not receive fixed selected metadata")
		}
		return map[string]SourceRecommendation{"0": {Text: "Contract-local suggestion", Reason: "Based on title/description", Provenance: Provenance{Processor: "test-cli", Mode: "contract_local"}}}, nil
	})
	_, err = s.RecommendSourceItems(ctx, human, r.Source.ID, RecommendSourceItemsCommand{CommandMeta: meta("recommend", r.Source.Version), ProjectID: "permitted-project", CLI: "selected-cli", Items: selection})
	if err != nil {
		t.Fatal(err)
	}
	if _, err = s.ProcessNextJob(ctx); err != nil {
		t.Fatal(err)
	}
	r, _ = s.GetTrackingSource(ctx, human, r.Source.ID)
	if modelCalls != 1 || bodyCalls != 0 || len(repo.state.Materials) != 0 {
		t.Fatal("recommendation downloaded body")
	}
	_, err = s.RecommendSourceItems(ctx, human, r.Source.ID, RecommendSourceItemsCommand{CommandMeta: meta("recommend-new-key", r.Source.Version), ProjectID: "permitted-project", CLI: "selected-cli", Items: selection})
	if err != nil {
		t.Fatal(err)
	}
	s.ProcessNextJob(ctx)
	if modelCalls != 1 {
		t.Fatal("new form key repeated identical metadata/configuration processing", modelCalls)
	}
	r, _ = s.GetTrackingSource(ctx, human, r.Source.ID)
	selected, err := s.SelectSourceItems(ctx, human, r.Source.ID, SelectSourceItemsCommand{CommandMeta: meta("select", r.Source.Version), Items: selection})
	if err != nil || len(selected.Jobs) != 1 {
		t.Fatal("human selection did not queue import", err)
	}
	if _, err = s.ProcessNextJob(ctx); err != nil {
		t.Fatal(err)
	}
	if bodyCalls != 1 || len(repo.state.Materials) != 1 {
		t.Fatal("selected import absent or unselected bodies downloaded")
	}
	for _, event := range repo.state.Events {
		if event.Type == "mission_created" {
			t.Fatal("tracking created mission")
		}
	}
}

func TestPartialCacheKnownFailureResumeAndImmutableUnknown(t *testing.T) {
	ctx := context.Background()
	s, repo, _ := fixture(t)
	r := bindFixture(t, s)
	fail := true
	calls := []string{}
	s.options.ListingReader = listingFunc(func(_ context.Context, _ TrackingSource, cursor string, _ int) (ListingPage, error) {
		calls = append(calls, cursor)
		if cursor == "" {
			next := "page2"
			return ListingPage{Items: []ListingMetadata{metadataItem("one")}, HasMore: true, NextCursor: &next}, nil
		}
		if fail {
			return ListingPage{}, &apierrors.ServiceError{Code: apierrors.ProviderUnavailable, Message: "known safe read failure", Retryable: true, RequiredAction: "retry_public_source"}
		}
		return ListingPage{Items: []ListingMetadata{metadataItem("two")}}, nil
	})
	queued, err := s.SyncTrackingSource(ctx, human, r.Source.ID, SyncTrackingSourceCommand{CommandMeta: meta("partial-sync", r.Source.Version)})
	if err != nil {
		t.Fatal(err)
	}
	s.ProcessNextJob(ctx)
	r, _ = s.GetTrackingSource(ctx, human, r.Source.ID)
	job, _ := s.GetJob(ctx, human, queued.Jobs[0].JobID)
	if job.Status != "failed" || len(r.Items) != 1 || !r.Items[0].Stale || r.Source.LastSuccessAt != nil || r.Source.LastError == nil {
		t.Fatal("failed partial page became empty/success", job, r)
	}
	op, deadline := job.OperationID, job.DeadlineAt
	fail = false
	now := s.options.Clock()
	s.options.Clock = func() time.Time { return now.Add(2 * time.Second) }
	if err = s.retrySafeJobs(ctx); err != nil {
		t.Fatal(err)
	}
	s.ProcessNextJob(ctx)
	r, _ = s.GetTrackingSource(ctx, human, r.Source.ID)
	job, _ = s.GetJob(ctx, human, job.JobID)
	if job.Status != "succeeded" || job.Attempts != 2 || job.OperationID != op || !job.DeadlineAt.Equal(deadline) || len(r.Items) != 2 || r.Items[0].Stale || r.Items[1].Stale || len(calls) != 3 || calls[2] != "page2" {
		t.Fatal("safe retry lost cursor/cache/budget", job, r, calls)
	}
	modelCalls := 0
	s.options.ListingRecommender = recommendationFunc(func(context.Context, ListingRecommendationInput) (map[string]SourceRecommendation, error) {
		modelCalls++
		return nil, domain.ErrUnknown
	})
	rec, err := s.RecommendSourceItems(ctx, human, r.Source.ID, RecommendSourceItemsCommand{CommandMeta: meta("unknown-rec", r.Source.Version), ProjectID: "project", CLI: "cli", Items: []SourceItemSelection{{"one", 1}}})
	if err != nil {
		t.Fatal(err)
	}
	s.ProcessNextJob(ctx)
	unknown := repo.state.Jobs[rec.Jobs[0].JobID]
	if !unknown.DeliveryUnknown || unknown.Status != "failed" {
		t.Fatal("unknown outcome lost", unknown)
	}
	s.RecoverJobs(ctx)
	s.retrySafeJobs(ctx)
	s.ProcessNextJob(ctx)
	if modelCalls != 1 || repo.state.Jobs[unknown.JobID].Version != unknown.Version {
		t.Fatal("old unknown mutated or replayed")
	}
	_, err = s.RetryJob(ctx, human, unknown.JobID, meta("no-replay", unknown.Version))
	errorCode(t, err, apierrors.DeliveryUnknown)
	r, _ = s.GetTrackingSource(ctx, human, r.Source.ID)
	_, err = s.RecommendSourceItems(ctx, human, r.Source.ID, RecommendSourceItemsCommand{CommandMeta: meta("unknown-new-key", r.Source.Version), ProjectID: "project", CLI: "cli", Items: []SourceItemSelection{{"one", 1}}})
	errorCode(t, err, apierrors.DeliveryUnknown)
}

func TestBoundedPrefixDoesNotMarkUnobservedOlderCacheStale(t *testing.T) {
	ctx := context.Background()
	s, _, _ := fixture(t)
	r := bindFixture(t, s)
	s.options.ListingReader = listingFunc(func(_ context.Context, _ TrackingSource, cursor string, _ int) (ListingPage, error) {
		start, end := 0, 100
		if cursor == "older" {
			start, end = 100, 110
		}
		items := []ListingMetadata{}
		for i := start; i < end; i++ {
			items = append(items, metadataItem(fmt.Sprint(i)))
		}
		if cursor == "older" {
			return ListingPage{Items: items}, nil
		}
		next := "older"
		return ListingPage{Items: items, HasMore: true, NextCursor: &next}, nil
	})
	for i, cursor := range []string{"", "older", ""} {
		_, err := s.SyncTrackingSource(ctx, human, r.Source.ID, SyncTrackingSourceCommand{CommandMeta: meta(fmt.Sprintf("prefix-%d", i), r.Source.Version), Cursor: cursor})
		if err != nil {
			t.Fatal(err)
		}
		if _, err = s.ProcessNextJob(ctx); err != nil {
			t.Fatal(err)
		}
		r, _ = s.GetTrackingSource(ctx, human, r.Source.ID)
	}
	if len(r.Items) != 110 || !r.HasMore {
		t.Fatal("prefix observation lost cache/window", r.Source, len(r.Items))
	}
	for _, item := range r.Items {
		if item.Stale {
			t.Fatal("unobserved older cached row claimed stale", item.ExternalID)
		}
	}
}

func TestSavedFinalListingPageCompletesWithoutReader(t *testing.T) {
	ctx := context.Background()
	s, repo, _ := fixture(t)
	r := bindFixture(t, s)
	queued, err := s.SyncTrackingSource(ctx, human, r.Source.ID, SyncTrackingSourceCommand{CommandMeta: meta("saved-final-page", r.Source.Version)})
	if err != nil {
		t.Fatal(err)
	}
	job := repo.state.Jobs[queued.Jobs[0].JobID]
	job.Payload, err = json.Marshal(syncListingPayload{SourceID: r.Source.ID, Limit: 100, Count: 1, Seen: []string{"one"}, Done: true})
	if err != nil {
		t.Fatal(err)
	}
	repo.state.Jobs[job.JobID] = job
	repo.state.SourceItems = map[string]SourceItem{r.Source.ID + ":one": {SourceID: r.Source.ID, ExternalID: "one", Revision: 1, Metadata: metadataItem("one"), Stale: true}}
	s.options.ListingReader = nil
	if _, err = s.ProcessNextJob(ctx); err != nil {
		t.Fatal(err)
	}
	r, err = s.GetTrackingSource(ctx, human, r.Source.ID)
	job, _ = s.GetJob(ctx, human, job.JobID)
	if err != nil || job.Status != "succeeded" || r.Source.LastSuccessAt == nil || len(r.Items) != 1 || r.Items[0].Stale {
		t.Fatal("durable final page was discarded or fetched again", job, r, err)
	}
}

func TestStartupResumesCurrentFailureButNotSupersededFailure(t *testing.T) {
	for _, superseded := range []bool{false, true} {
		t.Run(fmt.Sprint(superseded), func(t *testing.T) {
			ctx := context.Background()
			s, repo, _ := fixture(t)
			r := bindFixture(t, s)
			queued, err := s.SyncTrackingSource(ctx, human, r.Source.ID, SyncTrackingSourceCommand{CommandMeta: meta("old-failed-sync", r.Source.Version)})
			if err != nil {
				t.Fatal(err)
			}
			old := repo.state.Jobs[queued.Jobs[0].JobID]
			old.Status, old.Attempts = "failed", 1
			old.Error = retryableInterruption()
			repo.state.Jobs[old.JobID] = old
			source := repo.state.TrackingSources[r.Source.ID]
			source.Status = "failed"
			repo.state.TrackingSources[source.ID] = source
			if superseded {
				newer := old
				newer.JobID, newer.Status = "newer-completed-sync", "succeeded"
				newer.CreatedAt = old.CreatedAt.Add(time.Second)
				newer.Error = nil
				repo.state.Jobs[newer.JobID] = newer
			}
			before := len(repo.state.Jobs)
			if err = s.SyncSourcesOnce(ctx); err != nil {
				t.Fatal(err)
			}
			expected := before
			if superseded {
				expected++
			}
			if len(repo.state.Jobs) != expected || repo.state.Jobs[old.JobID].Status != "failed" {
				t.Fatal("startup replayed old failure or suppressed the current sync", len(repo.state.Jobs), expected)
			}
		})
	}
}

func TestUnknownSourceReadIsNotResentByStartupOrNewRequestKey(t *testing.T) {
	ctx := context.Background()
	s, repo, _ := fixture(t)
	source := bindFixture(t, s)
	s.options.ListingReader = listingFunc(func(context.Context, TrackingSource, string, int) (ListingPage, error) {
		return ListingPage{}, serviceError(apierrors.DeliveryUnknown, "contract_local lost browser read result", "inspect_browser_read_outcome")
	})
	queued, err := s.SyncTrackingSource(ctx, human, source.Source.ID, SyncTrackingSourceCommand{CommandMeta: meta("first-unknown-read", source.Source.Version)})
	if err != nil {
		t.Fatal(err)
	}
	if _, err = s.ProcessNextJob(ctx); err != nil {
		t.Fatal(err)
	}
	old := repo.state.Jobs[queued.Jobs[0].JobID]
	data, _ := json.Marshal(old)
	current, err := s.GetTrackingSource(ctx, human, source.Source.ID)
	if err != nil || !old.DeliveryUnknown || current.Source.LastError == nil || current.Source.LastError.Code != apierrors.DeliveryUnknown {
		t.Fatal("unknown result lost", old, current, err)
	}
	if err = s.SyncSourcesOnce(ctx); err != nil || len(repo.state.Jobs) != 1 {
		t.Fatal("startup resent unknown read", err, len(repo.state.Jobs))
	}
	_, err = s.SyncTrackingSource(ctx, human, source.Source.ID, SyncTrackingSourceCommand{CommandMeta: meta("new-key-unknown-read", current.Source.Version)})
	errorCode(t, err, apierrors.DeliveryUnknown)
	after, _ := json.Marshal(repo.state.Jobs[old.JobID])
	if string(data) != string(after) || len(repo.state.Jobs) != 1 {
		t.Fatal("new key rewrote/resent unknown observation")
	}
}
