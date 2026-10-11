package domain

import (
	"testing"
	"time"
)

func TestWorkIdentityIsStableAndDistinct(t *testing.T) {
	a := WorkIdentity("codex", "C:\\works\\app")
	b := WorkIdentity("codex", "C:\\works\\app")
	if a != b {
		t.Fatalf("identity must be stable: %q != %q", a, b)
	}
	if WorkIdentity("codex", "C:\\works\\app") == WorkIdentity("claude", "C:\\works\\app") {
		t.Fatal("platform must differentiate identity")
	}
	if WorkIdentity("codex", "C:\\works\\app") == WorkIdentity("codex", "C:\\works\\other") {
		t.Fatal("root must differentiate identity")
	}
}

func TestAttitudeAndBounds(t *testing.T) {
	for _, valid := range []string{AttitudeContinue, AttitudePark, AttitudeDone} {
		if !ValidAttitude(valid) {
			t.Fatalf("%q must be a valid attitude", valid)
		}
	}
	for _, invalid := range []string{"", "later", "CONTINUE", "done "} {
		if ValidAttitude(invalid) {
			t.Fatalf("%q must be rejected", invalid)
		}
	}
	if !ValidWorkPlatform("lingxi") || !ValidWorkPlatform("DeepSeek Harness") {
		t.Fatal("surveyed platforms must be accepted")
	}
	if ValidWorkPlatform("") || ValidWorkPlatform(string(make([]byte, 65))) {
		t.Fatal("empty or oversized platform must be rejected")
	}
	if !ValidWorkGroup("research") || ValidWorkGroup(string(make([]byte, 201))) {
		t.Fatal("group bounds are wrong")
	}
}

func TestHandedOverRequiresContent(t *testing.T) {
	if HandedOver(WorkHandoff{Content: "  "}) {
		t.Fatal("whitespace-only manual is not a handoff")
	}
	if !HandedOver(WorkHandoff{Content: "how to resume"}) {
		t.Fatal("content manual must count as handed over")
	}
}

func TestRecentlyActiveRequiresVerifiedMtime(t *testing.T) {
	now := time.Now()
	mtime := now.Add(-time.Hour)
	f := WorkFileFact{Mtime: &mtime, Verified: true}
	if !f.RecentlyActive(now, 24*time.Hour) {
		t.Fatal("recent verified mtime must count")
	}
	if f.RecentlyActive(now, time.Minute) {
		t.Fatal("mtime outside window must not count")
	}
	f.Verified = false
	if f.RecentlyActive(now, 24*time.Hour) {
		t.Fatal("unverified work must never count as recent")
	}
}

func sampleWorks() []Work {
	now := time.Now()
	recent := now.Add(-time.Hour)
	old := now.Add(-60 * 24 * time.Hour)
	return []Work{
		{ID: "w1", Platform: "codex", Name: "app", Kind: WorkKindFolder, Attitude: AttitudeContinue, Handoff: WorkHandoff{Content: "resume here"}, Files: WorkFileFact{Mtime: &recent, Verified: true}},
		{ID: "w2", Platform: "codex", Name: "deck", Kind: WorkKindFile, Attitude: AttitudeDone, Files: WorkFileFact{Mtime: &old, Verified: true}},
		{ID: "w3", Platform: "lingxi", Name: "report", Kind: WorkKindFolder, Attitude: AttitudePark, Hidden: true, Files: WorkFileFact{}},
		{ID: "w4", Platform: "lingxi", Name: "video", Kind: WorkKindFolder, Attitude: AttitudePark, Group: "media", Handoff: WorkHandoff{Content: "prompt kept"}, Files: WorkFileFact{Mtime: &recent, Verified: true}},
	}
}

func TestAggregateWorksStats(t *testing.T) {
	now := time.Now()
	stats := AggregateWorksStats(sampleWorks(), now, 30*24*time.Hour)
	if len(stats) != 2 {
		t.Fatalf("expected 2 platforms, got %d", len(stats))
	}
	if stats[0].Platform != "codex" || stats[1].Platform != "lingxi" {
		t.Fatalf("platforms must be sorted, got %q %q", stats[0].Platform, stats[1].Platform)
	}
	var codex, lingxi PlatformStat
	for _, s := range stats {
		switch s.Platform {
		case "codex":
			codex = s
		case "lingxi":
			lingxi = s
		}
	}
	if codex.Works != 2 || codex.Folders != 1 || codex.HandedOver != 1 || codex.Recent != 1 {
		t.Fatalf("codex stats wrong: %+v", codex)
	}
	if lingxi.Works != 2 || lingxi.Folders != 2 || lingxi.HandedOver != 1 || lingxi.Recent != 1 {
		t.Fatalf("lingxi stats wrong: %+v", lingxi)
	}
}

func TestFilterWorks(t *testing.T) {
	works := sampleWorks()
	if got := FilterWorks(works, WorkFilter{}); len(got) != 3 {
		t.Fatalf("hidden work excluded by default, got %d", len(got))
	}
	if got := FilterWorks(works, WorkFilter{IncludeHidden: true}); len(got) != 4 {
		t.Fatalf("include hidden expected 4, got %d", len(got))
	}
	if got := FilterWorks(works, WorkFilter{Platform: "lingxi", IncludeHidden: true}); len(got) != 2 {
		t.Fatalf("platform filter expected 2, got %d", len(got))
	}
	if got := FilterWorks(works, WorkFilter{Attitude: AttitudePark, IncludeHidden: true}); len(got) != 2 {
		t.Fatalf("attitude filter expected 2, got %d", len(got))
	}
	if got := FilterWorks(works, WorkFilter{Group: "media", IncludeHidden: true}); len(got) != 1 {
		t.Fatalf("group filter expected 1, got %d", len(got))
	}
}

func TestSortWorks(t *testing.T) {
	works := sampleWorks()
	byName := SortWorks(works, "name")
	if byName[0].Name != "app" || byName[len(byName)-1].Name != "video" {
		t.Fatalf("name sort wrong: %v", names(byName))
	}
	byRecent := SortWorks(works, "recent")
	if byRecent[0].Name != "app" {
		t.Fatalf("recent sort must put verified recent first, got %q", byRecent[0].Name)
	}
	byUpdated := SortWorks(works, "updated")
	if len(byUpdated) != 4 {
		t.Fatalf("updated sort must keep all works, got %d", len(byUpdated))
	}
}

func names(works []Work) []string {
	out := make([]string, len(works))
	for i, w := range works {
		out[i] = w.Name
	}
	return out
}
