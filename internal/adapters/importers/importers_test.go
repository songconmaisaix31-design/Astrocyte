package importers

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/songconmaisaix31-design/Astrocyte/internal/apierrors"
	"github.com/songconmaisaix31-design/Astrocyte/internal/attention/app"
)

func TestInstalledSummarizeExport(t *testing.T) {
	raw, err := os.ReadFile("testdata/summarize-arxiv-manual.json")
	if err != nil {
		t.Fatal(err)
	}
	v, err := ParseSummarizeJSON(raw)
	if err != nil {
		t.Fatal(err)
	}
	if v.URL != "https://info.arxiv.org/help/api/user-manual.html" || !strings.Contains(v.Text, "Atom") || len(v.Segments) != 0 {
		t.Fatal("lost real source/text or invented timing")
	}
}

func TestPaperExportIdentityRetainsOfficialVersionAndOriginal(t *testing.T) {
	for _, tc := range []struct{ kind, locator, expected string }{
		{"paper", "https://arxiv.org/html/2504.16054v1", "arxiv:2504.16054"},
		{"paper", "https://www.arxiv.org/pdf/2504.16054v2.pdf", "arxiv:2504.16054"},
		{"paper", "https://export.arxiv.org/abs/hep-th/9901001v1", "arxiv:hep-th/9901001"},
		{"video", "https://arxiv.org/html/2504.16054v1", "https://arxiv.org/html/2504.16054v1"},
		{"paper", "https://example.org/html/2504.16054v1", "https://example.org/html/2504.16054v1"},
		{"paper", "https://arxiv.org/html/not-a-paper", "https://arxiv.org/html/not-a-paper"},
	} {
		t.Run(tc.kind+tc.locator, func(t *testing.T) {
			if exportSourceKey(tc.locator, tc.kind) != tc.expected {
				t.Fatal("official version identity changed", tc.locator)
			}
		})
	}
	actual, err := installedExtractor(t).Extract(context.Background(), "https://arxiv.org/html/2501.12948v1")
	if err != nil {
		t.Fatal(err)
	}
	source, err := NewReader(nil).ReadSource(context.Background(), app.ImportMaterialCommand{Adapter: "summarize_json", Kind: "paper", SourceLocator: actual.URL, ExportText: string(actual.Original)})
	if err != nil || source.SourceKey != "arxiv:2501.12948" || source.SourceLocator != actual.URL || source.Provenance.Source != actual.URL || source.Provenance.Mode != "existing_json_export" || source.Provenance.Version != "0.21.8-format" || len(source.Attachments) != 1 || string(source.Attachments[0].Data) != string(actual.Original) {
		t.Fatal("actual export source/version/original changed", err)
	}
}

