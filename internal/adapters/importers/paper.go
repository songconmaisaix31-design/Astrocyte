package importers

import (
	"bytes"
	"context"
	_ "embed"
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/songconmaisaix31-design/Astrocyte/internal/apierrors"
	"github.com/songconmaisaix31-design/Astrocyte/internal/attention/app"
)

//go:embed paper-dom.mjs
var paperDOMModule []byte

//go:embed paper-dom-bridge.mjs
var paperDOMBridge []byte

// PaperSnapshot is adapter-local observation, not authority to import or process.
// Text is empty unless body evidence establishes readable full-paper structure.
type PaperSnapshot struct {
	SchemaVersion   int      `json:"schema_version"`
	SourceURL       string   `json:"source_url"`
	HostFamily      string   `json:"host_family"`
	SourceKey       string   `json:"source_key"`
	Title           string   `json:"title"`
	Abstract        string   `json:"abstract"`
	Authors         []string `json:"authors"`
	DOI             string   `json:"doi"`
	ArxivID         string   `json:"arxiv_id"`
	ObservedVersion string   `json:"observed_version"`
	PDFURLs         []string `json:"pdf_urls"`
	ContentState    string   `json:"content_state"`
	Text            string   `json:"text"`
	Warning         string   `json:"warning"`
	Truncated       bool     `json:"truncated"`
	Provenance      struct {
		Processor, Version, Mode, Source string
	} `json:"provenance"`
}

// InspectPaperHTML parses only supplied HTML bytes with existing pinned
// summarize DOM dependencies, without fetching URLs or invoking a model.
// It supports rendered current-page observations and explicit public snapshots
// without selecting either pending product path or freezing a transport.
func (e *SummarizeExtractor) InspectPaperHTML(ctx context.Context, locator string, html []byte) (PaperSnapshot, error) {
	var snapshot PaperSnapshot
	u, err := url.Parse(locator)
	if err != nil || u.Scheme != "https" || u.Hostname() == "" || u.User != nil || u.Port() != "" {
		return snapshot, invalid("paper observation requires a public HTTPS URL without credentials")
	}
	if len(html) == 0 || len(html) > 16<<20 {
		return snapshot, invalid("paper HTML must contain from 1 byte to 16MiB; no truncation")
	}
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	home, err := os.MkdirTemp("", "astrocyte-paper-dom-")
	if err != nil {
		return snapshot, err
	}
	defer os.RemoveAll(home)
	for name, data := range map[string][]byte{"paper-dom.mjs": paperDOMModule, "paper-dom-bridge.mjs": paperDOMBridge} {
		if err = os.WriteFile(filepath.Join(home, name), data, 0o600); err != nil {
			return snapshot, err
		}
	}
	cmd := exec.CommandContext(ctx, e.NodeExecutable, filepath.Join(home, "paper-dom-bridge.mjs"), e.CLIPath, locator)
	cmd.Env = e.isolatedEnv(home, false)
	cmd.Dir = home
	cmd.Stdin = bytes.NewReader(html)
	cmd.WaitDelay = 3 * time.Second
	var out, stderr cappedOutput
	cmd.Stdout, cmd.Stderr = &out, &stderr
	if err = cmd.Run(); err != nil {
		if ctx.Err() != nil {
			return snapshot, ctx.Err()
		}
		return snapshot, dependencyError(fmt.Sprintf("paper DOM extraction failed (%v): %s", err, stderr.String()))
	}
	if err = json.Unmarshal(out.Bytes(), &snapshot); err != nil {
		return snapshot, fmt.Errorf("paper DOM output: %w", err)
	}
	if snapshot.SourceURL != locator || snapshot.Truncated || (snapshot.ContentState != "readable_fulltext" && strings.TrimSpace(snapshot.Text) != "") {
		return PaperSnapshot{}, invalid("paper DOM result lacks matching untruncated provenance")
	}
	return snapshot, nil
}

