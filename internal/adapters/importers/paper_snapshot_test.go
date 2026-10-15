package importers

import (
	"bytes"
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

func TestPaperSnapshotFulltextProvenance(t *testing.T) {
	snapshot, _ := actualPublicPaperSnapshot(t, "plos-public.html", "https://journals.plos.org/digitalhealth/article?id=10.1371/journal.pdig.0000514")
	if snapshot.ContentState != "readable_fulltext" || len(snapshot.Text) < 30000 {
		t.Fatal("actual public PLOS full text required")
	}
	raw := snap(t, snapshot)
	src, err := paperSnapshotSource(raw, snapshot.SourceURL)
	if err != nil {
		t.Fatal(err)
	}
	if src.SourceKey != "doi:10.1371/journal.pdig.0000514" {
		t.Fatalf("source key not re-derived from DOI: %s", src.SourceKey)
	}
	if src.Provenance.Processor != "paper_snapshot" || src.Provenance.Mode != "browser_snapshot_fulltext" {
		t.Fatalf("provenance not marked as browser snapshot: %+v", src.Provenance)
	}
	if !strings.Contains(src.Text, "Full text (browser snapshot)") || !strings.Contains(src.Text, snapshot.Text) || src.Title != snapshot.Title || src.Summary != snapshot.Abstract {
		t.Fatalf("full text missing: %q", src.Text)
	}
	if len(src.Attachments) != 1 || src.Attachments[0].Name != "paper-snapshot.json" || !bytes.Equal(src.Attachments[0].Data, raw) {
		t.Fatalf("original snapshot attachment missing: %+v", src.Attachments)
	}
}

func TestPaperSnapshotAbstractOnlyNeverFullText(t *testing.T) {
	snapshot, _ := actualPublicPaperSnapshot(t, "acl-public.html", "https://aclanthology.org/2024.acl-long.1/")
	if snapshot.ContentState != "abstract_only" || snapshot.Abstract == "" || snapshot.Text != "" {
		t.Fatal("actual public ACL abstract-only observation required")
	}
	// 冲突 key 是明确规则输入，来源仍由实际 ACL URL 重新计算。
	snapshot.SourceKey = "explicit-conflicting-source-key-rule-input"
	raw := snap(t, snapshot)
	src, err := paperSnapshotSource(raw, snapshot.SourceURL)
	if err != nil {
		t.Fatal(err)
	}
	if src.SourceKey != "doi:10.18653/v1/2024.acl-long.1" {
		t.Fatalf("ACL identity not derived from URL: %s", src.SourceKey)
	}
	if src.Provenance.Mode != "browser_snapshot_abstract_only" {
		t.Fatalf("abstract marked as full text: %s", src.Provenance.Mode)
	}
	if strings.Contains(src.Text, "Full text") || !strings.Contains(src.Text, "metadata only") || !strings.Contains(src.Text, snapshot.Abstract) {
		t.Fatalf("metadata misrepresented: %q", src.Text)
	}
}

func TestPaperSnapshotTruncatedNeverFullText(t *testing.T) {
	snapshot, _ := actualPublicPaperSnapshot(t, "plos-public.html", "https://journals.plos.org/digitalhealth/article?id=10.1371/journal.pdig.0000514")
	// 截断标记是明确规则输入，不声明原始公开页面发生截断。
	snapshot.Truncated = true
	raw := snap(t, snapshot)
	src, err := paperSnapshotSource(raw, snapshot.SourceURL)
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
	snapshot, _ := actualPublicPaperSnapshot(t, "acl-public.html", "https://aclanthology.org/2024.acl-long.1/")
	// 访问限制是拒绝规则输入，不声明实际 ACL 页面受限。
	snapshot.ContentState = "restricted"
	raw := snap(t, snapshot)
	if _, err := paperSnapshotSource(raw, snapshot.SourceURL); err == nil {
		t.Fatal("restricted snapshot accepted for import")
	}
}

func TestPaperSnapshotSourceURLMustMatch(t *testing.T) {
	snapshot, _ := actualPublicPaperSnapshot(t, "acl-public.html", "https://aclanthology.org/2024.acl-long.1/")
	raw := snap(t, snapshot)
	if _, err := paperSnapshotSource(raw, "https://journals.plos.org/digitalhealth/article?id=10.1371/journal.pdig.0000514"); err == nil {
		t.Fatal("snapshot accepted with mismatched source_url")
	}
}

func TestPaperSnapshotArxivIdentity(t *testing.T) {
	snapshot, _ := actualPublicPaperSnapshot(t, "arxiv-v2-public.html", "https://arxiv.org/html/2501.12948v2")
	if snapshot.ArxivID != "2501.12948v2" || snapshot.ObservedVersion != "v2" {
		t.Fatal("actual arXiv source revision changed")
	}
	raw := snap(t, snapshot)
	src, err := paperSnapshotSource(raw, snapshot.SourceURL)
	if err != nil {
		t.Fatal(err)
	}
	if src.SourceKey != "arxiv:2501.12948" {
		t.Fatalf("arxiv identity not re-derived: %s", src.SourceKey)
	}
}

func TestPaperSnapshotReaderAdapter(t *testing.T) {
	r := NewReader(nil)
	snapshot, _ := actualPublicPaperSnapshot(t, "acl-public.html", "https://aclanthology.org/2024.acl-long.1/")
	raw := snap(t, snapshot)
	cmd := app.ImportMaterialCommand{Adapter: "paper_snapshot", Kind: "paper", SourceLocator: snapshot.SourceURL, ExportText: string(raw)}
	src, err := r.ReadSource(context.Background(), cmd)
	if err != nil {
		t.Fatal(err)
	}
	if src.SourceKey != "doi:10.18653/v1/2024.acl-long.1" || src.Kind != "paper" {
		t.Fatalf("adapter produced wrong source: %+v", src)
	}
	// 调用参数中的冲突 key 必须拒绝。
	cmd.SourceKey = "explicit-conflicting-import-key-rule-input"
	if _, err := r.ReadSource(context.Background(), cmd); err == nil {
		t.Fatal("conflicting source key accepted")
	}
	// 空导出和非论文类型必须在读取前拒绝。
	if _, err := r.ReadSource(context.Background(), app.ImportMaterialCommand{Adapter: "paper_snapshot", Kind: "paper", SourceLocator: "https://aclanthology.org/2024.acl-long.1/"}); err == nil {
		t.Fatal("empty snapshot accepted")
	}
	if _, err := r.ReadSource(context.Background(), app.ImportMaterialCommand{Adapter: "paper_snapshot", Kind: "video", SourceLocator: "https://aclanthology.org/2024.acl-long.1/", ExportText: string(raw)}); err == nil {
		t.Fatal("non-paper snapshot accepted")
	}
}
