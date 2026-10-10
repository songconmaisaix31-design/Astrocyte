package importers

import (
	"context"
	"os"
	"testing"
)

func TestScholarSearchValidation(t *testing.T) {
	s := NewScholar()
	ctx := context.Background()
	cases := []struct {
		name     string
		provider ScholarProvider
		query    string
		limit    int
	}{
		{"empty query", ScholarCrossref, "   ", 10},
		{"oversize query", ScholarCrossref, string(make([]byte, 513)), 10},
		{"zero limit", ScholarCrossref, "attention", 0},
		{"oversize limit", ScholarCrossref, "attention", 51},
		{"bad provider", ScholarProvider("scopus"), "attention", 10},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if _, err := s.Search(ctx, c.provider, c.query, c.limit); err == nil {
				t.Fatal("expected validation error")
			}
		})
	}
}

func TestParseCrossref(t *testing.T) {
	raw := []byte(`{"message":{"items":[{"DOI":"10.1038/s41586-020-2649-2","title":["Highly accurate protein structure prediction with AlphaFold"],"author":[{"given":"John","family":"Jumper"},{"given":"Richard","family":"Evans"}],"abstract":"Proteins are essential.","URL":"https://www.nature.com/articles/s41586-020-2649-2","published":{"date-parts":[[2021,7,15]]},"container-title":["Nature"]}]}}`)
	hits, err := parseCrossref(raw)
	if err != nil {
		t.Fatal(err)
	}
	if len(hits) != 1 {
		t.Fatalf("expected 1 hit, got %d", len(hits))
	}
	h := hits[0]
	if h.DOI != "10.1038/s41586-020-2649-2" || h.SourceKey != "doi:10.1038/s41586-020-2649-2" || h.Year != 2021 || h.Venue != "Nature" || len(h.Authors) != 2 || h.Authors[0] != "John Jumper" {
		t.Fatalf("crossref fields wrong: %+v", h)
	}
}

func TestParseEuropePMC(t *testing.T) {
	raw := []byte(`{"resultList":{"result":[{"id":"33334934","title":"A test paper","authorString":"Jumper J; Evans R","abstractText":"Abstract here.","doi":"10.1371/journal.pdig.0000514","pubYear":2023,"journalTitle":"PLOS Digital Health","pmcid":"PMC11135672","fullTextUrlList":{"fullTextUrl":[{"url":"https://www.ncbi.nlm.nih.gov/pmc/articles/PMC11135672/pdf/"}]}}]}}`)
	hits, err := parseEuropePMC(raw)
	if err != nil {
		t.Fatal(err)
	}
	if len(hits) != 1 {
		t.Fatalf("expected 1 hit, got %d", len(hits))
	}
	h := hits[0]
	if h.DOI != "10.1371/journal.pdig.0000514" || h.SourceKey != "doi:10.1371/journal.pdig.0000514" || h.Year != 2023 || len(h.Authors) != 2 || len(h.PDFURLs) != 1 {
		t.Fatalf("europepmc fields wrong: %+v", h)
	}
	if h.Locator != "https://pmc.ncbi.nlm.nih.gov/articles/PMC11135672/" {
		t.Fatalf("europepmc locator wrong: %q", h.Locator)
	}
}

func TestParseArxivSearch(t *testing.T) {
	raw := []byte(`<?xml version="1.0"?><feed xmlns="http://www.w3.org/2005/Atom"><entry><id>https://arxiv.org/abs/2504.16054v1</id><title> A   Sample Paper </title><summary>Abstract text.</summary><published>2025-04-16T00:00:00Z</published><author><name>Alice Author</name></author></entry></feed>`)
	hits, err := parseArxivSearch(raw)
	if err != nil {
		t.Fatal(err)
	}
	if len(hits) != 1 {
		t.Fatalf("expected 1 hit, got %d", len(hits))
	}
	h := hits[0]
	if h.ArxivID != "2504.16054v1" || h.SourceKey != "arxiv:2504.16054" || h.Year != 2025 || h.Locator != "https://arxiv.org/abs/2504.16054v1" || len(h.PDFURLs) != 1 {
		t.Fatalf("arxiv fields wrong: %+v", h)
	}
	if h.Title != "A Sample Paper" {
		t.Fatalf("arxiv title not normalized: %q", h.Title)
	}
}

// TestScholarSearchLive performs one credential-free query per provider. It runs
// only with an explicit opt-in and is an observation of real public metadata,
// never an import or a claim of full-text coverage.
func TestScholarSearchLive(t *testing.T) {
	if os.Getenv("ASTROCYTE_PAPER_SEARCH_LIVE") != "1" {
		t.Skip("live public metadata search not requested")
	}
	s := NewScholar()
	ctx := context.Background()
	for _, provider := range []ScholarProvider{ScholarCrossref, ScholarEuropePMC, ScholarArxiv} {
		t.Run(string(provider), func(t *testing.T) {
			hits, err := s.Search(ctx, provider, "attention mechanism transformer", 5)
			if err != nil {
				t.Fatalf("live %s search failed: %v", provider, err)
			}
			if len(hits) == 0 || len(hits) > 5 {
				t.Fatalf("live %s returned unexpected count %d", provider, len(hits))
			}
			for _, h := range hits {
				if h.Title == "" || h.SourceKey == "" || h.ContentState != "abstract_only" {
					t.Fatalf("live %s hit incomplete: %+v", provider, h)
				}
			}
			t.Logf("live %s returned %d hits; first=%q source=%s", provider, len(hits), hits[0].Title, hits[0].SourceKey)
		})
	}
}
