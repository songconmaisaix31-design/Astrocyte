package sqlite

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/songconmaisaix31-design/Astrocyte/internal/apierrors"
	"github.com/songconmaisaix31-design/Astrocyte/internal/attention/app"
	"github.com/songconmaisaix31-design/Astrocyte/internal/attention/domain"
)

// The processing port is contract_local; persistence and close/reopen are real.
type sqliteRecommendationProcessor struct {
	configuration string
	calls         int
	revoked       bool
	unknown       bool
}

func (p *sqliteRecommendationProcessor) ConfigurationID(context.Context, app.Principal, string, string) (string, error) {
	if p.revoked {
		return "", &apierrors.ServiceError{Code: apierrors.ScopeDenied, Message: "contract-local project permission revoked"}
	}
	return p.configuration, nil
}
func (p *sqliteRecommendationProcessor) Recommend(_ context.Context, input app.ListingRecommendationInput) (map[string]app.SourceRecommendation, error) {
	p.calls++
	if p.unknown {
		return nil, domain.ErrUnknown
	}
	output := map[string]app.SourceRecommendation{}
	for _, item := range input.Items {
		output[item.ExternalID] = app.SourceRecommendation{Text: "contract-local fixed metadata suggestion", Reason: "contract-local title and description", Provenance: app.Provenance{Processor: "test-cli", Mode: "contract_local"}}
	}
	return output, nil
}