func TestReaderCanonicalDigestOriginalAndFileScope(t *testing.T) {
	rawBytes, err := os.ReadFile("testdata/summarize-arxiv-manual.json")
	if err != nil {
		t.Fatal(err)
	}
	raw := string(rawBytes)
	actual, err := ParseSummarizeJSON(rawBytes)
	if err != nil {
		t.Fatal(err)
	}
	r := NewReader(nil)
	s, err := r.ReadSource(context.Background(), app.ImportMaterialCommand{Adapter: "summarize_json", SourceLocator: actual.URL, Kind: "paper", ExportText: raw})
	if err != nil || s.SourceKey != canonicalWebKey(actual.URL) || s.Text != actual.Text || s.Summary != actual.Summary || len(s.SourceSpans) != len(actual.Segments) || len(s.Attachments) != 1 || string(s.Attachments[0].Data) != raw {
		t.Fatal("actual reader export, digest source or original changed", err)
	}
	timed := videoActualExport(t, true)
	if len(exportSpans(timed)) != len(timed.Segments) || len(timed.Segments) == 0 {
		t.Fatal("actual media time spans missing")
	}
	sum := sha256.Sum256([]byte(s.Text))
	for _, digest := range []string{hex.EncodeToString(sum[:]), "caller-made-up-digest"} {
		_, err := r.ReadSource(context.Background(), app.ImportMaterialCommand{Adapter: "summarize_json", Kind: "paper", ExportText: raw, ContentDigest: digest})
		if (err == nil) != (digest == hex.EncodeToString(sum[:])) {
			t.Fatal("actual digest validation failed", err)
		}
	}
	root := t.TempDir()
	path := filepath.Join(root, "export.json")
	if err := os.WriteFile(path, []byte(raw), 0o600); err != nil {
		t.Fatal(err)
	}
	_, err = r.ReadSource(context.Background(), app.ImportMaterialCommand{Adapter: "summarize_json", Kind: "paper", LocalFileRef: path})
	var service *apierrors.ServiceError
	if !errors.As(err, &service) || service.Code != apierrors.ScopeDenied {
		t.Fatal("default file read allowed", err)
	}
	r.AllowedRoots = []string{root}
	if _, err := r.ReadSource(context.Background(), app.ImportMaterialCommand{Adapter: "summarize_json", Kind: "paper", LocalFileRef: path}); err != nil {
		t.Fatal(err)
	}
	outside := filepath.Join(t.TempDir(), "export.json")
	if err := os.WriteFile(outside, []byte(raw), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := r.ReadSource(context.Background(), app.ImportMaterialCommand{Adapter: "summarize_json", Kind: "paper", LocalFileRef: outside}); err == nil {
		t.Fatal("out of scope file read")
	}
	if _, err := r.ReadSource(context.Background(), app.ImportMaterialCommand{Adapter: "summarize", SourceLocator: "https://www.bilibili.com/video/BV1PReT6EEqR/"}); err == nil {
		t.Fatal("invented unavailable backend result")
	}
}

// Explicit opt-in uses the real official HTTP source and full original. It does
// not choose an acceptance paper for the user or silently fall back to mocks.
func TestArxivLiveOfficialSource(t *testing.T) {
	id := os.Getenv("ASTROCYTE_TEST_ARXIV_ID")
	if id == "" {
		t.Skip("set ASTROCYTE_TEST_ARXIV_ID to the approved public paper")
	}
	r := NewReader(nil)
	fullText := os.Getenv("ASTROCYTE_TEST_SUMMARIZE_CLI") != ""
	if fullText {
		extractor, err := NewSummarizeExtractor(os.Getenv("ASTROCYTE_TEST_SUMMARIZE_NODE"), os.Getenv("ASTROCYTE_TEST_SUMMARIZE_CLI"))
		if err != nil {
			t.Fatal(err)
		}
		r.Arxiv.TextExtractor = extractor
	}
	s, err := r.ReadSource(context.Background(), app.ImportMaterialCommand{Adapter: "arxiv", SourceLocator: id, Kind: "paper"})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(string(s.Attachments[0].Data), "%PDF-") {
		t.Fatalf("incomplete original source %+v", s)
	}
	if fullText {
		if len(s.Attachments) != 4 || len(s.Text) < 5000 || !strings.Contains(s.Text, "Introduction") || !strings.Contains(s.Text, "Original HTML text") || len(s.SourceSpans) != 2 || s.Provenance.Mode != "official_atom_pdf_and_html_text" {
			t.Fatal("lost original HTML text or provenance")
		}
		t.Logf("HTML_bytes=%d extracted_text_chars=%d mode=%s", len(s.Attachments[2].Data), len(s.Text), s.Provenance.Mode)
	} else if len(s.Attachments) != 2 || !strings.Contains(s.Text, "metadata only") || len(s.SourceSpans) != 1 {
		t.Fatal("metadata scope changed without extractor")
	}
	t.Logf("source=%s fixed=%s PDF_bytes=%d", s.SourceKey, s.SourceLocator, len(s.Attachments[0].Data))
	if dir := os.Getenv("ASTROCYTE_TEST_EXPORT_DIR"); dir != "" {
		if err := os.MkdirAll(dir, 0o700); err != nil {
			t.Fatal(err)
		}
		for _, a := range s.Attachments {
			if err := os.WriteFile(filepath.Join(dir, filepath.Base(a.Name)), a.Data, 0o600); err != nil {
				t.Fatal(err)
			}
		}
		if err := os.WriteFile(filepath.Join(dir, "arxiv-imported-text.md"), []byte(s.Text), 0o600); err != nil {
			t.Fatal(err)
		}
		sum := sha256.Sum256([]byte(s.Text))
		command := app.ImportMaterialCommand{CommandMeta: app.CommandMeta{SchemaVersion: 1, RequestID: "selected-arxiv-live-export", ExpectedVersion: 1}, Adapter: "manual", SourceKey: "", SourceLocator: s.SourceLocator, Kind: "paper", ContentDigest: hex.EncodeToString(sum[:]), ExportText: s.Text, SourceSpans: s.SourceSpans, Title: s.Title}
		data, err := json.MarshalIndent(command, "", "  ")
		if err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(dir, "arxiv-text-command.json"), data, 0o600); err != nil {
			t.Fatal(err)
		}
		t.Logf("actual original/metadata/text export=%s (manual text command is not full-original arxiv import)", dir)
	}
}

