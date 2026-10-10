package domain

import "strings"

// PaperSearchQuery selects one credential-free public academic index and a
// bounded free-text query. A query never carries an actor, entitlement or
// import decision; it is metadata discovery only.
type PaperSearchQuery struct {
	Provider string `json:"provider"`
	Query    string `json:"query"`
	Limit    int    `json:"limit"`
}

// Validate bounds a search before any provider is contacted.
func (q PaperSearchQuery) Validate() error {
	q.Provider = strings.TrimSpace(q.Provider)
	q.Query = strings.TrimSpace(q.Query)
	if q.Provider == "" || q.Query == "" {
		return ErrInvalid
	}
	if len(q.Query) > 512 {
		return ErrInvalid
	}
	if q.Limit < 1 || q.Limit > 50 {
		return ErrInvalid
	}
	return nil
}

// PaperSearchHit is provider metadata, never full text or a material revision.
type PaperSearchHit struct {
	Provider     string   `json:"provider"`
	Title        string   `json:"title"`
	Authors      []string `json:"authors"`
	DOI          string   `json:"doi"`
	ArxivID      string   `json:"arxiv_id"`
	Year         int      `json:"year"`
	Venue        string   `json:"venue"`
	Abstract     string   `json:"abstract"`
	Locator      string   `json:"locator"`
	SourceKey    string   `json:"source_key"`
	PDFURLs      []string `json:"pdf_urls"`
	ContentState string   `json:"content_state"`
}
