package domain

import (
	"bytes"
	"encoding/json"
	"testing"
	"time"
)

func equalRaw(a, b json.RawMessage) bool { return bytes.Equal(a, b) }

func TestConfirmKnowledge(t *testing.T) {
	if got, err := ConfirmKnowledge(KnowledgeInbox); err != nil || got != KnowledgeConfirmed {
		t.Fatalf("inbox -> %v, %v", got, err)
	}
	if got, err := ConfirmKnowledge(KnowledgeConfirmed); err != nil || got != KnowledgeConfirmed {
		t.Fatalf("confirmed -> %v, %v", got, err)
	}
	if _, err := ConfirmKnowledge("rejected"); err == nil {
		t.Fatal("unknown status must be rejected")
	}
}

func TestReferencableKnowledge(t *testing.T) {
	if ReferencableKnowledge(KnowledgeInbox) {
		t.Fatal("inbox must not be referenceable")
	}
	if !ReferencableKnowledge(KnowledgeConfirmed) {
		t.Fatal("confirmed must be referenceable")
	}
}

func TestAppendCrystallizedAppendsWithoutMutating(t *testing.T) {
	first := CrystallizedVersion{Content: json.RawMessage(`{"fact":"a"}`), RecordedAt: time.Unix(100, 0)}
	history := []CrystallizedVersion{first}
	orig := append([]CrystallizedVersion(nil), history...)

	second, err := AppendCrystallized(history, json.RawMessage(`{"fact":"b"}`), time.Unix(200, 0))
	if err != nil {
		t.Fatal(err)
	}
	if len(second) != 2 {
		t.Fatalf("len = %d, want 2", len(second))
	}
	if second[1].RecordedAt != time.Unix(200, 0) {
		t.Fatal("appended version wrong")
	}
	if len(history) != 1 || !equalRaw(history[0].Content, orig[0].Content) || history[0].RecordedAt != orig[0].RecordedAt {
		t.Fatal("input history was mutated")
	}
}

func TestAppendCrystallizedRejectsNonMonotonicTimestamp(t *testing.T) {
	history := []CrystallizedVersion{{Content: json.RawMessage(`{}`), RecordedAt: time.Unix(200, 0)}}
	if _, err := AppendCrystallized(history, json.RawMessage(`{}`), time.Unix(100, 0)); err != ErrVersion {
		t.Fatalf("err = %v, want ErrVersion", err)
	}
	if _, err := AppendCrystallized(history, nil, time.Unix(300, 0)); err != ErrEvidence {
		t.Fatalf("empty content err = %v, want ErrEvidence", err)
	}
}

func TestAppendCrystallizedOwnsSnapshotBytes(t *testing.T) {
	first := json.RawMessage(`{"fact":"a"}`)
	history, err := AppendCrystallized(nil, first, time.Unix(100, 0))
	if err != nil {
		t.Fatal(err)
	}
	first[9] = 'x'
	if string(history[0].Content) != `{"fact":"a"}` {
		t.Fatal("caller mutation changed the stored snapshot")
	}
	second, err := AppendCrystallized(history, json.RawMessage(`{"fact":"b"}`), time.Unix(200, 0))
	if err != nil {
		t.Fatal(err)
	}
	second[0].Content[9] = 'y'
	if string(history[0].Content) != `{"fact":"a"}` {
		t.Fatal("returned history mutation changed an earlier snapshot")
	}
}

func TestKnowledgeVisible(t *testing.T) {
	if !KnowledgeVisible("p1", nil, "p1") {
		t.Fatal("owner must see own entry")
	}
	if KnowledgeVisible("p1", nil, "p2") {
		t.Fatal("default isolation must hide other projects")
	}
	if !KnowledgeVisible("p1", []string{"p2"}, "p2") {
		t.Fatal("explicit share must be visible")
	}
}