func TestBilibiliTrackingCanonicalSource(t *testing.T) {
	base := "https://www.bilibili.com/video/BV1PReT6EEqR/"
	root := os.Getenv("ASTROCYTE_TEST_VIDEO_BASELINE_DIR")
	file := filepath.Join(root, "video-original-text.txt")
	if !filepath.IsAbs(file) {
		t.Fatal("actual video baseline directory required")
	}
	reader := NewReader([]string{root})
	actual := videoActualExport(t, false)
	for _, alias := range []string{base + "?spm_id_from=333.337.search-card.all.click", base + "?spm=another&p=1#reply", "http://m.bilibili.com/video/BV1PReT6EEqR?spm_id_from=old"} {
		if canonicalWebKey(alias) != base {
			t.Fatalf("tracking alias %s creates a distinct source", alias)
		}
		s, err := reader.ReadSource(context.Background(), app.ImportMaterialCommand{Adapter: "summarize_markdown", Kind: "video", SourceLocator: alias, LocalFileRef: file})
		if err != nil || s.SourceKey != base || canonicalWebKey(s.SourceLocator) != base || s.Text != actual.Text || len(s.SourceSpans) != 0 {
			t.Fatal("actual alias identity, transcript or missing timeline changed", err)
		}
	}
	for _, different := range []string{base + "?p=2", base + "?meaningful=other", "https://www.bilibili.com/video/BV1111111111/"} {
		if canonicalWebKey(different) == base {
			t.Fatalf("collapsed different content %s", different)
		}
	}
	other := "https://example.org/video/BV1PReT6EEqR/?spm_id_from=keep"
	if canonicalWebKey(other) != other {
		t.Fatal("applied Bilibili rule to another source")
	}
}