func TestRecommendationSemanticReuseSQLiteRestartAndScope(t *testing.T) {
	for _, unknown := range []bool{false, true} {
		t.Run(fmt.Sprint(unknown), func(t *testing.T) {
			ctx := context.Background()
			path := filepath.Join(t.TempDir(), "recommendations.sqlite")
			db := openAttentionDB(t, path)
			human := app.Principal{ID: "test-human", Kind: "human"}
			processor := &sqliteRecommendationProcessor{configuration: "fixed-config", unknown: unknown}
			title := "contract-local metadata"
			reader := sqliteListingFunc(func(context.Context, app.TrackingSource, string, int) (app.ListingPage, error) {
				return app.ListingPage{Items: []app.ListingMetadata{
					{ExternalID: "2:1", Locator: "https://www.bilibili.com/video/BV1PReT6EEqR/", Title: title},
					{ExternalID: "2:2", Locator: "https://www.bilibili.com/video/BV11t411C7Lk/", Title: title},
				}}, nil
			})
			options := app.ServiceOptions{ListingReader: reader, ListingRecommender: processor}
			service := app.NewAttentionService(NewAttentionRepository(db), nil, nil, options)
			source, err := service.BindTrackingSource(ctx, human, app.BindTrackingSourceCommand{CommandMeta: trackingMeta("bind", 1), Platform: "bilibili", SourceKind: "favorites", Locator: "https://space.bilibili.com/3494358764489275/favlist?fid=2356677875"})
			if err != nil {
				t.Fatal(err)
			}
			refresh := func(key string) {
				t.Helper()
				source, err = service.GetTrackingSource(ctx, human, source.Source.ID)
				if err != nil {
					t.Fatal(err)
				}
				if _, err = service.SyncTrackingSource(ctx, human, source.Source.ID, app.SyncTrackingSourceCommand{CommandMeta: trackingMeta(key, source.Source.Version)}); err != nil {
					t.Fatal(err)
				}
				if _, err = service.ProcessNextJob(ctx); err != nil {
					t.Fatal(err)
				}
			}
			refresh("sync")
			selections := []app.SourceItemSelection{{ExternalID: "2:1", Revision: 1}, {ExternalID: "2:2", Revision: 1}}
			project, cli := "project-one", "cli-one"
			recommend := func(key string) (app.TrackingSourceResult, error) {
				t.Helper()
				source, err = service.GetTrackingSource(ctx, human, source.Source.ID)
				if err != nil {
					t.Fatal(err)
				}
				return service.RecommendSourceItems(ctx, human, source.Source.ID, app.RecommendSourceItemsCommand{CommandMeta: trackingMeta(key, source.Source.Version), ProjectID: project, CLI: cli, Items: selections})
			}
			first, err := recommend("first")
			if err != nil {
				t.Fatal(err)
			}
			selections[0], selections[1] = selections[1], selections[0]
			pending, err := recommend("pending-new-key-reversed-input")
			if err != nil || len(pending.Jobs) != 1 || pending.Jobs[0].JobID != first.Jobs[0].JobID {
				t.Fatal("pending same fixed inputs duplicated original work", pending.Jobs, err)
			}
			if _, err = service.ProcessNextJob(ctx); err != nil {
				t.Fatal(err)
			}
			original, err := service.GetJob(ctx, human, first.Jobs[0].JobID)
			if err != nil || processor.calls != 1 {
				t.Fatal("original processing count", processor.calls, err)
			}
			var oldData, oldPayload string
			if err = db.conn.QueryRow("SELECT data,payload FROM attention_jobs WHERE id=?", original.JobID).Scan(&oldData, &oldPayload); err != nil {
				t.Fatal(err)
			}
			if err = db.Close(); err != nil {
				t.Fatal(err)
			}
			db = openAttentionDB(t, path)
			defer db.Close()
			service = app.NewAttentionService(NewAttentionRepository(db), nil, nil, options)
			for _, key := range []string{"restart-new-key", "another-new-key"} {
				duplicate, err := recommend(key)
				if unknown {
					var failure *apierrors.ServiceError
					if !errors.As(err, &failure) || failure.Code != apierrors.DeliveryUnknown {
						t.Fatal("new key replayed original UNKNOWN", err)
					}
				} else if err != nil || len(duplicate.Jobs) != 1 || duplicate.Jobs[0].JobID != original.JobID || duplicate.Jobs[0].Status != "succeeded" {
					t.Fatal("successful same metadata did not reuse original result/job", duplicate.Jobs, err)
				}
				if _, err = service.ProcessNextJob(ctx); err != nil || processor.calls != 1 {
					t.Fatal("new request key repeated processing", processor.calls, err)
				}
			}
			var afterData, afterPayload string
			if err = db.conn.QueryRow("SELECT data,payload FROM attention_jobs WHERE id=?", original.JobID).Scan(&afterData, &afterPayload); err != nil || oldData != afterData || oldPayload != afterPayload {
				t.Fatal("reuse changed original job/result/history", err)
			}
			processor.revoked = true
			_, err = recommend("revoked-new-key")
			var denied *apierrors.ServiceError
			if !errors.As(err, &denied) || denied.Code != apierrors.ScopeDenied {
				t.Fatal("cached result bypassed current project permission", err)
			}
			processor.revoked = false
			if unknown {
				return
			}
			for i, change := range []func(){
				func() { processor.configuration = "changed-config" },
				func() { project = "project-two" },
				func() { cli = "cli-two" },
				func() { human.ID = "another-human" },
				func() {
					title = "contract-local changed metadata"
					refresh("changed-metadata-sync")
					selections[0].Revision, selections[1].Revision = 2, 2
				},
			} {
				change()
				fresh, err := recommend(fmt.Sprintf("changed-%d", i))
				if err != nil || len(fresh.Jobs) != 1 || fresh.Jobs[0].JobID == original.JobID {
					t.Fatal("changed metadata/configuration/scope reused different processing", fresh.Jobs, err)
				}
				if _, err = service.ProcessNextJob(ctx); err != nil || processor.calls != i+2 {
					t.Fatal("changed processing did not run once", processor.calls, err)
				}
			}
			var current app.Job
			err = NewAttentionRepository(db).WithTx(ctx, func(tx app.AttentionTx) error { current, err = tx.LoadJob(original.JobID); return err })
			if err != nil || !reflect.DeepEqual(original, current) {
				t.Fatal("new scope modified original completed work", err)
			}
			var count int
			if err = db.conn.QueryRow("SELECT count(*) FROM attention_materials").Scan(&count); err != nil || count != 0 {
				t.Fatal("metadata recommendation imported body", count, err)
			}
			t.Logf("original job %s unchanged; contract-local processor calls=%d", original.JobID, processor.calls)
		})
	}
}
