package importers

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"testing"

	"github.com/songconmaisaix31-design/Astrocyte/internal/attention/app"
)

func TestPublicListingBoundsActualFavoriteRequestsIncludingSkippedRows(t *testing.T) {
	r := NewPublicListingReader("")
	folders, err := r.ListCollections(context.Background(), "bilibili", "3494358764489275")
	if err != nil {
		t.Fatal(err)
	}
	var selected *app.SourceCollection
	for i := range folders {
		if folders[i].ItemCount != nil && *folders[i].ItemCount > 100 {
			selected = &folders[i]
			break
		}
	}
	if selected == nil {
		t.Fatal("approved actual favorite folder capacity unavailable")
	}
	transport := &actualListingTransport{inner: http.DefaultTransport}
	r.Bilibili.Client.Transport = transport
	source := app.TrackingSource{Platform: "bilibili", SourceKind: "favorites", ExternalID: selected.ExternalID, OwnerID: selected.OwnerID}
	cursor := ""
	total, count := 0, 0
	for total < 100 {
		page, err := r.ReadPage(context.Background(), source, cursor, 100-total)
		if err != nil {
			t.Fatal(err)
		}
		if page.Observed <= 0 || page.Observed > 100-total || len(page.Items) > page.Observed {
			t.Fatal("actual page did not advance within the remaining metadata budget")
		}
		total += page.Observed
		count += len(page.Items)
		if page.NextCursor == nil {
			t.Fatal("lost continuation")
		}
		cursor = *page.NextCursor
	}
	if total != 100 || count != 100 || len(transport.positions) != 6 || cursor != "6:1" {
		t.Fatalf("actual observations%d metadata%d requests%d cursor%s", total, count, len(transport.positions), cursor)
	}
	for i := range transport.positions {
		expectedSize := 20
		if i == 5 {
			expectedSize = 1
		}
		if transport.positions[i] != i*20 || transport.sizes[i] != expectedSize {
			t.Fatal("actual request window changed or fetched beyond its remaining metadata limit")
		}
	}
	t.Logf("actual_folder=%s metadata_observed=100 provider_slots_requested=101 requests=6 next_cursor=%s", selected.ExternalID, cursor)
}

func TestNativeUploadLastPageUsesRemainingTenRows(t *testing.T) {
	// 下列数据只作为 JSON 解析参数与期望值，不声明为客户端或公开来源观测。
	entries := []string{}
	for i := 0; i < 10; i++ {
		entries = append(entries, fmt.Sprintf(`{"aid":%d,"mid":3494358764489275,"bvid":"BV1PReT6EEqR"}`, i+91))
	}
	raw := []byte(fmt.Sprintf(`{"code":0,"data":{"page":{"pn":10,"ps":10,"count":150},"list":{"vlist":[%s]}}}`, strings.Join(entries, ",")))
	page, err := parseBilibiliUploadsSize(raw, "3494358764489275", 10, 10)
	if err != nil || page.Observed != 10 || len(page.Items) != 10 || !page.HasMore {
		t.Fatal("bounded final native page malformed", err, page)
	}
}
