package importers

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/cookiejar"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/songconmaisaix31-design/Astrocyte/internal/apierrors"
)

func TestExplicitPublicFavoriteURLIdentity(t *testing.T) {
	for _, locator := range []string{"https://space.bilibili.com/84912/favlist?fid=1103407912&ftype=create", "https://www.bilibili.com/medialist/detail/ml1103407912"} {
		id, err := BilibiliCollectionID(locator)
		if err != nil || id != "1103407912" {
			t.Fatalf("%s: %s %v", locator, id, err)
		}
	}
	for _, locator := range []string{"https://space.bilibili.com/84912/video", "https://www.bilibili.com/watchlater", "https://evil.test/medialist/detail/ml1103407912", "https://www.bilibili.com:443/medialist/detail/ml1103407912", "https://user:pass@space.bilibili.com/84912/favlist?fid=1", "http://space.bilibili.com/84912/favlist?fid=1", "https://space.bilibili.com/84912/favlist?fid=-1", "https://space.bilibili.com/84912/favlist?fid=1/2"} {
		if _, err := BilibiliCollectionID(locator); err == nil {
			t.Fatalf("accepted unrelated or credential-bearing source %s", locator)
		}
	}
}

func TestPublicFavoriteMetadataDoesNotFetchMedia(t *testing.T) {
	raw := []byte(`{"code":0,"data":{"info":{"id":1103407912,"title":"public"},"has_more":true,"medias":[{"bvid":"BV1PReT6EEqR","type":2,"title":"original title","intro":"metadata only","upper":{"name":"creator"},"pubtime":123}]}}`)
	page, err := parseBilibiliFavorites(raw, "1103407912", 1)
	if err != nil || len(page.Items) != 1 || page.NextCursor != "2" || !page.HasMore {
		t.Fatalf("metadata page: %+v %v", page, err)
	}
	item := page.Items[0]
	if item.ExternalID != "BV1PReT6EEqR" || item.Locator != "https://www.bilibili.com/video/BV1PReT6EEqR/" || item.Description != "metadata only" || item.Author != "creator" {
		t.Fatalf("provider identity/metadata lost: %+v", item)
	}
	// Observed live upstream public example: code0, medias:null, count0.
	empty := []byte(`{"code":0,"data":{"info":{"id":1103407912,"media_count":0},"has_more":false,"medias":null}}`)
	page, err = parseBilibiliFavorites(empty, "1103407912", 1)
	if err != nil || page.Items == nil || len(page.Items) != 0 || page.HasMore || page.NextCursor != "" {
		t.Fatalf("explicit successful empty listing: %+v %v", page, err)
	}
}

func TestDiscoveryRejectsChallengeMalformedPartialAndPrivateResponses(t *testing.T) {
	for _, raw := range []string{
		`<html><script>challenge()</script></html>`, `{}`, `{"data":{"medias":[]}}`,
		`{"code":-352,"data":{"medias":[]}}`, `{"code":0,"data":null}`,
		`{"code":0,"data":{"info":{"id":1},"has_more":false,"medias":[]}}`,
		`{"code":0,"data":{"info":{"id":1103407912},"medias":[]}}`,
		`{"code":0,"data":{"info":{"id":1103407912},"has_more":true,"medias":[]}}`,
		`{"code":0,"data":{"info":{"id":1103407912},"has_more":false,"medias":null}}`,
		`{"code":0,"data":{"info":{"id":1103407912},"has_more":false,"medias":[{"type":2,"bvid":"BV1PReT6EEqR","title":"first"},{"type":2,"bvid":"BV1PReT6EEqR","title":"changed"}]}}`,
	} {
		if _, err := parseBilibiliFavorites([]byte(raw), "1103407912", 1); err == nil {
			t.Fatalf("malformed/challenged listing succeeded: %s", raw)
		}
	}
	_, err := parseBilibiliFavorites([]byte(`{"code":-403}`), "1103407912", 1)
	var service *apierrors.ServiceError
	if !errors.As(err, &service) || service.Code != apierrors.ScopeDenied {
		t.Fatalf("private listing should remain authorization denied: %v", err)
	}
}

