package importers

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"testing"

	"github.com/songconmaisaix31-design/Astrocyte/internal/attention/app"
)

// 观察包装只记录实际公开请求，响应始终来自原生产 Transport。
type actualListingTransport struct {
	inner     http.RoundTripper
	positions []int
	sizes     []int
	responses []*actualListingResponse
}

type actualListingResponse struct {
	status int
	body   bytes.Buffer
}

type actualListingBody struct {
	io.Reader
	io.Closer
}

func (r *actualListingTransport) RoundTrip(request *http.Request) (*http.Response, error) {
	observed := request.URL.Path == "/x/v3/fav/resource/list"
	if observed {
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
	response, err := r.inner.RoundTrip(request)
	if err != nil {
		return nil, err
	}
	if observed {
		original := &actualListingResponse{status: response.StatusCode}
		r.responses = append(r.responses, original)
		response.Body = actualListingBody{Reader: io.TeeReader(response.Body, &original.body), Closer: response.Body}
	}
	return response, nil
}

type actualListingEntry struct {
	Title string `json:"title"`
	Intro string `json:"intro"`
	Cover string `json:"cover"`
	Attr  *int   `json:"attr"`
	Pub   int64  `json:"pubtime"`
	Upper struct {
		Name string `json:"name"`
	} `json:"upper"`
}

func assertActualListingResponse(t *testing.T, response *actualListingResponse, folderID string, actual app.ListingPage) int {
	t.Helper()
	var original struct {
		Code *int `json:"code"`
		Data struct {
			Info struct {
				ID json.Number `json:"id"`
			} `json:"info"`
			HasMore *bool                `json:"has_more"`
			Medias  []actualListingEntry `json:"medias"`
		} `json:"data"`
	}
	if err := json.Unmarshal(response.body.Bytes(), &original); err != nil {
		t.Fatal("actual provider JSON could not be independently decoded", err)
	}
	if response.status != http.StatusOK || original.Code == nil || *original.Code != 0 || original.Data.Info.ID.String() != folderID || original.Data.HasMore == nil || actual.HasMore != *original.Data.HasMore || actual.Observed != len(original.Data.Medias) {
		t.Fatal("actual production page differs from independent provider identity, slots or continuation")
	}
	if len(original.Data.Medias) > 0 && len(actual.Items) == 0 {
		t.Fatal("actual ordinary source yielded no metadata for source-order verification")
	}
	// 仅核对同次来源的原字段及顺序，不复制生产身份转换或可用性过滤。
	position := 0
	for _, item := range actual.Items {
		matched := false
		for position < len(original.Data.Medias) {
			entry := original.Data.Medias[position]
			position++
			if item.Title == entry.Title && item.Description == entry.Intro && item.Author == entry.Upper.Name && item.Cover == entry.Cover && item.PublishedAt == entry.Pub && reflect.DeepEqual(item.ProviderStatus, entry.Attr) {
				matched = true
				break
			}
		}
		if !matched {
			t.Fatal("actual production metadata differs from the original response fields or source order")
		}
	}
	return len(original.Data.Medias)
}

func TestPublicListingActualWindowCursorWithUnavailableSlot(t *testing.T) {
	testActualListingWindows(t, true)
}

func TestPublicListingActualWindowMatchesProviderResponse(t *testing.T) {
	testActualListingWindows(t, false)
}

func testActualListingWindows(t *testing.T, requireOriginalMissingSlot bool) {
	t.Helper()
	ctx := context.Background()
	output := os.Getenv("ASTROCYTE_TEST_OUTPUT_ROOT")
	if !filepath.IsAbs(output) {
		t.Fatal("actual public listing originals require an explicit absolute owned output root")
	}
	if err := os.MkdirAll(output, 0700); err != nil {
		t.Fatal(err)
	}
	directory, err := os.MkdirTemp(output, "public-listing-actual-")
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("actual public response originals: %s", directory)
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
	cursor, observed, originalSlots := "", 0, 0
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
		if len(transport.responses) != page {
			t.Fatal("one actual resource request must produce one original response")
		}
		response := transport.responses[page-1]
		if err := os.WriteFile(filepath.Join(directory, "page-"+strconv.Itoa(page)+".json"), response.body.Bytes(), 0600); err != nil {
			t.Fatal(err)
		}
		slots := assertActualListingResponse(t, response, selected.ExternalID, actual)
		originalSlots += slots
		observed += actual.Observed
		cursor = *actual.NextCursor
	}
	if observed != originalSlots || len(transport.positions) != 3 {
		t.Fatalf("actual three-page count differs from original responses: observed=%d original_slots=%d requests=%d", observed, originalSlots, len(transport.positions))
	}
	if requireOriginalMissingSlot && observed != 59 {
		t.Fatalf("original actual omitted-slot cohort unavailable: observed=%d original_slots=%d requests=%d; one real omitted slot remains required", observed, originalSlots, len(transport.positions))
	}
	for i := range transport.positions {
		if transport.positions[i] != i*20 || transport.sizes[i] != 20 {
			t.Fatal("actual request windows overlapped or fetched outside their sizes")
		}
	}
	t.Logf("actual_folder=%s provider_slots_requested=60 independently_observed_slots=%d metadata_observed=%d next_cursor=%s", selected.ExternalID, originalSlots, observed, cursor)
	if originalSlots == 60 {
		t.Log("NOT_RUN: these actual responses contain no omitted provider slot; omitted-slot acceptance remains unverified")
	} else {
		t.Logf("actual omitted provider slots=%d; continuation retained exact requested windows", 60-originalSlots)
	}
}
