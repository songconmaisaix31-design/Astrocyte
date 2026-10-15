package main

import (
	"github.com/songconmaisaix31-design/Astrocyte/internal/adapters/events"
	"testing"
)

func TestPopulationCountsRetainedUnknownAndUnassignedSeparately(t *testing.T) {
	d := &diagnosis{Counts: map[string]int{}, Kinds: map[string]int{}, Topics: map[string]int{}}
	cases := []struct{ project, raw string }{
		{"/Astrocyte", `{"type":"response_item","payload":{"type":"custom_tool_call","name":"apply_patch","input":"LOCAL_RAW_SENTINEL"}}`},
		{"/outside", `{"type":"response_item","payload":{"type":"custom_tool_call","name":"apply_patch","input":"LOCAL_RAW_SENTINEL"}}`},
		{"/Astrocyte", `{"type":"response_item","payload":{"type":"custom_tool_call","name":"unknown","input":"LOCAL_RAW_SENTINEL"}}`},
	}
	for _, v := range cases {
		row, err := events.MapNativeToolRecord("codex", "fixture", v.project, []byte(v.raw), 0, events.NativeAssociation{Session: "fixture", Scope: v.project}, 0)
		if err != nil {
			t.Fatal("fixture rejected")
		}
		countPopulationRow(row, d)
	}
	if d.Counts["retained"] != 3 || d.Counts["mapped"] != 2 || d.Counts["unknown_kind"] != 1 || d.Counts["kind_topic_complete"] != 1 || d.Counts["unassigned"] != 1 || d.Counts["signal_eligible_strength_1"] != 1 {
		t.Fatal("population stages collapsed")
	}
}
