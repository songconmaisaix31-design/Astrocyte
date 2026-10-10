package httpapi

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/songconmaisaix31-design/Astrocyte/internal/attention/domain"
)

// paperSearchHitWire is the transport projection of one public academic search
// hit. It mirrors W1's canonical domain.PaperSearchHit (source_key/content_state/
// pdf_urls) rather than a reduced id/availability vocabulary, so the search UI
// can hand a hit straight back into ImportMaterial with its canonical source key.
type paperSearchHitWire struct {
	SourceKey    string   `json:"source_key"`
	Provider     string   `json:"provider"`
	Title        string   `json:"title"`
	Authors      []string `json:"authors"`
	Year         int      `json:"year"`
	Venue        string   `json:"venue"`
	ArxivID      string   `json:"arxiv_id"`
	DOI          string   `json:"doi"`
	Locator      string   `json:"locator"`
	Abstract     string   `json:"abstract"`
	ContentState string   `json:"content_state"`
	PDFURLs      []string `json:"pdf_urls"`
}

type paperSearchResultWire struct {
	SchemaVersion int                  `json:"schema_version"`
	Query         string               `json:"query"`
	Items         []paperSearchHitWire `json:"items"`
	NextCursor    *string              `json:"next_cursor"`
	HasMore       bool                 `json:"has_more"`
	Warnings      []string             `json:"warnings"`
}

// mapPaperSearchHit projects one domain hit to the wire. Content state is the
// raw domain vocabulary; PDF URLs are passed through verbatim for the human to
// select. No already_imported/dedup claim is fabricated: dedup resolves on the
// ImportMaterial job, not at search time.
func mapPaperSearchHit(h domain.PaperSearchHit) paperSearchHitWire {
	contentState := h.ContentState
	if contentState == "" {
		contentState = "abstract_only"
	}
	pdfs := h.PDFURLs
	if pdfs == nil {
		pdfs = []string{}
	}
	return paperSearchHitWire{
		SourceKey:    h.SourceKey,
		Provider:     h.Provider,
		Title:        h.Title,
		Authors:      h.Authors,
		Year:         h.Year,
		Venue:        h.Venue,
		ArxivID:      h.ArxivID,
		DOI:          h.DOI,
		Locator:      h.Locator,
		Abstract:     h.Abstract,
		ContentState: contentState,
		PDFURLs:      pdfs,
	}
}

// registerPaperSearch wires public academic metadata search (GET, human-only).
// It performs no import, no full-text extraction and no model call; a human
// selects a hit and imports it through the existing ImportMaterial path
// (adapter=paper_url|paper_pdf, source_key dedup, new revision). There is no
// separate batch-import route: the UI submits one ImportMaterial per selected
// hit, so a dead 501 batch entry is not retained.
func (h *handler) registerPaperSearch(mux *http.ServeMux) {
	s := h.services.PaperSearch
	if s == nil {
		mux.HandleFunc("GET /api/v1/papers/search", h.notImplemented("paper_search"))
		return
	}
	mux.HandleFunc("GET /api/v1/papers/search", func(w http.ResponseWriter, r *http.Request) {
		q := strings.TrimSpace(r.URL.Query().Get("q"))
		if q == "" {
			h.attentionResult(w, r, http.StatusOK, nil, invalid("query parameter q is required"))
			return
		}
		limit := 20
		if v := r.URL.Query().Get("limit"); v != "" {
			n, err := strconv.Atoi(v)
			if err != nil || n < 1 || n > 50 {
				h.attentionResult(w, r, http.StatusOK, nil, invalid("limit must be an integer between 1 and 50"))
				return
			}
			limit = n
		}
		provider := strings.TrimSpace(r.URL.Query().Get("provider"))
		if provider == "" {
			provider = "crossref"
		}
		query := domain.PaperSearchQuery{
			Provider: provider,
			Query:    q,
			Limit:    limit,
		}
		hits, err := s.SearchPapers(r.Context(), principal(r), query)
		if err != nil {
			h.attentionResult(w, r, http.StatusOK, nil, err)
			return
		}
		items := make([]paperSearchHitWire, 0, len(hits))
		for _, hit := range hits {
			items = append(items, mapPaperSearchHit(hit))
		}
		h.attentionResult(w, r, http.StatusOK, paperSearchResultWire{SchemaVersion: 1, Query: q, Items: items, Warnings: []string{}}, nil)
	})
}
