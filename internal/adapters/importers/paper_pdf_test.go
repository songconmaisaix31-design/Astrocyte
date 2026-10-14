package importers

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/songconmaisaix31-design/Astrocyte/internal/apierrors"
	"github.com/songconmaisaix31-design/Astrocyte/internal/attention/app"
)

func TestPaperPDFSourceKey(t *testing.T) {
	cases := []struct {
		pdf  string
		want string
	}{
		{"https://aclanthology.org/2024.acl-long.1.pdf", "doi:10.18653/v1/2024.acl-long.1"},
		{"https://aclanthology.org/P19-1001.pdf", "doi:10.18653/v1/P19-1001"},
		// 其他站点使用其 canonical URL。
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
	// 使用已完整取得的真实公开 HTML，核生产格式检查，不声明新的 HTTP 访问。
	directory := os.Getenv("ASTROCYTE_TEST_PUBLIC_PAPER_DIR")
	if directory == "" || !filepath.IsAbs(directory) {
		t.Fatal("explicit approved actual public paper directory required")
	}
	html, err := os.ReadFile(filepath.Join(directory, "acl-public.html"))
	if err != nil {
		t.Fatal(err)
	}
	if len(html) != 47355 {
		t.Fatal("approved full public ACL HTML changed")
	}
	err = requirePDFContent(html)
	var failure *apierrors.ServiceError
	if !errors.As(err, &failure) || failure.Code != apierrors.ValidationFailed || failure.Message != "paper PDF source did not return a PDF document" {
		t.Fatalf("actual public HTML did not reach the PDF format rejection: %v", err)
	}
}
