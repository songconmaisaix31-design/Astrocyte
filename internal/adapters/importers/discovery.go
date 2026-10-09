package importers

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/songconmaisaix31-design/Astrocyte/internal/apierrors"
	"github.com/songconmaisaix31-design/Astrocyte/internal/attention/domain"
)

// Discovery metadata is not imported material or a model recommendation. These
// adapter-local values deliberately do not choose the pending binding/API policy.
type DiscoveredVideo = domain.ListingItem

type DiscoveredCollection struct {
	ExternalID, OwnerID, Title, Locator string
	ItemCount                           int
}

type DiscoveryPage struct {
	Title      string
	Items      []DiscoveredVideo
	HasMore    bool
	NextCursor string
	Warnings   []string
}

type PublicBilibili struct{ Client *http.Client }

func NewPublicBilibili() *PublicBilibili {
	return &PublicBilibili{Client: &http.Client{Timeout: 30 * time.Second, CheckRedirect: func(_ *http.Request, _ []*http.Request) error {
		return fmt.Errorf("public discovery does not follow redirects")
	}}}
}

var decimalID = regexp.MustCompile(`^[1-9][0-9]{0,19}$`)
var biliVideoID = regexp.MustCompile(`^BV[0-9A-Za-z]{10}$`)

// BilibiliCollectionID accepts explicit public favorite-folder URLs only; account
// uploads and private watch-later are distinct sources, never implicit aliases.
func BilibiliCollectionID(locator string) (string, error) {
	u, err := url.Parse(strings.TrimSpace(locator))
	if err != nil || u.Scheme != "https" || u.User != nil || u.Port() != "" {
		return "", invalid("an official HTTPS Bilibili favorite-folder URL is required")
	}
	host := strings.ToLower(u.Hostname())
	id := ""
	if host == "space.bilibili.com" {
		parts := strings.Split(strings.Trim(u.Path, "/"), "/")
		if len(parts) == 2 && decimalID.MatchString(parts[0]) && parts[1] == "favlist" {
			id = u.Query().Get("fid")
		}
	} else if host == "www.bilibili.com" || host == "bilibili.com" {
		id = strings.TrimPrefix(u.Path, "/medialist/detail/ml")
	}
	if !decimalID.MatchString(id) {
		return "", invalid("an explicit Bilibili favorite-folder ID is required")
	}
	return id, nil
}

// Collections lists the specified account's publicly visible folders, without a
// cookie jar, browser, credentials, private-profile discovery or media fetching.
func (b *PublicBilibili) Collections(ctx context.Context, ownerID string) ([]DiscoveredCollection, error) {
	if !decimalID.MatchString(ownerID) {
		return nil, invalid("a positive Bilibili public UID is required")
	}
	raw, err := b.get(ctx, "https://api.bilibili.com/x/v3/fav/folder/created/list-all?up_mid="+ownerID)
	if err != nil {
		return nil, err
	}
	return parseBilibiliCollections(raw, ownerID)
}

// Page fetches metadata only. Pagination is explicit, so a caller cannot silently
// claim a complete sync from one page or trigger full-collection media extraction.
func (b *PublicBilibili) Page(ctx context.Context, folderID string, page int) (DiscoveryPage, error) {
	if !decimalID.MatchString(folderID) || page < 1 || page > 100000 {
		return DiscoveryPage{}, invalid("a favorite-folder ID and positive bounded page are required")
	}
	raw, err := b.get(ctx, fmt.Sprintf("https://api.bilibili.com/x/v3/fav/resource/list?media_id=%s&pn=%d&ps=20", folderID, page))
	if err != nil {
		return DiscoveryPage{}, err
	}
	return parseBilibiliFavorites(raw, folderID, page)
}

func (b *PublicBilibili) get(ctx context.Context, address string) ([]byte, error) {
	if b.Client == nil || b.Client.Jar != nil {
		return nil, discoveryUnavailable("public discovery requires a credential-free HTTP client", "configure_public_discovery")
	}
	// Copy the trusted transport for testability; never inherit its jar or redirect
	// policy. Only fixed HTTPS API paths are constructed by the two methods above.
	client := *b.Client
	client.Jar = nil
	client.CheckRedirect = func(_ *http.Request, _ []*http.Request) error { return http.ErrUseLastResponse }
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, address, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Referer", "https://www.bilibili.com/")
	resp, err := client.Do(req)
	if err != nil {
		if ctx.Err() != nil {
			return nil, ctx.Err()
		}
		return nil, discoveryUnavailable("Bilibili public metadata request failed", "check_public_source_access")
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, discoveryUnavailable(fmt.Sprintf("Bilibili public metadata HTTP %d", resp.StatusCode), "check_public_source_access")
	}
	raw, err := io.ReadAll(io.LimitReader(resp.Body, (4<<20)+1))
	if err != nil {
		return nil, err
	}
	if len(raw) > 4<<20 {
		return nil, discoveryUnavailable("public metadata exceeds size limit", "inspect_source_response")
	}
	return raw, nil
}

