package importers

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"testing"

	"github.com/songconmaisaix31-design/Astrocyte/internal/attention/app"
)

func TestPublicListingBoundsActualFavoriteRequestsIncludingSkippedRows(t *testing.T) {
	observed, requests := 0, 0
	r := NewPublicListingReader("")
	r.Bilibili.Client = &http.Client{Transport: roundTripFunc(func(request *http.Request) (*http.Response, error) {
		q := request.URL.Query()
		page, _ := strconv.Atoi(q.Get("pn"))
		size, _ := strconv.Atoi(q.Get("ps"))
		start := (page - 1) * size
		if q.Get("media_id") != "456" || size < 1 || size > 20 || observed+size > 100 {
			t.Fatal("provider request exceeded metadata cap", request.URL)
		}
		requests++
		observed += size
		entries := []string{}
		for i := 0; i < size; i++ {
			if start+i == 5 {
				entries = append(entries, `{"type":2,"title":"missing identity"}`)
			} else {
				entries = append(entries, fmt.Sprintf(`{"id":%d,"type":2,"bvid":"BV1PReT6EEqR","title":"contract-local row"}`, start+i+1))
			}
		}
		raw := fmt.Sprintf(`{"code":0,"data":{"info":{"id":456,"media_count":150},"has_more":true,"medias":[%s]}}`, strings.Join(entries, ","))
		return &http.Response{StatusCode: 200, Body: io.NopCloser(strings.NewReader(raw)), Header: http.Header{}}, nil
	})}
	source := app.TrackingSource{Platform: "bilibili", SourceKind: "favorites", ExternalID: "456"}
	cursor := ""
	total, count := 0, 0
	for total < 100 {
		page, err := r.ReadPage(context.Background(), source, cursor, 100-total)
		if err != nil {
			t.Fatal(err)
		}
		total += page.Observed
		count += len(page.Items)
		if page.NextCursor == nil {
			t.Fatal("lost continuation")
		}
		cursor = *page.NextCursor
	}
	if observed != 100 || total != 100 || count != 99 || requests != 5 || cursor != "6" {
		t.Fatalf("real request rows%d observations%d metadata%d requests%d cursor%s", observed, total, count, requests, cursor)
	}
}

func TestNativeUploadLastPageUsesRemainingTenRows(t *testing.T) {
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
