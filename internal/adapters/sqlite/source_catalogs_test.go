package sqlite

import (
	"context"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/songconmaisaix31-design/Astrocyte/internal/adapters/importers"
	"github.com/songconmaisaix31-design/Astrocyte/internal/apierrors"
	"github.com/songconmaisaix31-design/Astrocyte/internal/attention/app"
)

type catalogFixture struct {
	calls int
	fail  bool
}

func (*catalogFixture) ConfiguredSelectedOwner() string { return "MS4wLjAB_fixture" }
func (f *catalogFixture) ListCollections(_ context.Context, platform, owner string) ([]app.SourceCollection, error) {
	f.calls++
	if platform != "douyin" || owner != f.ConfiguredSelectedOwner() {
		panic("catalog scope expanded")
	}
	if f.fail {
		return nil, &apierrors.ServiceError{Code: apierrors.ProviderUnavailable, Message: "contract_local bridge disconnected"}
	}
	count := 3
	return []app.SourceCollection{{ExternalID: "123", OwnerID: owner, Title: "求职", Locator: importers.DouyinFolderLocator(owner, "123"), ItemCount: &count}}, nil
}

func TestSelectedCatalogExplicitDiscoveryCacheRestartAndBindingScope(t *testing.T) {
	ctx := context.Background()
	path := filepath.Join(t.TempDir(), "catalog.sqlite")
	db := openAttentionDB(t, path)
	defer func() { db.Close() }()
	f := &catalogFixture{}
	s := app.NewAttentionService(NewAttentionRepository(db), nil, nil, app.ServiceOptions{CollectionReader: f})
	p := app.Principal{ID: "catalog-human", Kind: "human"}
	if _, err := s.ListSourceCollections(ctx, p, "douyin", "self"); err == nil {
		t.Fatal("undiscovered became successful empty account")
	}
	if f.calls != 0 {
		t.Fatal("GET executed browser")
	}
	bind := app.BindTrackingSourceCommand{CommandMeta: trackingMeta("premature-bind", 1), Platform: "douyin", SourceKind: "favorites", AccessMode: "browser_selected", OwnerID: f.ConfiguredSelectedOwner(), ExternalID: "123", Locator: importers.DouyinFolderLocator(f.ConfiguredSelectedOwner(), "123")}
	if _, err := s.BindTrackingSource(ctx, p, bind); err == nil {
		t.Fatal("undiscovered browser folder bound")
	}
	c := app.DiscoverSourceCollectionsCommand{CommandMeta: trackingMeta("discover", 1), Platform: "douyin", OwnerID: "self", AccessMode: "browser_selected"}
	first, err := s.DiscoverSourceCollections(ctx, p, c)
	if err != nil || len(first.Items) != 1 || f.calls != 1 {
		t.Fatal(first, err, f.calls)
	}
	if _, err = s.DiscoverSourceCollections(ctx, p, c); err != nil || f.calls != 1 {
		t.Fatal("receipt replay executed browser", err, f.calls)
	}
	if _, err = s.ListSourceCollections(ctx, p, "douyin", "self"); err != nil || f.calls != 1 {
		t.Fatal("cache GET executed browser", err, f.calls)
	}
	bind.CommandMeta = trackingMeta("bind-selected", 1)
	bound, err := s.BindTrackingSource(ctx, p, bind)
	if err != nil || bound.Source.AccessMode != "browser_selected" || bound.Source.Title != "求职" {
		t.Fatal(bound, err)
	}
	bind.CommandMeta = trackingMeta("bind-selected-alias", 1)
	bind.Locator += "&from_tab_name=main"
	alias, err := s.BindTrackingSource(ctx, p, bind)
	if err != nil || alias.Source.ID != bound.Source.ID {
		t.Fatal("tracking URL created new binding", alias, err)
	}
	bind.CommandMeta = trackingMeta("bind-unselected", 1)
	bind.ExternalID = "456"
	bind.Locator = importers.DouyinFolderLocator(bind.OwnerID, "456")
	if _, err = s.BindTrackingSource(ctx, p, bind); err == nil {
		t.Fatal("undiscovered folder admitted")
	}
	if err = db.Close(); err != nil {
		t.Fatal(err)
	}
	db = openAttentionDB(t, path)
	s = app.NewAttentionService(NewAttentionRepository(db), nil, nil, app.ServiceOptions{CollectionReader: f})
	after, err := s.ListSourceCollections(ctx, p, "douyin", "self")
	if err != nil || !reflect.DeepEqual(first, after) || f.calls != 1 {
		t.Fatal("cold restart lost catalog or executed browser", after, err, f.calls)
	}
	f.fail = true
	c.CommandMeta = trackingMeta("discover-failed", 1)
	if _, err = s.DiscoverSourceCollections(ctx, p, c); err == nil {
		t.Fatal("disconnected bridge became success")
	}
	after, err = s.ListSourceCollections(ctx, p, "douyin", "self")
	if err != nil || !reflect.DeepEqual(first, after) || f.calls != 2 {
		t.Fatal("failed discovery destroyed cached catalog", after, err, f.calls)
	}
	if _, err = s.DiscoverSourceCollections(ctx, app.Principal{ID: "machine", Kind: "agent"}, c); err == nil || f.calls != 2 {
		t.Fatal("agent invoked browser", err, f.calls)
	}
	var sources, materials, jobs int
	db.conn.QueryRow("SELECT count(*) FROM attention_tracking_sources").Scan(&sources)
	db.conn.QueryRow("SELECT count(*) FROM attention_materials").Scan(&materials)
	db.conn.QueryRow("SELECT count(*) FROM attention_jobs").Scan(&jobs)
	if sources != 1 || materials != 0 || jobs != 0 {
		t.Fatal("catalog discovery triggered import/model work", sources, materials, jobs)
	}
}
