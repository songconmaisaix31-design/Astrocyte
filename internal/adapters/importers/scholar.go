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
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/songconmaisaix31-design/Astrocyte/internal/apierrors"
	"github.com/songconmaisaix31-design/Astrocyte/internal/attention/domain"
	"golang.org/x/net/html"
)

// ScholarProvider 选择公开学术元数据索引；查询结果仅包含元数据。
type ScholarProvider string

const (
	ScholarCrossref  ScholarProvider = "crossref"
	ScholarEuropePMC ScholarProvider = "europepmc"
	ScholarArxiv     ScholarProvider = "arxiv"
)

// ScholarHit 保存查询元数据，ContentState 固定为 abstract_only。
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

// Search 在官方 HTTPS 索引中执行一次有长度和数量限制的查询。
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

	// 完整 DOI 和 arXiv 标识符查询注册库；未知标识符返回明确的 404。
	if doi := parseDOIQuery(query); doi != "" {
		hit, err := s.crossrefWork(ctx, doi)
		if err != nil {
			return nil, err
		}
		hit.ContentState = "abstract_only"
		if hit.SourceKey == "" {
			hit.SourceKey = hit.Locator
		}
		return []ScholarHit{hit}, nil
	}
	if id, err := NormalizeArxivID(query); err == nil {
		hit, err := s.arxivIDLookup(ctx, id)
		if err != nil {
			return nil, err
		}
		hit.ContentState = "abstract_only"
		if hit.SourceKey == "" {
			hit.SourceKey = hit.Locator
		}
		return []ScholarHit{hit}, nil
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
	status, b, err := s.getStatus(ctx, locator, limit)
	if err != nil {
		return nil, err
	}
	if status != http.StatusOK {
		return nil, providerUnavailable(fmt.Errorf("provider HTTP %d", status))
	}
	return b, nil
}

// getStatus 保留实际 HTTP 状态，用于区分未知标识符和来源失败。
func (s *Scholar) getStatus(ctx context.Context, locator string, limit int64) (int, []byte, error) {
	u, err := url.Parse(locator)
	if err != nil || u.Scheme != "https" || u.User != nil || u.Port() != "" || !scholarHost(u.Hostname()) {
		return 0, nil, invalid("paper search requires an official public HTTPS endpoint")
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, locator, nil)
	if err != nil {
		return 0, nil, err
	}
	req.Header.Set("User-Agent", "Astrocyte/0.1 (paper metadata search; mailto:see-project)")
	resp, err := s.Client.Do(req)
	if err != nil {
		return 0, nil, providerUnavailable(err)
	}
	defer resp.Body.Close()
	b, err := io.ReadAll(io.LimitReader(resp.Body, limit+1))
	if err != nil {
		return resp.StatusCode, nil, providerUnavailable(err)
	}
	if int64(len(b)) > limit {
		return resp.StatusCode, nil, invalid("provider response exceeds search size limit")
	}
	return resp.StatusCode, b, nil
}

// providerUnavailable 保留网络或 HTTP 失败原因和所需操作。
func providerUnavailable(err error) error {
	return &apierrors.ServiceError{Code: apierrors.ProviderUnavailable, Message: "paper provider request failed: " + err.Error(), Retryable: true, RequiredAction: "retry_paper_request_or_check_network"}
}

func (s *Scholar) crossref(ctx context.Context, query string, limit int) ([]byte, error) {
	return s.get(ctx, scholarSearchTarget(ScholarCrossref, query, limit), 8<<20)
}

func scholarSearchTarget(provider ScholarProvider, query string, limit int) string {
	q := url.Values{}
	switch provider {
	case ScholarCrossref:
		q.Set("query", query)
		q.Set("rows", strconv.Itoa(limit))
		q.Set("select", "DOI,title,author,abstract,URL,published,container-title,type")
		return "https://api.crossref.org/works?" + q.Encode()
	case ScholarEuropePMC:
		q.Set("query", query)
		q.Set("resultType", "lite")
		q.Set("format", "json")
		q.Set("pageSize", strconv.Itoa(limit))
		return "https://www.ebi.ac.uk/europepmc/webservices/rest/search?" + q.Encode()
	case ScholarArxiv:
		q.Set("search_query", "all:"+query)
		q.Set("start", "0")
		q.Set("max_results", strconv.Itoa(limit))
		return "https://export.arxiv.org/api/query?" + q.Encode()
	default:
		panic("validated scholar provider required")
	}
}

