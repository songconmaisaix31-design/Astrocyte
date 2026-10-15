package importers

import (
	"bytes"
	"context"
	"encoding/json"
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
	directory := os.Getenv("ASTROCYTE_TEST_PDF_PUBLICATION")
	if directory == "" || !filepath.IsAbs(directory) {
		t.Fatal("explicit actual public PDF publication required")
	}
	read := func(name string) []byte {
		body, err := os.ReadFile(filepath.Join(directory, name))
		if err != nil {
			t.Fatal(err)
		}
		return body
	}
	var source struct {
		SourceURL string `json:"source_url"`
		Title     string `json:"title"`
	}
	if err := json.Unmarshal(read("source.json"), &source); err != nil {
		t.Fatal(err)
	}
	if source.SourceURL != "https://arxiv.org/pdf/2501.12948v1" || source.Title == "" {
		t.Fatal("actual fixed approved PDF identity changed")
	}
	pdf, original, text := read("paper.pdf"), read("summarize-pdf-original.json"), read("paper-original-text.txt")
	if err := requirePDFContent(pdf); err != nil {
		t.Fatal(err)
	}
	var output struct {
		Extracted struct {
			Content string `json:"content"`
		} `json:"extracted"`
	}
	if err := json.Unmarshal(original, &output); err != nil || output.Extracted.Content != string(text) || len(text) < 10000 {
		t.Fatal("actual PDF CLI text and retained original differ", err)
	}
	src := paperPDFSource(source.SourceURL, source.Title, string(text), pdf, original)
	if src.SourceKey != source.SourceURL {
		t.Fatalf("source key wrong: %s", src.SourceKey)
	}
	if src.Provenance.Mode != "public_pdf_fulltext" || src.Provenance.Processor != "paper_pdf+summarize" {
		t.Fatalf("provenance wrong: %+v", src.Provenance)
	}
	if !strings.Contains(src.Text, string(text)) {
		t.Fatal("actual extracted full text missing")
	}
	if len(src.Attachments) != 2 || src.Attachments[0].Name != "2501.12948v1.pdf" || src.Attachments[0].MediaType != "application/pdf" || !bytes.Equal(src.Attachments[0].Data, pdf) {
		t.Fatal("actual original PDF attachment missing or changed")
	}
	if src.Attachments[1].Name != "summarize-pdf-original.json" || src.Attachments[1].MediaType != app.NativePrivateAttachmentMediaType || !bytes.Equal(src.Attachments[1].Data, original) {
		t.Fatal("actual private CLI original attachment missing or changed")
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
