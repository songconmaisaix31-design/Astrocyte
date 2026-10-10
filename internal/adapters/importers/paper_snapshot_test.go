package importers

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	"github.com/songconmaisaix31-design/Astrocyte/internal/attention/app"
)

func snap(t *testing.T, s PaperSnapshot) []byte {
	t.Helper()
	b, err := json.Marshal(s)
	if err != nil {
		t.Fatal(err)
	}
	return b
}

func prov(p, v, m, s string) struct{ Processor, Version, Mode, Source string } {
	return struct{ Processor, Version, Mode, Source string }{p, v, m, s}
}

func TestPaperSnapshotFulltextProvenance(t *testing.T) {
	raw := snap(t, PaperSnapshot{
		SchemaVersion: 1, SourceURL: "https://journals.plos.org/x", HostFamily: "plos",
		SourceKey: "doi:10.1371/journal.pdig.0000514", Title: "Full Paper", Abstract: "Abs",
		Authors: []string{"A"}, DOI: "10.1371/journal.pdig.0000514",
		ContentState: "readable_fulltext", Text: "Introduction Methods Results Discussion",
		Provenance: prov("summarize-readability-approach", "browser readability", "readable_fulltext", "https://journals.plos.org/x"),
	})
	src, err := paperSnapshotSource(raw, "https://journals.plos.org/x")
	if err != nil {
		t.Fatal(err)
	}
	if src.SourceKey != "doi:10.1371/journal.pdig.0000514" {
		t.Fatalf("source key not re-derived from DOI: %s", src.SourceKey)
	}
	if src.Provenance.Processor != "paper_snapshot" || src.Provenance.Mode != "browser_snapshot_fulltext" {
		t.Fatalf("provenance not marked as browser snapshot: %+v", src.Provenance)
	}
	if !strings.Contains(src.Text, "Full text (browser snapshot)") || !strings.Contains(src.Text, "Introduction Methods Results Discussion") {
		t.Fatalf("full text missing: %q", src.Text)
	}
	if len(src.Attachments) != 1 || src.Attachments[0].Name != "paper-snapshot.json" {
		t.Fatalf("original snapshot attachment missing: %+v", src.Attachments)
	}
}

func TestPaperSnapshotAbstractOnlyNeverFullText(t *testing.T) {
	// The snapshot's self-reported source_key must not be trusted; a hand-edited
	// key is ignored in favor of the re-derived ACL URL identity.
	raw := snap(t, PaperSnapshot{
		SchemaVersion: 1, SourceURL: "https://aclanthology.org/2024.acl-long.1/", HostFamily: "acl",
		SourceKey: "doi:10.9999/fake", Title: "A Paper", Abstract: "Only abstract",
		DOI: "10.18653/v1/2024.acl-long.1", ContentState: "abstract_only",
		Provenance: prov("summarize-readability-approach", "browser", "abstract_only", "https://aclanthology.org/2024.acl-long.1/"),
	})
	src, err := paperSnapshotSource(raw, "https://aclanthology.org/2024.acl-long.1/")
	if err != nil {
		t.Fatal(err)
	}
	if src.SourceKey != "doi:10.18653/v1/2024.acl-long.1" {
		t.Fatalf("ACL identity not derived from URL: %s", src.SourceKey)
	}
	if src.Provenance.Mode != "browser_snapshot_abstract_only" {
		t.Fatalf("abstract marked as full text: %s", src.Provenance.Mode)
	}
	if strings.Contains(src.Text, "Full text") || !strings.Contains(src.Text, "metadata only") {
		t.Fatalf("metadata misrepresented: %q", src.Text)
	}
}

func TestPaperSnapshotTruncatedNeverFullText(t *testing.T) {
	raw := snap(t, PaperSnapshot{
		SchemaVersion: 1, SourceURL: "https://journals.plos.org/x", HostFamily: "plos",
		Title: "T", Abstract: "A", DOI: "10.1371/journal.pdig.0000514",
		ContentState: "readable_fulltext", Text: "partial body", Truncated: true,
		Provenance: prov("x", "v", "readable_fulltext", "https://journals.plos.org/x"),
	})
	src, err := paperSnapshotSource(raw, "https://journals.plos.org/x")
	if err != nil {
		t.Fatal(err)
	}
	if src.Provenance.Mode != "browser_snapshot_truncated" {
		t.Fatalf("truncated snapshot marked as full text: %s", src.Provenance.Mode)
	}
	if !strings.Contains(src.Text, "truncated") {
		t.Fatalf("truncation not recorded in content: %q", src.Text)
	}
}

