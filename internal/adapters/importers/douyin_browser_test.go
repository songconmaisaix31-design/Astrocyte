package importers

import (
	"context"
	"encoding/json"
	"reflect"
	"strings"
	"testing"

	"github.com/songconmaisaix31-design/Astrocyte/internal/attention/app"
)

type selectedBridgeFixture struct {
	page, catalog         string
	calls                 int
	owner, folder, cursor string
	limit                 int
}

func (f *selectedBridgeFixture) ReadSelectedFolders(_ context.Context, owner string, folders []string) ([]byte, error) {
	f.calls++
	if owner != "MS4wLjAB_fixture" || !reflect.DeepEqual(folders, []string{"123", "456"}) {
		panic("scope expanded")
	}
	return []byte(f.catalog), nil
}
func (f *selectedBridgeFixture) ReadSelectedFolderPage(_ context.Context, owner, folder, cursor string, limit int) ([]byte, error) {
	f.calls++
	f.owner, f.folder, f.cursor, f.limit = owner, folder, cursor, limit
	return []byte(f.page), nil
}

func TestSelectedDouyinScopeDeniesBeforeBrowserRead(t *testing.T) {
	fixture := &selectedBridgeFixture{}
	r := &DouyinBrowser{Bridge: fixture, OwnerID: "MS4wLjAB_fixture", FolderIDs: []string{"123", "456"}}
	for _, source := range []app.TrackingSource{
		{Platform: "douyin", SourceKind: "favorites", OwnerID: "self", ExternalID: "123"},
		{Platform: "douyin", SourceKind: "favorites", OwnerID: "MS4wLjAB_other", ExternalID: "123"},
		{Platform: "douyin", SourceKind: "favorites", OwnerID: r.OwnerID, ExternalID: "789"},
		{Platform: "douyin", SourceKind: "uploads", OwnerID: r.OwnerID, ExternalID: "123"},
	} {
		if _, err := r.ReadPage(context.Background(), source, "", 100); err == nil {
			t.Fatal("expanded scope", source)
		}
	}
	if fixture.calls != 0 {
		t.Fatal("denied scope touched browser")
	}
	if _, err := (*DouyinBrowser)(nil).ListCollections(context.Background(), "douyin", r.OwnerID); err == nil {
		t.Fatal("disconnected became empty")
	}
}

func TestSelectedDouyinCatalogReturnsOnlyVerifiedFolders(t *testing.T) {
	f := &selectedBridgeFixture{catalog: `{"owner_id":"MS4wLjAB_fixture","folders":[{"id":"123","title":"求职","count":3},{"id":"456","title":"科研","count":0}]}`}
	r := &DouyinBrowser{Bridge: f, OwnerID: "MS4wLjAB_fixture", FolderIDs: []string{"123", "456"}}
	rows, err := r.ListCollections(context.Background(), "douyin", r.OwnerID)
	if err != nil || len(rows) != 2 || *rows[0].ItemCount != 3 || *rows[1].ItemCount != 0 || !strings.Contains(rows[0].Locator, "collect_id=123") {
		t.Fatal(rows, err)
	}
	for _, raw := range []string{
		`{"owner_id":"MS4wLjAB_fixture","folders":[]}`,
		`{"owner_id":"MS4wLjAB_other","folders":[{"id":"123","title":"求职"},{"id":"456","title":"科研"}]}`,
		`{"owner_id":"MS4wLjAB_fixture","folders":[{"id":"123","title":"求职"},{"id":"789","title":"unselected"}]}`,
		`{"owner_id":"MS4wLjAB_fixture","folders":[{"id":"123","title":"求职"},{"id":"123","title":"duplicate"}]}`,
	} {
		f.catalog = raw
		if _, err = r.ListCollections(context.Background(), "douyin", r.OwnerID); err == nil {
			t.Fatal("incomplete/expanded catalog accepted", raw)
		}
	}
}

func TestSelectedDouyinMetadataPreservesExactIDsAndBoundedObservations(t *testing.T) {
	raw := `{"owner_id":"MS4wLjAB_fixture","folder_id":"123","status_code":0,"has_more":true,"cursor":"20","videos":[{"id":"7490000000000000001","title":"岗位","description":"完整简介"},{"id":"7490000000000000001","title":"岗位","description":"完整简介"},{"id":"7490000000000000002","unavailable_reason":"video removed"}]}`
	f := &selectedBridgeFixture{page: raw}
	r := &DouyinBrowser{Bridge: f, OwnerID: "MS4wLjAB_fixture", FolderIDs: []string{"123", "456"}}
	source := app.TrackingSource{Platform: "douyin", SourceKind: "favorites", OwnerID: r.OwnerID, ExternalID: "123"}
	page, err := r.ReadPage(context.Background(), source, "", 3)
	if err != nil || page.Observed != 3 || len(page.Items) != 2 || *page.NextCursor != "20" || page.Items[0].ExternalID != "7490000000000000001" || page.Items[0].Description != "完整简介" || page.Items[1].Locator != "" {
		t.Fatal(page, err)
	}
	if f.owner != r.OwnerID || f.folder != "123" || f.cursor != "" || f.limit != 3 {
		t.Fatal("request broadened", f)
	}
	var value map[string]any
	if err = json.Unmarshal([]byte(raw), &value); err != nil {
		t.Fatal(err)
	}
	for _, field := range []string{"owner_id", "folder_id", "status_code", "has_more", "videos"} {
		copy := map[string]any{}
		for k, v := range value {
			copy[k] = v
		}
		delete(copy, field)
		data, _ := json.Marshal(copy)
		if _, err = parseDouyinSelectedPage(data, r.OwnerID, "123", "", 3); err == nil {
			t.Fatal("missing required field accepted", field)
		}
	}
	for _, bad := range []string{
		strings.Replace(raw, `"folder_id":"123"`, `"folder_id":"789"`, 1),
		strings.Replace(raw, `"status_code":0`, `"status_code":8`, 1),
		strings.Replace(raw, `"cursor":"20"`, `"cursor":"0"`, 1),
		strings.Replace(raw, `"title":"岗位"`, `"title":"conflicting"`, 1),
		`<html>captcha</html>`,
	} {
		if _, err = parseDouyinSelectedPage([]byte(bad), r.OwnerID, "123", "0", 3); err == nil {
			t.Fatal("bad provider metadata accepted", bad)
		}
	}
	if _, err = parseDouyinSelectedPage([]byte(raw), r.OwnerID, "123", "", 2); err == nil {
		t.Fatal("metadata cap exceeded")
	}
	page, err = parseDouyinSelectedPage([]byte(`{"owner_id":"MS4wLjAB_fixture","folder_id":"456","status_code":0,"has_more":false,"cursor":"0","videos":[]}`), r.OwnerID, "456", "", 100)
	if err != nil || page.Observed != 0 || len(page.Items) != 0 || page.HasMore {
		t.Fatal("verified empty folder lost", page, err)
	}
}
