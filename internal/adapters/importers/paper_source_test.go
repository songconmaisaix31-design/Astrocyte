package importers

import (
	"net"
	"net/url"
	"strings"
	"testing"
)

func TestPaperSourceMetadataOnly(t *testing.T) {
	snap := PaperSnapshot{
		SchemaVersion: 1, SourceURL: "https://aclanthology.org/2024.acl-long.1/", HostFamily: "acl",
		SourceKey: "https://aclanthology.org/2024.acl-long.1/", Title: "A Paper", Abstract: "Only abstract",
		ContentState: "abstract_only", Provenance: struct{ Processor, Version, Mode, Source string }{Version: "v"},
	}
	src, err := paperSource(snap, []byte("<html></html>"))
	if err != nil {
		t.Fatal(err)
	}
	if src.Kind != "paper" || src.Title != "A Paper" || src.Summary != "Only abstract" {
		t.Fatalf("metadata fields lost: %+v", src)
	}
	if !strings.Contains(src.Text, "metadata only") || strings.Contains(src.Text, "Full text") {
		t.Fatalf("metadata presented as full text: %q", src.Text)
	}
	if src.Provenance.Mode != "public_html_abstract_only" {
		t.Fatalf("wrong provenance mode: %s", src.Provenance.Mode)
	}
	if len(src.Attachments) != 1 || src.Attachments[0].Name != "source.html" {
		t.Fatalf("source HTML attachment missing: %+v", src.Attachments)
	}
}

func TestPaperSourceFulltext(t *testing.T) {
	snap := PaperSnapshot{
		SchemaVersion: 1, SourceURL: "https://journals.plos.org/x", HostFamily: "plos",
		SourceKey: "doi:10.1371/journal.pdig.0000514", Title: "Full Paper", Abstract: "Abs",
		ContentState: "readable_fulltext", Text: "Introduction Methods Results Discussion",
		Provenance: struct{ Processor, Version, Mode, Source string }{Version: "v"},
	}
	src, err := paperSource(snap, []byte("<html></html>"))
	if err != nil {
		t.Fatal(err)
	}
	if src.SourceKey != "doi:10.1371/journal.pdig.0000514" {
		t.Fatalf("source key wrong: %s", src.SourceKey)
	}
	if !strings.Contains(src.Text, "Introduction Methods Results Discussion") {
		t.Fatalf("full text missing: %q", src.Text)
	}
	if src.Provenance.Mode != "public_html_fulltext" {
		t.Fatalf("wrong provenance mode: %s", src.Provenance.Mode)
	}
}

func TestPaperSourceRestrictedRefused(t *testing.T) {
	snap := PaperSnapshot{
		SchemaVersion: 1, SourceURL: "https://example.com/article", HostFamily: "generic",
		SourceKey: "https://example.com/article", Title: "Paywalled", Abstract: "Abs",
		ContentState: "restricted",
		Provenance:   struct{ Processor, Version, Mode, Source string }{Version: "v"},
	}
	if _, err := paperSource(snap, []byte("<html></html>")); err == nil {
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
		{"::1", false},
		{"fc00::1", false},
		{"fe80::1", false},
		{"2606:4700:4700::1111", true},
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
	if u, _ := url.Parse("https://example.com/x"); validatePublicPaperURL(u) != nil {
		t.Error("rejected compliant public HTTPS URL")
	}
}