type biliEnvelope struct {
	Code *int            `json:"code"`
	Data json.RawMessage `json:"data"`
}

func biliData(raw []byte) (json.RawMessage, error) {
	var envelope biliEnvelope
	if json.Unmarshal(raw, &envelope) != nil || envelope.Code == nil {
		return nil, discoveryUnavailable("Bilibili metadata response is malformed", "inspect_source_response")
	}
	if *envelope.Code == -403 || *envelope.Code == -101 {
		return nil, &apierrors.ServiceError{Code: apierrors.ScopeDenied, Message: "Bilibili source requires authorization; public discovery cannot read it", RequiredAction: "choose_authorized_binding"}
	}
	if *envelope.Code != 0 {
		return nil, discoveryUnavailable(fmt.Sprintf("Bilibili metadata code %d", *envelope.Code), "check_public_source_access")
	}
	if len(envelope.Data) == 0 || string(envelope.Data) == "null" {
		return nil, discoveryUnavailable("Bilibili metadata is missing", "inspect_source_response")
	}
	return envelope.Data, nil
}

func parseBilibiliCollections(raw []byte, ownerID string) ([]DiscoveredCollection, error) {
	data, err := biliData(raw)
	if err != nil {
		return nil, err
	}
	var value struct {
		List json.RawMessage `json:"list"`
	}
	if json.Unmarshal(data, &value) != nil || len(value.List) == 0 {
		return nil, discoveryUnavailable("Bilibili folder listing is missing", "inspect_source_response")
	}
	var entries []struct {
		ID         json.Number `json:"id"`
		Title      string      `json:"title"`
		MediaCount int         `json:"media_count"`
	}
	if json.Unmarshal(value.List, &entries) != nil {
		return nil, discoveryUnavailable("Bilibili folder listing is malformed", "inspect_source_response")
	}
	result := []DiscoveredCollection{}
	for _, entry := range entries {
		if !decimalID.MatchString(entry.ID.String()) || entry.MediaCount < 0 {
			return nil, discoveryUnavailable("Bilibili folder identity is malformed", "inspect_source_response")
		}
		result = append(result, DiscoveredCollection{ExternalID: entry.ID.String(), OwnerID: ownerID, Title: entry.Title, ItemCount: entry.MediaCount, Locator: "https://space.bilibili.com/" + ownerID + "/favlist?fid=" + entry.ID.String()})
	}
	return result, nil
}

func parseBilibiliFavorites(raw []byte, folderID string, page int) (DiscoveryPage, error) {
	data, err := biliData(raw)
	if err != nil {
		return DiscoveryPage{}, err
	}
	var value struct {
		Info *struct {
			ID         json.Number `json:"id"`
			Title      string      `json:"title"`
			MediaCount *int        `json:"media_count"`
		} `json:"info"`
		HasMore *bool           `json:"has_more"`
		Medias  json.RawMessage `json:"medias"`
	}
	if json.Unmarshal(data, &value) != nil || value.Info == nil || value.Info.ID.String() != folderID || value.HasMore == nil || len(value.Medias) == 0 {
		return DiscoveryPage{}, discoveryUnavailable("Bilibili favorite listing is malformed or mismatched", "inspect_source_response")
	}
	// The live public empty-folder response uses medias:null. Accept that only
	// when the provider explicitly reports no content and no continuation.
	if string(value.Medias) == "null" && (value.Info.MediaCount == nil || *value.Info.MediaCount != 0 || *value.HasMore) {
		return DiscoveryPage{}, discoveryUnavailable("Bilibili favorite entries are missing", "inspect_source_response")
	}
	var medias []struct {
		ID    json.Number `json:"id"`
		BVID  string      `json:"bvid"`
		Title string      `json:"title"`
		Intro string      `json:"intro"`
		Cover string      `json:"cover"`
		Type  int         `json:"type"`
		Pub   int64       `json:"pubtime"`
		Upper struct {
			Name string `json:"name"`
		} `json:"upper"`
	}
	if json.Unmarshal(value.Medias, &medias) != nil || (*value.HasMore && len(medias) == 0) {
		return DiscoveryPage{}, discoveryUnavailable("Bilibili favorite entries or pagination are malformed", "inspect_source_response")
	}
	result := DiscoveryPage{Title: value.Info.Title, Items: []DiscoveredVideo{}, HasMore: *value.HasMore, Warnings: []string{}}
	for _, entry := range medias {
		identity := entry.BVID
		if decimalID.MatchString(entry.ID.String()) {
			identity = fmt.Sprintf("%d:%s", entry.Type, entry.ID.String())
		}
		item := DiscoveredVideo{ExternalID: identity, Title: entry.Title, Description: entry.Intro, Author: entry.Upper.Name, Cover: entry.Cover, PublishedAt: entry.Pub}
		if entry.Type != 2 {
			item.UnavailableReason = "unsupported Bilibili collection entry type"
		} else if !biliVideoID.MatchString(entry.BVID) {
			item.UnavailableReason = "Bilibili collection entry has no readable video URL"
		} else {
			item.Locator = "https://www.bilibili.com/video/" + entry.BVID + "/"
		}
		if !decimalID.MatchString(entry.ID.String()) && !biliVideoID.MatchString(entry.BVID) {
			result.Warnings = append(result.Warnings, "collection entry skipped: provider identity is missing")
			continue
		}
		result.Items = append(result.Items, item)
	}
	result.Items, err = domain.UniqueListingItems(result.Items)
	if err != nil {
		return DiscoveryPage{}, discoveryUnavailable("favorite folder contains conflicting provider identities", "inspect_source_response")
	}
	if result.HasMore {
		result.NextCursor = strconv.Itoa(page + 1)
	}
	return result, nil
}

