package importers

import (
	"context"
	"encoding/json"
	"net/url"
	"regexp"
	"slices"
	"strings"

	"github.com/songconmaisaix31-design/Astrocyte/internal/apierrors"
	"github.com/songconmaisaix31-design/Astrocyte/internal/attention/app"
	"github.com/songconmaisaix31-design/Astrocyte/internal/attention/domain"
)

// SelectedFolderBridge reads the explicitly handed-off browser session. It must
// return only the requested folder metadata, never cookies, downloads, comments,
// account-wide favorites, or an inferred identity for /user/self.
type SelectedFolderBridge interface {
	ReadSelectedFolders(context.Context, string, []string) ([]byte, error)
	ReadSelectedFolderPage(context.Context, string, string, string, int) ([]byte, error)
}

// DouyinBrowser is configured by the host only after a browser ownership handoff.
// An HTTP caller cannot add another account/folder to this host allowlist.
type DouyinBrowser struct {
	Bridge    SelectedFolderBridge
	OwnerID   string
	FolderIDs []string
}

var browserDouyinOwner = regexp.MustCompile(`^[A-Za-z0-9_-]{8,200}$`)
var browserDouyinCursor = regexp.MustCompile(`^[0-9]{1,20}$`)

func browserScopeDenied(message string) error {
	return &apierrors.ServiceError{Code: apierrors.ScopeDenied, Message: message, RequiredAction: "choose_authorized_browser_folder"}
}

func (r *DouyinBrowser) validateOwner(owner string) error {
	if r == nil || r.Bridge == nil {
		return &apierrors.ServiceError{Code: apierrors.ProviderUnavailable, Message: "The selected Douyin browser bridge is not connected", RequiredAction: "connect_selected_douyin_browser"}
	}
	if !browserDouyinOwner.MatchString(owner) || owner != r.OwnerID || len(r.FolderIDs) == 0 || len(r.FolderIDs) > 100 {
		return browserScopeDenied("This browser account has not been handed off for selected-folder reads")
	}
	for _, id := range r.FolderIDs {
		if !decimalID.MatchString(id) {
			return browserScopeDenied("A stable selected favorite-folder ID is required")
		}
	}
	return nil
}

func DouyinFolderLocator(owner, folder string) string {
	q := url.Values{"showTab": {"favorite_collection"}, "showSubTab": {"favorite_folder"}, "collect_id": {folder}}
	return "https://www.douyin.com/user/" + owner + "?" + q.Encode()
}

func (r *DouyinBrowser) ListCollections(ctx context.Context, platform, owner string) ([]app.SourceCollection, error) {
	if platform != "douyin" {
		return nil, browserScopeDenied("The browser bridge only reads selected Douyin favorite folders")
	}
	if err := r.validateOwner(owner); err != nil {
		return nil, err
	}
	selected := slices.Clone(r.FolderIDs)
	raw, err := r.Bridge.ReadSelectedFolders(ctx, owner, selected)
	if err != nil {
		return nil, err
	}
	var value struct {
		OwnerID string `json:"owner_id"`
		Folders *[]struct {
			ID    string `json:"id"`
			Title string `json:"title"`
			Count *int   `json:"count"`
		} `json:"folders"`
	}
	if len(raw) > 4<<20 || json.Unmarshal(raw, &value) != nil || value.Folders == nil || value.OwnerID != owner {
		return nil, browserScopeDenied("The browser did not verify the selected account and folder metadata")
	}
	seen := map[string]bool{}
	result := []app.SourceCollection{}
	for _, folder := range *value.Folders {
		if !slices.Contains(selected, folder.ID) || seen[folder.ID] || strings.TrimSpace(folder.Title) == "" || (folder.Count != nil && *folder.Count < 0) {
			return nil, browserScopeDenied("The browser returned an unselected, duplicated, or unverifiable favorite folder")
		}
		seen[folder.ID] = true
		result = append(result, app.SourceCollection{OwnerID: owner, ExternalID: folder.ID, Title: folder.Title, Locator: DouyinFolderLocator(owner, folder.ID), ItemCount: folder.Count})
	}
	if len(result) != len(selected) {
		return nil, discoveryUnavailable("A selected favorite folder is missing; the browser result is not an empty account", "open_selected_douyin_folders")
	}
	return result, nil
}

