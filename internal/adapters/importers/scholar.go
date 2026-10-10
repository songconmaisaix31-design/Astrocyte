package importers

import (
	"context"
	"encoding/json"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/songconmaisaix31-design/Astrocyte/internal/apierrors"
	"github.com/songconmaisaix31-design/Astrocyte/internal/attention/domain"
)

// ScholarProvider selects one official, credential-free public academic
// metadata index. A search hit is metadata only; it never carries full text,
// an entitlement or an import decision.
type ScholarProvider string

const (
	ScholarCrossref  ScholarProvider = "crossref"
	ScholarEuropePMC ScholarProvider = "europepmc"
	ScholarArxiv     ScholarProvider = "arxiv"
)

// ScholarHit is a search metadata record, not a material revision. ContentState
// is always abstract_only because providers are metadata indexes.
type ScholarHit struct {
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

type Scholar struct {
	Client *http.Client
}

func NewScholar() *Scholar {
	return &Scholar{Client: &http.Client{Timeout: 30 * time.Second, CheckRedirect: scholarRedirect}}
}

func scholarRedirect(req *http.Request, via []*http.Request) error {
	if len(via) >= 5 {
		return errors.New("too many scholar redirects")
	}
	if req.URL.Scheme != "https" || !scholarHost(req.URL.Host) {
		return errors.New("scholar redirect outside official HTTPS hosts")
	}
	return nil
}

func scholarHost(host string) bool {
	switch host {
	case "api.crossref.org", "www.ebi.ac.uk", "export.arxiv.org", "arxiv.org", "api.openalex.org":
		return true
	}
	return false
}

// Search performs one bounded query against a single provider. The query must be
// nonempty and stay within a hard character limit; results are capped at the
// caller's limit. Providers are contacted over official public HTTPS only, with
// no credentials, cookies or private network addresses.
func (s *Scholar) Search(ctx context.Context, provider ScholarProvider, query string, limit int) ([]ScholarHit, error) {
	query = strings.TrimSpace(query)
	if query == "" {
		return nil, invalid("paper search query must not be empty")
	}
	if len(query) > 512 {
		return nil, invalid("paper search query exceeds 512 characters")
	}
	if limit < 1 || limit > 50 {
		return nil, invalid("paper search limit must be between 1 and 50")
	}
	switch provider {
	case ScholarCrossref, ScholarEuropePMC, ScholarArxiv:
	default:
		return nil, invalid("unsupported paper search provider")
	}
	var (
		raw []byte
		err error
	)
	switch provider {
	case ScholarCrossref:
		raw, err = s.crossref(ctx, query, limit)
	case ScholarEuropePMC:
		raw, err = s.europepmc(ctx, query, limit)
	case ScholarArxiv:
		raw, err = s.arxivSearch(ctx, query, limit)
	}
	if err != nil {
		return nil, err
	}
	hits, err := s.parse(provider, raw)
	if err != nil {
		return nil, err
	}
	if len(hits) > limit {
		hits = hits[:limit]
	}
	for i := range hits {
		hits[i].ContentState = "abstract_only"
		if hits[i].SourceKey == "" {
			hits[i].SourceKey = hits[i].Locator
		}
	}
	return hits, nil
}

func (s *Scholar) get(ctx context.Context, locator string, limit int64) ([]byte, error) {
	u, err := url.Parse(locator)
	if err != nil || u.Scheme != "https" || u.User != nil || u.Port() != "" || !scholarHost(u.Hostname()) {
		return nil, invalid("paper search requires an official public HTTPS endpoint")
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, locator, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "Astrocyte/0.1 (paper metadata search; mailto:see-project)")
	resp, err := s.Client.Do(req)
	if err != nil {
		return nil, providerUnavailable(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, providerUnavailable(fmt.Errorf("provider HTTP %d", resp.StatusCode))
	}
	b, err := io.ReadAll(io.LimitReader(resp.Body, limit+1))
	if err != nil {
		return nil, providerUnavailable(err)
	}
	if int64(len(b)) > limit {
		return nil, invalid("provider response exceeds search size limit")
	}
	return b, nil
}

// providerUnavailable marks a network/HTTP failure as retryable with a clear
// required action, preserving the cause for logs instead of masking it.
func providerUnavailable(err error) error {
	return &apierrors.ServiceError{Code: apierrors.ProviderUnavailable, Message: "paper provider request failed: " + err.Error(), Retryable: true, RequiredAction: "retry_paper_request_or_check_network"}
}

func (s *Scholar) crossref(ctx context.Context, query string, limit int) ([]byte, error) {
	q := url.Values{}
	q.Set("query", query)
	q.Set("rows", strconv.Itoa(limit))
	q.Set("select", "DOI,title,author,abstract,URL,published,container-title,type")
	u := "https://api.crossref.org/works?" + q.Encode()
	return s.get(ctx, u, 8<<20)
}

func (s *Scholar) europepmc(ctx context.Context, query string, limit int) ([]byte, error) {
	q := url.Values{}
	q.Set("query", query)
	q.Set("resultType", "lite")
	q.Set("format", "json")
	q.Set("pageSize", strconv.Itoa(limit))
	u := "https://www.ebi.ac.uk/europepmc/webservices/rest/search?" + q.Encode()
	return s.get(ctx, u, 8<<20)
}

func (s *Scholar) arxivSearch(ctx context.Context, query string, limit int) ([]byte, error) {
	q := url.Values{}
	q.Set("search_query", "all:"+query)
	q.Set("start", "0")
	q.Set("max_results", strconv.Itoa(limit))
	u := "https://export.arxiv.org/api/query?" + q.Encode()
	return s.get(ctx, u, 8<<20)
}

func (s *Scholar) parse(provider ScholarProvider, raw []byte) ([]ScholarHit, error) {
	switch provider {
	case ScholarCrossref:
		return parseCrossref(raw)
	case ScholarEuropePMC:
		return parseEuropePMC(raw)
	case ScholarArxiv:
		return parseArxivSearch(raw)
	}
	return nil, invalid("unsupported paper search provider")
}

func parseCrossref(raw []byte) ([]ScholarHit, error) {
	var payload struct {
		Message struct {
			Items []struct {
				DOI    string   `json:"DOI"`
				Title  []string `json:"title"`
				Author []struct {
					Family string `json:"family"`
					Given  string `json:"given"`
				} `json:"author"`
				Abstract  string `json:"abstract"`
				URL       string `json:"URL"`
				Published struct {
					DateParts [][]int `json:"date-parts"`
				} `json:"published"`
				ContainerTitle []string `json:"container-title"`
			} `json:"items"`
		} `json:"message"`
	}
	if err := json.Unmarshal(raw, &payload); err != nil {
		return nil, &apierrors.ServiceError{Code: apierrors.ProviderUnavailable, Message: "Crossref search response could not be parsed: " + err.Error(), Retryable: true, RequiredAction: "check_paper_search_provider_response"}
	}
	hits := make([]ScholarHit, 0, len(payload.Message.Items))
	for _, item := range payload.Message.Items {
		doi := normalizeDOIRaw(item.DOI)
		authors := []string{}
		for _, a := range item.Author {
			name := strings.TrimSpace(a.Given + " " + a.Family)
			if name != "" {
				authors = append(authors, name)
			}
		}
		title := strings.TrimSpace(strings.Join(item.Title, " "))
		locator := item.URL
		if locator == "" && doi != "" {
			locator = "https://doi.org/" + doi
		}
		year := 0
		if len(item.Published.DateParts) > 0 && len(item.Published.DateParts[0]) > 0 {
			year = item.Published.DateParts[0][0]
		}
		venue := strings.TrimSpace(strings.Join(item.ContainerTitle, " "))
		hits = append(hits, ScholarHit{Provider: string(ScholarCrossref), Title: title, Authors: authors, DOI: doi, Year: year, Venue: venue, Abstract: strings.TrimSpace(item.Abstract), Locator: locator, SourceKey: doiKey(doi, locator)})
	}
	return hits, nil
}

// europePMCInt accepts the mixed number/string year representation returned by
// the lite result type across query styles.
type europePMCInt int

func (f *europePMCInt) UnmarshalJSON(b []byte) error {
	s := strings.Trim(string(b), `"`)
	if s == "" || s == "null" {
		*f = 0
		return nil
	}
	n, err := strconv.Atoi(s)
	if err != nil {
		return err
	}
	*f = europePMCInt(n)
	return nil
}

func parseEuropePMC(raw []byte) ([]ScholarHit, error) {
	var payload struct {
		ResultList struct {
			Result []struct {
				ID           string       `json:"id"`
				Title        string       `json:"title"`
				AuthorString string       `json:"authorString"`
				AbstractText string       `json:"abstractText"`
				DOI          string       `json:"doi"`
				PubYear      europePMCInt `json:"pubYear"`
				JournalTitle string       `json:"journalTitle"`
				PMID         string       `json:"pmid"`
				PMCID        string       `json:"pmcid"`
				FullTextURLs struct {
					URLs []struct {
						URL string `json:"url"`
					} `json:"fullTextUrl"`
				} `json:"fullTextUrlList"`
			} `json:"result"`
		} `json:"resultList"`
	}
	if err := json.Unmarshal(raw, &payload); err != nil {
		return nil, &apierrors.ServiceError{Code: apierrors.ProviderUnavailable, Message: "Europe PMC search response could not be parsed: " + err.Error(), Retryable: true, RequiredAction: "check_paper_search_provider_response"}
	}
	hits := make([]ScholarHit, 0, len(payload.ResultList.Result))
	for _, item := range payload.ResultList.Result {
		doi := normalizeDOIRaw(item.DOI)
		authors := splitAuthors(item.AuthorString)
		locator := ""
		switch {
		case item.PMCID != "":
			locator = "https://pmc.ncbi.nlm.nih.gov/articles/" + item.PMCID + "/"
		case item.PMID != "":
			locator = "https://pubmed.ncbi.nlm.nih.gov/" + item.PMID + "/"
		case doi != "":
			locator = "https://doi.org/" + doi
		}
		pdfs := []string{}
		for _, u := range item.FullTextURLs.URLs {
			if u.URL != "" {
				pdfs = append(pdfs, u.URL)
			}
		}
		hits = append(hits, ScholarHit{Provider: string(ScholarEuropePMC), Title: strings.TrimSpace(item.Title), Authors: authors, DOI: doi, Year: int(item.PubYear), Venue: strings.TrimSpace(item.JournalTitle), Abstract: strings.TrimSpace(item.AbstractText), Locator: locator, SourceKey: doiKey(doi, locator), PDFURLs: pdfs})
	}
	return hits, nil
}

func parseArxivSearch(raw []byte) ([]ScholarHit, error) {
	var feed struct {
		Entries []struct {
			ID        string `xml:"id"`
			Title     string `xml:"title"`
			Summary   string `xml:"summary"`
			Published string `xml:"published"`
			Authors   []struct {
				Name string `xml:"name"`
			} `xml:"author"`
		} `xml:"entry"`
	}
	if err := xml.Unmarshal(raw, &feed); err != nil {
		return nil, &apierrors.ServiceError{Code: apierrors.ProviderUnavailable, Message: "arXiv search response could not be parsed: " + err.Error(), Retryable: true, RequiredAction: "check_paper_search_provider_response"}
	}
	hits := make([]ScholarHit, 0, len(feed.Entries))
	for _, e := range feed.Entries {
		id, err := NormalizeArxivID(e.ID)
		if err != nil {
			continue
		}
		authors := []string{}
		for _, a := range e.Authors {
			if strings.TrimSpace(a.Name) != "" {
				authors = append(authors, strings.TrimSpace(a.Name))
			}
		}
		year := 0
		if len(e.Published) >= 4 {
			year, _ = strconv.Atoi(e.Published[:4])
		}
		base := revisionSuffix.ReplaceAllString(id, "")
		locator := "https://arxiv.org/abs/" + id
		hits = append(hits, ScholarHit{Provider: string(ScholarArxiv), Title: strings.Join(strings.Fields(e.Title), " "), Authors: authors, ArxivID: id, Year: year, Abstract: strings.TrimSpace(e.Summary), Locator: locator, SourceKey: "arxiv:" + base, PDFURLs: []string{"https://arxiv.org/pdf/" + id}})
	}
	return hits, nil
}

func normalizeDOIRaw(raw string) string {
	if raw == "" {
		return ""
	}
	d := strings.TrimSpace(raw)
	d = strings.TrimPrefix(d, "doi:")
	d = strings.TrimPrefix(d, "https://doi.org/")
	d = strings.TrimPrefix(d, "http://dx.doi.org/")
	d = strings.TrimPrefix(d, "https://dx.doi.org/")
	if strings.HasPrefix(d, "10.") {
		return strings.ToLower(d)
	}
	return ""
}

func doiKey(doi, locator string) string {
	if doi != "" {
		return "doi:" + doi
	}
	return locator
}

func splitAuthors(s string) []string {
	sep := ";"
	if !strings.Contains(s, ";") {
		sep = ","
	}
	parts := strings.Split(s, sep)
	out := []string{}
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			out = append(out, p)
		}
	}
	return out
}

