package domain

import "testing"

func TestPaperSearchQueryValidate(t *testing.T) {
	cases := []struct {
		name string
		q    PaperSearchQuery
		ok   bool
	}{
		{"valid", PaperSearchQuery{Provider: "crossref", Query: "attention", Limit: 10}, true},
		{"empty provider", PaperSearchQuery{Provider: "", Query: "x", Limit: 10}, false},
		{"empty query", PaperSearchQuery{Provider: "arxiv", Query: "", Limit: 10}, false},
		{"oversize query", PaperSearchQuery{Provider: "arxiv", Query: string(make([]byte, 513)), Limit: 10}, false},
		{"zero limit", PaperSearchQuery{Provider: "arxiv", Query: "x", Limit: 0}, false},
		{"oversize limit", PaperSearchQuery{Provider: "arxiv", Query: "x", Limit: 51}, false},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if err := c.q.Validate(); (err == nil) != c.ok {
				t.Fatalf("Validate() = %v, want ok=%v", err, c.ok)
			}
		})
	}
}
