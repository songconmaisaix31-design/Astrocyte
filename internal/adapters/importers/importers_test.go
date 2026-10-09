package importers

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
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

func TestReaderCanonicalDigestOriginalAndFileScope(t *testing.T) {
	raw := `{"input":{"url":"https://www.youtube.com/watch?v=video&t=22"},"extracted":{"content":"original transcript","title":"real title","transcriptSegments":[{"startMs":22500,"text":"segment","endMs":25000}]},"summary":"actual summary"}`
	r := NewReader(nil)
	s, err := r.ReadSource(context.Background(), app.ImportMaterialCommand{Adapter: "summarize_json", SourceLocator: "https://youtu.be/video", Kind: "video", ExportText: raw})
	if err != nil || s.SourceKey != "youtube:video" || s.Text != "original transcript" || s.Summary != "actual summary" || len(s.SourceSpans) != 1 || len(s.Attachments) != 1 || string(s.Attachments[0].Data) != raw {
		t.Fatalf("%+v %v", s, err)
	}
	sum := sha256.Sum256([]byte(s.Text))
	for _, digest := range []string{hex.EncodeToString(sum[:]), "caller-made-up-digest"} {
		_, err := r.ReadSource(context.Background(), app.ImportMaterialCommand{Adapter: "summarize_json", Kind: "video", ExportText: raw, ContentDigest: digest})
		if (err == nil) != (digest == hex.EncodeToString(sum[:])) {
			t.Fatal(digest, err)
		}
	}
	root := t.TempDir()
	path := filepath.Join(root, "export.json")
	if err := os.WriteFile(path, []byte(raw), 0o600); err != nil {
		t.Fatal(err)
	}
	_, err = r.ReadSource(context.Background(), app.ImportMaterialCommand{Adapter: "summarize_json", LocalFileRef: path})
	var service *apierrors.ServiceError
	if !errors.As(err, &service) || service.Code != apierrors.ScopeDenied {
		t.Fatal("default file read allowed", err)
	}
	r.AllowedRoots = []string{root}
	if _, err := r.ReadSource(context.Background(), app.ImportMaterialCommand{Adapter: "summarize_json", LocalFileRef: path}); err != nil {
		t.Fatal(err)
	}
	outside := filepath.Join(t.TempDir(), "export.json")
	os.WriteFile(outside, []byte(raw), 0o600)
	if _, err := r.ReadSource(context.Background(), app.ImportMaterialCommand{Adapter: "summarize_json", LocalFileRef: outside}); err == nil {
		t.Fatal("out of scope file read")
	}
	if _, err := r.ReadSource(context.Background(), app.ImportMaterialCommand{Adapter: "summarize", SourceLocator: "https://youtube.com/watch?v=video"}); err == nil {
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
	s, err := r.ReadSource(context.Background(), app.ImportMaterialCommand{Adapter: "arxiv", SourceLocator: id, Kind: "paper"})
	if err != nil {
		t.Fatal(err)
	}
	if len(s.Attachments) != 2 || !strings.HasPrefix(string(s.Attachments[0].Data), "%PDF-") || !strings.Contains(s.Text, "metadata only") || len(s.SourceSpans) != 1 {
		t.Fatalf("incomplete original source %+v", s)
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
	for _, alias := range []string{base + "?spm_id_from=333.337.search-card.all.click", base + "?spm=another&p=1#reply", "http://m.bilibili.com/video/BV1PReT6EEqR?spm_id_from=old"} {
		if canonicalWebKey(alias) != base {
			t.Fatalf("tracking alias %s creates a distinct source", alias)
		}
		// Actual export source is unchanged while callers can use tracked links.
		raw := fmt.Sprintf(`{"input":{"url":%q},"extracted":{"content":"protocol fixture content"}}`, base)
		s, err := NewReader(nil).ReadSource(context.Background(), app.ImportMaterialCommand{Adapter: "summarize", Kind: "video", SourceLocator: alias, ExportText: raw})
		if err != nil || s.SourceKey != base || s.SourceLocator != base || len(s.SourceSpans) != 0 {
			t.Fatalf("alias mismatch or fake timestamp: %+v %v", s, err)
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

func TestSummarizeTimingAndMissingOriginal(t *testing.T) {
	for _, tc := range []struct {
		raw   string
		ok    bool
		count int
	}{
		{`{"input":{"url":"https://www.youtube.com/watch?v=x"},"extracted":{"content":"words","transcriptSegments":[{"startMs":1500,"endMs":2900,"text":"words"}]},"summary":"summary"}`, true, 1},
		{`{"input":{"url":"https://www.youtube.com/watch?v=x"},"summary":"legacy summary"}`, true, 0},
		{`{"input":{"url":"https://example.org"},"extracted":{"content":"words","transcriptSegments":[{"endMs":2900,"text":"words"}]}}`, false, 0},
		{`{"input":{"url":"https://example.org"},"extracted":{"content":"words","transcriptSegments":[{"startMs":2900,"endMs":1500,"text":"words"}]}}`, false, 0},
		{`{"input":{"url":"https://example.org"},"summary":""}`, false, 0},
	} {
		v, err := ParseSummarizeJSON([]byte(tc.raw))
		if (err == nil) != tc.ok || len(v.Segments) != tc.count {
			t.Fatalf("%s: %v %+v", tc.raw, err, v)
		}
	}
	v, err := ParseSummarizeMarkdown("https://youtu.be/x", "title", []byte("# Prior summary\n[01:20] not a verified transcript"))
	if err != nil || len(v.Segments) != 0 {
		t.Fatal("invented timestamp from markdown")
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
	var fetched []string
	a := NewArxiv()
	a.Client.Transport = roundTripFunc(func(r *http.Request) (*http.Response, error) {
		fetched = append(fetched, r.URL.String())
		body := `<feed xmlns="http://www.w3.org/2005/Atom"><entry><id>http://arxiv.org/abs/2501.12345v2</id><title> Actual title </title><summary>Abstract only</summary></entry></feed>`
		if r.URL.Host == "arxiv.org" {
			body = "%PDF-1.7\noriginal bytes"
		}
		return &http.Response{StatusCode: 200, Body: io.NopCloser(strings.NewReader(body)), Header: make(http.Header)}, nil
	})
	p, err := a.Read(context.Background(), "2501.12345")
	if err != nil || p.ID != "2501.12345v2" || p.SourceKey != "arxiv:2501.12345" || !strings.HasPrefix(string(p.PDF), "%PDF-") || len(fetched) != 2 {
		t.Fatalf("%+v %v", p, err)
	}
	if _, err := a.Read(context.Background(), "2501.12345v1"); err == nil {
		t.Fatal("accepted changed pinned revision")
	}
	a.Client.Transport = roundTripFunc(func(r *http.Request) (*http.Response, error) { return nil, fmt.Errorf("network down") })
	if _, err := a.Read(context.Background(), "2501.12345"); err == nil {
		t.Fatal("hid network failure")
	}
}
