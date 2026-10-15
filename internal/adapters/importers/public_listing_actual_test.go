package importers

import (
	"context"
	"net/http"
	"strconv"
	"testing"

	"github.com/songconmaisaix31-design/Astrocyte/internal/attention/app"
)

// 观察包装只记录实际公开请求，响应始终来自原生产 Transport。
type actualListingTransport struct {
	inner     http.RoundTripper
	positions []int
	sizes     []int
}

func (r *actualListingTransport) RoundTrip(request *http.Request) (*http.Response, error) {
	if request.URL.Path == "/x/v3/fav/resource/list" {
		page, err := strconv.Atoi(request.URL.Query().Get("pn"))
		if err != nil {
			return nil, err
		}
		size, err := strconv.Atoi(request.URL.Query().Get("ps"))
		if err != nil {
			return nil, err
		}
		r.positions = append(r.positions, (page-1)*size)
		r.sizes = append(r.sizes, size)
	}
	return r.inner.RoundTrip(request)
}

func TestPublicListingActualWindowCursorWithUnavailableSlot(t *testing.T) {
	ctx := context.Background()
	reader := NewPublicListingReader("")
	folders, err := reader.ListCollections(ctx, "bilibili", "3494358764489275")
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
		t.Fatal("approved actual folder with enough metadata unavailable")
	}
	transport := &actualListingTransport{inner: http.DefaultTransport}
	reader.Bilibili.Client.Transport = transport
	source := app.TrackingSource{Platform: "bilibili", SourceKind: "favorites", ExternalID: selected.ExternalID, OwnerID: selected.OwnerID}
	cursor, observed := "", 0
	for page := 1; page <= 3; page++ {
		actual, err := reader.ReadPage(ctx, source, cursor, 100-observed)
		if err != nil {
			t.Fatal(err)
		}
		if actual.NextCursor == nil || *actual.NextCursor != strconv.Itoa(page+1) || !actual.HasMore {
			t.Fatal("actual provider window continuation did not advance by its request size")
		}
		if actual.Observed < len(actual.Items) || actual.Observed > 20 {
			t.Fatal("actual observation count changed")
		}
		observed += actual.Observed
		cursor = *actual.NextCursor
	}
	if observed != 59 || len(transport.positions) != 3 {
		t.Fatalf("actual three-page cohort changed: observed=%d requests=%d", observed, len(transport.positions))
	}
	for i := range transport.positions {
		if transport.positions[i] != i*20 || transport.sizes[i] != 20 {
			t.Fatal("actual request windows overlapped or fetched outside their sizes")
		}
	}
	t.Logf("actual_folder=%s provider_slots_requested=60 metadata_observed=%d next_cursor=%s", selected.ExternalID, observed, cursor)
}
