package importers

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"testing"
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