func TestPaperSnapshotRestrictedRefused(t *testing.T) {
	raw := snap(t, PaperSnapshot{
		SchemaVersion: 1, SourceURL: "https://example.com/a", HostFamily: "generic",
		Title: "T", Abstract: "A", ContentState: "restricted",
		Provenance: prov("x", "v", "restricted", "https://example.com/a"),
	})
	if _, err := paperSnapshotSource(raw, "https://example.com/a"); err == nil {
		t.Fatal("restricted snapshot accepted for import")
	}
}

func TestPaperSnapshotSourceURLMustMatch(t *testing.T) {
	raw := snap(t, PaperSnapshot{
		SchemaVersion: 1, SourceURL: "https://example.com/a", HostFamily: "generic",
		Title: "T", Abstract: "A", ContentState: "abstract_only",
		Provenance: prov("x", "v", "abstract_only", "https://example.com/a"),
	})
	if _, err := paperSnapshotSource(raw, "https://example.com/b"); err == nil {
		t.Fatal("snapshot accepted with mismatched source_url")
	}
}

func TestPaperSnapshotArxivIdentity(t *testing.T) {
	raw := snap(t, PaperSnapshot{
		SchemaVersion: 1, SourceURL: "https://arxiv.org/abs/2504.16054v2", HostFamily: "arxiv",
		SourceKey: "arxiv:2504.16054", Title: "T", Abstract: "A",
		ArxivID: "2504.16054v2", ObservedVersion: "v2", PDFURLs: []string{"https://arxiv.org/pdf/2504.16054v2"},
		ContentState: "abstract_only",
		Provenance:   prov("x", "v", "abstract_only", "https://arxiv.org/abs/2504.16054v2"),
	})
	src, err := paperSnapshotSource(raw, "https://arxiv.org/abs/2504.16054v2")
	if err != nil {
		t.Fatal(err)
	}
	if src.SourceKey != "arxiv:2504.16054" {
		t.Fatalf("arxiv identity not re-derived: %s", src.SourceKey)
	}
}

func TestPaperSnapshotReaderAdapter(t *testing.T) {
	r := NewReader(nil)
	raw := snap(t, PaperSnapshot{
		SchemaVersion: 1, SourceURL: "https://aclanthology.org/2024.acl-long.1/", HostFamily: "acl",
		Title: "A Paper", Abstract: "Abs", DOI: "10.18653/v1/2024.acl-long.1",
		ContentState: "abstract_only",
		Provenance:   prov("x", "v", "abstract_only", "https://aclanthology.org/2024.acl-long.1/"),
	})
	cmd := app.ImportMaterialCommand{Adapter: "paper_snapshot", Kind: "paper", SourceLocator: "https://aclanthology.org/2024.acl-long.1/", ExportText: string(raw)}
	src, err := r.ReadSource(context.Background(), cmd)
	if err != nil {
		t.Fatal(err)
	}
	if src.SourceKey != "doi:10.18653/v1/2024.acl-long.1" || src.Kind != "paper" {
		t.Fatalf("adapter produced wrong source: %+v", src)
	}
	// A conflicting caller-supplied source key must be rejected, never merged.
	cmd.SourceKey = "doi:10.9999/other"
	if _, err := r.ReadSource(context.Background(), cmd); err == nil {
		t.Fatal("conflicting source key accepted")
	}
	// No export and a non-paper kind are rejected up front.
	if _, err := r.ReadSource(context.Background(), app.ImportMaterialCommand{Adapter: "paper_snapshot", Kind: "paper", SourceLocator: "https://aclanthology.org/2024.acl-long.1/"}); err == nil {
		t.Fatal("empty snapshot accepted")
	}
	if _, err := r.ReadSource(context.Background(), app.ImportMaterialCommand{Adapter: "paper_snapshot", Kind: "video", SourceLocator: "https://aclanthology.org/2024.acl-long.1/", ExportText: string(raw)}); err == nil {
		t.Fatal("non-paper snapshot accepted")
	}
}
