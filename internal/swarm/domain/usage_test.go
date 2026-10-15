package domain

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestUnknownUsageRemainsNullAndBoundedDimensionsRequireEvidence(t *testing.T) {
	calls, limit, tokens := int64(1), int64(4), int64(128)
	u := Usage{Calls: &calls}
	data, err := json.Marshal(u)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(data), `"tokens":null`) || !strings.Contains(string(data), `"money":null`) {
		t.Fatal("unknown usage was serialized as a numeric zero")
	}
	var restored Usage
	if err = json.Unmarshal(data, &restored); err != nil {
		t.Fatal(err)
	}
	if restored.Tokens != nil || restored.Money != nil {
		t.Fatal("unknown usage did not survive JSON persistence")
	}
	b := Budget{Calls: &limit}
	if !u.Within(b) || !u.WithinReservation(Resources{Calls: 1}, b) {
		t.Fatal("known call usage was rejected by an unbounded unknown token dimension")
	}
	b.Tokens = &tokens
	if u.Within(b) || u.WithinReservation(Resources{Calls: 1, Tokens: 128}, b) {
		t.Fatal("bounded tokens accepted missing actual token evidence")
	}
	if (Usage{}).Within(Budget{Calls: &limit}) {
		t.Fatal("bounded calls accepted missing actual call evidence")
	}
	u.Tokens = &tokens
	if !u.WithinReservation(Resources{Calls: 1}, Budget{Calls: &limit}) {
		t.Fatal("unbounded actual tokens were treated as a zero token limit")
	}
}

func TestActualUsageAggregationPreservesUnknownPriorDimensions(t *testing.T) {
	limit, calls, knownTokens, money := int64(4), int64(1), int64(128), int64(5)
	m := Mission{Grant: Grant{Budget: Budget{Calls: &limit}}}
	if err := m.RecordUsage(Usage{Calls: &calls}); err != nil {
		t.Fatal(err)
	}
	if err := m.RecordUsage(Usage{Calls: &calls, Tokens: &knownTokens, Money: &money, Currency: "USD"}); err != nil {
		t.Fatal(err)
	}
	if m.Used.Calls == nil || *m.Used.Calls != 2 || m.Used.Tokens != nil || m.Used.Money != nil || m.SettledOperations != 2 {
		t.Fatal("aggregate usage replaced unknown prior dimensions with zero")
	}
	m = Mission{Grant: Grant{Budget: Budget{Calls: &limit}}}
	if err := m.RecordUsage(Usage{Calls: &calls, Money: &money, Currency: "USD"}); err != nil {
		t.Fatal(err)
	}
	if m.Used.Money == nil || *m.Used.Money != 5 || m.Used.Currency != "USD" {
		t.Fatal("first actual known monetary usage was lost")
	}
}
