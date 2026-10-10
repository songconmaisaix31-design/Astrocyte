package app

import (
	"context"
	"testing"

	"github.com/songconmaisaix31-design/Astrocyte/internal/attention/domain"
)

type fakePaperSearcher struct{ hits []domain.PaperSearchHit }

func (f fakePaperSearcher) SearchPapers(context.Context, domain.PaperSearchQuery) ([]domain.PaperSearchHit, error) {
	return f.hits, nil
}

func TestSearchPapersAuthorizesAndNormalizes(t *testing.T) {
	repo := newMemoryRepo()
	s := NewAttentionService(repo, nil, nil, ServiceOptions{ScholarSearcher: fakePaperSearcher{hits: []domain.PaperSearchHit{{Title: "X", SourceKey: "doi:10.1/x", ContentState: ""}}}})
	if _, err := s.SearchPapers(context.Background(), Principal{ID: "h", Kind: "human"}, domain.PaperSearchQuery{Provider: "crossref", Query: "x", Limit: 10}); err != nil {
		t.Fatal(err)
	}
	// Agent principal has no material access; search is human-only metadata.
	if _, err := s.SearchPapers(context.Background(), Principal{ID: "a", Kind: "agent"}, domain.PaperSearchQuery{Provider: "crossref", Query: "x", Limit: 10}); err == nil {
		t.Fatal("agent search should be denied")
	}
}

func TestSearchPapersInvalidQuery(t *testing.T) {
	repo := newMemoryRepo()
	s := NewAttentionService(repo, nil, nil, ServiceOptions{ScholarSearcher: fakePaperSearcher{}})
	if _, err := s.SearchPapers(context.Background(), Principal{ID: "h", Kind: "human"}, domain.PaperSearchQuery{Provider: "crossref", Query: "", Limit: 10}); err == nil {
		t.Fatal("invalid query accepted")
	}
}