func (r *DouyinBrowser) ReadPage(ctx context.Context, source app.TrackingSource, cursor string, limit int) (app.ListingPage, error) {
	if err := r.validateOwner(source.OwnerID); err != nil {
		return app.ListingPage{}, err
	}
	if source.Platform != "douyin" || source.AccessMode != "browser_selected" || source.SourceKind != "favorites" || !slices.Contains(r.FolderIDs, source.ExternalID) {
		return app.ListingPage{}, browserScopeDenied("Only explicitly selected Douyin favorite folders may be synchronized")
	}
	if limit < 1 || limit > 100 || (cursor != "" && !browserDouyinCursor.MatchString(cursor)) {
		return app.ListingPage{}, invalid("a bounded selected-folder page and provider cursor are required")
	}
	raw, err := r.Bridge.ReadSelectedFolderPage(ctx, source.OwnerID, source.ExternalID, cursor, limit)
	if err != nil {
		return app.ListingPage{}, err
	}
	return parseDouyinSelectedPage(raw, source.OwnerID, source.ExternalID, cursor, limit)
}

func parseDouyinSelectedPage(raw []byte, owner, folder, cursor string, limit int) (app.ListingPage, error) {
	var value struct {
		OwnerID  string `json:"owner_id"`
		FolderID string `json:"folder_id"`
		Status   *int   `json:"status_code"`
		HasMore  *bool  `json:"has_more"`
		Cursor   string `json:"cursor"`
		Videos   *[]struct {
			ID                string `json:"id"`
			Title             string `json:"title"`
			Description       string `json:"description"`
			Author            string `json:"author"`
			Cover             string `json:"cover"`
			PublishedAt       int64  `json:"published_at"`
			UnavailableReason string `json:"unavailable_reason"`
		} `json:"videos"`
	}
	if len(raw) > 4<<20 || json.Unmarshal(raw, &value) != nil || value.Status == nil || value.HasMore == nil || value.Videos == nil {
		return app.ListingPage{}, discoveryUnavailable("Douyin selected-folder metadata is missing or malformed", "inspect_selected_douyin_folder")
	}
	if value.OwnerID != owner || value.FolderID != folder {
		return app.ListingPage{}, browserScopeDenied("The browser account or favorite folder changed during synchronization")
	}
	if *value.Status != 0 {
		return app.ListingPage{}, discoveryUnavailable("Douyin denied the selected-folder metadata request", "open_selected_douyin_folder")
	}
	if len(*value.Videos) > limit || (*value.HasMore && (!browserDouyinCursor.MatchString(value.Cursor) || value.Cursor == cursor || len(*value.Videos) == 0)) {
		return app.ListingPage{}, discoveryUnavailable("Douyin pagination did not advance within the requested metadata cap", "inspect_selected_douyin_folder")
	}
	items := []domain.ListingItem{}
	for _, video := range *value.Videos {
		if !decimalID.MatchString(video.ID) || video.PublishedAt < 0 || len(video.Description) > 128<<10 || len(video.Title) > 8<<10 {
			return app.ListingPage{}, discoveryUnavailable("Douyin video metadata has no stable identity or exceeds its bounds", "inspect_selected_douyin_folder")
		}
		item := domain.ListingItem{ExternalID: video.ID, Title: video.Title, Description: video.Description, Author: video.Author, Cover: video.Cover, PublishedAt: video.PublishedAt, UnavailableReason: video.UnavailableReason}
		if video.UnavailableReason == "" {
			item.Locator = "https://www.douyin.com/video/" + video.ID
		}
		items = append(items, item)
	}
	items, err := domain.UniqueListingItems(items)
	if err != nil {
		return app.ListingPage{}, discoveryUnavailable("Douyin returned conflicting metadata for the same video", "inspect_selected_douyin_folder")
	}
	page := app.ListingPage{Items: []app.ListingMetadata{}, Warnings: []string{}, Observed: len(*value.Videos), HasMore: *value.HasMore}
	if page.HasMore {
		page.NextCursor = &value.Cursor
	}
	for _, item := range items {
		page.Items = append(page.Items, app.ListingMetadata{ExternalID: item.ExternalID, Locator: item.Locator, Title: item.Title, Description: item.Description, Author: item.Author, Cover: item.Cover, PublishedAt: item.PublishedAt, UnavailableReason: item.UnavailableReason})
	}
	return page, nil
}