// paperSource maps a DOM observation to an importable source without ever
// presenting metadata or a PDF link as full text. A restricted page is refused;
// a metadata-only page imports only clearly-marked metadata.
func paperSource(snapshot PaperSnapshot, html []byte) (app.ImportedSource, error) {
	var source app.ImportedSource
	source.SourceKey = snapshot.SourceKey
	source.SourceLocator = snapshot.SourceURL
	source.Kind = "paper"
	source.Title = snapshot.Title
	source.Summary = snapshot.Abstract
	source.Attachments = []app.SourceAttachment{{Name: "source.html", MediaType: "text/html", Data: html, SourceLocator: snapshot.SourceURL}}
	source.SourceSpans = []string{"whole HTML document: " + snapshot.SourceURL}
	version := "summarize-readability-approach " + snapshot.Provenance.Version
	switch snapshot.ContentState {
	case "readable_fulltext":
		source.Text = fmt.Sprintf("# %s\n\nSource: %s\n\n## Abstract\n\n%s\n\n## Full text (HTML extraction)\n\n%s", snapshot.Title, snapshot.SourceURL, snapshot.Abstract, snapshot.Text)
		source.Provenance = app.Provenance{Processor: "paper_url+summarize-readability", Version: version, Mode: "public_html_fulltext", Source: snapshot.SourceURL}
	case "abstract_only":
		source.Text = fmt.Sprintf("# %s\n\nSource: %s\n\n## Abstract (metadata only)\n\n%s", snapshot.Title, snapshot.SourceURL, snapshot.Abstract)
		source.Provenance = app.Provenance{Processor: "paper_url+summarize-readability", Version: version, Mode: "public_html_abstract_only", Source: snapshot.SourceURL}
	default:
		return source, &apierrors.ServiceError{Code: apierrors.EvidenceMissing, Message: "Access restriction detected; no entitlement bypass", RequiredAction: "choose_accessible_public_paper_or_provide_existing_export"}
	}
	return source, nil
}

// aclAnthologyID matches the ACL Anthology identifier embedded in its DOI and
// PDF/landing URLs (e.g. 2024.acl-long.1, P19-1001, D19-1001).
var aclAnthologyID = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9._-]*$`)

// paperPDFSourceKey derives a stable material identity from a public PDF URL.
// The ACL Anthology embeds its identifier and thus its DOI in the URL, so the
// landing page and PDF share one source key; other sites fall back to the
// canonical URL. A PDF link is never presented as full text itself.
func paperPDFSourceKey(pdfURL string) string {
	u, err := url.Parse(pdfURL)
	if err == nil && (u.Hostname() == "aclanthology.org" || u.Hostname() == "www.aclanthology.org") {
		id := strings.TrimSuffix(strings.TrimPrefix(u.Path, "/"), ".pdf")
		if id != "" && !strings.Contains(id, "/") && aclAnthologyID.MatchString(id) {
			return "doi:10.18653/v1/" + id
		}
	}
	return canonicalWebKey(pdfURL)
}

// pdfName returns a stable attachment stem for a PDF URL, used for the stored
// original document and its extraction metadata.
func pdfName(pdfURL string) string {
	u, err := url.Parse(pdfURL)
	if err != nil {
		return "paper"
	}
	parts := strings.Split(strings.Trim(u.Path, "/"), "/")
	last := parts[len(parts)-1]
	if strings.HasSuffix(strings.ToLower(last), ".pdf") {
		last = last[:len(last)-4]
	}
	if last == "" {
		return "paper"
	}
	return last
}

// paperPDFSource builds an importable paper source from an already-downloaded
// public PDF and its summarize extraction. The PDF is stored as an original
// attachment; its extracted markdown is the full text. The PDF URL is never
// emitted as metadata full text, and a restricted/paywalled or empty document
// never reaches this point.
func paperPDFSource(pdfURL, title string, text string, pdf, original []byte) app.ImportedSource {
	name := pdfName(pdfURL)
	if title == "" {
		title = name
	}
	return app.ImportedSource{
		SourceKey:     paperPDFSourceKey(pdfURL),
		SourceLocator: pdfURL,
		Kind:          "paper",
		Title:         title,
		Text:          fmt.Sprintf("# %s\n\nSource: %s\n\n## Full text (PDF extraction)\n\n%s", title, pdfURL, text),
		SourceSpans:   []string{"whole PDF document: " + pdfURL},
		Provenance:    app.Provenance{Processor: "paper_pdf+summarize", Version: "summarize 0.25.1 + markitdown", Mode: "public_pdf_fulltext", Source: pdfURL},
		Attachments: []app.SourceAttachment{
			{Name: name + ".pdf", MediaType: "application/pdf", Data: pdf, SourceLocator: pdfURL},
			{Name: "summarize-pdf-original.json", MediaType: "application/json", Data: original, SourceLocator: pdfURL},
		},
	}
}
