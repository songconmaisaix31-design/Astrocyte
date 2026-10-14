package s1_test

import (
	"bytes"
	"context"
	"strings"
	"testing"

	"github.com/songconmaisaix31-design/Astrocyte/internal/adapters/importers"
	"github.com/songconmaisaix31-design/Astrocyte/internal/attention/app"
)

// 实际官方 Atom/PDF 接口验证固定版本的各种公开 URL。
func TestPublicArxivAliasesPinOriginalVersion(t *testing.T) {
	reader := importers.NewReader(nil)
	var first app.ImportedSource
	for i, locator := range []string{"2501.12948v1", "https://arxiv.org/abs/2501.12948v1", "https://arxiv.org/pdf/2501.12948v1.pdf#page=1", "https://arxiv.org/html/2501.12948v1"} {
		source, err := reader.ReadSource(context.Background(), app.ImportMaterialCommand{Adapter: "arxiv", Kind: "paper", SourceLocator: locator})
		if err != nil {
			t.Fatalf("%s: %v", locator, err)
		}
		if source.SourceKey != "arxiv:2501.12948" || source.SourceLocator != "https://arxiv.org/abs/2501.12948v1" {
			t.Fatalf("alias changed identity/version: %+v", source)
		}
		if !strings.Contains(source.Text, "metadata only") || len(source.Attachments) != 2 || !bytes.HasPrefix(source.Attachments[0].Data, []byte("%PDF-")) || len(source.Attachments[0].Data) < 1000 || len(source.Attachments[1].Data) < 100 || source.Title == "" {
			t.Fatalf("original attachment/metadata distinction lost: %+v", source)
		}
		if i == 0 {
			first = source
		} else if first.Text != source.Text || !bytes.Equal(first.Attachments[0].Data, source.Attachments[0].Data) {
			t.Fatal("aliases changed fixed-version source content")
		}
	}
	second, err := reader.ReadSource(context.Background(), app.ImportMaterialCommand{Adapter: "arxiv", Kind: "paper", SourceLocator: "https://arxiv.org/abs/2501.12948v2"})
	if err != nil || second.SourceLocator != "https://arxiv.org/abs/2501.12948v2" || second.SourceKey != first.SourceKey || bytes.Equal(first.Attachments[0].Data, second.Attachments[0].Data) {
		t.Fatalf("actual published versions lost: %+v %v", second.Provenance, err)
	}
	_, err = reader.ReadSource(context.Background(), app.ImportMaterialCommand{Adapter: "arxiv", Kind: "paper", SourceLocator: "https://arxiv.org/abs/2501.12948v999999"})
	if err == nil {
		t.Fatal("unpublished requested version accepted another source response")
	}
}
