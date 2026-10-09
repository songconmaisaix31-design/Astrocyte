package importers

import (
	"bytes"
	"context"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"
)

var arxivID = regexp.MustCompile(`^(?:[0-9]{4}\.[0-9]{4,5}|[a-z][a-z0-9.-]*(?:\.[A-Z]{2})?/[0-9]{7})(?:v[1-9][0-9]*)?$`)
var revisionSuffix = regexp.MustCompile(`v[1-9][0-9]*$`)

func NormalizeArxivID(locator string) (string, error) {
	id := strings.TrimSpace(locator)
	if strings.HasPrefix(id, "arXiv:") {
		id = strings.TrimPrefix(id, "arXiv:")
	}
	if strings.Contains(id, "://") {
		u, err := url.Parse(id)
		if err != nil || (u.Scheme != "http" && u.Scheme != "https") || u.User != nil || (u.Host != "arxiv.org" && u.Host != "www.arxiv.org" && u.Host != "export.arxiv.org") {
			return "", errors.New("arXiv source must use an official URL or ID")
		}
		id = ""
		for _, prefix := range []string{"/abs/", "/pdf/", "/html/"} {
			if strings.HasPrefix(u.Path, prefix) {
				id = strings.TrimPrefix(u.Path, prefix)
				break
			}
		}
		id = strings.TrimSuffix(id, ".pdf")
	}
	if !arxivID.MatchString(id) {
		return "", errors.New("invalid arXiv ID")
	}
	return id, nil
}

// Paper preserves the full official PDF and exact version returned by Atom.
// An abstract is metadata, never reported as the paper's full text.
type Paper struct {
	ID, SourceKey, Title, Abstract, PDFURL string
	PDF, Metadata                          []byte
}

type Arxiv struct{ Client *http.Client }

func NewArxiv() *Arxiv {
	return &Arxiv{Client: &http.Client{Timeout: 45 * time.Second, CheckRedirect: officialArxivRedirect}}
}

func officialArxivRedirect(req *http.Request, via []*http.Request) error {
	if len(via) >= 5 {
		return errors.New("too many arXiv redirects")
	}
	if req.URL.Scheme != "https" || !officialArxivHost(req.URL.Host) {
		return errors.New("arXiv redirect outside official HTTPS hosts")
	}
	return nil
}

func officialArxivHost(host string) bool {
	return host == "arxiv.org" || host == "export.arxiv.org" || host == "www.arxiv.org"
}

func (a *Arxiv) Read(ctx context.Context, locator string) (Paper, error) {
	id, err := NormalizeArxivID(locator)
	if err != nil {
		return Paper{}, err
	}
	metadata, err := a.get(ctx, "https://export.arxiv.org/api/query?id_list="+url.QueryEscape(id), 4<<20)
	if err != nil {
		return Paper{}, err
	}
	var feed struct {
		Entries []struct {
			ID      string `xml:"id"`
			Title   string `xml:"title"`
			Summary string `xml:"summary"`
		} `xml:"entry"`
	}
	if err := xml.Unmarshal(metadata, &feed); err != nil {
		return Paper{}, fmt.Errorf("arXiv Atom: %w", err)
	}
	if len(feed.Entries) != 1 {
		return Paper{}, errors.New("arXiv did not return exactly one paper")
	}
	e := feed.Entries[0]
	fixed, err := NormalizeArxivID(e.ID)
	if err != nil || !revisionSuffix.MatchString(fixed) || revisionSuffix.ReplaceAllString(fixed, "") != revisionSuffix.ReplaceAllString(id, "") {
		return Paper{}, errors.New("arXiv response lacks matching fixed revision")
	}
	if revisionSuffix.MatchString(id) && fixed != id {
		return Paper{}, errors.New("arXiv returned a different requested revision")
	}
	pdfURL := "https://arxiv.org/pdf/" + fixed
	pdf, err := a.get(ctx, pdfURL, 64<<20)
	if err != nil {
		return Paper{}, err
	}
	if !bytes.HasPrefix(bytes.TrimSpace(pdf), []byte("%PDF-")) {
		return Paper{}, errors.New("arXiv original response is not a PDF")
	}
	return Paper{ID: fixed, SourceKey: "arxiv:" + revisionSuffix.ReplaceAllString(fixed, ""), Title: strings.Join(strings.Fields(e.Title), " "), Abstract: strings.TrimSpace(e.Summary), PDFURL: pdfURL, PDF: pdf, Metadata: metadata}, nil
}

func (a *Arxiv) get(ctx context.Context, locator string, limit int64) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, locator, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "Astrocyte/0.1 (single-paper import)")
	resp, err := a.Client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("arXiv unavailable: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("arXiv HTTP %d", resp.StatusCode)
	}
	b, err := io.ReadAll(io.LimitReader(resp.Body, limit+1))
	if err != nil {
		return nil, err
	}
	if int64(len(b)) > limit {
		return nil, errors.New("arXiv response exceeds import size limit")
	}
	return b, nil
}