func TestStaleCollectionEntryDoesNotBlockHumanSelectionOfValidVideos(t *testing.T) {
	raw := []byte(`{"code":0,"data":{"info":{"id":1103407912},"has_more":false,"medias":[{"id":123,"type":2,"bvid":"BV1PReT6EEqR","title":"usable"},{"id":456,"type":2,"bvid":"","title":"deleted"},{"id":789,"type":12,"title":"not video"},{"type":2,"title":"identity unavailable"}]}}`)
	page, err := parseBilibiliFavorites(raw, "1103407912", 1)
	if err != nil || len(page.Items) != 3 || len(page.Warnings) != 1 {
		t.Fatalf("one unusable entry blocked usable videos: %+v %v", page, err)
	}
	if page.Items[0].Locator == "" || page.Items[0].ExternalID != "2:123" || page.Items[0].UnavailableReason != "" || page.Items[1].Locator != "" || page.Items[1].ExternalID != "2:456" || page.Items[1].UnavailableReason == "" || page.Items[2].Locator != "" || page.Items[2].UnavailableReason == "" {
		t.Fatalf("unusable entries fabricated URLs or lost identity: %+v", page)
	}
}

func TestCollectionDiscoveryPreservesExactProviderIDs(t *testing.T) {
	collections, err := parseBilibiliCollections([]byte(`{"code":0,"data":{"list":[{"id":9007199254740993,"title":"public","media_count":2}]}}`), "84912")
	if err != nil || len(collections) != 1 || collections[0].ExternalID != "9007199254740993" || collections[0].OwnerID != "84912" {
		t.Fatalf("large provider ID precision: %+v %v", collections, err)
	}
}

type discoveryTransport func(*http.Request) (*http.Response, error)

func (f discoveryTransport) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

func TestPublicDiscoveryCredentialsRedirectsAndCancellation(t *testing.T) {
	calls := 0
	b := &PublicBilibili{Client: &http.Client{Transport: discoveryTransport(func(r *http.Request) (*http.Response, error) {
		calls++
		if r.URL.Host != "api.bilibili.com" || r.Header.Get("Cookie") != "" || r.Header.Get("Authorization") != "" || r.URL.Query().Get("media_id") != "1103407912" {
			t.Fatalf("unsafe discovery request: %v", r)
		}
		return &http.Response{StatusCode: 302, Header: http.Header{"Location": {"https://private.test/"}}, Body: io.NopCloser(strings.NewReader("redirect")), Request: r}, nil
	})}}
	if _, err := b.Page(context.Background(), "1103407912", 1); err == nil || calls != 1 {
		t.Fatalf("redirect followed or reported successful: %d %v", calls, err)
	}
	jar, _ := cookiejar.New(nil)
	b.Client.Jar = jar
	if _, err := b.Page(context.Background(), "1103407912", 1); err == nil || calls != 1 {
		t.Fatalf("credential-enabled client used: %d %v", calls, err)
	}
	b.Client.Jar = nil
	b.Client.Transport = discoveryTransport(func(r *http.Request) (*http.Response, error) { return nil, r.Context().Err() })
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := b.Page(ctx, "1103407912", 1); !errors.Is(err, context.Canceled) {
		t.Fatalf("lost cancellation: %v", err)
	}
}

func TestDouyinOfficialAccountResponseParserIsNotAuthorization(t *testing.T) {
	// Shape from official Douyin video.list docs; local contract evidence only.
	raw := []byte(`{"extra":{"error_code":0},"data":{"error_code":0,"has_more":true,"cursor":9007199254740993,"list":[{"item_id":"opaque-provider-item","video_id":"70000000001","title":"original title","create_time":1571075129,"video_status":5}]}}`)
	page, err := ParseDouyinAccountVideos(raw)
	if err != nil || len(page.Items) != 1 || page.NextCursor != "9007199254740993" || page.Items[0].ExternalID != "opaque-provider-item" || page.Items[0].Locator != "https://www.douyin.com/video/70000000001" {
		t.Fatalf("official metadata decoding: %+v %v", page, err)
	}
	for _, raw := range []string{`<html>challenge</html>`, `{"data":{"error_code":2100005}}`, `{"data":{"error_code":0}}`, `{"data":{"error_code":0,"has_more":true,"cursor":0,"list":[]}}`, `{"data":{"error_code":0,"has_more":false,"cursor":0.5,"list":[]}}`} {
		if _, err := ParseDouyinAccountVideos([]byte(raw)); err == nil {
			t.Fatalf("invalid/blocked account listing accepted: %s", raw)
		}
	}
}

// Explicit public upstream example; metadata only, never account binding,
// media/ASR/model processing or evidence of the user's own account acceptance.
func TestPublicBilibiliLiveMetadata(t *testing.T) {
	if os.Getenv("ASTROCYTE_TEST_PUBLIC_DISCOVERY") != "1" {
		t.Skip("opt-in public metadata diagnosis")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 40*time.Second)
	defer cancel()
	page, err := NewPublicBilibili().Page(ctx, "1103407912", 1)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("public example title=%q items=%d has_more=%t next_cursor=%q", page.Title, len(page.Items), page.HasMore, page.NextCursor)
	if page.Items == nil {
		t.Fatal("successful response must expose an explicit listing")
	}
}
