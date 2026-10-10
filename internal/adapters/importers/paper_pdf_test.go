package importers

import (
	"context"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/songconmaisaix31-design/Astrocyte/internal/attention/app"
)

func TestPaperPDFSourceKey(t *testing.T) {
	cases := []struct {
		pdf  string
		want string
	}{
		{"https://aclanthology.org/2024.acl-long.1.pdf", "doi:10.18653/v1/2024.acl-long.1"},
		{"https://aclanthology.org/P19-1001.pdf", "doi:10.18653/v1/P19-1001"},
		// Non-ACL PDFs fall back to the canonical URL, never a fabricated DOI.
		{"https://proceedings.mlr.press/v202/dettmers23a/dettmers23a.pdf", "https://proceedings.mlr.press/v202/dettmers23a/dettmers23a.pdf"},
	}
	for _, c := range cases {
		if got := paperPDFSourceKey(c.pdf); got != c.want {
			t.Errorf("paperPDFSourceKey(%s) = %q, want %q", c.pdf, got, c.want)
		}
	}
}

func TestPaperPDFSourceNeverPresentsURLAsFullText(t *testing.T) {
	src := paperPDFSource("https://aclanthology.org/2024.acl-long.1.pdf", "A Paper", "extracted full paper text", []byte("%PDF-1.5 original"), []byte(`{"extracted":{"content":"extracted full paper text"}}`))
	if src.SourceKey != "doi:10.18653/v1/2024.acl-long.1" {
		t.Fatalf("source key wrong: %s", src.SourceKey)
	}
	if src.Provenance.Mode != "public_pdf_fulltext" || src.Provenance.Processor != "paper_pdf+summarize" {
		t.Fatalf("provenance wrong: %+v", src.Provenance)
	}
	if !strings.Contains(src.Text, "extracted full paper text") {
		t.Fatalf("extracted full text missing: %q", src.Text)
	}
	if len(src.Attachments) != 2 || src.Attachments[0].Name != "2024.acl-long.1.pdf" || src.Attachments[0].MediaType != "application/pdf" {
		t.Fatalf("original PDF attachment missing: %+v", src.Attachments)
	}
	if src.Attachments[1].Name != "summarize-pdf-original.json" {
		t.Fatalf("summarize extraction attachment missing: %+v", src.Attachments)
	}
	if !strings.Contains(src.SourceSpans[0], "whole PDF document") {
		t.Fatalf("PDF span missing: %v", src.SourceSpans)
	}
}

func TestPaperPDFReaderValidation(t *testing.T) {
	r := NewReader(nil)
	if _, err := r.ReadSource(context.Background(), app.ImportMaterialCommand{Adapter: "paper_pdf", Kind: "paper", SourceLocator: "https://aclanthology.org/2024.acl-long.1.pdf"}); err == nil {
		t.Fatal("paper_pdf without extractor accepted")
	}
	if _, err := r.ReadSource(context.Background(), app.ImportMaterialCommand{Adapter: "paper_pdf", Kind: "video", SourceLocator: "https://aclanthology.org/2024.acl-long.1.pdf"}); err == nil {
		t.Fatal("paper_pdf accepted non-paper kind")
	}
	if _, err := r.ReadSource(context.Background(), app.ImportMaterialCommand{Adapter: "paper_pdf", Kind: "paper", ExportText: "existing"}); err == nil {
		t.Fatal("paper_pdf accepted existing export")
	}
}

func TestPaperPDFRequiresPDFMagic(t *testing.T) {
	w := NewPaperWeb()
	w.Client.Transport = roundTripFunc(func(req *http.Request) (*http.Response, error) {
		return &http.Response{StatusCode: 200, Body: io.NopCloser(strings.NewReader(`<html>login required</html>`)), Header: make(http.Header)}, nil
	})
	if _, err := w.FetchPDF(context.Background(), "https://aclanthology.org/2024.acl-long.1.pdf", 64<<20); err == nil {
		t.Fatal("HTML wall accepted as a PDF")
	}
}
