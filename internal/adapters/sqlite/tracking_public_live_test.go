package sqlite

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"reflect"
	"testing"
	"time"

	"github.com/songconmaisaix31-design/Astrocyte/internal/adapters/importers"
	"github.com/songconmaisaix31-design/Astrocyte/internal/attention/app"
)

type publicMetadataCapture struct {
	directory string
	requests  int
	observed  int
}

func (c *publicMetadataCapture) RoundTrip(request *http.Request) (*http.Response, error) {
	if request.URL.Host != "api.bilibili.com" || request.Header.Get("Cookie") != "" || request.Header.Get("Authorization") != "" {
		return nil, fmt.Errorf("expected credential-free public metadata request")
	}
	response, err := http.DefaultTransport.RoundTrip(request)
	if err != nil {
		return nil, err
	}
	raw, err := io.ReadAll(io.LimitReader(response.Body, (4<<20)+1))
	response.Body.Close()
	if err != nil {
		return nil, err
	}
	c.requests++
	if err = os.WriteFile(filepath.Join(c.directory, fmt.Sprintf("public-response-%02d.json", c.requests)), raw, 0600); err != nil {
		return nil, err
	}
	if request.URL.Path == "/x/v3/fav/resource/list" {
		var envelope struct {
			Data struct {
				Medias []json.RawMessage `json:"medias"`
			} `json:"data"`
		}
		if json.Unmarshal(raw, &envelope) == nil {
			c.observed += len(envelope.Data.Medias)
		}
	}
	response.Body = io.NopCloser(bytes.NewReader(raw))
	return response, nil
}

// Opt-in genuine public HTTP -> application -> SQLite, with no model, media,
// fake provider response or personal database. Keep raw responses and the DB.
func TestPublicFavoritesLiveSyncReuseAndRestart(t *testing.T) {
	root := os.Getenv("ASTROCYTE_TEST_PUBLIC_TRACKING_DIR")
	if root == "" {
		t.Skip("explicit public metadata-only acceptance directory required")
	}
	if !filepath.IsAbs(root) {
		t.Fatal("absolute acceptance directory required")
	}
	if err := os.MkdirAll(root, 0700); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(root, "tracking.sqlite")
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Fatal("use a fresh directory; preserve prior live observations")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()
	capture := &publicMetadataCapture{directory: root}
	reader := importers.NewPublicListingReader("")
	reader.Bilibili.Client.Transport = capture
	db := openAttentionDB(t, path)
	human := app.Principal{ID: "public-favorites-acceptance", Kind: "human"}
	service := app.NewAttentionService(NewAttentionRepository(db), nil, nil, app.ServiceOptions{ListingReader: reader, JobTimeout: 2 * time.Minute, MaxAttempts: 1})
	bound, err := service.BindTrackingSource(ctx, human, app.BindTrackingSourceCommand{CommandMeta: trackingMeta("bind-public-default", 1), Platform: "bilibili", SourceKind: "favorites", Locator: "https://space.bilibili.com/3494358764489275/favlist?fid=2356677875"})
	if err != nil {
		t.Fatal(err)
	}
	alias, err := service.BindTrackingSource(ctx, human, app.BindTrackingSourceCommand{CommandMeta: trackingMeta("bind-alias", 1), Platform: "bilibili", SourceKind: "favorites", OwnerID: "3494358764489275", Locator: "https://www.bilibili.com/medialist/detail/ml2356677875"})
	if err != nil || alias.Source.ID != bound.Source.ID {
		t.Fatal("same genuine folder created another binding", err)
	}
	var before app.TrackingSourceResult
	for round := 0; round < 2; round++ {
		observed := capture.observed
		queued, err := service.SyncTrackingSource(ctx, human, bound.Source.ID, app.SyncTrackingSourceCommand{CommandMeta: trackingMeta(fmt.Sprintf("sync-%d", round), bound.Source.Version)})
		if err != nil {
			t.Fatal(err)
		}
		if _, err = service.ProcessNextJob(ctx); err != nil {
			t.Fatal(err)
		}
		job, err := service.GetJob(ctx, human, queued.Jobs[0].JobID)
		if err != nil || job.Status != "succeeded" || job.Attempts != 1 {
			t.Fatalf("preserved live metadata failure: %+v %v", job, err)
		}
		bound, err = service.GetTrackingSource(ctx, human, bound.Source.ID)
		// Actual page overlap or omitted provider rows can produce fewer than
		// 100 unique entries; the limit bounds observations, not unique IDs.
		if err != nil || bound.Source.Status != "succeeded" || len(bound.Items) == 0 || len(bound.Items) > 100 || capture.observed-observed != 100 || !bound.HasMore || bound.NextCursor == nil || *bound.NextCursor != "6" {
			t.Fatalf("actual metadata cap/cursor: observed%d items%d source%+v %v", capture.observed-observed, len(bound.Items), bound.Source, err)
		}
		if round == 0 {
			before = bound
		} else {
			prior := map[string]app.SourceItem{}
			for _, item := range before.Items {
				prior[item.ExternalID] = item
			}
			for _, item := range bound.Items {
				old, exists := prior[item.ExternalID]
				if exists && reflect.DeepEqual(old.Metadata, item.Metadata) && item.Revision != old.Revision {
					t.Fatal("unchanged real metadata created new revision", item.ExternalID)
				}
			}
		}
		t.Logf("round%d items%d provider_rows%d has_more%t cursor%s", round+1, len(bound.Items), capture.observed-observed, bound.HasMore, *bound.NextCursor)
	}
	selected := false
	for _, item := range bound.Items {
		if item.Metadata.Locator == "https://www.bilibili.com/video/BV1PReT6EEqR/" && !item.Stale && item.Metadata.UnavailableReason == "" {
			selected = true
			t.Logf("original selected video external_id=%s revision=%d title=%q", item.ExternalID, item.Revision, item.Metadata.Title)
		}
		if item.Selected || item.ImportJobID != nil || item.Recommendation != nil {
			t.Fatal("metadata sync invented recommendation/selection/import")
		}
	}
	if !selected {
		t.Fatal("original selected video absent from current real first100")
	}
	for _, table := range []string{"attention_materials", "attention_distillations", "attention_opportunities"} {
		var count int
		if err = db.conn.QueryRow("SELECT count(*) FROM " + table).Scan(&count); err != nil || count != 0 {
			t.Fatal("metadata created KB/model/candidate state", table, count, err)
		}
	}
	data, _ := json.MarshalIndent(bound, "", "  ")
	if err = os.WriteFile(filepath.Join(root, "actual-source-after-repeat.json"), data, 0600); err != nil {
		t.Fatal(err)
	}
	if err = db.Close(); err != nil {
		t.Fatal(err)
	}
	db = openAttentionDB(t, path)
	defer db.Close()
	service = app.NewAttentionService(NewAttentionRepository(db), nil, nil, app.ServiceOptions{})
	after, err := service.GetTrackingSource(ctx, human, bound.Source.ID)
	if err != nil || !reflect.DeepEqual(bound, after) {
		t.Fatal("real close/reopen changed exact persisted source/metadata", err)
	}
	var sources int
	if err = db.conn.QueryRow("SELECT count(*) FROM attention_tracking_sources").Scan(&sources); err != nil || sources != 1 {
		t.Fatal("canonical binding not unique after restart", sources, err)
	}
}
