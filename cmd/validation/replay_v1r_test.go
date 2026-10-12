package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/songconmaisaix31-design/Astrocyte/internal/adapters/events"
)

func TestV1RInputBoundsAndOriginalV1Compatibility(t *testing.T) {
	population := filepath.Join(v1EvidenceRoot, "v21-w1")
	git := filepath.Join(v1EvidenceRoot, "diagnosis-w1/p0-git/git-log-local.jsonl")
	source, out := filepath.Join(v1rRoot, "source-test"), filepath.Join(v1rRoot, "replay-test")
	if err := validateV1RPaths(population, git, source, out); err != nil {
		t.Fatal(err)
	}
	for _, paths := range [][4]string{
		{"C:/personal", git, source, out},
		{population, "C:/personal/git.jsonl", source, out},
		{population, git, "C:/personal/source", out},
		{population, git, source, filepath.Join(v1EvidenceRoot, "v1-w1/replay")},
		{population, git, source, source},
		{population, git, filepath.Join(source, "nested"), out},
	} {
		if validateV1RPaths(paths[0], paths[1], paths[2], paths[3]) == nil {
			t.Fatal("V1R loosened the fixed input/evidence bounds")
		}
	}
	// The old V1 mode rejects V1R output roots before opening any source.
	if runV1(population, git, out) == nil {
		t.Fatal("original V1 accepted an expanded output root")
	}
	if _, err := readMcPetSource(t.TempDir()); err == nil {
		t.Fatal("missing explicit source accepted")
	}
}

func TestV1RFixedSeedMcPetSupplementExcludesPrimary(t *testing.T) {
	for _, primaryTopic := range []string{"Morphogenesis", "McPet"} {
		out := t.TempDir()
		primary := []map[string]any{{"replay_id": "primary", "event": replayEventRef{EventID: "primary", Topic: primaryTopic}}}
		if err := writeExclusive(filepath.Join(out, "weights-50-safe.json"), map[string]any{"samples": primary}); err != nil {
			t.Fatal(err)
		}
		f, err := os.Create(filepath.Join(out, "operations-safe.jsonl"))
		if err != nil {
			t.Fatal(err)
		}
		for i := 0; i < 15; i++ {
			id := fmt.Sprintf("candidate-%02d", i)
			if err := json.NewEncoder(f).Encode(replayOperation{Phase: "event", ReplayID: id, Event: &replayEventRef{EventID: id, Topic: "McPet"}, Strength: 1, Weight: .6}); err != nil {
				t.Fatal(err)
			}
		}
		f.Close()
		if err := writeMcPetSupplement(out); err != nil {
			t.Fatal(err)
		}
		var sample struct {
			Count   int `json:"supplement_count"`
			Samples []struct {
				ReplayID string `json:"replay_id"`
			} `json:"samples"`
		}
		readReplayJSON(t, filepath.Join(out, "mcpet-supplement-safe.json"), &sample)
		want := 10
		if primaryTopic == "McPet" {
			want = 0
		}
		if sample.Count != want || len(sample.Samples) != want {
			t.Fatal("McPet supplementary sample violates primary coverage rule")
		}
		ids := map[string]bool{}
		for _, row := range sample.Samples {
			if row.ReplayID == "primary" || ids[row.ReplayID] {
				t.Fatal("supplement overlaps itself or primary")
			}
			ids[row.ReplayID] = true
		}
	}
}

func TestV1RCommitFactsDeduplicateWithoutDoubleCounting(t *testing.T) {
	base := t.TempDir()
	root := filepath.Join(base, "maimai-monster")
	if err := os.Mkdir(root, 0700); err != nil {
		t.Fatal(err)
	}
	if b, err := exec.Command("git", "-C", root, "init", "-q").CombinedOutput(); err != nil {
		t.Fatalf("synthetic Git init: %v %s", err, b)
	}
	mapper := events.NewNativeSemanticMapper()
	first := events.GitLogRecord{Commit: strings.Repeat("a", 40), TS: v1Window().UnixMilli() + 1000, Subject: "synthetic source fact"}
	second := events.GitLogRecord{Commit: strings.Repeat("b", 40), TS: first.TS, Subject: "synthetic second fact"}
	raw, _ := json.Marshal(first)
	old, err := mapper.MapGit(root, raw, 0)
	if err != nil {
		t.Fatal(err)
	}
	write := func(records []events.GitLogRecord) string {
		t.Helper()
		dir := t.TempDir()
		f, err := os.Create(filepath.Join(dir, "git-log-local.jsonl"))
		if err != nil {
			t.Fatal(err)
		}
		defer f.Close()
		for _, record := range records {
			if err := json.NewEncoder(f).Encode(map[string]any{"project": root, "record": record}); err != nil {
				t.Fatal(err)
			}
		}
		return dir
	}
	source := mcpetSource{Root: root, UniqueCommits: 2, WindowStart: v1Window().UnixMilli(), WindowEnd: v1Window().AddDate(0, 0, 30).UnixMilli()}
	merged, added, duplicates, err := mergeMcPetFacts([]replayInput{{Row: old}}, source, write([]events.GitLogRecord{first, second}), mapper)
	if err != nil || len(merged) != 2 || added != 1 || duplicates != 1 || merged[0].StableID >= merged[1].StableID {
		t.Fatalf("additional facts double-counted: added=%d duplicates=%d error=%v", added, duplicates, err)
	}
	changed := first
	changed.Subject = "different source bytes"
	if _, _, _, err := mergeMcPetFacts([]replayInput{{Row: old}}, source, write([]events.GitLogRecord{changed, second}), mapper); err == nil {
		t.Fatal("same SHA changing its committed fact was accepted")
	}
	if _, _, _, err := mergeMcPetFacts(nil, source, write([]events.GitLogRecord{first, first}), mapper); err == nil {
		t.Fatal("duplicate within captured source was accepted")
	}
}