func (s *Scholar) europepmc(ctx context.Context, query string, limit int) ([]byte, error) {
	return s.get(ctx, scholarSearchTarget(ScholarEuropePMC, query, limit), 8<<20)
}

func (s *Scholar) arxivSearch(ctx context.Context, query string, limit int) ([]byte, error) {
	return s.get(ctx, scholarSearchTarget(ScholarArxiv, query, limit), 8<<20)
}

func crossrefWorkTarget(doi string) string {
	return "https://api.crossref.org/works/" + url.PathEscape(doi)
}

func arxivIDTarget(id string) string {
	return "https://export.arxiv.org/api/query?" + url.Values{"id_list": {id}}.Encode()
}

// completeDOI 要求完整的 10. 注册前缀和非空后缀。
var completeDOI = regexp.MustCompile(`^10\.[0-9]{4,9}/[^\s]+$`)

// parseDOIQuery 识别 DOI、doi: 前缀和 doi.org URL；普通关键词返回空字符串。
func parseDOIQuery(query string) string {
	q := strings.TrimSpace(query)
	if q == "" {
		return ""
	}
	if strings.Contains(q, "://") {
		u, err := url.Parse(q)
		if err != nil || (u.Scheme != "http" && u.Scheme != "https") || u.User != nil {
			return ""
		}
		if u.Host != "doi.org" && u.Host != "www.doi.org" && u.Host != "dx.doi.org" && u.Host != "www.dx.doi.org" {
			return ""
		}
		q = strings.TrimPrefix(u.Path, "/")
		q = strings.TrimSuffix(q, "/")
	}
	d := normalizeDOIRaw(q)
	if !completeDOI.MatchString(d) {
		return ""
	}
	return d
}

// crossrefWork 使用官方 GET /works/{doi}，未知 DOI 返回 NotFound。
func (s *Scholar) crossrefWork(ctx context.Context, doi string) (ScholarHit, error) {
	u := crossrefWorkTarget(doi)
	status, body, err := s.getStatus(ctx, u, 8<<20)
	if err != nil {
		return ScholarHit{}, err
	}
	if status == http.StatusNotFound {
		return ScholarHit{}, notFoundDOI(doi)
	}
	if status != http.StatusOK {
		return ScholarHit{}, providerUnavailable(fmt.Errorf("Crossref work HTTP %d", status))
	}
	hit, err := parseCrossrefWork(body)
	if err != nil {
		return ScholarHit{}, err
	}
	return hit, nil
}

// arxivIDLookup 使用官方 id_list 查询，并严格核对请求版本。
func (s *Scholar) arxivIDLookup(ctx context.Context, id string) (ScholarHit, error) {
	u := arxivIDTarget(id)
	raw, err := s.get(ctx, u, 8<<20)
	if err != nil {
		return ScholarHit{}, err
	}
	hits, err := parseArxivSearch(raw)
	if err != nil {
		return ScholarHit{}, err
	}
	return selectArxivHit(id, hits)
}

func selectArxivHit(id string, hits []ScholarHit) (ScholarHit, error) {
	if len(hits) != 1 {
		return ScholarHit{}, arxivNotFound(id)
	}
	hit := hits[0]
	base := revisionSuffix.ReplaceAllString(id, "")
	if revisionSuffix.ReplaceAllString(hit.ArxivID, "") != base {
		return ScholarHit{}, arxivNotFound(id)
	}
	if revisionSuffix.MatchString(id) && hit.ArxivID != id {
		return ScholarHit{}, arxivNotFound(id)
	}
	return hit, nil
}

func notFoundDOI(doi string) error {
	return &apierrors.ServiceError{Code: apierrors.NotFound, Message: "DOI " + doi + " was not found in Crossref", RequiredAction: "check_doi_and_retry"}
}

func arxivNotFound(id string) error {
	return &apierrors.ServiceError{Code: apierrors.NotFound, Message: "arXiv identifier " + id + " was not found", RequiredAction: "check_arxiv_id_and_retry"}
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
			Items []crossrefWorkMessage `json:"items"`
		} `json:"message"`
	}
	if err := json.Unmarshal(raw, &payload); err != nil {
		return nil, crossrefParseError(err)
	}
	hits := make([]ScholarHit, 0, len(payload.Message.Items))
	for _, item := range payload.Message.Items {
		hit, err := crossrefHit(item)
		if err != nil {
			return nil, err
		}
		hits = append(hits, hit)
	}
	return hits, nil
}

