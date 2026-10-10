package importers

import "testing"

func TestBilibiliUploadsMetadataIdentityPaginationAndUnavailable(t *testing.T) {
	raw := []byte(`{"code":0,"data":{"page":{"pn":1,"ps":30,"count":31},"list":{"vlist":[{"aid":9007199254740993,"mid":3494358764489275,"bvid":"BV1PReT6EEqR","title":"Real provider title","description":"Provider description","created":123},{"aid":123,"mid":3494358764489275,"title":"Unavailable"}]}}}`)
	page, err := parseBilibiliUploads(raw, "3494358764489275", 1)
	if err != nil || len(page.Items) != 2 || !page.HasMore || page.NextCursor != "2" {
		t.Fatalf("page %+v %v", page, err)
	}
	if page.Items[0].ExternalID != "2:9007199254740993" || page.Items[0].Description != "Provider description" || page.Items[1].UnavailableReason == "" || page.Items[1].Locator != "" {
		t.Fatalf("lost identity/metadata/unavailable: %+v", page.Items)
	}
	for _, bad := range []string{`{"code":-352,"data":{}}`, `{"code":0,"data":{"page":{"pn":1,"ps":30,"count":3},"list":{"vlist":[]}}}`, `{"code":0,"data":{"page":{"pn":1,"ps":30,"count":1},"list":{"vlist":[{"aid":123,"mid":1,"bvid":"BV1PReT6EEqR"}]}}}`} {
		if _, err := parseBilibiliUploads([]byte(bad), "3494358764489275", 1); err == nil {
			t.Fatal("challenge/missing/owner mismatch accepted", bad)
		}
	}
}
