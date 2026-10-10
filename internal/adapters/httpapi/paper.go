package httpapi

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/songconmaisaix31-design/Astrocyte/internal/attention/domain"
)

// paperAvailabilityWire is the transport availability vocabulary consumed by
// the paper search UI. It is derived from the domain content_state; the raw
// domain value is retained in Detail so no provider precision is dropped.
type paperAvailabilityWire struct {
	Status string  `json:"status"`
	Detail *string `json:"detail"`
}

type paperSearchHitWire struct {
	ID               string                `json:"id"`
	Title            string                `json:"title"`
	Authors          []string              `json:"authors"`
	Year             int                   `json:"year"`
	Venue            string                `json:"venue"`
	SourceType       string                `json:"source_type"`
	ArxivID          string                `json:"arxiv_id"`
	DOI              string                `json:"doi"`
	Locator          string                `json:"locator"`
	Abstract         string                `json:"abstract"`
	Availability     paperAvailabilityWire `json:"availability"`
	AlreadyImported  bool                  `json:"already_imported"`
	ImportMaterialID *string               `json:"import_material_id"`
}

type paperSearchResultWire struct {
	SchemaVersion int                  `json:"schema_version"`
	Query         string               `json:"query"`
	Items         []paperSearchHitWire `json:"items"`
	NextCursor    *string              `json:"next_cursor"`
	HasMore       bool                 `json:"has_more"`
	Warnings      []string             `json:"warnings"`
}

func availabilityStatus(contentState string) string {
	switch contentState {
	case "readable_fulltext":
		return "full_text"
	case "abstract_only":
		return "metadata_only"
	case "paywall", "restricted":
		return "restricted"
	default:
		return "unknown"
	}
}

func mapPaperSearchHit(h domain.PaperSearchHit) paperSearchHitWire {
	detail := h.ContentState
	if detail == "" {
		detail = "abstract_only"
	}
	return paperSearchHitWire{
		ID:           h.SourceKey,
		Title:        h.Title,
		Authors:      h.Authors,
		Year:         h.Year,
		Venue:        h.Venue,
		SourceType:   h.Provider,
		ArxivID:      h.ArxivID,
		DOI:          h.DOI,
		Locator:      h.Locator,
		Abstract:     h.Abstract,
		Availability: paperAvailabilityWire{Status: availabilityStatus(h.ContentState), Detail: &detail},
	}
}

// registerPaperSearch wires public academic metadata search (GET, human-only).
// It performs no import, no full-text extraction and no model call; a human
// selects a hit and imports it through the existing ImportMaterial path. The
// already_imported/import_material_id dedup columns are reserved by the wire
// contract but are not yet populated by the search domain (W1 remaining work).
func (h *handler) registerPaperSearch(mux *http.ServeMux) {
	s := h.services.PaperSearch
	if s == nil {
		mux.HandleFunc("GET /api/v1/papers/search", h.notImplemented("paper_search"))
		mux.HandleFunc("POST /api/v1/papers/import", h.notImplemented("paper_import"))
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
	// Batch import is a human-session write over selected search hits. The
	// domain mapping onto ImportMaterial/job dedup is W1's remaining work; until
	// it is assembled the route stays an explicit 501, never a silent success.
	mux.HandleFunc("POST /api/v1/papers/import", h.notImplemented("paper_import"))
}