// parseCrossrefWork 解析单篇查询返回的 message 对象。
func parseCrossrefWork(raw []byte) (ScholarHit, error) {
	var payload struct {
		Message crossrefWorkMessage `json:"message"`
	}
	if err := json.Unmarshal(raw, &payload); err != nil {
		return ScholarHit{}, crossrefParseError(err)
	}
	if normalizeDOIRaw(payload.Message.DOI) == "" {
		return ScholarHit{}, &apierrors.ServiceError{Code: apierrors.ProviderUnavailable, Message: "Crossref work response lacks a DOI", Retryable: true, RequiredAction: "check_paper_search_provider_response"}
	}
	return crossrefHit(payload.Message)
}

func crossrefParseError(err error) error {
	return &apierrors.ServiceError{Code: apierrors.ProviderUnavailable, Message: "Crossref search response could not be parsed: " + err.Error(), Retryable: true, RequiredAction: "check_paper_search_provider_response"}
}

// crossrefWorkMessage 是关键词查询与单篇查询共用的记录格式。
type crossrefWorkMessage struct {
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
}

// 通过成熟 HTML tokenizer 提取 JATS 摘要的文字和字符引用。
func plainAbstract(raw string) (string, error) {
	z := html.NewTokenizer(strings.NewReader(raw))
	var text strings.Builder
	for {
		switch z.Next() {
		case html.ErrorToken:
			if err := z.Err(); err != io.EOF {
				return "", err
			}
			return strings.Join(strings.Fields(text.String()), " "), nil
		case html.TextToken:
			text.Write(z.Text())
		case html.StartTagToken, html.EndTagToken, html.SelfClosingTagToken:
			text.WriteByte(' ')
		}
	}
}

func crossrefHit(msg crossrefWorkMessage) (ScholarHit, error) {
	abstract, err := plainAbstract(msg.Abstract)
	if err != nil {
		return ScholarHit{}, err
	}
	doi := normalizeDOIRaw(msg.DOI)
	authors := []string{}
	for _, a := range msg.Author {
		name := strings.TrimSpace(a.Given + " " + a.Family)
		if name != "" {
			authors = append(authors, name)
		}
	}
	title := strings.TrimSpace(strings.Join(msg.Title, " "))
	locator := msg.URL
	if locator == "" && doi != "" {
		locator = "https://doi.org/" + doi
	}
	year := 0
	if len(msg.Published.DateParts) > 0 && len(msg.Published.DateParts[0]) > 0 {
		year = msg.Published.DateParts[0][0]
	}
	venue := strings.TrimSpace(strings.Join(msg.ContainerTitle, " "))
	return ScholarHit{Provider: string(ScholarCrossref), Title: title, Authors: authors, DOI: doi, Year: year, Venue: venue, Abstract: abstract, Locator: locator, SourceKey: doiKey(doi, locator)}, nil
}

// europePMCInt 接收 Europe PMC lite 返回的整数或字符串年份。
type europePMCInt int

func (f *europePMCInt) UnmarshalJSON(b []byte) error {
	var value any
	decoder := json.NewDecoder(strings.NewReader(string(b)))
	decoder.UseNumber()
	if err := decoder.Decode(&value); err != nil {
		return err
	}
	if value == nil || value == "" {
		*f = 0
		return nil
	}
	var s string
	switch v := value.(type) {
	case string:
		s = v
	case json.Number:
		s = v.String()
	default:
		return fmt.Errorf("Europe PMC year requires a JSON integer or string")
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
			return nil, fmt.Errorf("arXiv search entry identifier: %w", err)
		}
		authors := []string{}
		for _, a := range e.Authors {
			if strings.TrimSpace(a.Name) != "" {
				authors = append(authors, strings.TrimSpace(a.Name))
			}
		}
		year := 0
		if e.Published != "" {
			published, err := time.Parse(time.RFC3339, e.Published)
			if err != nil {
				return nil, fmt.Errorf("arXiv search entry publication: %w", err)
			}
			year = published.Year()
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

// SortHits 按来源标识和标题稳定排序。
func SortHits(hits []ScholarHit) {
	sort.SliceStable(hits, func(i, j int) bool {
		if hits[i].SourceKey != hits[j].SourceKey {
			return hits[i].SourceKey < hits[j].SourceKey
		}
		return hits[i].Title < hits[j].Title
	})
}

// SearchPapers 实现应用 PaperSearcher 接口，返回公开元数据。
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
