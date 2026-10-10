package importers

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func installedPaperExtractor(t *testing.T) *SummarizeExtractor {
	t.Helper()
	node, err := exec.LookPath("node")
	if err != nil {
		t.Skip("Node unavailable")
	}
	node, _ = filepath.Abs(node)
	cli, err := filepath.Abs("../../../node_modules/@steipete/summarize/dist/cli.js")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(cli); err != nil {
		t.Skip("install locked project dependencies before installed parser verification")
	}
	e, err := NewSummarizeExtractor(node, cli)
	if err != nil {
		t.Fatal(err)
	}
	return e
}

func TestPaperDOMInstalledPinAndIsolation(t *testing.T) {
	e := installedPaperExtractor(t)
	t.Setenv("NODE_OPTIONS", "--require nonexisting-sentinel")
	t.Setenv("OPENAI_API_KEY", "synthetic-unavailable")
	html := []byte(`<html><head><meta name="citation_title" content="Paper"><meta name="citation_doi" content="10.1000/ABC"><meta name="citation_abstract" content="Only abstract"></head><body><article><p>Only abstract</p></article></body></html>`)
	result, err := e.InspectPaperHTML(context.Background(), "https://dl.acm.org/doi/10.1000/ABC", html)
	if err != nil {
		t.Fatal(err)
	}
	if result.SourceKey != "doi:10.1000/abc" || result.Title != "Paper" || result.Text != "" || result.Provenance.Version == "" {
		t.Fatalf("metadata lost or misrepresented: %+v", result)
	}
	for _, locator := range []string{"file:///private/file", "https://user:pass@nature.com/article", "http://nature.com/article"} {
		if _, err := e.InspectPaperHTML(context.Background(), locator, html); err == nil {
			t.Fatal("accepted invalid source", locator)
		}
	}
	if _, err := e.InspectPaperHTML(context.Background(), "https://nature.com/article", make([]byte, (16<<20)+1)); err == nil {
		t.Fatal("oversize observation accepted")
	}
}

// Opt-in captured public-source check never navigates or imports into a library.
// Captures are research observations from explicit official URLs, not fixtures.
func TestPaperDOMActualPublicHostCaptures(t *testing.T) {
	root := os.Getenv("ASTROCYTE_PAPER_CAPTURE_ROOT")
	if root == "" {
		t.Skip("actual public HTML capture root not provided")
	}
	e := installedPaperExtractor(t)
	for _, source := range []struct{ name, url, family string }{
		{"acl", "https://aclanthology.org/2024.acl-long.1/", "acl"},
		{"pmlr", "https://proceedings.mlr.press/v202/dettmers23a.html", "pmlr"},
		{"cvf", "https://openaccess.thecvf.com/content_cvpr_2016/html/He_Deep_Residual_Learning_CVPR_2016_paper.html", "cvf"},
	} {
		t.Run(source.name, func(t *testing.T) {
			html, err := os.ReadFile(filepath.Join(root, source.name+".html"))
			if err != nil {
				t.Fatal(err)
			}
			result, err := e.InspectPaperHTML(context.Background(), source.url, html)
			if err != nil {
				t.Fatal(err)
			}
			if result.Title == "" || result.HostFamily != source.family || len(result.PDFURLs) == 0 || result.Text != "" || result.ContentState == "readable_fulltext" {
				t.Fatalf("landing metadata incorrectly classified: %+v", result)
			}
			t.Logf("actual %s title=%q abstract_bytes=%d doi=%q state=%s pdfs=%d", source.family, result.Title, len(result.Abstract), result.DOI, result.ContentState, len(result.PDFURLs))
		})
	}
}
