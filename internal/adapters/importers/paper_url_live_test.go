package importers

import (
	"context"
	"os"
	"testing"

	"github.com/songconmaisaix31-design/Astrocyte/internal/attention/app"
)

// TestPaperURLLive exercises the paper_url reader adapter against real public
// pages. It runs only with an explicit opt-in; observations are metadata/fulltext
// extraction only, never a library import or model call. Full text is required
// only where the page openly publishes a complete scholarly body.
func TestPaperURLLive(t *testing.T) {
	if os.Getenv("ASTROCYTE_PAPER_URL_LIVE") != "1" {
		t.Skip("live paper_url extraction not requested")
	}
	e := installedPaperExtractor(t)
	reader := &Reader{Arxiv: NewArxiv(), Summarize: e, Web: NewPaperWeb()}
	ctx := context.Background()
	cases := []struct {
		name, url, wantState string
	}{
		{"plos-fulltext", "https://journals.plos.org/digitalhealth/article?id=10.1371/journal.pdig.0000514", "readable_fulltext"},
		{"acl-metadata", "https://aclanthology.org/2024.acl-long.1/", "abstract_only"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			src, err := reader.ReadSource(ctx, app.ImportMaterialCommand{SourceLocator: c.url, Kind: "paper", Adapter: "paper_url"})
			if err != nil {
				t.Fatalf("paper_url %s failed: %v", c.name, err)
			}
			if src.SourceKey == "" || src.Title == "" || src.Kind != "paper" {
				t.Fatalf("paper_url %s incomplete: %+v", c.name, src)
			}
			t.Logf("live paper_url %s: title=%q source_key=%s mode=%s text_bytes=%d attachments=%d", c.name, src.Title, src.SourceKey, src.Provenance.Mode, len(src.Text), len(src.Attachments))
		})
	}
}
