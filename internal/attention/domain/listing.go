package domain

import (
	"errors"
	"reflect"
	"strings"
)

// ListingItem is provider metadata, not a material revision, recommendation,
// authorization grant or a human attention event.
type ListingItem struct {
	ExternalID, Locator, Title, Description, Author, Cover string
	PublishedAt                                            int64
	ProviderStatus                                         *int
	UnavailableReason                                      string
}

type ListingRecord struct {
	Item     ListingItem
	Revision int
}

type ListingSnapshot struct {
	Items    []ListingItem
	Complete bool
}

type ListingChange struct {
	Record ListingRecord
	Kind   string // new, updated, unchanged
}

type ListingComparison struct {
	Records []ListingRecord
	Changes []ListingChange
	Missing []string // Observed absence, not permission to remove cached data.
}

var ErrIncompleteListing = errors.New("listing_not_complete")

// UniqueListingItems collapses overlapping page entries. Conflicting metadata
// for one ID means the listing changed while paging; don't pick an arbitrary
// winner or let an incomplete observation replace the previous cache.
func UniqueListingItems(items []ListingItem) ([]ListingItem, error) {
	result := []ListingItem{}
	indices := map[string]int{}
	for _, item := range items {
		if strings.TrimSpace(item.ExternalID) == "" || (strings.TrimSpace(item.Locator) == "" && item.UnavailableReason == "") {
			return nil, ErrInvalid
		}
		if i, ok := indices[item.ExternalID]; ok {
			if !reflect.DeepEqual(result[i], item) {
				return nil, ErrVersion
			}
			continue
		}
		indices[item.ExternalID] = len(result)
		result = append(result, cloneListingItem(item))
	}
	return result, nil
}

func cloneListingItem(item ListingItem) ListingItem {
	if item.ProviderStatus != nil {
		status := *item.ProviderStatus
		item.ProviderStatus = &status
	}
	return item
}

// CompareListing applies only complete successful observations. Callers must
// commit the result in their existing transaction; failed/partial reads leave
// prior metadata, versions, selections and successful cursors untouched.
// Missing entries stay cached until removal semantics are explicitly decided.
func CompareListing(previous []ListingRecord, snapshot ListingSnapshot) (ListingComparison, error) {
	result := ListingComparison{Records: []ListingRecord{}, Changes: []ListingChange{}, Missing: []string{}}
	indices := map[string]int{}
	for _, record := range previous {
		if strings.TrimSpace(record.Item.ExternalID) == "" || record.Revision < 1 {
			return ListingComparison{}, ErrInvalid
		}
		if _, exists := indices[record.Item.ExternalID]; exists {
			return ListingComparison{}, ErrInvalid
		}
		indices[record.Item.ExternalID] = len(result.Records)
		record.Item = cloneListingItem(record.Item)
		result.Records = append(result.Records, record)
	}
	if !snapshot.Complete {
		return result, ErrIncompleteListing
	}
	items, err := UniqueListingItems(snapshot.Items)
	if err != nil {
		return result, err
	}
	seen := map[string]bool{}
	for _, item := range items {
		seen[item.ExternalID] = true
		record, kind := ListingRecord{Item: item, Revision: 1}, "new"
		if index, exists := indices[item.ExternalID]; exists {
			old := result.Records[index]
			record.Revision, kind = old.Revision, "unchanged"
			if !reflect.DeepEqual(old.Item, item) {
				record.Revision++
				kind = "updated"
			}
			result.Records[index] = record
		} else {
			result.Records = append(result.Records, record)
		}
		record.Item = cloneListingItem(record.Item)
		result.Changes = append(result.Changes, ListingChange{Record: record, Kind: kind})
	}
	for _, record := range previous {
		if !seen[record.Item.ExternalID] {
			result.Missing = append(result.Missing, record.Item.ExternalID)
		}
	}
	return result, nil
}