// SortHits stabilizes provider result order by a stable key so identical
// queries produce deterministic output for tests and dedupe.
func SortHits(hits []ScholarHit) {
	sort.SliceStable(hits, func(i, j int) bool {
		if hits[i].SourceKey != hits[j].SourceKey {
			return hits[i].SourceKey < hits[j].SourceKey
		}
		return hits[i].Title < hits[j].Title
	})
}

// SearchPapers implements the application PaperSearcher port, mapping the
// domain query to one bounded provider call. Hits remain metadata only.
func (s *Scholar) SearchPapers(ctx context.Context, q domain.PaperSearchQuery) ([]domain.PaperSearchHit, error) {
	if err := q.Validate(); err != nil {
		return nil, invalid("invalid paper search query")
	}
	hits, err := s.Search(ctx, ScholarProvider(q.Provider), q.Query, q.Limit)
	if err != nil {
		return nil, err
	}
	out := make([]domain.PaperSearchHit, 0, len(hits))
	for _, h := range hits {
		out = append(out, domain.PaperSearchHit{Provider: h.Provider, Title: h.Title, Authors: h.Authors, DOI: h.DOI, ArxivID: h.ArxivID, Year: h.Year, Venue: h.Venue, Abstract: h.Abstract, Locator: h.Locator, SourceKey: h.SourceKey, PDFURLs: h.PDFURLs, ContentState: h.ContentState})
	}
	return out, nil
}
