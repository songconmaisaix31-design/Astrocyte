package domain

import (
	"errors"
	"reflect"
	"testing"
)

func TestListingUpdatesOnlyObservedMetadataAndRetainsPriorItems(t *testing.T) {
	one := ListingItem{ExternalID: "provider1", Locator: "https://source/1", Title: "old"}
	two := ListingItem{ExternalID: "provider2", Locator: "https://source/2", Title: "retained"}
	previous := []ListingRecord{{Item: one, Revision: 3}, {Item: two, Revision: 2}}
	updated := one
	updated.Title = "new title"
	newItem := ListingItem{ExternalID: "provider3", Locator: "https://source/3"}
	result, err := CompareListing(previous, ListingSnapshot{Complete: true, Items: []ListingItem{updated, newItem, updated}})
	if err != nil || len(result.Records) != 3 || len(result.Changes) != 2 || !reflect.DeepEqual(result.Missing, []string{"provider2"}) {
		t.Fatalf("complete overlapping listing: %+v %v", result, err)
	}
	if result.Changes[0].Kind != "updated" || result.Changes[0].Record.Revision != 4 || result.Changes[1].Kind != "new" || result.Changes[1].Record.Revision != 1 || result.Records[1] != previous[1] || previous[0].Item.Title != "old" {
		t.Fatalf("revision/retention semantics: %+v previous=%+v", result, previous)
	}
	replay, err := CompareListing(result.Records, ListingSnapshot{Complete: true, Items: []ListingItem{updated, newItem}})
	if err != nil || replay.Changes[0].Kind != "unchanged" || replay.Changes[0].Record.Revision != 4 || !reflect.DeepEqual(replay.Records, result.Records) {
		t.Fatalf("same metadata advanced revision: %+v %v", replay, err)
	}
}

func TestPartialOrConflictingListingNeverReplacesCache(t *testing.T) {
	old := ListingItem{ExternalID: "provider1", Locator: "https://source/1", Title: "keep"}
	previous := []ListingRecord{{Item: old, Revision: 7}}
	changed := old
	changed.Title = "new"
	for _, snapshot := range []ListingSnapshot{{Items: []ListingItem{changed}}, {Complete: true, Items: []ListingItem{old, changed}}} {
		result, err := CompareListing(previous, snapshot)
		if err == nil || !reflect.DeepEqual(previous, result.Records) || len(result.Changes) != 0 || len(result.Missing) != 0 {
			t.Fatalf("unsafe partial observation: %+v %v", result, err)
		}
	}
	result, err := CompareListing(previous, ListingSnapshot{Complete: true})
	if err != nil || !reflect.DeepEqual(result.Records, previous) || !reflect.DeepEqual(result.Missing, []string{"provider1"}) {
		t.Fatalf("successful empty listing destroyed cache: %+v %v", result, err)
	}
	_, err = CompareListing(previous, ListingSnapshot{})
	if !errors.Is(err, ErrIncompleteListing) {
		t.Fatal(err)
	}
}

func TestListingStateDoesNotShareMutableProviderStatus(t *testing.T) {
	status := 5
	item := ListingItem{ExternalID: "provider1", Locator: "https://source/1", ProviderStatus: &status}
	result, err := CompareListing(nil, ListingSnapshot{Complete: true, Items: []ListingItem{item}})
	if err != nil {
		t.Fatal(err)
	}
	status = 1
	*result.Changes[0].Record.Item.ProviderStatus = 2
	if *result.Records[0].Item.ProviderStatus != 5 {
		t.Fatal("metadata changed after provider buffer/returned change was modified")
	}
}
