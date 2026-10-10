package app

import (
	"context"

	"github.com/songconmaisaix31-design/Astrocyte/internal/apierrors"
	"github.com/songconmaisaix31-design/Astrocyte/internal/attention/domain"
)

// PaperSearcher queries one credential-free public academic metadata index.
// It is read-only metadata discovery and never imports, navigates a browser,
// or establishes full text. The importers.Scholar adapter implements this port.
type PaperSearcher interface {
	SearchPapers(context.Context, domain.PaperSearchQuery) ([]domain.PaperSearchHit, error)
}

// PaperSearchService exposes bounded academic search to the transport layer.
// It is deliberately separate from AttentionService so W0 can wire the HTTP
// route without widening any write/approval boundary.
type PaperSearchService interface {
	SearchPapers(context.Context, Principal, domain.PaperSearchQuery) ([]domain.PaperSearchHit, error)
}

func (s *Service) SearchPapers(ctx context.Context, p Principal, q domain.PaperSearchQuery) ([]domain.PaperSearchHit, error) {
	if err := authorize(p, false); err != nil {
		return nil, err
	}
	if err := q.Validate(); err != nil {
		return nil, mapError(domain.ErrInvalid, "")
	}
	if s.options.ScholarSearcher == nil {
		return nil, serviceError(apierrors.ProviderUnavailable, "Paper search is not configured", "configure_paper_search_provider")
	}
	hits, err := s.options.ScholarSearcher.SearchPapers(ctx, q)
	if err != nil {
		return nil, mapError(err, "")
	}
	for i := range hits {
		if hits[i].ContentState == "" {
			hits[i].ContentState = "abstract_only"
		}
	}
	return hits, nil
}