func discoveryUnavailable(message, action string) *apierrors.ServiceError {
	return &apierrors.ServiceError{Code: apierrors.ProviderUnavailable, Message: message, RequiredAction: action}
}

// ParseDouyinAccountVideos decodes the official authorized-account video-list
// response. Transport/authorization is deliberately absent until the user chooses
// binding; accepting a JSON response is not evidence of public account access.
func ParseDouyinAccountVideos(raw []byte) (DiscoveryPage, error) {
	var value struct {
		Data *struct {
			ErrorCode *int            `json:"error_code"`
			HasMore   *bool           `json:"has_more"`
			Cursor    *json.Number    `json:"cursor"`
			List      json.RawMessage `json:"list"`
		} `json:"data"`
	}
	if json.Unmarshal(raw, &value) != nil || value.Data == nil || value.Data.ErrorCode == nil {
		return DiscoveryPage{}, discoveryUnavailable("Douyin account listing is malformed or challenged", "choose_authorized_binding")
	}
	data := value.Data
	if *data.ErrorCode != 0 {
		return DiscoveryPage{}, discoveryUnavailable(fmt.Sprintf("Douyin account listing code %d", *data.ErrorCode), "check_authorized_source_access")
	}
	if data.HasMore == nil || data.Cursor == nil || len(data.List) == 0 || string(data.List) == "null" {
		return DiscoveryPage{}, discoveryUnavailable("Douyin account pagination or listing is missing", "inspect_source_response")
	}
	if _, err := strconv.ParseUint(data.Cursor.String(), 10, 64); err != nil {
		return DiscoveryPage{}, discoveryUnavailable("Douyin account cursor is malformed", "inspect_source_response")
	}
	var entries []struct {
		ID          string `json:"item_id"`
		VideoID     string `json:"video_id"`
		Title       string `json:"title"`
		Cover       string `json:"cover"`
		Created     int64  `json:"create_time"`
		VideoStatus *int   `json:"video_status"`
	}
	if json.Unmarshal(data.List, &entries) != nil || (*data.HasMore && len(entries) == 0) {
		return DiscoveryPage{}, discoveryUnavailable("Douyin account entries are malformed", "inspect_source_response")
	}
	result := DiscoveryPage{Items: []DiscoveredVideo{}, HasMore: *data.HasMore}
	for _, entry := range entries {
		if entry.ID == "" || !decimalID.MatchString(entry.VideoID) {
			return DiscoveryPage{}, discoveryUnavailable("Douyin entry lacks a unique public video identity", "inspect_collection_entries")
		}
		result.Items = append(result.Items, DiscoveredVideo{ExternalID: entry.ID, Locator: "https://www.douyin.com/video/" + entry.VideoID, Title: entry.Title, Cover: entry.Cover, PublishedAt: entry.Created, ProviderStatus: entry.VideoStatus})
	}
	var err error
	result.Items, err = domain.UniqueListingItems(result.Items)
	if err != nil {
		return DiscoveryPage{}, discoveryUnavailable("Douyin account contains conflicting provider identities", "inspect_source_response")
	}
	if result.HasMore {
		result.NextCursor = data.Cursor.String()
	}
	return result, nil
}
