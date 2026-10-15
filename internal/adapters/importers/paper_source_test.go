package importers

import (
	"bytes"
	"context"
	"net"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func actualPublicPaperSnapshot(t *testing.T, name, locator string) (PaperSnapshot, []byte) {
	t.Helper()
	directory := os.Getenv("ASTROCYTE_TEST_PUBLIC_PAPER_DIR")
	if directory == "" || !filepath.IsAbs(directory) {
		t.Fatal("explicit approved public paper originals required")
	}
	body, err := os.ReadFile(filepath.Join(directory, name))
	if err != nil {
		t.Fatal(err)
	}
	snapshot, err := installedPaperExtractor(t).InspectPaperHTML(context.Background(), locator, body)
	if err != nil {
		t.Fatal(err)
	}
	return snapshot, body
}

func TestPaperSourceMetadataOnly(t *testing.T) {
	snap, html := actualPublicPaperSnapshot(t, "acl-public.html", "https://aclanthology.org/2024.acl-long.1/")
	if snap.ContentState != "abstract_only" || snap.Abstract == "" || snap.Text != "" {
		t.Fatal("actual public ACL metadata observation changed")
	}
	src, err := paperSource(snap, html)
	if err != nil {
		t.Fatal(err)
	}
	if src.Kind != "paper" || src.Title != snap.Title || src.Summary != snap.Abstract || src.SourceKey != "doi:10.18653/v1/2024.acl-long.1" {
		t.Fatal("actual metadata fields lost")
	}
	if !strings.Contains(src.Text, "metadata only") || strings.Contains(src.Text, "Full text") {
		t.Fatalf("metadata presented as full text: %q", src.Text)
	}
	if src.Provenance.Mode != "public_html_abstract_only" {
		t.Fatalf("wrong provenance mode: %s", src.Provenance.Mode)
	}
	if len(src.Attachments) != 1 || src.Attachments[0].Name != "source.html" || !bytes.Equal(src.Attachments[0].Data, html) {
		t.Fatal("actual source HTML attachment missing or changed")
	}
}

func TestPaperSourceFulltext(t *testing.T) {
	snap, html := actualPublicPaperSnapshot(t, "plos-public.html", "https://journals.plos.org/digitalhealth/article?id=10.1371/journal.pdig.0000514")
	if snap.ContentState != "readable_fulltext" || len(snap.Text) < 30000 || snap.Truncated {
		t.Fatal("actual full paper observation changed")
	}
	src, err := paperSource(snap, html)
	if err != nil {
		t.Fatal(err)
	}
	if src.SourceKey != "doi:10.1371/journal.pdig.0000514" {
		t.Fatalf("source key wrong: %s", src.SourceKey)
	}
	if !strings.Contains(src.Text, snap.Text) || src.Title != snap.Title || len(src.Attachments) != 1 || !bytes.Equal(src.Attachments[0].Data, html) {
		t.Fatal("actual full text or original attachment missing")
	}
	if src.Provenance.Mode != "public_html_fulltext" {
		t.Fatalf("wrong provenance mode: %s", src.Provenance.Mode)
	}
}

func TestPaperSourceRestrictedRefused(t *testing.T) {
	snap, html := actualPublicPaperSnapshot(t, "acl-public.html", "https://aclanthology.org/2024.acl-long.1/")
	// 此状态是拒绝规则的明确函数参数，不声明观察到真实付费页面。
	snap.ContentState = "restricted"
	if _, err := paperSource(snap, html); err == nil {
		t.Fatal("restricted page accepted as importable source")
	}
}

func TestIsPublicIP(t *testing.T) {
	cases := []struct {
		addr string
		want bool
	}{
		{"8.8.8.8", true},
		{"1.1.1.1", true},
		{"127.0.0.1", false},
		{"10.0.0.1", false},
		{"172.16.0.1", false},
		{"192.168.1.1", false},
		{"169.254.1.1", false},
		{"0.0.0.0", false},
		{"100.64.0.1", false},
		{"100.127.255.254", false},
		{"192.0.0.9", false},
		{"192.0.2.1", false},
		{"198.18.0.1", false},
		{"198.19.255.254", false},
		{"198.51.100.1", false},
		{"203.0.113.1", false},
		{"240.0.0.1", false},
		{"255.255.255.255", false},
		{"::1", false},
		{"fc00::1", false},
		{"fe80::1", false},
		{"::ffff:10.0.0.1", false},
		{"::ffff:192.168.1.1", false},
		{"2606:4700:4700::1111", true},
		{"2001:4860:4860::8888", true},
		{"2a00:1450:4001::", true},
		// 文档地址和丢弃专用地址保留拒绝规则。
		{"2001:db8::1", false},
		{"2001:db8:0:0:0:0:0:1", false},
		{"2001:db8:ffff:ffff:ffff:ffff:ffff:ffff", false},
		{"100::", false},
		{"100::1", false},
	}
	for _, c := range cases {
		if got := isPublicIP(net.ParseIP(c.addr)); got != c.want {
			t.Errorf("isPublicIP(%s) = %v, want %v", c.addr, got, c.want)
		}
	}
}

func TestValidatePublicPaperURL(t *testing.T) {
	for _, raw := range []string{"http://example.com/x", "https://user:pass@example.com/x", "https://example.com:8080/x"} {
		u, err := url.Parse(raw)
		if err != nil {
			t.Fatalf("parse %s: %v", raw, err)
		}
		if err := validatePublicPaperURL(u); err == nil {
			t.Errorf("accepted non-compliant URL %s", raw)
		}
	}
	u, err := url.Parse("https://example.com/x")
	if err != nil {
		t.Fatal(err)
	}
	if validatePublicPaperURL(u) != nil {
		t.Error("rejected compliant public HTTPS URL")
	}
}
