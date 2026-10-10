package importers

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestPaperPDFExtractLive verifies the real summarize local-PDF extraction path
// (uvx + markitdown) through the production isolated child environment against
// an already-downloaded public ACL Anthology PDF. It never fetches a URL and
// never calls a model; the PDF bytes are supplied from disk under opt-in.
func TestPaperPDFExtractLive(t *testing.T) {
	if os.Getenv("ASTROCYTE_PAPER_PDF_LIVE") != "1" {
		t.Skip("live paper PDF extraction not requested")
	}
	pdfPath := os.Getenv("ASTROCYTE_PAPER_PDF_FILE")
	if pdfPath == "" {
		pdfPath = filepath.Join(os.TempDir(), "opencode", "w1-paper", "acl.pdf")
	}
	pdf, err := os.ReadFile(pdfPath)
	if err != nil {
		t.Fatalf("read acl.pdf: %v", err)
	}
	if !strings.HasPrefix(strings.TrimSpace(string(pdf)), "%PDF-") {
		t.Fatalf("not a PDF: %s", pdfPath)
	}
	e := installedPaperExtractor(t)
	result, err := e.ExtractPDF(context.Background(), "https://aclanthology.org/2024.acl-long.1.pdf", pdf)
	if err != nil {
		t.Fatalf("ExtractPDF: %v", err)
	}
	if len(result.Text) < 10000 {
		t.Fatalf("extracted full text too small: %d bytes", len(result.Text))
	}
	if len(result.Original) == 0 {
		t.Fatal("original summarize JSON missing")
	}
	t.Logf("live PDF extraction: text_bytes=%d original_json_bytes=%d", len(result.Text), len(result.Original))
}
