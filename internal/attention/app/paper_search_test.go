package app_test

import (
	"context"
	"testing"

	"github.com/songconmaisaix31-design/Astrocyte/internal/adapters/importers"
	"github.com/songconmaisaix31-design/Astrocyte/internal/attention/app"
	"github.com/songconmaisaix31-design/Astrocyte/internal/attention/domain"
)

func TestSearchPapersAuthorizesAndNormalizes(t *testing.T) {
	s := app.NewAttentionService(nil, nil, nil, app.ServiceOptions{ScholarSearcher: importers.NewScholar()})
	hits, err := s.SearchPapers(context.Background(), app.Principal{ID: "h", Kind: "human"}, domain.PaperSearchQuery{Provider: "crossref", Query: "10.1371/journal.pdig.0000514", Limit: 10})
	if err != nil {
		t.Fatal(err)
	}
	if len(hits) != 1 || hits[0].DOI != "10.1371/journal.pdig.0000514" || hits[0].ContentState != "abstract_only" || hits[0].Title == "" {
		t.Fatalf("公开 DOI 查询缺少元数据: %+v", hits)
	}
	// Agent 权限检查在实际来源查询之前失败。
	if _, err := s.SearchPapers(context.Background(), app.Principal{ID: "a", Kind: "agent"}, domain.PaperSearchQuery{Provider: "crossref", Query: "x", Limit: 10}); err == nil {
		t.Fatal("agent search should be denied")
	}
}

func TestSearchPapersInvalidQuery(t *testing.T) {
	s := app.NewAttentionService(nil, nil, nil, app.ServiceOptions{ScholarSearcher: importers.NewScholar()})
	if _, err := s.SearchPapers(context.Background(), app.Principal{ID: "h", Kind: "human"}, domain.PaperSearchQuery{Provider: "crossref", Query: "", Limit: 10}); err == nil {
		t.Fatal("invalid query accepted")
	}
}
