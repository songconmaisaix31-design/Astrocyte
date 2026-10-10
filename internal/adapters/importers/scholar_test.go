package importers

import (
	"context"
	"errors"
	"io"
	"net/http"
	"os"
	"strings"
	"testing"

	"github.com/songconmaisaix31-design/Astrocyte/internal/apierrors"
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

// newMockScholar returns a Scholar whose HTTP client is stubbed by handler so
// protocol regression tests can assert the exact endpoint an identifier query
// is routed to without touching the network.
func newMockScholar(handler func(*http.Request) (int, string)) *Scholar {
	s := NewScholar()
	s.Client = &http.Client{
		CheckRedirect: scholarRedirect,
		Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
			status, body := handler(r)
			return &http.Response{
				StatusCode: status,
				Body:       io.NopCloser(strings.NewReader(body)),
				Header:     http.Header{},
				Request:    r,
			}, nil
		}),
	}
	return s
}

func TestParseDOIQuery(t *testing.T) {
	cases := []struct{ in, want string }{
		{"10.1371/journal.pdig.0000514", "10.1371/journal.pdig.0000514"},
		{"doi:10.1371/journal.pdig.0000514", "10.1371/journal.pdig.0000514"},
		{"https://doi.org/10.1371/journal.pdig.0000514", "10.1371/journal.pdig.0000514"},
		{"http://dx.doi.org/10.1371/journal.pdig.0000514", "10.1371/journal.pdig.0000514"},
		{"10.7717/peerj-cs.3663/table-101", "10.7717/peerj-cs.3663/table-101"},
		{"attention mechanism transformer", ""},
		{"10.1234", ""},
		{"2504.16054", ""},
		{"https://example.com/10.1371/journal.pdig.0000514", ""},
		{"", ""},
	}
	for _, c := range cases {
		if got := parseDOIQuery(c.in); got != c.want {
			t.Errorf("parseDOIQuery(%q) = %q, want %q", c.in, got, c.want)
		}
	}
}

func TestScholarExactDOIRouting(t *testing.T) {
	var gotPath string
	s := newMockScholar(func(r *http.Request) (int, string) {
		gotPath = r.URL.EscapedPath()
		return http.StatusOK, `{"status":"ok","message":{"DOI":"10.1371/journal.pdig.0000514","title":["A test paper"],"author":[{"given":"Ada","family":"Lovelace"}],"URL":"https://journals.plos.org/digitalhealth/article?id=10.1371/journal.pdig.0000514","published":{"date-parts":[[2023,5,30]]},"container-title":["PLOS Digital Health"]}}`
	})
	hits, err := s.Search(context.Background(), ScholarCrossref, "https://doi.org/10.1371/journal.pdig.0000514", 3)
	if err != nil {
		t.Fatal(err)
	}
	if len(hits) != 1 || hits[0].DOI != "10.1371/journal.pdig.0000514" || hits[0].SourceKey != "doi:10.1371/journal.pdig.0000514" || hits[0].ContentState != "abstract_only" {
		t.Fatalf("unexpected hits: %+v", hits)
	}
	if gotPath != "/works/10.1371%2Fjournal.pdig.0000514" {
		t.Fatalf("expected exact /works/{doi} lookup, got path %q", gotPath)
	}
}

func TestScholarUnknownDOIExplicitNotFound(t *testing.T) {
	s := newMockScholar(func(r *http.Request) (int, string) {
		return http.StatusNotFound, `Resource not found.`
	})
	_, err := s.Search(context.Background(), ScholarCrossref, "10.9999/nonexistent.12345678", 3)
	var se *apierrors.ServiceError
	if !errors.As(err, &se) || se.Code != apierrors.NotFound {
		t.Fatalf("expected explicit NotFound error, got %v", err)
	}
}

