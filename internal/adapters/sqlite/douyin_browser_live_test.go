package sqlite

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"testing"
	"time"

	"github.com/songconmaisaix31-design/Astrocyte/internal/adapters/importers"
	"github.com/songconmaisaix31-design/Astrocyte/internal/attention/app"
)

// Opt-in scoped real browser -> service -> SQLite. Preserve actual metadata only,
// a dedicated DB and the first failure; never request media/model processing.
func TestSelectedDouyinBrowserLiveSyncRepeatColdRestart(t *testing.T) {
	root := os.Getenv("ASTROCYTE_TEST_SELECTED_DOUYIN_DIR")
	if root == "" {
		t.Skip("explicit selected-browser metadata acceptance directory required")
	}
	if !filepath.IsAbs(root) {
		t.Fatal("absolute independent acceptance directory required")
	}
	if err := os.MkdirAll(root, 0700); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(root, "tracking.sqlite")
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Fatal("fresh directory required; preserve first observations")
	}
	var folders []string
	if err := json.Unmarshal([]byte(os.Getenv("ASTROCYTE_DOUYIN_FOLDER_IDS")), &folders); err != nil || len(folders) != 1 {
		t.Fatal("one explicitly authorized selected folder required", err)
	}
	owner := os.Getenv("ASTROCYTE_DOUYIN_OWNER_ID")
	bridge := importers.NewOpenCLISelectedFolderBridge(os.Getenv("ASTROCYTE_DOUYIN_OPENCLI_NODE"), os.Getenv("ASTROCYTE_DOUYIN_OPENCLI_MAIN"), os.Getenv("ASTROCYTE_DOUYIN_OPENCLI_PROFILE"), os.Getenv("ASTROCYTE_DOUYIN_OPENCLI_SESSION"))
	reader := importers.NewPublicListingReader("")
	reader.Douyin = &importers.DouyinBrowser{Bridge: bridge, OwnerID: owner, FolderIDs: folders}
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()
	db := openAttentionDB(t, path)
	defer func() { db.Close() }()
	s := app.NewAttentionService(NewAttentionRepository(db), nil, nil, app.ServiceOptions{ListingReader: reader, CollectionReader: reader, JobTimeout: time.Minute, MaxAttempts: 1})
	p := app.Principal{ID: "selected-douyin-acceptance", Kind: "human"}
	catalog, err := s.DiscoverSourceCollections(ctx, p, app.DiscoverSourceCollectionsCommand{CommandMeta: trackingMeta("discover-selected", 1), Platform: "douyin", AccessMode: "browser_selected", OwnerID: "self"})
	if err != nil || len(catalog.Items) != 1 {
		t.Fatal("actual selected catalog", catalog, err)
	}
	row, ok := catalog.Items[0].(app.SourceCollection)
	if !ok || row.OwnerID != owner || row.ExternalID != folders[0] || row.ItemCount == nil || *row.ItemCount < 1 {
		t.Fatal("real nonempty selected-folder identity required", row)
	}
	data, _ := json.MarshalIndent(catalog, "", "  ")
	if err = os.WriteFile(filepath.Join(root, "actual-selected-catalog.json"), data, 0600); err != nil {
		t.Fatal(err)
	}
	bound, err := s.BindTrackingSource(ctx, p, app.BindTrackingSourceCommand{CommandMeta: trackingMeta("bind-selected", 1), Platform: "douyin", SourceKind: "favorites", AccessMode: "browser_selected", OwnerID: row.OwnerID, ExternalID: row.ExternalID, Title: row.Title, Locator: row.Locator})
	if err != nil {
		t.Fatal(err)
	}
	var before app.TrackingSourceResult
	for i := 0; i < 2; i++ {
		queued, err := s.SyncTrackingSource(ctx, p, bound.Source.ID, app.SyncTrackingSourceCommand{CommandMeta: trackingMeta("sync-selected-"+string(rune('0'+i)), bound.Source.Version)})
		if err != nil {
			t.Fatal(err)
		}
		if _, err = s.ProcessNextJob(ctx); err != nil {
			t.Fatal(err)
		}
		job, err := s.GetJob(ctx, p, queued.Jobs[0].JobID)
		if err != nil || job.Status != "succeeded" || job.Attempts != 1 {
			t.Fatal("actual metadata job", job, err)
		}
		current, err := s.GetTrackingSource(ctx, p, bound.Source.ID)
		if err != nil || current.Source.Status != "succeeded" || current.Source.HasMore || len(current.Items) != *row.ItemCount {
			t.Fatal("actual selected metadata", current, err)
		}
		for _, item := range current.Items {
			if item.Revision != 1 || item.Metadata.Description == "" || item.SourceID != bound.Source.ID || item.Selected || item.ImportJobID != nil || item.MaterialID != nil || item.Recommendation != nil || item.Stale {
				t.Fatal("metadata expanded into body/model work", item)
			}
		}
		if i == 1 && !reflect.DeepEqual(before.Items, current.Items) {
			t.Fatal("repeat changed revisions or duplicated rows")
		}
		before, bound = current, current
	}
	if err = db.Close(); err != nil {
		t.Fatal(err)
	}
	db = openAttentionDB(t, path)
	s = app.NewAttentionService(NewAttentionRepository(db), nil, nil, app.ServiceOptions{ListingReader: reader, CollectionReader: reader, JobTimeout: time.Minute, MaxAttempts: 1})
	after, err := s.GetTrackingSource(ctx, p, before.Source.ID)
	if err != nil || !reflect.DeepEqual(before, after) {
		t.Fatal("real SQLite cold restart changed metadata", after, err)
	}
	cached, err := s.ListSourceCollections(ctx, p, "douyin", "self")
	if err != nil || !reflect.DeepEqual(catalog, cached) {
		t.Fatal("catalog cold restart changed", cached, err)
	}
	for _, table := range []string{"attention_materials", "attention_distillations", "attention_opportunities"} {
		var count int
		if err = db.conn.QueryRow("SELECT count(*) FROM " + table).Scan(&count); err != nil || count != 0 {
			t.Fatal("metadata created unselected artifacts", table, count, err)
		}
	}
	var jobs int
	if err = db.conn.QueryRow("SELECT count(*) FROM attention_jobs WHERE json_extract(data,'$.kind') != 'source_sync'").Scan(&jobs); err != nil || jobs != 0 {
		t.Fatal("metadata invoked media/model", jobs, err)
	}
	data, _ = json.MarshalIndent(after, "", "  ")
	if err = os.WriteFile(filepath.Join(root, "actual-selected-after-repeat-restart.json"), data, 0600); err != nil {
		t.Fatal(err)
	}
	t.Logf("real selected browser folder %s title=%s items=%d, repeat revision1 and SQLite cold restart preserved; no model/media", row.ExternalID, row.Title, len(after.Items))
}