func TestBilibiliPageOnlyIsNotVideoEvidence(t *testing.T) {
	cli, _, _ := videoActualInputs(t, false)
	for _, samePageSummary := range []bool{false, true} {
		raw := cli
		if samePageSummary {
			var value map[string]any
			if err := json.Unmarshal(cli, &value); err != nil {
				t.Fatal(err)
			}
			// 显式无字幕页面即使复制页面到 summary 字段，也必须拒绝。
			value["summary"] = value["extracted"].(map[string]any)["content"]
			var err error
			raw, err = json.Marshal(value)
			if err != nil {
				t.Fatal(err)
			}
		}
		_, err := NewReader(nil).ReadSource(context.Background(), app.ImportMaterialCommand{Adapter: "summarize_json", Kind: "video", ExportText: string(raw)})
		var service *apierrors.ServiceError
		if !errors.As(err, &service) || service.Code != apierrors.EvidenceMissing {
			t.Fatalf("page-only export accepted: %v", err)
		}
	}
	actual := videoActualExport(t, false)
	if e, err := ParseSummarizeMarkdown(actual.URL, actual.Title, []byte(actual.Text)); err != nil || e.Text != actual.Text || len(e.Segments) != 0 || e.HasOriginal {
		t.Fatal("actual legacy plain export rejected or claimed verified time/original", err)
	}
	// Opt in to the preserved actual installed CLI output, outside this repo.
	if path := os.Getenv("ASTROCYTE_TEST_BILIBILI_EXPORT"); path != "" {
		raw, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		_, err = ParseSummarizeJSON(raw)
		var service *apierrors.ServiceError
		if !errors.As(err, &service) || service.Code != apierrors.EvidenceMissing {
			t.Fatalf("real page-only export accepted: %v", err)
		}
	}
}

func TestSummarizeTimingAndMissingOriginal(t *testing.T) {
	actual := videoActualExport(t, true)
	if !actual.HasOriginal || len(actual.Segments) != 952 || actual.Segments[0].EndMS == nil || actual.Segments[0].StartMS < 0 || *actual.Segments[0].EndMS < actual.Segments[0].StartMS {
		t.Fatal("actual native time or original missing")
	}
	cli, _, _ := videoActualInputs(t, true)
	if _, err := ParseSummarizeJSON(cli); err == nil {
		t.Fatal("actual CLI page-only output accepted as timed video")
	}
	// 以下仅为 JSON 字段校验的非法输入，不代表外部成功输出。
	for _, raw := range []string{
		`{"input":{"url":"https://example.org"},"extracted":{"content":"words","transcriptSegments":[{"endMs":2900,"text":"words"}]}}`,
		`{"input":{"url":"https://example.org"},"extracted":{"content":"words","transcriptSegments":[{"startMs":2900,"endMs":1500,"text":"words"}]}}`,
		`{"input":{"url":"https://example.org"},"summary":""}`,
	} {
		if _, err := ParseSummarizeJSON([]byte(raw)); err == nil {
			t.Fatal("accepted invalid JSON time or empty content")
		}
	}
	v, err := ParseSummarizeMarkdown(actual.URL, actual.Title, []byte(actual.Text))
	if err != nil || v.Text != actual.Text || len(v.Segments) != 0 || v.HasOriginal {
		t.Fatal("plain export fabricated verified timeline/original", err)
	}
}

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

func TestArxivRevisionAndOriginal(t *testing.T) {
	for _, id := range []string{"2501.12345v2", "https://arxiv.org/pdf/2501.12345v2.pdf", "arXiv:hep-th/9901001v1"} {
		if _, err := NormalizeArxivID(id); err != nil {
			t.Fatal(err)
		}
	}
	for _, id := range []string{"https://evil.test/abs/2501.12345", "https://arxiv.org:99/abs/2501.12345", "../../x", "2501.12345v0"} {
		if _, err := NormalizeArxivID(id); err == nil {
			t.Fatal("accepted", id)
		}
	}
	a := NewArxiv()
	p, err := a.Read(context.Background(), "2501.12948v2")
	if err != nil || p.ID != "2501.12948v2" || p.SourceKey != "arxiv:2501.12948" || p.Title == "" || len(p.Metadata) == 0 || len(p.PDF) < 1000 || !strings.HasPrefix(string(p.PDF), "%PDF-") {
		t.Fatal("actual fixed official revision or original unavailable", err)
	}
	if _, err := a.Read(context.Background(), "2501.12948v999999"); err == nil {
		t.Fatal("accepted unpublished fixed revision")
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := a.Read(ctx, "2501.12948v2"); !errors.Is(err, context.Canceled) {
		t.Fatal("actual HTTP cancellation failure was hidden", err)
	}
}