func TestScholarExactArxivRouting(t *testing.T) {
	var gotQuery string
	s := newMockScholar(func(r *http.Request) (int, string) {
		gotQuery = r.URL.RawQuery
		return http.StatusOK, `<?xml version="1.0"?><feed xmlns="http://www.w3.org/2005/Atom"><entry><id>http://arxiv.org/abs/2504.16054v2</id><title> A   Paper </title><summary>Abstract.</summary><published>2025-04-16T00:00:00Z</published><author><name>Alice Author</name></author></entry></feed>`
	})
	hits, err := s.Search(context.Background(), ScholarCrossref, "2504.16054v2", 3)
	if err != nil {
		t.Fatal(err)
	}
	if len(hits) != 1 || hits[0].ArxivID != "2504.16054v2" || hits[0].SourceKey != "arxiv:2504.16054" {
		t.Fatalf("unexpected hits: %+v", hits)
	}
	if gotQuery != "id_list=2504.16054v2" {
		t.Fatalf("expected id_list exact-version lookup, got query %q", gotQuery)
	}
}

func TestScholarExactArxivVersionMismatch(t *testing.T) {
	s := newMockScholar(func(r *http.Request) (int, string) {
		return http.StatusOK, `<?xml version="1.0"?><feed xmlns="http://www.w3.org/2005/Atom"><entry><id>http://arxiv.org/abs/2504.16054v1</id><title>A Paper</title><summary>Abstract.</summary><published>2025-04-16T00:00:00Z</published><author><name>Alice Author</name></author></entry></feed>`
	})
	_, err := s.Search(context.Background(), ScholarCrossref, "2504.16054v2", 3)
	var se *apierrors.ServiceError
	if !errors.As(err, &se) || se.Code != apierrors.NotFound {
		t.Fatalf("expected version mismatch NotFound, got %v", err)
	}
}

func TestScholarKeywordQueryUsesProviderSearch(t *testing.T) {
	var gotURL string
	s := newMockScholar(func(r *http.Request) (int, string) {
		gotURL = r.URL.String()
		return http.StatusOK, `{"message":{"items":[]}}`
	})
	if _, err := s.Search(context.Background(), ScholarCrossref, "attention mechanism transformer", 3); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(gotURL, "/works?") || !strings.Contains(gotURL, "query=attention+mechanism+transformer") {
		t.Fatalf("expected provider keyword search, got %q", gotURL)
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

// TestScholarExactIDLive verifies one exact identifier against each official
// registry (no full text, no paid access). It runs only with an explicit opt-in.
func TestScholarExactIDLive(t *testing.T) {
	if os.Getenv("ASTROCYTE_PAPER_SEARCH_LIVE") != "1" {
		t.Skip("live public metadata lookup not requested")
	}
	s := NewScholar()
	ctx := context.Background()

	t.Run("crossref doi", func(t *testing.T) {
		hits, err := s.Search(ctx, ScholarCrossref, "10.1371/journal.pdig.0000514", 3)
		if err != nil {
			t.Fatalf("live DOI lookup failed: %v", err)
		}
		if len(hits) != 1 || hits[0].DOI != "10.1371/journal.pdig.0000514" || hits[0].Title == "" {
			t.Fatalf("live DOI lookup returned unexpected hits: %+v", hits)
		}
		t.Logf("live crossref doi title=%q year=%d", hits[0].Title, hits[0].Year)
	})

	t.Run("arxiv id", func(t *testing.T) {
		hits, err := s.Search(ctx, ScholarCrossref, "2504.16054", 3)
		if err != nil {
			t.Fatalf("live arXiv id_list lookup failed: %v", err)
		}
		if len(hits) != 1 || hits[0].ArxivID == "" || hits[0].SourceKey != "arxiv:2504.16054" {
			t.Fatalf("live arXiv id_list returned unexpected hits: %+v", hits)
		}
		t.Logf("live arxiv id_list version=%q title=%q", hits[0].ArxivID, hits[0].Title)
	})
}
