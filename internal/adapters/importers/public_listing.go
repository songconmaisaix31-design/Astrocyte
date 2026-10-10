package importers

import (
	"context"
	"strconv"
	"strings"

	"github.com/songconmaisaix31-design/Astrocyte/internal/apierrors"
	"github.com/songconmaisaix31-design/Astrocyte/internal/attention/app"
)

type PublicListing struct {
	Bilibili *PublicBilibili
	Uploads  *BilibiliUploads
}

var _ app.PublicListingReader = (*PublicListing)(nil)
var _ app.PublicCollectionReader = (*PublicListing)(nil)

func NewPublicListingReader(pythonPath string) *PublicListing {
	return &PublicListing{Bilibili: NewPublicBilibili(), Uploads: &BilibiliUploads{PythonPath: pythonPath}}
}

func (r *PublicListing) ReadPage(ctx context.Context, source app.TrackingSource, cursor string, limit int) (app.ListingPage, error) {
	if source.Platform != "bilibili" {
		return app.ListingPage{}, &apierrors.ServiceError{Code: apierrors.UnsupportedCapability, Message: "The current upstream adapter has no anonymous Douyin public-profile/favorite listing transport", RequiredAction: "provide_supported_public_source"}
	}
	if !decimalID.MatchString(source.ExternalID) || limit < 1 || limit > 100 {
		return app.ListingPage{}, invalid("a bound public identity and limit from1to100 are required")
	}
	window := 30
	if source.SourceKind == "favorites" {
		window = 20
	}
	page, offset := 1, 0
	if cursor != "" {
		parts := strings.Split(cursor, ":")
		var err error
		page, err = strconv.Atoi(parts[0])
		if err != nil || len(parts) > 2 || page < 1 || page > 100000 {
			return app.ListingPage{}, invalid("invalid public listing cursor")
		}
		if len(parts) == 2 {
			offset, err = strconv.Atoi(parts[1])
			if err != nil || offset < 1 || offset >= window {
				return app.ListingPage{}, invalid("invalid public listing offset")
			}
		}
	}
	start := (page-1)*window + offset
	// Page-number APIs use (pn-1)*ps. Choose a divisor of the exact offset,
	// bounded by remaining observations, so the final request never fetches
	// a hidden extra 20 rows merely to return a 100-item view.
	size := min(window, limit)
	for size > 1 && start%size != 0 {
		size--
	}
	providerPage := start/size + 1
	var raw DiscoveryPage
	var err error
	switch source.SourceKind {
	case "uploads":
		if source.OwnerID != source.ExternalID {
			return app.ListingPage{}, invalid("Bilibili uploads owner differs from its public UID")
		}
		raw, err = r.Uploads.PageSize(ctx, source.ExternalID, providerPage, size)
	case "favorites":
		if source.OwnerID != "" {
			folders, e := r.Bilibili.Collections(ctx, source.OwnerID)
			if e != nil {
				return app.ListingPage{}, e
			}
			matched := false
			for _, folder := range folders {
				if folder.ExternalID == source.ExternalID {
					matched = true
					break
				}
			}
			if !matched {
				return app.ListingPage{}, &apierrors.ServiceError{Code: apierrors.ScopeDenied, Message: "The bound folder is not publicly listed for this account", RequiredAction: "choose_public_folder_for_account"}
			}
		}
		raw, err = r.Bilibili.PageSize(ctx, source.ExternalID, providerPage, size)
	default:
		return app.ListingPage{}, invalid("unsupported public source kind")
	}
	if err != nil {
		return app.ListingPage{}, err
	}
	if raw.Observed > size || len(raw.Items) > raw.Observed {
		return app.ListingPage{}, discoveryUnavailable("provider exceeded the requested public metadata page size", "inspect_public_listing")
	}
	items := raw.Items
	result := app.ListingPage{Items: []app.ListingMetadata{}, HasMore: raw.HasMore, Warnings: raw.Warnings, Observed: raw.Observed}
	if raw.HasMore {
		position := start + raw.Observed
		next := strconv.Itoa(position/window + 1)
		if position%window != 0 {
			next += ":" + strconv.Itoa(position%window)
		}
		result.NextCursor = &next
	}
	for _, item := range items {
		result.Items = append(result.Items, app.ListingMetadata{ExternalID: item.ExternalID, Locator: item.Locator, Title: item.Title, Description: item.Description, Author: item.Author, Cover: item.Cover, PublishedAt: item.PublishedAt, ProviderStatus: item.ProviderStatus, UnavailableReason: item.UnavailableReason})
	}
	return result, nil
}

func (r *PublicListing) ListCollections(ctx context.Context, platform, ownerID string) ([]app.SourceCollection, error) {
	if platform != "bilibili" {
		return nil, &apierrors.ServiceError{Code: apierrors.UnsupportedCapability, Message: "Anonymous public collections are only supported for Bilibili", RequiredAction: "choose_supported_public_collection"}
	}
	folders, err := r.Bilibili.Collections(ctx, ownerID)
	if err != nil {
		return nil, err
	}
	result := []app.SourceCollection{}
	for _, folder := range folders {
		count := folder.ItemCount
		result = append(result, app.SourceCollection{ExternalID: folder.ExternalID, OwnerID: folder.OwnerID, Title: folder.Title, Locator: folder.Locator, ItemCount: &count})
	}
	return result, nil
}
