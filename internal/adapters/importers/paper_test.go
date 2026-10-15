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
		t.Fatal(err)
	}
	node, err = filepath.Abs(node)
	if err != nil {
		t.Fatal(err)
	}
	cli, err := filepath.Abs("../../../node_modules/@steipete/summarize/dist/cli.js")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(cli); err != nil {
		t.Fatal(err)
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
	directory := os.Getenv("ASTROCYTE_TEST_PUBLIC_PAPER_DIR")
	if directory == "" || !filepath.IsAbs(directory) {
		t.Fatal("explicit actual public paper originals required")
	}
	html, err := os.ReadFile(filepath.Join(directory, "acl-public.html"))
	if err != nil {
		t.Fatal(err)
	}
	result, err := e.InspectPaperHTML(context.Background(), "https://aclanthology.org/2024.acl-long.1/", html)
	if err != nil {
		t.Fatal(err)
	}
	if result.SourceKey != "doi:10.18653/v1/2024.acl-long.1" || result.Title == "" || result.Abstract == "" || result.Text != "" || result.Provenance.Version == "" {
		t.Fatal("actual metadata lost or misrepresented")
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

func TestPaperDOMActualCrossHostFullTextAndIdentity(t *testing.T) {
	root := os.Getenv("ASTROCYTE_PAPER_CAPTURE_ROOT")
	if root == "" {
		t.Skip("actual public HTML capture root not provided")
	}
	e := installedPaperExtractor(t)
	keys := []string{}
	for _, source := range []struct{ name, url string }{
		{"pmc", "https://pmc.ncbi.nlm.nih.gov/articles/PMC11135672/"},
		{"plos", "https://journals.plos.org/digitalhealth/article?id=10.1371/journal.pdig.0000514"},
	} {
		html, err := os.ReadFile(filepath.Join(root, source.name+".html"))
		if err != nil {
			t.Fatal(err)
		}
		result, err := e.InspectPaperHTML(context.Background(), source.url, html)
		if err != nil {
			t.Fatal(err)
		}
		if result.ContentState != "readable_fulltext" || len(result.Text) < 30000 || len(result.Abstract) < 1000 || result.Truncated {
			t.Fatalf("actual full text missing: state=%s text=%d abstract=%d", result.ContentState, len(result.Text), len(result.Abstract))
		}
		keys = append(keys, result.SourceKey)
		t.Logf("actual %s fulltext_bytes=%d abstract_bytes=%d identity=%s", source.name, len(result.Text), len(result.Abstract), result.SourceKey)
	}
	if keys[0] != keys[1] || keys[0] != "doi:10.1371/journal.pdig.0000514" {
		t.Fatal("publisher and archive identity not shared", keys)
	}
}
